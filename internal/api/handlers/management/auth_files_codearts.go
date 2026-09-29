package management

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// codeartsOAuthService abstracts the credential acquisition used by the login
// handler so tests can substitute a fake.
type codeartsOAuthService interface {
	BuildAuthorizationURL(verifier, port, ticketID string) (string, error)
	ExchangeCode(ctx context.Context, code, verifier string, keyPair *codeartsauth.DpopKeyPair, port string) (*codeartsauth.TokenData, error)
	GetUserInfo(ctx context.Context, accessKey, secretKey, securityToken string) (*codeartsauth.UserInfo, error)
}

var newCodeArtsOAuthService = func(cfg *config.Config) codeartsOAuthService {
	if cfg != nil {
		return codeartsauth.NewClient(cfg)
	}
	return codeartsauth.NewClient(nil)
}

// RequestCodeArtsToken starts the Huawei Cloud CodeArts browser login flow.
//
// CodeArts has no standalone OAuth server: the desktop client opens the CodeArts
// portal authorize page and listens on a loopback callback. The portal redirects
// back with an authorization code, which is exchanged against Huawei Cloud STS
// for a temporary AK/SK/security-token triple (OAuth2 + PKCE + DPoP).
//
// This endpoint returns the login URL plus a state and waits for the code in the
// background. The flow completes through whichever of these the caller can use:
//
//   - POST /v0/management/oauth-callback with provider=codearts
//   - GET  /v0/management/oauth-callback?provider=codearts&code=...
//   - POST /v0/management/codearts-auth-callback with the callback URL
func (h *Handler) RequestCodeArtsToken(c *gin.Context) {
	ctx := PopulateAuthContext(context.Background(), c)

	state, errState := misc.GenerateRandomState()
	if errState != nil {
		log.Errorf("codearts: failed to generate state parameter: %v", errState)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state parameter"})
		return
	}

	verifier, _, errPKCE := codeartsauth.GeneratePKCEPair()
	if errPKCE != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PKCE codes"})
		return
	}
	keyPair, errKey := codeartsauth.GenerateDpopKeyPair()
	if errKey != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate DPoP key pair"})
		return
	}

	// The portal requires a loopback port; the callback itself is never served by
	// the proxy, users paste the URL back (or the loopback listener on the
	// operator's machine completes it).
	port := strings.TrimSpace(c.Query("port"))
	if port == "" {
		port = "10000"
	}
	ticketID, errTicket := misc.GenerateRandomState()
	if errTicket != nil || strings.TrimSpace(ticketID) == "" {
		ticketID = codeartsauth.StableHexID(state, 32)
	}

	authSvc := newCodeArtsOAuthService(h.cfg)
	authURL, errURL := authSvc.BuildAuthorizationURL(verifier, port, ticketID)
	if errURL != nil || authURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build authorization url"})
		return
	}

	RegisterOAuthSession(state, constant.CodeArts)

	authDir := ""
	if h != nil && h.cfg != nil {
		authDir = h.cfg.AuthDir
	}
	callbackPath := codeartsCallbackFilePath(authDir, state)
	login := &codeartsLoginContext{
		verifier:     verifier,
		keyPair:      keyPair,
		port:         port,
		ticketID:     ticketID,
		callbackPath: callbackPath,
	}
	storeCodeArtsLoginContext(state, login)
	go h.runCodeArtsLogin(ctx, state, login)

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"url":      authURL,
		"state":    state,
		"flow":     "callback",
		"callback": codeartsauth.CallbackURL(port),
	})
}

// codeartsLoginContext carries the per-session secrets the completion step needs.
type codeartsLoginContext struct {
	verifier     string
	keyPair      *codeartsauth.DpopKeyPair
	port         string
	ticketID     string
	callbackPath string
}

// runCodeArtsLogin waits for the authorization callback and completes the
// credential exchange.
func (h *Handler) runCodeArtsLogin(ctx context.Context, state string, login *codeartsLoginContext) {
	defer forgetCodeArtsLoginContext(state)
	code, errWait := waitCodeArtsCallback(ctx, state, login)
	if errWait != nil {
		if !errors.Is(errWait, errOAuthSessionNotPending) {
			SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errWait))
			fmt.Printf("CodeArts authentication failed: %v\n", errWait)
		}
		return
	}
	if errComplete := h.completeCodeArtsLogin(ctx, state, login, code); errComplete != nil {
		fmt.Printf("CodeArts authentication failed: %v\n", errComplete)
	}
}

// codeartsCallbackFilePath mirrors writeOAuthCallbackFile naming so this handler
// can wait on the file the shared callback endpoints publish.
func codeartsCallbackFilePath(authDir, state string) string {
	authDir = strings.TrimSpace(authDir)
	if authDir == "" {
		return ""
	}
	return filepath.Join(authDir, fmt.Sprintf(".oauth-%s-%s.oauth", constant.CodeArts, state))
}

// waitCodeArtsCallback blocks until the callback file appears or the pending
// session ends.
func waitCodeArtsCallback(ctx context.Context, state string, login *codeartsLoginContext) (string, error) {
	if login == nil || strings.TrimSpace(login.callbackPath) == "" {
		return "", errors.New("auth directory is not configured")
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(oauthSessionTTL)
	defer deadline.Stop()

	for {
		if !IsOAuthSessionPending(state, constant.CodeArts) {
			return "", errOAuthSessionNotPending
		}
		if ctx.Err() != nil {
			return "", errors.New("timeout waiting for OAuth callback")
		}

		data, errRead := os.ReadFile(login.callbackPath)
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

// completeCodeArtsLogin exchanges the authorization code and persists the
// credential record.
func (h *Handler) completeCodeArtsLogin(ctx context.Context, state string, login *codeartsLoginContext, code string) error {
	if errGuard := guardOAuthSessionPendingForSave(state, constant.CodeArts); errGuard != nil {
		return errGuard
	}
	if login == nil {
		return errors.New("missing login context")
	}

	authSvc := newCodeArtsOAuthService(h.cfg)
	token, errExchange := authSvc.ExchangeCode(ctx, code, login.verifier, login.keyPair, login.port)
	if errExchange != nil {
		log.Errorf("CodeArts token exchange failed: %v", errExchange)
		SetOAuthSessionError(state, oauthSessionErrorWithCause("Authentication failed", errExchange))
		return errExchange
	}
	if !IsOAuthSessionPending(state, constant.CodeArts) {
		return errOAuthSessionNotPending
	}

	userInfo, errUser := authSvc.GetUserInfo(ctx, token.AccessKey, token.SecretKey, token.SecurityToken)
	if errUser != nil {
		// Identity is informational; the credential triple is already usable.
		log.Warnf("CodeArts caller identity lookup failed: %v", errUser)
	}

	fileName := codeArtsAuthFileName(token, userInfo)
	metadata := codeArtsMetadata(token, userInfo, login)
	tokenStorage := &codeartsauth.TokenStorage{
		AccessKey:      token.AccessKey,
		SecretKey:      token.SecretKey,
		SecurityToken:  token.SecurityToken,
		RefreshToken:   token.RefreshToken,
		Expired:        token.ExpiresAt.UTC().Format(time.RFC3339),
		DpopPrivateKey: token.DpopKeyPair.PrivateKey,
		DpopPublicKey:  token.DpopKeyPair.PublicKey,
		CodeVerifier:   token.CodeVerifier,
	}
	if userInfo != nil {
		tokenStorage.UserID = userInfo.UserID
		tokenStorage.UserName = userInfo.UserName
		tokenStorage.DomainID = userInfo.DomainID
	}
	tokenStorage.SetMetadata(metadata)

	label := "CodeArts User"
	if userInfo != nil && strings.TrimSpace(userInfo.UserName) != "" {
		label = "CodeArts " + userInfo.UserName
	}

	record := &coreauth.Auth{
		ID:         fileName,
		Provider:   constant.CodeArts,
		FileName:   fileName,
		Label:      label,
		Storage:    tokenStorage,
		Metadata:   metadata,
		Attributes: codeArtsAttributes(token, userInfo),
	}

	if errGuard := guardOAuthSessionPendingForSave(state, constant.CodeArts); errGuard != nil {
		return errGuard
	}
	savedPath, errSave := h.saveTokenRecord(ctx, record)
	if errSave != nil {
		log.Errorf("failed to save CodeArts credentials: %v", errSave)
		SetOAuthSessionError(state, "Failed to save authentication tokens")
		return errSave
	}

	CompleteOAuthSession(state)
	fmt.Printf("CodeArts authentication successful! Token saved to %s\n", savedPath)
	return nil
}

// PostCodeArtsAuthCallback accepts a pasted callback URL (or a bare authorization
// code) and completes the flow.
func (h *Handler) PostCodeArtsAuthCallback(c *gin.Context) {
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
	login := loadCodeArtsLoginContext(state)
	if login == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "login context not found"})
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		parsed, errParse := parseCodeArtsCallback(req.RedirectURL)
		if errParse != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "failed to parse callback url"})
			return
		}
		code = parsed
	}

	ctx := PopulateAuthContext(context.Background(), c)
	if errComplete := h.completeCodeArtsLogin(ctx, state, login, code); errComplete != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": errComplete.Error()})
		return
	}
	forgetCodeArtsLoginContext(state)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// codeArtsMetadata builds the persisted metadata map for a login.
func codeArtsMetadata(token *codeartsauth.TokenData, userInfo *codeartsauth.UserInfo, login *codeartsLoginContext) map[string]any {
	metadata := map[string]any{
		"type":           constant.CodeArts,
		"auth_kind":      coreauth.AuthKindOAuth,
		"timestamp":      time.Now().UnixMilli(),
		"access_key":     token.AccessKey,
		"secret_key":     token.SecretKey,
		"security_token": token.SecurityToken,
		"base_url":       codeartsauth.InferHubBaseURL,
		"token_host":     codeartsauth.IAMSTSHost,
		"expired":        token.ExpiresAt.UTC().Format(time.RFC3339),
	}
	if token.RefreshToken != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if token.CodeVerifier != "" {
		metadata["code_verifier"] = token.CodeVerifier
	}
	if token.DpopKeyPair != nil {
		metadata["dpop_private_key"] = token.DpopKeyPair.PrivateKey
		metadata["dpop_public_key"] = token.DpopKeyPair.PublicKey
	}
	if login != nil && strings.TrimSpace(login.ticketID) != "" {
		metadata["ticket_id"] = strings.TrimSpace(login.ticketID)
	}
	if userInfo != nil {
		if strings.TrimSpace(userInfo.UserID) != "" {
			metadata["user_id"] = userInfo.UserID
		}
		if strings.TrimSpace(userInfo.UserName) != "" {
			metadata["user_name"] = userInfo.UserName
		}
		if strings.TrimSpace(userInfo.DomainID) != "" {
			metadata["domain_id"] = userInfo.DomainID
		}
	}
	// Keep the session id stable across restarts so the upstream correlation
	// header does not churn.
	seed := token.AccessKey
	if userInfo != nil && strings.TrimSpace(userInfo.UserID) != "" {
		seed = userInfo.UserID + "|" + token.AccessKey
	}
	metadata["session_id"] = codeartsauth.StableHexID(seed, 16)
	return metadata
}

// codeArtsAttributes builds the immutable attributes for a login record.
func codeArtsAttributes(token *codeartsauth.TokenData, userInfo *codeartsauth.UserInfo) map[string]string {
	attributes := map[string]string{
		coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
		"api_key":                  token.AccessKey,
		"secret_key":               token.SecretKey,
		"security_token":           token.SecurityToken,
		"base_url":                 codeartsauth.InferHubBaseURL,
		"source":                   "oauth:codearts",
	}
	if userInfo != nil && strings.TrimSpace(userInfo.UserName) != "" {
		attributes["user_name"] = userInfo.UserName
	}
	return attributes
}

// codeArtsAuthFileName derives a stable file name from the account identity so a
// repeated login for the same account overwrites the existing credential.
func codeArtsAuthFileName(token *codeartsauth.TokenData, userInfo *codeartsauth.UserInfo) string {
	seed := ""
	if userInfo != nil {
		seed = strings.TrimSpace(userInfo.UserID)
		if seed == "" {
			seed = strings.TrimSpace(userInfo.UserName)
		}
	}
	if seed == "" && token != nil {
		seed = token.AccessKey
	}
	digest := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("codearts-%s.json", hex.EncodeToString(digest[:16]))
}

// parseCodeArtsCallback extracts the authorization code from a pasted callback
// URL. Unlike the deep-link providers, CodeArts uses a real http loopback
// redirect, so a bare code is also accepted.
func parseCodeArtsCallback(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("empty callback url")
	}
	if !strings.Contains(trimmed, "://") && !strings.Contains(trimmed, "?") {
		return trimmed, nil
	}
	parsed, errParse := url.Parse(trimmed)
	if errParse != nil {
		return "", errors.New("invalid callback url")
	}
	query := parsed.Query()
	if code := strings.TrimSpace(query.Get("code")); code != "" {
		return code, nil
	}
	if fragment := strings.TrimSpace(parsed.Fragment); fragment != "" {
		if values, errFragment := url.ParseQuery(fragment); errFragment == nil {
			if code := strings.TrimSpace(values.Get("code")); code != "" {
				return code, nil
			}
		}
	}
	return "", errors.New("callback url has no code parameter")
}

// codeartsLoginRegistry keeps the per-session login contexts in memory.
var codeartsLoginRegistry = newCodeArtsLoginRegistry()

// codeartsLoginStore is a tiny concurrency-safe map for login contexts.
type codeartsLoginStore struct {
	mutex  sync.Mutex
	values map[string]*codeartsLoginContext
}

func newCodeArtsLoginRegistry() *codeartsLoginStore {
	return &codeartsLoginStore{values: make(map[string]*codeartsLoginContext)}
}

func storeCodeArtsLoginContext(state string, login *codeartsLoginContext) {
	key := strings.TrimSpace(state)
	if key == "" || login == nil {
		return
	}
	codeartsLoginRegistry.mutex.Lock()
	defer codeartsLoginRegistry.mutex.Unlock()
	codeartsLoginRegistry.values[key] = login
}

func loadCodeArtsLoginContext(state string) *codeartsLoginContext {
	key := strings.TrimSpace(state)
	if key == "" {
		return nil
	}
	codeartsLoginRegistry.mutex.Lock()
	defer codeartsLoginRegistry.mutex.Unlock()
	return codeartsLoginRegistry.values[key]
}

// forgetCodeArtsLoginContext drops a completed login context so the DPoP key
// pair and PKCE verifier are not retained in memory after the flow finishes.
func forgetCodeArtsLoginContext(state string) {
	key := strings.TrimSpace(state)
	if key == "" {
		return
	}
	codeartsLoginRegistry.mutex.Lock()
	defer codeartsLoginRegistry.mutex.Unlock()
	delete(codeartsLoginRegistry.values, key)
}
