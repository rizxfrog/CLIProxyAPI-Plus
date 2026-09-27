package registry

import (
	"testing"
)

// TestClineModelsSmoke verifies the Cline builtin catalog is seeded from the
// live GET /api/v1/models response (458 models, 2026-09-26) rather than a
// hand-curated subset. ClinePass/Free/Cloud tier ids (e.g. cline-pass/*) are not
// returned by /api/v1/models and are therefore intentionally absent here.
func TestClineModelsSmoke(t *testing.T) {
	models := GetClineModels()
	if len(models) < 450 {
		t.Fatalf("expected ~458 cline models, got %d", len(models))
	}
	present := map[string]bool{}
	for _, m := range models {
		present[m.ID] = true
	}
	// A representative sample of real Cline /api/v1/models ids. These were
	// previously assumed (incorrectly) to be OmniRoute fabrications; the live
	// endpoint confirms they are genuine Cline model ids.
	for _, id := range []string{
		"typesafe/jev-router",
		"openai/gpt-6-luna",
		"z-ai/glm-5.2",
		"x-ai/grok-4.5",
		"openai/gpt-5.6-sol",
		"anthropic/claude-opus-4.8",
		"openrouter/free",
		"minimax/minimax-m3",
		"moonshotai/kimi-k3",
	} {
		if !present[id] {
			t.Errorf("missing expected real cline model id %q", id)
		}
	}
	// ClinePass-tier ids are NOT in /api/v1/models (separate subscription tier).
	if present["cline-pass/glm-5.3"] {
		t.Errorf("cline-pass/glm-5.3 should not be in /api/v1/models catalog")
	}
}
