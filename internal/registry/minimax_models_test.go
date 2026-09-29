package registry

import "testing"

func TestGetMinimaxModels(t *testing.T) {
	models := GetMinimaxModels()
	if len(models) != 3 {
		t.Fatalf("GetMinimaxModels() returned %d models, want 3", len(models))
	}
	ids := map[string]*ModelInfo{}
	for _, m := range models {
		ids[m.ID] = m
	}
	m3, ok := ids["MiniMax-M3"]
	if !ok {
		t.Fatalf("MiniMax-M3 missing: %+v", models)
	}
	if m3.ContextLength != 512000 {
		t.Fatalf("MiniMax-M3 context = %d", m3.ContextLength)
	}
	if _, ok := ids["MiniMax-M2.7"]; !ok {
		t.Fatal("MiniMax-M2.7 missing")
	}
	if _, ok := ids["MiniMax-M2.7-highspeed"]; !ok {
		t.Fatal("MiniMax-M2.7-highspeed missing")
	}
}

func TestGetStaticModelDefinitionsByMinimaxChannel(t *testing.T) {
	for _, channel := range []string{"minimax", "minimax-cn"} {
		models := GetStaticModelDefinitionsByChannel(channel)
		if len(models) != 3 {
			t.Fatalf("channel %q returned %d models, want 3", channel, len(models))
		}
	}
}

func TestLookupMinimaxModelInfo(t *testing.T) {
	for _, provider := range []string{"minimax", "minimax-cn"} {
		info := LookupModelInfo("MiniMax-M3", provider)
		if info == nil {
			t.Fatalf("LookupModelInfo(MiniMax-M3, %s) = nil", provider)
		}
		if info.ContextLength != 512000 || info.MaxCompletionTokens != 128000 {
			t.Fatalf("%s info = %+v", provider, info)
		}
		if info.Thinking == nil {
			t.Fatalf("%s thinking metadata missing", provider)
		}
	}
}
