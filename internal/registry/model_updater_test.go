package registry

import (
	"reflect"
	"testing"
)

func TestDefaultGeneralCatalogSource(t *testing.T) {
	t.Run("no env override keeps official remote sources", func(t *testing.T) {
		t.Setenv("MODELS_FILE", "")
		t.Setenv("MODELS_URL", "")

		if got := defaultGeneralCatalogSource(); got != "" {
			t.Fatalf("defaultGeneralCatalogSource() = %q, want empty so the official URLs are used", got)
		}
	})

	t.Run("custom URL", func(t *testing.T) {
		t.Setenv("MODELS_FILE", "")
		t.Setenv("MODELS_URL", " https://example.com/models.json ")

		if got := defaultGeneralCatalogSource(); got != "https://example.com/models.json" {
			t.Fatalf("defaultGeneralCatalogSource() = %q, want %q", got, "https://example.com/models.json")
		}
	})

	t.Run("file takes precedence", func(t *testing.T) {
		t.Setenv("MODELS_FILE", " /data/models.json ")
		t.Setenv("MODELS_URL", "https://example.com/models.json")

		if got := defaultGeneralCatalogSource(); got != "/data/models.json" {
			t.Fatalf("defaultGeneralCatalogSource() = %q, want %q", got, "/data/models.json")
		}
	})
}

func TestEffectiveCatalogSourcesRespectsEnvDefault(t *testing.T) {
	t.Setenv("MODELS_FILE", "/data/models.json")
	t.Setenv("MODELS_URL", "https://example.com/models.json")

	got := effectiveCatalogSources(CatalogSources{}, false, false)
	if got.Catalog != "/data/models.json" {
		t.Fatalf("env default catalog = %q, want %q", got.Catalog, "/data/models.json")
	}

	// An explicit models.catalog value overrides the environment default.
	explicit := effectiveCatalogSources(CatalogSources{Catalog: "https://example.com/explicit.json"}, false, false)
	if explicit.Catalog != "https://example.com/explicit.json" {
		t.Fatalf("explicit catalog = %q, want the configured source", explicit.Catalog)
	}

	// --local-model selects the embedded catalog over the environment default.
	local := effectiveCatalogSources(CatalogSources{}, true, false)
	if local.Catalog != embeddedCatalogSource {
		t.Fatalf("local catalog = %q, want %q", local.Catalog, embeddedCatalogSource)
	}

	// Home mode disables the general catalog regardless of the environment.
	home := effectiveCatalogSources(CatalogSources{}, false, true)
	if home.Catalog != "disabled" {
		t.Fatalf("home catalog = %q, want disabled", home.Catalog)
	}
}

func TestDetectChangedProvidersIncludesLocalProviders(t *testing.T) {
	oldData := &staticModelsJSON{
		CodeBuddyCN: []*ModelInfo{{ID: "codebuddy-old"}},
		DeepSeekWeb: []*ModelInfo{{ID: "deepseek-old"}},
	}
	newData := &staticModelsJSON{
		CodeBuddyCN: []*ModelInfo{{ID: "codebuddy-new"}},
		DeepSeekWeb: []*ModelInfo{{ID: "deepseek-new"}},
	}

	got := detectChangedProviders(oldData, newData)
	want := []string{"codebuddy-cn", "deepseek-web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detectChangedProviders() = %v, want %v", got, want)
	}
}

func TestDetectChangedProviders_CodexConfigurationUpdate(t *testing.T) {
	oldData := &staticModelsJSON{
		CodexFree: []*ModelInfo{{ID: "gpt-6-luna"}},
	}
	newData := &staticModelsJSON{
		CodexFree: []*ModelInfo{{ID: "gpt-6-luna", SupportConfigurationUpdate: true}},
	}

	changed := detectChangedProviders(oldData, newData)
	if len(changed) != 1 || changed[0] != "codex" {
		t.Fatalf("configuration_update-only change: got providers %v, want [codex]", changed)
	}
}

func TestDetectChangedProviders_KimiAliases(t *testing.T) {
	oldData := &staticModelsJSON{
		Kimi: []*ModelInfo{{ID: "kimi-k2"}},
	}
	newData := &staticModelsJSON{
		Kimi: []*ModelInfo{{ID: "kimi-k2"}, {ID: "kimi-k3"}},
	}

	changed := detectChangedProviders(oldData, newData)
	expected := map[string]bool{
		"kimi":     false,
		"kimi-ai":  false,
		"kimi.ai":  false,
		"kimi.com": false,
	}

	for _, p := range changed {
		if _, ok := expected[p]; ok {
			expected[p] = true
		}
	}

	for p, found := range expected {
		if !found {
			t.Errorf("expected changed provider %q to be reported, got %v", p, changed)
		}
	}
}

// TestDetectChangedProvidersIncludesMergedForkSections pins that the sections
// migrated from dedicated embedded files are tracked by change detection so a
// remote update to them re-registers the affected credentials.
func TestDetectChangedProvidersIncludesMergedForkSections(t *testing.T) {
	oldData := &staticModelsJSON{
		Xiaohuanxiong: []*ModelInfo{{ID: "glm-5-3"}},
		CodeArts:      []*ModelInfo{{ID: "GLM-5.1"}},
		Loomy:         []*ModelInfo{{ID: "spark-x2.6"}},
	}
	newData := &staticModelsJSON{
		Xiaohuanxiong: []*ModelInfo{{ID: "glm-5-3"}, {ID: "kimi-k3"}},
		CodeArts:      []*ModelInfo{{ID: "GLM-5.2"}},
		Loomy:         []*ModelInfo{{ID: "spark-x2.6", DisplayName: "Spark X2.6"}},
	}

	got := detectChangedProviders(oldData, newData)
	want := map[string]bool{"xiaohuanxiong": false, "codearts": false, "loomy": false}
	for _, provider := range got {
		if _, ok := want[provider]; ok {
			want[provider] = true
		}
	}
	for provider, seen := range want {
		if !seen {
			t.Errorf("detectChangedProviders() = %v, missing changed provider %q", got, provider)
		}
	}
}
