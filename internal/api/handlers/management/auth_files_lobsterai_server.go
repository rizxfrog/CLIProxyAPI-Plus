package management

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// RequestLobsteraiServerToken starts the LobsterAI login flow.
//
// The proxy serves a loopback callback so a local browser completes the login
// without any copying. When the port is unavailable, or the operator's browser
// cannot reach this host (remote management), the pasted-callback path is still
// offered: the Web UI and TUI can submit the callback URL through the shared
// /v0/management/oauth-callback endpoint, or PostLobsteraiServerAuthCallback.
func (h *Handler) RequestLobsteraiServerToken(c *gin.Context) {
	state, errState := misc.GenerateRandomState()
	if errState != nil {
		log.Errorf("Failed to generate LobsterAI state: %v", errState)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start LobsterAI authorization"})
		return
	}
	client := lobsteraiserverauth.NewClient(h.cfg)
	RegisterOAuthSession(state, lobsteraiserverauth.Provider)

	// Prefer a real loopback listener so a local browser finishes the login by
	// itself. Its absence only degrades the flow to pasting the callback URL.
	callbackServer := lobsteraiserverauth.NewOAuthServer(lobsteraiserverauth.DefaultLoopbackCallbackPort)
	redirectURI := ""
	if port, errStart := callbackServer.Start(); errStart == nil {
		redirectURI = fmt.Sprintf("http://127.0.0.1:%d%s", port, lobsteraiserverauth.CallbackPath)
		go h.waitLobsteraiServerLoopback(state, callbackServer)
	} else {
		log.Warnf("LobsterAI: loopback callback unavailable, falling back to paste: %v", errStart)
	}

	loginURL := client.BuildLoginURL(state, redirectURI)
	go h.waitLobsteraiServerCallback(state)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"url":    loginURL,
		"state":  state,
		"flow":   "manual",
	})
}

// waitLobsteraiServerLoopback captures the browser redirect on the local
// callback server and completes the login, then releases the listener.
func (h *Handler) waitLobsteraiServerLoopback(state string, server *lobsteraiserverauth.OAuthServer) {
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if errStop := server.Stop(stopCtx); errStop != nil {
			log.Warnf("LobsterAI: stop loopback callback server: %v", errStop)
		}
	}()
	result, errWait := server.WaitForCallback(10 * time.Minute)
	if errWait != nil {
		return
	}
	if result == nil {
		return
	}
	if result.Error != "" {
		SetOAuthSessionError(state, "Authentication failed: "+result.Error)
		return
	}
	if result.State != "" && result.State != state {
		SetOAuthSessionError(state, "State verification failed")
		return
	}
	if errComplete := h.completeLobsteraiServerLogin(state, result.Code, ""); errComplete != nil {
		log.Errorf("LobsterAI loopback authentication failed: %v", errComplete)
		SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errComplete))
	}
}

// waitLobsteraiServerCallback blocks until the pasted callback file appears (or
// the session is cancelled or times out), then completes the credential exchange.
func (h *Handler) waitLobsteraiServerCallback(state string) {
	waitFile := filepath.Join(h.cfg.AuthDir, fmt.Sprintf(".oauth-%s-%s.oauth", lobsteraiserverauth.Provider, state))
	deadline := time.Now().Add(10 * time.Minute)
	for {
		if !IsOAuthSessionPending(state, lobsteraiserverauth.Provider) {
			return
		}
		if time.Now().After(deadline) {
			SetOAuthSessionError(state, "Timeout waiting for LobsterAI callback")
			return
		}
		data, errRead := os.ReadFile(waitFile)
		if errRead == nil {
			_ = os.Remove(waitFile)
			var payload map[string]string
			_ = json.Unmarshal(data, &payload)
			if errComplete := h.completeLobsteraiServerLogin(state, payload["code"], payload["error"]); errComplete != nil {
				log.Errorf("LobsterAI authentication failed: %v", errComplete)
				SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errComplete))
				return
			}
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// PostLobsteraiServerAuthCallback receives the LobsterAI callback URL pasted by
// the user, exchanges the code, and persists the credential.
func (h *Handler) PostLobsteraiServerAuthCallback(c *gin.Context) {
	var req struct {
		State       string `json:"state"`
		RedirectURL string `json:"redirect_url"`
	}
	if errBind := c.ShouldBindJSON(&req); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid body"})
		return
	}
	state := strings.TrimSpace(req.State)
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "state is required"})
		return
	}
	if errState := ValidateOAuthState(state); errState != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid state"})
		return
	}
	if errGuard := guardOAuthSessionPendingForSave(state, lobsteraiserverauth.Provider); errGuard != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "error": errGuard.Error()})
		return
	}
	redirectURL := strings.TrimSpace(req.RedirectURL)
	if redirectURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "redirect_url is required"})
		return
	}

	code, returnedState, errParse := parseLobsteraiServerCallbackURL(redirectURL)
	if errParse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": errParse.Error()})
		return
	}
	if returnedState == "" {
		returnedState = state
	} else if returnedState != state {
		SetOAuthSessionError(state, "State verification failed")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "state mismatch"})
		return
	}

	if errComplete := h.completeLobsteraiServerLogin(state, code, ""); errComplete != nil {
		log.Errorf("LobsterAI callback processing failed: %v", errComplete)
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": "failed to complete LobsterAI authentication"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// parseLobsteraiServerCallbackURL extracts the authorization code and state from
// a pasted callback URL.
func parseLobsteraiServerCallbackURL(raw string) (string, string, error) {
	parsed, errParse := url.Parse(strings.TrimSpace(raw))
	if errParse != nil {
		return "", "", fmt.Errorf("failed to parse LobsterAI callback URL")
	}
	query := parsed.Query()
	if errParam := strings.TrimSpace(query.Get("error")); errParam != "" {
		description := strings.TrimSpace(query.Get("error_description"))
		if description != "" {
			return "", "", fmt.Errorf("LobsterAI authorization failed (%s: %s)", errParam, description)
		}
		return "", "", fmt.Errorf("LobsterAI authorization failed (%s)", errParam)
	}
	code := strings.TrimSpace(query.Get("code"))
	if code == "" {
		return "", "", fmt.Errorf("callback URL missing code")
	}
	return code, strings.TrimSpace(query.Get("state")), nil
}

// completeLobsteraiServerLogin exchanges the authorization code and persists the
// credential. It is shared by the callback-file waiter and the direct POST
// endpoint.
func (h *Handler) completeLobsteraiServerLogin(state, code, errParam string) error {
	if trimmed := strings.TrimSpace(errParam); trimmed != "" {
		return fmt.Errorf("authorization failed: %s", trimmed)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("missing LobsterAI authorization code")
	}
	ctx := context.Background()
	client := lobsteraiserverauth.NewClient(h.cfg)
	token, user, quota, errExchange := client.ExchangeCode(ctx, code, state)
	if errExchange != nil {
		return errExchange
	}
	// The exchange carries the numeric account id but not the login identity;
	// /api/user/profile fills in yid (the email) and nickname. It is best-effort
	// so a profile outage cannot fail an otherwise valid login.
	if profile, errProfile := client.FetchUserProfile(ctx, token.AccessToken); errProfile != nil {
		log.Warnf("LobsterAI user profile fetch failed: %v", errProfile)
	} else {
		user = mergeLobsteraiServerUser(user, profile)
	}
	// An enterprise account must send its identity on every request, so the
	// context is read once here and persisted with the credential. Failure leaves
	// the account personal, which is the correct default for the common case.
	enterprise, errEnterprise := client.FetchEnterpriseContext(ctx, token.AccessToken)
	if errEnterprise != nil {
		log.Warnf("LobsterAI enterprise context fetch failed: %v", errEnterprise)
	}
	record := buildLobsteraiServerRecord(token, user, quota, enterprise)
	if errGuard := guardOAuthSessionPendingForSave(state, lobsteraiserverauth.Provider); errGuard != nil {
		return errGuard
	}
	savedPath, errSave := h.saveTokenRecord(ctx, record)
	if errSave != nil {
		return errSave
	}
	CompleteOAuthSession(state)
	fmt.Printf("LobsterAI authentication successful! Token saved to %s\n", savedPath)
	return nil
}

// buildLobsteraiServerRecord assembles the persisted credential from the
// exchanged token and account identity.
func buildLobsteraiServerRecord(token *lobsteraiserverauth.TokenData, user *lobsteraiserverauth.User, quota *lobsteraiserverauth.Quota, enterprise *lobsteraiserverauth.EnterpriseContext) *coreauth.Auth {
	fileName := lobsteraiServerRecordFileName(user)
	metadata := map[string]any{
		"type":           lobsteraiserverauth.Provider,
		"auth_kind":      "oauth",
		"access_token":   token.AccessToken,
		"base_url":       lobsteraiserverauth.DefaultProxyBaseURL,
		"backend_url":    lobsteraiserverauth.DefaultBackendURL,
		"client_version": lobsteraiserverauth.ClientVersion,
		"timestamp":      time.Now().UnixMilli(),
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		metadata["expires_in"] = token.ExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if user != nil {
		if user.ID != "" {
			metadata["user_id"] = user.ID
		}
		if user.Email != "" {
			metadata["email"] = user.Email
			metadata["user_email"] = user.Email
		}
		if user.Name != "" {
			metadata["user_name"] = user.Name
		}
		if user.Avatar != "" {
			metadata["user_avatar"] = user.Avatar
		}
	}
	if quota != nil {
		if quota.SubscriptionStatus != "" {
			metadata["subscription_status"] = quota.SubscriptionStatus
		}
		if quota.CreditsKnown {
			metadata["quota_remaining"] = quota.Credits
		}
		if quota.PurchasedCredits > 0 {
			metadata["purchased_credits"] = quota.PurchasedCredits
		}
	}
	// A personal account persists no enterprise identity, which is what makes the
	// proxy omit the two enterprise headers on its requests.
	if enterprise != nil && enterprise.EnterpriseID != "" {
		metadata["account_mode"] = lobsteraiserverauth.AccountModeEnterprise
		metadata["enterprise_id"] = enterprise.EnterpriseID
		if enterprise.EnterpriseName != "" {
			metadata["enterprise_name"] = enterprise.EnterpriseName
		}
		if enterprise.Role != "" {
			metadata["enterprise_role"] = enterprise.Role
		}
		if enterprise.MemberID != "" {
			metadata["enterprise_member_id"] = enterprise.MemberID
		}
	}

	attributes := map[string]string{
		coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
		"base_url":                 lobsteraiserverauth.DefaultProxyBaseURL,
		"backend_url":              lobsteraiserverauth.DefaultBackendURL,
		"access_token":             token.AccessToken,
	}
	return &coreauth.Auth{
		ID:         fileName,
		Provider:   lobsteraiserverauth.Provider,
		FileName:   fileName,
		Label:      lobsteraiServerRecordLabel(user),
		Metadata:   metadata,
		Attributes: attributes,
	}
}

// mergeLobsteraiServerUser fills missing identity fields from the profile read.
// The exchange response carries the numeric account id while /api/user/profile
// carries yid (the login email) and nickname, so neither alone is complete.
func mergeLobsteraiServerUser(base, extra *lobsteraiserverauth.User) *lobsteraiserverauth.User {
	if base == nil {
		return extra
	}
	if extra == nil {
		return base
	}
	if base.ID == "" {
		base.ID = extra.ID
	}
	if base.Email == "" {
		base.Email = extra.Email
	}
	if base.Name == "" {
		base.Name = extra.Name
	}
	if base.Avatar == "" {
		base.Avatar = extra.Avatar
	}
	if base.Nickname == "" {
		base.Nickname = extra.Nickname
	}
	return base
}

// lobsteraiServerRecordLabel derives a display label for the credential.
func lobsteraiServerRecordLabel(user *lobsteraiserverauth.User) string {
	if user != nil {
		if user.Email != "" {
			return user.Email
		}
		if user.Nickname != "" {
			return user.Nickname
		}
		if user.Name != "" {
			return user.Name
		}
		if user.ID != "" {
			return user.ID
		}
	}
	return lobsteraiserverauth.Label
}

// lobsteraiServerRecordFileName derives a stable, filesystem-safe credential
// file name.
func lobsteraiServerRecordFileName(user *lobsteraiserverauth.User) string {
	identity := ""
	if user != nil {
		identity = strings.TrimSpace(user.ID)
		if identity == "" {
			identity = strings.TrimSpace(user.Email)
		}
	}
	if identity == "" {
		identity = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", "@", "_at_", " ", "_", ":", "_").Replace(identity)
	return fmt.Sprintf("lobsterai-server-%s.json", safe)
}
