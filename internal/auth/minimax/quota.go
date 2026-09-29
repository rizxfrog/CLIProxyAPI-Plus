package minimax

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	log "github.com/sirupsen/logrus"
)

// Signed matrix endpoints used to resolve the coding-plan quota.
const (
	userInfoPath         = "/v1/api/user/info"
	userExtraInfoPath    = "/matrix/api/v1/user/get_user_extra_info"
	membershipInfoPath   = "/matrix/api/v1/commerce/get_membership_info"
	tokenPlanRemainsPath = "/v1/api/openplatform/coding_plan/remains"

	// signatureSalt is the shared wire constant from the client's matrix
	// account client. It tags a request as first-party; authorization is still
	// the Bearer token. Treat it as a protocol constant.
	signatureSalt     = "I*7Cf%WZ#S&%1RlZJ&C2"
	signatureSuffix   = "ooui"
	matrixVersionCode = "22201"
	matrixAppID       = "3001"
	matrixBizID       = "3"
	matrixUserAgent   = "MiniMaxCode"
)

// QuotaWindow is one rate-limit window (5-hour "interval" or "weekly").
type QuotaWindow struct {
	// RemainingPercent is the remaining allowance percentage (0-100). It is nil
	// when Unlimited is true or when the upstream reported nothing usable.
	RemainingPercent *int
	// ResetAtMs is the window reset instant in epoch milliseconds, 0 when absent.
	ResetAtMs int64
	// Unlimited reports the upstream "no limit" status.
	Unlimited bool
}

// QuotaVideo is the optional video-generation bucket.
type QuotaVideo struct {
	RemainingCount *int
	TotalCount     *int
	ResetAtMs      int64
	Unlimited      bool
}

// QuotaSnapshot is the coding-plan quota for one account.
type QuotaSnapshot struct {
	FiveHour QuotaWindow
	Weekly   QuotaWindow
	Video    *QuotaVideo
}

// Membership is the token-plan membership summary for one account.
type Membership struct {
	HasTokenPlan  *bool
	OpGroupID     string
	Tier          string
	ExpiresAtMs   int64
	CreditBalance string
}

// AccountQuota is the full quota observation for one credential: the plan
// membership plus, when the account is subscribed and an operation group is
// known, the coding-plan rate-limit windows.
type AccountQuota struct {
	Membership Membership
	Snapshot   *QuotaSnapshot
	// NotSubscribed reports an explicit has_token_plan == false.
	NotSubscribed bool
}

// FetchAccountQuota resolves the account identity, membership and coding-plan
// quota using the same signed matrix calls and open-platform probe as the
// native MiniMax Code client. It mirrors getTokenPlanAccountStatus:
//
//   - realUserID from GET /v1/api/user/info (signed)
//   - workspace + membership from POST /matrix/api/v1/user/get_user_extra_info,
//     falling back to POST /matrix/api/v1/commerce/get_membership_info
//   - the coding-plan windows from GET /v1/api/openplatform/coding_plan/remains
//     with X-Group-Id set to the personal operation group id
func (c *Client) FetchAccountQuota(ctx context.Context, accessToken string) (*AccountQuota, error) {
	if c == nil {
		return nil, fmt.Errorf("minimax: client is nil")
	}
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("minimax: access token is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	matrixOrigin := c.matrixOrigin()

	realUserID, errIdentity := c.fetchRealUserID(ctx, matrixOrigin, accessToken)
	if errIdentity != nil {
		return nil, errIdentity
	}
	nowMs := time.Now().UnixMilli()

	// Personal workspace first; its membership carries the op_group_id the quota
	// probe needs. A failure falls through to the plain membership read.
	membership := Membership{}
	var personalOpGroupID string
	hasPersonalWorkspace := false
	if extra, errExtra := c.postMatrixJSON(ctx, matrixOrigin, userExtraInfoPath, "{}", accessToken, realUserID, nowMs); errExtra == nil {
		if workspaceID, personal, ok := projectPersonalWorkspace(extra); ok {
			hasPersonalWorkspace = true
			membership = personal
			personalOpGroupID = personal.OpGroupID
			// A scoped membership read can add tier/credit fields the personal
			// workspace omitted; keep the personal values when it fails.
			body := fmt.Sprintf(`{"workspace_id":%s}`, workspaceID)
			if scoped, errScoped := c.postMatrixJSON(ctx, matrixOrigin, membershipInfoPath, body, accessToken, realUserID, nowMs); errScoped == nil {
				membership = mergeMembership(personal, projectMembership(scoped))
				if membership.OpGroupID == "" {
					membership.OpGroupID = personalOpGroupID
				}
				personalOpGroupID = membership.OpGroupID
			}
		}
	}
	if !hasPersonalWorkspace {
		if plain, errPlain := c.postMatrixJSON(ctx, matrixOrigin, membershipInfoPath, "{}", accessToken, realUserID, nowMs); errPlain == nil {
			membership = projectMembership(plain)
			personalOpGroupID = membership.OpGroupID
		}
	}

	result := &AccountQuota{Membership: membership}
	if membership.HasTokenPlan != nil && !*membership.HasTokenPlan {
		result.NotSubscribed = true
		return result, nil
	}
	if personalOpGroupID == "" {
		// Without an operation group the upstream cannot address the plan quota.
		return result, nil
	}
	snapshot, errQuota := c.fetchCodingPlanQuota(ctx, accessToken, personalOpGroupID)
	if errQuota != nil {
		return result, errQuota
	}
	result.Snapshot = snapshot
	return result, nil
}

// fetchRealUserID reads the account identity and returns its realUserID.
func (c *Client) fetchRealUserID(ctx context.Context, matrixOrigin, accessToken string) (string, error) {
	body, errGet := c.signedGet(ctx, matrixOrigin, userInfoPath, accessToken, "0", time.Now().UnixMilli())
	if errGet != nil {
		return "", errGet
	}
	payload, ok := body.(map[string]any)
	if !ok {
		return "", fmt.Errorf("minimax: identity response is not an object")
	}
	data := asMap(payload["data"])
	userInfo := firstMap(
		asMap(data["userInfo"]), asMap(data["user_info"]),
		asMap(payload["userInfo"]), asMap(payload["user_info"]),
	)
	if userInfo == nil {
		return "", fmt.Errorf("minimax: identity response has no user info")
	}
	for _, key := range []string{"realUserID", "real_user_id"} {
		if value := readMapString(userInfo, key); value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf("minimax: identity response has no account id")
}

// fetchCodingPlanQuota reads the rate-limit windows from the open-platform probe.
func (c *Client) fetchCodingPlanQuota(ctx context.Context, accessToken, opGroupID string) (*QuotaSnapshot, error) {
	endpoint := c.openPlatformOrigin() + tokenPlanRemainsPath
	req, errReq := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if errReq != nil {
		return nil, fmt.Errorf("minimax: create quota request: %w", errReq)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", matrixUserAgent)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if trimmed := strings.TrimSpace(opGroupID); trimmed != "" {
		req.Header.Set("X-Group-Id", trimmed)
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("minimax: quota request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("minimax: close quota response body: %v", errClose)
		}
	}()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, fmt.Errorf("minimax: read quota response: %w", errRead)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("minimax: quota request failed with status %d", resp.StatusCode)
	}
	var payload map[string]any
	if errDecode := json.Unmarshal(raw, &payload); errDecode != nil {
		return nil, fmt.Errorf("minimax: decode quota response: %w", errDecode)
	}
	if base := asMap(payload["base_resp"]); base != nil {
		if code, ok := base["status_code"].(float64); ok && code != 0 {
			return nil, fmt.Errorf("minimax: quota rejected with status %v", code)
		}
	}
	remains, _ := payload["model_remains"].([]any)
	if len(remains) == 0 {
		return nil, nil
	}
	primary := asMap(remains[0])
	if primary == nil {
		return nil, nil
	}
	snapshot := &QuotaSnapshot{
		FiveHour: readQuotaWindow(primary, "interval"),
		Weekly:   readQuotaWindow(primary, "weekly"),
	}
	for _, entry := range remains {
		item := asMap(entry)
		if item == nil {
			continue
		}
		name := readMapString(item, "model_name")
		if strings.Contains(strings.ToLower(name), "video") {
			if total, ok := finiteNumber(item["current_interval_total_count"]); !ok || total <= 0 {
				continue
			}
			snapshot.Video = readVideoQuota(item)
			break
		}
	}
	return snapshot, nil
}

// signedGet performs a signed matrix GET.
func (c *Client) signedGet(ctx context.Context, origin, path, accessToken, realUserID string, nowMs int64) (any, error) {
	req, errReq := c.buildSignedMatrixRequest(ctx, origin, http.MethodGet, path, accessToken, realUserID, nowMs, "")
	if errReq != nil {
		return nil, errReq
	}
	return c.doMatrixJSON(req)
}

// postMatrixJSON performs a signed matrix POST with a JSON body.
func (c *Client) postMatrixJSON(ctx context.Context, origin, path, body, accessToken, realUserID string, nowMs int64) (map[string]any, error) {
	req, errReq := c.buildSignedMatrixRequest(ctx, origin, http.MethodPost, path, accessToken, realUserID, nowMs, body)
	if errReq != nil {
		return nil, errReq
	}
	payload, errDo := c.doMatrixJSON(req)
	if errDo != nil {
		return nil, errDo
	}
	result, _ := payload.(map[string]any)
	if result == nil {
		return nil, fmt.Errorf("minimax: matrix response is not an object")
	}
	return result, nil
}

func (c *Client) buildSignedMatrixRequest(ctx context.Context, origin, method, path, accessToken, realUserID string, nowMs int64, body string) (*http.Request, error) {
	query := matrixQuery(realUserID, nowMs)
	pathWithSearch := path + "?" + query
	yyBody := "{}"
	signatureBody := ""
	if body != "" {
		yyBody = body
		signatureBody = body
	}
	second := nowMs / 1000
	req, errReq := http.NewRequestWithContext(ctx, method, origin+pathWithSearch, strings.NewReader(body))
	if errReq != nil {
		return nil, fmt.Errorf("minimax: create %s request: %w", path, errReq)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", matrixUserAgent)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("yy", md5Hex(encodeURIComponent(pathWithSearch)+"_"+yyBody+md5Hex(fmt.Sprintf("%d", nowMs))+signatureSuffix))
	req.Header.Set("x-timestamp", fmt.Sprintf("%d", second))
	req.Header.Set("x-signature", md5Hex(fmt.Sprintf("%d%s%s", second, signatureSalt, signatureBody)))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) doMatrixJSON(req *http.Request) (any, error) {
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("minimax: matrix request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("minimax: close matrix response body: %v", errClose)
		}
	}()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, fmt.Errorf("minimax: read matrix response: %w", errRead)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &AuthError{StatusCode: resp.StatusCode}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("minimax: matrix request failed with status %d", resp.StatusCode)
	}
	var payload any
	if errDecode := json.Unmarshal(raw, &payload); errDecode != nil {
		return nil, fmt.Errorf("minimax: decode matrix response: %w", errDecode)
	}
	return payload, nil
}

// AuthError marks an upstream rejection that a token refresh may resolve.
type AuthError struct {
	StatusCode int
}

func (e *AuthError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("minimax: upstream rejected the credential (status %d)", e.StatusCode)
}

// IsAuthError reports whether err is an upstream authentication rejection.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// matrixOrigin returns the matrix API origin for the client's region, honoring
// a test override when set.
func (c *Client) matrixOrigin() string {
	if trimmed := strings.TrimSpace(c.matrixOriginOverride); trimmed != "" {
		return strings.TrimRight(trimmed, "/")
	}
	if c.region == RegionCN {
		return "https://agent.minimaxi.com"
	}
	return "https://agent.minimax.io"
}

// openPlatformOrigin returns the open-platform origin, honoring a test override.
func (c *Client) openPlatformOrigin() string {
	if trimmed := strings.TrimSpace(c.openPlatformOriginOverride); trimmed != "" {
		return strings.TrimRight(trimmed, "/")
	}
	return c.region.OpenPlatformOrigin()
}

// matrixQuery builds the fixed-order query string the signed requests carry.
// Order matters because the signature covers the encoded path+query.
func matrixQuery(realUserID string, nowMs int64) string {
	language := "en"
	if realUserID == "" {
		realUserID = "0"
	}
	params := [][2]string{
		{"device_platform", "mcode"},
		{"biz_id", matrixBizID},
		{"app_id", matrixAppID},
		{"version_code", matrixVersionCode},
		{"unix", fmt.Sprintf("%d", nowMs)},
		{"timezone_offset", fmt.Sprintf("%d", timezoneOffsetSeconds())},
		{"sys_language", language},
		{"lang", language},
		{"device_id", "0"},
		{"os_name", platformName()},
		{"browser_name", "mcode"},
		{"user_id", strings.TrimSpace(realUserID)},
		{"client", "mcode"},
	}
	parts := make([]string, 0, len(params))
	for _, kv := range params {
		parts = append(parts, kv[0]+"="+kv[1])
	}
	return strings.Join(parts, "&")
}

func timezoneOffsetSeconds() int {
	_, offset := time.Now().Zone()
	return offset
}

func platformName() string {
	switch runtime.GOOS {
	case "windows":
		return "win32"
	case "darwin":
		return "darwin"
	default:
		return runtime.GOOS
	}
}

// readQuotaWindow reads one interval/weekly window from a model_remains entry.
func readQuotaWindow(value map[string]any, kind string) QuotaWindow {
	prefix := "current_interval"
	if kind == "weekly" {
		prefix = "current_weekly"
	}
	status, statusOK := finiteNumber(value[prefix+"_status"])
	unlimited := statusOK && status == 3
	remaining, remainingOK := finiteNumber(value[prefix+"_remaining_percent"])
	total, totalOK := finiteNumber(value[prefix+"_total_count"])
	legacyRemaining, legacyOK := finiteNumber(value[prefix+"_usage_count"])
	var percent *int
	if !remainingOK && totalOK && total > 0 && legacyOK {
		calculated := (legacyRemaining / total) * 100
		remaining = calculated
		remainingOK = true
	}
	if !unlimited && remainingOK {
		clamped := int(remaining + roundingBias(remaining))
		if clamped < 0 {
			clamped = 0
		}
		if clamped > 100 {
			clamped = 100
		}
		percent = &clamped
	}
	resetKey := "end_time"
	if kind == "weekly" {
		resetKey = "weekly_end_time"
	}
	resetAtMs := readInt64(value[resetKey])
	return QuotaWindow{RemainingPercent: percent, ResetAtMs: resetAtMs, Unlimited: unlimited}
}

// roundingBias reproduces JavaScript Math.round for non-negative values.
func roundingBias(value float64) float64 {
	if value < 0 {
		return -0.5
	}
	return 0.5
}

func readVideoQuota(value map[string]any) *QuotaVideo {
	status, _ := finiteNumber(value["current_interval_status"])
	video := &QuotaVideo{Unlimited: status == 3}
	if total, ok := finiteNumber(value["current_interval_total_count"]); ok {
		clamped := int(total)
		if clamped < 0 {
			clamped = 0
		}
		video.TotalCount = &clamped
	}
	if remaining, ok := finiteNumber(value["current_interval_usage_count"]); ok {
		clamped := int(remaining)
		if clamped < 0 {
			clamped = 0
		}
		video.RemainingCount = &clamped
	}
	video.ResetAtMs = readInt64(value["end_time"])
	return video
}

// projectMembership mirrors the client's projectMembership.
func projectMembership(body map[string]any) Membership {
	data := asMap(body["data"])
	membership := Membership{}
	if has, ok := readBoolAlt(body, data, "has_token_plan"); ok {
		membership.HasTokenPlan = &has
	}
	membership.OpGroupID = firstNonEmpty(readMapStringAlt(body, data, "op_group_id"))
	membership.Tier = readMapStringAlt(body, data, "token_plan_tier")
	membership.ExpiresAtMs = readPositiveInt64Alt(body, data, "token_plan_expires_at")
	creditSummary := asMap(body["op_credit_summary"])
	if creditSummary == nil {
		creditSummary = asMap(data["op_credit_summary"])
	}
	if creditSummary != nil {
		membership.CreditBalance = readMapString(creditSummary, "total_remaining_amount")
	}
	if membership.CreditBalance == "" {
		membership.CreditBalance = readNumberishStringAlt(body, data, "opcredit_balance")
	}
	return membership
}

// projectPersonalWorkspace mirrors the client's projectPersonalWorkspace,
// returning the personal workspace id and its membership.
func projectPersonalWorkspace(body map[string]any) (string, Membership, bool) {
	data := asMap(body["data"])
	workspaces, _ := body["workspaces"].([]any)
	if workspaces == nil {
		workspaces, _ = data["workspaces"].([]any)
	}
	for _, raw := range workspaces {
		workspace := asMap(raw)
		if workspace == nil {
			continue
		}
		if workspaceType, ok := finiteNumber(workspace["workspace_type"]); !ok || workspaceType != 0 {
			continue
		}
		workspaceID, ok := readWorkspaceID(workspace["workspace_id"])
		if !ok {
			continue
		}
		return workspaceID, projectMembership(workspace), true
	}
	return "", Membership{}, false
}

// mergeMembership mirrors the client's mergeMembership: personal and scoped
// readings are unioned, preferring a definite has_token_plan and the first
// non-empty summary field.
func mergeMembership(personal, scoped Membership) Membership {
	merged := Membership{}
	switch {
	case personal.HasTokenPlan != nil && *personal.HasTokenPlan, scoped.HasTokenPlan != nil && *scoped.HasTokenPlan:
		value := true
		merged.HasTokenPlan = &value
	case personal.HasTokenPlan != nil && !*personal.HasTokenPlan, scoped.HasTokenPlan != nil && !*scoped.HasTokenPlan:
		value := false
		merged.HasTokenPlan = &value
	}
	merged.OpGroupID = firstNonEmpty(personal.OpGroupID, scoped.OpGroupID)
	merged.Tier = firstNonEmpty(personal.Tier, scoped.Tier)
	merged.CreditBalance = firstNonEmpty(personal.CreditBalance, scoped.CreditBalance)
	if personal.ExpiresAtMs > 0 {
		merged.ExpiresAtMs = personal.ExpiresAtMs
	} else {
		merged.ExpiresAtMs = scoped.ExpiresAtMs
	}
	return merged
}

func readWorkspaceID(value any) (string, bool) {
	switch typed := value.(type) {
	case float64:
		if typed >= 0 {
			return fmt.Sprintf("%d", int64(typed)), true
		}
	case string:
		if trimmed := strings.TrimSpace(typed); trimmed != "" {
			return trimmed, true
		}
	}
	return "", false
}

// md5Hex returns the lowercase hex MD5 digest. It reproduces the client's
// request-signing wire format; it is an integrity marker, not a security
// primitive.
func md5Hex(value string) string {
	// MD5 is mandated by the MiniMax matrix request-signing protocol: the gateway
	// recomputes and compares the same digest, so a stronger primitive cannot be
	// substituted.
	// nosemgrep: go.lang.security.audit.crypto.use_of_weak_crypto.use-of-md5
	sum := md5.Sum([]byte(value)) //nolint:gosec // #nosec G401 -- protocol-defined MiniMax signature requires MD5.
	return hex.EncodeToString(sum[:])
}

// encodeURIComponent reproduces JavaScript's encodeURIComponent: everything
// except the unreserved set is percent-encoded byte-by-byte from UTF-8.
func encodeURIComponent(value string) string {
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.!~*'()"
	var builder strings.Builder
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if strings.IndexByte(unreserved, ch) >= 0 {
			builder.WriteByte(ch)
			continue
		}
		builder.WriteString(fmt.Sprintf("%%%02X", ch))
	}
	return builder.String()
}

// finiteNumber returns a finite JSON number.
func finiteNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func readInt64(value any) int64 {
	if number, ok := finiteNumber(value); ok && number > 0 {
		return int64(number)
	}
	return 0
}

func asMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	result, _ := value.(map[string]any)
	return result
}

func firstMap(candidates ...map[string]any) map[string]any {
	for _, candidate := range candidates {
		if candidate != nil {
			return candidate
		}
	}
	return nil
}

func readMapString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	if value, ok := payload[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func readMapStringAlt(primary, fallback map[string]any, key string) string {
	if value := readMapString(primary, key); value != "" {
		return value
	}
	return readMapString(fallback, key)
}

func readNumberishStringAlt(primary, fallback map[string]any, key string) string {
	for _, payload := range []map[string]any{primary, fallback} {
		if payload == nil {
			continue
		}
		switch value := payload[key].(type) {
		case string:
			return strings.TrimSpace(value)
		case float64:
			return fmt.Sprintf("%v", value)
		}
	}
	return ""
}

func readBoolAlt(primary, fallback map[string]any, key string) (bool, bool) {
	for _, payload := range []map[string]any{primary, fallback} {
		if payload == nil {
			continue
		}
		if value, ok := payload[key].(bool); ok {
			return value, true
		}
	}
	return false, false
}

func readPositiveInt64Alt(primary, fallback map[string]any, key string) int64 {
	for _, payload := range []map[string]any{primary, fallback} {
		if payload == nil {
			continue
		}
		if value, ok := finiteNumber(payload[key]); ok && value > 0 {
			return int64(value)
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// NewQuotaClient builds a client for quota reads, reusing the auth client's
// proxy-aware HTTP transport.
func NewQuotaClient(cfg *config.Config, region Region, proxyURL string) *Client {
	return NewClientWithProxyURL(cfg, region, proxyURL)
}
