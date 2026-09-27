package auth

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

func TestProviderModelsDisableResolvedAliasPoolTargets(t *testing.T) {
	executor := &openAICompatPoolExecutor{id: openAICompatPoolProviderKey}
	models := []config.OpenAICompatibilityModel{{Name: "old", Alias: "shared"}, {Name: "kept", Alias: "shared"}}
	manager := newOpenAICompatPoolTestManager(t, "shared", models, executor)
	cfg := &config.Config{OpenAICompatibility: []config.OpenAICompatibility{{Name: "pool", Models: models}}, ProviderModels: map[string]config.ProviderModels{openAICompatPoolProviderKey: {Disabled: []string{"old"}}}}
	manager.SetConfig(cfg)
	for i := 0; i < 3; i++ {
		if _, err := manager.Execute(context.Background(), []string{openAICompatPoolProviderKey}, cliproxyexecutor.Request{Model: "shared"}, cliproxyexecutor.Options{}); err != nil {
			t.Fatal(err)
		}
	}
	for _, model := range executor.ExecuteModels() {
		if model != "kept" {
			t.Fatalf("executed disabled upstream: %s", model)
		}
	}
	auth := manager.List()[0]
	if got := manager.filterExecutionModels(auth, "shared", []string{"old(high)", "kept"}, true); len(got) != 1 || got[0] != "kept" {
		t.Fatalf("suffix bypass: %v", got)
	}
	cfg.Home.Enabled = true
	manager.SetConfig(cfg)
	if manager.providerModelDisabled(auth, "old") {
		t.Fatal("local overlay enforced in Home mode")
	}
}
