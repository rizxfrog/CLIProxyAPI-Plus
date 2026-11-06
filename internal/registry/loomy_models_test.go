package registry

import "testing"

// TestGetLoomyModelsLoadsSharedCatalog verifies the loomy section of the shared
// models.json is loaded so a Loomy credential has a fallback catalog before the
// per-account gateway probe publishes the live list.
func TestGetLoomyModelsLoadsSharedCatalog(t *testing.T) {
	models := GetLoomyModels()
	if len(models) == 0 {
		t.Fatal("loomy section of the shared catalog is empty")
	}
	ids := make(map[string]*ModelInfo, len(models))
	for _, model := range models {
		if model == nil || model.ID == "" {
			t.Fatalf("catalog contains an unusable entry: %+v", model)
		}
		if _, dup := ids[model.ID]; dup {
			t.Fatalf("duplicate model id %q", model.ID)
		}
		if model.Type != "loomy" {
			t.Fatalf("model %q type = %q, want loomy", model.ID, model.Type)
		}
		if model.Object == "" || model.OwnedBy == "" {
			t.Fatalf("model %q missing object/owned_by metadata: %+v", model.ID, model)
		}
		ids[model.ID] = model
	}

	spark := ids["spark-x2.6"]
	if spark == nil {
		t.Fatalf("catalog is missing spark-x2.6: %v", ids)
	}
	if spark.Thinking == nil || !spark.Thinking.ZeroAllowed || len(spark.Thinking.Levels) == 0 {
		t.Fatalf("spark-x2.6 must declare a disable-able thinking menu: %+v", spark.Thinking)
	}

	// Multi-modal rows must advertise their input modalities or the request
	// validator would reject image and video attachments.
	if mimo := ids["mimo-v2.5"]; mimo == nil || !containsModality(mimo.SupportedInputModalities, "image") {
		t.Fatalf("mimo-v2.5 must advertise image input: %+v", mimo)
	}
}

// TestLoomyModelsResolveThroughStaticLookup verifies the channel dispatch and the
// generic lookup both reach the shared loomy section.
func TestLoomyModelsResolveThroughStaticLookup(t *testing.T) {
	byChannel := LookupStaticModelInfoByChannel("mimo-v2.5", "loomy")
	if byChannel == nil || byChannel.ID != "mimo-v2.5" {
		t.Fatalf("channel lookup = %+v", byChannel)
	}
	if generic := LookupStaticModelInfo("mimo-v2.5"); generic == nil {
		t.Fatal("generic static lookup does not see the loomy section")
	}
	if len(GetStaticModelDefinitionsByChannel("loomy")) != len(GetLoomyModels()) {
		t.Fatal("channel dispatch does not return the shared loomy section")
	}
}

// TestGetLoomyModelsReturnsClone pins that callers cannot mutate the catalog.
func TestGetLoomyModelsReturnsClone(t *testing.T) {
	first := GetLoomyModels()
	if len(first) == 0 {
		t.Fatal("empty catalog")
	}
	first[0].ID = "mutated"
	if second := GetLoomyModels(); second[0].ID == "mutated" {
		t.Fatal("the catalog was mutated through a returned slice")
	}
}

func containsModality(modalities []string, want string) bool {
	for _, modality := range modalities {
		if modality == want {
			return true
		}
	}
	return false
}
