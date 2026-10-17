package registry

import (
	"testing"
)

// TestGetXiaohuanxiongModelsLoadsEmbeddedCatalog verifies the dedicated embedded
// file is parsed. It is separate from models.json so a remote catalog refresh
// cannot erase the provider's models.
func TestGetXiaohuanxiongModelsLoadsEmbeddedCatalog(t *testing.T) {
	models := GetXiaohuanxiongModels()
	if len(models) == 0 {
		t.Fatal("embedded xiaohuanxiong catalog is empty")
	}
	ids := make(map[string]*ModelInfo, len(models))
	for _, model := range models {
		if model == nil || model.ID == "" {
			t.Fatalf("catalog contains an unusable entry: %+v", model)
		}
		if _, dup := ids[model.ID]; dup {
			t.Fatalf("duplicate model id %q", model.ID)
		}
		ids[model.ID] = model
	}

	// The default desktop model and the documented GLM id must be present.
	for _, want := range []string{"raccoon-chat-ml-5-5", "glm-5-3"} {
		if _, ok := ids[want]; !ok {
			t.Fatalf("catalog is missing %q", want)
		}
	}
	current := ids["sn-deepseek-v4-1-flash"]
	if current == nil || current.DisplayName != "DeepSeek V4.1 Flash" {
		t.Fatalf("current DeepSeek model missing or mislabeled: %+v", current)
	}
	if current.ContextLength != 1000000 || current.MaxCompletionTokens != 100000 {
		t.Fatalf("current DeepSeek model limits do not match upstream: %+v", current)
	}
	if ids["deepseek-v4-pro-0813"] != nil {
		t.Fatal("retired DeepSeek V4 Pro remains in the Xiaohuanxiong catalog")
	}
}

func TestCodeBuddyCNIncludesStep5Preview(t *testing.T) {
	for _, model := range GetCodeBuddyCNModels() {
		if model != nil && model.ID == "step-5-preview" {
			if model.DisplayName != "Step-5-Preview" {
				t.Fatalf("Step-5 display name = %q", model.DisplayName)
			}
			return
		}
	}
	t.Fatal("WorkBuddy model step-5-preview is absent")
}

// TestXiaohuanxiongCatalogThinkingMetadata guards the capability data that drives
// whether thinking is offered to clients.
func TestXiaohuanxiongCatalogThinkingMetadata(t *testing.T) {
	if model := LookupXiaohuanxiongModel("glm-5-3"); model == nil {
		t.Fatal("glm-5-3 not found")
	} else if model.Thinking == nil {
		t.Fatal("glm-5-3 must declare thinking support")
	} else {
		if !model.Thinking.ZeroAllowed {
			t.Error("glm-5-3 thinking should allow zero (disabled)")
		}
		if len(model.Thinking.Levels) == 0 {
			t.Error("glm-5-3 thinking should declare discrete levels")
		}
	}

	// SenseNova does not expose a thinking channel; it must not advertise one.
	if model := LookupXiaohuanxiongModel("sn-sensenova-6-8-flash-lite"); model == nil {
		t.Fatal("sn-sensenova-6-8-flash-lite not found")
	} else if model.Thinking != nil {
		t.Fatalf("sensenova must not advertise thinking support: %+v", model.Thinking)
	}
}

// TestXiaohuanxiongContextWindows pins the documented Claude Code ceilings.
func TestXiaohuanxiongContextWindows(t *testing.T) {
	defaultModel := LookupXiaohuanxiongModel("raccoon-chat-ml-5-5")
	if defaultModel == nil {
		t.Fatal("default model not found")
	}
	if defaultModel.ContextLength != 180000 {
		t.Fatalf("default context = %d, want 180000", defaultModel.ContextLength)
	}
	if defaultModel.MaxCompletionTokens != 80000 {
		t.Fatalf("default max tokens = %d, want 80000 (hosted default)", defaultModel.MaxCompletionTokens)
	}
}

// TestLookupXiaohuanxiongModelMissesCleanly avoids accidental cross-provider hits.
func TestLookupXiaohuanxiongModelMissesCleanly(t *testing.T) {
	for _, miss := range []string{"", "   ", "does-not-exist", "claude-sonnet-4"} {
		if got := LookupXiaohuanxiongModel(miss); got != nil {
			t.Fatalf("LookupXiaohuanxiongModel(%q) = %+v, want nil", miss, got)
		}
	}
}

// TestXiaohuanxiongChannelDispatch ensures the model-list endpoint can resolve
// this provider by channel name.
func TestXiaohuanxiongChannelDispatch(t *testing.T) {
	models := GetStaticModelDefinitionsByChannel("xiaohuanxiong")
	if len(models) == 0 {
		t.Fatal("channel dispatch returned no models for xiaohuanxiong")
	}
	if len(models) != len(GetXiaohuanxiongModels()) {
		t.Fatalf("channel dispatch count %d != catalog count %d", len(models), len(GetXiaohuanxiongModels()))
	}
}

// TestXiaohuanxiongModelsAreNotInSharedCatalog documents the durability contract:
// models.json must not carry the section, because the remote updater replaces
// that file's contents wholesale and would drop the provider.
func TestXiaohuanxiongModelsAreNotInSharedCatalog(t *testing.T) {
	cached := getModels()
	if cached == nil {
		t.Skip("shared catalog not loaded in this test context")
	}
	// The type intentionally has no Xiaohuanxiong field. Assert indirectly:
	// channel dispatch must still resolve models from the embedded store.
	models := GetStaticModelDefinitionsByChannel("xiaohuanxiong")
	if len(models) == 0 {
		t.Fatal("xiaohuanxiong models must resolve from the embedded store, not models.json")
	}
}
