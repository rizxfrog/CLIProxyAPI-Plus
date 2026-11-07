package registry

import "testing"

// TestGetDumateModelsLoadsSharedCatalog verifies the dumate section of the
// shared models.json is loaded. The Dumate client's local backend exposes no
// model-list endpoint, so this catalog is the only source of routable IDs.
func TestGetDumateModelsLoadsSharedCatalog(t *testing.T) {
	models := GetDumateModels()
	if len(models) == 0 {
		t.Fatal("dumate section of the shared catalog is empty")
	}
	ids := make(map[string]*ModelInfo, len(models))
	for _, model := range models {
		if model == nil || model.ID == "" {
			t.Fatalf("catalog contains an unusable entry: %+v", model)
		}
		if _, dup := ids[model.ID]; dup {
			t.Fatalf("duplicate model id %q", model.ID)
		}
		if model.Type != "dumate" {
			t.Fatalf("model %q type = %q, want dumate", model.ID, model.Type)
		}
		if model.Object == "" || model.OwnedBy == "" {
			t.Fatalf("model %q missing object/owned_by metadata: %+v", model.ID, model)
		}
		// The backend accepts reasoning_effort on every verified model, so the
		// catalog must advertise a disable-able thinking menu.
		if model.Thinking == nil || !model.Thinking.ZeroAllowed || len(model.Thinking.Levels) == 0 {
			t.Fatalf("model %q must declare a disable-able thinking menu: %+v", model.ID, model.Thinking)
		}
		ids[model.ID] = model
	}

	// IDs confirmed against a running client; each resolves to a vendor model
	// reported in the X-Dumate-Gw-Realmodel response header.
	for _, want := range []string{"glm-5", "deepseek-v4-pro", "deepseek-v4-flash-0731", "deepseek-v41-flash"} {
		if ids[want] == nil {
			t.Fatalf("catalog is missing %s: %v", want, ids)
		}
	}
}

// TestDumateModelsResolveThroughStaticLookup verifies the channel dispatch and
// the generic lookup both reach the shared dumate section.
func TestDumateModelsResolveThroughStaticLookup(t *testing.T) {
	byChannel := LookupStaticModelInfoByChannel("glm-5", "dumate")
	if byChannel == nil || byChannel.ID != "glm-5" {
		t.Fatalf("channel lookup = %+v", byChannel)
	}
	if generic := LookupStaticModelInfo("glm-5"); generic == nil {
		t.Fatal("generic static lookup does not see the dumate section")
	}
	if len(GetStaticModelDefinitionsByChannel("dumate")) != len(GetDumateModels()) {
		t.Fatal("channel dispatch does not return the shared dumate section")
	}
}

// TestGetDumateModelsReturnsClone pins that callers cannot mutate the catalog.
func TestGetDumateModelsReturnsClone(t *testing.T) {
	first := GetDumateModels()
	if len(first) == 0 {
		t.Fatal("empty catalog")
	}
	first[0].ID = "mutated"
	if second := GetDumateModels(); second[0].ID == "mutated" {
		t.Fatal("the catalog was mutated through a returned slice")
	}
}
