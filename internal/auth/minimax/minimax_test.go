package minimax

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient wires a Client to mock endpoints.
func newTestClient(deviceURL, tokenURL, revokeURL string) *Client {
	return &Client{
		httpClient: http.DefaultClient,
		deviceURL:  deviceURL,
		tokenURL:   tokenURL,
		revokeURL:  revokeURL,
		region:     RegionEN,
		sleep:      func(context.Context, time.Duration) error { return nil },
	}
}

func TestRegionEndpoints(t *testing.T) {
	if got := RegionEN.AccountOrigin(); got != "https://account.minimax.io" {
		t.Fatalf("EN AccountOrigin = %q", got)
	}
	if got := RegionCN.AccountOrigin(); got != "https://account.minimax.cn" {
		t.Fatalf("CN AccountOrigin = %q", got)
	}
	if got := RegionEN.InferenceBaseURL(); got != "https://agent.minimax.io/mavis/api/v1/llm" {
		t.Fatalf("EN InferenceBaseURL = %q", got)
	}
	if got := RegionCN.InferenceBaseURL(); got != "https://agent.minimax.cn/mavis/api/v1/llm" {
		t.Fatalf("CN InferenceBaseURL = %q", got)
	}
}

func TestNormalizeRegion(t *testing.T) {
	cases := map[string]Region{
		"":             RegionEN,
		"en":           RegionEN,
		"us":           RegionEN,
		"cn":           RegionCN,
		"China":        RegionCN,
		"minimax.cn":   RegionCN,
		"minimaxi.com": RegionCN,
	}
	for input, want := range cases {
		if got := NormalizeRegion(input); got != want {
			t.Fatalf("NormalizeRegion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestStartDeviceAuthorizationSendsPKCE(t *testing.T) {
	var gotForm map[string][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"dev-1","user_code":"USER-1","verification_uri":"https://account.minimax.cn/oauth-authorize","verification_uri_complete":"https://account.minimax.cn/oauth-authorize?user_code=USER-1","expires_in":900,"interval":5}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, server.URL, server.URL)
	device, err := client.StartDeviceAuthorization(context.Background())
	if err != nil {
		t.Fatalf("StartDeviceAuthorization error: %v", err)
	}
	if device.DeviceCode != "dev-1" || device.UserCode != "USER-1" {
		t.Fatalf("unexpected device auth: %+v", device)
	}
	if device.VerificationURI != "https://account.minimax.cn/oauth-authorize" || device.ExpiresIn != 900 || device.Interval != 5 {
		t.Fatalf("unexpected device auth fields: %+v", device)
	}
	if device.VerificationURIComplete != "https://account.minimax.cn/oauth-authorize?user_code=USER-1" {
		t.Fatalf("verification_uri_complete not preserved: %+v", device)
	}
	if got := gotForm["client_id"]; len(got) != 1 || got[0] != ClientID {
		t.Fatalf("client_id = %q", got)
	}
	if got := gotForm["scope"]; len(got) != 1 || got[0] != Scope {
		t.Fatalf("scope = %q", got)
	}
	if got := gotForm["audience"]; len(got) != 1 || got[0] != Audience {
		t.Fatalf("audience = %q", got)
	}
	if got := gotForm["code_challenge_method"]; len(got) != 1 || got[0] != "S256" {
		t.Fatalf("code_challenge_method = %q", got)
	}
	if got := gotForm["code_challenge"]; len(got) != 1 || got[0] == "" {
		t.Fatal("code_challenge missing")
	}
	if device.CodeVerifier == "" {
		t.Fatal("code verifier missing")
	}
}

// TestAuthorizationURLMatchesNativeClient reproduces the native client's
// markTuiAuthorizationUrl output: the user_code-bearing complete URL with the
// tui surface and download-source attribution parameters appended.
func TestAuthorizationURLMatchesNativeClient(t *testing.T) {
	device := &DeviceAuthorization{
		UserCode:                "B5X9-ZB2Q",
		VerificationURI:         "https://account.minimax.cn/oauth-authorize",
		VerificationURIComplete: "https://account.minimax.cn/oauth-authorize?user_code=B5X9-ZB2Q",
	}
	got := device.AuthorizationURL()
	want := "https://account.minimax.cn/oauth-authorize?client_surface=tui&download_source=mcode-internal&user_code=B5X9-ZB2Q"
	if got != want {
		t.Fatalf("AuthorizationURL() = %q, want %q", got, want)
	}
}

func TestAuthorizationURLFallsBackAndInjectsUserCode(t *testing.T) {
	// No complete URL: the bare verification URI is extended with the user code.
	device := &DeviceAuthorization{
		UserCode:        "ABCD-EFGH",
		VerificationURI: "https://account.minimax.io/oauth-authorize",
	}
	got := device.AuthorizationURL()
	want := "https://account.minimax.io/oauth-authorize?client_surface=tui&download_source=mcode-internal&user_code=ABCD-EFGH"
	if got != want {
		t.Fatalf("AuthorizationURL() = %q, want %q", got, want)
	}
}

func TestStartDeviceAuthorizationRejectsIncompleteResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"dev-1"}`))
	}))
	defer server.Close()
	client := newTestClient(server.URL, server.URL, server.URL)
	if _, err := client.StartDeviceAuthorization(context.Background()); err == nil {
		t.Fatal("expected error for incomplete device authorization response")
	}
}

func TestPollDeviceTokenPendingThenSuccess(t *testing.T) {
	jwt := makeJWT(t, map[string]any{"sub": "user-1", "account_id": "acct-1"})
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polls++
		w.Header().Set("Content-Type", "application/json")
		if polls == 1 {
			_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"` + jwt + `","refresh_token":"refresh-1","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, server.URL, server.URL)
	token, err := client.PollDeviceToken(context.Background(), &DeviceAuthorization{
		DeviceCode:   "dev-1",
		CodeVerifier: "verifier",
		ExpiresIn:    30,
		Interval:     0, // force immediate-ish polling; interval seconds default to 5 but deadline allows
	})
	if err != nil {
		t.Fatalf("PollDeviceToken error: %v", err)
	}
	if token.AccessToken != jwt || token.RefreshToken != "refresh-1" {
		t.Fatalf("unexpected token: %+v", token)
	}
	if token.Subject != "user-1" || token.AccountID != "acct-1" {
		t.Fatalf("unexpected JWT claims: sub=%q account=%q", token.Subject, token.AccountID)
	}
	if token.ExpiresAt.IsZero() {
		t.Fatal("ExpiresAt not set")
	}
}

func TestPollDeviceTokenRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":"access_denied"}`))
	}))
	defer server.Close()
	client := newTestClient(server.URL, server.URL, server.URL)
	if _, err := client.PollDeviceToken(context.Background(), &DeviceAuthorization{DeviceCode: "d", CodeVerifier: "v", ExpiresIn: 30, Interval: 1}); err == nil {
		t.Fatal("expected rejection error")
	}
}

func TestRefreshRotatesTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Errorf("grant_type = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","token_type":"Bearer","expires_in":1800}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, server.URL, server.URL)
	token, err := client.Refresh(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("Refresh error: %v", err)
	}
	if token.AccessToken != "new-access" {
		t.Fatalf("access = %q", token.AccessToken)
	}
	// Rotation omitted by upstream: the previous refresh token must be retained.
	if token.RefreshToken != "old-refresh" {
		t.Fatalf("refresh token not preserved: %q", token.RefreshToken)
	}
}

func TestRevoke(t *testing.T) {
	var hit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := newTestClient(server.URL, server.URL, server.URL)
	if err := client.Revoke(context.Background(), "refresh-1"); err != nil {
		t.Fatalf("Revoke error: %v", err)
	}
	if !hit {
		t.Fatal("revoke endpoint not hit")
	}
	if err := client.Revoke(context.Background(), "  "); err == nil {
		t.Fatal("expected error for empty refresh token")
	}
}

func TestNewClientUsesRegionOrigins(t *testing.T) {
	en := NewClient(nil, RegionEN)
	if !strings.HasPrefix(en.deviceURL, "https://account.minimax.io") {
		t.Fatalf("EN device URL = %q", en.deviceURL)
	}
	cn := NewClient(nil, RegionCN)
	if !strings.HasPrefix(cn.tokenURL, "https://account.minimax.cn") {
		t.Fatalf("CN token URL = %q", cn.tokenURL)
	}
}

func makeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}
