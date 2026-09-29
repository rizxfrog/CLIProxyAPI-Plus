package pluginhost

import (
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
)

// ModelCatalog returns the unfiltered plugin catalog, including disabled models.
func (h *Host) ModelCatalog() map[string][]*registry.ModelInfo {
	out := map[string][]*registry.ModelInfo{}
	for _, registration := range h.snapshotModelRegistrations() {
		appendModelsForProvider(out, registration.provider, registration.models)
	}
	return out
}

func (h *Host) registerProviderModelClient(target modelRegistry, registration modelClientRegistration) {
	models := registration.models
	cfg := h.currentRuntimeConfig()
	if cfg != nil && !cfg.Home.Enabled {
		overlay := cfg.ProviderModels[registration.provider]
		models = config.MergeProviderModels(models, registration.provider, overlay, true)
		models = config.FilterProviderModels(models, overlay)
	}
	if len(models) == 0 {
		target.UnregisterClient(registration.clientID)
		return
	}
	target.RegisterClient(registration.clientID, registration.provider, models)
}
