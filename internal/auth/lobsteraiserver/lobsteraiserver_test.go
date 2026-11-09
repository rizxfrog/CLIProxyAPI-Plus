package lobsteraiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

func TestBuildLoginURLUsesPortalSurfaceAndHashQuery(t *testing.T) {
	// A backend override that is neither test nor production keeps the lookup
	// away from the live config endpoint, so the portal fallback is exercised.
	client := NewClientWithProxyURL(nil, "", "https://lobsterai-server.youdao.com")
	got := client.BuildLoginURL("state-123", "http://127.0.0.1:9999/auth/callback")
	if !strings.HasPrefix(got, "https://lobsterai.youdao.com/portal#/login?") {
		t.Fatalf("login URL = %q", got)
	}
	// The parameters belong to the hash query, not the regular query.
	if !strings.Contains(got, "state=state-123") {
		t.Fatalf("login URL missing state: %q", got)
	}
	if !strings.Contains(got, "redirect_uri=http%3A%2F%2F127.0.0.1%3A9999%2Fauth%2Fcallback") {
		t.Fatalf("login URL missing encoded redirect_uri: %q", got)
	}
	// The portal only issues the authorization-code redirect for a recognized
	// source; "electron" is the value the desktop client sends.
	if !strings.Contains(got, "source=electron") {
		t.Fatalf("login URL missing electron source: %q", got)
	}
	if strings.Contains(got, "?state=state-123") {
		t.Fatalf("parameters must live in the hash query, not the regular query: %q", got)
	}
}

func TestBuildLoginURLDefaultsRedirectURIWhenCallerHasNoListener(t *testing.T) {
	got := appendLoginParams("https://example.com/login", map[string]string{"state": "abc", "empty": "  "})
	if got != "https://example.com/login?state=abc" {
		t.Fatalf("appendLoginParams = %q", got)
	}
}

func TestIsTestBackend(t *testing.T) {
	if !NewClientWithProxyURL(nil, "", DefaultTestBackendURL).isTestBackend() {
		t.Fatal("inner backend must be detected as test")
	}
	if NewClientWithProxyURL(nil, "", DefaultBackendURL).isTestBackend() {
		t.Fatal("production backend must not be detected as test")
	}
}

func TestExchangeCodeParsesEnvelopeAndKeyfrom(t *testing.T) {
	var sawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/exchange" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "ok",
			"data": map[string]any{
				"accessToken":  "acc-1",
				"refreshToken": "ref-1",
				"expiresIn":    3600,
				"user":         map[string]any{"id": "u1", "email": "a@b.c", "name": "Ann"},
				"quota":        map[string]any{"subscriptionStatus": "active", "credits": 42},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(&config.Config{}, "", server.URL)
	client.SetKeyfromIdentity("official", "official", "2026.10.13", "uuid-1")
	token, user, quota, errExchange := client.ExchangeCode(context.Background(), "code-1", "state-1")
	if errExchange != nil {
		t.Fatalf("ExchangeCode: %v", errExchange)
	}
	if token.AccessToken != "acc-1" || token.RefreshToken != "ref-1" || token.ExpiresIn != 3600 {
		t.Fatalf("token = %+v", token)
	}
	if token.ExpiresAt.IsZero() {
		t.Fatal("ExpiresAt not derived from expiresIn")
	}
	if user == nil || user.ID != "u1" || user.Email != "a@b.c" {
		t.Fatalf("user = %+v", user)
	}
	if quota == nil || quota.SubscriptionStatus != "active" || !quota.CreditsKnown || quota.Credits != 42 {
		t.Fatalf("quota = %+v", quota)
	}
	if sawBody["authCode"] != "code-1" || sawBody["state"] != "state-1" {
		t.Fatalf("exchange body = %+v", sawBody)
	}
	// The upstream requires the attribution payload on login requests.
	if sawBody["firstKeyfrom"] != "official" || sawBody["latestKeyfrom"] != "official" {
		t.Fatalf("exchange body missing keyfrom: %+v", sawBody)
	}
	if sawBody["version"] != "2026.10.13" || sawBody["uuid"] != "uuid-1" {
		t.Fatalf("exchange body missing version/uuid: %+v", sawBody)
	}
}

func TestExchangeCodeErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	if _, _, _, errExchange := client.ExchangeCode(context.Background(), "bad", "state"); errExchange == nil {
		t.Fatal("expected error for non-2xx exchange")
	}
}

func TestExchangeCodeSurfacesNonZeroEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 40100, "message": "auth code expired"})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	_, _, _, errExchange := client.ExchangeCode(context.Background(), "expired", "state")
	if errExchange == nil || !strings.Contains(errExchange.Error(), "40100") {
		t.Fatalf("expected backend error code surfaced, got %v", errExchange)
	}
}

func TestRefreshPreservesTokenWhenOmitted(t *testing.T) {
	var sawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/refresh" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "ok",
			"data": map[string]any{"accessToken": "acc-2", "expiresIn": 1800},
		})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	token, errRefresh := client.Refresh(context.Background(), "ref-old")
	if errRefresh != nil {
		t.Fatalf("Refresh: %v", errRefresh)
	}
	if token.AccessToken != "acc-2" {
		t.Fatalf("access token = %q", token.AccessToken)
	}
	if token.RefreshToken != "ref-old" {
		t.Fatalf("refresh token = %q, want the preserved previous token", token.RefreshToken)
	}
	if sawBody["refreshToken"] != "ref-old" {
		t.Fatalf("refresh body = %+v", sawBody)
	}
	if sawBody["version"] == nil {
		t.Fatalf("refresh body missing keyfrom version: %+v", sawBody)
	}
}

func TestFetchModelsFiltersInaccessibleAndParsesIdentity(t *testing.T) {
	var sawAuth, sawCaps, sawVersion string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/models/available" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization")
		sawCaps = r.Header.Get("X-LobsterAI-Client-Capabilities")
		sawVersion = r.Header.Get("X-LobsterAI-Client-Version")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": []any{
				map[string]any{"modelId": "deepseek-v4-flash", "modelName": "DeepSeek V4 Flash", "apiFormat": "openai", "supportsThinking": true, "accessible": true},
				map[string]any{"modelId": "deepseek-v4-pro", "modelName": "DeepSeek V4 Pro", "apiFormat": "openai", "supportsImage": true, "contextWindow": 200000, "accessible": true},
				map[string]any{"modelId": "kimi-k3", "apiFormat": "openai", "accessible": false},
				map[string]any{"modelId": "deepseek-v4-flash", "apiFormat": "openai", "accessible": true},
			},
		})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	models, errModels := client.FetchModels(context.Background(), "acc-1")
	if errModels != nil {
		t.Fatalf("FetchModels: %v", errModels)
	}
	// The inaccessible kimi-k3 entry is dropped, and the duplicate collapses.
	if len(models) != 2 {
		t.Fatalf("models = %+v, want 2 after entitlement filter and dedupe", models)
	}
	if models[0].ID != "deepseek-v4-flash" || models[0].DisplayName != "DeepSeek V4 Flash" || !models[0].Thinking {
		t.Fatalf("models[0] = %+v", models[0])
	}
	if models[1].ID != "deepseek-v4-pro" || models[1].APIFormat != "openai" || !models[1].Image || models[1].ContextLimit != 200000 {
		t.Fatalf("models[1] = %+v", models[1])
	}
	if sawAuth != "Bearer acc-1" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if !strings.Contains(sawCaps, "kimi-k3-agentic-v1") {
		t.Fatalf("capabilities header = %q", sawCaps)
	}
	if sawVersion == "" {
		t.Fatal("client version header missing")
	}
}

func TestFetchQuotaHonestAboutUnknownCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"subscriptionStatus": "free", "planName": "免费"},
		})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	quota, errQuota := client.FetchQuota(context.Background(), "acc-1")
	if errQuota != nil {
		t.Fatalf("FetchQuota: %v", errQuota)
	}
	if quota.SubscriptionStatus != "free" {
		t.Fatalf("subscription = %q", quota.SubscriptionStatus)
	}
	if quota.CreditsKnown {
		t.Fatalf("credits must stay unknown when the backend omits them: %+v", quota)
	}
}

func TestFetchActivitySnapshotReadsSlotAndContext(t *testing.T) {
	var sawPlacement, sawContainerVersion, sawPlatform string
	var contextCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/client-activities/slot":
			sawPlacement = r.URL.Query().Get("placement")
			sawContainerVersion = r.URL.Query().Get("containerApiVersion")
			sawPlatform = r.URL.Query().Get("platform")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"slotState": "available",
					"activity":  map[string]any{"activityCode": "daily_check_in", "configRevision": 7},
				},
			})
		case r.URL.Path == "/api/client-activities/daily_check_in/context":
			contextCalled = true
			if got := r.URL.Query().Get("configRevision"); got != "7" {
				t.Errorf("configRevision = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{"dailyCheckInRewardPerDay": 20, "dailyCheckInProgress": "1/7"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	snapshot, errSnapshot := client.FetchActivitySnapshot(context.Background(), "acc-1", ActivityPlacementDesktopSidebar, ActivityContainerAPIVersionDailyCheckIn)
	if errSnapshot != nil {
		t.Fatalf("FetchActivitySnapshot: %v", errSnapshot)
	}
	if !contextCalled {
		t.Fatal("activity context was not read")
	}
	if snapshot.SlotState != "available" || snapshot.ActivityCode != "daily_check_in" || snapshot.ConfigRevision != 7 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if snapshot.Context["dailyCheckInRewardPerDay"] != float64(20) {
		t.Fatalf("context = %+v", snapshot.Context)
	}
	if sawPlacement != ActivityPlacementDesktopSidebar {
		t.Fatalf("placement = %q", sawPlacement)
	}
	if sawContainerVersion != "2" {
		t.Fatalf("containerApiVersion = %q", sawContainerVersion)
	}
	// The backend keys activity availability on a real electron platform token;
	// an unrecognized value (e.g. "cli") silently yields an empty slot.
	if sawPlatform == "" || sawPlatform == "cli" {
		t.Fatalf("platform = %q, want a desktop platform token", sawPlatform)
	}
}

func TestFetchActivitySnapshotSkipsEmptySlot(t *testing.T) {
	var contextCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/context") {
			contextCalled = true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"slotState": "empty"}})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	snapshot, errSnapshot := client.FetchActivitySnapshot(context.Background(), "acc-1", ActivityPlacementDesktopSidebar, 0)
	if errSnapshot != nil {
		t.Fatalf("FetchActivitySnapshot: %v", errSnapshot)
	}
	if snapshot.SlotState != "empty" || snapshot.ActivityCode != "" {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if contextCalled {
		t.Fatal("context must not be read for an empty slot")
	}
}

func TestFetchQuotaMapsLiveCreditFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"subscriptionStatus":   "free",
				"planName":             "免费",
				"freeCreditsTotal":     300,
				"freeCreditsRemaining": 299.5,
				"freeCreditsUsed":      0.5,
			},
		})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	quota, errQuota := client.FetchQuota(context.Background(), "acc-1")
	if errQuota != nil {
		t.Fatalf("FetchQuota: %v", errQuota)
	}
	if quota.SubscriptionStatus != "free" {
		t.Fatalf("subscription = %q", quota.SubscriptionStatus)
	}
	if !quota.CreditsKnown || quota.Credits != 299 {
		t.Fatalf("credits = %d known=%v, want 299/true", quota.Credits, quota.CreditsKnown)
	}
}

func TestParseUserAcceptsNumericIDAndYid(t *testing.T) {
	// /api/user/profile reports the id as a NUMBER and the login identifier as
	// "yid"; a string-only reader silently dropped the whole identity.
	user := parseUser(map[string]any{
		"yid":      "urs-phoneyd.bc69332f76b2498aa@163.com",
		"nickname": "130****2189",
		"id":       float64(123318),
	})
	if user.ID != "123318" {
		t.Fatalf("id = %q, want 123318", user.ID)
	}
	if user.Email != "urs-phoneyd.bc69332f76b2498aa@163.com" {
		t.Fatalf("email = %q", user.Email)
	}
	if user.Name != "130****2189" {
		t.Fatalf("name = %q", user.Name)
	}
}

func TestFetchCreditSummaryParsesTotalAndPackages(t *testing.T) {
	var sawPath, sawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath, sawQuery = r.URL.Path, r.URL.RawQuery
		if got := r.Header.Get("Authorization"); got != "Bearer acc-1" {
			t.Errorf("authorization = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "success",
			"data": map[string]any{
				"id":                    float64(123318),
				"nickname":              "130****2189",
				"totalCreditsRemaining": 5297.72,
				"creditItems": []any{
					map[string]any{"type": "free", "label": "免费额度", "labelEn": "Free", "creditsRemaining": 299.86, "expiresAt": "2026-10-29"},
					map[string]any{"type": "bonus", "label": "每日登录奖励", "labelEn": "Daily login", "creditsRemaining": 100.0, "expiresAt": nil},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(&config.Config{}, "", server.URL)
	client.SetKeyfromIdentity("official", "official", "2026.10.13", "")
	summary, errSummary := client.FetchCreditSummary(context.Background(), "acc-1")
	if errSummary != nil {
		t.Fatalf("FetchCreditSummary: %v", errSummary)
	}
	if sawPath != "/api/user/profile-summary" {
		t.Fatalf("path = %q", sawPath)
	}
	// The account-menu read carries the same attribution identity as the other
	// account reads.
	if !strings.Contains(sawQuery, "firstKeyfrom=official") || !strings.Contains(sawQuery, "version=2026.10.13") {
		t.Fatalf("keyfrom query = %q", sawQuery)
	}
	if !summary.TotalKnown || summary.TotalCreditsRemaining != 5297.72 {
		t.Fatalf("total = %v known=%v", summary.TotalCreditsRemaining, summary.TotalKnown)
	}
	if len(summary.Packages) != 2 {
		t.Fatalf("packages = %+v", summary.Packages)
	}
	if summary.Packages[1].Type != "bonus" || summary.Packages[1].Label != "每日登录奖励" || summary.Packages[1].CreditsRemaining != 100 {
		t.Fatalf("bonus package = %+v", summary.Packages[1])
	}
	if summary.UserID != "123318" {
		t.Fatalf("user id = %q", summary.UserID)
	}
}

func TestFetchCreditSummaryKeepsAnAbsentTotalUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "success", "data": map[string]any{}})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	summary, errSummary := client.FetchCreditSummary(context.Background(), "acc-1")
	if errSummary != nil {
		t.Fatalf("FetchCreditSummary: %v", errSummary)
	}
	// Absent must not become a zero balance: the handler reports it as unknown.
	if summary.TotalKnown || summary.TotalCreditsRemaining != 0 {
		t.Fatalf("total = %v known=%v", summary.TotalCreditsRemaining, summary.TotalKnown)
	}
	if len(summary.Packages) != 0 {
		t.Fatalf("packages = %+v", summary.Packages)
	}
}

func TestFetchCreditSummaryRequiresAToken(t *testing.T) {
	client := NewClientWithProxyURL(nil, "", DefaultBackendURL)
	if _, errSummary := client.FetchCreditSummary(context.Background(), "  "); errSummary == nil {
		t.Fatal("expected an error without an access token")
	}
}

func TestClaimActivityPostsTheCheckInAndParsesTheGrant(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod, sawPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "success",
			"data": map[string]any{"result": map[string]any{
				"creditsGranted": 100.0, "periodKey": "2026-10-15", "message": "签到成功",
			}},
		})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(&config.Config{}, "", server.URL)
	claim, errClaim := client.ClaimActivity(context.Background(), "acc-1", ActivityClaimRequest{
		ActivityCode:   "daily-check-in-evergreen-prod-20260814",
		ActionID:       ActivityActionCheckIn,
		ConfigRevision: 7,
	})
	if errClaim != nil {
		t.Fatalf("ClaimActivity: %v", errClaim)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("method = %q", sawMethod)
	}
	want := "/api/client-activities/daily-check-in-evergreen-prod-20260814/actions/check_in"
	if sawPath != want {
		t.Fatalf("path = %q, want %q", sawPath, want)
	}
	if sawBody["configRevision"] != float64(7) {
		t.Fatalf("configRevision = %v", sawBody["configRevision"])
	}
	if key, _ := sawBody["idempotencyKey"].(string); key == "" {
		t.Fatalf("idempotencyKey missing: %v", sawBody)
	}
	if _, ok := sawBody["payload"].(map[string]any); !ok {
		t.Fatalf("empty payload object required: %v", sawBody)
	}
	if !claim.GrantedKnown || claim.Granted != 100 {
		t.Fatalf("granted = %v known=%v", claim.Granted, claim.GrantedKnown)
	}
	if claim.PeriodKey != "2026-10-15" {
		t.Fatalf("periodKey = %q", claim.PeriodKey)
	}
}

func TestClaimActivityMapsTheDuplicateClaimCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": ActivityAlreadyClaimedCode, "message": "已领取"})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	_, errClaim := client.ClaimActivity(context.Background(), "acc-1", ActivityClaimRequest{
		ActivityCode: "daily-check-in", ActionID: ActivityActionCheckIn, ConfigRevision: 1,
	})
	if !errors.Is(errClaim, ErrActivityAlreadyClaimed) {
		t.Fatalf("err = %v, want ErrActivityAlreadyClaimed", errClaim)
	}
}

func TestClaimActivityRejectsAResponseWithoutABusinessCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"unexpected":true}`))
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	if _, errClaim := client.ClaimActivity(context.Background(), "acc-1", ActivityClaimRequest{
		ActivityCode: "daily-check-in", ActionID: ActivityActionCheckIn, ConfigRevision: 1,
	}); errClaim == nil {
		t.Fatal("a response without a business code must not read as success")
	}
}

func TestClaimActivityRequiresAnActivityCode(t *testing.T) {
	client := NewClientWithProxyURL(nil, "", DefaultBackendURL)
	if _, errClaim := client.ClaimActivity(context.Background(), "acc-1", ActivityClaimRequest{ActionID: ActivityActionCheckIn}); errClaim == nil {
		t.Fatal("expected an error without an activity code")
	}
}

func TestActivitySnapshotExposesTheClaimState(t *testing.T) {
	snapshot := &ActivitySnapshot{
		SlotState:      "available",
		ActivityCode:   "daily-check-in",
		ConfigRevision: 7,
		Context: map[string]any{
			"state": map[string]any{
				"claimedToday":   true,
				"completed":      false,
				"claimedDays":    float64(3),
				"claimedCredits": 300.0,
				"rewardCredits":  100.0,
			},
			"actions": []any{"check_in"},
		},
	}
	if !snapshot.Available() {
		t.Fatal("an open slot with a revision must be available")
	}
	if !snapshot.ClaimedToday() || snapshot.Completed() {
		t.Fatalf("claim state = today:%v completed:%v", snapshot.ClaimedToday(), snapshot.Completed())
	}
	if snapshot.ClaimedDays() != 3 || snapshot.ClaimedCredits() != 300 || snapshot.RewardCredits() != 100 {
		t.Fatalf("counters = days:%d credits:%v reward:%v", snapshot.ClaimedDays(), snapshot.ClaimedCredits(), snapshot.RewardCredits())
	}
	if actions := snapshot.Actions(); len(actions) != 1 || actions[0] != ActivityActionCheckIn {
		t.Fatalf("actions = %+v", actions)
	}
	// A slot without a config revision cannot be read any further, so it must
	// not be reported as claimable.
	if (&ActivitySnapshot{SlotState: "available", ActivityCode: "x"}).Available() {
		t.Fatal("a zero config revision must not be claimable")
	}
}

func TestActivityPlatformForMapsOnlyElectronHosts(t *testing.T) {
	// The backend only recognizes the desktop client's own process.platform
	// values; anything else makes it answer an empty slot.
	for goos, want := range map[string]string{
		"windows": "win32",
		"darwin":  "darwin",
		"linux":   "linux",
		"Windows": "win32",
	} {
		if got := ActivityPlatformFor(goos); got != want {
			t.Fatalf("ActivityPlatformFor(%q) = %q, want %q", goos, got, want)
		}
	}
	// An unrecognized host must not be reported as a platform the client could
	// have sent.
	for _, goos := range []string{"freebsd", "plan9", "", "  "} {
		if got := ActivityPlatformFor(goos); got != "" {
			t.Fatalf("ActivityPlatformFor(%q) = %q, want empty", goos, got)
		}
	}
}

func TestActivitySlotSendsTheResolvedPlatform(t *testing.T) {
	var platforms []string
	var enterpriseMode, enterpriseID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/client-activities/slot") {
			http.NotFound(w, r)
			return
		}
		platforms = append(platforms, r.URL.Query().Get("platform"))
		enterpriseMode = r.Header.Get(HeaderAccountMode)
		enterpriseID = r.Header.Get(HeaderEnterpriseID)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"slotState": "empty"}})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(&config.Config{}, "", server.URL)
	if _, errSlot := client.FetchActivitySnapshot(context.Background(), "acc-1", "", 0); errSlot != nil {
		t.Fatalf("FetchActivitySnapshot: %v", errSlot)
	}
	if got := platforms[0]; got != ActivityPlatformFor(runtime.GOOS) {
		t.Fatalf("platform = %q, want the host mapping", got)
	}
	// A personal account sends neither enterprise header.
	if enterpriseMode != "" || enterpriseID != "" {
		t.Fatalf("enterprise headers leaked: %q %q", enterpriseMode, enterpriseID)
	}

	// An explicit override wins over the host, and an enterprise context adds the
	// two headers to the activity read as well.
	client.SetActivityPlatform("darwin")
	client.SetEnterpriseContext(AccountModeEnterprise, "1001")
	if _, errSlot := client.FetchActivitySnapshot(context.Background(), "acc-1", "", 0); errSlot != nil {
		t.Fatalf("FetchActivitySnapshot: %v", errSlot)
	}
	if got := platforms[1]; got != "darwin" {
		t.Fatalf("overridden platform = %q", got)
	}
	if enterpriseMode != AccountModeEnterprise || enterpriseID != "1001" {
		t.Fatalf("enterprise headers = %q %q", enterpriseMode, enterpriseID)
	}

	// Clearing the override restores the host mapping.
	client.SetActivityPlatform("  ")
	if _, errSlot := client.FetchActivitySnapshot(context.Background(), "acc-1", "", 0); errSlot != nil {
		t.Fatalf("FetchActivitySnapshot: %v", errSlot)
	}
	if got := platforms[2]; got != ActivityPlatformFor(runtime.GOOS) {
		t.Fatalf("platform after clearing = %q", got)
	}
}

func TestSetEnterpriseContextRejectsAnIncompleteIdentity(t *testing.T) {
	newClient := func() *Client { return NewClientWithProxyURL(nil, "", DefaultBackendURL) }
	for _, tt := range []struct {
		name string
		mode string
		id   string
	}{
		{"personal mode", AccountModePersonal, "1001"},
		{"empty mode", "", "1001"},
		{"enterprise without id", AccountModeEnterprise, "  "},
	} {
		client := newClient()
		client.SetEnterpriseContext(tt.mode, tt.id)
		if client.accountMode != "" || client.enterpriseID != "" {
			t.Fatalf("%s: context = %q %q, want empty", tt.name, client.accountMode, client.enterpriseID)
		}
	}
	client := newClient()
	client.SetEnterpriseContext(AccountModeEnterprise, "1001")
	client.SetEnterpriseContext(AccountModePersonal, "")
	if client.accountMode != "" || client.enterpriseID != "" {
		t.Fatal("a later personal context must clear the previous enterprise one")
	}
}

func TestEnterpriseHeadersAreSentOnAuthenticatedRequests(t *testing.T) {
	var mode, id string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode, id = r.Header.Get(HeaderAccountMode), r.Header.Get(HeaderEnterpriseID)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"credits": 1}})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(nil, "", server.URL)
	client.SetEnterpriseContext(AccountModeEnterprise, "1001")
	if _, errQuota := client.FetchQuota(context.Background(), "acc-1"); errQuota != nil {
		t.Fatalf("FetchQuota: %v", errQuota)
	}
	if mode != AccountModeEnterprise || id != "1001" {
		t.Fatalf("enterprise headers = %q %q", mode, id)
	}
}

func TestFetchEnterpriseContextParsesTheIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/enterprise/context" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "success",
			"data": map[string]any{
				"accountMode":    AccountModeEnterprise,
				"enterpriseId":   float64(1001),
				"memberId":       float64(7),
				"enterpriseName": "Example Enterprise",
				"role":           "super_admin",
			},
		})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(&config.Config{}, "", server.URL)
	enterprise, errContext := client.FetchEnterpriseContext(context.Background(), "acc-1")
	if errContext != nil {
		t.Fatalf("FetchEnterpriseContext: %v", errContext)
	}
	if enterprise == nil {
		t.Fatal("expected an enterprise context")
	}
	if enterprise.EnterpriseID != "1001" || enterprise.MemberID != "7" || enterprise.Role != "super_admin" {
		t.Fatalf("context = %+v", enterprise)
	}
	if enterprise.EnterpriseName != "Example Enterprise" {
		t.Fatalf("enterprise name = %q", enterprise.EnterpriseName)
	}
}

func TestEnterpriseContextIsRejectedWhenIncomplete(t *testing.T) {
	// A half-populated payload must not produce a header pair that scopes requests
	// to an enterprise the account may not belong to.
	for name, data := range map[string]map[string]any{
		"personal mode":     {"accountMode": AccountModePersonal, "enterpriseId": float64(1001), "enterpriseName": "X", "role": "member"},
		"non-positive id":   {"enterpriseId": float64(0), "enterpriseName": "X", "role": "member"},
		"missing name":      {"enterpriseId": float64(1001), "role": "member"},
		"unknown role":      {"enterpriseId": float64(1001), "enterpriseName": "X", "role": "owner"},
		"empty payload":     {},
		"id as a big float": {"enterpriseId": 1.5, "enterpriseName": "X", "role": "member"},
	} {
		if got := parseEnterpriseContext(data); got != nil {
			t.Fatalf("%s: context = %+v, want nil", name, got)
		}
	}
	// A nested payload is accepted, as the backend nests it differently across
	// surfaces.
	nested := map[string]any{"data": map[string]any{"enterpriseContext": map[string]any{
		"enterpriseId": "1001", "enterpriseName": "X", "role": "member",
	}}}
	if got := parseEnterpriseContext(nested); got == nil || got.EnterpriseID != "1001" {
		t.Fatalf("nested context = %+v", got)
	}
}

func TestFetchEnterpriseContextDistinguishesTerminalFromTransient(t *testing.T) {
	status := 0
	code := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != 0 {
			w.WriteHeader(status)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": "nope"})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)

	// Terminal codes mean the account has no context any more: clear it, and treat
	// the answer as a normal "personal account" rather than a failure.
	for _, terminal := range []int{EnterpriseErrorNotFound, EnterpriseErrorNotMember, EnterpriseErrorAccountModeMismatch} {
		code, status = terminal, 200
		enterprise, errContext := client.FetchEnterpriseContext(context.Background(), "acc-1")
		if errContext != nil || enterprise != nil {
			t.Fatalf("code %d: context=%+v err=%v", terminal, enterprise, errContext)
		}
	}

	// Anything else must not erase a known context, so it is reported as an error.
	code, status = 41601, 200
	if _, errContext := client.FetchEnterpriseContext(context.Background(), "acc-1"); errContext == nil {
		t.Fatal("a transient business code must be reported as an error")
	}
	code, status = 0, http.StatusInternalServerError
	if _, errContext := client.FetchEnterpriseContext(context.Background(), "acc-1"); errContext == nil {
		t.Fatal("a 5xx must be reported as an error")
	}
	// A response without the envelope is a failure, not a personal account.
	code, status = 0, 200
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"unexpected":true}`))
	}))
	defer empty.Close()
	if _, errContext := NewClientWithProxyURL(nil, "", empty.URL).FetchEnterpriseContext(context.Background(), "acc-1"); errContext == nil {
		t.Fatal("a response without a business code must be an error")
	}
	if _, errContext := client.FetchEnterpriseContext(context.Background(), "  "); errContext == nil {
		t.Fatal("expected an error without an access token")
	}
}
