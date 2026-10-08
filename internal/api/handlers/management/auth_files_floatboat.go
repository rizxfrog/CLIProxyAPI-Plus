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
	floatboatauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/floatboat"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// RequestFloatboatToken starts the FloatBoat (aoe.chat) login flow.
//
// FloatBoat returns the authorization code on the aoe:// deep link registered by
// its desktop app, which CLIProxyAPI cannot own. The Web UI and TUI therefore
// show the sign-in URL and submit the pasted callback URL through the shared
// /v0/management/oauth-callback endpoint, which writes a callback file consumed
// by the waiter started here.
func (h *Handler) RequestFloatboatToken(c *gin.Context) {
	state, errState := misc.GenerateRandomState()
	if errState != nil {
		log.Errorf("Failed to generate FloatBoat state: %v", errState)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start FloatBoat authorization"})
		return
	}
	client := floatboatauth.NewClient(h.cfg)
	loginURL := client.BuildSignInURL(state)
	RegisterOAuthSession(state, floatboatauth.Provider)
	go h.waitFloatboatCallback(state)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"url":    loginURL,
		"state":  state,
		"flow":   "manual",
	})
}

// waitFloatboatCallback blocks until the pasted callback file appears (or the
// session is cancelled or times out), then completes the credential exchange.
func (h *Handler) waitFloatboatCallback(state string) {
	waitFile := filepath.Join(h.cfg.AuthDir, fmt.Sprintf(".oauth-%s-%s.oauth", floatboatauth.Provider, state))
	deadline := time.Now().Add(10 * time.Minute)
	for {
		if !IsOAuthSessionPending(state, floatboatauth.Provider) {
			return
		}
		if time.Now().After(deadline) {
			SetOAuthSessionError(state, "Timeout waiting for FloatBoat callback")
			return
		}
		data, errRead := os.ReadFile(waitFile)
		if errRead == nil {
			_ = os.Remove(waitFile)
			var payload map[string]string
			_ = json.Unmarshal(data, &payload)
			if errComplete := h.completeFloatboatLogin(state, payload["code"], payload["error"]); errComplete != nil {
				log.Errorf("FloatBoat authentication failed: %v", errComplete)
				SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errComplete))
				return
			}
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// PostFloatboatAuthCallback receives the FloatBoat callback URL pasted by the
// user, exchanges the code, mints the inference key, and persists the record.
func (h *Handler) PostFloatboatAuthCallback(c *gin.Context) {
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
	if errGuard := guardOAuthSessionPendingForSave(state, floatboatauth.Provider); errGuard != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "error": errGuard.Error()})
		return
	}
	redirectURL := strings.TrimSpace(req.RedirectURL)
	if redirectURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "redirect_url is required"})
		return
	}

	code, returnedState, errParse := parseFloatboatCallbackURL(redirectURL)
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

	if errComplete := h.completeFloatboatLogin(state, code, ""); errComplete != nil {
		log.Errorf("FloatBoat callback processing failed: %v", errComplete)
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": "failed to complete FloatBoat authentication"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// parseFloatboatCallbackURL extracts the authorization code and state from a
// pasted aoe:// deep link or an equivalent https callback URL.
func parseFloatboatCallbackURL(raw string) (string, string, error) {
	parsed, errParse := url.Parse(strings.TrimSpace(raw))
	if errParse != nil {
		return "", "", fmt.Errorf("failed to parse FloatBoat callback URL")
	}
	query := parsed.Query()
	if errParam := strings.TrimSpace(query.Get("error")); errParam != "" {
		return "", "", fmt.Errorf("FloatBoat authorization failed: %s", errParam)
	}
	code := strings.TrimSpace(query.Get("code"))
	if code == "" {
		return "", "", fmt.Errorf("callback URL missing code")
	}
	return code, strings.TrimSpace(query.Get("state")), nil
}

// completeFloatboatLogin exchanges the authorization code, mints the inference
// key, and persists the credential. It is shared by the callback-file waiter and
// the direct POST endpoint.
func (h *Handler) completeFloatboatLogin(state, code, errParam string) error {
	if trimmed := strings.TrimSpace(errParam); trimmed != "" {
		return fmt.Errorf("authorization failed: %s", trimmed)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("missing FloatBoat authorization code")
	}
	ctx := context.Background()
	client := floatboatauth.NewClient(h.cfg)
	token, errExchange := client.ExchangeCode(ctx, code, state, "")
	if errExchange != nil {
		return errExchange
	}
	profile, errProfile := client.FetchUserProfile(ctx, token.AccessToken)
	if errProfile != nil {
		log.Warnf("FloatBoat user profile fetch failed: %v", errProfile)
		profile = &floatboatauth.UserProfile{}
	}
	apiKey, errAPIKey := client.FetchNewAPIKey(ctx, token.AccessToken)
	if errAPIKey != nil {
		log.Warnf("FloatBoat newapi key fetch failed: %v", errAPIKey)
	}
	record := buildFloatboatRecord(token, profile, apiKey)
	if errGuard := guardOAuthSessionPendingForSave(state, floatboatauth.Provider); errGuard != nil {
		return errGuard
	}
	savedPath, errSave := h.saveTokenRecord(ctx, record)
	if errSave != nil {
		return errSave
	}
	CompleteOAuthSession(state)
	fmt.Printf("FloatBoat authentication successful! Token saved to %s\n", savedPath)
	return nil
}

// buildFloatboatRecord assembles the persisted credential from the exchanged
// token, the account profile, and the minted inference key.
func buildFloatboatRecord(token *floatboatauth.TokenData, profile *floatboatauth.UserProfile, apiKey string) *coreauth.Auth {
	fileName := floatboatFileName(profile)
	metadata := map[string]any{
		"type":         floatboatauth.Provider,
		"auth_kind":    "oauth",
		"access_token": token.AccessToken,
		"base_url":     floatboatauth.DefaultInferenceBaseURL,
		"backend_url":  floatboatauth.DefaultBackendURL,
		"timestamp":    time.Now().UnixMilli(),
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		metadata["expires_in"] = token.ExpiresIn
	}
	if token.RefreshExpiresIn > 0 {
		metadata["refresh_expires_in"] = token.RefreshExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(apiKey) != "" {
		metadata["api_key"] = apiKey
	}
	if profile != nil {
		if profile.ID != "" {
			metadata["user_id"] = profile.ID
		}
		if profile.Email != "" {
			metadata["email"] = profile.Email
			metadata["user_email"] = profile.Email
		}
		if profile.Name != "" {
			metadata["user_name"] = profile.Name
		}
		if profile.Image != "" {
			metadata["user_avatar"] = profile.Image
		}
		if profile.MembershipLevel != "" {
			metadata["membership_level"] = profile.MembershipLevel
		}
		metadata["has_active_subscription"] = profile.HasSubscription
		metadata["quota_remaining"] = profile.Credits
	}

	attributes := map[string]string{
		coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
		"base_url":                 floatboatauth.DefaultInferenceBaseURL,
		"access_token":             token.AccessToken,
	}
	if apiKey != "" {
		attributes["api_key"] = apiKey
	}
	return &coreauth.Auth{
		ID:         fileName,
		Provider:   floatboatauth.Provider,
		FileName:   fileName,
		Label:      floatboatLabel(profile),
		Metadata:   metadata,
		Attributes: attributes,
	}
}

// floatboatLabel derives a display label for the credential.
func floatboatLabel(profile *floatboatauth.UserProfile) string {
	if profile != nil {
		if profile.Email != "" {
			return profile.Email
		}
		if profile.ID != "" {
			return profile.ID
		}
	}
	return floatboatauth.Label
}

// floatboatFileName derives a stable, filesystem-safe credential file name.
func floatboatFileName(profile *floatboatauth.UserProfile) string {
	identity := ""
	if profile != nil {
		identity = strings.TrimSpace(profile.ID)
		if identity == "" {
			identity = strings.TrimSpace(profile.Email)
		}
	}
	if identity == "" {
		identity = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", "@", "_at_", " ", "_").Replace(identity)
	return fmt.Sprintf("floatboat-%s.json", safe)
}
