package auth

import (
	"os"
	"strings"
	"sync"
	"time"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

// systemPromptOverrideRules is the runtime view of config SystemPromptOverride.
type systemPromptOverrideRules struct {
	Enabled           bool
	Prompt            string
	PromptFile        string
	Providers         []string
	ExcludedProviders []string
	Models            []string
	Replacements      []internalconfig.PromptReplacementRule
	ToolReplacements  []internalconfig.PromptReplacementRule
}

// systemPromptOverrideRulesFromConfig builds the runtime rule view.
func systemPromptOverrideRulesFromConfig(cfg internalconfig.SystemPromptOverrideConfig) systemPromptOverrideRules {
	return systemPromptOverrideRules{
		Enabled:           cfg.Enabled,
		Prompt:            cfg.Prompt,
		PromptFile:        cfg.PromptFile,
		Providers:         cfg.Providers,
		ExcludedProviders: cfg.ExcludedProviders,
		Models:            cfg.Models,
		Replacements:      cfg.Replacements,
		ToolReplacements:  cfg.ToolDescriptionReplacements,
	}
}

// promptFileCache caches prompt-file content keyed by resolved path, re-reading
// only when mtime or size changed. Cached reads keep the injected text stable
// (prompt-cache friendly) while still picking up edits without a config reload.
var promptFileCache struct {
	sync.Mutex
	entries map[string]promptFileEntry
}

type promptFileEntry struct {
	content string
	modTime time.Time
	size    int64
	readErr bool
	lastTry time.Time
}

const promptFileRetryInterval = 30 * time.Second

// resolvePromptText returns the effective prompt text: inline prompt wins over
// prompt-file; prompt-file is loaded through the mtime cache.
func resolvePromptText(rules systemPromptOverrideRules) string {
	if strings.TrimSpace(rules.Prompt) != "" {
		return rules.Prompt
	}
	path := strings.TrimSpace(rules.PromptFile)
	if path == "" {
		return ""
	}
	return loadPromptFile(path)
}

func loadPromptFile(path string) string {
	promptFileCache.Lock()
	defer promptFileCache.Unlock()
	if promptFileCache.entries == nil {
		promptFileCache.entries = make(map[string]promptFileEntry)
	}
	info, errStat := os.Stat(path)
	if errStat != nil {
		entry, ok := promptFileCache.entries[path]
		if ok && !entry.readErr && time.Since(entry.lastTry) < promptFileRetryInterval {
			// Serve the last good content briefly while the file is missing (e.g.
			// an editor doing an atomic replace) instead of dropping injections.
			return entry.content
		}
		promptFileCache.entries[path] = promptFileEntry{readErr: true, lastTry: time.Now()}
		return ""
	}
	if info.IsDir() {
		return ""
	}
	if entry, ok := promptFileCache.entries[path]; ok && !entry.readErr &&
		entry.modTime.Equal(info.ModTime()) && entry.size == info.Size() {
		return entry.content
	}
	data, errRead := os.ReadFile(path)
	if errRead != nil {
		promptFileCache.entries[path] = promptFileEntry{readErr: true, lastTry: time.Now()}
		return ""
	}
	content := string(data)
	promptFileCache.entries[path] = promptFileEntry{
		content: content,
		modTime: info.ModTime(),
		size:    info.Size(),
		lastTry: time.Now(),
	}
	return content
}

// systemPromptOverrideConfigSnapshot returns the effective override config
// from the manager's current runtime config, or nil when unavailable.
func (m *Manager) systemPromptOverrideConfigSnapshot() *internalconfig.SystemPromptOverrideConfig {
	cfg := m.runtimeConfigSnapshot()
	if cfg == nil {
		return nil
	}
	return &cfg.SystemPromptOverride
}

// applySystemPromptOverrideForAuth applies the configured system prompt override
// for the selected provider when enabled. It is a no-op when the config
// snapshot is absent, the override is disabled, or the request does not match.
func (m *Manager) applySystemPromptOverrideForAuth(provider string, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Request, cliproxyexecutor.Options) {
	if m == nil {
		return req, opts
	}
	cfg := m.systemPromptOverrideConfigSnapshot()
	if cfg == nil || !cfg.Enabled {
		return req, opts
	}
	rules := systemPromptOverrideRulesFromConfig(*cfg)
	// Feature gate: proceed when any actionable knob is set (prompt text, file,
	// system replacements, or tool description replacements).
	hasPrompt := strings.TrimSpace(rules.Prompt) != "" || strings.TrimSpace(rules.PromptFile) != ""
	hasReplacements := len(rules.Replacements) > 0 || len(rules.ToolReplacements) > 0
	if !hasPrompt && !hasReplacements {
		return req, opts
	}
	return applySystemPromptOverride(provider, req, opts, rules)
}
