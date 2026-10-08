package registry

import "testing"

// TestGetFloatboatModels pins the FloatBoat managed-account catalogue. These
// entries are the fallback used when the live gateway probe cannot run, so they
// carry the models that were verified callable on the base plan. The catalogue
// is superseded at runtime by the credential-scoped probe, which re-reads the
// gateway's own /api/pricing.
func TestGetFloatboatModels(t *testing.T) {
	models := GetFloatboatModels()
	if len(models) == 0 {
		t.Fatal("GetFloatboatModels() returned no models")
	}
	ids := map[string]*ModelInfo{}
	for _, m := range models {
		if m == nil {
			t.Fatal("nil model in floatboat catalogue")
		}
		ids[m.ID] = m
	}
	for _, want := range []string{"deepseek-flash", "gemini-3.8-flash", "glm-5.2", "kimi-k3"} {
		if _, ok := ids[want]; !ok {
			t.Fatalf("floatboat catalogue missing %q: %+v", want, ids)
		}
	}
	// Models that the gateway does not publish must not come back as fallbacks.
	for _, gone := range []string{"claude-sonnet-4-6", "claude-opus-4-6"} {
		if _, ok := ids[gone]; ok {
			t.Fatalf("floatboat catalogue must not contain unpublished model %q", gone)
		}
	}
	flash := ids["deepseek-flash"]
	if flash.ContextLength != 128000 {
		t.Fatalf("deepseek-flash context = %d, want 128000", flash.ContextLength)
	}
}

func TestGetStaticModelDefinitionsByFloatboatChannel(t *testing.T) {
	models := GetStaticModelDefinitionsByChannel("floatboat")
	if len(models) == 0 {
		t.Fatal(`GetStaticModelDefinitionsByChannel("floatboat") returned no models`)
	}
	if len(GetStaticModelDefinitionsByChannel("FloatBoat")) == 0 {
		t.Fatal("floatboat channel lookup must be case-insensitive")
	}
}

func TestLookupFloatboatModelInfo(t *testing.T) {
	info := LookupModelInfo("deepseek-flash", "floatboat")
	if info == nil {
		t.Fatal(`LookupModelInfo(deepseek-flash, floatboat) = nil`)
	}
	if info.ContextLength != 128000 || info.MaxCompletionTokens != 65536 {
		t.Fatalf("info = %+v", info)
	}
}
