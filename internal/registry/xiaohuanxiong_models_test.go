package registry

import (
	"encoding/json"
	"testing"
)

// TestGetXiaohuanxiongModelsLoadsSharedCatalog verifies the xiaohuanxiong
// section of the shared models.json is loaded.
func TestGetXiaohuanxiongModelsLoadsSharedCatalog(t *testing.T) {
	models := GetXiaohuanxiongModels()
	if len(models) == 0 {
		t.Fatal("xiaohuanxiong section of the shared catalog is empty")
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
	if model := LookupStaticModelInfoByChannel("glm-5-3", "xiaohuanxiong"); model == nil {
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
	if model := LookupStaticModelInfoByChannel("sn-sensenova-6-8-flash-lite", "xiaohuanxiong"); model == nil {
		t.Fatal("sn-sensenova-6-8-flash-lite not found")
	} else if model.Thinking != nil {
		t.Fatalf("sensenova must not advertise thinking support: %+v", model.Thinking)
	}
}

// TestXiaohuanxiongContextWindows pins the documented Claude Code ceilings.
func TestXiaohuanxiongContextWindows(t *testing.T) {
	defaultModel := LookupStaticModelInfoByChannel("raccoon-chat-ml-5-5", "xiaohuanxiong")
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

// TestXiaohuanxiongChannelLookupMissesCleanly avoids accidental cross-provider hits.
func TestXiaohuanxiongChannelLookupMissesCleanly(t *testing.T) {
	for _, miss := range []string{"", "   ", "does-not-exist", "claude-sonnet-4"} {
		if got := LookupStaticModelInfoByChannel(miss, "xiaohuanxiong"); got != nil {
			t.Fatalf("LookupStaticModelInfoByChannel(%q, xiaohuanxiong) = %+v, want nil", miss, got)
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

// TestXiaohuanxiongRefreshKeepsSection documents the durability contract: the
// section ships in models.json, and a remote catalog that omits it must not
// drop the provider (preserveLocalCatalogSections).
func TestXiaohuanxiongRefreshKeepsSection(t *testing.T) {
	previous := getModels()
	if previous == nil || len(previous.Xiaohuanxiong) == 0 {
		t.Fatal("test setup: shared catalog has no xiaohuanxiong section")
	}
	t.Cleanup(func() {
		modelsCatalogStore.mu.Lock()
		modelsCatalogStore.data = previous
		modelsCatalogStore.mu.Unlock()
	})

	// An upstream catalog that carries official providers only.
	upstream, errMarshal := json.Marshal(&staticModelsJSON{Claude: []*ModelInfo{{ID: "claude-opus-5-5"}}})
	if errMarshal != nil {
		t.Fatal(errMarshal)
	}
	if _, errPublish := publishCatalogBytes(upstream); errPublish != nil {
		t.Fatalf("publishCatalogBytes() error = %v", errPublish)
	}
	if got := len(getModels().Xiaohuanxiong); got != len(previous.Xiaohuanxiong) {
		t.Fatalf("refresh dropped the xiaohuanxiong section: got %d models, want %d", got, len(previous.Xiaohuanxiong))
	}
}
