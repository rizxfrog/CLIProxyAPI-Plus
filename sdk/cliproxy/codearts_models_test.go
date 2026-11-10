package cliproxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

// TestCodeArtsModelsForCataloguePublishesLiveMetadata pins that the live
// catalogue, not the static registry, decides a CodeArts model's limits.
//
// The static list is keyed by model id across every provider, so a
// coincidentally equal id would otherwise inherit limits CodeArts never
// declared — the account would be told it can send a context the gateway then
// rejects.
func TestCodeArtsModelsForCataloguePublishesLiveMetadata(t *testing.T) {
	auth := &coreauth.Auth{
		ID:         "codearts-catalogue-metadata",
		Provider:   constant.CodeArts,
		Attributes: map[string]string{"auth_kind": coreauth.AuthKindOAuth},
	}
	service := &Service{cfg: &config.Config{}}

	catalogue := []*codeartsauth.Model{
		{ID: "GLM-5.2", Name: "GLM-5.2", ContextWindow: 202752, OutputContextWindow: 65536},
		{ID: "vision-live", Name: "Vision Live", ContextWindow: 128000, SupportsImages: true},
	}
	models := service.codeartsModelsForCatalogue(auth, catalogue)
	byID := make(map[string]*ModelInfo, len(models))
	for _, model := range models {
		byID[model.ID] = model
	}

	glm, ok := byID["GLM-5.2"]
	if !ok {
		t.Fatalf("GLM-5.2 missing from %v", byID)
	}
	if glm.ContextLength != 202752 {
		t.Fatalf("context = %d, want the live value 202752", glm.ContextLength)
	}
	if glm.MaxCompletionTokens != 65536 {
		t.Fatalf("max completion = %d, want 65536", glm.MaxCompletionTokens)
	}
	if glm.Type != constant.CodeArts || glm.OwnedBy != constant.CodeArts {
		t.Fatalf("model not attributed to codearts: %+v", glm)
	}

	vision, ok := byID["vision-live"]
	if !ok {
		t.Fatalf("vision-live missing from %v", byID)
	}
	if len(vision.SupportedInputModalities) != 2 {
		t.Fatalf("image capability lost: %+v", vision.SupportedInputModalities)
	}
}

// TestCodeArtsModelsForCatalogueRegistrationRoundTrip pins that the mapped
// catalogue actually reaches the global registry, which is where the model list
// endpoint and selection read it from.
func TestCodeArtsModelsForCatalogueRegistrationRoundTrip(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	auth := &coreauth.Auth{
		ID:         "codearts-catalogue-roundtrip",
		Provider:   constant.CodeArts,
		Attributes: map[string]string{"auth_kind": coreauth.AuthKindOAuth},
	}
	service := &Service{cfg: &config.Config{}}

	catalogue := []*codeartsauth.Model{
		{ID: "openpangu-2.0-pro", Name: "OpenPangu-2.0-Pro", ContextWindow: 524288},
	}
	models := service.codeartsModelsForCatalogue(auth, catalogue)
	if len(models) != 1 {
		t.Fatalf("models = %d, want 1", len(models))
	}
	service.registerResolvedModelsForAuth(auth, constant.CodeArts, models)
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	info := registry.LookupModelInfo("openpangu-2.0-pro", constant.CodeArts)
	if info == nil {
		t.Fatal("openpangu-2.0-pro was not registered")
	}
	if info.ContextLength != 524288 {
		t.Fatalf("context = %d, want 524288", info.ContextLength)
	}
}

// TestCodeArtsCatalogueModelInfoSkipsBlankID pins that a malformed entry cannot
// register a nameless model.
func TestCodeArtsCatalogueModelInfoSkipsBlankID(t *testing.T) {
	if info := codeartsCatalogueModelInfo(&codeartsauth.Model{ID: "   "}); info != nil {
		t.Fatalf("blank id produced %+v", info)
	}
	if info := codeartsCatalogueModelInfo(nil); info != nil {
		t.Fatalf("nil entry produced %+v", info)
	}
}

// TestCodeArtsCatalogueModelInfoPublishesDeclaredThinking pins that a model the
// catalogue marks as reasoning-capable gets a level list.
//
// Without model-level thinking metadata the shared pipeline drops
// reasoning_effort from the upstream body, so the client's choice silently never
// arrives.
func TestCodeArtsCatalogueModelInfoPublishesDeclaredThinking(t *testing.T) {
	info := codeartsCatalogueModelInfo(&codeartsauth.Model{ID: "GLM-5.2", ThinkLevel: "high"})
	if info == nil || info.Thinking == nil {
		t.Fatalf("thinking metadata missing: %+v", info)
	}
	if len(info.Thinking.Levels) == 0 {
		t.Fatal("a reasoning model must publish its level list")
	}

	plain := codeartsCatalogueModelInfo(&codeartsauth.Model{ID: "GLM-5.1"})
	if plain == nil || plain.Thinking != nil {
		t.Fatalf("a model without think_level must not carry metadata: %+v", plain.Thinking)
	}
}

// TestCodeArtsProbeSkipsExplicitlyConfiguredModels pins that an operator's
// pinned model list is not replaced by a discovered catalogue.
func TestCodeArtsProbeSkipsExplicitlyConfiguredModels(t *testing.T) {
	auth := &coreauth.Auth{
		ID:       "codearts-catalogue-pinned",
		Provider: constant.CodeArts,
		Attributes: map[string]string{
			"auth_kind": coreauth.AuthKindOAuth,
			"api_key":   "AK_PIN",
			"base_url":  codeartsauth.InferHubBaseURL,
		},
	}
	service := &Service{cfg: &config.Config{
		CodeArtsKey: []config.CodeArtsKey{{
			APIKey:  "AK_PIN",
			BaseURL: codeartsauth.InferHubBaseURL,
			Models:  []config.CodeArtsModel{{Name: "pinned-only"}},
		}},
	}}
	// A catalogue fetch would fail here because no upstream is reachable; the
	// point is that it must not even be attempted.
	service.executeCodeArtsCatalogueProbe(
		context.Background(), auth, constant.CodeArts, codeartsCatalogueKey(auth), auth.RegistrationEpoch,
	)
	if entry := service.resolveConfigCodeArtsKey(auth); entry == nil || len(entry.Models) != 1 {
		t.Fatalf("config entry was disturbed: %+v", entry)
	}
}

// TestCodeArtsCatalogueProbeEndToEnd pins the whole discovery path: a probe
// against a serving upstream must publish the account's live models into the
// global registry, which is what the model list endpoint and selection read.
//
// This is the behaviour the feature exists for: before it, a CodeArts account
// was pinned to the compiled-in static list, so a model the vendor added or
// removed was invisible until a rebuild.
func TestCodeArtsCatalogueProbeEndToEnd(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent-center/agents/useragents":
			_, _ = w.Write([]byte(`{"agents":[{"agent_id":"code-agent-1","agent_name":"CodeAgent","alias":{"alias_zh_cn":"智能体"},"show_in_ide":true}]}`))
		case "/v1/agent-center/agents/detail":
			if got := r.URL.Query().Get("agent_id"); got != "code-agent-1" {
				t.Errorf("agent_id = %q", got)
			}
			_, _ = w.Write([]byte(`{"gpts":{"models":[{
				"model_name":"live-model","model_type":"builtin","think_level":"high",
				"model_parameters":{"model_id":"live-model","context_window":123456,"output_context_window":4321,"supports_images":true}
			}]}}`))
		case "/api/v1/gateway/config":
			_, _ = w.Write([]byte(`{"error_code":"0000","result":{"base_url":"","models":[{"model_id":"live-benefit","model_name":"live-benefit"}]}}`))
		default:
			// Catalogue probing tolerates unknown endpoints, so answer 404 and
			// let the merge skip that source.
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	reg := registry.GetGlobalRegistry()
	auth := &coreauth.Auth{
		ID:       "codearts-catalogue-e2e",
		Provider: constant.CodeArts,
		Attributes: map[string]string{
			"auth_kind":       coreauth.AuthKindOAuth,
			"api_key":         "AK_E2E",
			"secret_key":      "SK_E2E",
			"security_token":  "ST_E2E",
			"snap_engine_url": upstream.URL,
			"benefit_api_url": upstream.URL,
		},
	}
	service := &Service{cfg: &config.Config{}}
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	catalogue, err := service.fetchCodeArtsCatalogue(context.Background(), auth)
	if err != nil {
		t.Fatalf("fetch catalogue: %v", err)
	}
	if len(catalogue) == 0 {
		t.Fatal("catalogue is empty")
	}

	models := service.codeartsModelsForCatalogue(auth, catalogue)
	service.registerResolvedModelsForAuth(auth, constant.CodeArts, models)

	info := registry.LookupModelInfo("live-model", constant.CodeArts)
	if info == nil {
		t.Fatalf("live-model was not published; got %d models", len(models))
	}
	if info.ContextLength != 123456 || info.MaxCompletionTokens != 4321 {
		t.Fatalf("live limits not published: %+v", info)
	}
	if len(info.SupportedInputModalities) != 2 {
		t.Fatalf("image capability not published: %+v", info.SupportedInputModalities)
	}
	if info.Thinking == nil || len(info.Thinking.Levels) == 0 {
		t.Fatalf("thinking levels not published: %+v", info.Thinking)
	}

	// The benefit model must have registered its routing channel, otherwise a
	// request for it would omit maas_type and be rejected upstream.
	if channel, ok := codeartsauth.BenefitRouteFor("live-benefit"); !ok || channel != codeartsauth.BenefitChannelBenefit {
		t.Fatalf("benefit route = (%q, %v)", channel, ok)
	}
}
