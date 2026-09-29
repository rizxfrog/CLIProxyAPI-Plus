package management

import (
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	"strings"
)

func configuredProviderCatalog(cfg *config.Config) map[string][]*registry.ModelInfo {
	out := map[string][]*registry.ModelInfo{}
	for _, entry := range cfg.GeminiKey {
		out["gemini"] = append(out["gemini"], configuredCatalogModels(entry.Models, "gemini")...)
	}
	for _, entry := range cfg.InteractionsKey {
		out["gemini-interactions"] = append(out["gemini-interactions"], configuredCatalogModels(entry.Models, "gemini-interactions")...)
	}
	for _, entry := range cfg.CodexKey {
		out["codex"] = append(out["codex"], configuredCatalogModels(entry.Models, "codex")...)
	}
	for _, entry := range cfg.XAIKey {
		out["xai"] = append(out["xai"], configuredCatalogModels(entry.Models, "xai")...)
	}
	for _, entry := range cfg.CodeBuddyCNKey {
		out["codebuddy-cn"] = append(out["codebuddy-cn"], configuredCatalogModels(entry.Models, "codebuddy-cn")...)
	}
	for _, entry := range cfg.CodeBuddyAIKey {
		out["codebuddy-ai"] = append(out["codebuddy-ai"], configuredCatalogModels(entry.Models, "codebuddy-ai")...)
	}
	for _, entry := range cfg.DeepSeekWebKey {
		out["deepseek-web"] = append(out["deepseek-web"], configuredCatalogModels(entry.Models, "deepseek-web")...)
	}
	for _, entry := range cfg.XiaohuanxiongKey {
		out["xiaohuanxiong"] = append(out["xiaohuanxiong"], configuredCatalogModels(entry.Models, "xiaohuanxiong")...)
	}
	for _, entry := range cfg.CodeArtsKey {
		out["codearts"] = append(out["codearts"], configuredCatalogModels(entry.Models, "codearts")...)
	}
	for _, entry := range cfg.TraeKey {
		out["trae"] = append(out["trae"], configuredCatalogModels(entry.Models, "trae")...)
	}
	for _, entry := range cfg.ClineKey {
		out["cline"] = append(out["cline"], configuredCatalogModels(entry.Models, "cline")...)
	}
	for _, entry := range cfg.QoderCNKey {
		out["qoder-cn"] = append(out["qoder-cn"], configuredCatalogModels(entry.Models, "qoder-cn")...)
	}
	for _, entry := range cfg.QoderAIKey {
		out["qoder-ai"] = append(out["qoder-ai"], configuredCatalogModels(entry.Models, "qoder-ai")...)
	}
	for _, entry := range cfg.MetaKey {
		out["meta"] = append(out["meta"], configuredCatalogModels(entry.Models, "meta")...)
	}
	for _, entry := range cfg.ClaudeKey {
		out["claude"] = append(out["claude"], configuredCatalogModels(entry.Models, "claude")...)
	}
	for _, entry := range cfg.VertexCompatAPIKey {
		out["vertex"] = append(out["vertex"], configuredCatalogModels(entry.Models, "vertex")...)
	}
	for _, entry := range cfg.OpenAICompatibility {
		provider := util.OpenAICompatibleProviderKey(entry.Name)
		out[provider] = append(out[provider], configuredCatalogModels(entry.Models, provider)...)
	}
	return out
}

func configuredCatalogModels[T interface {
	GetName() string
	GetDisplayName() string
	GetThinking() *registry.ThinkingSupport
}](models []T, provider string) []*registry.ModelInfo {
	out := make([]*registry.ModelInfo, 0, len(models))
	for _, model := range models {
		id := strings.TrimSpace(model.GetName())
		if id == "" {
			continue
		}
		out = append(out, &registry.ModelInfo{ID: id, DisplayName: model.GetDisplayName(), OwnedBy: provider, Type: provider, Object: "model", Thinking: model.GetThinking()})
	}
	return out
}
