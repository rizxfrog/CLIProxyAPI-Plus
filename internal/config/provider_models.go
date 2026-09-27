package config

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
)

// ProviderModels is an incremental overlay, never a snapshot of the default catalog.
type ProviderModels struct {
	Disabled []string        `yaml:"disabled" json:"disabled"`
	Custom   []ProviderModel `yaml:"custom" json:"custom"`
}

// ProviderModel describes a real upstream model, not a routing alias.
type ProviderModel struct {
	ID                  string                    `yaml:"id" json:"id"`
	DisplayName         string                    `yaml:"display-name,omitempty" json:"display-name,omitempty"`
	ContextLength       *int                      `yaml:"context-length,omitempty" json:"context-length,omitempty"`
	MaxCompletionTokens *int                      `yaml:"max-completion-tokens,omitempty" json:"max-completion-tokens,omitempty"`
	Thinking            *registry.ThinkingSupport `yaml:"thinking,omitempty" json:"thinking,omitempty"`
}

func validProviderModelID(id string) bool {
	return id != "" && !strings.ContainsAny(id, "*?()") && strings.IndexFunc(id, unicode.IsSpace) < 0 && strings.IndexFunc(id, unicode.IsControl) < 0
}

// NormalizeProviderModels validates the complete overlay without silently dropping mistakes.
func NormalizeProviderModels(entries map[string]ProviderModels) (map[string]ProviderModels, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	out := make(map[string]ProviderModels, len(entries))
	for rawProvider, entry := range entries {
		provider := strings.ToLower(strings.TrimSpace(rawProvider))
		if !validProviderModelID(provider) || strings.Contains(provider, "/") {
			return nil, fmt.Errorf("invalid provider %q", rawProvider)
		}
		if _, exists := out[provider]; exists {
			return nil, fmt.Errorf("duplicate provider %q", provider)
		}
		normalized := ProviderModels{Disabled: []string{}, Custom: []ProviderModel{}}
		seen := map[string]bool{}
		for _, rawID := range entry.Disabled {
			id := strings.TrimSpace(rawID)
			if !validProviderModelID(id) {
				return nil, fmt.Errorf("invalid disabled model ID for %s", provider)
			}
			if !seen[id] {
				normalized.Disabled = append(normalized.Disabled, id)
				seen[id] = true
			}
		}
		sort.Strings(normalized.Disabled)
		seen = map[string]bool{}
		for _, model := range entry.Custom {
			model.ID = strings.TrimSpace(model.ID)
			model.DisplayName = strings.TrimSpace(model.DisplayName)
			if !validProviderModelID(model.ID) {
				return nil, fmt.Errorf("invalid custom model ID for %s", provider)
			}
			if seen[model.ID] {
				return nil, fmt.Errorf("duplicate custom model %q", model.ID)
			}
			seen[model.ID] = true
			if model.ContextLength != nil && *model.ContextLength <= 0 {
				return nil, fmt.Errorf("context-length must be positive")
			}
			if model.MaxCompletionTokens != nil && *model.MaxCompletionTokens <= 0 {
				return nil, fmt.Errorf("max-completion-tokens must be positive")
			}
			if model.Thinking != nil && (model.Thinking.Min < 0 || model.Thinking.Max < 0 || model.Thinking.Max < model.Thinking.Min) {
				return nil, fmt.Errorf("invalid thinking budget range")
			}
			normalized.Custom = append(normalized.Custom, model)
		}
		out[provider] = normalized
	}
	return out, nil
}

// MergeProviderModels preserves upstream metadata unless the user explicitly overrides it.
// allowAdd is false for accounts with explicit model whitelists.
func MergeProviderModels(models []*registry.ModelInfo, provider string, overlay ProviderModels, allowAdd bool) []*registry.ModelInfo {
	out := make([]*registry.ModelInfo, 0, len(models)+len(overlay.Custom))
	custom := make(map[string]ProviderModel, len(overlay.Custom))
	for _, model := range overlay.Custom {
		custom[model.ID] = model
	}
	seen := map[string]bool{}
	apply := func(base *registry.ModelInfo, model ProviderModel) *registry.ModelInfo {
		clone := *base
		if model.DisplayName != "" {
			clone.DisplayName = model.DisplayName
		}
		if model.ContextLength != nil {
			clone.ContextLength = *model.ContextLength
			clone.MaxContextLength = *model.ContextLength
		}
		if model.MaxCompletionTokens != nil {
			clone.MaxCompletionTokens = *model.MaxCompletionTokens
		}
		if model.Thinking != nil {
			thinking := *model.Thinking
			thinking.Levels = append([]string(nil), thinking.Levels...)
			clone.Thinking = &thinking
			clone.ExplicitThinking = true
		}
		return &clone
	}
	for _, base := range models {
		if base == nil {
			continue
		}
		id := base.MetadataModelID
		if id == "" {
			id = base.ID
		}
		seen[id] = true
		if model, ok := custom[id]; ok {
			out = append(out, apply(base, model))
		} else {
			out = append(out, base)
		}
	}
	if allowAdd {
		for _, model := range overlay.Custom {
			if seen[model.ID] {
				continue
			}
			seen[model.ID] = true
			base := &registry.ModelInfo{ID: model.ID, Object: "model", OwnedBy: provider, Type: provider, DisplayName: model.ID}
			out = append(out, apply(base, model))
		}
	}
	return out
}

// FilterProviderModels operates on upstream metadata IDs so aliases and prefixes cannot bypass a disable.
func FilterProviderModels(models []*registry.ModelInfo, overlay ProviderModels) []*registry.ModelInfo {
	disabled := make(map[string]bool, len(overlay.Disabled))
	for _, id := range overlay.Disabled {
		disabled[id] = true
	}
	out := make([]*registry.ModelInfo, 0, len(models))
	for _, model := range models {
		if model == nil {
			continue
		}
		id := model.MetadataModelID
		if id == "" {
			id = model.ID
		}
		if !disabled[id] {
			out = append(out, model)
		}
	}
	return out
}
