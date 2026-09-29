package management

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	xiaohuanxiongauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// RequestXiaohuanxiongToken starts the SenseTime Xiaohuanxiong (商汤小浣熊 /
// Raccoon) browser authorization flow.
//
// Xiaohuanxiong has no standard OAuth2 authorization server. The desktop client
// opens a web login page that redirects to an office-raccoon:// deep link
// carrying a one-time authorization code, then exchanges that code for tokens.
//
// This endpoint returns the login URL plus a state and waits for the code in the
// background. The flow completes through whichever of these the caller can use:
//
//   - POST /v0/management/oauth-callback with provider=xiaohuanxiong
//   - GET  /v0/management/oauth-callback?provider=xiaohuanxiong&code=...
//   - POST /v0/management/xiaohuanxiong-auth-callback with the callback URL
func (h *Handler) RequestXiaohuanxiongToken(c *gin.Context) {
	ctx := PopulateAuthContext(context.Background(), c)

	state, errState := misc.GenerateRandomState()
	if errState != nil {
		log.Errorf("failed to generate state parameter: %v", errState)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state parameter"})
		return
	}

	// An explicit login_url override keeps self-hosted or branded deployments
	// working without a rebuild.
	authURL := xiaohuanxiongauth.AuthorizationURL(strings.TrimSpace(c.Query("login_url")))
	if authURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build authorization url"})
		return
	}

	RegisterOAuthSession(state, constant.Xiaohuanxiong)

	callbackPath := xiaohuanxiongCallbackFilePath(h, state)
	go runXiaohuanxiongLogin(ctx, h, state, callbackPath)

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"url":      authURL,
		"state":    state,
		"flow":     "callback",
		"callback": xiaohuanxiongauth.CallbackURL(),
	})
}

// xiaohuanxiongCallbackFilePath mirrors writeOAuthCallbackFile's naming so this
// handler can wait on the file the shared callback endpoints publish.
func xiaohuanxiongCallbackFilePath(h *Handler, state string) string {
	authDir := ""
	if h != nil && h.cfg != nil {
		authDir = h.cfg.AuthDir
	}
	if strings.TrimSpace(authDir) == "" {
		return ""
	}
	return filepath.Join(authDir, fmt.Sprintf(".oauth-%s-%s.oauth", constant.Xiaohuanxiong, state))
}

// runXiaohuanxiongLogin waits for the authorization callback, exchanges the
// one-time code for tokens, and persists the credential record.
func runXiaohuanxiongLogin(ctx context.Context, h *Handler, state, callbackPath string) {
	code, errWait := waitXiaohuanxiongCallback(ctx, state, callbackPath)
	if errWait != nil {
		if !errors.Is(errWait, errOAuthSessionNotPending) {
			SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errWait))
			fmt.Printf("Xiaohuanxiong authentication failed: %v\n", errWait)
		}
		return
	}
	if errComplete := completeXiaohuanxiongLogin(ctx, h, state, code); errComplete != nil {
		fmt.Printf("Xiaohuanxiong authentication failed: %v\n", errComplete)
	}
}

// waitXiaohuanxiongCallback blocks until the callback file appears or the
// pending session ends.
func waitXiaohuanxiongCallback(ctx context.Context, state, callbackPath string) (string, error) {
	if strings.TrimSpace(callbackPath) == "" {
		return "", errors.New("auth directory is not configured")
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(oauthSessionTTL)
	defer deadline.Stop()

	for {
		if !IsOAuthSessionPending(state, constant.Xiaohuanxiong) {
			return "", errOAuthSessionNotPending
		}
		if ctx.Err() != nil {
			return "", errors.New("timeout waiting for OAuth callback")
		}

		data, errRead := os.ReadFile(callbackPath)
		if errRead == nil {
			var payload oauthCallbackFilePayload
			if errDecode := json.Unmarshal(data, &payload); errDecode != nil {
				return "", errors.New("invalid OAuth callback")
			}
			if strings.TrimSpace(payload.Error) != "" {
				return "", fmt.Errorf("authorization error: %s", strings.TrimSpace(payload.Error))
			}
			if strings.TrimSpace(payload.Code) == "" {
				return "", errors.New("OAuth callback has no code")
			}
			return strings.TrimSpace(payload.Code), nil
		}
		if !errors.Is(errRead, os.ErrNotExist) {
			return "", errors.New("failed to read OAuth callback")
		}

		select {
		case <-ctx.Done():
			return "", errors.New("timeout waiting for OAuth callback")
		case <-deadline.C:
			return "", errors.New("timeout waiting for OAuth callback")
		case <-ticker.C:
		}
	}
}

// completeXiaohuanxiongLogin exchanges a one-time code and persists the
// credential. It returns an error for every non-success path so callers never
// have to infer the outcome from session state.
func completeXiaohuanxiongLogin(ctx context.Context, h *Handler, state, code string) error {
	if errGuard := guardOAuthSessionPendingForSave(state, constant.Xiaohuanxiong); errGuard != nil {
		return errGuard
	}

	client := xiaohuanxiongauth.NewClient(nil)
	token, errExchange := client.ExchangeAuthorizationCode(ctx, code)
	if errExchange != nil {
		log.Errorf("Xiaohuanxiong token exchange failed: %v", errExchange)
		SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errExchange))
		return errExchange
	}
	if !IsOAuthSessionPending(state, constant.Xiaohuanxiong) {
		return errOAuthSessionNotPending
	}

	fileName := xiaohuanxiongAuthFileName(token)
	metadata := xiaohuanxiongMetadata(token)
	tokenStorage := &xiaohuanxiongauth.TokenStorage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		Expired:        xiaohuanxiongauth.ExpiryRFC3339(token.AccessToken),
		OfficeIdentity: token.OfficeIdentity,
		OfficeOrgName:  token.OfficeOrgName,
		OfficeOrgRole:  token.OfficeOrgRole,
	}
	tokenStorage.SetMetadata(metadata)

	record := &coreauth.Auth{
		ID:         fileName,
		Provider:   constant.Xiaohuanxiong,
		FileName:   fileName,
		Label:      "Xiaohuanxiong User",
		Storage:    tokenStorage,
		Metadata:   metadata,
		Attributes: xiaohuanxiongAttributes(token),
	}

	if errGuard := guardOAuthSessionPendingForSave(state, constant.Xiaohuanxiong); errGuard != nil {
		return errGuard
	}
	savedPath, errSave := h.saveTokenRecord(ctx, record)
	if errSave != nil {
		log.Errorf("Failed to save Xiaohuanxiong authentication tokens: %v", errSave)
		SetOAuthSessionError(state, "Failed to save authentication tokens")
		return errSave
	}

	CompleteOAuthSession(state)
	fmt.Printf("Xiaohuanxiong authentication successful! Token saved to %s\n", savedPath)
	return nil
}

// PostXiaohuanxiongAuthCallback accepts a pasted callback URL (or a bare
// authorization code) and completes the flow. It is the escape hatch for
// environments where the office-raccoon deep link cannot reach the proxy.
func (h *Handler) PostXiaohuanxiongAuthCallback(c *gin.Context) {
	var req struct {
		State       string `json:"state"`
		Code        string `json:"code"`
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
	if errGuard := guardOAuthSessionPendingForSave(state, constant.Xiaohuanxiong); errGuard != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "error": errGuard.Error()})
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		parsed, errParse := xiaohuanxiongauth.ParseCallbackCode(req.RedirectURL)
		if errParse != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "failed to parse callback url"})
			return
		}
		code = parsed
	}

	ctx := PopulateAuthContext(context.Background(), c)
	if errComplete := completeXiaohuanxiongLogin(ctx, h, state, code); errComplete != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": errComplete.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// xiaohuanxiongMetadata builds the persisted metadata map for a login.
func xiaohuanxiongMetadata(token *xiaohuanxiongauth.TokenData) map[string]any {
	metadata := map[string]any{
		"type":       constant.Xiaohuanxiong,
		"auth_kind":  coreauth.AuthKindOAuth,
		"timestamp":  time.Now().UnixMilli(),
		"base_url":   xiaohuanxiongauth.LLMBaseURL,
		"token_host": xiaohuanxiongauth.BaseURL,
	}
	if strings.TrimSpace(token.AccessToken) != "" {
		metadata["access_token"] = token.AccessToken
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if expiry := xiaohuanxiongauth.ExpiryRFC3339(token.AccessToken); expiry != "" {
		metadata["expired"] = expiry
	}
	if strings.TrimSpace(token.OfficeIdentity) != "" {
		metadata["office_identity"] = token.OfficeIdentity
	}
	if strings.TrimSpace(token.OfficeOrgName) != "" {
		metadata["office_org_name"] = token.OfficeOrgName
	}
	if strings.TrimSpace(token.OfficeOrgRole) != "" {
		metadata["office_org_role"] = token.OfficeOrgRole
	}
	return metadata
}

// xiaohuanxiongAttributes builds the immutable attributes for a login record.
// The executor reads base_url and api_key from here.
func xiaohuanxiongAttributes(token *xiaohuanxiongauth.TokenData) map[string]string {
	attributes := map[string]string{
		coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
		"api_key":                  token.AccessToken,
		"base_url":                 xiaohuanxiongauth.LLMBaseURL,
		"source":                   "oauth:xiaohuanxiong",
	}
	if strings.TrimSpace(token.OfficeIdentity) != "" {
		attributes["office_identity"] = token.OfficeIdentity
	}
	return attributes
}

// xiaohuanxiongAuthFileName derives a stable file name from the account identity
// so a repeated login for the same account overwrites the existing credential.
func xiaohuanxiongAuthFileName(token *xiaohuanxiongauth.TokenData) string {
	seed := strings.TrimSpace(token.OfficeIdentity)
	if seed == "" {
		seed = strings.TrimSpace(token.OfficeOrgName)
	}
	if seed == "" {
		seed = token.AccessToken
	}
	digest := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("xiaohuanxiong-%s.json", hex.EncodeToString(digest[:16]))
}
