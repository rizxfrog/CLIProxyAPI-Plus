package pluginhost

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
)

func TestProviderModelsSyntheticPluginRegistrations(t *testing.T) {
	host := New()
	host.runtimeConfig = &config.Config{ProviderModels: map[string]config.ProviderModels{"plugin-provider": {Disabled: []string{"old"}, Custom: []config.ProviderModel{{ID: "new"}}}}}
	target := newFakeModelRegistry()
	registration := modelClientRegistration{clientID: "plugin:sample", provider: "plugin-provider", models: []*registry.ModelInfo{{ID: "old"}}}
	for i := 0; i < 2; i++ {
		host.registerProviderModelClient(target, registration)
		models := target.clients[registration.clientID].models
		if len(models) != 1 || models[0].ID != "new" {
			t.Fatalf("plugin refresh bypassed overlay: %#v", models)
		}
	}
	host.runtimeConfig = &config.Config{}
	host.registerProviderModelClient(target, registration)
	if target.clients[registration.clientID].models[0].ID != "old" {
		t.Fatal("reset failed")
	}
	host.runtimeConfig = &config.Config{ProviderModels: map[string]config.ProviderModels{"plugin-provider": {Disabled: []string{"old"}}}}
	host.registerProviderModelClient(target, registration)
	if _, ok := target.clients[registration.clientID]; ok {
		t.Fatal("empty provider remained registered")
	}
}
