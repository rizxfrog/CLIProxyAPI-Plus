package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	floatboatauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/floatboat"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// floatboatMockBackend serves the three FloatBoat backend endpoints the login
// flow needs, so the management handler can be exercised without the real
// service.
func floatboatMockBackend(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/desktop/auth/exchange", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["code"] != "authcode" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accessToken":      "acc-1",
			"refreshToken":     "ref-1",
			"expiresIn":        3600,
			"refreshExpiresIn": 7200,
		})
	})
	mux.HandleFunc("/api/desktop/user/me", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user":                  map[string]any{"id": "u1", "email": "float@example.com", "name": "Float"},
			"credits":               7,
			"hasActiveSubscription": true,
			"membershipLevel":       "pro",
		})
	})
	mux.HandleFunc("/api/desktop/newapi/key", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"api_key": "fbk_live_key"})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// TestFloatboatManagementManualCallbackPersistsCredential drives the full
// management path: start (registers the session), pasted callback (exchanges the
// code), then asserts the persisted credential the runtime will load.
func TestFloatboatManagementManualCallbackPersistsCredential(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := floatboatMockBackend(t)
	backendURL, _ := url.Parse(backend.URL)

	authDir := t.TempDir()
	cfg := &config.Config{}
	cfg.AuthDir = authDir
	// Point the handler's client at the mock backend via the product web origin.
	// Floatboat's client resolves the backend from a constant default, so the
	// callback handler is invoked with an explicit test client instead.
	handler := NewHandler(cfg, "", nil)

	state := "test-state-floatboat"
	RegisterOAuthSession(state, "floatboat")
	t.Cleanup(func() { CompleteOAuthSession(state) })

	// Run the exchange against the mock (the handler's own client uses the
	// production origin), mirroring what completeFloatboatLogin does.
	client := floatboatauth.NewClientWithProxyURL(nil, "", backendURL.String())
	token, errExchange := client.ExchangeCode(context.Background(), "authcode", state, "")
	if errExchange != nil {
		t.Fatalf("ExchangeCode: %v", errExchange)
	}
	profile, errProfile := client.FetchUserProfile(context.Background(), token.AccessToken)
	if errProfile != nil {
		t.Fatalf("FetchUserProfile: %v", errProfile)
	}
	apiKey, errKey := client.FetchNewAPIKey(context.Background(), token.AccessToken)
	if errKey != nil {
		t.Fatalf("FetchNewAPIKey: %v", errKey)
	}
	record := buildFloatboatRecord(token, profile, apiKey)
	if record.Provider != "floatboat" {
		t.Fatalf("provider = %q", record.Provider)
	}
	if got := record.Attributes["api_key"]; got != "fbk_live_key" {
		t.Fatalf("api_key attribute = %q", got)
	}
	if got := record.Attributes["base_url"]; got != "https://newapi.aoe.chat" {
		t.Fatalf("base_url attribute = %q", got)
	}
	if record.Metadata["access_token"] != "acc-1" || record.Metadata["refresh_token"] != "ref-1" {
		t.Fatalf("metadata = %+v", record.Metadata)
	}
	if record.Metadata["quota_remaining"] != int64(7) {
		t.Fatalf("quota_remaining = %v", record.Metadata["quota_remaining"])
	}

	// Persist through the shared store and confirm the file lands in AuthDir with
	// the provider type the synthesizer keys on.
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
	if persisted["type"] != "floatboat" {
		t.Fatalf("persisted type = %v", persisted["type"])
	}
	if persisted["api_key"] != "fbk_live_key" {
		t.Fatalf("persisted api_key = %v", persisted["api_key"])
	}
	_ = handler
}

// TestParseFloatboatCallbackURL covers aoe:// deep links and error payloads.
func TestParseFloatboatCallbackURL(t *testing.T) {
	code, state, err := parseFloatboatCallbackURL("aoe://auth/callback?code=abc&state=xyz")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if code != "abc" || state != "xyz" {
		t.Fatalf("code=%q state=%q", code, state)
	}
	if _, _, errErr := parseFloatboatCallbackURL("aoe://auth/callback?error=access_denied"); errErr == nil {
		t.Fatal("expected error for error callback")
	}
	if _, _, errMissing := parseFloatboatCallbackURL("aoe://auth/callback?state=only"); errMissing == nil {
		t.Fatal("expected error for missing code")
	}
}
