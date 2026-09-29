// Package xiaohuanxiong implements the SenseTime Xiaohuanxiong (商汤小浣熊 /
// Raccoon) deep-thinking dialect for the xiaohuanxiong.com OpenAI-compatible
// gateway.
//
// The gateway fronts several upstream model families and each family expresses
// "deep thinking" differently. This mirrors the box-agent client's
// _apply_thinking_params dialect selection:
//
//	glm-5-3 / glm-5.3           -> reasoning_effort: high | low
//	deepseek / doubao           -> extra_body.thinking.type: enabled | disabled
//	qwen                        -> extra_body.enable_thinking: true | false
//	gemini                      -> reasoning_effort: high | none
//	                               (gemini-2.5-pro / gemini-3.1-pro never
//	                                receive an explicit disable value)
//	sensenova- / sn-sensenova-  -> reasoning_effort: high | none
//	everything else             -> reasoning_effort: high | removed
//
// The gateway speaks the OpenAI wire protocol, so the shared thinking pipeline
// dispatches on the "openai" format and cannot know these per-family dialects.
// The executor therefore applies them directly to the final upstream body.
package xiaohuanxiong

import (
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// reasoningEffortField is the canonical discrete reasoning level field.
const reasoningEffortField = "reasoning_effort"

// Model-family markers, matched case-insensitively as substrings of the model ID.
var (
	glm53Markers          = []string{"glm-5-3", "glm-5.3"}
	deepSeekMarkers       = []string{"deepseek", "doubao"}
	qwenMarkers           = []string{"qwen"}
	geminiMarkers         = []string{"gemini"}
	sensenovaPrefixes     = []string{"sensenova-", "sn-sensenova-"}
	geminiNoDisableMarker = []string{"gemini-2.5-pro", "gemini-3.1-pro"}
)

// ModelFamily identifies the deep-thinking dialect an upstream model expects.
type ModelFamily int

const (
	// FamilyDefault uses reasoning_effort and drops the field when disabled.
	FamilyDefault ModelFamily = iota
	// FamilyGLM53 uses a two-value reasoning_effort switch (high/low).
	FamilyGLM53
	// FamilyDeepSeek uses extra_body.thinking.type (enabled/disabled).
	FamilyDeepSeek
	// FamilyQwen uses extra_body.enable_thinking (boolean).
	FamilyQwen
	// FamilyGemini uses reasoning_effort and never disables the pro tiers.
	FamilyGemini
	// FamilySenseNova uses reasoning_effort with an explicit none when disabled.
	FamilySenseNova
)

// DetectFamily classifies a model ID into its thinking dialect.
func DetectFamily(model string) ModelFamily {
	normalized := strings.ToLower(strings.TrimSpace(model))
	switch {
	case containsAny(normalized, glm53Markers):
		return FamilyGLM53
	case containsAny(normalized, deepSeekMarkers):
		return FamilyDeepSeek
	case containsAny(normalized, qwenMarkers):
		return FamilyQwen
	case containsAny(normalized, geminiMarkers):
		return FamilyGemini
	case hasAnyPrefix(normalized, sensenovaPrefixes):
		return FamilySenseNova
	default:
		return FamilyDefault
	}
}

// ThinkingEnabled maps a canonical ThinkingConfig onto the upstream on/off flag.
//
// ModeNone and an explicit zero budget both mean disabled; every other mode
// (level, budget, auto) requests provider-side thinking.
func ThinkingEnabled(config thinking.ThinkingConfig) bool {
	switch config.Mode {
	case thinking.ModeNone:
		return false
	case thinking.ModeBudget:
		return config.Budget != 0
	default:
		return true
	}
}

// ApplyThinking rewrites body with the thinking dialect required by model.
//
// enabled selects the on/off branch; each family decides how that maps onto the
// wire. Bodies that are empty or not valid JSON are treated as an empty object.
func ApplyThinking(body []byte, model string, enabled bool) ([]byte, error) {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		body = []byte(`{}`)
	}
	normalized := strings.ToLower(strings.TrimSpace(model))
	sensenova := hasAnyPrefix(normalized, sensenovaPrefixes)

	switch DetectFamily(model) {
	case FamilyGLM53:
		// GLM-5.3 exposes a two-value effort switch only.
		return setString(body, reasoningEffortField, effortFor(enabled, "high", "low"))

	case FamilyDeepSeek:
		// DeepSeek/Doubao use the nested thinking object.
		return setString(body, "extra_body.thinking.type", effortFor(enabled, "enabled", "disabled"))

	case FamilyQwen:
		return setBool(body, "extra_body.enable_thinking", enabled)

	case FamilyGemini:
		if enabled {
			return setString(body, reasoningEffortField, "high")
		}
		// The pro tiers reject an explicit disable value.
		if containsAny(normalized, geminiNoDisableMarker) {
			return body, nil
		}
		return setString(body, reasoningEffortField, string(thinking.LevelNone))
	}

	if sensenova {
		// SenseNova shares the generic effort field but pins the disabled value
		// to the explicit "none" level instead of dropping the field.
		return setString(body, reasoningEffortField, effortFor(enabled, string(thinking.LevelHigh), string(thinking.LevelNone)))
	}

	// Default family: enable with high effort, disable by removing the field.
	if enabled {
		return setString(body, reasoningEffortField, string(thinking.LevelHigh))
	}
	return deleteField(body, reasoningEffortField)
}

// ApplyThinkingForModel derives the on/off flag from a canonical ThinkingConfig
// and applies the model family dialect.
func ApplyThinkingForModel(body []byte, model string, config thinking.ThinkingConfig) ([]byte, error) {
	return ApplyThinking(body, model, ThinkingEnabled(config))
}

// TranslateRequestThinking rewrites the generic OpenAI thinking fields on a
// translated upstream body into the dialect required by model.
//
// The shared thinking pipeline runs before this point and, because the gateway
// speaks the OpenAI protocol, emits a generic `reasoning_effort`. That single
// value is re-expressed per family here: DeepSeek/Doubao and Qwen move it into
// `extra_body`, while the effort-valued families keep or drop it.
//
// A body without `reasoning_effort` carries no thinking intent and is returned
// unchanged, so a request that never asked for thinking stays untouched.
func TranslateRequestThinking(body []byte, model string) ([]byte, error) {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return body, nil
	}
	effort := gjson.GetBytes(body, reasoningEffortField)
	if !effort.Exists() {
		return body, nil
	}

	enabled := !strings.EqualFold(strings.TrimSpace(effort.String()), string(thinking.LevelNone))

	var updated []byte
	var errApply error
	switch DetectFamily(model) {
	case FamilyDeepSeek, FamilyQwen:
		// These families express thinking through extra_body, so the generic
		// top-level field must not leak to the upstream.
		updated, errApply = ApplyThinking(body, model, enabled)
		if errApply != nil {
			return body, errApply
		}
		stripped, errStrip := deleteField(updated, reasoningEffortField)
		if errStrip != nil {
			return updated, nil
		}
		return stripped, nil
	default:
		return ApplyThinking(body, model, enabled)
	}
}

// effortFor returns enabled or disabled depending on the thinking flag.
func effortFor(enabled bool, on, off string) string {
	if enabled {
		return on
	}
	return off
}

// setString sets a JSON string field, ignoring sjson's typed error.
func setString(body []byte, path, value string) ([]byte, error) {
	updated, errSet := sjson.SetBytes(body, path, value)
	if errSet != nil {
		return body, nil
	}
	return updated, nil
}

// setBool sets a JSON boolean field, ignoring sjson's typed error.
func setBool(body []byte, path string, value bool) ([]byte, error) {
	updated, errSet := sjson.SetBytes(body, path, value)
	if errSet != nil {
		return body, nil
	}
	return updated, nil
}

// deleteField removes a field, ignoring sjson's typed error.
func deleteField(body []byte, path string) ([]byte, error) {
	updated, errDelete := sjson.DeleteBytes(body, path)
	if errDelete != nil {
		return body, nil
	}
	return updated, nil
}

// containsAny reports whether haystack contains any of the markers.
func containsAny(haystack string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(haystack, marker) {
			return true
		}
	}
	return false
}

// hasAnyPrefix reports whether haystack starts with any of the prefixes.
func hasAnyPrefix(haystack string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(haystack, prefix) {
			return true
		}
	}
	return false
}

// FamilyFromModelInfo is a convenience wrapper for registry metadata.
func FamilyFromModelInfo(modelInfo *registry.ModelInfo) ModelFamily {
	if modelInfo == nil {
		return FamilyDefault
	}
	return DetectFamily(modelInfo.ID)
}
