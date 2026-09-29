// Package cline implements the Cline (cline.bot) OAuth login flow.
//
// Cline's OpenAI-compatible API lives at https://api.cline.bot/api/v1 and is
// authenticated with a WorkOS access token. Unlike Codex/Claude (device or
// loopback PKCE), the Cline extension publishes an authorization-code flow whose
// callback carries the credential bundle: the desktop client is redirected to a
// loopback callback_url, and the resulting code is a base64-encoded JSON document
// holding accessToken/refreshToken/email. For CLIProxyAPI deployments (often
// remote) the user signs in at app.cline.bot, copies the full callback URL from
// the browser address bar, and pastes it back so the server can extract tokens.
//
// The upstream bearer token is the bare WorkOS access token; Cline rejects the
// "workos:" prefix that older integrations assumed. AuthorizationHeader therefore
// returns "Bearer <token>" without a prefix.
package cline

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	// APIHost is the Cline API origin.
	APIHost = "https://api.cline.bot"
	// BaseURL is the OpenAI-compatible base URL (without /chat/completions).
	BaseURL = APIHost + "/api/v1"
	// AuthorizePath is the WorkOS authorization endpoint.
	AuthorizePath = "/api/v1/auth/authorize"
	// TokenPath exchanges an authorization code for tokens.
	TokenPath = "/api/v1/auth/token"
	// RefreshPath rotates a refresh token for a fresh access token.
	RefreshPath = "/api/v1/auth/refresh"
	// ModelsPath is Cline's public model catalog endpoint.
	ModelsPath = "/api/v1/ai/cline/models"

	// ClientType is the OAuth client type sent by the extension flow.
	ClientType = "extension"
	// CallbackPath is the loopback callback path embedded in callback_url.
	CallbackPath = "/callback"
	// CallbackPort is the loopback port embedded in callback_url. The callback
	// itself is never served; users paste the URL back manually.
	CallbackPort = "18080"

	// UserAgent is the Cline client user agent prefix.
	UserAgent = "Cline"

	// WorkOSPrefix is the legacy access-token prefix. Cline's current API rejects
	// this prefix and expects the bare WorkOS access token; it is retained for
	// reference and is intentionally NOT applied to outbound Authorization headers.
	WorkOSPrefix = "workos:"
)

// TokenData is the normalized Cline credential bundle produced by a successful
// login or refresh.
type TokenData struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64 // Unix seconds; 0 when unknown.
	Email        string
	FirstName    string
	LastName     string
}

// Client performs Cline credential-acquisition requests.
type Client struct {
	httpClient *http.Client
	apiHost    string
}

// NewClient creates a proxy-aware Cline OAuth client.
func NewClient(cfg *config.Config) *Client {
	return NewClientWithProxyURL(cfg, "")
}

// NewClientWithProxyURL creates a client with an optional per-auth proxy override.
func NewClientWithProxyURL(cfg *config.Config, proxyURL string) *Client {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	var sdkCfg config.SDKConfig
	if cfg != nil {
		sdkCfg = cfg.SDKConfig
	}
	if strings.TrimSpace(proxyURL) != "" {
		sdkCfg.ProxyURL = strings.TrimSpace(proxyURL)
	}
	httpClient = util.SetProxy(&sdkCfg, httpClient)
	return &Client{httpClient: httpClient, apiHost: APIHost}
}

// LoginURL carries the generated authorization URL and its state.
type LoginURL struct {
	URL   string
	State string
}

// BuildLoginURL generates the Cline authorization URL. The callback URL uses a
// 127.0.0.1 loopback host because the extension flow requires a loopback
// callback_url.
func (c *Client) BuildLoginURL() (*LoginURL, error) {
	state := mustRandomHex(16)
	params := url.Values{}
	params.Set("client_type", ClientType)
	params.Set("callback_url", "http://127.0.0.1:"+CallbackPort+CallbackPath)
	params.Set("redirect_uri", "http://127.0.0.1:"+CallbackPort+CallbackPath)
	params.Set("state", state)
	return &LoginURL{
		URL:   strings.TrimRight(c.apiHost, "/") + AuthorizePath + "?" + params.Encode(),
		State: state,
	}, nil
}

// ParseCallback accepts either the full callback URL pasted by the user or a
// bare authorization code, and returns the normalized credential bundle.
func (c *Client) ParseCallback(callback string) (*TokenData, error) {
	trimmed := strings.TrimSpace(callback)
	if trimmed == "" {
		return nil, fmt.Errorf("cline: callback is required")
	}
	code := trimmed
	if parsed, errParse := url.Parse(trimmed); errParse == nil && parsed.Scheme != "" {
		if q := parsed.Query(); q.Get("code") != "" {
			code = q.Get("code")
		}
	}

	// The extension flow embeds the credential bundle as base64 JSON in the code.
	if data, ok := decodeEmbeddedTokens(code); ok {
		if strings.TrimSpace(data.AccessToken) != "" {
			return data, nil
		}
	}

	// Fall back to the authorization-code token exchange.
	return c.exchangeCode(context.Background(), code)
}

// decodeEmbeddedTokens base64-decodes the extension code and extracts the token
// bundle (accessToken/refreshToken/email/firstName/lastName/expiresAt).
func decodeEmbeddedTokens(code string) (*TokenData, bool) {
	decoded, errDecode := base64.StdEncoding.DecodeString(padBase64(code))
	if errDecode != nil {
		if unescaped, errUnescape := url.QueryUnescape(code); errUnescape == nil {
			decoded, errDecode = base64.StdEncoding.DecodeString(padBase64(unescaped))
		}
		if errDecode != nil {
			return nil, false
		}
	}
	raw := strings.TrimSpace(string(decoded))
	// Cline appends non-JSON trailer data after the JSON object; trim to the
	// last closing brace.
	if lastBrace := strings.LastIndex(raw, "}"); lastBrace >= 0 {
		raw = raw[:lastBrace+1]
	}
	if !strings.HasPrefix(raw, "{") {
		return nil, false
	}
	var payload struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresAt    string `json:"expiresAt"`
		Email        string `json:"email"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
	}
	if errUnmarshal := json.Unmarshal([]byte(raw), &payload); errUnmarshal != nil {
		return nil, false
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return nil, false
	}
	out := &TokenData{
		AccessToken:  strings.TrimSpace(payload.AccessToken),
		RefreshToken: strings.TrimSpace(payload.RefreshToken),
		Email:        strings.TrimSpace(payload.Email),
		FirstName:    strings.TrimSpace(payload.FirstName),
		LastName:     strings.TrimSpace(payload.LastName),
	}
	out.ExpiresAt = parseExpiresAt(payload.ExpiresAt)
	return out, true
}

// exchangeCode swaps an authorization code for tokens via the token endpoint.
func (c *Client) exchangeCode(ctx context.Context, code string) (*TokenData, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	body, _ := json.Marshal(map[string]any{
		"grant_type":   "authorization_code",
		"code":         code,
		"client_type":  ClientType,
		"redirect_uri": "http://127.0.0.1:" + CallbackPort + CallbackPath,
	})
	raw, errPost := c.postJSON(ctx, strings.TrimRight(c.apiHost, "/")+TokenPath, body)
	if errPost != nil {
		return nil, errPost
	}
	return parseTokenEnvelope(raw)
}

// ExchangeToken swaps a refresh token for a fresh access token. The upstream
// rotates the refresh token on success.
func (c *Client) ExchangeToken(ctx context.Context, refreshToken string) (*TokenData, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("cline: refresh token is required")
	}
	body, _ := json.Marshal(map[string]any{
		"refreshToken": refreshToken,
		"grantType":    "refresh_token",
		"clientType":   ClientType,
	})
	raw, errPost := c.postJSON(ctx, strings.TrimRight(c.apiHost, "/")+RefreshPath, body)
	if errPost != nil {
		return nil, errPost
	}
	token, errParse := parseTokenEnvelope(raw)
	if errParse != nil {
		return nil, errParse
	}
	if strings.TrimSpace(token.RefreshToken) == "" {
		token.RefreshToken = refreshToken
	}
	return token, nil
}

// parseTokenEnvelope accepts either {data:{...}} or a flat object and normalizes
// the camelCase Cline token fields.
func parseTokenEnvelope(raw []byte) (*TokenData, error) {
	var envelope struct {
		Data *struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			ExpiresAt    string `json:"expiresAt"`
			Email        string `json:"email"`
			UserInfo     *struct {
				Email string `json:"email"`
			} `json:"userInfo"`
		} `json:"data"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresAt    string `json:"expiresAt"`
		Email        string `json:"email"`
	}
	if errUnmarshal := json.Unmarshal(raw, &envelope); errUnmarshal != nil {
		return nil, fmt.Errorf("cline: parse token response: %w", errUnmarshal)
	}
	out := &TokenData{}
	expiresRaw := ""
	if envelope.Data != nil {
		out.AccessToken = strings.TrimSpace(envelope.Data.AccessToken)
		out.RefreshToken = strings.TrimSpace(envelope.Data.RefreshToken)
		out.Email = strings.TrimSpace(envelope.Data.Email)
		expiresRaw = envelope.Data.ExpiresAt
		if out.Email == "" && envelope.Data.UserInfo != nil {
			out.Email = strings.TrimSpace(envelope.Data.UserInfo.Email)
		}
	} else {
		out.AccessToken = strings.TrimSpace(envelope.AccessToken)
		out.RefreshToken = strings.TrimSpace(envelope.RefreshToken)
		out.Email = strings.TrimSpace(envelope.Email)
		expiresRaw = envelope.ExpiresAt
	}
	if out.AccessToken == "" {
		return nil, fmt.Errorf("cline: token response contained no access token")
	}
	out.ExpiresAt = parseExpiresAt(expiresRaw)
	return out, nil
}

// postJSON sends a JSON POST and returns the response body for 2xx responses.
func (c *Client) postJSON(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	req, errRequest := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if errRequest != nil {
		return nil, errRequest
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("cline: post %s: %w", endpoint, errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("cline: close response body: %v", errClose)
		}
	}()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, errRead
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cline: post %s HTTP %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

// AccountLabel returns a display label for the credential.
func (t *TokenData) AccountLabel() string {
	if t == nil {
		return ""
	}
	name := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(t.FirstName), strings.TrimSpace(t.LastName)}, " "))
	if name != "" {
		return name
	}
	return strings.TrimSpace(t.Email)
}

// AccessTokenWithPrefix normalizes a raw token. The current Cline API authenticates
// with the bare WorkOS access token and rejects the legacy "workos:" prefix, so
// this helper returns the token unchanged. It remains idempotent for callers that
// may pass an already-prefixed value. BYOK API keys (sk-...) are also passed
// through as-is.
func AccessTokenWithPrefix(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, WorkOSPrefix) {
		return strings.TrimPrefix(trimmed, WorkOSPrefix)
	}
	return trimmed
}

// AuthorizationHeader builds the full Authorization header value for a token.
func AuthorizationHeader(token string) string {
	value := AccessTokenWithPrefix(token)
	if value == "" {
		return ""
	}
	return "Bearer " + value
}

func parseExpiresAt(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if parsed, errParse := time.Parse(time.RFC3339, raw); errParse == nil {
		return parsed.Unix()
	}
	return 0
}

func padBase64(value string) string {
	trimmed := strings.TrimSpace(value)
	if pad := len(trimmed) % 4; pad != 0 {
		trimmed += strings.Repeat("=", 4-pad)
	}
	return trimmed
}

func mustRandomHex(bytesCount int) string {
	buf := make([]byte, bytesCount)
	if _, errRead := rand.Read(buf); errRead != nil {
		// crypto/rand failure is unrecoverable for login URL generation.
		panic(errRead)
	}
	return hex.EncodeToString(buf)
}
