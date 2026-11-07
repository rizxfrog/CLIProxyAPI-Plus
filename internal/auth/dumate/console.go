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
	Message string          `json:"message"`
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
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
	applyConsoleHeaders(req, cookie)

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
			Message:    strings.TrimSpace(envelope.Message),
			LogID:      envelope.LogID,
		}
	}
	return envelope.Result, nil
}

// applyConsoleHeaders mirrors the identity headers the desktop client sends.
// The session itself is carried entirely by the cookie.
func applyConsoleHeaders(req *http.Request, cookie string) {
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", ConsoleBaseURL)
	req.Header.Set("Referer", ConsoleBaseURL+"/")
	req.Header.Set("User-Agent", consoleUserAgent)
}

// consoleUserAgent matches the desktop client's Electron identity so the
// console does not treat the request as an unknown client.
const consoleUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/150.0.7871.114 Electron/43.1.1 Safari/537.36"
