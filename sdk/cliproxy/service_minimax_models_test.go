package cliproxy

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/config"
)

func assertMinimaxModelsRegistered(t *testing.T, provider, baseURL string) {
	t.Helper()
	authID := provider + "-oauth-models"
	modelRegistry := registry.GetGlobalRegistry()
	modelRegistry.UnregisterClient(authID)
	t.Cleanup(func() {
		modelRegistry.UnregisterClient(authID)
	})

	service := &Service{cfg: &config.Config{}}
	auth := &coreauth.Auth{
		ID:       authID,
		Provider: provider,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"base_url":                 baseURL,
		},
		Metadata: map[string]any{"access_token": "managed-token"},
	}

	service.registerModelsForAuth(context.Background(), auth)

	got := modelRegistry.GetModelsForClient(authID)
	want := registry.GetMinimaxModels()
	if len(got) != len(want) {
		t.Fatalf("%s registered models = %d, want %d", provider, len(got), len(want))
	}
	seen := map[string]bool{}
	for _, m := range got {
		seen[m.ID] = true
	}
	for _, id := range []string{"MiniMax-M3", "MiniMax-M2.7", "MiniMax-M2.7-highspeed"} {
		if !seen[id] {
			t.Fatalf("%s missing registered model %q", provider, id)
		}
	}
}

func TestRegisterModelsForAuthMinimaxOAuth(t *testing.T) {
	assertMinimaxModelsRegistered(t, constant.Minimax, "https://agent.minimax.io/mavis/api/v1/llm")
}

func TestRegisterModelsForAuthMinimaxCNOAuth(t *testing.T) {
	assertMinimaxModelsRegistered(t, constant.MinimaxCN, "https://agent.minimax.cn/mavis/api/v1/llm")
}
