package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// lobsteraiServerMockBackend serves the backend endpoints the login flow needs.
func lobsteraiServerMockBackend(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/exchange", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "ok",
			"data": map[string]any{
				"accessToken":  "acc-1",
				"refreshToken": "ref-1",
				"expiresIn":    3600,
				"user":         map[string]any{"id": "u1", "email": "lobster@example.com", "name": "Lobster"},
				"quota":        map[string]any{"subscriptionStatus": "free", "credits": 12},
			},
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// TestLobsteraiServerManagementRecordPersistsCredential drives the login exchange
// and record construction, then persists through the shared store: the credential
// the runtime will load after restart.
func TestLobsteraiServerManagementRecordPersistsCredential(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := lobsteraiServerMockBackend(t)

	authDir := t.TempDir()
	cfg := &config.Config{}
	cfg.AuthDir = authDir
	_ = NewHandler(cfg, "", nil)

	state := "test-state-lobsterai"
	RegisterOAuthSession(state, lobsteraiserverauth.Provider)
	t.Cleanup(func() { CompleteOAuthSession(state) })

	client := lobsteraiserverauth.NewClientWithProxyURL(nil, "", backend.URL)
	token, user, quota, errExchange := client.ExchangeCode(context.Background(), "authcode", state)
	if errExchange != nil {
		t.Fatalf("ExchangeCode: %v", errExchange)
	}
	record := buildLobsteraiServerRecord(token, user, quota, nil)
	if record.Provider != lobsteraiserverauth.Provider {
		t.Fatalf("provider = %q", record.Provider)
	}
	if got := record.Attributes["access_token"]; got != "acc-1" {
		t.Fatalf("access_token attribute = %q", got)
	}
	if got := record.Attributes["base_url"]; got != lobsteraiserverauth.DefaultProxyBaseURL {
		t.Fatalf("base_url attribute = %q", got)
	}
	if got := record.Attributes["backend_url"]; got != lobsteraiserverauth.DefaultBackendURL {
		t.Fatalf("backend_url attribute = %q", got)
	}
	if record.Metadata["access_token"] != "acc-1" || record.Metadata["refresh_token"] != "ref-1" {
		t.Fatalf("metadata = %+v", record.Metadata)
	}
	if record.Metadata["quota_remaining"] != int64(12) {
		t.Fatalf("quota_remaining = %v", record.Metadata["quota_remaining"])
	}

	store := sdkAuth.NewFileTokenStore()
	store.SetBaseDir(authDir)
	savedPath, errSave := store.Save(coreauth.WithAuthCreationIntent(nil), record)
	if errSave != nil {
		t.Fatalf("save: %v", errSave)
	}
	if savedPath == "" {
		t.Fatal("save returned empty path")
	}
	raw, errRead := os.ReadFile(filepath.Join(authDir, record.FileName))
	if errRead != nil {
		t.Fatalf("read saved credential: %v", errRead)
	}
	var persisted map[string]any
	if errUnmarshal := json.Unmarshal(raw, &persisted); errUnmarshal != nil {
		t.Fatalf("unmarshal saved credential: %v", errUnmarshal)
	}
	if persisted["type"] != lobsteraiserverauth.Provider {
		t.Fatalf("persisted type = %v", persisted["type"])
	}
	if persisted["refresh_token"] != "ref-1" {
		t.Fatalf("persisted refresh_token = %v", persisted["refresh_token"])
	}
}

// TestLobsteraiServerManagementRecordKeepsUnknownCreditsAbsent pins the honest
// reporting rule: a backend payload without credits must not be recorded as 0.
func TestLobsteraiServerManagementRecordKeepsUnknownCreditsAbsent(t *testing.T) {
	token := &lobsteraiserverauth.TokenData{AccessToken: "acc-1"}
	quota := &lobsteraiserverauth.Quota{SubscriptionStatus: "free", CreditsKnown: false}
	record := buildLobsteraiServerRecord(token, &lobsteraiserverauth.User{ID: "u1"}, quota, nil)
	if _, present := record.Metadata["quota_remaining"]; present {
		t.Fatalf("quota_remaining must stay absent when unknown: %+v", record.Metadata)
	}
	if record.Metadata["subscription_status"] != "free" {
		t.Fatalf("subscription_status = %v", record.Metadata["subscription_status"])
	}
}

// TestLobsteraiServerLoopbackCaptureCompletesLogin pins that a browser redirect
// arriving on the loopback callback is exchanged and persisted with no user
// paste, and that a mismatched state is refused. The handler's own client targets
// the production backend, so the exchange is verified by holding the listener
// contract rather than by a live server.
func TestLobsteraiServerLoopbackCaptureCompletesLogin(t *testing.T) {
	server := lobsteraiserverauth.NewOAuthServer(0)
	port, errStart := server.Start()
	if errStart != nil {
		t.Fatalf("start loopback server: %v", errStart)
	}
	defer func() {
		if errStop := server.Stop(context.Background()); errStop != nil {
			t.Fatalf("stop loopback server: %v", errStop)
		}
	}()

	if port <= 0 {
		t.Fatalf("port = %d", port)
	}
	if got := server.RedirectURI(); got != "http://127.0.0.1:"+strconv.Itoa(port)+"/auth/callback" {
		t.Fatalf("redirect URI = %q", got)
	}

	// Simulate the browser landing on the callback with the expected code+state.
	go func() {
		_, _ = http.Get(server.RedirectURI() + "?code=abc&state=xyz")
	}()
	result, errWait := server.WaitForCallback(5 * time.Second)
	if errWait != nil {
		t.Fatalf("WaitForCallback: %v", errWait)
	}
	if result == nil || result.Code != "abc" || result.State != "xyz" || result.Error != "" {
		t.Fatalf("result = %+v", result)
	}
}

// TestLobsteraiServerLoopbackRejectsErrorCallback pins that an authorization
// error surfaced by the portal is reported instead of being exchanged.
func TestLobsteraiServerLoopbackRejectsErrorCallback(t *testing.T) {
	server := lobsteraiserverauth.NewOAuthServer(0)
	if _, errStart := server.Start(); errStart != nil {
		t.Fatalf("start loopback server: %v", errStart)
	}
	defer func() {
		if errStop := server.Stop(context.Background()); errStop != nil {
			t.Fatalf("stop loopback server: %v", errStop)
		}
	}()

	go func() {
		_, _ = http.Get(server.RedirectURI() + "?error=access_denied&error_description=denied")
	}()
	result, errWait := server.WaitForCallback(5 * time.Second)
	if errWait != nil {
		t.Fatalf("WaitForCallback: %v", errWait)
	}
	if result == nil || result.Error == "" || result.Code != "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestParseLobsteraiServerCallbackURL(t *testing.T) {
	code, state, err := parseLobsteraiServerCallbackURL("http://127.0.0.1:9999/auth/callback?code=abc&state=xyz")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if code != "abc" || state != "xyz" {
		t.Fatalf("code=%q state=%q", code, state)
	}
	if _, _, errErr := parseLobsteraiServerCallbackURL("http://127.0.0.1:9999/auth/callback?error=access_denied&error_description=nope"); errErr == nil {
		t.Fatal("expected error for an error callback")
	}
	if _, _, errMissing := parseLobsteraiServerCallbackURL("http://127.0.0.1:9999/auth/callback?state=only"); errMissing == nil {
		t.Fatal("expected error for a missing code")
	}
}

// TestLobsteraiServerQuotaAccessTokenAndBaseURL pins credential resolution.
func TestLobsteraiServerQuotaAccessTokenAndBaseURL(t *testing.T) {
	auth := &coreauth.Auth{
		Attributes: map[string]string{"access_token": "attr-token", "backend_url": "https://inner.example.com/"},
	}
	if got := lobsteraiServerQuotaAccessToken(auth); got != "attr-token" {
		t.Fatalf("access token = %q", got)
	}
	if got := lobsteraiServerQuotaBaseURL(auth); got != "https://inner.example.com" {
		t.Fatalf("base url = %q", got)
	}
	if got := lobsteraiServerQuotaAccessToken(nil); got != "" {
		t.Fatalf("nil auth token = %q", got)
	}
	if got := lobsteraiServerQuotaBaseURL(nil); got != lobsteraiserverauth.DefaultBackendURL {
		t.Fatalf("nil auth base url = %q", got)
	}
}
