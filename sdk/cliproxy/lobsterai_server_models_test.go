package cliproxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

// newLobsteraiServerCatalogueServer serves the backend's /api/models/available
// shape used by the catalogue probe.
func newLobsteraiServerCatalogueServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/models/available" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": []any{
				map[string]any{"modelId": "deepseek-v4-flash", "modelName": "DeepSeek V4 Flash", "apiFormat": "openai", "accessible": true},
				map[string]any{"modelId": "deepseek-v4-pro", "apiFormat": "openai", "contextWindow": 200000, "accessible": true},
				map[string]any{"modelId": "kimi-k3", "apiFormat": "openai", "accessible": false},
			},
		})
	}))
}

// TestLobsteraiServerModelsForCatalogueCarriesFormat pins the critical wiring
// between the backend catalogue and the executor's per-model protocol selection:
// each registered model must carry its own apiFormat through the shared
// exclusion/alias/prefix pipeline, because the executor reads it back from the
// global registry to choose Chat Completions / Responses / Messages.
func TestLobsteraiServerModelsForCatalogueCarriesFormat(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	auth := &coreauth.Auth{
		ID:       "lobsterai-catalogue-format",
		Provider: constant.LobsteraiServer,
		Attributes: map[string]string{
			"auth_kind": "oauth",
		},
	}
	service := &Service{cfg: &config.Config{}}

	catalogue := []lobsteraiserverauth.Model{
		{ID: "deepseek-v4-flash", DisplayName: "DeepSeek V4 Flash", APIFormat: "openai", Accessible: true},
		{ID: "deepseek-v4-pro", DisplayName: "DeepSeek V4 Pro", APIFormat: "openai", Accessible: true, ContextLimit: 200000},
		{ID: "kimi-k3", APIFormat: "openai", Accessible: true},
	}
	models := service.lobsteraiServerModelsForCatalogue(auth, catalogue)
	if len(models) != 3 {
		t.Fatalf("models = %d, want 3", len(models))
	}

	// The pipeline clones models; identity and limits must survive on the clones
	// that are actually registered.
	registered := service.lobsteraiServerModelsForCatalogue(auth, catalogue)
	byID := make(map[string]*ModelInfo, len(registered))
	for _, model := range registered {
		byID[model.ID] = model
	}
	if byID["deepseek-v4-pro"].DisplayName != "DeepSeek V4 Pro" {
		t.Fatalf("display name lost: %+v", byID["deepseek-v4-pro"])
	}
	if byID["deepseek-v4-pro"].ContextLength != 200000 {
		t.Fatalf("context window lost: %+v", byID["deepseek-v4-pro"])
	}

	// Registration must round-trip the format through the global registry, which
	// is exactly where the executor reads it from.
	service.registerResolvedModelsForAuth(auth, constant.LobsteraiServer, models)
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	if info := registry.LookupModelInfo("deepseek-v4-pro", constant.LobsteraiServer); info == nil || info.Type != constant.LobsteraiServer {
		t.Fatalf("registered deepseek-v4-pro lost: %+v", info)
	}
}

// TestLobsteraiServerCatalogueModelInfoFillsContextWindow pins that per-model
// limits reported by the backend are carried onto the registered entry.
func TestLobsteraiServerCatalogueModelInfoFillsContextWindow(t *testing.T) {
	info := lobsteraiServerCatalogueModelInfo(nil, lobsteraiserverauth.Model{
		ID:           "big-context",
		APIFormat:    "openai",
		ContextLimit: 1_000_000,
		MaxTokens:    65_536,
	})
	if info == nil {
		t.Fatal("model info is nil")
	}
	if info.ContextLength != 1_000_000 || info.MaxCompletionTokens != 65_536 {
		t.Fatalf("limits not carried: %+v", info)
	}
	if info.Type != constant.LobsteraiServer {
		t.Fatalf("type = %q", info.Type)
	}
}

// TestLobsteraiServerEmptyCatalogueModelInfoSkipsBlankID pins that a malformed
// catalogue entry cannot register a nameless model.
func TestLobsteraiServerCatalogueModelInfoSkipsBlankID(t *testing.T) {
	if info := lobsteraiServerCatalogueModelInfo(nil, lobsteraiserverauth.Model{ID: "  "}); info != nil {
		t.Fatalf("blank id produced %+v", info)
	}
}

// TestLobsteraiServerCatalogueModelInfoPublishesDeclaredThinking pins that the
// backend's own reasoning vocabulary reaches the registry.
//
// Without model-level thinking metadata the shared pipeline leaves
// `reasoning_effort` out of the upstream body entirely, so every request ran at
// the backend default no matter which level the client asked for.
func TestLobsteraiServerCatalogueModelInfoPublishesDeclaredThinking(t *testing.T) {
	info := lobsteraiServerCatalogueModelInfo(nil, lobsteraiserverauth.Model{
		ID:              "deepseek-v4-flash",
		APIFormat:       "openai",
		Thinking:        true,
		ReasoningLevels: []string{"off", "high", "xhigh"},
	})
	if info == nil || info.Thinking == nil {
		t.Fatalf("thinking metadata missing: %+v", info)
	}
	if got := info.Thinking.Levels; len(got) != 3 || got[0] != "off" || got[2] != "xhigh" {
		t.Fatalf("levels = %+v", got)
	}
	// "off" is a real level here (the backend's own token), so a model that
	// declares it can be switched off.
	if !info.Thinking.ZeroAllowed {
		t.Fatal("a declared off level must mark the model as disableable")
	}
	// The published slice must not alias the catalogue entry's backing array.
	levels := []string{"off", "high"}
	copied := lobsteraiServerCatalogueModelInfo(nil, lobsteraiserverauth.Model{
		ID: "copy", Thinking: true, ReasoningLevels: levels,
	})
	levels[0] = "mutated"
	if copied.Thinking.Levels[0] != "off" {
		t.Fatalf("published levels alias the catalogue entry: %+v", copied.Thinking.Levels)
	}
}

// TestLobsteraiServerCatalogueModelInfoOwnsItsThinkingMetadata pins that the
// catalogue entry, not a same-named entry from another provider, decides this
// channel's reasoning vocabulary.
//
// The static registry is keyed by model id across every provider, and ids such
// as "kimi-k3" exist in several channels, so an inheriting clone would enforce
// another provider's level list against this backend.
func TestLobsteraiServerCatalogueModelInfoOwnsItsThinkingMetadata(t *testing.T) {
	// Thinking reported, no levels declared: forward the client's choice without
	// inventing a restriction, and replace anything inherited.
	undeclared := lobsteraiServerCatalogueModelInfo(nil, lobsteraiserverauth.Model{
		ID:       "kimi-k3",
		Thinking: true,
	})
	if undeclared == nil || undeclared.Thinking == nil {
		t.Fatalf("thinking metadata missing: %+v", undeclared)
	}
	if len(undeclared.Thinking.Levels) != 0 || undeclared.Thinking.ZeroAllowed {
		t.Fatalf("undeclared levels must stay unconstrained: %+v", undeclared.Thinking)
	}

	// No thinking at all: an effort must not be forwarded.
	plain := lobsteraiServerCatalogueModelInfo(nil, lobsteraiserverauth.Model{ID: "no-thinking"})
	if plain == nil || plain.Thinking != nil {
		t.Fatalf("non-thinking model carries metadata: %+v", plain.Thinking)
	}
}

// TestLobsteraiServerCatalogueProbePublishesFetchedModels exercises the whole
// probe path (fetch -> publish) against a mock, since the live catalogue is the
// authoritative model source for this provider.
func TestLobsteraiServerCatalogueProbePublishesFetchedModels(t *testing.T) {
	server := newLobsteraiServerCatalogueServer(t)
	defer server.Close()

	manager := coreauth.NewManager(nil, &coreauth.RoundRobinSelector{}, nil)
	service := &Service{cfg: &config.Config{}, coreManager: manager}

	auth := &coreauth.Auth{
		ID:       "lobsterai-probe-publish",
		Provider: constant.LobsteraiServer,
		Attributes: map[string]string{
			"auth_kind":    "oauth",
			"access_token": "acct-token",
			"base_url":     server.URL,
			"backend_url":  server.URL,
		},
	}
	if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
		t.Fatalf("register auth: %v", errRegister)
	}
	t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })

	service.executeLobsteraiServerCatalogueProbe(context.Background(), auth.Clone(), constant.LobsteraiServer, "probe-key", auth.RegistrationEpoch)

	// Assert on the client's own registration: LookupModelInfo would also see the
	// global static catalog, which does not reflect this account's entitlements.
	registered := map[string]bool{}
	for _, model := range registry.GetGlobalRegistry().GetModelsForClient(auth.ID) {
		registered[model.ID] = true
	}
	if !registered["deepseek-v4-pro"] {
		t.Fatalf("deepseek-v4-pro was not registered by the probe: %v", registered)
	}
	// An inaccessible model is filtered upstream (40301), so it must not register.
	if registered["kimi-k3"] {
		t.Fatalf("inaccessible model registered: %v", registered)
	}
}
