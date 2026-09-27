package cliproxy

import (
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"strings"
)

// applyProviderModelOverlay runs before aliases/prefixes and after all catalog sources.
func (s *Service) applyProviderModelOverlay(provider string, models []*ModelInfo, allowAdd bool, excluded []string) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg != nil && !cfg.Home.Enabled {
		overlay := cfg.ProviderModels[provider]
		models = config.MergeProviderModels(models, provider, overlay, allowAdd)
		models = config.FilterProviderModels(models, overlay)
	}
	return applyExcludedModels(models, excluded)
}

// Plugins can handle native accounts before the normal provider switch. Preserve
// those accounts' explicit config whitelists instead of broadening them.
func (s *Service) hasExplicitProviderModels(auth *coreauth.Auth, provider string) bool {
	if s.cfg == nil {
		return false
	}
	switch provider {
	case "gemini":
		if entry := s.resolveConfigGeminiKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "gemini-interactions":
		if entry := s.resolveConfigInteractionsKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "vertex":
		if entry := s.resolveConfigVertexCompatKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "claude":
		if entry := s.resolveConfigClaudeKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "codex":
		if entry := s.resolveConfigCodexKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "codebuddy-cn":
		if entry := s.resolveConfigCodeBuddyCNKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "codebuddy-ai":
		if entry := s.resolveConfigCodeBuddyAIKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "cline":
		if entry := s.resolveConfigClineKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "deepseek-web":
		if entry := s.resolveConfigDeepSeekWebKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "codearts":
		if entry := s.resolveConfigCodeArtsKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "xiaohuanxiong":
		if entry := s.resolveConfigXiaohuanxiongKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "xai":
		if entry := s.resolveConfigXAIKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "trae":
		if entry := s.resolveConfigTraeKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "qoder-cn":
		if entry := s.resolveConfigQoderCNKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "qoder-ai":
		if entry := s.resolveConfigQoderAIKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	case "meta":
		if entry := s.resolveConfigMetaKey(auth); entry != nil {
			return len(entry.Models) > 0
		}
	}
	if _, name, ok := openAICompatInfoFromAuth(auth); ok {
		if entry := configEntryForAuthIndex(auth, s.cfg.OpenAICompatibility); entry != nil {
			return len(entry.Models) > 0
		}
		for _, entry := range s.cfg.OpenAICompatibility {
			if strings.EqualFold(entry.Name, name) {
				return len(entry.Models) > 0
			}
		}
	}
	return false
}
