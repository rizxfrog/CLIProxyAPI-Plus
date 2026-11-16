// Console API client for the Baidu Cloud console endpoints that back the Dumate
// points wallet and the daily sign-in bonus.
//
// These endpoints live on https://console.bce.baidu.com and are NOT reached
// through the client's local backend: the Electron main process calls them
// directly with the decrypted console cookie. The proxy therefore needs the
// cookie in the credential file to read the balance or claim the daily bonus.
//
// Every response shares the envelope
//
//	{"logId":"<uuid>","code":0,"message":"","success":true,"result":<T>}
//
// where only code == 0 && success == true means the call succeeded; HTTP 200 on
// its own proves nothing.
package dumate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	// ConsoleBaseURL is the Baidu Cloud console origin.
	ConsoleBaseURL = "https://console.bce.baidu.com"

	// ConsoleTimezoneDefault is the timezone used for day boundaries when the
	// caller does not specify one. It only shifts which calendar day the
	// daily-bonus bookkeeping uses; it does not change the totals.
	ConsoleTimezoneDefault = "Asia/Shanghai"

	consoleQuotaOverviewPath = "/api/dumate/points/quota_overview"
	consoleLoginBonusPath    = "/api/dumate/points/loginBonus"
	consoleLoginBonusInfo    = "/api/dumate/points/loginBonusInfo"

	consoleResponseLimit = 1 << 20
)

// ConsoleError is a structured console API failure.
type ConsoleError struct {
	StatusCode int
	Code       int
	Message    string
	LogID      string
	// SessionInvalid marks a failure that no retry can fix because the stored
	// session cannot authorize the call at all.
	SessionInvalid bool
}

func (e *ConsoleError) Error() string {
	switch {
	case e.Message != "":
		return fmt.Sprintf("dumate console error: %s (http %d, code %d)", e.Message, e.StatusCode, e.Code)
	case e.StatusCode != 0:
		return fmt.Sprintf("dumate console error: http %d", e.StatusCode)
	default:
		return "dumate console error"
	}
}

// SessionExpired reports whether the console refused the request because the
// stored login session is no longer usable.
//
// The session check lives with the client because a stale session is not a
// transient upstream fault: the credential cannot recover on its own, so
// callers must surface it as a re-login requirement instead of retrying.
func (e *ConsoleError) SessionExpired() bool {
	if e == nil {
		return false
	}
	if e.SessionInvalid {
		return true
	}
	if e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden {
		return true
	}
	// The console answers a rejected session with HTTP 200 and a localized
	// message, so the text is the only signal available. Sending the CSRF header
	// keeps this text meaningful: without it, every write reports the same thing.
	return strings.Contains(e.Message, "登录凭证已过期") ||
		strings.Contains(e.Message, "请重新登录") ||
		strings.Contains(strings.ToLower(e.Message), "not logged in") ||
		strings.Contains(strings.ToLower(e.Message), "login expired")
}

// ConsoleClient performs authenticated console requests with an injected HTTP
// client so per-credential proxy settings and cancellation still apply.
type ConsoleClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewConsoleClient wraps an already proxy-aware HTTP client.
func NewConsoleClient(httpClient *http.Client) *ConsoleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ConsoleClient{httpClient: httpClient, baseURL: ConsoleBaseURL}
}

// NewConsoleClientWithBaseURL points the client at another origin. Tests and
// self-hosted mirrors use it; production callers keep the fixed console origin
// so credential metadata can never redirect a privileged read.
func NewConsoleClientWithBaseURL(httpClient *http.Client, baseURL string) *ConsoleClient {
	client := NewConsoleClient(httpClient)
	client.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return client
}

// QuotaPackage is one points bucket: a subscription window, a purchased bundle
// or a granted bonus.
type QuotaPackage struct {
	PackageID        string `json:"packageId"`
	UsedPoints       string `json:"usedPoints"`
	TotalPoints      string `json:"totalPoints"`
	StartDate        int64  `json:"startDate"`
	ExpireDate       int64  `json:"expireDate"`
	PointsExpireDate int64  `json:"pointsExpireDate,omitempty"`
	PackageType      string `json:"packageType"`
	Source           string `json:"source"`
	Status           string `json:"status"`
	AutoRenewStatus  string `json:"autoRenewStatus,omitempty"`
	NextAutoRenewAt  int64  `json:"nextAutoRenewAt,omitempty"`
	Product          string `json:"product,omitempty"`
}

// QuotaThrottle reports whether the account is currently rate limited.
type QuotaThrottle struct {
	Throttled    bool   `json:"throttled"`
	Reason       string `json:"reason"`
	ThrottleType string `json:"throttleType"`
}

// PointsOverview is the points wallet shown by Settings -> Points.
//
// Subscription holds the plan windows; Incremental holds purchased and granted
// buckets. TotalPoints already accounts for the used portion of every active
// bucket.
type PointsOverview struct {
	IsSubscribed bool `json:"isSubscribed"`
	// UsedPoints and TotalPoints are decimal strings, not numbers.
	UsedPoints        string         `json:"usedPoints"`
	TotalPoints       string         `json:"totalPoints"`
	Subscription      []QuotaPackage `json:"subscription"`
	Incremental       []QuotaPackage `json:"incremental"`
	ModelThrottleInfo QuotaThrottle  `json:"modelThrottleInfo"`
}

// LoginBonusInfo is the daily sign-in state.
type LoginBonusInfo struct {
	TotalTimes  int      `json:"totalTimes"`
	TotalPoints float64  `json:"totalPoints"`
	HasIssued   bool     `json:"hasIssued"`
	SignInDays  []string `json:"signInDays"`
}

// SignedInOn reports whether the given calendar day (YYYY-MM-DD) is already in
// the server's sign-in list.
func (i *LoginBonusInfo) SignedInOn(day string) bool {
	if i == nil || strings.TrimSpace(day) == "" {
		return false
	}
	for _, entry := range i.SignInDays {
		if strings.TrimSpace(entry) == day {
			return true
		}
	}
	return false
}

type consoleEnvelope struct {
	LogID   string          `json:"logId"`
	Code    int             `json:"code"`
	Message consoleMessage  `json:"message"`
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
}

// consoleMessage accepts both message shapes the console returns: a plain
// string on success, and an object such as {"global":"..."} on failure. The
// object form is what a rejected session looks like, so decoding it as a string
// would turn every expired cookie into an opaque parse error.
type consoleMessage struct {
	Text string
}

func (m *consoleMessage) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		m.Text = ""
		return nil
	}
	if trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		m.Text = strings.TrimSpace(text)
		return nil
	}
	if trimmed[0] != '{' {
		// Any other scalar shape is unexpected but must not fail the response.
		m.Text = trimmed
		return nil
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	// "global" is the console's general failure slot.
	if global, ok := fields["global"].(string); ok {
		if text := strings.TrimSpace(global); text != "" {
			m.Text = text
			return nil
		}
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if text := strings.TrimSpace(fmt.Sprint(fields[key])); text != "" {
			parts = append(parts, key+": "+text)
		}
	}
	m.Text = strings.Join(parts, "; ")
	return nil
}

// FetchPointsOverview reads the points wallet.
func (c *ConsoleClient) FetchPointsOverview(ctx context.Context, cookie, timezone string) (*PointsOverview, error) {
	zone := strings.TrimSpace(timezone)
	if zone == "" {
		zone = ConsoleTimezoneDefault
	}
	query := url.Values{}
	query.Set("timezone", zone)
	query.Set("clientType", "desktop")
	// Match the client: an unclaimed daily bonus is not part of the displayed balance.
	query.Set("ignoreLoginBonus", "true")
	raw, err := c.get(ctx, cookie, consoleQuotaOverviewPath+"?"+query.Encode())
	if err != nil {
		return nil, err
	}
	var overview PointsOverview
	if err := json.Unmarshal(raw, &overview); err != nil {
		return nil, fmt.Errorf("dumate console: decode points overview: %w", err)
	}
	return &overview, nil
}

// FetchLoginBonusInfo reads the daily sign-in state without claiming it.
func (c *ConsoleClient) FetchLoginBonusInfo(ctx context.Context, cookie string) (*LoginBonusInfo, error) {
	raw, err := c.get(ctx, cookie, consoleLoginBonusInfo)
	if err != nil {
		return nil, err
	}
	var info LoginBonusInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("dumate console: decode login bonus info: %w", err)
	}
	return &info, nil
}

// ClaimLoginBonus claims today's sign-in bonus.
//
// The upstream has no idempotency key, so callers must check
// FetchLoginBonusInfo first; this call performs the state change.
func (c *ConsoleClient) ClaimLoginBonus(ctx context.Context, cookie string) (bool, error) {
	raw, err := c.post(ctx, cookie, consoleLoginBonusPath)
	if err != nil {
		return false, err
	}
	var claimed bool
	if err := json.Unmarshal(raw, &claimed); err != nil {
		return false, fmt.Errorf("dumate console: decode login bonus result: %w", err)
	}
	return claimed, nil
}

func (c *ConsoleClient) get(ctx context.Context, cookie, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, cookie, path)
}

func (c *ConsoleClient) post(ctx context.Context, cookie, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, cookie, path)
}

func (c *ConsoleClient) do(ctx context.Context, method, cookie, path string) (json.RawMessage, error) {
	if strings.TrimSpace(cookie) == "" {
		return nil, &ConsoleError{Message: "console cookie is missing"}
	}
	base := c.baseURL
	if base == "" {
		base = ConsoleBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, nil)
	if err != nil {
		return nil, fmt.Errorf("dumate console: build request: %w", err)
	}
	csrfToken := consoleCSRFToken(cookie)
	if csrfToken == "" && method != http.MethodGet && method != http.MethodHead {
		// Fail before the call: the console would answer a tokenless write with a
		// misleading "login expired" body, turning a header bug into a phantom
		// credential problem.
		return nil, &ConsoleError{Message: "console session is missing its CSRF token", SessionInvalid: true}
	}
	applyConsoleHeaders(req, cookie, csrfToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dumate console: request failed: %w", err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			_ = errClose
		}
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, consoleResponseLimit))
	if err != nil {
		return nil, fmt.Errorf("dumate console: read response: %w", err)
	}

	var envelope consoleEnvelope
	if errUnmarshal := json.Unmarshal(body, &envelope); errUnmarshal != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, &ConsoleError{StatusCode: resp.StatusCode}
		}
		return nil, fmt.Errorf("dumate console: decode response: %w", errUnmarshal)
	}
	if resp.StatusCode != http.StatusOK || !envelope.Success || envelope.Code != 0 {
		return nil, &ConsoleError{
			StatusCode: resp.StatusCode,
			Code:       envelope.Code,
			Message:    envelope.Message.Text,
			LogID:      envelope.LogID,
		}
	}
	return envelope.Result, nil
}

// consoleCSRFToken returns the CSRF token the console requires on writes.
//
// The desktop client reads it from the bce-user-info cookie, stripping the
// surrounding quotes; sending it back verbatim as the csrftoken header is what
// authorizes a state change. Reads work without it, so a missing token only
// surfaces when a claim is attempted.
func consoleCSRFToken(cookie string) string {
	for _, part := range strings.Split(cookie, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || name != "bce-user-info" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}

// applyConsoleHeaders mirrors the identity headers the desktop client sends.
// The session itself is carried entirely by the cookie.
//
// The csrftoken header is not optional: the console accepts reads without it but
// rejects every write with a misleading "login expired" body, which is why a
// claim must reproduce the client's own header set.
func applyConsoleHeaders(req *http.Request, cookie, csrfToken string) {
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	if csrfToken != "" {
		req.Header.Set("csrftoken", csrfToken)
	}
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		// The client's own write requests declare a JSON body even when empty.
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	req.Header.Set("Origin", ConsoleBaseURL)
	req.Header.Set("Referer", ConsoleBaseURL+"/")
	req.Header.Set("User-Agent", consoleUserAgent)
}

// consoleUserAgent matches the desktop client's Electron identity so the
// console does not treat the request as an unknown client.
const consoleUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/150.0.7871.114 Electron/43.1.1 Safari/537.36"
