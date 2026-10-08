package floatboat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

func TestBuildSignInURL(t *testing.T) {
	client := NewClientWithProxyURL(nil, "", "")
	got := client.BuildSignInURL("state-123")
	if !strings.HasPrefix(got, DefaultBackendURL+DefaultSignInPath+"?") {
		t.Fatalf("sign-in URL = %q", got)
	}
	if !strings.Contains(got, "state=state-123") {
		t.Fatalf("sign-in URL missing state: %q", got)
	}
}

func TestExchangeCodeAndFetchProfileAndKey(t *testing.T) {
	var sawExchangeBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/desktop/auth/exchange":
			_ = json.NewDecoder(r.Body).Decode(&sawExchangeBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"accessToken":      "acc-1",
				"refreshToken":     "ref-1",
				"expiresIn":        3600,
				"refreshExpiresIn": 7200,
			})
		case "/api/desktop/user/me":
			if got := r.Header.Get("Authorization"); got != "Bearer acc-1" {
				t.Errorf("user/me Authorization = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user":                  map[string]any{"id": "u1", "email": "a@b.c", "name": "Ann"},
				"credits":               42,
				"hasActiveSubscription": true,
				"membershipLevel":       "pro",
			})
		case "/api/desktop/newapi/key":
			if got := r.Header.Get("Authorization"); got != "Bearer acc-1" {
				t.Errorf("newapi/key Authorization = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"api_key": "fbk_live"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClientWithProxyURL(&config.Config{}, "", server.URL)
	token, errExchange := client.ExchangeCode(context.Background(), "code-1", "state-1", "")
	if errExchange != nil {
		t.Fatalf("ExchangeCode: %v", errExchange)
	}
	if token.AccessToken != "acc-1" || token.RefreshToken != "ref-1" || token.ExpiresIn != 3600 {
		t.Fatalf("token = %+v", token)
	}
	if token.ExpiresAt.IsZero() {
		t.Fatal("ExpiresAt not set from expiresIn")
	}
	if sawExchangeBody["code"] != "code-1" || sawExchangeBody["state"] != "state-1" {
		t.Fatalf("exchange body = %+v", sawExchangeBody)
	}

	profile, errProfile := client.FetchUserProfile(context.Background(), token.AccessToken)
	if errProfile != nil {
		t.Fatalf("FetchUserProfile: %v", errProfile)
	}
	if profile.ID != "u1" || profile.Email != "a@b.c" || profile.Credits != 42 || !profile.HasSubscription {
		t.Fatalf("profile = %+v", profile)
	}

	key, errKey := client.FetchNewAPIKey(context.Background(), token.AccessToken)
	if errKey != nil {
		t.Fatalf("FetchNewAPIKey: %v", errKey)
	}
	if key != "fbk_live" {
		t.Fatalf("api key = %q", key)
	}
}

func TestFetchNewAPIKeyDataEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_key": "wrapped"}})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	key, errKey := client.FetchNewAPIKey(context.Background(), "acc")
	if errKey != nil {
		t.Fatalf("FetchNewAPIKey: %v", errKey)
	}
	if key != "wrapped" {
		t.Fatalf("api key = %q, want wrapped", key)
	}
}

func TestRefreshRotatesAndPreservesToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/desktop/auth/refresh" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "ok",
			"data": map[string]any{"access_token": "acc-2", "expires_in": 1800},
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
}

func TestExchangeCodeErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	if _, err := client.ExchangeCode(context.Background(), "bad", "s", ""); err == nil {
		t.Fatal("expected exchange error")
	}
}

// TestExchangeCodeEnvelopeShape pins the real backend response observed live:
// {"code":0,"message":"ok","data":{"access_token":...,"refresh_token":...}}.
// The desktop JWT carries no expires_in, so the lifetime comes from exp-iat.
func TestExchangeCodeEnvelopeShape(t *testing.T) {
	// JWT with iat=1000, exp=4600 => 3600s lifetime; fid/email in the payload.
	jwt := "eyJhbGciOiJIUzI1NiJ9." +
		"eyJmaWQiOiJmYWtlLWZpZCIsImVtYWlsIjoiYUBiLmMiLCJkaXNwbGF5X25hbWUiOiJBbm4iLCJleHAiOjQ2MDAsImlhdCI6MTAwMH0." +
		"sig"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "ok",
			"data": map[string]any{
				"access_token":  jwt,
				"refresh_token": "ref-x",
			},
		})
	}))
	defer server.Close()

	client := NewClientWithProxyURL(nil, "", server.URL)
	token, errExchange := client.ExchangeCode(context.Background(), "c", "s", "")
	if errExchange != nil {
		t.Fatalf("ExchangeCode: %v", errExchange)
	}
	if token.AccessToken != jwt || token.RefreshToken != "ref-x" {
		t.Fatalf("token = %+v", token)
	}
	if token.ExpiresIn != 3600 {
		t.Fatalf("ExpiresIn = %d, want 3600 derived from exp-iat", token.ExpiresIn)
	}
	if token.ExpiresAt.Unix() != 4600 {
		t.Fatalf("ExpiresAt = %v, want unix 4600", token.ExpiresAt)
	}

	// The profile endpoint is offline here, so the identity must come from JWT claims.
	profile, errProfile := client.FetchUserProfile(context.Background(), jwt)
	if errProfile != nil {
		t.Fatalf("FetchUserProfile: %v", errProfile)
	}
	if profile.ID != "fake-fid" || profile.Email != "a@b.c" || profile.Name != "Ann" {
		t.Fatalf("profile from JWT = %+v", profile)
	}
}

// TestExchangeCodeEnvelopeErrorSurfaces rejects a non-zero envelope code instead
// of silently parsing an error payload as credentials.
func TestExchangeCodeEnvelopeErrorSurfaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 40101, "message": "code expired"})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	_, err := client.ExchangeCode(context.Background(), "c", "s", "")
	if err == nil {
		t.Fatal("expected envelope error to surface")
	}
	if !strings.Contains(err.Error(), "code expired") {
		t.Fatalf("error = %v, want it to carry the backend message", err)
	}
}

// TestFetchNewAPIKeyEnvelopeShape covers the minted-key envelope.
func TestFetchNewAPIKeyEnvelopeShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "ok",
			"data": map[string]any{"api_key": "fbk_envelope"},
		})
	}))
	defer server.Close()
	client := NewClientWithProxyURL(nil, "", server.URL)
	key, errKey := client.FetchNewAPIKey(context.Background(), "acc")
	if errKey != nil {
		t.Fatalf("FetchNewAPIKey: %v", errKey)
	}
	if key != "fbk_envelope" {
		t.Fatalf("api key = %q", key)
	}
}

func TestDefaultEndpoints(t *testing.T) {
	endpoints := DefaultEndpoints()
	if endpoints.Exchange != "/api/desktop/auth/exchange" {
		t.Fatalf("exchange = %q", endpoints.Exchange)
	}
	// The desktop backend keeps the refresh route under /api/desktop/auth/, not
	// the /api/v1/ namespace the client bundle suggested. A wrong path here
	// makes every expired token unrecoverable.
	if endpoints.Refresh != "/api/desktop/auth/refresh" {
		t.Fatalf("refresh = %q", endpoints.Refresh)
	}
	if endpoints.Pricing != "/api/pricing" {
		t.Fatalf("pricing = %q", endpoints.Pricing)
	}
	if DefaultInferenceBaseURL != "https://newapi.aoe.chat" {
		t.Fatalf("inference base = %q", DefaultInferenceBaseURL)
	}
}
