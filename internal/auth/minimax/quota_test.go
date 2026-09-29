package minimax

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func intPtr(value int) *int { return &value }

func TestRegionOpenPlatformOrigin(t *testing.T) {
	if got := RegionEN.OpenPlatformOrigin(); got != "https://platform.minimax.io" {
		t.Fatalf("EN open platform origin = %q", got)
	}
	if got := RegionCN.OpenPlatformOrigin(); got != "https://www.minimaxi.com" {
		t.Fatalf("CN open platform origin = %q", got)
	}
}

func TestReadQuotaWindowPercentAndReset(t *testing.T) {
	value := map[string]any{
		"current_interval_status":            float64(1),
		"current_interval_remaining_percent": float64(73.6),
		"end_time":                           float64(1900000000000),
	}
	window := readQuotaWindow(value, "interval")
	if window.RemainingPercent == nil || *window.RemainingPercent != 74 {
		t.Fatalf("percent = %v, want 74 (rounded)", window.RemainingPercent)
	}
	if window.ResetAtMs != 1900000000000 {
		t.Fatalf("reset = %d", window.ResetAtMs)
	}
	if window.Unlimited {
		t.Fatal("unlimited = true, want false")
	}
}

func TestReadQuotaWindowUnlimited(t *testing.T) {
	value := map[string]any{"current_weekly_status": float64(3)}
	window := readQuotaWindow(value, "weekly")
	if !window.Unlimited {
		t.Fatal("unlimited = false, want true")
	}
	if window.RemainingPercent != nil {
		t.Fatalf("percent = %v, want nil when unlimited", *window.RemainingPercent)
	}
	if window.ResetAtMs != 0 {
		t.Fatalf("reset = %d, want 0", window.ResetAtMs)
	}
}

func TestReadQuotaWindowLegacyUsageCount(t *testing.T) {
	value := map[string]any{
		"current_interval_status":      float64(1),
		"current_interval_total_count": float64(200),
		"current_interval_usage_count": float64(50),
	}
	window := readQuotaWindow(value, "interval")
	if window.RemainingPercent == nil || *window.RemainingPercent != 25 {
		t.Fatalf("percent = %v, want 25", window.RemainingPercent)
	}
}

func TestReadQuotaWindowClamps(t *testing.T) {
	over := readQuotaWindow(map[string]any{"current_interval_status": float64(1), "current_interval_remaining_percent": float64(150)}, "interval")
	if over.RemainingPercent == nil || *over.RemainingPercent != 100 {
		t.Fatalf("clamp high = %v, want 100", over.RemainingPercent)
	}
	under := readQuotaWindow(map[string]any{"current_interval_status": float64(1), "current_interval_remaining_percent": float64(-5)}, "interval")
	if under.RemainingPercent == nil || *under.RemainingPercent != 0 {
		t.Fatalf("clamp low = %v, want 0", under.RemainingPercent)
	}
}

func TestProjectMembership(t *testing.T) {
	body := map[string]any{
		"data": map[string]any{
			"has_token_plan":        true,
			"op_group_id":           "grp-9",
			"token_plan_tier":       "MAX",
			"token_plan_expires_at": float64(1900000000000),
			"op_credit_summary":     map[string]any{"total_remaining_amount": "5000"},
		},
	}
	membership := projectMembership(body)
	if membership.HasTokenPlan == nil || !*membership.HasTokenPlan {
		t.Fatal("has_token_plan not read")
	}
	if membership.OpGroupID != "grp-9" || membership.Tier != "MAX" {
		t.Fatalf("membership = %+v", membership)
	}
	if membership.ExpiresAtMs != 1900000000000 || membership.CreditBalance != "5000" {
		t.Fatalf("membership = %+v", membership)
	}
}

func TestMergeMembershipPrefersDefiniteAndNonEmpty(t *testing.T) {
	personal := Membership{Tier: "PRO", OpGroupID: "grp-1"}
	scoped := Membership{Tier: "MAX", CreditBalance: "1000"}
	merged := mergeMembership(personal, scoped)
	if merged.Tier != "PRO" {
		t.Fatalf("tier = %q, want PRO (personal wins)", merged.Tier)
	}
	if merged.CreditBalance != "1000" {
		t.Fatalf("credit = %q, want 1000 (scoped fills)", merged.CreditBalance)
	}
	if merged.OpGroupID != "grp-1" {
		t.Fatalf("op group = %q", merged.OpGroupID)
	}
}

func TestEncodeURIComponentMatchesJavaScript(t *testing.T) {
	if got := encodeURIComponent("/a b?c=1"); got != "%2Fa%20b%3Fc%3D1" {
		t.Fatalf("encode = %q", got)
	}
	if got := encodeURIComponent("Az-_.!~*'()"); got != "Az-_.!~*'()" {
		t.Fatalf("unreserved changed: %q", got)
	}
	if got := encodeURIComponent("中"); got != "%E4%B8%AD" {
		t.Fatalf("utf8 = %q", got)
	}
}

func TestMD5HexMatchesReference(t *testing.T) {
	// Matches `printf %s abc | md5sum`; guards the wire signature digest.
	if got := md5Hex("abc"); got != "900150983cd24fb0d6963f7d28e17f72" {
		t.Fatalf("md5Hex(abc) = %q", got)
	}
}

func TestFetchAccountQuotaEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/api/openplatform/coding_plan/remains"):
			// The probe is a plain bearer read with the operation-group header;
			// only the matrix calls carry the signed attribution headers.
			if r.Header.Get("X-Group-Id") != "grp-x" {
				t.Errorf("X-Group-Id = %q", r.Header.Get("X-Group-Id"))
			}
			_, _ = w.Write([]byte(`{"base_resp":{"status_code":0},"model_remains":[
				{"model_name":"MiniMax-M3","current_interval_status":1,"current_interval_remaining_percent":90,"end_time":1900000000000,"current_weekly_status":3},
				{"model_name":"video-01","current_interval_status":1,"current_interval_total_count":10,"current_interval_usage_count":3,"end_time":1900000000000}]}`))
		case strings.HasSuffix(r.URL.Path, "/v1/api/user/info"):
			if r.Header.Get("yy") == "" || r.Header.Get("x-signature") == "" || r.Header.Get("x-timestamp") == "" {
				t.Errorf("missing signature headers on %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"base_resp":{"status_code":0},"data":{"user_info":{"real_user_id":"user-7"}}}`))
		case strings.HasSuffix(r.URL.Path, "/get_user_extra_info"):
			if r.Header.Get("yy") == "" || r.Header.Get("x-signature") == "" {
				t.Errorf("missing signature headers on %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"workspaces":[{"workspace_type":0,"workspace_id":12,"has_token_plan":true,"op_group_id":"grp-x","token_plan_tier":"PRO"}]}`))
		case strings.HasSuffix(r.URL.Path, "/get_membership_info"):
			_, _ = w.Write([]byte(`{"data":{"op_credit_summary":{"total_remaining_amount":"42"}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := (&Client{httpClient: http.DefaultClient, region: RegionEN}).
		WithOrigins(server.URL, server.URL)
	quota, err := client.FetchAccountQuota(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("FetchAccountQuota: %v", err)
	}
	if quota.NotSubscribed {
		t.Fatal("not_subscribed = true")
	}
	if quota.Membership.Tier != "PRO" || quota.Membership.OpGroupID != "grp-x" {
		t.Fatalf("membership = %+v", quota.Membership)
	}
	if quota.Membership.CreditBalance != "42" {
		t.Fatalf("credit = %q (scoped read wins)", quota.Membership.CreditBalance)
	}
	if quota.Snapshot == nil {
		t.Fatal("snapshot missing")
	}
	if quota.Snapshot.FiveHour.RemainingPercent == nil || *quota.Snapshot.FiveHour.RemainingPercent != 90 {
		t.Fatalf("five hour = %+v", quota.Snapshot.FiveHour)
	}
	if !quota.Snapshot.Weekly.Unlimited {
		t.Fatalf("weekly should be unlimited: %+v", quota.Snapshot.Weekly)
	}
	if quota.Snapshot.Video == nil || quota.Snapshot.Video.TotalCount == nil || *quota.Snapshot.Video.TotalCount != 10 {
		t.Fatalf("video = %+v", quota.Snapshot.Video)
	}
}

func TestFetchAccountQuotaUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	client := (&Client{httpClient: http.DefaultClient, region: RegionEN}).WithOrigins(server.URL, server.URL)
	_, err := client.FetchAccountQuota(context.Background(), "token")
	if !IsAuthError(err) {
		t.Fatalf("err = %v, want AuthError", err)
	}
}

func TestFetchAccountQuotaRejectsEmptyToken(t *testing.T) {
	client := &Client{httpClient: http.DefaultClient, region: RegionEN}
	if _, err := client.FetchAccountQuota(context.Background(), "  "); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestQuotaSnapshotJSONShape(t *testing.T) {
	// Guards that the exported shapes stay JSON-serializable for the handler.
	payload, errMarshal := json.Marshal(QuotaSnapshot{FiveHour: QuotaWindow{RemainingPercent: intPtr(50), ResetAtMs: 1}})
	if errMarshal != nil {
		t.Fatalf("marshal: %v", errMarshal)
	}
	if !strings.Contains(string(payload), "FiveHour") {
		t.Fatalf("unexpected snapshot JSON: %s", payload)
	}
}
