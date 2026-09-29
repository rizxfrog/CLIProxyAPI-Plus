package xiaohuanxiong

import (
	"encoding/json"
	"testing"

	xhxthinking "github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
	"github.com/tidwall/gjson"
)

// TestTranslateRequestThinking verifies that the generic reasoning_effort the
// shared thinking pipeline emits is re-expressed in each model family's dialect.
func TestTranslateRequestThinking(t *testing.T) {
	tests := []struct {
		name  string
		model string
		body  string
		// want is the full expected body when wantBody is set, otherwise the
		// assertions below are applied to the result.
		wantEffort    *string
		wantExtraBody map[string]any
	}{
		{
			name:       "glm-5-3 enabled maps to high",
			model:      "glm-5-3",
			body:       `{"model":"glm-5-3","reasoning_effort":"high"}`,
			wantEffort: strPtr("high"),
		},
		{
			name:       "glm-5-3 disabled maps to low",
			model:      "glm-5-3",
			body:       `{"model":"glm-5-3","reasoning_effort":"none"}`,
			wantEffort: strPtr("low"),
		},
		{
			name:       "glm-5.3 dotted variant also maps to low",
			model:      "glm-5.3",
			body:       `{"reasoning_effort":"none"}`,
			wantEffort: strPtr("low"),
		},
		{
			name:          "deepseek enabled moves into extra_body.thinking",
			model:         "deepseek-v4-pro-0813",
			body:          `{"reasoning_effort":"high"}`,
			wantExtraBody: map[string]any{"thinking": map[string]any{"type": "enabled"}},
		},
		{
			name:          "deepseek disabled moves into extra_body.thinking",
			model:         "deepseek-v4-pro-0813",
			body:          `{"reasoning_effort":"none"}`,
			wantExtraBody: map[string]any{"thinking": map[string]any{"type": "disabled"}},
		},
		{
			name:          "doubao shares the deepseek dialect",
			model:         "doubao-seed-1-6",
			body:          `{"reasoning_effort":"high"}`,
			wantExtraBody: map[string]any{"thinking": map[string]any{"type": "enabled"}},
		},
		{
			name:          "qwen enabled maps to extra_body.enable_thinking true",
			model:         "qwen3-max",
			body:          `{"reasoning_effort":"high"}`,
			wantExtraBody: map[string]any{"enable_thinking": true},
		},
		{
			name:          "qwen disabled maps to extra_body.enable_thinking false",
			model:         "qwen3-max",
			body:          `{"reasoning_effort":"none"}`,
			wantExtraBody: map[string]any{"enable_thinking": false},
		},
		{
			name:       "gemini enabled maps to high",
			model:      "gemini-3-pro",
			body:       `{"reasoning_effort":"high"}`,
			wantEffort: strPtr("high"),
		},
		{
			name:       "gemini disabled maps to none",
			model:      "gemini-3-pro",
			body:       `{"reasoning_effort":"none"}`,
			wantEffort: strPtr("none"),
		},
		{
			name:       "sensenova enabled maps to high",
			model:      "sensenova-6-8-flash",
			body:       `{"reasoning_effort":"low"}`,
			wantEffort: strPtr("high"),
		},
		{
			name:       "sn-sensenova disabled keeps explicit none",
			model:      "sn-sensenova-6-8-flash-lite",
			body:       `{"reasoning_effort":"none"}`,
			wantEffort: strPtr("none"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := TranslateRequestThinking([]byte(tc.body), tc.model)
			if err != nil {
				t.Fatalf("TranslateRequestThinking returned error: %v", err)
			}
			if !gjson.ValidBytes(out) {
				t.Fatalf("result is not valid JSON: %s", out)
			}
			if tc.wantEffort != nil {
				got := gjson.GetBytes(out, "reasoning_effort").String()
				if got != *tc.wantEffort {
					t.Fatalf("reasoning_effort = %q, want %q (body=%s)", got, *tc.wantEffort, out)
				}
			}
			if tc.wantExtraBody != nil {
				got := gjson.GetBytes(out, "extra_body").Raw
				var decoded map[string]any
				if errUnmarshal := json.Unmarshal([]byte(got), &decoded); errUnmarshal != nil {
					t.Fatalf("extra_body is not an object: %s", got)
				}
				for key, want := range tc.wantExtraBody {
					assertDeepEqual(t, key, decoded[key], want)
				}
			}
		})
	}
}

// TestTranslateRequestThinkingStripsGenericEffortForExtraBodyFamilies asserts the
// generic top-level field never leaks to families that use extra_body.
func TestTranslateRequestThinkingStripsGenericEffortForExtraBodyFamilies(t *testing.T) {
	for _, model := range []string{"deepseek-v4-pro-0813", "qwen3-max"} {
		t.Run(model, func(t *testing.T) {
			out, err := TranslateRequestThinking([]byte(`{"reasoning_effort":"high"}`), model)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gjson.GetBytes(out, "reasoning_effort").Exists() {
				t.Fatalf("reasoning_effort leaked upstream for %s: %s", model, out)
			}
		})
	}
}

// TestTranslateRequestThinkingLeavesUnrequestedBodiesUntouched guarantees the
// gateway only rewrites a body that carried thinking intent.
func TestTranslateRequestThinkingLeavesUnrequestedBodiesUntouched(t *testing.T) {
	bodies := []string{
		`{"model":"glm-5-3","messages":[]}`,
		`{"model":"deepseek-v4-pro-0813","messages":[{"role":"user","content":"hi"}]}`,
		`{}`,
	}
	for _, body := range bodies {
		out, err := TranslateRequestThinking([]byte(body), "glm-5-3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(out) != body {
			t.Fatalf("body changed without thinking intent:\n got: %s\nwant: %s", out, body)
		}
	}
}

// TestTranslateRequestThinkingHandlesInvalidJSON ensures a malformed body is
// returned rather than panicking or emitting invalid JSON.
func TestTranslateRequestThinkingHandlesInvalidJSON(t *testing.T) {
	out, err := TranslateRequestThinking([]byte(``), "glm-5-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty body passthrough, got %q", out)
	}
}

// TestDetectFamily covers the documented model-family markers.
func TestDetectFamily(t *testing.T) {
	cases := map[string]ModelFamily{
		"glm-5-3":                     FamilyGLM53,
		"GLM-5.3":                     FamilyGLM53,
		"deepseek-v4-pro-0813":        FamilyDeepSeek,
		"doubao-seed-1-6":             FamilyDeepSeek,
		"qwen3-max":                   FamilyQwen,
		"gemini-3-pro":                FamilyGemini,
		"sensenova-6-8-flash":         FamilySenseNova,
		"sn-sensenova-6-8-flash-lite": FamilySenseNova,
		"raccoon-chat-ml-5-5":         FamilyDefault,
		"kimi-k3":                     FamilyDefault,
	}
	for model, want := range cases {
		if got := DetectFamily(model); got != want {
			t.Errorf("DetectFamily(%q) = %v, want %v", model, got, want)
		}
	}
}

// TestApplyThinkingDisabledNeverEnables verifies explicit disabled is never
// turned into an enabled state for any family.
func TestApplyThinkingDisabledNeverEnables(t *testing.T) {
	for _, model := range []string{
		"glm-5-3", "deepseek-v4-pro-0813", "qwen3-max", "gemini-3-pro",
		"sensenova-6-8-flash", "raccoon-chat-ml-5-5",
	} {
		out, err := ApplyThinking([]byte(`{}`), model, false)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", model, err)
		}
		raw := string(out)
		if gjson.GetBytes(out, "reasoning_effort").String() == "high" {
			t.Errorf("%s: disabled thinking produced high effort: %s", model, raw)
		}
		if gjson.GetBytes(out, "extra_body.thinking.type").String() == "enabled" {
			t.Errorf("%s: disabled thinking produced enabled: %s", model, raw)
		}
		if gjson.GetBytes(out, "extra_body.enable_thinking").Bool() {
			t.Errorf("%s: disabled thinking produced enable_thinking=true: %s", model, raw)
		}
	}
}

// TestGeminiProTiersNeverReceiveDisableValue covers the documented carve-out.
func TestGeminiProTiersNeverReceiveDisableValue(t *testing.T) {
	for _, model := range []string{"gemini-2.5-pro", "gemini-3.1-pro"} {
		out, err := ApplyThinking([]byte(`{"reasoning_effort":"high"}`), model, false)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", model, err)
		}
		if got := gjson.GetBytes(out, "reasoning_effort").String(); got == "none" {
			t.Errorf("%s: pro tier received explicit disable value: %s", model, out)
		}
	}
}

// TestThinkingEnabledFromConfig maps canonical modes onto the on/off flag.
func TestThinkingEnabledFromConfig(t *testing.T) {
	tests := []struct {
		name   string
		config xhxthinking.ThinkingConfig
		want   bool
	}{
		{"mode none is disabled", xhxthinking.ThinkingConfig{Mode: xhxthinking.ModeNone}, false},
		{"zero budget is disabled", xhxthinking.ThinkingConfig{Mode: xhxthinking.ModeBudget, Budget: 0}, false},
		{"positive budget is enabled", xhxthinking.ThinkingConfig{Mode: xhxthinking.ModeBudget, Budget: 1024}, true},
		{"auto is enabled", xhxthinking.ThinkingConfig{Mode: xhxthinking.ModeAuto}, true},
		{"level is enabled", xhxthinking.ThinkingConfig{Mode: xhxthinking.ModeLevel, Level: xhxthinking.LevelHigh}, true},
		{"zero-value config is treated as unspecified and stays off", xhxthinking.ThinkingConfig{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ThinkingEnabled(tc.config); got != tc.want {
				t.Fatalf("ThinkingEnabled(%+v) = %v, want %v", tc.config, got, tc.want)
			}
		})
	}
}

// TestApplyThinkingForModel mirrors the canonical-config entry point.
func TestApplyThinkingForModel(t *testing.T) {
	out, err := ApplyThinkingForModel([]byte(`{}`), "glm-5-3", xhxthinking.ThinkingConfig{Mode: xhxthinking.ModeLevel, Level: xhxthinking.LevelHigh})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := gjson.GetBytes(out, "reasoning_effort").String(); got != "high" {
		t.Fatalf("reasoning_effort = %q, want high", got)
	}
}

func strPtr(value string) *string { return &value }

// assertDeepEqual compares nested map/slice structures from JSON decoding.
func assertDeepEqual(t *testing.T, label string, got, want any) {
	t.Helper()
	gotJSON, errGot := json.Marshal(got)
	if errGot != nil {
		t.Fatalf("%s: marshal got: %v", label, errGot)
	}
	wantJSON, errWant := json.Marshal(want)
	if errWant != nil {
		t.Fatalf("%s: marshal want: %v", label, errWant)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("%s = %s, want %s", label, gotJSON, wantJSON)
	}
}
