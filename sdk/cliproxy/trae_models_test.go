package cliproxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

// TestTraeCatalogueProbeRepublishesLiveModels pins the live-fetch contract: the
// gateway's /api/ide/v1/get_detail_param replaces the static catalogue for the
// credential, keeps the alias/prefix pipeline, and routes the config_name the
// chat endpoint expects.
func TestTraeCatalogueProbeRepublishesLiveModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ide/v1/get_detail_param" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Cloud-IDE-JWT live-token" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte(`{"code":0,"config_info_list":[
			{"config_name":"glm-5.3","config_switch":true,"context_window_tokens":{"dev":128000,"max":200000},"max_tokens":65536,
			 "display_config":{"display_name":"GLM-5.3","multimodal":true}},
			{"config_name":"kimi-k3","config_switch":true,"context_window_size":{"default":192000,"max":256000},
			 "display_config":{"display_name":"Kimi-K3"}}
		]}`))
	}))
	defer server.Close()

	auth := &coreauth.Auth{
		ID:       t.Name(),
		Provider: "trae",
		Attributes: map[string]string{
			"api_key":   "live-token",
			"base_url":  server.URL,
			"auth_kind": coreauth.AuthKindOAuth,
			"uid":       "42",
		},
	}
	t.Cleanup(func() { GlobalModelRegistry().UnregisterClient(auth.ID) })

	s := &Service{cfg: &config.Config{}}
	catalogue, errFetch := s.fetchTraeCatalogue(context.Background(), auth)
	if errFetch != nil {
		t.Fatalf("fetchTraeCatalogue: %v", errFetch)
	}
	if len(catalogue.Models) != 2 {
		t.Fatalf("catalogue = %+v", catalogue.Models)
	}
	models := s.traeModelsForCatalogue(auth, catalogue.Models)
	if len(models) != 2 {
		t.Fatalf("models = %+v", models)
	}
	byID := map[string]*ModelInfo{}
	for _, model := range models {
		byID[model.ID] = model
	}
	glm := byID["glm-5.3"]
	if glm == nil {
		t.Fatalf("glm-5.3 missing from live catalogue: %+v", byID)
	}
	if glm.DisplayName != "GLM-5.3" || glm.ContextLength != 200000 || glm.MaxCompletionTokens != 65536 {
		t.Fatalf("glm-5.3 = %+v", glm)
	}
	if glm.Type != "trae" || glm.OwnedBy != "trae" {
		t.Fatalf("glm-5.3 routing metadata = %+v", glm)
	}
	if byID["kimi-k3"] == nil {
		t.Fatal("kimi-k3 missing from live catalogue")
	}

	// Publishing must reach the registry so the model list and routing agree.
	s.registerResolvedModelsForAuth(auth, "trae", models)
	for _, id := range []string{"glm-5.3", "kimi-k3"} {
		if !registry.GetGlobalRegistry().ClientSupportsModel(auth.ID, id) {
			t.Fatalf("registry does not route live model %q", id)
		}
	}
}

// TestTraeCatalogueKeepsFallbackWhenProbeFails pins that a failed fetch leaves
// the credential's current models untouched and schedules a retry.
func TestTraeCatalogueKeepsFallbackWhenProbeFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":1001,"message":"unauthenticated","config_info_list":[]}`))
	}))
	defer server.Close()

	auth := &coreauth.Auth{
		ID:         t.Name(),
		Provider:   "trae",
		Attributes: map[string]string{"api_key": "expired", "base_url": server.URL},
	}
	t.Cleanup(func() { GlobalModelRegistry().UnregisterClient(auth.ID) })

	s := &Service{cfg: &config.Config{}}
	if _, errFetch := s.fetchTraeCatalogue(context.Background(), auth); errFetch == nil {
		t.Fatal("expected an unauthenticated catalogue to fail")
	}

	key := traeCatalogueKey(auth)
	s.deferTraeCatalogueRetry(key)
	traeCatalogueMu.RLock()
	entry := traeCatalogue[key]
	traeCatalogueMu.RUnlock()
	if entry.nextRetry.IsZero() {
		t.Fatal("failed probe must schedule a retry")
	}
	if !entry.expiresAt.IsZero() {
		t.Fatal("failed probe must not mark the catalogue fresh")
	}
}

// TestTraeRegistrationTriggersLiveCatalogueProbe pins the wiring: registering a
// trae credential starts the gateway fetch and republishes the live rows in the
// registry, which is what the management "supported models" view reads.
func TestTraeRegistrationTriggersLiveCatalogueProbe(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte(`{"code":0,"config_info_list":[
			{"config_name":"glm-5.3","config_switch":true,"context_window_tokens":{"max":200000},
			 "display_config":{"display_name":"GLM-5.3"}}
		]}`))
	}))
	defer server.Close()

	manager := coreauth.NewManager(nil, nil, nil)
	svc := &Service{cfg: &config.Config{}, coreManager: manager}
	auth := &coreauth.Auth{
		ID:       t.Name(),
		Provider: "trae",
		Attributes: map[string]string{
			"api_key":   "live-token",
			"base_url":  server.URL,
			"auth_kind": coreauth.AuthKindOAuth,
		},
	}
	registered, errRegister := manager.Register(context.Background(), auth)
	if errRegister != nil {
		t.Fatalf("Register: %v", errRegister)
	}
	t.Cleanup(func() {
		GlobalModelRegistry().UnregisterClient(registered.ID)
		traeProbeWg.Wait()
	})
	traeProbeWg.Wait()

	svc.registerModelsForAuth(context.Background(), registered)
	traeProbeWg.Wait()

	if atomic.LoadInt32(&hits) == 0 {
		t.Fatal("registering a trae credential did not fetch the live catalogue")
	}
	if !registry.GetGlobalRegistry().ClientSupportsModel(registered.ID, "glm-5.3") {
		t.Fatal("live catalogue model was not published to the registry")
	}
}
