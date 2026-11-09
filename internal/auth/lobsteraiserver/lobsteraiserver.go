// Package lobsteraiserver implements account authentication and quota reads for
// the LobsterAI (NetEase/Youdao) desktop client's managed backend.
//
// The protocol is recovered from the LobsterAI 2026.10.13 Electron client
// (LobsterAI.exe NSIS installer → resources/app.asar → dist-electron/main;
// see docs/providers/lobsterai-server-integration.md for the evidence
// worksheet):
//
//   - Login is a browser authorization-code flow. The client opens
//     {base}/login?source=<source>&redirect_uri=<loopback>%3Freturn_to=…&state=<random>
//     and receives the authorization code back on the local callback URL
//     http://127.0.0.1:<port>/auth/callback?code=…&state=… (the client also
//     supports a lobsterai:// deep link; CLIProxyAPI uses the loopback path).
//   - The backend exchanges the code at POST /api/auth/exchange with body
//     {authCode, firstKeyfrom, latestKeyfrom, version, uuid?} and returns
//     {code:0, data:{accessToken, refreshToken, user, quota}}.
//   - POST /api/auth/refresh with {refreshToken, …keyfrom} rotates the pair;
//     a response that omits refreshToken keeps the stored one.
//   - Inference is forwarded to {base}/api/proxy/v1/* (OpenAI Chat Completions,
//     OpenAI Responses and Anthropic Messages); this package only reads the
//     per-model apiFormat and the account catalogue, the executor performs the
//     actual proxying.
//   - GET /api/models/available lists the account's models; GET /api/user/quota
//     reports credits/plan; GET /api/client-activities/slot and its context
//     report the daily check-in / one-time credit activity state.
//
// The provider identity ("lobsterai-server") is kept separate from each wire
// format, so CLIProxyAPI reuses the existing OpenAI / Codex / Claude
// translators instead of impersonating another provider.
package lobsteraiserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	// Provider is the canonical provider key.
	Provider = "lobsterai-server"
	// Label is the human readable provider name.
	Label = "LobsterAI (网易有道)"

	// DefaultBackendURL is the production backend origin.
	DefaultBackendURL = "https://lobsterai-server.youdao.com"
	// DefaultTestBackendURL is the internal/test backend origin.
	DefaultTestBackendURL = "https://lobsterai-server.inner.youdao.com"

	// DefaultLoginURL is the browser login surface. It lives on the product
	// portal (hash route), NOT on the backend origin: {base}/login is an
	// authenticated API that answers a stray browser request with HTTP 200 and
	// a JSON envelope (code 40100), so it must never be used as the sign-in page.
	DefaultLoginURL = "https://lobsterai.youdao.com/portal#/login"
	// DefaultLoginURLTest is the internal/test portal login surface.
	DefaultLoginURLTest = "https://lobsterai.inner.youdao.com/portal#/login"

	// DefaultLoginURLConfigProd and DefaultLoginURLConfigTest serve the
	// server-controlled login URL under data.value. The desktop client resolves
	// the login surface from here first and only falls back to the portal route,
	// so a config change does not require a client release.
	DefaultLoginURLConfigProd = "https://api-overmind.youdao.com/openapi/get/luna/hardware/lobsterai/prod/login-url"
	DefaultLoginURLConfigTest = "https://api-overmind.youdao.com/openapi/get/luna/hardware/lobsterai/test/login-url"

	// loginURLConfigTimeout bounds the login-surface config lookup. It is part of
	// credential acquisition, so a deadline is permitted here.
	loginURLConfigTimeout = 5 * time.Second

	// DefaultProxyBasePath prefixes the managed inference gateway. The executor
	// forwards /v1/* under it, so the effective base URL is base+"/api/proxy/v1".
	DefaultProxyBasePath = "/api/proxy/v1"

	// DefaultProxyBaseURL is the managed inference gateway's OpenAI-compatible
	// base URL. It is what the credential's base_url attribute carries.
	DefaultProxyBaseURL = DefaultBackendURL + DefaultProxyBasePath

	// UserAgent identifies proxy automation against the backend.
	UserAgent = "CLIProxyAPI/" + "lobsterai-server"

	// ClientVersion is sent as X-LobsterAI-Client-Version. The upstream treats a
	// missing/empty version as a legacy client, so a concrete value is required.
	ClientVersion = "2026.10.13"

	// ClientCapabilities is sent as X-LobsterAI-Client-Capabilities. The desktop
	// client always advertises these; the gateway gates some request shapes on it.
	ClientCapabilities = "kimi-k3-agentic-v1,thinking-level-control-v1"

	// HeaderAccountMode carries the account mode on every authenticated request
	// once the client knows which enterprise the account belongs to. A personal
	// account sends neither of the two enterprise headers.
	HeaderAccountMode = "X-LobsterAI-Account-Mode"
	// HeaderEnterpriseID carries the joined enterprise id, as a decimal string.
	HeaderEnterpriseID = "X-LobsterAI-Enterprise-Id"

	// AccountModePersonal and AccountModeEnterprise are the modes the backend
	// reports. Only the enterprise mode carries an enterprise id.
	AccountModePersonal   = "personal"
	AccountModeEnterprise = "enterprise"

	// Enterprise context business codes. The first three mean the account has no
	// usable enterprise context, so a persisted one must be dropped; any other
	// non-zero code is transient and must not erase a known context.
	EnterpriseErrorNotFound            = 41600
	EnterpriseErrorNotMember           = 41602
	EnterpriseErrorAccountModeMismatch = 41612

	// KeyfromOfficial is the default attribution channel for the public build.
	KeyfromOfficial = "official"

	// LoginSource identifies the proxy as the login surface's client. The desktop
	// client sends "electron"; the portal only performs the authorization-code
	// redirect for a recognized source, so an arbitrary value (which leaves the
	// user on the profile page with no code) must not be used.
	LoginSource = "electron"

	// DefaultLoopbackCallbackPort is the redirect target used when the caller has
	// no listener of its own (management/manual-paste flows). Nothing needs to
	// listen there: the portal appends code+state and the browser shows a
	// connection error, but the address bar carries the callback URL the user
	// pastes back.
	DefaultLoopbackCallbackPort = 18924

	// defaultRequestTimeout bounds credential acquisition calls only.
	// Inference requests establish their own context and carry no deadline.
	defaultRequestTimeout = 30 * time.Second
)

// Model is one entry from /api/models/available.
//
// The gateway implements only the OpenAI Chat Completions surface: /messages
// answers every model with code 40300 ("不支持的模型") and /responses is 404, so
// the raw apiFormat is carried for observability only.
type Model struct {
	ID          string
	DisplayName string
	APIFormat   string
	Thinking    bool
	Image       bool
	Video       bool
	// Accessible is the account's entitlement for this model. An inaccessible
	// model always answers 40301 ("模型不可见或无访问权限"), so only accessible
	// entries are routable.
	Accessible bool
	// SupportsToolCalling is the backend's own tool-calling flag (nullable).
	SupportsToolCalling bool
	// RuntimeProfile is the backend's named runtime profile, when any.
	RuntimeProfile string
	ContextLimit   int
	MaxTokens      int
	// ReasoningLevels is the model's declared reasoning-effort vocabulary, taken
	// from thinkingConfig.options[].openclawLevel.
	//
	// openclawLevel is the wire value of `reasoning_effort` (the product-side
	// `level` is a different, display-only name: the strongest option is "max" in
	// the UI but "xhigh" on the wire). Only the models that actually declare a
	// thinkingConfig have this list; the rest report supportsThinking without
	// exposing any choice, so an empty list means "no declared control" and not
	// "no thinking".
	ReasoningLevels []string
}

// Endpoints groups the backend paths used during login, refresh and reads.
type Endpoints struct {
	Exchange          string
	Refresh           string
	UserProfile       string
	EnterpriseContext string
	Models            string
	Quota             string
	CreditSummary     string
	ActivitySlot      string
	// ActivityContext is a prefix; the activity code and query are appended.
	ActivityContext string
}

// DefaultEndpoints returns the backend paths observed in the client.
func DefaultEndpoints() Endpoints {
	return Endpoints{
		Exchange:          "/api/auth/exchange",
		Refresh:           "/api/auth/refresh",
		UserProfile:       "/api/user/profile",
		EnterpriseContext: "/api/enterprise/context",
		Models:            "/api/models/available",
		Quota:             "/api/user/quota",
		CreditSummary:     "/api/user/profile-summary",
		ActivitySlot:      "/api/client-activities/slot",
		ActivityContext:   "/api/client-activities",
	}
}

// TokenData is the normalized exchange/refresh result.
type TokenData struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	ExpiresAt    time.Time
}

// User is the account identity returned by the exchange/profile endpoints.
type User struct {
	ID       string
	Email    string
	Name     string
	Avatar   string
	Nickname string
}

// EnterpriseContext is the account's joined-enterprise identity.
//
// The desktop client sends it as two request headers on every authenticated
// call, and the backend uses the pair to scope the account to an enterprise
// pool. A personal account has no context and sends neither header.
type EnterpriseContext struct {
	// AccountMode is "enterprise" for a context that was accepted.
	AccountMode string
	// EnterpriseID is the joined enterprise id, as a decimal string.
	EnterpriseID string
	// EnterpriseName and Role are informational (management display).
	EnterpriseName string
	Role           string
	// MemberID is the account's id inside the enterprise.
	MemberID string
}

// Quota is the normalized credential quota snapshot.
//
// These are the values the backend reports at login/cache-warmup time. They are
// a plan/credit snapshot, not a scheduler cooldown, and may be absent: an
// unknown value stays zero with Known=false rather than being reported as
// unlimited or exhausted.
type Quota struct {
	// SubscriptionStatus is the raw plan/status string (e.g. "free", "active").
	SubscriptionStatus string
	// Credits is the remaining credit balance when the backend reports one.
	Credits int64
	// CreditsKnown reports whether Credits was present in the payload.
	CreditsKnown bool
	// PurchasedCredits is a second credit bucket some plans carry separately.
	PurchasedCredits int64
	// Raw is the untrimmed data object, kept for diagnostics.
	Raw map[string]any
}

// CreditPackage is one credit bucket from GET /api/user/profile-summary.
//
// The backend groups credits by grant source (subscription, boost, free,
// bonus, invitation, campaign) and reports the remaining amount per group.
type CreditPackage struct {
	// Type is the machine group code (e.g. "subscription", "bonus").
	Type string
	// Label and LabelEn are the localized display names the client shows.
	Label   string
	LabelEn string
	// CreditsRemaining is the untouched remaining amount in this bucket.
	CreditsRemaining float64
	// ExpiresAt is the backend's own expiry string (ISO 8601, empty when the
	// bucket does not expire).
	ExpiresAt string
}

// CreditSummary is the account's credit total with its per-bucket breakdown.
//
// This is the source the desktop client's account menu reads
// (`profileSummary.totalCreditsRemaining` and `profileSummary.creditItems`).
// GET /api/user/quota is a different view: it carries only the free-tier bucket
// (freeCreditsTotal/freeCreditsRemaining) and drives the low-credit purchase
// gate, so it understates an account whose credits also include activity and
// daily-login grants.
type CreditSummary struct {
	// TotalCreditsRemaining is the backend's own account total.
	TotalCreditsRemaining float64
	// TotalKnown reports whether the total was present; false means unknown,
	// never zero.
	TotalKnown bool
	// Packages is the per-bucket breakdown, in backend order.
	Packages []CreditPackage
	// UserID and Nickname are the account identity the summary carries.
	UserID   string
	Nickname string
	// Raw is the untrimmed data object, kept for diagnostics.
	Raw map[string]any
}

// ActivitySnapshot is the read-only state of one backend activity (daily
// check-in / one-time credit). CLIProxyAPI reports this but never executes an
// activity action on the user's behalf.
type ActivitySnapshot struct {
	// SlotState is "available" or "empty".
	SlotState string
	// ActivityCode identifies the activity (empty when none is available).
	ActivityCode string
	// ConfigRevision is the revision required to read the context.
	ConfigRevision int64
	// Placement is the requested slot placement.
	Placement string
	// Context is the raw activity context payload when it could be read.
	Context map[string]any
}

// activityState returns the context's "state" object, if any.
func (s *ActivitySnapshot) activityState() map[string]any {
	if s == nil || s.Context == nil {
		return nil
	}
	return decodeJSONObjectField(s.Context, "state")
}

// Available reports whether the slot exposes a claimable activity.
func (s *ActivitySnapshot) Available() bool {
	return s != nil && s.SlotState == "available" && s.ActivityCode != "" && s.ConfigRevision > 0
}

// ClaimedToday reports the backend's own "already claimed today" flag.
func (s *ActivitySnapshot) ClaimedToday() bool {
	return readBool(s.activityState(), "claimedToday")
}

// Completed reports whether the whole activity period is finished, which makes
// today's reward moot.
func (s *ActivitySnapshot) Completed() bool {
	return readBool(s.activityState(), "completed")
}

// ClaimedDays is the number of days already claimed in this activity.
func (s *ActivitySnapshot) ClaimedDays() int64 {
	return readInt64(s.activityState(), "claimedDays")
}

// ClaimedCredits is the total granted by this activity so far.
func (s *ActivitySnapshot) ClaimedCredits() float64 {
	value, _ := firstNumeric(s.activityState(), "claimedCredits")
	return value
}

// RewardCredits is the reward one claim is worth, when the activity advertises
// it.
func (s *ActivitySnapshot) RewardCredits() float64 {
	value, _ := firstNumeric(s.activityState(), "rewardCredits")
	return value
}

// Actions lists the action ids the activity currently accepts.
func (s *ActivitySnapshot) Actions() []string {
	if s == nil || s.Context == nil {
		return nil
	}
	raw, ok := s.Context["actions"].([]any)
	if !ok {
		return nil
	}
	actions := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, okText := item.(string); okText && strings.TrimSpace(text) != "" {
			actions = append(actions, strings.TrimSpace(text))
		}
	}
	return actions
}

// ActivityActionCheckIn is the daily check-in action id.
const ActivityActionCheckIn = "check_in"

// ActivityAlreadyClaimedCode is the backend business code for a duplicate
// claim. The client reports it as success.
const ActivityAlreadyClaimedCode = 51104

// ErrActivityAlreadyClaimed marks a claim the backend accounted for already.
var ErrActivityAlreadyClaimed = errors.New("lobsterai-server: activity already claimed")

// ActivityClaimRequest describes one activity action execution.
type ActivityClaimRequest struct {
	// ActivityCode is the activity the action belongs to.
	ActivityCode string
	// ActionID is the action to run (see ActivityActionCheckIn).
	ActionID string
	// ConfigRevision is the revision read from the activity slot.
	ConfigRevision int64
	// IdempotencyKey deduplicates a retried claim. The backend accepts
	// ^[A-Za-z0-9._:-]{1,64}$; an empty value lets the client generate one.
	IdempotencyKey string
}

// ActivityClaim is the normalized result of an activity action.
type ActivityClaim struct {
	// Granted is the credit amount the backend reports for this claim.
	Granted float64
	// GrantedKnown reports whether Granted was present.
	GrantedKnown bool
	// PeriodKey identifies the accounted period (e.g. "2026-10-15").
	PeriodKey string
	// Message is the backend's own human-readable message, when any.
	Message string
	// Raw is the untrimmed action result payload.
	Raw map[string]any
}

// Client performs LobsterAI backend credential operations.
type Client struct {
	httpClient *http.Client
	backendURL string
	endpoints  Endpoints
	// Keyfrom identity appended to login-related requests. These are
	// attribution fields, not credentials.
	firstKeyfrom  string
	latestKeyfrom string
	version       string
	uuid          string
	// Enterprise context sent as two headers on authenticated requests. Empty
	// for a personal account, which is the default.
	accountMode  string
	enterpriseID string
	// activityPlatformOverride overrides the environment token sent to the
	// activity surface (see SetActivityPlatform).
	activityPlatformOverride string
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
		httpClient:    httpClient,
		backendURL:    base,
		endpoints:     DefaultEndpoints(),
		firstKeyfrom:  KeyfromOfficial,
		latestKeyfrom: KeyfromOfficial,
		version:       ClientVersion,
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
	if strings.TrimSpace(endpoints.Refresh) != "" {
		c.endpoints.Refresh = endpoints.Refresh
	}
	if strings.TrimSpace(endpoints.UserProfile) != "" {
		c.endpoints.UserProfile = endpoints.UserProfile
	}
	if strings.TrimSpace(endpoints.EnterpriseContext) != "" {
		c.endpoints.EnterpriseContext = endpoints.EnterpriseContext
	}
	if strings.TrimSpace(endpoints.Models) != "" {
		c.endpoints.Models = endpoints.Models
	}
	if strings.TrimSpace(endpoints.Quota) != "" {
		c.endpoints.Quota = endpoints.Quota
	}
	if strings.TrimSpace(endpoints.CreditSummary) != "" {
		c.endpoints.CreditSummary = endpoints.CreditSummary
	}
	if strings.TrimSpace(endpoints.ActivitySlot) != "" {
		c.endpoints.ActivitySlot = endpoints.ActivitySlot
	}
	if strings.TrimSpace(endpoints.ActivityContext) != "" {
		c.endpoints.ActivityContext = endpoints.ActivityContext
	}
}

// SetKeyfromIdentity overrides the attribution identity appended to login
// requests. Empty values fall back to the defaults (tests and multi-install
// callers).
func (c *Client) SetKeyfromIdentity(first, latest, version, uuid string) {
	if c == nil {
		return
	}
	if trimmed := strings.TrimSpace(first); trimmed != "" {
		c.firstKeyfrom = trimmed
	}
	if trimmed := strings.TrimSpace(latest); trimmed != "" {
		c.latestKeyfrom = trimmed
	}
	if trimmed := strings.TrimSpace(version); trimmed != "" {
		c.version = trimmed
	}
	c.uuid = strings.TrimSpace(uuid)
}

// SetEnterpriseContext sets the enterprise identity sent as request headers.
//
// A personal account must call this with empty values (or not at all): sending a
// fabricated enterprise id would scope the request to a pool the account is not
// a member of. The id must therefore come from the backend's own context read.
func (c *Client) SetEnterpriseContext(accountMode, enterpriseID string) {
	if c == nil {
		return
	}
	mode := strings.TrimSpace(accountMode)
	id := strings.TrimSpace(enterpriseID)
	if !strings.EqualFold(mode, AccountModeEnterprise) || id == "" {
		c.accountMode = ""
		c.enterpriseID = ""
		return
	}
	c.accountMode = AccountModeEnterprise
	c.enterpriseID = id
}

// SetActivityPlatform overrides the environment token sent to the activity
// surface.
//
// The desktop client sends its own process.platform (win32 / darwin / linux),
// and the backend treats any other value as an unrecognized environment and
// answers an empty slot. A proxy host is not always one of those three, and an
// operator may also know which environment the account was set up for, so an
// explicit value wins over the host mapping. An empty value restores the host
// default.
func (c *Client) SetActivityPlatform(platform string) {
	if c == nil {
		return
	}
	c.activityPlatformOverride = strings.TrimSpace(platform)
}

// keyfromPayload builds the attribution fields the upstream expects on login and
// refresh requests.
func (c *Client) keyfromPayload() map[string]any {
	payload := map[string]any{
		"firstKeyfrom":  c.firstKeyfrom,
		"latestKeyfrom": c.latestKeyfrom,
		"version":       c.version,
	}
	if trimmed := strings.TrimSpace(c.uuid); trimmed != "" {
		payload["uuid"] = trimmed
	}
	return payload
}

// appendKeyfromQuery appends the non-empty keyfrom payload as query parameters.
func (c *Client) appendKeyfromQuery(rawURL string) string {
	parsed, errParse := url.Parse(rawURL)
	if errParse != nil {
		return rawURL
	}
	query := parsed.Query()
	for key, value := range c.keyfromPayload() {
		if text := strings.TrimSpace(fmt.Sprintf("%v", value)); text != "" && text != "<nil>" {
			query.Set(key, text)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// LoginURL returns the browser login surface.
//
// The desktop client resolves it from a server-controlled config endpoint (so a
// changed login surface needs no client release) and falls back to the portal
// hash route. The backend origin's own /login path is deliberately NOT used: it
// is an authenticated API that answers a stray browser request with a JSON
// envelope instead of a page.
func (c *Client) LoginURL() string {
	if c == nil {
		return DefaultLoginURL
	}
	configEndpoint := DefaultLoginURLConfigProd
	fallback := DefaultLoginURL
	if c.isTestBackend() {
		configEndpoint = DefaultLoginURLConfigTest
		fallback = DefaultLoginURLTest
	}
	if resolved := c.fetchLoginURLFromConfig(configEndpoint); resolved != "" {
		return resolved
	}
	return fallback
}

// isTestBackend reports whether the client targets the internal/test backend.
func (c *Client) isTestBackend() bool {
	base := strings.ToLower(strings.TrimRight(strings.TrimSpace(c.backendURL), "/"))
	return strings.Contains(base, ".inner.") || strings.Contains(base, "test")
}

// fetchLoginURLFromConfig reads the server-controlled login URL. A failure is
// non-fatal: the caller falls back to the portal route.
func (c *Client) fetchLoginURLFromConfig(endpoint string) string {
	if c == nil || strings.TrimSpace(endpoint) == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), loginURLConfigTimeout)
	defer cancel()
	body, status, errGet := c.getJSON(ctx, endpoint, "")
	if errGet != nil || status < 200 || status >= 300 {
		return ""
	}
	payload := decodeJSONObject(body)
	value := readString(decodeJSONObjectField(payload, "data"), "value")
	if value == "" {
		return ""
	}
	return value
}

// BuildLoginURL constructs the browser login URL the user must open, carrying
// the CSRF state and the redirect URI.
//
// Every dependency is required for the portal to perform the authorization-code
// redirect, mirroring the desktop client: the login surface resolves from the
// config endpoint, source identifies the client, and redirect_uri gives the
// portal somewhere to return the code. A login that omits the code (the user
// simply landing on the portal profile page) means one of these was missing.
func (c *Client) BuildLoginURL(state, redirectURI string) string {
	base := DefaultLoginURL
	if c != nil {
		base = c.LoginURL()
	}
	redirect := strings.TrimSpace(redirectURI)
	if redirect == "" {
		redirect = DefaultRedirectURI()
	}
	return appendLoginParams(base, map[string]string{
		"state":        strings.TrimSpace(state),
		"redirect_uri": redirect,
		"source":       LoginSource,
	})
}

// DefaultRedirectURI is the loopback callback the portal returns the
// authorization code to when the caller has no listener of its own. It matches
// the path the desktop client registers.
func DefaultRedirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", DefaultLoopbackCallbackPort, CallbackPath)
}

// appendLoginParams merges parameters into a URL, placing them in the hash query
// when the surface is a hash route (#/login?…) and in the regular query
// otherwise, mirroring the desktop client's appendLoginParams.
//
// The hash case is assembled as a string rather than through url.URL.String():
// that method escapes the fragment, which would double-encode the percent
// escapes produced by url.Values.Encode().
func appendLoginParams(rawURL string, params map[string]string) string {
	parsed, errParse := url.Parse(strings.TrimSpace(rawURL))
	if errParse != nil {
		return rawURL
	}
	values := url.Values{}
	if parsed.Fragment != "" {
		hashPath := parsed.Fragment
		if idx := strings.Index(parsed.Fragment, "?"); idx >= 0 {
			hashPath = parsed.Fragment[:idx]
			if existing, errExisting := url.ParseQuery(parsed.Fragment[idx+1:]); errExisting == nil {
				values = existing
			}
		}
		for key, value := range params {
			if strings.TrimSpace(value) != "" {
				values.Set(key, value)
			}
		}
		base := parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
		if parsed.RawQuery != "" {
			base += "?" + parsed.RawQuery
		}
		if encoded := values.Encode(); encoded != "" {
			return base + "#" + hashPath + "?" + encoded
		}
		return base + "#" + hashPath
	}
	values = parsed.Query()
	for key, value := range params {
		if strings.TrimSpace(value) != "" {
			values.Set(key, value)
		}
	}
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

// ExchangeCode trades the authorization code for a token pair.
func (c *Client) ExchangeCode(ctx context.Context, code, state string) (*TokenData, *User, *Quota, error) {
	payload := map[string]any{"authCode": strings.TrimSpace(code)}
	if trimmed := strings.TrimSpace(state); trimmed != "" {
		payload["state"] = trimmed
	}
	for key, value := range c.keyfromPayload() {
		payload[key] = value
	}
	body, status, errPost := c.postJSON(ctx, c.endpoints.Exchange, payload, "")
	if errPost != nil {
		return nil, nil, nil, errPost
	}
	if status < 200 || status >= 300 {
		return nil, nil, nil, fmt.Errorf("lobsterai-server: code exchange failed with status %d: %s", status, summarizeBody(body))
	}
	data, errEnvelope := unwrapEnvelope(body)
	if errEnvelope != nil {
		return nil, nil, nil, errEnvelope
	}
	token, errToken := parseTokenData(data, "")
	if errToken != nil {
		return nil, nil, nil, errToken
	}
	user := parseUser(data)
	quota := parseQuota(data)
	return token, user, quota, nil
}

// Refresh rotates the token pair using the stored refresh token.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*TokenData, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("lobsterai-server: refresh token is required")
	}
	payload := map[string]any{"refreshToken": refreshToken}
	for key, value := range c.keyfromPayload() {
		payload[key] = value
	}
	body, status, errPost := c.postJSON(ctx, c.endpoints.Refresh, payload, "")
	if errPost != nil {
		return nil, errPost
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: refresh failed with status %d: %s", status, summarizeBody(body))
	}
	data, errEnvelope := unwrapEnvelope(body)
	if errEnvelope != nil {
		return nil, errEnvelope
	}
	// The refresh endpoint may not reissue the refresh token; callers keep the
	// stored one when it returns empty.
	return parseTokenData(data, refreshToken)
}

// FetchUserProfile reads the account profile with the access token. It is
// best-effort: the exchange already carries the identity.
func (c *Client) FetchUserProfile(ctx context.Context, accessToken string) (*User, error) {
	body, status, errGet := c.getJSON(ctx, c.appendKeyfromQuery(c.backendURL+c.endpoints.UserProfile), accessToken)
	if errGet != nil {
		return nil, errGet
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: user profile failed with status %d: %s", status, summarizeBody(body))
	}
	data, errEnvelope := unwrapEnvelope(body)
	if errEnvelope != nil {
		return nil, errEnvelope
	}
	return parseUser(data), nil
}

// FetchEnterpriseContext reads the account's joined-enterprise context.
//
// A nil context with a nil error is a normal answer meaning the account is
// personal (or is no longer a member of its enterprise): the caller must clear
// any persisted enterprise identity rather than keep sending a stale one.
// Errors are reserved for outcomes that must not erase a known context, which
// mirrors the desktop client keeping its last valid context on a transient
// failure.
func (c *Client) FetchEnterpriseContext(ctx context.Context, accessToken string) (*EnterpriseContext, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: access token is required")
	}
	body, status, errGet := c.getJSON(ctx, c.backendURL+c.endpoints.EnterpriseContext, accessToken)
	if errGet != nil {
		return nil, errGet
	}
	payload := decodeJSONObject(body)
	if payload == nil {
		return nil, fmt.Errorf("lobsterai-server: empty enterprise context response: %s", summarizeBody(body))
	}
	// An absent business code is a failure, not "personal": this endpoint always
	// answers the envelope, so its absence means an unexpected shape.
	code, codePresent := firstNumeric(payload, "code")
	if !codePresent {
		return nil, fmt.Errorf("lobsterai-server: enterprise context response has no business code: %s", summarizeBody(body))
	}
	if code != 0 {
		switch int64(code) {
		case EnterpriseErrorNotFound, EnterpriseErrorNotMember, EnterpriseErrorAccountModeMismatch:
			// Terminal: this account has no usable enterprise context any more.
			return nil, nil
		}
		message := firstNonEmpty(readString(payload, "message"), readString(payload, "msg"))
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("lobsterai-server: enterprise context failed (code %d): %s", int64(code), message)
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: enterprise context failed with status %d: %s", status, summarizeBody(body))
	}
	data, _ := payload["data"].(map[string]any)
	return parseEnterpriseContext(data), nil
}

// parseEnterpriseContext normalizes the enterprise context payload.
//
// It mirrors the desktop client's rules: an explicit personal mode, or a payload
// without a usable identity, yields no context. The id must be a positive
// integer, and the name and role must be present, so a half-populated payload
// cannot produce a header pair that scopes requests to an enterprise the account
// may not belong to.
func parseEnterpriseContext(data map[string]any) *EnterpriseContext {
	candidate := enterpriseContextCandidate(data)
	if candidate == nil {
		return nil
	}
	enterpriseID := positiveIntString(candidate["enterpriseId"])
	enterpriseName := readString(candidate, "enterpriseName")
	role := strings.ToLower(readString(candidate, "role"))
	if enterpriseID == "" || enterpriseName == "" || (role != "super_admin" && role != "member") {
		return nil
	}
	return &EnterpriseContext{
		AccountMode:    AccountModeEnterprise,
		EnterpriseID:   enterpriseID,
		EnterpriseName: enterpriseName,
		Role:           role,
		MemberID:       positiveIntString(candidate["memberId"]),
	}
}

// enterpriseContextCandidate finds the object carrying the enterprise identity.
//
// The backend nests it differently across surfaces (directly, or under one of
// several wrapper keys), and an explicit personal mode anywhere in the chain
// means the account is not an enterprise account.
func enterpriseContextCandidate(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	if strings.EqualFold(readString(value, "accountMode"), AccountModePersonal) {
		return nil
	}
	if value["enterpriseId"] != nil && readString(value, "enterpriseName") != "" {
		return value
	}
	for _, key := range []string{"enterpriseContext", "accountContext", "organizationContext", "context", "user", "data"} {
		nested := decodeJSONObjectField(value, key)
		if nested == nil {
			continue
		}
		if candidate := enterpriseContextCandidate(nested); candidate != nil {
			return candidate
		}
	}
	return nil
}

// positiveIntString reads a positive integer identity as a decimal string.
// A zero or negative id is rejected the same way the client rejects it.
func positiveIntString(value any) string {
	switch typed := value.(type) {
	case string:
		parsed, errParse := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if errParse == nil && parsed > 0 {
			return strconv.FormatInt(parsed, 10)
		}
	case float64:
		if typed > 0 && typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
	case json.Number:
		parsed, errParse := typed.Int64()
		if errParse == nil && parsed > 0 {
			return strconv.FormatInt(parsed, 10)
		}
	}
	return ""
}

// FetchModels reads the account's model catalogue with the access token.
// Entitlement is decided by the backend; the returned list is the routable
// candidate set for this credential.
func (c *Client) FetchModels(ctx context.Context, accessToken string) ([]Model, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: access token is required")
	}
	endpoint := c.appendKeyfromQuery(c.backendURL + c.endpoints.Models)
	body, status, errGet := c.getJSON(ctx, endpoint, accessToken)
	if errGet != nil {
		return nil, errGet
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: models failed with status %d: %s", status, summarizeBody(body))
	}
	payload := decodeJSONObject(body)
	if !isZeroCode(payload["code"]) {
		if rawCode, ok := payload["code"]; ok {
			return nil, fmt.Errorf("lobsterai-server: backend error (code %v): %s", rawCode, readString(payload, "message"))
		}
	}
	return parseModels(payload["data"]), nil
}

// FetchQuota reads the account credit/plan snapshot with the access token.
func (c *Client) FetchQuota(ctx context.Context, accessToken string) (*Quota, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: access token is required")
	}
	endpoint := c.appendKeyfromQuery(c.backendURL + c.endpoints.Quota)
	body, status, errGet := c.getJSON(ctx, endpoint, accessToken)
	if errGet != nil {
		return nil, errGet
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: quota failed with status %d: %s", status, summarizeBody(body))
	}
	data, errEnvelope := unwrapEnvelope(body)
	if errEnvelope != nil {
		return nil, errEnvelope
	}
	return parseQuota(data), nil
}

// FetchCreditSummary reads the account's credit total with its per-bucket
// breakdown.
//
// This is the read the desktop client's account menu renders
// (`profileSummary.totalCreditsRemaining` + `profileSummary.creditItems`), so it
// is the view that matches what the user sees. GET /api/user/quota is a
// different view (see FetchQuota): it carries only the free-tier bucket and
// drives the low-credit purchase gate, so using it alone understates an account
// whose credits also include activity and daily-login grants. The two are
// reported separately rather than one standing in for the other.
func (c *Client) FetchCreditSummary(ctx context.Context, accessToken string) (*CreditSummary, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: access token is required")
	}
	endpoint := c.appendKeyfromQuery(c.backendURL + c.endpoints.CreditSummary)
	body, status, errGet := c.getJSON(ctx, endpoint, accessToken)
	if errGet != nil {
		return nil, errGet
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: credit summary failed with status %d: %s", status, summarizeBody(body))
	}
	data, errEnvelope := unwrapEnvelope(body)
	if errEnvelope != nil {
		return nil, errEnvelope
	}
	return parseCreditSummary(data), nil
}

// ActivityPlacementDesktopSidebar is the placement the desktop client uses for
// the daily check-in activity.
const ActivityPlacementDesktopSidebar = "desktop_sidebar"

// ActivityContainerAPIVersionDailyCheckIn is the container API version the
// desktop client advertises when reading the daily check-in slot.
const ActivityContainerAPIVersionDailyCheckIn = 2

// ActivityPlatformFor maps a Go GOOS to the platform token the desktop client
// sends (its own process.platform: win32 / darwin / linux).
//
// Only those three are meaningful to the backend, which treats the value as an
// environment discriminator and answers an empty slot for anything else (an
// unrecognized token such as "cli" hides every activity). An unknown GOOS
// therefore yields "" rather than a fabricated "linux", so a caller can supply
// an explicit override and report the gap instead of silently claiming an
// environment the host is not.
func ActivityPlatformFor(goos string) string {
	switch strings.ToLower(strings.TrimSpace(goos)) {
	case "windows":
		return "win32"
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return ""
	}
}

// activityPlatform resolves the platform token for this client: an explicit
// override wins, then the host mapping, then the closest recognized environment.
func (c *Client) activityPlatform() string {
	if c != nil && c.activityPlatformOverride != "" {
		return c.activityPlatformOverride
	}
	if platform := ActivityPlatformFor(runtime.GOOS); platform != "" {
		return platform
	}
	// This host has no Electron equivalent. "linux" is the nearest recognized
	// environment, and SetActivityPlatform is how an operator corrects it.
	return "linux"
}

// FetchActivitySnapshot reads the read-only state of one activity placement.
//
// It never executes an activity action: the daily check-in is a user action in
// the desktop client, not something a proxy may perform on the account's
// behalf. A missing slot or an unauthenticated response is reported honestly as
// an empty snapshot rather than fabricated credit.
func (c *Client) FetchActivitySnapshot(ctx context.Context, accessToken, placement string, containerAPIVersion int) (*ActivitySnapshot, error) {
	placement = strings.TrimSpace(placement)
	if placement == "" {
		placement = ActivityPlacementDesktopSidebar
	}
	if containerAPIVersion <= 0 {
		containerAPIVersion = ActivityContainerAPIVersionDailyCheckIn
	}
	query := url.Values{}
	query.Set("placement", placement)
	query.Set("clientVersion", c.version)
	query.Set("containerApiVersion", strconv.Itoa(containerAPIVersion))
	query.Set("platform", c.activityPlatform())
	snapshot := &ActivitySnapshot{Placement: placement}
	endpoint := c.backendURL + c.endpoints.ActivitySlot + "?" + query.Encode()
	body, status, errGet := c.getJSON(ctx, endpoint, accessToken)
	if errGet != nil {
		return snapshot, errGet
	}
	if status < 200 || status >= 300 {
		return snapshot, fmt.Errorf("lobsterai-server: activity slot failed with status %d: %s", status, summarizeBody(body))
	}
	data, errEnvelope := unwrapEnvelope(body)
	if errEnvelope != nil {
		return snapshot, errEnvelope
	}
	snapshot.SlotState = readString(data, "slotState")
	if activity := decodeJSONObjectField(data, "activity"); activity != nil {
		snapshot.ActivityCode = readString(activity, "activityCode")
		snapshot.ConfigRevision = readInt64(activity, "configRevision")
	}
	if snapshot.SlotState != "available" || snapshot.ActivityCode == "" {
		return snapshot, nil
	}
	contextQuery := url.Values{}
	contextQuery.Set("configRevision", strconv.FormatInt(snapshot.ConfigRevision, 10))
	contextEndpoint := c.backendURL + c.endpoints.ActivityContext + "/" +
		url.PathEscape(snapshot.ActivityCode) + "/context?" + contextQuery.Encode()
	contextBody, contextStatus, errContext := c.getJSON(ctx, contextEndpoint, accessToken)
	if errContext != nil {
		return snapshot, errContext
	}
	if contextStatus < 200 || contextStatus >= 300 {
		return snapshot, fmt.Errorf("lobsterai-server: activity context failed with status %d: %s", contextStatus, summarizeBody(contextBody))
	}
	contextData, errContextEnvelope := unwrapEnvelope(contextBody)
	if errContextEnvelope != nil {
		return snapshot, errContextEnvelope
	}
	snapshot.Context = contextData
	return snapshot, nil
}

// ClaimActivity executes one activity action, which is what the daily check-in
// does inside the desktop client.
//
// This is the only write this client performs against the activity surface;
// FetchActivitySnapshot stays read-only, so a caller that only wants to observe
// never grants anything. The backend deduplicates by idempotencyKey and answers
// ActivityAlreadyClaimedCode once the period is accounted for, which is reported
// as ErrActivityAlreadyClaimed rather than as a failure: the reward was granted
// earlier, so a retry must not be escalated as an error.
func (c *Client) ClaimActivity(ctx context.Context, accessToken string, request ActivityClaimRequest) (*ActivityClaim, error) {
	if c == nil {
		return nil, fmt.Errorf("lobsterai-server: client is nil")
	}
	activityCode := strings.TrimSpace(request.ActivityCode)
	if activityCode == "" {
		return nil, fmt.Errorf("lobsterai-server: activity code is required")
	}
	actionID := strings.TrimSpace(request.ActionID)
	if actionID == "" {
		actionID = ActivityActionCheckIn
	}
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: access token is required")
	}
	idempotencyKey := strings.TrimSpace(request.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = newActivityIdempotencyKey()
	}
	payload := map[string]any{
		"configRevision": request.ConfigRevision,
		"idempotencyKey": idempotencyKey,
		"payload":        map[string]any{},
	}
	path := c.endpoints.ActivityContext + "/" + url.PathEscape(activityCode) +
		"/actions/" + url.PathEscape(actionID)
	body, status, errPost := c.postJSON(ctx, path, payload, accessToken)
	if errPost != nil {
		return nil, errPost
	}
	raw := decodeJSONObject(body)
	if raw == nil {
		return nil, fmt.Errorf("lobsterai-server: empty action response: %s", summarizeBody(body))
	}
	// An absent business code is a failure, not a success: the backend always
	// answers the envelope, so a missing code means the response is not the
	// contract this code understands.
	code, codePresent := firstNumeric(raw, "code")
	if !codePresent {
		return nil, fmt.Errorf("lobsterai-server: action response has no business code: %s", summarizeBody(body))
	}
	if int64(code) == ActivityAlreadyClaimedCode {
		return nil, ErrActivityAlreadyClaimed
	}
	if code != 0 {
		message := firstNonEmpty(readString(raw, "message"), readString(raw, "msg"))
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("lobsterai-server: activity action failed (code %d): %s", int64(code), message)
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("lobsterai-server: activity action failed with status %d: %s", status, summarizeBody(body))
	}
	data, _ := raw["data"].(map[string]any)
	result := decodeJSONObjectField(data, "result")
	if result == nil {
		result = data
	}
	claim := &ActivityClaim{Raw: cloneAnyMap(result)}
	// The granted amount is named differently across activity types.
	if granted, ok := firstNumeric(claim.Raw, "creditsGranted", "rewardCredits", "credits"); ok {
		claim.Granted = granted
		claim.GrantedKnown = true
	}
	claim.PeriodKey = readString(claim.Raw, "periodKey", "period_key")
	claim.Message = firstNonEmpty(readString(claim.Raw, "message"), readString(claim.Raw, "msg"))
	return claim, nil
}

// newActivityIdempotencyKey builds the client-side dedupe key the backend
// accepts, which must match ^[A-Za-z0-9._:-]{1,64}$.
func newActivityIdempotencyKey() string {
	buffer := make([]byte, 8)
	if _, errRead := rand.Read(buffer); errRead != nil {
		return fmt.Sprintf("cli-%d", time.Now().UnixMilli())
	}
	return fmt.Sprintf("cli-%d-%s", time.Now().UnixMilli(), hex.EncodeToString(buffer))
}

func (c *Client) getJSON(ctx context.Context, endpoint, token string) ([]byte, int, error) {
	return c.doJSONURL(ctx, http.MethodGet, endpoint, nil, token)
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, token string) ([]byte, int, error) {
	return c.doJSONURL(ctx, http.MethodPost, c.backendURL+path, payload, token)
}

func (c *Client) doJSONURL(ctx context.Context, method, endpoint string, payload any, token string) ([]byte, int, error) {
	if c == nil {
		return nil, 0, fmt.Errorf("lobsterai-server: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var reader io.Reader
	if payload != nil {
		encoded, errMarshal := json.Marshal(payload)
		if errMarshal != nil {
			return nil, 0, fmt.Errorf("lobsterai-server: encode request: %w", errMarshal)
		}
		reader = bytes.NewReader(encoded)
	}
	req, errReq := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if errReq != nil {
		return nil, 0, fmt.Errorf("lobsterai-server: create request: %w", errReq)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-LobsterAI-Client-Capabilities", ClientCapabilities)
	req.Header.Set("X-LobsterAI-Client-Version", c.version)
	// The enterprise pair is sent on every authenticated request once known; a
	// personal account has none and therefore sends neither header.
	if strings.EqualFold(c.accountMode, AccountModeEnterprise) && c.enterpriseID != "" {
		req.Header.Set(HeaderAccountMode, AccountModeEnterprise)
		req.Header.Set(HeaderEnterpriseID, c.enterpriseID)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if trimmed := strings.TrimSpace(token); trimmed != "" {
		req.Header.Set("Authorization", "Bearer "+trimmed)
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, 0, fmt.Errorf("lobsterai-server: request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("lobsterai-server: close response body: %v", errClose)
		}
	}()
	body, errRead := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if errRead != nil {
		return nil, resp.StatusCode, fmt.Errorf("lobsterai-server: read response: %w", errRead)
	}
	return body, resp.StatusCode, nil
}

// unwrapEnvelope flattens the backend's {code,message,data:{...}} envelope. A
// non-zero code is reported so callers fail loudly instead of parsing an error
// payload as success.
func unwrapEnvelope(body []byte) (map[string]any, error) {
	payload := decodeJSONObject(body)
	if payload == nil {
		return nil, fmt.Errorf("lobsterai-server: empty response body: %s", summarizeBody(body))
	}
	if rawCode, ok := payload["code"]; ok && !isZeroCode(rawCode) {
		message := readString(payload, "message")
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("lobsterai-server: backend error (code %v): %s", rawCode, message)
	}
	if data, ok := payload["data"].(map[string]any); ok {
		return data, nil
	}
	return payload, nil
}

func isZeroCode(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
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

func parseTokenData(data map[string]any, previousRefreshToken string) (*TokenData, error) {
	accessToken := readString(data, "accessToken", "access_token")
	refreshToken := readString(data, "refreshToken", "refresh_token")
	if refreshToken == "" {
		refreshToken = previousRefreshToken
	}
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: token response missing accessToken")
	}
	token := &TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    readPositiveInt(data, "expiresIn", "expires_in"),
		ExpiresAt:    expiryTime(data),
	}
	if token.ExpiresAt.IsZero() && token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	return token, nil
}

// expiryTime reads an absolute access-token expiry when the backend reports one.
func expiryTime(data map[string]any) time.Time {
	for _, key := range []string{"expired", "expiresAt", "expireAt"} {
		raw, ok := data[key]
		if !ok {
			continue
		}
		switch value := raw.(type) {
		case string:
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
				if parsed, errParse := time.Parse(layout, trimmed); errParse == nil {
					return parsed
				}
			}
		case float64:
			// Heuristic used elsewhere in this repo: seconds vs milliseconds.
			seconds := int64(value)
			if seconds > 1_000_000_000_000 {
				return time.UnixMilli(seconds)
			}
			return time.Unix(seconds, 0)
		}
	}
	return time.Time{}
}

// parseUser extracts the account identity from an exchange/user payload. The
// backend nests it under "user"; a flat payload is also accepted.
//
// The live profile reports the stable account id as a NUMBER and the login
// identifier as "yid" (the exchange response carries "id"), so both spellings
// and both scalar types are accepted.
func parseUser(data map[string]any) *User {
	user := data
	if nested := decodeJSONObjectField(data, "user"); nested != nil {
		user = nested
	}
	return &User{
		ID:       firstNonEmpty(readString(user, "id", "userId", "user_id", "uid"), readScalarString(user, "id", "userId", "user_id", "uid")),
		Email:    readString(user, "email", "mail", "yid"),
		Name:     readString(user, "name", "displayName", "display_name", "nickname"),
		Avatar:   readString(user, "avatarUrl", "avatar", "image", "picture"),
		Nickname: readString(user, "nickname", "nickName"),
	}
}

// parseQuota flattens the credit payload into a normalized snapshot. The
// exchange nests it under "quota"; /api/user/quota returns it directly.
//
// The live payload reports the plan as subscriptionStatus/planName and the
// credit balance as freeCreditsRemaining/purchasedCredits. Unknown values stay
// absent rather than being reported as zero or unlimited.
func parseQuota(data map[string]any) *Quota {
	if data == nil {
		return &Quota{}
	}
	if nested := decodeJSONObjectField(data, "quota"); nested != nil {
		data = nested
	}
	quota := &Quota{Raw: cloneAnyMap(data)}
	quota.SubscriptionStatus = readString(data, "subscriptionStatus", "subscription_status", "planType", "plan", "planName", "plan_name")
	if credits, ok := firstNumeric(data, "freeCreditsRemaining", "free_credits_remaining", "credits", "credit", "remainingCredits", "remaining_credits", "quotaRemaining", "quota_remaining", "balance"); ok {
		quota.Credits = int64(credits)
		quota.CreditsKnown = true
	}
	if purchased, ok := firstNumeric(data, "purchasedCredits", "purchased_credits", "paidCredits", "paid_credits"); ok {
		quota.PurchasedCredits = int64(purchased)
	}
	return quota
}

// parseCreditSummary maps the profile-summary payload into a CreditSummary.
//
// The total stays unknown unless the backend reports it: a missing
// totalCreditsRemaining is not a zero balance, and the handler must be able to
// tell the two apart before it renders anything.
func parseCreditSummary(data map[string]any) *CreditSummary {
	summary := &CreditSummary{}
	if data == nil {
		return summary
	}
	summary.Raw = cloneAnyMap(data)
	if total, ok := firstNumeric(data, "totalCreditsRemaining", "total_credits_remaining"); ok {
		summary.TotalCreditsRemaining = total
		summary.TotalKnown = true
	}
	summary.UserID = readScalarString(data, "id", "userId", "user_id")
	summary.Nickname = readString(data, "nickname", "nickName")
	if items, ok := data["creditItems"].([]any); ok {
		packages := make([]CreditPackage, 0, len(items))
		for _, item := range items {
			entry, okEntry := item.(map[string]any)
			if !okEntry {
				continue
			}
			remaining, _ := firstNumeric(entry, "creditsRemaining", "credits_remaining")
			packages = append(packages, CreditPackage{
				Type:             readString(entry, "type"),
				Label:            readString(entry, "label"),
				LabelEn:          readString(entry, "labelEn", "label_en"),
				CreditsRemaining: remaining,
				ExpiresAt:        readString(entry, "expiresAt", "expires_at"),
			})
		}
		summary.Packages = packages
	}
	return summary
}

// parseModels maps the /api/models/available data array into normalized models.
//
// The backend reports the identity as modelId/modelName (a bare "id"/"name" is
// also accepted defensively) and marks unroutable entries with accessible=false.
// Entitlement is enforced upstream with code 40301, so inaccessible models are
// dropped here rather than registered and then rejected at request time.
func parseModels(raw any) []Model {
	entries, ok := raw.([]any)
	if !ok {
		return nil
	}
	models := make([]Model, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, item := range entries {
		entry, okEntry := item.(map[string]any)
		if !okEntry {
			continue
		}
		if accessible, present := entry["accessible"].(bool); present && !accessible {
			continue
		}
		id := readString(entry, "modelId", "model_id", "id", "model", "name")
		if id == "" {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		models = append(models, Model{
			ID:                  id,
			DisplayName:         readString(entry, "modelName", "displayName", "display_name", "name", "label"),
			APIFormat:           readString(entry, "apiFormat", "api_format", "format", "protocol"),
			Thinking:            readBool(entry, "supportsThinking", "supports_thinking") || decodeJSONObjectField(entry, "thinkingConfig") != nil,
			Image:               readBool(entry, "supportsImage", "supports_image"),
			Video:               readBool(entry, "supportsVideo", "supports_video"),
			Accessible:          true,
			SupportsToolCalling: readBool(entry, "supportsToolCalling", "supports_tool_calling"),
			RuntimeProfile:      readString(entry, "runtimeProfile", "runtime_profile"),
			ContextLimit:        int(readInt64(entry, "contextWindow", "context_window", "contextLength", "context_length", "contextLimit")),
			MaxTokens:           int(readInt64(entry, "maxTokens", "max_tokens", "maxOutputTokens")),
			ReasoningLevels:     parseReasoningLevels(entry),
		})
	}
	return models
}

// parseReasoningLevels extracts the model's declared reasoning-effort wire values.
//
// The backend reports both a product-side `level` (off/minimal/…/max, used by its
// own UI) and the `openclawLevel` it actually sends. Only the latter belongs in
// `reasoning_effort`, and they differ exactly at the top of the scale: the UI
// calls it "max" while the wire value is "xhigh". Sending the UI name would look
// like it worked (the gateway answers 200) while quietly running at the default.
//
// A model with no usable options yields no levels, which keeps the request field
// untouched rather than inventing a vocabulary the backend never declared.
func parseReasoningLevels(entry map[string]any) []string {
	config := decodeJSONObjectField(entry, "thinkingConfig")
	if config == nil {
		return nil
	}
	options, _ := config["options"].([]any)
	if len(options) == 0 {
		return nil
	}
	levels := make([]string, 0, len(options))
	seen := make(map[string]struct{}, len(options))
	for _, item := range options {
		option, okOption := item.(map[string]any)
		if !okOption {
			continue
		}
		level := strings.TrimSpace(readString(option, "openclawLevel", "openclaw_level"))
		if level == "" {
			continue
		}
		key := strings.ToLower(level)
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		levels = append(levels, level)
	}
	if len(levels) == 0 {
		return nil
	}
	return levels
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func cloneAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	clone := make(map[string]any, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func decodeJSONObject(body []byte) map[string]any {
	if len(bytes.TrimSpace(body)) == 0 {
		return map[string]any{}
	}
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

// readScalarString returns the textual form of a string or numeric field. The
// backend mixes types for the account id (number on /api/user/profile, string in
// the exchange response), so a string-only reader would silently drop it.
func readScalarString(payload map[string]any, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		switch value := payload[key].(type) {
		case string:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		case float64:
			if value == float64(int64(value)) {
				return strconv.FormatInt(int64(value), 10)
			}
			return strconv.FormatFloat(value, 'f', -1, 64)
		}
	}
	return ""
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
		if value, ok := numericValue(payload[key]); ok && value > 0 {
			return int(value)
		}
	}
	return 0
}

func readInt64(payload map[string]any, keys ...string) int64 {
	if payload == nil {
		return 0
	}
	for _, key := range keys {
		if value, ok := numericValue(payload[key]); ok {
			return int64(value)
		}
	}
	return 0
}

func firstNumeric(payload map[string]any, keys ...string) (float64, bool) {
	if payload == nil {
		return 0, false
	}
	for _, key := range keys {
		if value, ok := numericValue(payload[key]); ok {
			return value, true
		}
	}
	return 0, false
}

func numericValue(raw any) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case json.Number:
		parsed, errParse := value.Float64()
		return parsed, errParse == nil
	case string:
		parsed, errParse := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed, errParse == nil
	default:
		return 0, false
	}
}

func summarizeBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if len(trimmed) > 512 {
		return trimmed[:512] + "…"
	}
	return trimmed
}
