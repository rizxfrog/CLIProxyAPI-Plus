package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	minimaxauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/minimax"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// registerMinimaxAuth builds a manager carrying one MiniMax Code credential.
func registerMinimaxAuth(t *testing.T, provider, accessToken string, metadata map[string]any) (*coreauth.Manager, *coreauth.Auth) {
	t.Helper()
	manager := coreauth.NewManager(nil, nil, nil)
	values := map[string]any{
		"type":         provider,
		"auth_kind":    coreauth.AuthKindOAuth,
		"access_token": accessToken,
		"expired":      time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}
	for key, value := range metadata {
		values[key] = value
	}
	auth := &coreauth.Auth{
		ID:       provider + ":oauth:test",
		Provider: provider,
		Metadata: values,
	}
	if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
		t.Fatalf("register %s auth: %v", provider, errRegister)
	}
	return manager, auth
}

// withMinimaxTestOrigins points the quota handler at a stub matrix+open-platform
// origin for the duration of one test.
func withMinimaxTestOrigins(t *testing.T, matrixOrigin, openOrigin string) func() {
	t.Helper()
	previous := minimaxQuotaOriginOverride
	minimaxQuotaOriginOverride = func(minimaxauth.Region) (string, string) {
		return matrixOrigin, openOrigin
	}
	return func() { minimaxQuotaOriginOverride = previous }
}

// newMinimaxQuotaServer serves the signed identity/membership reads and the
// coding-plan probe. requestLog records the paths for assertions.
func newMinimaxQuotaServer(t *testing.T, identityBody, extraBody, membershipBody, remainsBody string, requestLog *[]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestLog != nil {
			*requestLog = append(*requestLog, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/api/user/info"):
			_, _ = w.Write([]byte(identityBody))
		case strings.HasSuffix(r.URL.Path, "/get_user_extra_info"):
			_, _ = w.Write([]byte(extraBody))
		case strings.HasSuffix(r.URL.Path, "/get_membership_info"):
			_, _ = w.Write([]byte(membershipBody))
		case strings.HasSuffix(r.URL.Path, "/coding_plan/remains"):
			if group := r.Header.Get("X-Group-Id"); group != "grp-1" {
				t.Errorf("X-Group-Id = %q, want grp-1", group)
			}
			_, _ = w.Write([]byte(remainsBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

const (
	minimaxIdentityBody = `{"base_resp":{"status_code":0},"data":{"userInfo":{"realUserID":"user-1"}}}`
	minimaxExtraBody    = `{"base_resp":{"status_code":0},"workspaces":[{"workspace_type":0,"workspace_id":"ws-1","has_token_plan":true,"op_group_id":"grp-1","token_plan_tier":"PRO"}]}`
	minimaxMemberBody   = `{"base_resp":{"status_code":0},"data":{"has_token_plan":true,"op_group_id":"grp-1","op_credit_summary":{"total_remaining_amount":"1000"}}}`
	minimaxRemainsBody  = `{"base_resp":{"status_code":0},"model_remains":[{"model_name":"MiniMax-M3","current_interval_status":1,"current_interval_remaining_percent":80,"end_time":1900000000000,"current_weekly_status":1,"current_weekly_remaining_percent":40,"weekly_end_time":1900600000000}]}`
)

func TestMinimaxQuotaBuildsWindows(t *testing.T) {
	manager, auth := registerMinimaxAuth(t, constant.Minimax, "access-1", map[string]any{"region": "en"})
	var paths []string
	server := newMinimaxQuotaServer(t, minimaxIdentityBody, minimaxExtraBody, minimaxMemberBody, minimaxRemainsBody, &paths)
	restore := withMinimaxTestOrigins(t, server.URL, server.URL)
	defer restore()

	handler := &Handler{authManager: manager}
	router := gin.New()
	router.GET("/minimax-quota", handler.GetMinimaxQuota)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/minimax-quota?auth_index="+auth.EnsureIndex(), nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	var payload MinimaxQuota
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &payload); errDecode != nil {
		t.Fatalf("decode: %v", errDecode)
	}
	if payload.Plan != "PRO" {
		t.Fatalf("plan = %q, want PRO", payload.Plan)
	}
	if payload.CreditBalance != "1000" {
		t.Fatalf("credit_balance = %q, want 1000", payload.CreditBalance)
	}
	if len(payload.Windows) != 2 {
		t.Fatalf("windows = %d, want 2 (%+v)", len(payload.Windows), payload.Windows)
	}
	if payload.Windows[0].Kind != "five_hour" || payload.Windows[1].Kind != "weekly" {
		t.Fatalf("window kinds = %q/%q", payload.Windows[0].Kind, payload.Windows[1].Kind)
	}
	if payload.Windows[0].RemainingPercent == nil || *payload.Windows[0].RemainingPercent != 80 {
		t.Fatalf("five_hour percent = %v, want 80", payload.Windows[0].RemainingPercent)
	}
	if payload.Windows[1].RemainingPercent == nil || *payload.Windows[1].RemainingPercent != 40 {
		t.Fatalf("weekly percent = %v, want 40", payload.Windows[1].RemainingPercent)
	}
	// The identity read carries user_id=0 (no realUserID yet); later reads carry
	// the resolved id, and the probe carries the operation group header.
	if len(paths) < 3 {
		t.Fatalf("expected identity+membership+probe calls, got %v", paths)
	}
}

func TestMinimaxQuotaNotSubscribed(t *testing.T) {
	manager, auth := registerMinimaxAuth(t, constant.MinimaxCN, "access-2", nil)
	member := `{"base_resp":{"status_code":0},"data":{"has_token_plan":false}}`
	server := newMinimaxQuotaServer(t, minimaxIdentityBody, `{"base_resp":{"status_code":0},"workspaces":[]}`, member, minimaxRemainsBody, nil)
	restore := withMinimaxTestOrigins(t, server.URL, server.URL)
	defer restore()

	handler := &Handler{authManager: manager}
	router := gin.New()
	router.GET("/minimax-quota", handler.GetMinimaxQuota)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/minimax-quota?auth_index="+auth.EnsureIndex(), nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	var payload MinimaxQuota
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &payload); errDecode != nil {
		t.Fatalf("decode: %v", errDecode)
	}
	if !payload.NotSubscribed {
		t.Fatalf("not_subscribed = false, want true")
	}
	if len(payload.Windows) != 0 {
		t.Fatalf("windows = %d, want 0 when not subscribed", len(payload.Windows))
	}
}

func TestMinimaxQuotaUnsupportedProvider(t *testing.T) {
	manager, _ := registerMinimaxAuth(t, "claude", "access-3", nil)
	handler := &Handler{authManager: manager}
	router := gin.New()
	router.GET("/minimax-quota", handler.GetMinimaxQuota)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/minimax-quota?auth_index=claude:oauth:test", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestMinimaxQuotaMissingAuthIndex(t *testing.T) {
	manager, _ := registerMinimaxAuth(t, constant.Minimax, "access-4", nil)
	handler := &Handler{authManager: manager}
	router := gin.New()
	router.GET("/minimax-quota", handler.GetMinimaxQuota)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/minimax-quota", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}
