package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
)

func TestProviderModelsNormalization(t *testing.T) {
	entries, err := NormalizeProviderModels(map[string]ProviderModels{" TRAE ": {Disabled: []string{" old ", "old"}, Custom: []ProviderModel{{ID: " new ", DisplayName: " New "}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries["trae"].Disabled) != 1 || entries["trae"].Custom[0].ID != "new" {
		t.Fatalf("unexpected normalization: %#v", entries)
	}
	for _, entry := range []ProviderModels{
		{Disabled: []string{"*"}}, {Custom: []ProviderModel{{ID: ""}}}, {Custom: []ProviderModel{{ID: "a"}, {ID: " a "}}},
		{Custom: []ProviderModel{{ID: "a", Thinking: &registry.ThinkingSupport{Min: 10, Max: 1}}}},
	} {
		if _, err := NormalizeProviderModels(map[string]ProviderModels{"trae": entry}); err == nil {
			t.Fatalf("accepted invalid overlay: %#v", entry)
		}
	}
	if _, err := NormalizeProviderModels(map[string]ProviderModels{"TRAE": {}, "trae": {}}); err == nil {
		t.Fatal("accepted duplicate normalized provider")
	}
}

func TestProviderModelsMergeAndFilter(t *testing.T) {
	base := []*registry.ModelInfo{{ID: "alias", MetadataModelID: "old", DisplayName: "Old", ContextLength: 100, Thinking: &registry.ThinkingSupport{Levels: []string{"low"}}}}
	overlay := ProviderModels{Disabled: []string{"old"}, Custom: []ProviderModel{{ID: "old", DisplayName: "Updated"}, {ID: "new"}}}
	merged := MergeProviderModels(base, "trae", overlay, true)
	if len(merged) != 2 || merged[0].ID != "alias" || merged[0].DisplayName != "Updated" || merged[0].ContextLength != 100 || merged[0].Thinking == nil {
		t.Fatalf("bad merge: %#v", merged)
	}
	if base[0].DisplayName != "Old" {
		t.Fatal("mutated catalog")
	}
	filtered := FilterProviderModels(merged, overlay)
	if len(filtered) != 1 || filtered[0].ID != "new" {
		t.Fatalf("alias bypassed disable: %#v", filtered)
	}
	if got := MergeProviderModels(base, "trae", overlay, false); len(got) != 1 {
		t.Fatal("expanded explicit whitelist")
	}
	if len(FilterProviderModels(merged, ProviderModels{})) != 2 {
		t.Fatal("leaked disable to another provider")
	}
}

func TestProviderModelsPersistenceResetAndParse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("# keep me\nport: 8317\nprovider-models:\n  trae:\n    disabled: [old]\n    custom:\n      - id: new\n        display-name: New\n        context-length: 100\n  kimi:\n    disabled: [other]\n    custom: []\n")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseConfigBytes(data)
	if err != nil || len(parsed.ProviderModels) != 2 {
		t.Fatalf("parse failed: %v", err)
	}
	clone := cfg.CloneForRuntime()
	entry := clone.ProviderModels["trae"]
	entry.Custom[0].DisplayName = "Changed"
	if cfg.ProviderModels["trae"].Custom[0].DisplayName != "New" {
		t.Fatal("clone shares overlay")
	}
	cfg.ProviderModels["trae"] = ProviderModels{Disabled: []string{}, Custom: []ProviderModel{{ID: "new"}}}
	delete(cfg.ProviderModels, "kimi")
	if err := SaveConfigPreserveComments(path, cfg); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ProviderModels) != 1 || cfg.ProviderModels["trae"].Custom[0].ContextLength != nil || cfg.ProviderModels["trae"].Custom[0].DisplayName != "" {
		t.Fatalf("stale deleted fields: %#v", cfg.ProviderModels)
	}
	cfg.ProviderModels = nil
	if err := SaveConfigPreserveComments(path, cfg); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig(path)
	if err != nil || len(cfg.ProviderModels) != 0 {
		t.Fatalf("reset did not persist: %v", err)
	}
	if _, err := ParseConfigBytes([]byte("provider-models:\n  trae:\n    custom: [{id: a}, {id: a}]\n")); err == nil {
		t.Fatal("parser accepted duplicate models")
	}
}
