package codearts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
)

// Upstream endpoints and constants, extracted from the CodeArts desktop
// install package (resources/product.json and resources/app.asar).
const (
	// Provider is the CLIProxyAPI provider identifier.
	Provider = "codearts"

	// ClientID is the public OAuth client id shipped with the desktop client.
	ClientID = "codearts"
	// Region is the Huawei Cloud region the desktop client targets.
	Region = "cn-north-4"
	// PluginName is the desktop plugin identity sent to the portal and snap API.
	PluginName = "snap_AIIDE"
	// PluginVersion is the desktop plugin version reported upstream.
	PluginVersion = "5.1.0"

	// PortalHost is the CodeArts portal that serves the authorize page.
	PortalHost = "https://codearts.huaweicloud.com/portal"
	// AuthorizePath is the browser authorization endpoint.
	AuthorizePath = "/authorize"
	// AuthRedirectPath is the loopback path the desktop client listens on.
	AuthRedirectPath = "/oauth/callback"

	// IAMSTSHost is the Huawei Cloud STS service that issues the temporary
	// AK/SK/security-token triple.
	IAMSTSHost = "https://sts.cn-north-4.myhuaweicloud.com"
	// TokenPath exchanges/refreshes the temporary credentials.
	TokenPath = "/v1/oauth2/tokens"
	// CallerIdentityPath resolves the logged-in account identity.
	CallerIdentityPath = "/v5/caller-identity"

	// SnapEngineURL is the CodeArts snap engine origin (model catalog, tickets,
	// and the LLM gateway).
	SnapEngineURL = "https://snap-access.cn-north-4.myhuaweicloud.com"
	// SnapEngineAPIHost serves the ticket fallback login.
	SnapEngineAPIHost = SnapEngineURL + "/snap-manager"
	// InferHubBaseURL is the OpenAI-compatible LLM gateway base URL.
	InferHubBaseURL = SnapEngineURL + "/api/v2/"
	// BuiltinModelsPath lists the platform model catalog.
	BuiltinModelsPath = "/v1/model/builtin"
	// VendorsPath lists third-party LLM vendors.
	VendorsPath = "/v1/llm/vendors"
	// TicketPath is the legacy ticket-polling login fallback.
	TicketPath = "/v1/login/ticket"

	// BenefitAPIURL is the developer gateway that serves the simple check-in.
	BenefitAPIURL = "https://opengw.developer.huaweicloud.com"
	// BenefitClaimPath queries/claims the simple daily benefit.
	BenefitClaimPath = "/api/v1/benefit/claim"
	// BalancePath reports the remaining token balance.
	BalancePath = "/api/v1/user/tokens/balance"
	// GatewayConfigPath lists the limited-time free models.
	GatewayConfigPath = "/api/v1/gateway/config"

	// WelfareDeliveryPath lists the welfare activities (including daily check-in).
	WelfareDeliveryPath = "/v1/ops/delivery"
	// WelfareClaimPath claims one welfare activity.
	WelfareClaimPath = "/v1/ops/claim"
	// WelfareConfirmPath confirms a claimed welfare activity.
	WelfareConfirmPath = "/v1/ops/confirm"

	// AgentTypePromptCenter is the Agent-Type value for model/policy calls.
	AgentTypePromptCenter = "PromptCenter"

	// DefaultHTTPTimeout bounds credential-acquisition requests only.
	DefaultHTTPTimeout = 30 * time.Second

	// RefreshWindowSeconds is how long before expiry credentials are refreshed.
	RefreshWindowSeconds = 300
)

// freeBenefitModelIDs are the limited-time free models (限时福利) served by the
// developer gateway. They are listed dynamically by /api/v1/gateway/config, and
// the fixed set below is used only to recognise them for request routing: these
// ids are rejected with InferHub.002002009.404 unless the request carries the
// maas_type=benefit header, so the id must be classifiable on the wire.
var freeBenefitModelIDs = map[string]struct{}{
	"deepseek-v4-flash-0731": {},
	"deepseek-v4-pro-0813":   {},
	"glm-5.3-flash":          {},
}

// IsFreeBenefitModel reports whether the model id is served from the
// limited-time free pool and therefore needs benefit routing.
func IsFreeBenefitModel(modelID string) bool {
	_, ok := freeBenefitModelIDs[strings.ToLower(strings.TrimSpace(modelID))]
	return ok
}

// TokenData is the temporary credential bundle produced by the OAuth exchange.
type TokenData struct {
	// AccessKey is the temporary Huawei Cloud access key id.
	AccessKey string
	// SecretKey is the temporary Huawei Cloud secret access key.
	SecretKey string
	// SecurityToken is the temporary Huawei Cloud security token.
	SecurityToken string
	// ExpiresAt is the credential expiry.
	ExpiresAt time.Time
	// RefreshToken rotates the credential bundle.
	RefreshToken string
	// DpopKeyPair is the P-256 key pair that signed the exchange. STS requires
	// the same key for refresh, so it must be persisted.
	DpopKeyPair *DpopKeyPair
	// CodeVerifier is the PKCE verifier used for the exchange. Huawei Cloud also
	// requires it on refresh.
	CodeVerifier string
}

// UserInfo is the normalized caller identity.
type UserInfo struct {
	UserID   string
	UserName string
	DomainID string
}

// Client performs CodeArts credential and catalog requests.
type Client struct {
	httpClient *http.Client
	stsHost    string
	portalHost string
	snapHost   string
	benefitURL string
}

// NewClient creates a proxy-aware CodeArts client.
func NewClient(cfg *config.Config) *Client {
	return NewClientWithProxyURL(cfg, "")
}

// NewClientWithProxyURL creates a client with an optional per-auth proxy override.
func NewClientWithProxyURL(cfg *config.Config, proxyURL string) *Client {
	httpClient := &http.Client{Timeout: DefaultHTTPTimeout}
	var sdkCfg config.SDKConfig
	if cfg != nil {
		sdkCfg = cfg.SDKConfig
	}
	if strings.TrimSpace(proxyURL) != "" {
		sdkCfg.ProxyURL = strings.TrimSpace(proxyURL)
	}
	httpClient = util.SetProxy(&sdkCfg, httpClient)
	return &Client{
		httpClient: httpClient,
		stsHost:    IAMSTSHost,
		portalHost: PortalHost,
		snapHost:   SnapEngineURL,
		benefitURL: BenefitAPIURL,
	}
}

// GeneratePKCEPair creates the PKCE verifier/challenge pair.
//
// The desktop client derives both with a plain hex verifier and a base64url
// SHA-256 challenge; Huawei Cloud accepts it as-is.
func GeneratePKCEPair() (verifier, challenge string, err error) {
	verifier = randomHex(64)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// BuildAuthorizationURL builds the browser login URL. ticketID is an opaque
// correlation value echoed back during the ticket fallback flow.
func (c *Client) BuildAuthorizationURL(verifier, port, ticketID string) (string, error) {
	challenge := challengeFromVerifier(verifier)
	base := strings.TrimRight(c.portalHost, "/")
	u, errParse := url.Parse(base + AuthorizePath)
	if errParse != nil {
		return "", fmt.Errorf("codearts: build authorize url: %w", errParse)
	}
	q := u.Query()
	q.Set("theme", "2")
	q.Set("locale", "zh-cn")
	q.Set("uri_scheme", ClientID)
	q.Set("client_id", ClientID)
	q.Set("port", strings.TrimSpace(port))
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "SHA-256")
	q.Set("ticket_id", strings.TrimSpace(ticketID))
	q.Set("plugin-name", PluginName)
	q.Set("plugin-version", PluginVersion)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// CallbackURL returns the loopback redirect URI for the given port.
func CallbackURL(port string) string {
	return "http://127.0.0.1:" + strings.TrimSpace(port) + AuthRedirectPath
}

// challengeFromVerifier derives the PKCE S256 challenge from a verifier.
func challengeFromVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// tokenResponse mirrors the Huawei Cloud STS /v1/oauth2/tokens payload.
type tokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	Credentials  struct {
		AccessKeyID     string `json:"access_key_id"`
		SecretAccessKey string `json:"secret_access_key"`
		SecurityToken   string `json:"security_token"`
		Expiration      string `json:"expiration"`
	} `json:"credentials"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	ErrorMsg  string `json:"error_msg"`
	ErrorCode string `json:"error_code"`
}

// ExchangeCode trades the authorization code for temporary credentials.
func (c *Client) ExchangeCode(ctx context.Context, code, verifier string, keyPair *DpopKeyPair, port string) (*TokenData, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("codearts: authorization code is empty")
	}
	if keyPair == nil {
		return nil, fmt.Errorf("codearts: DPoP key pair is required")
	}
	form := url.Values{}
	form.Set("client_id", ClientID)
	form.Set("code", strings.TrimSpace(code))
	form.Set("code_verifier", verifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", CallbackURL(port))

	return c.postToken(ctx, form, keyPair, verifier)
}

// Refresh rotates the temporary credentials using the stored refresh token.
//
// Refresh requires the same DPoP key and PKCE verifier as the original exchange
// because STS re-validates the proof of possession.
func (c *Client) Refresh(ctx context.Context, refreshToken, verifier string, keyPair *DpopKeyPair) (*TokenData, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("codearts: refresh token is empty")
	}
	if keyPair == nil {
		return nil, fmt.Errorf("codearts: DPoP key pair is required to refresh credentials")
	}
	form := url.Values{}
	form.Set("client_id", ClientID)
	form.Set("code_verifier", verifier)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", strings.TrimSpace(refreshToken))

	return c.postToken(ctx, form, keyPair, verifier)
}

// postToken performs the signed form POST shared by exchange and refresh.
func (c *Client) postToken(ctx context.Context, form url.Values, keyPair *DpopKeyPair, verifier string) (*TokenData, error) {
	endpoint := strings.TrimRight(c.stsHost, "/") + TokenPath
	body := form.Encode()

	dpopProof, errProof := SignDpopProof(keyPair, http.MethodPost, endpoint)
	if errProof != nil {
		return nil, errProof
	}

	req, errReq := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if errReq != nil {
		return nil, fmt.Errorf("codearts: create token request: %w", errReq)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("DPoP", dpopProof)

	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("codearts: token request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			return
		}
	}()

	data, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, fmt.Errorf("codearts: read token response: %w", errRead)
	}

	var parsed tokenResponse
	if errUnmarshal := json.Unmarshal(data, &parsed); errUnmarshal != nil {
		if resp.StatusCode >= 400 {
			return nil, &Error{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(data))}
		}
		return nil, fmt.Errorf("codearts: invalid token response (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{StatusCode: resp.StatusCode, Code: parsed.ErrorCode, Message: tokenErrorMessage(parsed, data)}
	}
	if strings.TrimSpace(parsed.Credentials.AccessKeyID) == "" || strings.TrimSpace(parsed.Credentials.SecretAccessKey) == "" {
		return nil, fmt.Errorf("codearts: token response has no credentials")
	}

	token := &TokenData{
		AccessKey:     strings.TrimSpace(parsed.Credentials.AccessKeyID),
		SecretKey:     strings.TrimSpace(parsed.Credentials.SecretAccessKey),
		SecurityToken: strings.TrimSpace(parsed.Credentials.SecurityToken),
		RefreshToken:  strings.TrimSpace(parsed.RefreshToken),
		DpopKeyPair:   keyPair,
		CodeVerifier:  verifier,
	}
	if expiry := strings.TrimSpace(parsed.Credentials.Expiration); expiry != "" {
		if parsedTime, errTime := time.Parse(time.RFC3339, expiry); errTime == nil {
			token.ExpiresAt = parsedTime.UTC()
		}
	}
	if token.ExpiresAt.IsZero() {
		token.ExpiresAt = time.Now().UTC().Add(time.Hour)
	}
	return token, nil
}

// tokenErrorMessage prefers the structured OAuth error, then the gateway error
// fields, then the raw body.
func tokenErrorMessage(parsed tokenResponse, raw []byte) string {
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return strings.TrimSpace(parsed.Error.Message)
	}
	if strings.TrimSpace(parsed.ErrorMsg) != "" {
		return strings.TrimSpace(parsed.ErrorMsg)
	}
	return strings.TrimSpace(string(raw))
}

// GetUserInfo resolves the account identity with signed AK/SK credentials.
func (c *Client) GetUserInfo(ctx context.Context, accessKey, secretKey, securityToken string) (*UserInfo, error) {
	endpoint := strings.TrimRight(c.stsHost, "/") + CallerIdentityPath
	headers := Sign(SignOptions{
		Method: http.MethodGet,
		URL:    endpoint,
		Headers: map[string]string{
			SecurityTokenHeader: securityToken,
			"Content-Type":      "application/json",
		},
	}, accessKey, secretKey)

	req, errReq := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if errReq != nil {
		return nil, fmt.Errorf("codearts: create caller-identity request: %w", errReq)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("codearts: caller-identity request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			return
		}
	}()
	data, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, fmt.Errorf("codearts: read caller-identity response: %w", errRead)
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}

	var payload struct {
		PrincipalURN  string `json:"principal_urn"`
		PrincipalID   string `json:"principal_id"`
		UserID        string `json:"user_id"`
		UserIDCamel   string `json:"userId"`
		UserName      string `json:"user_name"`
		UserNameCamel string `json:"userName"`
		AccountID     string `json:"account_id"`
		DomainID      string `json:"domain_id"`
		DomainIDCamel string `json:"domainId"`
	}
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil {
		return nil, fmt.Errorf("codearts: invalid caller-identity response: %w", errUnmarshal)
	}

	userName := strings.TrimSpace(payload.UserName)
	if userName == "" {
		userName = strings.TrimSpace(payload.UserNameCamel)
	}
	if userName == "" {
		userName = urnLeaf(payload.PrincipalURN)
	}
	userID := firstNonEmpty(payload.PrincipalID, payload.UserID, payload.UserIDCamel)
	domainID := firstNonEmpty(payload.AccountID, payload.DomainID, payload.DomainIDCamel)

	if userID == "" && userName == "" {
		return nil, fmt.Errorf("codearts: caller identity is empty")
	}
	return &UserInfo{UserID: userID, UserName: userName, DomainID: domainID}, nil
}

// TokenBalance is the credit balance of the signed-in account, returned by the
// developer gateway token balance endpoint.
type TokenBalance struct {
	// Channel is the benefit channel ("codearts").
	Channel string
	// TotalQuota is the total credit quota granted to the account.
	TotalQuota int64
	// TotalBalance is the remaining credit balance.
	TotalBalance int64
	// UsedAmount is the credits consumed so far.
	UsedAmount int64
	// DailyTokenLimit is the daily token cap (0 means unlimited).
	DailyTokenLimit int64
	// DailyTokensUsed is the tokens consumed today.
	DailyTokensUsed int64
	// MonthlyTokenLimit is the monthly token cap (0 means unlimited).
	MonthlyTokenLimit int64
	// MonthlyTokensUsed is the tokens consumed this month.
	MonthlyTokensUsed int64
	// ExpireTime is the quota expiry timestamp in milliseconds (0 means none).
	ExpireTime int64
}

// FetchTokenBalance queries the developer gateway for the account's credit
// balance. It signs the request with the temporary AK/SK triple the same way
// the inference endpoint does.
func (c *Client) FetchTokenBalance(ctx context.Context, accessKey, secretKey, securityToken string) (*TokenBalance, error) {
	endpoint := strings.TrimRight(c.benefitURL, "/") + "/api/v1/user/tokens/balance"
	req, errReq := c.signedRequest(ctx, http.MethodGet, endpoint, nil, Credentials{
		AccessKey:     accessKey,
		SecretKey:     secretKey,
		SecurityToken: securityToken,
	}, map[string]string{"X-Language": "zh-cn"})
	if errReq != nil {
		return nil, errReq
	}

	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("codearts: token balance request failed: %w", errDo)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	data, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, fmt.Errorf("codearts: read token balance response: %w", errRead)
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}

	var payload struct {
		ErrorCode string `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		Result    struct {
			Channel           string `json:"channel"`
			TotalQuota        int64  `json:"total_quota"`
			TotalBalance      int64  `json:"total_balance"`
			UsedAmount        int64  `json:"used_amount"`
			DailyTokenLimit   int64  `json:"daily_token_limit"`
			DailyTokensUsed   int64  `json:"daily_tokens_used"`
			MonthlyTokenLimit int64  `json:"monthly_token_limit"`
			MonthlyTokensUsed int64  `json:"monthly_tokens_used"`
			ExpireTime        int64  `json:"expire_time"`
		} `json:"result"`
	}
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil {
		return nil, fmt.Errorf("codearts: invalid token balance response: %w", errUnmarshal)
	}
	if payload.ErrorCode != "" && payload.ErrorCode != "0000" {
		return nil, &Error{StatusCode: resp.StatusCode, Code: payload.ErrorCode, Message: payload.ErrorMsg}
	}

	return &TokenBalance{
		Channel:           payload.Result.Channel,
		TotalQuota:        payload.Result.TotalQuota,
		TotalBalance:      payload.Result.TotalBalance,
		UsedAmount:        payload.Result.UsedAmount,
		DailyTokenLimit:   payload.Result.DailyTokenLimit,
		DailyTokensUsed:   payload.Result.DailyTokensUsed,
		MonthlyTokenLimit: payload.Result.MonthlyTokenLimit,
		MonthlyTokensUsed: payload.Result.MonthlyTokensUsed,
		ExpireTime:        payload.Result.ExpireTime,
	}, nil
}

// urnLeaf extracts the trailing account/user name from a principal URN.
func urnLeaf(urn string) string {
	trimmed := strings.TrimSpace(urn)
	if trimmed == "" {
		return ""
	}
	if index := strings.LastIndex(trimmed, ":"); index >= 0 {
		trimmed = trimmed[index+1:]
	}
	if index := strings.Index(trimmed, "/"); index >= 0 {
		trimmed = trimmed[index+1:]
	}
	return strings.TrimSpace(trimmed)
}

// firstNonEmpty returns the first trimmed non-empty value.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// Error is a structured upstream API error.
type Error struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Code is the upstream business error code, when available.
	Code string
	// Message is the upstream error message.
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return "codearts: unknown error"
	}
	if e.Code != "" {
		return fmt.Sprintf("codearts: HTTP %d code %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("codearts: HTTP %d: %s", e.StatusCode, e.Message)
}

// signedRequest builds a request whose headers carry a Huawei Cloud signature.
func (c *Client) signedRequest(ctx context.Context, method, endpoint string, body []byte, creds Credentials, extraHeaders map[string]string) (*http.Request, error) {
	headers := map[string]string{}
	for key, value := range extraHeaders {
		headers[key] = value
	}
	if strings.TrimSpace(creds.SecurityToken) != "" {
		headers[SecurityTokenHeader] = creds.SecurityToken
	}
	signed := Sign(SignOptions{Method: method, URL: endpoint, Headers: headers, Body: body}, creds.AccessKey, creds.SecretKey)

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, errReq := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if errReq != nil {
		return nil, fmt.Errorf("codearts: create request: %w", errReq)
	}
	for key, value := range signed {
		req.Header.Set(key, value)
	}
	return req, nil
}

// Credentials is the signed-request credential triple.
type Credentials struct {
	AccessKey     string
	SecretKey     string
	SecurityToken string
}
