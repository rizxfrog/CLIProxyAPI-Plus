package cliproxy

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

func TestRegisterModelsForAuthCodeBuddyCNOAuthIgnoresAPIKeyModels(t *testing.T) {
	const authID = "codebuddy-cn-oauth-models"
	modelRegistry := registry.GetGlobalRegistry()
	modelRegistry.UnregisterClient(authID)
	t.Cleanup(func() {
		modelRegistry.UnregisterClient(authID)
	})

	service := &Service{cfg: &config.Config{
		CodeBuddyCNKey: []config.CodeBuddyCNKey{
			{
				APIKey:  "api-key",
				BaseURL: "https://copilot.tencent.com/v2",
				Models: []config.CodeBuddyCNModel{
					{Name: "hy3"},
				},
			},
		},
	}}
	auth := &coreauth.Auth{
		ID:       authID,
		Provider: constant.CodeBuddyCN,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"base_url":                 "https://copilot.tencent.com/v2",
		},
	}

	service.registerModelsForAuth(context.Background(), auth)

	got := modelRegistry.GetModelsForClient(authID)
	want := registry.GetCodeBuddyCNModels()
	if len(got) != len(want) {
		t.Fatalf("registered models = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID {
			t.Fatalf("registered model[%d] = %q, want %q", i, got[i].ID, want[i].ID)
		}
	}
}

func TestRegisterModelsForAuthCodeBuddyAIOAuthIgnoresAPIKeyModels(t *testing.T) {
	const authID = "codebuddy-ai-oauth-models"
	modelRegistry := registry.GetGlobalRegistry()
	modelRegistry.UnregisterClient(authID)
	t.Cleanup(func() {
		modelRegistry.UnregisterClient(authID)
	})

	service := &Service{cfg: &config.Config{
		CodeBuddyAIKey: []config.CodeBuddyAIKey{
			{
				APIKey:  "api-key",
				BaseURL: "https://www.codebuddy.ai/v2",
				Models: []config.CodeBuddyAIModel{
					{Name: "gpt-5.5"},
				},
			},
		},
	}}
	auth := &coreauth.Auth{
		ID:       authID,
		Provider: constant.CodeBuddyAI,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"base_url":                 "https://www.codebuddy.ai/v2",
		},
	}

	service.registerModelsForAuth(context.Background(), auth)

	got := modelRegistry.GetModelsForClient(authID)
	want := registry.GetCodeBuddyAIModels()
	if len(got) != len(want) {
		t.Fatalf("registered models = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID {
			t.Fatalf("registered model[%d] = %q, want %q", i, got[i].ID, want[i].ID)
		}
	}
}

func TestRegisterModelsForAuthCodeBuddyAIAPIKeyUsesConfiguredModels(t *testing.T) {
	const authID = "codebuddy-ai-apikey-models"
	modelRegistry := registry.GetGlobalRegistry()
	modelRegistry.UnregisterClient(authID)
	t.Cleanup(func() {
		modelRegistry.UnregisterClient(authID)
	})

	service := &Service{cfg: &config.Config{
		CodeBuddyAIKey: []config.CodeBuddyAIKey{
			{
				APIKey:  "cbai-api-key",
				BaseURL: "https://www.codebuddy.ai/v2",
				Models: []config.CodeBuddyAIModel{
					{Name: "gpt-5.5", Alias: "gpt-latest"},
				},
			},
		},
	}}
	auth := &coreauth.Auth{
		ID:       authID,
		Provider: constant.CodeBuddyAI,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindAPIKey,
			"api_key":                  "cbai-api-key",
			"base_url":                 "https://www.codebuddy.ai/v2",
		},
	}

	service.registerModelsForAuth(context.Background(), auth)

	got := modelRegistry.GetModelsForClient(authID)
	if len(got) != 1 {
		t.Fatalf("registered models = %d, want 1", len(got))
	}
	if got[0].ID != "gpt-latest" {
		t.Fatalf("registered model = %q, want gpt-latest", got[0].ID)
	}
}
