// Package floatboat implements account authentication for the FloatBoat
// (aoe.chat "Agent OS") desktop client.
//
// The protocol is recovered from the FloatBoat 0.5.9 Electron client
// (resources/app.asar → .vite/build/index.js and floatboatWorker.js; see
// docs/providers/floatboat-integration.md for the evidence worksheet):
//
//   - Login is a browser + deep-link authorization-code flow. The client opens
//     https://floatboat.ai/desktop-sign-in?state=<uuid>&callback_scheme=<version>
//     and receives the authorization code back on the "aoe://" deep link.
//   - The backend (api-backend-url = https://floatboat.ai) exchanges the code at
//     POST /api/desktop/auth/exchange for an access/refresh token pair.
//   - GET /api/desktop/user/me (Bearer access token) returns the account profile.
//   - POST /api/desktop/newapi/key (Bearer access token) mints the api_key that
//     authorizes inference against the Anthropic-Messages gateway
//     https://newapi.aoe.chat/v1/messages.
//   - GET /api/v1/auth/refresh rotates the token pair; POST
//     /api/desktop/auth/sign-out revokes the session.
//
// The provider identity ("floatboat") is kept separate from the wire format
// (Anthropic Messages); CLIProxyAPI therefore reuses the existing Claude
// translators.
package floatboat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	// Provider is the canonical provider key.
	Provider = "floatboat"
	// Label is the human readable provider name.
	Label = "FloatBoat (aoe.chat)"

	// DefaultBackendURL is the account/product backend origin.
	DefaultBackendURL = "https://floatboat.ai"
	// DefaultInferenceBaseURL is the Anthropic-Messages inference gateway.
	DefaultInferenceBaseURL = "https://newapi.aoe.chat"
	// DefaultSignInPath is the desktop sign-in surface on the product web.
	DefaultSignInPath = "/desktop-sign-in"

	// DeepLinkScheme is the deep-link scheme the desktop client registers.
	DeepLinkScheme = "aoe"

	// UserAgent identifies proxy automation against the backend.
	UserAgent = "Floatboat/0.5.9 CLIProxyAPI"

	// DefaultPricingPath lists the gateway catalogue with an inference API key.
	// Served by the new-api gateway (base_url), not the product web.
	DefaultPricingPath = "/api/pricing"

	// DefaultBillingSubscriptionPath and DefaultBillingUsagePath are the
	// gateway's OpenAI-compatible billing reads. They report a measured
	// allowance/spend in USD, not model pricing metadata.
	DefaultBillingSubscriptionPath = "/v1/dashboard/billing/subscription"
	DefaultBillingUsagePath        = "/v1/dashboard/billing/usage"

	// defaultRequestTimeout bounds credential acquisition calls only.
	defaultRequestTimeout = 30 * time.Second
)

// chatEndpointTypeSet are the upstream endpoint families this provider can
// drive through its Anthropic Messages executor. Catalogue entries that
// advertise none of them (audio, speech, image-generation-only, typesafe) are
// not routable here and are excluded from the registered catalogue.
var chatEndpointTypeSet = map[string]struct{}{
	"openai":    {},
	"anthropic": {},
	"gemini":    {},
}

// Endpoints groups the backend paths used during login and refresh.
type Endpoints struct {
	Exchange   string
	UserMe     string
	NewAPIKey  string
	Refresh    string
	SignOut    string
	DevicePair string
	// Pricing lists the inference gateway catalogue. It is served by the
	// new-api gateway itself and is the only authoritative model source: the
	// product web has no model-list endpoint of its own.
	Pricing string
	// BillingSubscription and BillingUsage are the gateway's OpenAI-compatible
	// billing reads. They report the account's measured allowance and spend,
	// which is distinct from model pricing metadata.
	BillingSubscription string
	BillingUsage        string
}

// DefaultEndpoints returns the backend paths observed in the client. Every path
// here was exercised against the live backend; the refresh path in particular is
// under /api/desktop/auth/ and not the /api/v1/ namespace.
func DefaultEndpoints() Endpoints {
	return Endpoints{
		Exchange:            "/api/desktop/auth/exchange",
		UserMe:              "/api/desktop/user/me",
		NewAPIKey:           "/api/desktop/newapi/key",
		Refresh:             "/api/desktop/auth/refresh",
		SignOut:             "/api/desktop/auth/sign-out",
		DevicePair:          "/api/desktop/auth/device-pair",
		Pricing:             DefaultPricingPath,
		BillingSubscription: DefaultBillingSubscriptionPath,
		BillingUsage:        DefaultBillingUsagePath,
	}
}

// TokenData is the normalized exchange/refresh result.
type TokenData struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
}

// UserProfile is the account identity returned by /api/desktop/user/me.
type UserProfile struct {
	ID              string
	Email           string
	Name            string
	Image           string
	IsAdmin         bool
	Credits         int64
	HasSubscription bool
	MembershipLevel string
}

// Client performs FloatBoat backend credential operations.
type Client struct {
	httpClient *http.Client
	backendURL string
	signInURL  string
	endpoints  Endpoints
}

// NewClient builds a proxy-aware client for the backend.
func NewClient(cfg *config.Config) *Client {
	return NewClientWithProxyURL(cfg, "", "")
}

// NewClientWithProxyURL builds a client honoring a per-auth proxy override and
// optional backend URL override (useful for tests).
func NewClientWithProxyURL(cfg *config.Config, proxyURL, backendURL string) *Client {
	base := strings.TrimRight(strings.TrimSpace(backendURL), "/")
	if base == "" {
		base = DefaultBackendURL
	}
	httpClient := &http.Client{Timeout: defaultRequestTimeout}
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
		backendURL: base,
		signInURL:  base + DefaultSignInPath,
		endpoints:  DefaultEndpoints(),
	}
}

// BackendURL returns the resolved backend origin.
func (c *Client) BackendURL() string {
	if c == nil {
		return DefaultBackendURL
	}
	return c.backendURL
}

// SetEndpoints overrides endpoint paths (tests only).
func (c *Client) SetEndpoints(endpoints Endpoints) {
	if c == nil {
		return
	}
	if strings.TrimSpace(endpoints.Exchange) != "" {
		c.endpoints.Exchange = endpoints.Exchange
	}
	if strings.TrimSpace(endpoints.UserMe) != "" {
		c.endpoints.UserMe = endpoints.UserMe
	}
	if strings.TrimSpace(endpoints.NewAPIKey) != "" {
		c.endpoints.NewAPIKey = endpoints.NewAPIKey
	}
	if strings.TrimSpace(endpoints.Refresh) != "" {
		c.endpoints.Refresh = endpoints.Refresh
	}
	if strings.TrimSpace(endpoints.SignOut) != "" {
		c.endpoints.SignOut = endpoints.SignOut
	}
}

// BuildSignInURL constructs the login URL the user must open in a browser.
func (c *Client) BuildSignInURL(state string) string {
	state = strings.TrimSpace(state)
	parsed, errParse := url.Parse(c.signInURL)
	if errParse != nil {
		return c.signInURL
	}
	query := parsed.Query()
	if state != "" {
		query.Set("state", state)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// ExchangeCode trades the deep-link authorization code for a token pair.
func (c *Client) ExchangeCode(ctx context.Context, code, state, deviceID string) (*TokenData, error) {
	payload := map[string]any{
		"code":  strings.TrimSpace(code),
		"state": strings.TrimSpace(state),
	}
	if trimmed := strings.TrimSpace(deviceID); trimmed != "" {
		payload["deviceId"] = trimmed
	}
	body, status, errPost := c.postJSON(ctx, c.endpoints.Exchange, payload, "")
	if errPost != nil {
		return nil, errPost
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("floatboat: code exchange failed with status %d: %s", status, summarizeBody(body))
	}
	return parseTokenData(body, "")
}

// FetchUserProfile reads the account profile with the access token. When the
// profile endpoint is unavailable it falls back to the desktop session JWT
// claims, which already carry the account identity.
func (c *Client) FetchUserProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	body, status, errGet := c.getJSON(ctx, c.endpoints.UserMe, accessToken)
	if errGet != nil {
		if fromJWT := userProfileFromJWT(accessToken); fromJWT != nil {
			return fromJWT, nil
		}
		return nil, errGet
	}
	if status < 200 || status >= 300 {
		if fromJWT := userProfileFromJWT(accessToken); fromJWT != nil {
			return fromJWT, nil
		}
		return nil, fmt.Errorf("floatboat: user profile failed with status %d: %s", status, summarizeBody(body))
	}
	profile := parseUserProfile(body)
	if profile.ID == "" && profile.Email == "" {
		if fromJWT := userProfileFromJWT(accessToken); fromJWT != nil {
			if profile.Email == "" {
				profile.Email = fromJWT.Email
			}
			if profile.ID == "" {
				profile.ID = fromJWT.ID
			}
			if profile.Name == "" {
				profile.Name = fromJWT.Name
			}
			if profile.Image == "" {
				profile.Image = fromJWT.Image
			}
		}
	}
	return profile, nil
}

// userProfileFromJWT extracts the account identity from the desktop session JWT.
func userProfileFromJWT(accessToken string) *UserProfile {
	claims := decodeJWTPayload(accessToken)
	if claims == nil {
		return nil
	}
	profile := &UserProfile{
		ID:      readString(claims, "fid", "sub", "userId", "user_id"),
		Email:   readString(claims, "email"),
		Name:    readString(claims, "display_name", "name"),
		Image:   readString(claims, "avatar", "image"),
		IsAdmin: readBool(claims, "isAdmin", "is_admin"),
	}
	profile.HasSubscription = readBool(claims, "isPremium", "is_premium")
	if profile.ID == "" && profile.Email == "" {
		return nil
	}
	return profile
}

// FetchNewAPIKey mints the inference api_key with the access token.
func (c *Client) FetchNewAPIKey(ctx context.Context, accessToken string) (string, error) {
	body, status, errPost := c.postJSON(ctx, c.endpoints.NewAPIKey, map[string]any{}, accessToken)
	if errPost != nil {
		return "", errPost
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("floatboat: newapi key failed with status %d: %s", status, summarizeBody(body))
	}
	return parseAPIKey(body), nil
}

// CatalogueEntry is one routable model from the gateway catalogue.
type CatalogueEntry struct {
	ID            string
	DisplayName   string
	EndpointTypes []string
	ModelRatio    float64
	// EnabledGroups is the cross-account union of groups the gateway advertises
	// this model for. It is informational only; entitlement is decided by
	// Catalogue.Groups.
	EnabledGroups []string
}

// Catalogue is the entitlement-filtered model catalogue for one account.
type Catalogue struct {
	Entries []CatalogueEntry
	// Groups are the gateway groups the api_key belongs to, used to decide
	// entitlement. Empty means the gateway did not report any.
	Groups []string
}

// FetchCatalogue reads the inference gateway catalogue with the account's
// api_key and keeps only the models the account is entitled to and that this
// provider can actually drive.
//
// The gateway's `/api/pricing` is the authoritative catalogue: the product web
// exposes no model endpoint, and the base entitlement group available to this
// account is discoverable only at runtime (it varies per key, so it cannot be
// hard-coded). Entitlement alone is necessary but not sufficient — some
// entitled entries are still rejected by a per-model policy — so this list is
// the routable candidate set, not a guarantee.
func (c *Client) FetchCatalogue(ctx context.Context, baseURL, apiKey string) (*Catalogue, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("floatboat: inference api_key is required")
	}
	baseURL = strings.TrimRight(firstNonEmpty(baseURL, DefaultInferenceBaseURL), "/")
	endpoint := baseURL + firstNonEmpty(c.endpoints.Pricing, DefaultPricingPath)
	body, status, errGet := c.getJSONWithURL(ctx, endpoint, apiKey)
	if errGet != nil {
		return nil, errGet
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("floatboat: pricing failed with status %d: %s", status, summarizeBody(body))
	}
	return parseCatalogue(body)
}

// parseCatalogue filters the gateway catalogue down to the models this account
// can route, and reports the account's own entitlement groups.
//
// Entitlement is the INTERSECTION of the key's own groups (`auto_groups`) with
// each entry's `enable_groups`, because `enable_groups` is only the cross-account
// union of groups a model is published to. Trusting `enable_groups` alone would
// register models for groups this key is not in, which the gateway rejects at
// request time. Entries must also advertise a chat endpoint family this
// provider's executor can drive.
func parseCatalogue(body []byte) (*Catalogue, error) {
	payload, errEnvelope := unwrapEnvelope(decodeJSONObject(body))
	if errEnvelope != nil {
		return nil, errEnvelope
	}
	rawEntries, ok := payload["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("floatboat: pricing response missing data array")
	}
	groups := readStringSlice(payload, "auto_groups")
	owned := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		owned[strings.TrimSpace(group)] = struct{}{}
	}
	catalogue := &Catalogue{Groups: groups}
	seen := make(map[string]struct{}, len(rawEntries))
	for _, raw := range rawEntries {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id := readString(entry, "model_name")
		if id == "" {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		types := readStringSlice(entry, "supported_endpoint_types")
		if !hasChatEndpoint(types) {
			continue
		}
		enabledGroups := readStringSlice(entry, "enable_groups")
		if len(owned) > 0 && !intersectsGroup(enabledGroups, owned) {
			continue
		}
		seen[id] = struct{}{}
		catalogue.Entries = append(catalogue.Entries, CatalogueEntry{
			ID:            id,
			DisplayName:   readString(entry, "display_name", "name"),
			EndpointTypes: types,
			ModelRatio:    readFloat64(entry, "model_ratio"),
			EnabledGroups: enabledGroups,
		})
	}
	return catalogue, nil
}

func intersectsGroup(groups []string, owned map[string]struct{}) bool {
	for _, group := range groups {
		if _, ok := owned[strings.TrimSpace(group)]; ok {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// Billing is the gateway's measured allowance and spend for one account.
//
// These are real observations from the gateway's billing reads; they are not
// model pricing metadata and are not a scheduler cooldown. Unknown values stay
// zero with OK=false rather than being reported as unlimited or exhausted.
type Billing struct {
	// Currency is the denomination the gateway reports amounts in.
	Currency string
	// SoftLimitUSD and HardLimitUSD are the account's spending ceilings.
	SoftLimitUSD float64
	HardLimitUSD float64
	// TotalUsageUSD is the spend accumulated against the limit.
	TotalUsageUSD float64
	// HasPaymentMethod mirrors the gateway's own flag.
	HasPaymentMethod bool
	// SubscriptionOK and UsageOK report which reads actually succeeded. A
	// caller must not derive a remaining balance unless both are true.
	SubscriptionOK bool
	UsageOK        bool
}

// Remaining returns the remaining allowance when both billing reads succeeded.
func (b Billing) Remaining() (float64, bool) {
	if !b.SubscriptionOK || !b.UsageOK {
		return 0, false
	}
	return b.HardLimitUSD - b.TotalUsageUSD, true
}

// FetchBilling reads the gateway's measured allowance and usage. Each read is
// independent, so a partially available billing view is reported honestly as
// such instead of failing the whole call: the subscription read is required,
// the usage read is best-effort.
func (c *Client) FetchBilling(ctx context.Context, baseURL, apiKey string) (*Billing, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("floatboat: inference api_key is required")
	}
	baseURL = strings.TrimRight(firstNonEmpty(baseURL, DefaultInferenceBaseURL), "/")

	subBody, subStatus, errSub := c.getJSONWithURL(ctx, baseURL+firstNonEmpty(c.endpoints.BillingSubscription, DefaultBillingSubscriptionPath), apiKey)
	if errSub != nil {
		return nil, errSub
	}
	if subStatus < 200 || subStatus >= 300 {
		return nil, fmt.Errorf("floatboat: billing subscription failed with status %d: %s", subStatus, summarizeBody(subBody))
	}
	billing := parseBillingSubscription(subBody)
	billing.SubscriptionOK = true

	usageBody, usageStatus, errUsage := c.getJSONWithURL(ctx, baseURL+firstNonEmpty(c.endpoints.BillingUsage, DefaultBillingUsagePath), apiKey)
	if errUsage == nil && usageStatus >= 200 && usageStatus < 300 {
		billing.TotalUsageUSD = readFloat64(decodeJSONObject(usageBody), "total_usage")
		billing.UsageOK = true
	}
	return billing, nil
}

// parseBillingSubscription maps the gateway's subscription payload.
func parseBillingSubscription(body []byte) *Billing {
	payload := decodeJSONObject(body)
	billing := &Billing{
		Currency:         firstNonEmpty(readString(payload, "currency"), "USD"),
		SoftLimitUSD:     readFloat64(payload, "soft_limit_usd"),
		HardLimitUSD:     readFloat64(payload, "hard_limit_usd"),
		HasPaymentMethod: readBool(payload, "has_payment_method"),
	}
	return billing
}

// Refresh rotates the token pair using the stored refresh token.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*TokenData, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("floatboat: refresh token is required")
	}
	payload := map[string]any{"refresh_token": refreshToken}
	body, status, errPost := c.postJSON(ctx, c.endpoints.Refresh, payload, "")
	if errPost != nil {
		return nil, errPost
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("floatboat: refresh failed with status %d: %s", status, summarizeBody(body))
	}
	token, errParse := parseTokenData(body, refreshToken)
	if errParse != nil {
		return nil, errParse
	}
	// The refresh endpoint may not reissue the inference key; callers keep the
	// stored one when it returns empty.
	return token, nil
}

// SignOut best-effort revokes the backend session.
func (c *Client) SignOut(ctx context.Context, accessToken string) error {
	body, status, errPost := c.postJSON(ctx, c.endpoints.SignOut, map[string]any{}, accessToken)
	if errPost != nil {
		return errPost
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("floatboat: sign-out failed with status %d: %s", status, summarizeBody(body))
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path, token string) ([]byte, int, error) {
	return c.doJSON(ctx, http.MethodGet, path, nil, token)
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, token string) ([]byte, int, error) {
	return c.doJSON(ctx, http.MethodPost, path, payload, token)
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload any, token string) ([]byte, int, error) {
	if c == nil {
		return nil, 0, fmt.Errorf("floatboat: client is nil")
	}
	return c.doJSONURL(ctx, method, c.backendURL+path, payload, token)
}

// hasChatEndpoint reports whether the gateway advertises any endpoint family
// this provider's executor can drive.
func hasChatEndpoint(types []string) bool {
	for _, endpointType := range types {
		if _, ok := chatEndpointTypeSet[strings.ToLower(strings.TrimSpace(endpointType))]; ok {
			return true
		}
	}
	return false
}

// getJSONWithURL performs a GET against the inference gateway authenticated
// with the inference api_key (new-api Bearer convention).
func (c *Client) getJSONWithURL(ctx context.Context, endpoint, apiKey string) ([]byte, int, error) {
	return c.doJSONURL(ctx, http.MethodGet, endpoint, nil, apiKey)
}

func (c *Client) doJSONURL(ctx context.Context, method, endpoint string, payload any, token string) ([]byte, int, error) {
	if c == nil {
		return nil, 0, fmt.Errorf("floatboat: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var reader io.Reader
	if payload != nil {
		encoded, errMarshal := json.Marshal(payload)
		if errMarshal != nil {
			return nil, 0, fmt.Errorf("floatboat: encode request: %w", errMarshal)
		}
		reader = bytes.NewReader(encoded)
	}
	req, errReq := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if errReq != nil {
		return nil, 0, fmt.Errorf("floatboat: create request: %w", errReq)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if trimmed := strings.TrimSpace(token); trimmed != "" {
		req.Header.Set("Authorization", "Bearer "+trimmed)
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, 0, fmt.Errorf("floatboat: request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("floatboat: close response body: %v", errClose)
		}
	}()
	body, errRead := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if errRead != nil {
		return nil, resp.StatusCode, fmt.Errorf("floatboat: read response: %w", errRead)
	}
	return body, resp.StatusCode, nil
}

// unwrapEnvelope flattens the backend's {code,message,data:{...}} response
// envelope, which FloatBoat wraps around most authenticated endpoints. Flat
// responses pass through unchanged so tests and future shape changes keep
// working. A non-zero code is reported so callers can fail loudly instead of
// silently parsing an error payload.
func unwrapEnvelope(payload map[string]any) (map[string]any, error) {
	if payload == nil {
		return nil, nil
	}
	if rawCode, ok := payload["code"]; ok && !isZeroCode(rawCode) {
		message := readString(payload, "message")
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("floatboat: backend error (code %v): %s", rawCode, message)
	}
	if data, ok := payload["data"].(map[string]any); ok {
		return data, nil
	}
	return payload, nil
}

func isZeroCode(value any) bool {
	switch typed := value.(type) {
	case float64:
		return typed == 0
	case int:
		return typed == 0
	case json.Number:
		parsed, errParse := typed.Int64()
		return errParse == nil && parsed == 0
	case string:
		return strings.TrimSpace(typed) == "0"
	default:
		return false
	}
}

// decodeJWTPayload decodes the payload segment of a JWT without verifying the
// signature. The backend issues desktop_session tokens carrying the account
// identity and lifetime, which lets the credential survive when a profile
// endpoint is unavailable.
func decodeJWTPayload(token string) map[string]any {
	segments := strings.Split(strings.TrimSpace(token), ".")
	if len(segments) < 2 {
		return nil
	}
	decoded, errDecode := base64.RawURLEncoding.DecodeString(segments[1])
	if errDecode != nil {
		return nil
	}
	var payload map[string]any
	if errUnmarshal := json.Unmarshal(decoded, &payload); errUnmarshal != nil {
		return nil
	}
	return payload
}

// jwtLifetimeSeconds derives the token lifetime from exp-iat when the response
// omits expires_in.
func jwtLifetimeSeconds(token string) int {
	claims := decodeJWTPayload(token)
	if claims == nil {
		return 0
	}
	exp := readInt64(claims, "exp")
	if exp <= 0 {
		return 0
	}
	iat := readInt64(claims, "iat")
	if iat <= 0 {
		return int(exp - time.Now().Unix())
	}
	return int(exp - iat)
}

func parseTokenData(body []byte, previousRefreshToken string) (*TokenData, error) {
	payload, errEnvelope := unwrapEnvelope(decodeJSONObject(body))
	if errEnvelope != nil {
		return nil, errEnvelope
	}
	accessToken := readString(payload, "accessToken", "access_token")
	refreshToken := readString(payload, "refreshToken", "refresh_token")
	if refreshToken == "" {
		refreshToken = previousRefreshToken
	}
	expiresIn := readPositiveInt(payload, "expiresIn", "expires_in")
	refreshExpiresIn := readPositiveInt(payload, "refreshExpiresIn", "refresh_expires_in")
	if accessToken == "" {
		return nil, fmt.Errorf("floatboat: invalid token response: %s", summarizeBody(body))
	}
	// The desktop exchange returns a plain JWT pair without expires_in, so the
	// lifetime is derived from the token's own exp/iat claims.
	jwtExpiry := jwtExpiryTime(accessToken)
	if expiresIn == 0 {
		expiresIn = jwtLifetimeSeconds(accessToken)
	}
	token := &TokenData{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}
	if !jwtExpiry.IsZero() {
		token.ExpiresAt = jwtExpiry
	} else if expiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	if refreshExpiresIn > 0 {
		token.RefreshExpiresAt = time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)
	}
	return token, nil
}

// jwtExpiryTime returns the access token's absolute expiry from its exp claim.
func jwtExpiryTime(token string) time.Time {
	claims := decodeJWTPayload(token)
	if claims == nil {
		return time.Time{}
	}
	exp := readInt64(claims, "exp")
	if exp <= 0 {
		return time.Time{}
	}
	return time.Unix(exp, 0)
}

func parseAPIKey(body []byte) string {
	payload, errEnvelope := unwrapEnvelope(decodeJSONObject(body))
	if errEnvelope != nil || payload == nil {
		return ""
	}
	return readString(payload, "api_key", "apiKey", "key")
}

func parseUserProfile(body []byte) *UserProfile {
	payload, errEnvelope := unwrapEnvelope(decodeJSONObject(body))
	if errEnvelope != nil || payload == nil {
		return &UserProfile{}
	}
	user := payload
	if nested := decodeJSONObjectField(payload, "user"); nested != nil {
		user = nested
	}
	profile := &UserProfile{
		ID:              readString(user, "id", "userId", "user_id", "fid", "sub"),
		Email:           readString(user, "email"),
		Name:            readString(user, "name", "display_name", "displayName"),
		Image:           readString(user, "image", "avatar"),
		IsAdmin:         readBool(user, "isAdmin", "is_admin"),
		Credits:         readInt64(payload, "credits", "quota_remaining"),
		MembershipLevel: readString(payload, "membershipLevel", "membership_level"),
	}
	profile.HasSubscription = readBool(payload, "hasActiveSubscription", "has_active_subscription")
	return profile
}

func decodeJSONObject(body []byte) map[string]any {
	var payload map[string]any
	if errUnmarshal := json.Unmarshal(body, &payload); errUnmarshal != nil {
		return map[string]any{}
	}
	return payload
}

func decodeJSONObjectField(payload map[string]any, key string) map[string]any {
	if payload == nil {
		return nil
	}
	raw, ok := payload[key]
	if !ok {
		return nil
	}
	nested, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return nested
}

func readString(payload map[string]any, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := payload[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func readBool(payload map[string]any, keys ...string) bool {
	if payload == nil {
		return false
	}
	for _, key := range keys {
		if value, ok := payload[key].(bool); ok {
			return value
		}
	}
	return false
}

func readPositiveInt(payload map[string]any, keys ...string) int {
	if payload == nil {
		return 0
	}
	for _, key := range keys {
		switch value := payload[key].(type) {
		case float64:
			if value > 0 {
				return int(value)
			}
		case json.Number:
			if parsed, errParse := strconv.Atoi(value.String()); errParse == nil && parsed > 0 {
				return parsed
			}
		case string:
			if parsed, errParse := strconv.Atoi(strings.TrimSpace(value)); errParse == nil && parsed > 0 {
				return parsed
			}
		}
	}
	return 0
}

func readInt64(payload map[string]any, keys ...string) int64 {
	if payload == nil {
		return 0
	}
	for _, key := range keys {
		switch value := payload[key].(type) {
		case float64:
			return int64(value)
		case json.Number:
			if parsed, errParse := value.Int64(); errParse == nil {
				return parsed
			}
		case string:
			if parsed, errParse := strconv.ParseInt(strings.TrimSpace(value), 10, 64); errParse == nil {
				return parsed
			}
		}
	}
	return 0
}

func readFloat64(payload map[string]any, keys ...string) float64 {
	if payload == nil {
		return 0
	}
	for _, key := range keys {
		switch value := payload[key].(type) {
		case float64:
			return value
		case json.Number:
			if parsed, errParse := value.Float64(); errParse == nil {
				return parsed
			}
		case string:
			if parsed, errParse := strconv.ParseFloat(strings.TrimSpace(value), 64); errParse == nil {
				return parsed
			}
		}
	}
	return 0
}

func readStringSlice(payload map[string]any, keys ...string) []string {
	if payload == nil {
		return nil
	}
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		switch value := raw.(type) {
		case []any:
			out := make([]string, 0, len(value))
			for _, item := range value {
				if s, okString := item.(string); okString && strings.TrimSpace(s) != "" {
					out = append(out, strings.TrimSpace(s))
				}
			}
			return out
		case []string:
			return value
		case string:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return []string{trimmed}
			}
		}
	}
	return nil
}

func summarizeBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if len(trimmed) > 512 {
		return trimmed[:512] + "…"
	}
	return trimmed
}
