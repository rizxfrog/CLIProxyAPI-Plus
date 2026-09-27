package cliproxy

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/config"
)

func TestProviderModelsRegistrationAliasesPrefixesAndRefresh(t *testing.T) {
	cfg := &config.Config{ProviderModels: map[string]config.ProviderModels{"trae": {Disabled: []string{"glm-5.2"}, Custom: []config.ProviderModel{{ID: "glm-5.4"}, {ID: "account-hidden"}}}}, OAuthModelAlias: map[string][]config.OAuthModelAlias{"trae": {{Name: "glm-5.2", Alias: "old-alias", Fork: true}, {Name: "glm-5.4", Alias: "new-alias", Fork: true}}}}
	s := &Service{cfg: cfg}
	auth := &coreauth.Auth{ID: t.Name(), Provider: "trae", Prefix: "team", Attributes: map[string]string{"auth_kind": "oauth", "excluded_models": "account-hidden"}}
	t.Cleanup(func() { GlobalModelRegistry().UnregisterClient(auth.ID) })
	for i := 0; i < 2; i++ {
		s.registerModelsForAuth(context.Background(), auth)
		models := registry.GetGlobalRegistry().GetModelsForClient(auth.ID)
		ids := map[string]bool{}
		for _, model := range models {
			ids[model.ID] = true
		}
		for _, id := range []string{"glm-5.2", "old-alias", "team/glm-5.2", "team/old-alias", "account-hidden", "team/account-hidden"} {
			if ids[id] {
				t.Fatalf("blocked model registered: %s", id)
			}
		}
		for _, id := range []string{"glm-5.4", "new-alias", "team/glm-5.4", "team/new-alias"} {
			if !ids[id] {
				t.Fatalf("missing model: %s (%v)", id, ids)
			}
		}
	}
	// Reset rebinds the current default catalog, rather than a saved snapshot.
	s.cfg = &config.Config{}
	s.registerModelsForAuth(context.Background(), auth)
	if !registry.GetGlobalRegistry().ClientSupportsModel(auth.ID, "glm-5.2") || registry.GetGlobalRegistry().ClientSupportsModel(auth.ID, "glm-5.4") {
		t.Fatal("reset did not restore default catalog")
	}
}

func TestProviderModelsExplicitWhitelistAndProviderIsolation(t *testing.T) {
	cfg := &config.Config{
		ProviderModels: map[string]config.ProviderModels{"trae": {Disabled: []string{"upstream-old"}, Custom: []config.ProviderModel{{ID: "not-whitelisted"}, {ID: "upstream-kept", DisplayName: "New title"}}}},
		TraeKey:        []config.TraeKey{{APIKey: "test", Models: []config.TraeModel{{Name: "upstream-old", Alias: "old"}, {Name: "upstream-kept", Alias: "kept"}}}},
	}
	s := &Service{cfg: cfg}
	auth := &coreauth.Auth{ID: t.Name(), Provider: "trae", Attributes: map[string]string{"api_key": "test", "auth_kind": "apikey"}}
	t.Cleanup(func() { GlobalModelRegistry().UnregisterClient(auth.ID) })
	s.registerModelsForAuth(context.Background(), auth)
	models := registry.GetGlobalRegistry().GetModelsForClient(auth.ID)
	if len(models) != 1 || models[0].ID != "kept" || models[0].DisplayName != "New title" {
		t.Fatalf("whitelist/alias broken: %#v", models)
	}
	isolated := s.applyProviderModelOverlay("other", []*ModelInfo{{ID: "upstream-old"}}, true, nil)
	if len(isolated) != 1 {
		t.Fatal("provider disable leaked")
	}
}

func TestProviderModelsCompatWhitelistAndFinalDefense(t *testing.T) {
	cfg := &config.Config{ProviderModels: map[string]config.ProviderModels{"openai-compatible-compat": {Disabled: []string{"old"}, Custom: []config.ProviderModel{{ID: "new"}}}}, OpenAICompatibility: []config.OpenAICompatibility{{Name: "compat", Models: []config.OpenAICompatibilityModel{{Name: "old", Alias: "shared"}, {Name: "kept", Alias: "shared"}}}}}
	s := &Service{cfg: cfg}
	auth := &coreauth.Auth{ID: t.Name(), Provider: "compat", Prefix: "p", Attributes: map[string]string{"api_key": "test", "compat_name": "compat", "provider_key": "compat"}}
	t.Cleanup(func() { GlobalModelRegistry().UnregisterClient(auth.ID) })
	s.registerModelsForAuth(context.Background(), auth)
	models := registry.GetGlobalRegistry().GetModelsForClient(auth.ID)
	if len(models) != 2 {
		t.Fatalf("expected shared + prefixed alias, got %#v", models)
	}
	for _, model := range models {
		if model.MetadataModelID != "kept" {
			t.Fatalf("disabled or unlisted model: %#v", model)
		}
	}
	s.registerResolvedModelsForAuth(auth, "openai-compatible-compat", []*ModelInfo{{ID: "p/old-alias", MetadataModelID: "old"}})
	if len(registry.GetGlobalRegistry().GetModelsForClient(auth.ID)) != 0 {
		t.Fatal("final registration allowed disabled alias")
	}
}

func TestProviderModelsConfigHotReloadWithoutPluginHost(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{ID: t.Name(), Provider: "trae", Attributes: map[string]string{"auth_kind": "oauth"}}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { GlobalModelRegistry().UnregisterClient(auth.ID) })
	s := &Service{cfg: &config.Config{}, coreManager: manager}
	s.registerModelsForAuth(context.Background(), auth)
	cfg := &config.Config{ProviderModels: map[string]config.ProviderModels{"trae": {Disabled: []string{"glm-5.2"}, Custom: []config.ProviderModel{{ID: "glm-5.4"}}}}}
	if !s.applyConfigUpdateWithAuthSynthesis(context.Background(), cfg, false) {
		t.Fatal("reload failed")
	}
	if registry.GetGlobalRegistry().ClientSupportsModel(auth.ID, "glm-5.2") || !registry.GetGlobalRegistry().ClientSupportsModel(auth.ID, "glm-5.4") {
		t.Fatal("reload did not rebind models")
	}
}
