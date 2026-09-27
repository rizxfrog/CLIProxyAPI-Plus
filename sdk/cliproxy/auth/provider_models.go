package auth

import (
	"github.com/router-for-me/CLIProxyAPI/v7/internal/util"
	"strings"
)

// providerModelDisabled checks the resolved upstream model, including alias-pool candidates.
func (m *Manager) providerModelDisabled(auth *Auth, upstreamModel string) bool {
	if auth == nil {
		return false
	}
	cfg := m.runtimeConfigSnapshot()
	if cfg == nil || cfg.Home.Enabled {
		return false
	}
	provider := strings.ToLower(strings.TrimSpace(auth.Provider))
	if value := strings.TrimSpace(auth.Attributes["provider_key"]); value != "" {
		provider = strings.ToLower(value)
	}
	if strings.TrimSpace(auth.Attributes["compat_name"]) != "" {
		provider = util.OpenAICompatibleProviderKey(provider)
	}
	model := canonicalModelKey(upstreamModel)
	for _, disabled := range cfg.ProviderModels[provider].Disabled {
		if model == disabled {
			return true
		}
	}
	return false
}
