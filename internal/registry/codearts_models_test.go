package registry

import (
	"testing"
)

// codeArtsGatewayMaxTokens is the output ceiling the CodeArts gateway accepts.
// It is deliberately 65536 rather than the per-model values the upstream
// catalog advertises (131072-393216): the gateway validates max_tokens
// server-side and rejects anything above 65536 with an InferHub
// RANGE_VALIDATOR 400.
const codeArtsGatewayMaxTokens = 65536

// TestGetCodeArtsModels verifies the shared catalog's codearts section.
func TestGetCodeArtsModels(t *testing.T) {
	models := GetCodeArtsModels()
	if len(models) == 0 {
		t.Fatal("codearts section of the shared catalog is empty")
	}

	byID := make(map[string]*ModelInfo, len(models))
	for _, model := range models {
		byID[model.ID] = model
		if model.Type != "codearts" {
			t.Errorf("model %s type = %q, want codearts", model.ID, model.Type)
		}
		if model.OwnedBy == "" {
			t.Errorf("model %s has no owned_by", model.ID)
		}
		if model.MaxCompletionTokens <= 0 {
			t.Errorf("model %s has no max_completion_tokens", model.ID)
		}
	}

	glm52, ok := byID["GLM-5.2"]
	if !ok {
		t.Fatalf("GLM-5.2 missing from the catalog: %v", keys(byID))
	}
	if glm52.ContextLength != 202752 {
		t.Errorf("GLM-5.2 context_length = %d, want 202752", glm52.ContextLength)
	}
	if glm52.Thinking == nil || len(glm52.Thinking.Levels) == 0 {
		t.Error("GLM-5.2 should advertise thinking levels")
	}

	// The gateway rejects max_tokens above 65536, so every catalog entry must
	// stay within the server-enforced ceiling regardless of the larger values
	// the upstream catalog advertises.
	for _, model := range models {
		if model.MaxCompletionTokens > codeArtsGatewayMaxTokens {
			t.Errorf("model %s declares max_completion_tokens %d, above the gateway ceiling %d",
				model.ID, model.MaxCompletionTokens, codeArtsGatewayMaxTokens)
		}
	}

	// Models captured live from the upstream agent-center/gateway-config lists.
	for _, want := range []string{
		"GLM-5.2",
		"glm-5.2-sft-harmony",
		"openpangu-2.0-pro",
		"openpangu-2.0-flash",
		"deepseek-v4-flash-0731",
		"deepseek-v4-pro-0813",
		"glm-5.3-flash",
	} {
		if _, ok := byID[want]; !ok {
			t.Errorf("%s missing from the catalog: %v", want, keys(byID))
		}
	}

	// The wire id and the display name differ for these models; the display name
	// is what the desktop client shows, the id is what the gateway routes.
	display := map[string]string{
		"glm-5.2-sft-harmony": "GLM-5.2-ArkTS-SPARK",
		"openpangu-2.0-pro":   "OpenPangu-2.0-Pro",
	}
	for id, want := range display {
		model, ok := byID[id]
		if !ok {
			t.Errorf("%s missing", id)
			continue
		}
		if model.DisplayName != want {
			t.Errorf("%s display_name = %q, want %q", id, model.DisplayName, want)
		}
	}

	// Qwen3-VL-235B is the multimodal entry; image input must be advertised or
	// the aggregator would reject image attachments.
	if qwen, ok := byID["Qwen3-VL-235B"]; !ok {
		t.Errorf("Qwen3-VL-235B missing from the catalog: %v", keys(byID))
	} else {
		found := false
		for _, modality := range qwen.SupportedInputModalities {
			if modality == "image" {
				found = true
			}
		}
		if !found {
			t.Errorf("Qwen3-VL-235B should accept image input: %v", qwen.SupportedInputModalities)
		}
	}
}

// TestGetCodeArtsModelsReturnsClone pins that callers cannot mutate the catalog.
func TestGetCodeArtsModelsReturnsClone(t *testing.T) {
	first := GetCodeArtsModels()
	if len(first) == 0 {
		t.Fatal("empty catalog")
	}
	original := first[0].ID
	first[0].ID = "mutated"
	second := GetCodeArtsModels()
	if second[0].ID == "mutated" {
		t.Fatal("the catalog was mutated through a returned slice")
	}
	if second[0].ID != original {
		t.Fatalf("clone lost the original id: %q", second[0].ID)
	}
}

// TestGetStaticModelDefinitionsByChannelCodeArts pins the channel routing used by
// the management model-definitions API.
func TestGetStaticModelDefinitionsByChannelCodeArts(t *testing.T) {
	models := GetStaticModelDefinitionsByChannel("codearts")
	if len(models) == 0 {
		t.Fatal("codearts channel returned no models")
	}
	for _, model := range models {
		if model.ID == "" {
			t.Error("a model has an empty id")
		}
	}
	// Case-insensitive lookup must work too.
	if len(GetStaticModelDefinitionsByChannel("CodeArts")) == 0 {
		t.Fatal("channel lookup should be case-insensitive")
	}
}

// TestLookupStaticModelInfoFindsCodeArts pins the shared fallback lookup.
func TestLookupStaticModelInfoFindsCodeArts(t *testing.T) {
	if got := LookupStaticModelInfo("GLM-5.2"); got == nil || got.ID != "GLM-5.2" {
		t.Fatalf("LookupStaticModelInfo(GLM-5.2) = %+v", got)
	}
}

func keys(m map[string]*ModelInfo) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}
