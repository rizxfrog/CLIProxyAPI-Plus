// Package minimax implements OAuth 2.0 Device Authorization Grant (RFC 8628)
// authentication for the MiniMax Code (MCode) managed account backend.
//
// The protocol is recovered from the MiniMax Code client
// (packages/oauth-core/src/oauth-client.ts and endpoint-config.ts):
//
//   - Public client id "mcode-public", scope "agent.default", audience
//     "agent-backend", PKCE S256.
//   - Device authorization, token, and revocation endpoints live under the
//     regional account origin (account.minimax.io / account.minimax.cn).
//   - The token response carries access_token, refresh_token, token_type,
//     expires_in, and (from the JWT) sub / account_id.
package minimax

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	// ClientID is the public OAuth client id shipped with the MiniMax Code client.
	ClientID = "mcode-public"
	// Scope is the single OAuth scope requested by the client.
	Scope = "agent.default"
	// Audience is the OAuth audience identifying the agent backend.
	Audience = "agent-backend"

	// DeviceGrantType is the RFC 8628 device-code grant.
	DeviceGrantType = "urn:ietf:params:oauth:grant-type:device_code"

	// UserAgent identifies proxy automation against the account service.
	UserAgent = "MiniMaxAgent"

	// DefaultRegion is used when no region is configured.
	DefaultRegion = "en"

	defaultPollInterval = 5 * time.Second
	defaultExpiry       = 5 * time.Minute
)

// Region identifies which MiniMax account and agent backend to target.
type Region string

const (
	// RegionEN is the international (minimax.io) deployment.
	RegionEN Region = "en"
	// RegionCN is the mainland China (minimax.cn) deployment.
	RegionCN Region = "cn"
)

// NormalizeRegion maps user input to a supported region, defaulting to EN.
func NormalizeRegion(value string) Region {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "cn", "china", "minimax.cn", "minimaxi.com":
		return RegionCN
	default:
		return RegionEN
	}
}

// AccountOrigin returns the OAuth account service origin for the region.
func (r Region) AccountOrigin() string {
	if r == RegionCN {
		return "https://account.minimax.cn"
	}
	return "https://account.minimax.io"
}

// InferenceBaseURL returns the Anthropic-Messages base URL for the region. The
// messages endpoint is InferenceBaseURL + "/v1/messages".
func (r Region) InferenceBaseURL() string {
	if r == RegionCN {
		return "https://agent.minimax.cn/mavis/api/v1/llm"
	}
	return "https://agent.minimax.io/mavis/api/v1/llm"
}

// DeviceAuthorization is the device-code grant start response.
type DeviceAuthorization struct {
	DeviceCode      string
	CodeVerifier    string
	UserCode        string
	VerificationURI string
	// VerificationURIComplete is the upstream-provided URL that already embeds the
	// user_code. The native client opens this (with TUI attribution params) rather
	// than the bare VerificationURI.
	VerificationURIComplete string
	ExpiresIn               int
	Interval                int
}

// tuiDownloadSource preserves the attribution value the native MiniMax Code
// client sends (packages/tui/src/auth/authorization-url.ts).
const tuiDownloadSource = "mcode-internal"

// AuthorizationURL returns the browser URL the user should open to approve the
// device authorization. It mirrors the native client's markTuiAuthorizationUrl:
// prefer the complete URL that embeds user_code, then append the client_surface
// and download_source attribution parameters. Without these the upstream
// authorize page cannot pre-fill the code, which is why the bare
// verification_uri is not usable on its own.
func (d *DeviceAuthorization) AuthorizationURL() string {
	if d == nil {
		return ""
	}
	base := strings.TrimSpace(d.VerificationURIComplete)
	if base == "" {
		base = strings.TrimSpace(d.VerificationURI)
	}
	if base == "" {
		return ""
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	query := parsed.Query()
	if query.Get("user_code") == "" && strings.TrimSpace(d.UserCode) != "" {
		query.Set("user_code", strings.TrimSpace(d.UserCode))
	}
	query.Set("client_surface", "tui")
	query.Set("download_source", tuiDownloadSource)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// TokenData is the normalized token exchange/refresh result.
type TokenData struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	ExpiresAt    time.Time
	Subject      string
	AccountID    string
}

// Client performs MiniMax MCode device-flow credential operations.
type Client struct {
	httpClient *http.Client
	deviceURL  string
	tokenURL   string
	revokeURL  string
	region     Region
	// sleep is injectable so tests can avoid real polling delays.
	sleep func(context.Context, time.Duration) error
}

// NewClient builds a proxy-aware client for the given region.
func NewClient(cfg *config.Config, region Region) *Client {
	return NewClientWithProxyURL(cfg, region, "")
}

// NewClientWithProxyURL builds a client honoring a per-auth proxy override.
func NewClientWithProxyURL(cfg *config.Config, region Region, proxyURL string) *Client {
	region = NormalizeRegion(string(region))
	origin := region.AccountOrigin()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	var sdkCfg config.SDKConfig
	if cfg != nil {
		sdkCfg = cfg.SDKConfig
	}
	if trimmed := strings.TrimSpace(proxyURL); trimmed != "" {
		sdkCfg.ProxyURL = trimmed
	}
	httpClient = util.SetProxy(&sdkCfg, httpClient)
	return &Client{
		httpClient: httpClient,
		deviceURL:  origin + "/oauth2/device/code",
		tokenURL:   origin + "/oauth2/token",
		revokeURL:  origin + "/oauth2/revoke",
		region:     region,
		sleep:      sleepCtx,
	}
}

// Region returns the region this client targets.
func (c *Client) Region() Region {
	if c == nil {
		return RegionEN
	}
	return c.region
}

// StartDeviceAuthorization requests a device code and user verification URI.
func (c *Client) StartDeviceAuthorization(ctx context.Context) (*DeviceAuthorization, error) {
	if c == nil {
		return nil, fmt.Errorf("minimax: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	verifier, challenge, errPKCE := generatePKCE()
	if errPKCE != nil {
		return nil, fmt.Errorf("minimax: generate PKCE: %w", errPKCE)
	}
	form := url.Values{}
	form.Set("client_id", ClientID)
	form.Set("scope", Scope)
	form.Set("audience", Audience)
	form.Set("code_challenge", challenge)
	form.Set("code_challenge_method", "S256")

	body, status, errPost := c.postForm(ctx, c.deviceURL, form)
	if errPost != nil {
		return nil, errPost
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("minimax: device authorization failed with status %d: %s", status, strings.TrimSpace(string(body)))
	}
	var payload map[string]any
	if errDecode := json.Unmarshal(body, &payload); errDecode != nil {
		return nil, fmt.Errorf("minimax: decode device authorization: %w", errDecode)
	}
	deviceCode := readString(payload, "device_code")
	userCode := readString(payload, "user_code")
	verificationURI := readString(payload, "verification_uri")
	if verificationURI == "" {
		verificationURI = readString(payload, "verification_url")
	}
	verificationURIComplete := readString(payload, "verification_uri_complete")
	expiresIn := readPositiveInt(payload, "expires_in")
	interval := readPositiveInt(payload, "interval")
	if deviceCode == "" || userCode == "" || verificationURI == "" || expiresIn == 0 {
		return nil, fmt.Errorf("minimax: invalid device authorization response")
	}
	if interval == 0 {
		interval = int(defaultPollInterval / time.Second)
	}
	return &DeviceAuthorization{
		DeviceCode:              deviceCode,
		CodeVerifier:            verifier,
		UserCode:                userCode,
		VerificationURI:         verificationURI,
		VerificationURIComplete: verificationURIComplete,
		ExpiresIn:               expiresIn,
		Interval:                interval,
	}, nil
}

// PollDeviceToken polls the token endpoint until authorization completes.
func (c *Client) PollDeviceToken(ctx context.Context, auth *DeviceAuthorization) (*TokenData, error) {
	if c == nil {
		return nil, fmt.Errorf("minimax: client is nil")
	}
	if auth == nil || strings.TrimSpace(auth.DeviceCode) == "" {
		return nil, fmt.Errorf("minimax: device authorization is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	interval := auth.Interval
	if interval <= 0 {
		interval = int(defaultPollInterval / time.Second)
	}
	deadline := time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	if auth.ExpiresIn <= 0 {
		deadline = time.Now().Add(defaultExpiry)
	}
	for {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("minimax: device authorization cancelled: %w", ctx.Err())
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("minimax: device authorization expired")
		}
		form := url.Values{}
		form.Set("grant_type", DeviceGrantType)
		form.Set("device_code", auth.DeviceCode)
		form.Set("client_id", ClientID)
		form.Set("code_verifier", auth.CodeVerifier)
		body, status, errPost := c.postForm(ctx, c.tokenURL, form)
		if errPost != nil {
			return nil, errPost
		}
		var payload map[string]any
		if errDecode := json.Unmarshal(body, &payload); errDecode != nil {
			return nil, fmt.Errorf("minimax: decode token response: %w", errDecode)
		}
		switch errCode := readString(payload, "error"); errCode {
		case "":
		case "authorization_pending":
			if errSleep := c.sleepFor(ctx, time.Duration(interval)*time.Second); errSleep != nil {
				return nil, errSleep
			}
			continue
		case "slow_down":
			interval += 5
			if errSleep := c.sleepFor(ctx, time.Duration(interval)*time.Second); errSleep != nil {
				return nil, errSleep
			}
			continue
		default:
			return nil, fmt.Errorf("minimax: device authorization rejected (%s, status %d)", errCode, status)
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("minimax: token poll failed with status %d: %s", status, strings.TrimSpace(string(body)))
		}
		return parseTokenGrant(payload, "")
	}
}

// Refresh exchanges a refresh token for a new token pair.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*TokenData, error) {
	if c == nil {
		return nil, fmt.Errorf("minimax: client is nil")
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("minimax: refresh token is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", ClientID)
	form.Set("scope", Scope)
	form.Set("audience", Audience)
	body, status, errPost := c.postForm(ctx, c.tokenURL, form)
	if errPost != nil {
		return nil, errPost
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("minimax: refresh failed with status %d: %s", status, strings.TrimSpace(string(body)))
	}
	var payload map[string]any
	if errDecode := json.Unmarshal(body, &payload); errDecode != nil {
		return nil, fmt.Errorf("minimax: decode refresh response: %w", errDecode)
	}
	return parseTokenGrant(payload, refreshToken)
}

// Revoke revokes a refresh token. A network or 4xx failure is returned to the
// caller; callers treat revocation as best-effort during logout.
func (c *Client) Revoke(ctx context.Context, refreshToken string) error {
	if c == nil {
		return fmt.Errorf("minimax: client is nil")
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return fmt.Errorf("minimax: refresh token is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	form := url.Values{}
	form.Set("token", refreshToken)
	form.Set("token_type_hint", "refresh_token")
	form.Set("client_id", ClientID)
	body, status, errPost := c.postForm(ctx, c.revokeURL, form)
	if errPost != nil {
		return errPost
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("minimax: revoke failed with status %d: %s", status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) sleepFor(ctx context.Context, duration time.Duration) error {
	if c != nil && c.sleep != nil {
		return c.sleep(ctx, duration)
	}
	return sleepCtx(ctx, duration)
}

func (c *Client) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, int, error) {
	req, errReq := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if errReq != nil {
		return nil, 0, fmt.Errorf("minimax: create request: %w", errReq)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, 0, fmt.Errorf("minimax: request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("minimax: close response body: %v", errClose)
		}
	}()
	body, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, resp.StatusCode, fmt.Errorf("minimax: read response: %w", errRead)
	}
	return body, resp.StatusCode, nil
}

func parseTokenGrant(payload map[string]any, previousRefreshToken string) (*TokenData, error) {
	accessToken := readString(payload, "access_token")
	refreshToken := readString(payload, "refresh_token")
	if refreshToken == "" {
		refreshToken = previousRefreshToken
	}
	expiresIn := readPositiveInt(payload, "expires_in")
	if accessToken == "" || expiresIn == 0 {
		return nil, fmt.Errorf("minimax: invalid token response")
	}
	tokenType := readString(payload, "token_type")
	if tokenType == "" {
		tokenType = "Bearer"
	}
	token := &TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresIn:    expiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	if claims := decodeJWTPayload(accessToken); claims != nil {
		token.Subject = readString(claims, "sub")
		token.AccountID = readString(claims, "account_id")
	}
	return token, nil
}

func generatePKCE() (verifier, challenge string, err error) {
	raw := make([]byte, 32)
	if _, errRead := rand.Read(raw); errRead != nil {
		return "", "", errRead
	}
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func decodeJWTPayload(token string) map[string]any {
	segments := strings.Split(token, ".")
	if len(segments) != 3 || segments[1] == "" {
		return nil
	}
	decoded, errDecode := base64.RawURLEncoding.DecodeString(segments[1])
	if errDecode != nil {
		// Tolerate padded encodings too.
		decoded, errDecode = base64.URLEncoding.DecodeString(segments[1])
		if errDecode != nil {
			return nil
		}
	}
	var payload map[string]any
	if errUnmarshal := json.Unmarshal(decoded, &payload); errUnmarshal != nil {
		return nil
	}
	return payload
}

func readString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	if value, ok := payload[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func readPositiveInt(payload map[string]any, key string) int {
	if payload == nil {
		return 0
	}
	switch value := payload[key].(type) {
	case float64:
		if value > 0 {
			return int(value)
		}
	case json.Number:
		if parsed, err := value.Int64(); err == nil && parsed > 0 {
			return int(parsed)
		}
	}
	return 0
}

func sleepCtx(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("minimax: cancelled: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
