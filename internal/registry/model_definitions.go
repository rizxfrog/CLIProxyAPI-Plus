// Package registry provides model definitions and lookup helpers for various AI providers.
// Static model metadata is loaded from the embedded models.json file and can be refreshed from network.
package registry

import (
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/constant"
)

const (
	codexBuiltinImage15ModelID         = "gpt-image-1.5"
	codexBuiltinImageModelID           = "gpt-image-2"
	codexBuiltinImage25FlareModelID    = "gpt-image-2.5-flare"
	codexBuiltinImage25SunburstModelID = "gpt-image-2.5-sunburst"
	codexBuiltinImage25ModelID         = "gpt-image-2.5"
	xaiBuiltinImageModelID             = "grok-imagine-image"
	xaiBuiltinImageQualityModelID      = "grok-imagine-image-quality"
	xaiBuiltinImage20ModelID           = "grok-imagine-image-2.0"
	xaiBuiltinVideoModelID             = "grok-imagine-video"
	xaiBuiltinVideo15ModelID           = "grok-imagine-video-1.5"
	xaiBuiltinVideo15PreviewID         = "grok-imagine-video-1.5-preview"
)

// staticModelsJSON mirrors the top-level structure of models.json.
type staticModelsJSON struct {
	Claude      []*ModelInfo `json:"claude"`
	Gemini      []*ModelInfo `json:"gemini"`
	Vertex      []*ModelInfo `json:"vertex"`
	AIStudio    []*ModelInfo `json:"aistudio"`
	CodexFree   []*ModelInfo `json:"codex-free"`
	CodexTeam   []*ModelInfo `json:"codex-team"`
	CodexPlus   []*ModelInfo `json:"codex-plus"`
	CodexPro    []*ModelInfo `json:"codex-pro"`
	Kimi        []*ModelInfo `json:"kimi"`
	Antigravity []*ModelInfo `json:"antigravity"`
	CodeBuddyCN []*ModelInfo `json:"codebuddy-cn"`
	CodeBuddyAI []*ModelInfo `json:"codebuddy-ai"`
	DeepSeekWeb []*ModelInfo `json:"deepseek-web"`
	XAI         []*ModelInfo `json:"xai"`
	Devin       []*ModelInfo `json:"devin"`
	Trae        []*ModelInfo `json:"trae"`
	Cline       []*ModelInfo `json:"cline"`
	QoderCN     []*ModelInfo `json:"qoder-cn"`
	QoderAI     []*ModelInfo `json:"qoder-ai"`
	Meta        []*ModelInfo `json:"meta"`
}

// GetClaudeModels returns the standard Claude model definitions.
func GetClaudeModels() []*ModelInfo {
	return cloneModelInfos(getModels().Claude)
}

// GetGeminiModels returns the standard Gemini model definitions.
func GetGeminiModels() []*ModelInfo {
	return cloneModelInfos(getModels().Gemini)
}

// GetGeminiVertexModels returns Gemini model definitions for Vertex AI.
func GetGeminiVertexModels() []*ModelInfo {
	return cloneModelInfos(getModels().Vertex)
}

// GetAIStudioModels returns model definitions for AI Studio.
func GetAIStudioModels() []*ModelInfo {
	return cloneModelInfos(getModels().AIStudio)
}

// GetCodexFreeModels returns model definitions for the Codex free plan tier.
func GetCodexFreeModels() []*ModelInfo {
	return WithCodexBuiltins(cloneModelInfos(getModels().CodexFree))
}

// GetCodexTeamModels returns model definitions for the Codex team plan tier.
func GetCodexTeamModels() []*ModelInfo {
	return WithCodexBuiltins(cloneModelInfos(getModels().CodexTeam))
}

// GetCodexPlusModels returns model definitions for the Codex plus plan tier.
func GetCodexPlusModels() []*ModelInfo {
	return WithCodexBuiltins(cloneModelInfos(getModels().CodexPlus))
}

// GetCodexProModels returns model definitions for the Codex pro plan tier.
func GetCodexProModels() []*ModelInfo {
	return WithCodexBuiltins(cloneModelInfos(getModels().CodexPro))
}

// GetKimiModels returns the standard Kimi (Moonshot AI) model definitions.
func GetKimiModels() []*ModelInfo {
	return cloneModelInfos(getModels().Kimi)
}

// GetCodeBuddyCNModels returns the standard CodeBuddy CN (Tencent) model definitions.
func GetCodeBuddyCNModels() []*ModelInfo {
	return cloneModelInfos(getModels().CodeBuddyCN)
}

// GetCodeBuddyAIModels returns the standard CodeBuddy AI (international) model definitions.
func GetCodeBuddyAIModels() []*ModelInfo {
	return cloneModelInfos(getModels().CodeBuddyAI)
}

// GetDeepSeekWebModels returns the standard DeepSeek Web model definitions.
func GetDeepSeekWebModels() []*ModelInfo {
	return cloneModelInfos(getModels().DeepSeekWeb)
}

// GetAntigravityModels returns the standard Antigravity model definitions.
func GetAntigravityModels() []*ModelInfo {
	return cloneModelInfos(getModels().Antigravity)
}

var staticDevinModels = []*ModelInfo{
	devinBuiltinSWE16SlowModelInfo(),
	{
		ID:                  "devin/swe-2",
		Type:                "devin",
		OwnedBy:             "cognition",
		DisplayName:         "SWE-2",
		ContextLength:       262000,
		MaxCompletionTokens: 128000,
		Thinking: &ThinkingSupport{
			Levels: []string{"medium", "high", "max"},
		},
	},
	{
		ID:                  "devin/claude-fable-5-1",
		Type:                "devin",
		OwnedBy:             "anthropic",
		DisplayName:         "Claude Fable 5.1",
		ContextLength:       1000000,
		MaxCompletionTokens: 64000,
		Thinking: &ThinkingSupport{
			Levels: []string{"low", "medium", "high", "xhigh", "max"},
		},
	},
	{
		ID:                  "devin/gpt-6-astra",
		Type:                "devin",
		OwnedBy:             "openai",
		DisplayName:         "GPT-6 Astra",
		ContextLength:       1000000,
		MaxCompletionTokens: 64000,
		Thinking: &ThinkingSupport{
			Levels: []string{"low", "medium", "high", "xhigh", "max"},
		},
	},
	{
		ID:                  "devin/glm-5-2",
		Type:                "devin",
		OwnedBy:             "zhipu",
		DisplayName:         "GLM-5.2",
		ContextLength:       200000,
		MaxCompletionTokens: 64000,
		Thinking: &ThinkingSupport{
			Levels: []string{"none", "high"},
		},
	},
	{
		ID:                  "devin/glm-5-3",
		Type:                "devin",
		OwnedBy:             "zhipu",
		DisplayName:         "GLM-5.3",
		ContextLength:       1048576,
		MaxCompletionTokens: 128000,
		Thinking: &ThinkingSupport{
			Levels: []string{"low", "high", "max"},
		},
	},
	{
		ID:                  "devin/glm-5-3-flash",
		Type:                "devin",
		OwnedBy:             "zhipu",
		DisplayName:         "GLM-5.3 Flash",
		ContextLength:       1000000,
		MaxCompletionTokens: 128000,
		Thinking: &ThinkingSupport{
			Levels: []string{"low", "high", "max"},
		},
	},
	{
		ID:                  "devin/gpt-5-6-sol",
		Type:                "devin",
		OwnedBy:             "openai",
		DisplayName:         "GPT-5.6 Sol",
		ContextLength:       1000000,
		MaxCompletionTokens: 128000,
		Thinking: &ThinkingSupport{
			Levels: []string{"none", "low", "medium", "high", "xhigh", "max"},
		},
	},
	{
		ID:                  "devin/gemini-3-8-flash",
		Type:                "devin",
		OwnedBy:             "google",
		DisplayName:         "Gemini 3.8 Flash",
		ContextLength:       1048576,
		MaxCompletionTokens: 65536,
		Thinking: &ThinkingSupport{
			Levels: []string{"low", "medium", "high"},
		},
	},
	{
		ID:                  "devin/grok-4-6",
		Type:                "devin",
		OwnedBy:             "xai",
		DisplayName:         "Grok 4.6",
		ContextLength:       500000,
		MaxCompletionTokens: 131072,
		Thinking: &ThinkingSupport{
			Levels: []string{"low", "medium", "high", "xhigh"},
		},
	},
	{
		ID:                  "devin/deepseek-v4-flash",
		Type:                "devin",
		OwnedBy:             "deepseek",
		DisplayName:         "DeepSeek V4 Flash",
		ContextLength:       1048576,
		MaxCompletionTokens: 64000,
		Thinking: &ThinkingSupport{
			Levels: []string{"high", "max"},
		},
	},
	{
		ID:                  "devin/deepseek-v4-1-flash",
		Type:                "devin",
		OwnedBy:             "deepseek",
		DisplayName:         "DeepSeek V4.1 Flash",
		ContextLength:       1048576,
		MaxCompletionTokens: 64000,
		Thinking: &ThinkingSupport{
			Levels: []string{"high", "max"},
		},
	},
}

// AntigravityWebSearchModelFor returns the Antigravity model that should run a
// native web search request for modelID.
func AntigravityWebSearchModelFor(modelID string) string {
	modelID = normalizeAntigravityCapabilityModelID(modelID)
	if modelID == "" {
		return ""
	}
	for _, model := range GetGlobalRegistry().GetAvailableModelsByProvider("antigravity") {
		if model == nil {
			continue
		}
		currentModelID := normalizeAntigravityCapabilityModelID(model.ID)
		if currentModelID == "" {
			continue
		}
		if currentModelID == modelID {
			if model.SupportsWebSearch {
				return currentModelID
			}
			return ""
		}
	}
	return ""
}

// GetXAIModels returns the standard xAI Grok model definitions.
func GetXAIModels() []*ModelInfo {
	return WithXAIBuiltins(cloneModelInfos(getModels().XAI))
}

// GetTraeModels returns the standard TRAE SOLO CN model definitions.
func GetTraeModels() []*ModelInfo {
	return cloneModelInfos(getModels().Trae)
}

// GetClineModels returns the standard Cline (cline.bot) model definitions.
func GetClineModels() []*ModelInfo {
	return WithClineBuiltins(cloneModelInfos(getModels().Cline))
}

// clineBuiltinModels is the offline fallback catalog for the Cline
// (api.cline.bot) OpenAI-compatible gateway. It is seeded from the live
// GET /api/v1/models response (458 models, 2026-09-26) and only fills IDs
// the remote catalog omits. Cline's /api/v1/models endpoint returns model
// ids without context limits, so ContextLength/MaxCompletionTokens are left 0;
// the executor applies provider defaults at request time.
var clineBuiltinModels = []*ModelInfo{
	{ID: "typesafe/jev-router", Object: "model", OwnedBy: "typesafe", Type: "cline", DisplayName: "jev-router"},
	{ID: "perceptron/perceptron-mk1.5", Object: "model", OwnedBy: "perceptron", Type: "cline", DisplayName: "perceptron-mk1.5"},
	{ID: "fireworks/ember-1", Object: "model", OwnedBy: "fireworks", Type: "cline", DisplayName: "ember-1"},
	{ID: "z-ai/glm-5.3-prime", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.3-prime"},
	{ID: "qwen/qwen3.8-max-prime", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-max-prime"},
	{ID: "stealth/space-bunny-alpha", Object: "model", OwnedBy: "stealth", Type: "cline", DisplayName: "space-bunny-alpha"},
	{ID: "aion-labs/aion-3.5-mini", Object: "model", OwnedBy: "aion-labs", Type: "cline", DisplayName: "aion-3.5-mini"},
	{ID: "aion-labs/aion-3.5", Object: "model", OwnedBy: "aion-labs", Type: "cline", DisplayName: "aion-3.5"},
	{ID: "upstage/solar-mini4", Object: "model", OwnedBy: "upstage", Type: "cline", DisplayName: "solar-mini4"},
	{ID: "cohere/command-a-plus", Object: "model", OwnedBy: "cohere", Type: "cline", DisplayName: "command-a-plus"},
	{ID: "openai/gpt-6-luna-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-luna-pro"},
	{ID: "openai/gpt-6-luna-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-luna-pro:batch"},
	{ID: "openai/gpt-6-luna", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-luna"},
	{ID: "openai/gpt-6-luna:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-luna:batch"},
	{ID: "openai/gpt-6-sol-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-sol-pro"},
	{ID: "openai/gpt-6-sol-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-sol-pro:batch"},
	{ID: "openai/gpt-6-sol", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-sol"},
	{ID: "openai/gpt-6-sol:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-sol:batch"},
	{ID: "anthropic/claude-opus-5.5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-5.5"},
	{ID: "anthropic/claude-opus-5.5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-5.5:batch"},
	{ID: "xiaomi/mimo-v2.6-pro-ultraspeed", Object: "model", OwnedBy: "xiaomi", Type: "cline", DisplayName: "mimo-v2.6-pro-ultraspeed"},
	{ID: "xiaomi/mimo-v2.6-flash", Object: "model", OwnedBy: "xiaomi", Type: "cline", DisplayName: "mimo-v2.6-flash"},
	{ID: "xiaomi/mimo-v2.6-pro", Object: "model", OwnedBy: "xiaomi", Type: "cline", DisplayName: "mimo-v2.6-pro"},
	{ID: "x-ai/grok-4.7", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.7"},
	{ID: "qwen/qwen3.8-omni-flash", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-omni-flash"},
	{ID: "prism-ml/ternary-bonsai-2-27b", Object: "model", OwnedBy: "prism-ml", Type: "cline", DisplayName: "ternary-bonsai-2-27b"},
	{ID: "z-ai/glm-5.3-flashx", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.3-flashx"},
	{ID: "unbiased/pareto", Object: "model", OwnedBy: "unbiased", Type: "cline", DisplayName: "pareto"},
	{ID: "~deepseek/deepseek-pro-latest", Object: "model", OwnedBy: "~deepseek", Type: "cline", DisplayName: "deepseek-pro-latest"},
	{ID: "~deepseek/deepseek-flash-latest", Object: "model", OwnedBy: "~deepseek", Type: "cline", DisplayName: "deepseek-flash-latest"},
	{ID: "inference-net/schematron-v2-turbo", Object: "model", OwnedBy: "inference-net", Type: "cline", DisplayName: "schematron-v2-turbo"},
	{ID: "inference-net/schematron-v2-small", Object: "model", OwnedBy: "inference-net", Type: "cline", DisplayName: "schematron-v2-small"},
	{ID: "~openai/gpt-astra-latest", Object: "model", OwnedBy: "~openai", Type: "cline", DisplayName: "gpt-astra-latest"},
	{ID: "~openai/gpt-sol-latest", Object: "model", OwnedBy: "~openai", Type: "cline", DisplayName: "gpt-sol-latest"},
	{ID: "~openai/gpt-terra-latest", Object: "model", OwnedBy: "~openai", Type: "cline", DisplayName: "gpt-terra-latest"},
	{ID: "~openai/gpt-luna-latest", Object: "model", OwnedBy: "~openai", Type: "cline", DisplayName: "gpt-luna-latest"},
	{ID: "sakana/fugu-ultra-v2", Object: "model", OwnedBy: "sakana", Type: "cline", DisplayName: "fugu-ultra-v2"},
	{ID: "sakana/fugu-max", Object: "model", OwnedBy: "sakana", Type: "cline", DisplayName: "fugu-max"},
	{ID: "inclusionai/ling-3.0-flash-vl", Object: "model", OwnedBy: "inclusionai", Type: "cline", DisplayName: "ling-3.0-flash-vl"},
	{ID: "deepseek/deepseek-v4.1-flash", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4.1-flash"},
	{ID: "deepseek/deepseek-v4.1-flash:batch", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4.1-flash:batch"},
	{ID: "inception/mercury-2.5", Object: "model", OwnedBy: "inception", Type: "cline", DisplayName: "mercury-2.5"},
	{ID: "openai/gpt-6-astra", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-astra"},
	{ID: "openai/gpt-6-astra:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-astra:batch"},
	{ID: "openai/gpt-6-astra-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-astra-pro"},
	{ID: "openai/gpt-6-astra-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-6-astra-pro:batch"},
	{ID: "inclusionai/ling-3.0-flash-sante:free", Object: "model", OwnedBy: "inclusionai", Type: "cline", DisplayName: "ling-3.0-flash-sante:free"},
	{ID: "qwen/qwen3.8-max-0902", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-max-0902"},
	{ID: "meta/muse-spark-1.3-contributor", Object: "model", OwnedBy: "meta", Type: "cline", DisplayName: "muse-spark-1.3-contributor"},
	{ID: "meta/muse-spark-1.3", Object: "model", OwnedBy: "meta", Type: "cline", DisplayName: "muse-spark-1.3"},
	{ID: "google/gemini-3.8-flash", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.8-flash"},
	{ID: "google/gemini-3.8-flash:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.8-flash:batch"},
	{ID: "anthropic/claude-fable-5.1", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-fable-5.1"},
	{ID: "anthropic/claude-fable-5.1:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-fable-5.1:batch"},
	{ID: "ibm-granite/granite-4.2-8b", Object: "model", OwnedBy: "ibm-granite", Type: "cline", DisplayName: "granite-4.2-8b"},
	{ID: "tencent/hy4-preview", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hy4-preview"},
	{ID: "inclusionai/ling-3.0-flash-fin", Object: "model", OwnedBy: "inclusionai", Type: "cline", DisplayName: "ling-3.0-flash-fin"},
	{ID: "inclusionai/ling-3.0-flash-fin:free", Object: "model", OwnedBy: "inclusionai", Type: "cline", DisplayName: "ling-3.0-flash-fin:free"},
	{ID: "~z-ai/glm-flash-latest", Object: "model", OwnedBy: "~z-ai", Type: "cline", DisplayName: "glm-flash-latest"},
	{ID: "qwen/qwen3.8-flash", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-flash"},
	{ID: "z-ai/glm-5.3-flash", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.3-flash"},
	{ID: "z-ai/glm-5.3-flash:batch", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.3-flash:batch"},
	{ID: "meta/muse-spark-1.2-contributor", Object: "model", OwnedBy: "meta", Type: "cline", DisplayName: "muse-spark-1.2-contributor"},
	{ID: "deepseek/deepseek-v4-flash-vision-exp", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4-flash-vision-exp"},
	{ID: "tencent/hy-mt2-1.8b", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hy-mt2-1.8b"},
	{ID: "tencent/hy-mt2-30b-a3b", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hy-mt2-30b-a3b"},
	{ID: "~z-ai/glm-latest", Object: "model", OwnedBy: "~z-ai", Type: "cline", DisplayName: "glm-latest"},
	{ID: "tencent/hy-mt2-7b", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hy-mt2-7b"},
	{ID: "z-ai/glm-5.3", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.3"},
	{ID: "z-ai/glm-5.3:batch", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.3:batch"},
	{ID: "qwen/qwen3.8-27b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-27b"},
	{ID: "qwen/qwen3.8-27b:free", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-27b:free"},
	{ID: "dots-studio/dots-3-note-preview:free", Object: "model", OwnedBy: "dots-studio", Type: "cline", DisplayName: "dots-3-note-preview:free"},
	{ID: "google/gemini-3.7-flash", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.7-flash"},
	{ID: "google/gemini-3.7-flash:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.7-flash:batch"},
	{ID: "bytedance-seed/seed-2-1-turbo", Object: "model", OwnedBy: "bytedance-seed", Type: "cline", DisplayName: "seed-2-1-turbo"},
	{ID: "qwen/qwen3.8-2.4t-a95b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.8-2.4t-a95b"},
	{ID: "bytedance-seed/seed-2.0-code", Object: "model", OwnedBy: "bytedance-seed", Type: "cline", DisplayName: "seed-2.0-code"},
	{ID: "deepseek/deepseek-v4-pro-0813", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4-pro-0813"},
	{ID: "x-ai/grok-4.6", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.6"},
	{ID: "liquid/lfm-2.5-2.6b:free", Object: "model", OwnedBy: "liquid", Type: "cline", DisplayName: "lfm-2.5-2.6b:free"},
	{ID: "nvidia/nemotron-3.5-lightning", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3.5-lightning"},
	{ID: "nvidia/nemotron-3.5-lightning:free", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3.5-lightning:free"},
	{ID: "sakana/sakana-namazu", Object: "model", OwnedBy: "sakana", Type: "cline", DisplayName: "sakana-namazu"},
	{ID: "upstage/solar-pro4", Object: "model", OwnedBy: "upstage", Type: "cline", DisplayName: "solar-pro4"},
	{ID: "meta/muse-glimmer-30b", Object: "model", OwnedBy: "meta", Type: "cline", DisplayName: "muse-glimmer-30b"},
	{ID: "meta/muse-spark-1.2", Object: "model", OwnedBy: "meta", Type: "cline", DisplayName: "muse-spark-1.2"},
	{ID: "~deepseek/deepseek-v4-flash-latest", Object: "model", OwnedBy: "~deepseek", Type: "cline", DisplayName: "deepseek-v4-flash-latest"},
	{ID: "deepseek/deepseek-v4-flash-0731", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4-flash-0731"},
	{ID: "thinkingmachines/inkling-small", Object: "model", OwnedBy: "thinkingmachines", Type: "cline", DisplayName: "inkling-small"},
	{ID: "thinkingmachines/inkling-small:free", Object: "model", OwnedBy: "thinkingmachines", Type: "cline", DisplayName: "inkling-small:free"},
	{ID: "qwen/qwen3.7-flash", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.7-flash"},
	{ID: "anthropic/claude-opus-5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-5"},
	{ID: "anthropic/claude-opus-5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-5:batch"},
	{ID: "inclusionai/ling-3.0-flash", Object: "model", OwnedBy: "inclusionai", Type: "cline", DisplayName: "ling-3.0-flash"},
	{ID: "poolside/laguna-s-2.1", Object: "model", OwnedBy: "poolside", Type: "cline", DisplayName: "laguna-s-2.1"},
	{ID: "poolside/laguna-s-2.1:free", Object: "model", OwnedBy: "poolside", Type: "cline", DisplayName: "laguna-s-2.1:free"},
	{ID: "google/gemini-3.6-flash", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.6-flash"},
	{ID: "google/gemini-3.6-flash:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.6-flash:batch"},
	{ID: "google/gemini-3.5-flash-lite", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.5-flash-lite"},
	{ID: "google/gemini-3.5-flash-lite:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.5-flash-lite:batch"},
	{ID: "meituan/longcat-2.0", Object: "model", OwnedBy: "meituan", Type: "cline", DisplayName: "longcat-2.0"},
	{ID: "thinkingmachines/inkling", Object: "model", OwnedBy: "thinkingmachines", Type: "cline", DisplayName: "inkling"},
	{ID: "thinkingmachines/inkling:free", Object: "model", OwnedBy: "thinkingmachines", Type: "cline", DisplayName: "inkling:free"},
	{ID: "openrouter/auto-beta", Object: "model", OwnedBy: "openrouter", Type: "cline", DisplayName: "auto-beta"},
	{ID: "moonshotai/kimi-k3", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k3"},
	{ID: "moonshotai/kimi-k3:batch", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k3:batch"},
	{ID: "meta/muse-spark-1.1", Object: "model", OwnedBy: "meta", Type: "cline", DisplayName: "muse-spark-1.1"},
	{ID: "kwaipilot/kat-coder-pro-v2.5", Object: "model", OwnedBy: "kwaipilot", Type: "cline", DisplayName: "kat-coder-pro-v2.5"},
	{ID: "openai/gpt-5.6-luna-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-luna-pro"},
	{ID: "openai/gpt-5.6-luna-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-luna-pro:batch"},
	{ID: "openai/gpt-5.6-luna", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-luna"},
	{ID: "openai/gpt-5.6-luna:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-luna:batch"},
	{ID: "openai/gpt-5.6-terra-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-terra-pro"},
	{ID: "openai/gpt-5.6-terra-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-terra-pro:batch"},
	{ID: "openai/gpt-5.6-terra", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-terra"},
	{ID: "openai/gpt-5.6-terra:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-terra:batch"},
	{ID: "openai/gpt-5.6-sol-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-sol-pro"},
	{ID: "openai/gpt-5.6-sol-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-sol-pro:batch"},
	{ID: "openai/gpt-5.6-sol", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-sol"},
	{ID: "openai/gpt-5.6-sol:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.6-sol:batch"},
	{ID: "x-ai/grok-4.5", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.5"},
	{ID: "~x-ai/grok-latest", Object: "model", OwnedBy: "~x-ai", Type: "cline", DisplayName: "grok-latest"},
	{ID: "aion-labs/aion-3.0-mini", Object: "model", OwnedBy: "aion-labs", Type: "cline", DisplayName: "aion-3.0-mini"},
	{ID: "aion-labs/aion-3.0", Object: "model", OwnedBy: "aion-labs", Type: "cline", DisplayName: "aion-3.0"},
	{ID: "tencent/hy3", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hy3"},
	{ID: "poolside/laguna-xs-2.1", Object: "model", OwnedBy: "poolside", Type: "cline", DisplayName: "laguna-xs-2.1"},
	{ID: "poolside/laguna-xs-2.1:free", Object: "model", OwnedBy: "poolside", Type: "cline", DisplayName: "laguna-xs-2.1:free"},
	{ID: "anthropic/claude-sonnet-5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-5"},
	{ID: "anthropic/claude-sonnet-5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-5:batch"},
	{ID: "google/gemini-3.1-flash-lite-image", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-flash-lite-image"},
	{ID: "sakana/fugu-ultra", Object: "model", OwnedBy: "sakana", Type: "cline", DisplayName: "fugu-ultra"},
	{ID: "google/gemini-3.1-flash-image", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-flash-image"},
	{ID: "google/gemini-3-pro-image", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3-pro-image"},
	{ID: "cohere/north-mini-code:free", Object: "model", OwnedBy: "cohere", Type: "cline", DisplayName: "north-mini-code:free"},
	{ID: "z-ai/glm-5.2", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.2"},
	{ID: "openrouter/fusion", Object: "model", OwnedBy: "openrouter", Type: "cline", DisplayName: "fusion"},
	{ID: "moonshotai/kimi-k2.7-code", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k2.7-code"},
	{ID: "~anthropic/claude-fable-latest", Object: "model", OwnedBy: "~anthropic", Type: "cline", DisplayName: "claude-fable-latest"},
	{ID: "anthropic/claude-fable-5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-fable-5"},
	{ID: "anthropic/claude-fable-5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-fable-5:batch"},
	{ID: "nvidia/nemotron-3.5-content-safety", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3.5-content-safety"},
	{ID: "nvidia/nemotron-3.5-content-safety:free", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3.5-content-safety:free"},
	{ID: "nvidia/nemotron-3-ultra-550b-a55b", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3-ultra-550b-a55b"},
	{ID: "nvidia/nemotron-3-ultra-550b-a55b:free", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3-ultra-550b-a55b:free"},
	{ID: "qwen/qwen3.7-plus", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.7-plus"},
	{ID: "minimax/minimax-m3", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m3"},
	{ID: "stepfun/step-3.7-flash", Object: "model", OwnedBy: "stepfun", Type: "cline", DisplayName: "step-3.7-flash"},
	{ID: "anthropic/claude-opus-4.8", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.8"},
	{ID: "anthropic/claude-opus-4.8:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.8:batch"},
	{ID: "qwen/qwen3.7-max", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.7-max"},
	{ID: "x-ai/grok-build-0.1", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-build-0.1"},
	{ID: "google/gemini-3.5-flash", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.5-flash"},
	{ID: "google/gemini-3.5-flash:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.5-flash:batch"},
	{ID: "perceptron/perceptron-mk1", Object: "model", OwnedBy: "perceptron", Type: "cline", DisplayName: "perceptron-mk1"},
	{ID: "google/gemini-3.1-flash-lite", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-flash-lite"},
	{ID: "google/gemini-3.1-flash-lite:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-flash-lite:batch"},
	{ID: "openai/gpt-chat-latest", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-chat-latest"},
	{ID: "x-ai/grok-4.3", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.3"},
	{ID: "x-ai/grok-4.3:batch", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.3:batch"},
	{ID: "mistralai/mistral-medium-3-5", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-medium-3-5"},
	{ID: "mistralai/mistral-medium-3-5:batch", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-medium-3-5:batch"},
	{ID: "nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3-nano-omni-30b-a3b-reasoning:free"},
	{ID: "~anthropic/claude-haiku-latest", Object: "model", OwnedBy: "~anthropic", Type: "cline", DisplayName: "claude-haiku-latest"},
	{ID: "~openai/gpt-mini-latest", Object: "model", OwnedBy: "~openai", Type: "cline", DisplayName: "gpt-mini-latest"},
	{ID: "~google/gemini-pro-latest", Object: "model", OwnedBy: "~google", Type: "cline", DisplayName: "gemini-pro-latest"},
	{ID: "~moonshotai/kimi-latest", Object: "model", OwnedBy: "~moonshotai", Type: "cline", DisplayName: "kimi-latest"},
	{ID: "~google/gemini-flash-latest", Object: "model", OwnedBy: "~google", Type: "cline", DisplayName: "gemini-flash-latest"},
	{ID: "~anthropic/claude-sonnet-latest", Object: "model", OwnedBy: "~anthropic", Type: "cline", DisplayName: "claude-sonnet-latest"},
	{ID: "qwen/qwen3.5-plus-20260420", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-plus-20260420"},
	{ID: "qwen/qwen3.6-flash", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.6-flash"},
	{ID: "qwen/qwen3.6-35b-a3b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.6-35b-a3b"},
	{ID: "qwen/qwen3.6-max-preview", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.6-max-preview"},
	{ID: "qwen/qwen3.6-27b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.6-27b"},
	{ID: "openai/gpt-5.5-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.5-pro"},
	{ID: "openai/gpt-5.5-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.5-pro:batch"},
	{ID: "openai/gpt-5.5", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.5"},
	{ID: "openai/gpt-5.5:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.5:batch"},
	{ID: "deepseek/deepseek-v4-pro", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4-pro"},
	{ID: "deepseek/deepseek-v4-flash", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v4-flash"},
	{ID: "tencent/hy3-preview", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hy3-preview"},
	{ID: "xiaomi/mimo-v2.5-pro", Object: "model", OwnedBy: "xiaomi", Type: "cline", DisplayName: "mimo-v2.5-pro"},
	{ID: "xiaomi/mimo-v2.5", Object: "model", OwnedBy: "xiaomi", Type: "cline", DisplayName: "mimo-v2.5"},
	{ID: "openai/gpt-5.4-image-2", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-image-2"},
	{ID: "~anthropic/claude-opus-latest", Object: "model", OwnedBy: "~anthropic", Type: "cline", DisplayName: "claude-opus-latest"},
	{ID: "openrouter/pareto-code", Object: "model", OwnedBy: "openrouter", Type: "cline", DisplayName: "pareto-code"},
	{ID: "moonshotai/kimi-k2.6", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k2.6"},
	{ID: "anthropic/claude-opus-4.7", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.7"},
	{ID: "anthropic/claude-opus-4.7:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.7:batch"},
	{ID: "z-ai/glm-5.1", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5.1"},
	{ID: "google/gemma-4-26b-a4b-it", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-4-26b-a4b-it"},
	{ID: "google/gemma-4-26b-a4b-it:free", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-4-26b-a4b-it:free"},
	{ID: "google/gemma-4-31b-it", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-4-31b-it"},
	{ID: "google/gemma-4-31b-it:free", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-4-31b-it:free"},
	{ID: "qwen/qwen3.6-plus", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.6-plus"},
	{ID: "z-ai/glm-5v-turbo", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5v-turbo"},
	{ID: "arcee-ai/trinity-large-thinking", Object: "model", OwnedBy: "arcee-ai", Type: "cline", DisplayName: "trinity-large-thinking"},
	{ID: "x-ai/grok-4.20-multi-agent", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.20-multi-agent"},
	{ID: "x-ai/grok-4.20", Object: "model", OwnedBy: "x-ai", Type: "cline", DisplayName: "grok-4.20"},
	{ID: "google/lyria-3-pro-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "lyria-3-pro-preview"},
	{ID: "google/lyria-3-clip-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "lyria-3-clip-preview"},
	{ID: "rekaai/reka-edge", Object: "model", OwnedBy: "rekaai", Type: "cline", DisplayName: "reka-edge"},
	{ID: "minimax/minimax-m2.7", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m2.7"},
	{ID: "openai/gpt-5.4-nano", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-nano"},
	{ID: "openai/gpt-5.4-nano:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-nano:batch"},
	{ID: "openai/gpt-5.4-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-mini"},
	{ID: "openai/gpt-5.4-mini:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-mini:batch"},
	{ID: "mistralai/mistral-small-2603", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-small-2603"},
	{ID: "mistralai/mistral-small-2603:batch", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-small-2603:batch"},
	{ID: "z-ai/glm-5-turbo", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5-turbo"},
	{ID: "nvidia/nemotron-3-super-120b-a12b", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3-super-120b-a12b"},
	{ID: "nvidia/nemotron-3-super-120b-a12b:free", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3-super-120b-a12b:free"},
	{ID: "bytedance-seed/seed-2.0-lite", Object: "model", OwnedBy: "bytedance-seed", Type: "cline", DisplayName: "seed-2.0-lite"},
	{ID: "qwen/qwen3.5-9b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-9b"},
	{ID: "openai/gpt-5.4-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-pro"},
	{ID: "openai/gpt-5.4-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4-pro:batch"},
	{ID: "openai/gpt-5.4", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4"},
	{ID: "openai/gpt-5.4:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.4:batch"},
	{ID: "inception/mercury-2", Object: "model", OwnedBy: "inception", Type: "cline", DisplayName: "mercury-2"},
	{ID: "google/gemini-3.1-flash-lite-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-flash-lite-preview"},
	{ID: "bytedance-seed/seed-2.0-mini", Object: "model", OwnedBy: "bytedance-seed", Type: "cline", DisplayName: "seed-2.0-mini"},
	{ID: "google/gemini-3.1-flash-image-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-flash-image-preview"},
	{ID: "qwen/qwen3.5-35b-a3b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-35b-a3b"},
	{ID: "qwen/qwen3.5-27b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-27b"},
	{ID: "qwen/qwen3.5-122b-a10b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-122b-a10b"},
	{ID: "qwen/qwen3.5-flash-02-23", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-flash-02-23"},
	{ID: "google/gemini-3.1-pro-preview-customtools", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-pro-preview-customtools"},
	{ID: "openai/gpt-5.3-codex", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.3-codex"},
	{ID: "aion-labs/aion-2.0", Object: "model", OwnedBy: "aion-labs", Type: "cline", DisplayName: "aion-2.0"},
	{ID: "google/gemini-3.1-pro-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-pro-preview"},
	{ID: "google/gemini-3.1-pro-preview:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3.1-pro-preview:batch"},
	{ID: "anthropic/claude-sonnet-4.6", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-4.6"},
	{ID: "anthropic/claude-sonnet-4.6:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-4.6:batch"},
	{ID: "qwen/qwen3.5-plus-02-15", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-plus-02-15"},
	{ID: "qwen/qwen3.5-397b-a17b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3.5-397b-a17b"},
	{ID: "minimax/minimax-m2.5", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m2.5"},
	{ID: "z-ai/glm-5", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-5"},
	{ID: "qwen/qwen3-max-thinking", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-max-thinking"},
	{ID: "anthropic/claude-opus-4.6", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.6"},
	{ID: "anthropic/claude-opus-4.6:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.6:batch"},
	{ID: "qwen/qwen3-coder-next", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-coder-next"},
	{ID: "openrouter/free", Object: "model", OwnedBy: "openrouter", Type: "cline", DisplayName: "free"},
	{ID: "stepfun/step-3.5-flash", Object: "model", OwnedBy: "stepfun", Type: "cline", DisplayName: "step-3.5-flash"},
	{ID: "moonshotai/kimi-k2.5", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k2.5"},
	{ID: "upstage/solar-pro-3", Object: "model", OwnedBy: "upstage", Type: "cline", DisplayName: "solar-pro-3"},
	{ID: "minimax/minimax-m2-her", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m2-her"},
	{ID: "writer/palmyra-x5", Object: "model", OwnedBy: "writer", Type: "cline", DisplayName: "palmyra-x5"},
	{ID: "openai/gpt-audio", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-audio"},
	{ID: "openai/gpt-audio-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-audio-mini"},
	{ID: "z-ai/glm-4.7-flash", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.7-flash"},
	{ID: "openai/gpt-5.2-codex", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.2-codex"},
	{ID: "bytedance-seed/seed-1.6-flash", Object: "model", OwnedBy: "bytedance-seed", Type: "cline", DisplayName: "seed-1.6-flash"},
	{ID: "bytedance-seed/seed-1.6", Object: "model", OwnedBy: "bytedance-seed", Type: "cline", DisplayName: "seed-1.6"},
	{ID: "minimax/minimax-m2.1", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m2.1"},
	{ID: "z-ai/glm-4.7", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.7"},
	{ID: "google/gemini-3-flash-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3-flash-preview"},
	{ID: "google/gemini-3-flash-preview:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3-flash-preview:batch"},
	{ID: "nvidia/nemotron-3-nano-30b-a3b", Object: "model", OwnedBy: "nvidia", Type: "cline", DisplayName: "nemotron-3-nano-30b-a3b"},
	{ID: "openai/gpt-5.2-chat", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.2-chat"},
	{ID: "openai/gpt-5.2-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.2-pro"},
	{ID: "openai/gpt-5.2-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.2-pro:batch"},
	{ID: "openai/gpt-5.2", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.2"},
	{ID: "openai/gpt-5.2:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.2:batch"},
	{ID: "mistralai/devstral-2512", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "devstral-2512"},
	{ID: "relace/relace-search", Object: "model", OwnedBy: "relace", Type: "cline", DisplayName: "relace-search"},
	{ID: "z-ai/glm-4.6v", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.6v"},
	{ID: "openrouter/bodybuilder", Object: "model", OwnedBy: "openrouter", Type: "cline", DisplayName: "bodybuilder"},
	{ID: "openai/gpt-5.1-codex-max", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.1-codex-max"},
	{ID: "amazon/nova-2-lite-v1", Object: "model", OwnedBy: "amazon", Type: "cline", DisplayName: "nova-2-lite-v1"},
	{ID: "mistralai/ministral-14b-2512", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "ministral-14b-2512"},
	{ID: "mistralai/ministral-8b-2512", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "ministral-8b-2512"},
	{ID: "mistralai/ministral-8b-2512:batch", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "ministral-8b-2512:batch"},
	{ID: "mistralai/ministral-3b-2512", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "ministral-3b-2512"},
	{ID: "mistralai/mistral-large-2512", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-large-2512"},
	{ID: "mistralai/mistral-large-2512:batch", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-large-2512:batch"},
	{ID: "deepseek/deepseek-v3.2", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v3.2"},
	{ID: "anthropic/claude-opus-4.5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.5"},
	{ID: "anthropic/claude-opus-4.5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.5:batch"},
	{ID: "google/gemini-3-pro-image-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-3-pro-image-preview"},
	{ID: "openai/gpt-5.1", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.1"},
	{ID: "openai/gpt-5.1:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.1:batch"},
	{ID: "openai/gpt-5.1-codex", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.1-codex"},
	{ID: "openai/gpt-5.1-codex-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5.1-codex-mini"},
	{ID: "moonshotai/kimi-k2-thinking", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k2-thinking"},
	{ID: "amazon/nova-premier-v1", Object: "model", OwnedBy: "amazon", Type: "cline", DisplayName: "nova-premier-v1"},
	{ID: "perplexity/sonar-pro-search", Object: "model", OwnedBy: "perplexity", Type: "cline", DisplayName: "sonar-pro-search"},
	{ID: "mistralai/voxtral-small-24b-2507", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "voxtral-small-24b-2507"},
	{ID: "openai/gpt-oss-safeguard-20b", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-oss-safeguard-20b"},
	{ID: "minimax/minimax-m2", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m2"},
	{ID: "qwen/qwen3-vl-32b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-32b-instruct"},
	{ID: "ibm-granite/granite-4.0-h-micro", Object: "model", OwnedBy: "ibm-granite", Type: "cline", DisplayName: "granite-4.0-h-micro"},
	{ID: "openai/gpt-5-image-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-image-mini"},
	{ID: "anthropic/claude-haiku-4.5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-haiku-4.5"},
	{ID: "anthropic/claude-haiku-4.5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-haiku-4.5:batch"},
	{ID: "qwen/qwen3-vl-8b-thinking", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-8b-thinking"},
	{ID: "qwen/qwen3-vl-8b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-8b-instruct"},
	{ID: "openai/gpt-5-image", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-image"},
	{ID: "google/gemini-2.5-flash-image", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-flash-image"},
	{ID: "qwen/qwen3-vl-30b-a3b-thinking", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-30b-a3b-thinking"},
	{ID: "qwen/qwen3-vl-30b-a3b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-30b-a3b-instruct"},
	{ID: "openai/gpt-5-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-pro"},
	{ID: "openai/gpt-5-pro:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-pro:batch"},
	{ID: "z-ai/glm-4.6", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.6"},
	{ID: "anthropic/claude-sonnet-4.5", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-4.5"},
	{ID: "anthropic/claude-sonnet-4.5:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-4.5:batch"},
	{ID: "deepseek/deepseek-v3.2-exp", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v3.2-exp"},
	{ID: "thedrummer/cydonia-24b-v4.1", Object: "model", OwnedBy: "thedrummer", Type: "cline", DisplayName: "cydonia-24b-v4.1"},
	{ID: "relace/relace-apply-3", Object: "model", OwnedBy: "relace", Type: "cline", DisplayName: "relace-apply-3"},
	{ID: "qwen/qwen3-vl-235b-a22b-thinking", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-235b-a22b-thinking"},
	{ID: "qwen/qwen3-vl-235b-a22b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-vl-235b-a22b-instruct"},
	{ID: "qwen/qwen3-max", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-max"},
	{ID: "qwen/qwen3-coder-plus", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-coder-plus"},
	{ID: "deepseek/deepseek-v3.1-terminus", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-v3.1-terminus"},
	{ID: "qwen/qwen3-coder-flash", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-coder-flash"},
	{ID: "qwen/qwen3-next-80b-a3b-thinking", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-next-80b-a3b-thinking"},
	{ID: "qwen/qwen3-next-80b-a3b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-next-80b-a3b-instruct"},
	{ID: "qwen/qwen-plus-2025-07-28", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen-plus-2025-07-28"},
	{ID: "moonshotai/kimi-k2-0905", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k2-0905"},
	{ID: "qwen/qwen3-30b-a3b-thinking-2507", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-30b-a3b-thinking-2507"},
	{ID: "nousresearch/hermes-4-405b", Object: "model", OwnedBy: "nousresearch", Type: "cline", DisplayName: "hermes-4-405b"},
	{ID: "deepseek/deepseek-chat-v3.1", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-chat-v3.1"},
	{ID: "mistralai/mistral-medium-3.1", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-medium-3.1"},
	{ID: "mistralai/mistral-medium-3.1:batch", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-medium-3.1:batch"},
	{ID: "z-ai/glm-4.5v", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.5v"},
	{ID: "openai/gpt-5", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5"},
	{ID: "openai/gpt-5:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5:batch"},
	{ID: "openai/gpt-5-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-mini"},
	{ID: "openai/gpt-5-mini:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-mini:batch"},
	{ID: "openai/gpt-5-nano", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-nano"},
	{ID: "openai/gpt-5-nano:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-5-nano:batch"},
	{ID: "openai/gpt-oss-120b", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-oss-120b"},
	{ID: "openai/gpt-oss-120b:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-oss-120b:batch"},
	{ID: "openai/gpt-oss-20b", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-oss-20b"},
	{ID: "openai/gpt-oss-20b:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-oss-20b:batch"},
	{ID: "anthropic/claude-opus-4.1", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.1"},
	{ID: "anthropic/claude-opus-4.1:batch", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-opus-4.1:batch"},
	{ID: "mistralai/codestral-2508", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "codestral-2508"},
	{ID: "mistralai/codestral-2508:batch", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "codestral-2508:batch"},
	{ID: "qwen/qwen3-coder-30b-a3b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-coder-30b-a3b-instruct"},
	{ID: "qwen/qwen3-30b-a3b-instruct-2507", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-30b-a3b-instruct-2507"},
	{ID: "z-ai/glm-4.5", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.5"},
	{ID: "z-ai/glm-4.5-air", Object: "model", OwnedBy: "z-ai", Type: "cline", DisplayName: "glm-4.5-air"},
	{ID: "qwen/qwen3-235b-a22b-thinking-2507", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-235b-a22b-thinking-2507"},
	{ID: "qwen/qwen3-coder", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-coder"},
	{ID: "bytedance/ui-tars-1.5-7b", Object: "model", OwnedBy: "bytedance", Type: "cline", DisplayName: "ui-tars-1.5-7b"},
	{ID: "google/gemini-2.5-flash-lite", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-flash-lite"},
	{ID: "google/gemini-2.5-flash-lite:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-flash-lite:batch"},
	{ID: "qwen/qwen3-235b-a22b-2507", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-235b-a22b-2507"},
	{ID: "moonshotai/kimi-k2", Object: "model", OwnedBy: "moonshotai", Type: "cline", DisplayName: "kimi-k2"},
	{ID: "cognitivecomputations/dolphin-mistral-24b-venice-edition", Object: "model", OwnedBy: "cognitivecomputations", Type: "cline", DisplayName: "dolphin-mistral-24b-venice-edition"},
	{ID: "tencent/hunyuan-a13b-instruct", Object: "model", OwnedBy: "tencent", Type: "cline", DisplayName: "hunyuan-a13b-instruct"},
	{ID: "morph/morph-v3-large", Object: "model", OwnedBy: "morph", Type: "cline", DisplayName: "morph-v3-large"},
	{ID: "morph/morph-v3-fast", Object: "model", OwnedBy: "morph", Type: "cline", DisplayName: "morph-v3-fast"},
	{ID: "baidu/ernie-4.5-vl-424b-a47b", Object: "model", OwnedBy: "baidu", Type: "cline", DisplayName: "ernie-4.5-vl-424b-a47b"},
	{ID: "mistralai/mistral-small-3.2-24b-instruct", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-small-3.2-24b-instruct"},
	{ID: "minimax/minimax-m1", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-m1"},
	{ID: "google/gemini-2.5-flash", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-flash"},
	{ID: "google/gemini-2.5-flash:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-flash:batch"},
	{ID: "google/gemini-2.5-pro", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-pro"},
	{ID: "google/gemini-2.5-pro:batch", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-pro:batch"},
	{ID: "openai/o3-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o3-pro"},
	{ID: "google/gemini-2.5-pro-preview", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemini-2.5-pro-preview"},
	{ID: "deepseek/deepseek-r1-0528", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-r1-0528"},
	{ID: "anthropic/claude-sonnet-4", Object: "model", OwnedBy: "anthropic", Type: "cline", DisplayName: "claude-sonnet-4"},
	{ID: "mistralai/mistral-medium-3", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-medium-3"},
	{ID: "meta-llama/llama-guard-4-12b", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-guard-4-12b"},
	{ID: "qwen/qwen3-30b-a3b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-30b-a3b"},
	{ID: "qwen/qwen3-8b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-8b"},
	{ID: "qwen/qwen3-14b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-14b"},
	{ID: "qwen/qwen3-32b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-32b"},
	{ID: "qwen/qwen3-235b-a22b", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen3-235b-a22b"},
	{ID: "openai/o4-mini-high", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o4-mini-high"},
	{ID: "openai/o3", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o3"},
	{ID: "openai/o3:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o3:batch"},
	{ID: "openai/o4-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o4-mini"},
	{ID: "openai/o4-mini:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o4-mini:batch"},
	{ID: "openai/gpt-4.1", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4.1"},
	{ID: "openai/gpt-4.1:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4.1:batch"},
	{ID: "openai/gpt-4.1-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4.1-mini"},
	{ID: "openai/gpt-4.1-mini:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4.1-mini:batch"},
	{ID: "openai/gpt-4.1-nano", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4.1-nano"},
	{ID: "openai/gpt-4.1-nano:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4.1-nano:batch"},
	{ID: "meta-llama/llama-4-maverick", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-4-maverick"},
	{ID: "meta-llama/llama-4-scout", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-4-scout"},
	{ID: "deepseek/deepseek-chat-v3-0324", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-chat-v3-0324"},
	{ID: "openai/o1-pro", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o1-pro"},
	{ID: "mistralai/mistral-small-3.1-24b-instruct", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-small-3.1-24b-instruct"},
	{ID: "google/gemma-3-4b-it", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-3-4b-it"},
	{ID: "google/gemma-3-12b-it", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-3-12b-it"},
	{ID: "cohere/command-a", Object: "model", OwnedBy: "cohere", Type: "cline", DisplayName: "command-a"},
	{ID: "rekaai/reka-flash-3", Object: "model", OwnedBy: "rekaai", Type: "cline", DisplayName: "reka-flash-3"},
	{ID: "google/gemma-3-27b-it", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-3-27b-it"},
	{ID: "thedrummer/skyfall-36b-v2", Object: "model", OwnedBy: "thedrummer", Type: "cline", DisplayName: "skyfall-36b-v2"},
	{ID: "perplexity/sonar-reasoning-pro", Object: "model", OwnedBy: "perplexity", Type: "cline", DisplayName: "sonar-reasoning-pro"},
	{ID: "perplexity/sonar-pro", Object: "model", OwnedBy: "perplexity", Type: "cline", DisplayName: "sonar-pro"},
	{ID: "perplexity/sonar-deep-research", Object: "model", OwnedBy: "perplexity", Type: "cline", DisplayName: "sonar-deep-research"},
	{ID: "mistralai/mistral-saba", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-saba"},
	{ID: "openai/o3-mini-high", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o3-mini-high"},
	{ID: "aion-labs/aion-rp-llama-3.1-8b", Object: "model", OwnedBy: "aion-labs", Type: "cline", DisplayName: "aion-rp-llama-3.1-8b"},
	{ID: "qwen/qwen2.5-vl-72b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen2.5-vl-72b-instruct"},
	{ID: "qwen/qwen-plus", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen-plus"},
	{ID: "openai/o3-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o3-mini"},
	{ID: "openai/o3-mini:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o3-mini:batch"},
	{ID: "mistralai/mistral-small-24b-instruct-2501", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-small-24b-instruct-2501"},
	{ID: "perplexity/sonar", Object: "model", OwnedBy: "perplexity", Type: "cline", DisplayName: "sonar"},
	{ID: "deepseek/deepseek-r1-distill-llama-70b", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-r1-distill-llama-70b"},
	{ID: "deepseek/deepseek-r1", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-r1"},
	{ID: "minimax/minimax-01", Object: "model", OwnedBy: "minimax", Type: "cline", DisplayName: "minimax-01"},
	{ID: "microsoft/phi-4", Object: "model", OwnedBy: "microsoft", Type: "cline", DisplayName: "phi-4"},
	{ID: "deepseek/deepseek-chat", Object: "model", OwnedBy: "deepseek", Type: "cline", DisplayName: "deepseek-chat"},
	{ID: "sao10k/l3.3-euryale-70b", Object: "model", OwnedBy: "sao10k", Type: "cline", DisplayName: "l3.3-euryale-70b"},
	{ID: "openai/o1", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "o1"},
	{ID: "cohere/command-r7b-12-2024", Object: "model", OwnedBy: "cohere", Type: "cline", DisplayName: "command-r7b-12-2024"},
	{ID: "meta-llama/llama-3.3-70b-instruct", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-3.3-70b-instruct"},
	{ID: "amazon/nova-lite-v1", Object: "model", OwnedBy: "amazon", Type: "cline", DisplayName: "nova-lite-v1"},
	{ID: "amazon/nova-micro-v1", Object: "model", OwnedBy: "amazon", Type: "cline", DisplayName: "nova-micro-v1"},
	{ID: "amazon/nova-pro-v1", Object: "model", OwnedBy: "amazon", Type: "cline", DisplayName: "nova-pro-v1"},
	{ID: "openai/gpt-4o-2024-11-20", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o-2024-11-20"},
	{ID: "mistralai/mistral-large-2407", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-large-2407"},
	{ID: "qwen/qwen-2.5-coder-32b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen-2.5-coder-32b-instruct"},
	{ID: "thedrummer/unslopnemo-12b", Object: "model", OwnedBy: "thedrummer", Type: "cline", DisplayName: "unslopnemo-12b"},
	{ID: "anthracite-org/magnum-v4-72b", Object: "model", OwnedBy: "anthracite-org", Type: "cline", DisplayName: "magnum-v4-72b"},
	{ID: "qwen/qwen-2.5-7b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen-2.5-7b-instruct"},
	{ID: "meta-llama/llama-3.2-1b-instruct", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-3.2-1b-instruct"},
	{ID: "meta-llama/llama-3.2-3b-instruct", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-3.2-3b-instruct"},
	{ID: "qwen/qwen-2.5-72b-instruct", Object: "model", OwnedBy: "qwen", Type: "cline", DisplayName: "qwen-2.5-72b-instruct"},
	{ID: "cohere/command-r-08-2024", Object: "model", OwnedBy: "cohere", Type: "cline", DisplayName: "command-r-08-2024"},
	{ID: "cohere/command-r-plus-08-2024", Object: "model", OwnedBy: "cohere", Type: "cline", DisplayName: "command-r-plus-08-2024"},
	{ID: "sao10k/l3.1-euryale-70b", Object: "model", OwnedBy: "sao10k", Type: "cline", DisplayName: "l3.1-euryale-70b"},
	{ID: "nousresearch/hermes-3-llama-3.1-70b", Object: "model", OwnedBy: "nousresearch", Type: "cline", DisplayName: "hermes-3-llama-3.1-70b"},
	{ID: "nousresearch/hermes-3-llama-3.1-405b", Object: "model", OwnedBy: "nousresearch", Type: "cline", DisplayName: "hermes-3-llama-3.1-405b"},
	{ID: "sao10k/l3-lunaris-8b", Object: "model", OwnedBy: "sao10k", Type: "cline", DisplayName: "l3-lunaris-8b"},
	{ID: "openai/gpt-4o-2024-08-06", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o-2024-08-06"},
	{ID: "meta-llama/llama-3.1-70b-instruct", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-3.1-70b-instruct"},
	{ID: "meta-llama/llama-3.1-8b-instruct", Object: "model", OwnedBy: "meta-llama", Type: "cline", DisplayName: "llama-3.1-8b-instruct"},
	{ID: "mistralai/mistral-nemo", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-nemo"},
	{ID: "openai/gpt-4o-mini", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o-mini"},
	{ID: "openai/gpt-4o-mini-2024-07-18", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o-mini-2024-07-18"},
	{ID: "openai/gpt-4o-mini:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o-mini:batch"},
	{ID: "google/gemma-2-27b-it", Object: "model", OwnedBy: "google", Type: "cline", DisplayName: "gemma-2-27b-it"},
	{ID: "openai/gpt-4o", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o"},
	{ID: "openai/gpt-4o-2024-05-13", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o-2024-05-13"},
	{ID: "openai/gpt-4o:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4o:batch"},
	{ID: "mistralai/mixtral-8x22b-instruct", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mixtral-8x22b-instruct"},
	{ID: "microsoft/wizardlm-2-8x22b", Object: "model", OwnedBy: "microsoft", Type: "cline", DisplayName: "wizardlm-2-8x22b"},
	{ID: "openai/gpt-4-turbo", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4-turbo"},
	{ID: "openai/gpt-4-turbo:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4-turbo:batch"},
	{ID: "mistralai/mistral-large", Object: "model", OwnedBy: "mistralai", Type: "cline", DisplayName: "mistral-large"},
	{ID: "openai/gpt-3.5-turbo-0613", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-3.5-turbo-0613"},
	{ID: "openrouter/auto", Object: "model", OwnedBy: "openrouter", Type: "cline", DisplayName: "auto"},
	{ID: "openai/gpt-3.5-turbo-instruct", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-3.5-turbo-instruct"},
	{ID: "openai/gpt-3.5-turbo-16k", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-3.5-turbo-16k"},
	{ID: "mancer/weaver", Object: "model", OwnedBy: "mancer", Type: "cline", DisplayName: "weaver"},
	{ID: "undi95/remm-slerp-l2-13b", Object: "model", OwnedBy: "undi95", Type: "cline", DisplayName: "remm-slerp-l2-13b"},
	{ID: "gryphe/mythomax-l2-13b", Object: "model", OwnedBy: "gryphe", Type: "cline", DisplayName: "mythomax-l2-13b"},
	{ID: "openai/gpt-3.5-turbo", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-3.5-turbo"},
	{ID: "openai/gpt-3.5-turbo:batch", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-3.5-turbo:batch"},
	{ID: "openai/gpt-4", Object: "model", OwnedBy: "openai", Type: "cline", DisplayName: "gpt-4"},
}

// WithClineBuiltins appends builtin Cline models that are missing from the
// loaded catalog, without overriding live-fetched entries (which carry richer
// modalities/capability metadata).
func WithClineBuiltins(models []*ModelInfo) []*ModelInfo {
	existing := make(map[string]struct{}, len(models))
	for _, model := range models {
		if model == nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(model.ID))
		if key != "" {
			existing[key] = struct{}{}
		}
	}
	for _, extra := range clineBuiltinModels {
		if extra == nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(extra.ID))
		if key == "" {
			continue
		}
		if _, ok := existing[key]; ok {
			continue
		}
		existing[key] = struct{}{}
		models = append(models, cloneModelInfo(extra))
	}
	return models
}

// qoderCNBuiltinModels holds the Qoder CN model catalog fallback. The authoritative
// catalog is the models.json "qoder-cn" section, which mirrors the server-driven,
// WASM-encrypted catalog (~/.qoder/.models/<uid>/catalog-v{5,6}, decrypted via
// qoder_auth_wasm model_cache_decrypt). These identifiers are the offline fallbacks
// used before that catalog has been loaded, and only fill IDs absent from models.json.
//
// The set matches the current CN catalog (qodercli, 2026-09): Auto plus the Qwen,
// DeepSeek, GLM, Kimi and MiniMax entries shown in the model picker.
var qoderCNBuiltinModels = []*ModelInfo{
	{ID: "auto", DisplayName: "Auto"},
	{ID: "qmodel_38max", DisplayName: "Qwen3.8-Max"},
	{ID: "qfmodel", DisplayName: "Qwen3.8-Flash"},
	{ID: "qmodel_latest", DisplayName: "Qwen3.7-Max"},
	{ID: "qmodel", DisplayName: "Qwen3.7-Plus"},
	{ID: "q37fmodel", DisplayName: "Qwen3.7-Flash"},
	{ID: "dmodel", DisplayName: "DeepSeek-V4-Pro"},
	{ID: "dfmodel", DisplayName: "DeepSeek-Flash"},
	{ID: "gmodel", DisplayName: "GLM-5.3"},
	{ID: "gfmodel", DisplayName: "GLM-5.3-Flash"},
	{ID: "gm51model", DisplayName: "GLM-5.2"},
	{ID: "kmodel_latest", DisplayName: "Kimi-K3"},
	{ID: "kmodel", DisplayName: "Kimi-K2.8-Preview"},
	{ID: "mmodel", DisplayName: "MiniMax-M2.7"},
}

// GetQoderCNModels returns the Qoder CN model definitions. The models.json
// "qoder-cn" section is authoritative; the hard-coded identifiers only fill in
// gaps as an offline fallback before the catalog has been loaded.
func GetQoderCNModels() []*ModelInfo {
	return fillModelGaps(cloneModelInfos(getModels().QoderCN), qoderCNBuiltinModels...)
}

// GetQoderAIModels returns the international Qoder AI model definitions.
// Qoder AI shares the CN catalog shape, so the same fallback list applies.
func GetQoderAIModels() []*ModelInfo {
	return fillModelGaps(cloneModelInfos(getModels().QoderAI), qoderCNBuiltinModels...)
}

// fillModelGaps returns models followed by any fallbacks whose ID is not already
// present (case-insensitive). Existing entries always win, so a richer catalog
// definition is never shadowed by a minimal hard-coded fallback.
func fillModelGaps(models []*ModelInfo, fallbacks ...*ModelInfo) []*ModelInfo {
	if len(fallbacks) == 0 {
		return models
	}

	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		if model == nil {
			continue
		}
		id := strings.ToLower(strings.TrimSpace(model.ID))
		if id != "" {
			seen[id] = struct{}{}
		}
	}

	out := models
	for _, fallback := range fallbacks {
		if fallback == nil {
			continue
		}
		id := strings.ToLower(strings.TrimSpace(fallback.ID))
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, cloneModelInfo(fallback))
	}
	return out
}

// WithCodexBuiltins injects hard-coded Codex-only model definitions that should
// not depend on remote models.json updates. Built-ins replace any matching IDs
// already present in the provided slice.
func WithCodexBuiltins(models []*ModelInfo) []*ModelInfo {
	return upsertModelInfos(models,
		codexBuiltinImage15ModelInfo(),
		codexBuiltinImageModelInfo(),
		codexBuiltinImage25FlareModelInfo(),
		codexBuiltinImage25SunburstModelInfo(),
		codexBuiltinImage25ModelInfo(),
	)
}

// WithXAIBuiltins injects hard-coded xAI image/video model definitions that should
// not depend on remote models.json updates.
func WithXAIBuiltins(models []*ModelInfo) []*ModelInfo {
	return upsertModelInfos(models, xaiBuiltinImageModelInfo(), xaiBuiltinImageQualityModelInfo(), xaiBuiltinImage20ModelInfo(), xaiBuiltinVideoModelInfo(), xaiBuiltinVideo15ModelInfo(), xaiBuiltinVideo15PreviewModelInfo())
}

func normalizeAntigravityCapabilityModelID(modelID string) string {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	if open := strings.LastIndex(modelID, "("); open >= 0 && strings.HasSuffix(modelID, ")") {
		modelID = strings.TrimSpace(modelID[:open])
	}
	return modelID
}

func codexBuiltinImage15ModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          codexBuiltinImage15ModelID,
		Object:      "model",
		Created:     1704067200, // 2024-01-01
		OwnedBy:     "openai",
		Type:        "openai",
		DisplayName: "GPT Image 1.5",
		Version:     codexBuiltinImage15ModelID,
	}
}

func codexBuiltinImageModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          codexBuiltinImageModelID,
		Object:      "model",
		Created:     1704067200, // 2024-01-01
		OwnedBy:     "openai",
		Type:        "openai",
		DisplayName: "GPT Image 2",
		Version:     codexBuiltinImageModelID,
	}
}

func codexBuiltinImage25FlareModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          codexBuiltinImage25FlareModelID,
		Object:      "model",
		Created:     1704067200, // 2024-01-01
		OwnedBy:     "openai",
		Type:        "openai",
		DisplayName: "GPT Image 2.5 Flare",
		Version:     codexBuiltinImage25FlareModelID,
	}
}

func codexBuiltinImage25SunburstModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          codexBuiltinImage25SunburstModelID,
		Object:      "model",
		Created:     1704067200, // 2024-01-01
		OwnedBy:     "openai",
		Type:        "openai",
		DisplayName: "GPT Image 2.5 Sunburst",
		Version:     codexBuiltinImage25SunburstModelID,
	}
}

func codexBuiltinImage25ModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          codexBuiltinImage25ModelID,
		Object:      "model",
		Created:     1704067200, // 2024-01-01
		OwnedBy:     "openai",
		Type:        "openai",
		DisplayName: "GPT Image 2.5",
		Version:     codexBuiltinImage25ModelID,
	}
}

func xaiBuiltinImageModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          xaiBuiltinImageModelID,
		Object:      "model",
		Created:     1735689600, // 2025-01-01
		OwnedBy:     "xai",
		Type:        "xai",
		DisplayName: "Grok Imagine Image",
		Name:        xaiBuiltinImageModelID,
		Description: "xAI Grok image generation model.",
	}
}

func xaiBuiltinImageQualityModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          xaiBuiltinImageQualityModelID,
		Object:      "model",
		Created:     1735689600, // 2025-01-01
		OwnedBy:     "xai",
		Type:        "xai",
		DisplayName: "Grok Imagine Image Quality",
		Name:        xaiBuiltinImageQualityModelID,
		Description: "xAI Grok higher-fidelity image generation model.",
	}
}

func xaiBuiltinImage20ModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          xaiBuiltinImage20ModelID,
		Object:      "model",
		Created:     1786060800, // 2026-08-07
		OwnedBy:     "xai",
		Type:        "xai",
		DisplayName: "Grok Imagine Image 2.0",
		Name:        xaiBuiltinImage20ModelID,
		Description: "xAI Grok image generation model.",
	}
}

func xaiBuiltinVideoModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          xaiBuiltinVideoModelID,
		Object:      "model",
		Created:     1735689600, // 2025-01-01
		OwnedBy:     "xai",
		Type:        "xai",
		DisplayName: "Grok Imagine Video",
		Name:        xaiBuiltinVideoModelID,
		Description: "xAI Grok video generation model.",
	}
}

func xaiBuiltinVideo15ModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          xaiBuiltinVideo15ModelID,
		Object:      "model",
		Created:     1735689600, // 2025-01-01
		OwnedBy:     "xai",
		Type:        "xai",
		DisplayName: "Grok Imagine Video 1.5",
		Name:        xaiBuiltinVideo15ModelID,
		Description: "xAI Grok video generation model.",
	}
}

func xaiBuiltinVideo15PreviewModelInfo() *ModelInfo {
	return &ModelInfo{
		ID:          xaiBuiltinVideo15PreviewID,
		Object:      "model",
		Created:     1735689600, // 2025-01-01
		OwnedBy:     "xai",
		Type:        "xai",
		DisplayName: "Grok Imagine Video 1.5 Preview",
		Name:        xaiBuiltinVideo15PreviewID,
		Description: "Compatibility alias for the xAI Grok video generation model.",
	}
}

func upsertModelInfos(models []*ModelInfo, extras ...*ModelInfo) []*ModelInfo {
	if len(extras) == 0 {
		return models
	}

	extraIDs := make(map[string]struct{}, len(extras))
	extraList := make([]*ModelInfo, 0, len(extras))
	for _, extra := range extras {
		if extra == nil {
			continue
		}
		id := strings.TrimSpace(extra.ID)
		if id == "" {
			continue
		}
		key := strings.ToLower(id)
		if _, exists := extraIDs[key]; exists {
			continue
		}
		extraIDs[key] = struct{}{}
		extraList = append(extraList, cloneModelInfo(extra))
	}

	if len(extraList) == 0 {
		return models
	}

	filtered := make([]*ModelInfo, 0, len(models)+len(extraList))
	for _, model := range models {
		if model == nil {
			continue
		}
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		if _, exists := extraIDs[strings.ToLower(id)]; exists {
			continue
		}
		filtered = append(filtered, model)
	}

	filtered = append(filtered, extraList...)
	return filtered
}

// cloneModelInfos returns a shallow copy of the slice with each element deep-cloned.
func cloneModelInfos(models []*ModelInfo) []*ModelInfo {
	if len(models) == 0 {
		return nil
	}
	out := make([]*ModelInfo, len(models))
	for i, m := range models {
		out[i] = cloneModelInfo(m)
	}
	return out
}

// GetStaticModelDefinitionsByChannel returns static model definitions for a given channel/provider.
// It returns nil when the channel is unknown.
//
// Supported channels:
//   - claude
//   - gemini
//   - gemini-interactions
//   - vertex
//   - aistudio
//   - codex
//   - kimi
//   - codebuddy-cn
//   - codebuddy-ai
//   - deepseek-web
//   - antigravity
//   - xai
//   - xiaohuanxiong
//   - codearts
//   - devin
//   - meta
func GetStaticModelDefinitionsByChannel(channel string) []*ModelInfo {
	key := strings.ToLower(strings.TrimSpace(channel))
	switch key {
	case "trae":
		return GetTraeModels()
	case "cline":
		return GetClineModels()
	case "qwen-web":
		return GetQwenWebModels()
	case "claude":
		return GetClaudeModels()
	case "gemini":
		return GetGeminiModels()
	case "gemini-interactions":
		return GetGeminiModels()
	case "vertex":
		return GetGeminiVertexModels()
	case "aistudio":
		return GetAIStudioModels()
	case "codex":
		return GetCodexProModels()
	case "kimi", "kimi-ai", "kimi.ai", "kimi.com":
		return GetKimiModels()
	case "codebuddy-cn":
		return GetCodeBuddyCNModels()
	case "codebuddy-ai":
		return GetCodeBuddyAIModels()
	case "deepseek-web":
		return GetDeepSeekWebModels()
	case constant.Xiaohuanxiong:
		return GetXiaohuanxiongModels()
	case constant.CodeArts:
		return GetCodeArtsModels()
	case constant.QoderCN:
		return GetQoderCNModels()
	case constant.QoderAI:
		return GetQoderAIModels()
	case "antigravity":
		return GetAntigravityModels()
	case "xai", "x-ai", "grok":
		return GetXAIModels()
	case "devin":
		return GetDevinModels()
	case "meta", "muse":
		return GetMetaModels()
	default:
		return nil
	}
}

// LookupStaticModelInfoByChannel searches one provider-specific static section.
// It does not fall back across providers, so callers can preserve provenance.
func LookupStaticModelInfoByChannel(modelID, channel string) *ModelInfo {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return nil
	}
	for _, model := range GetStaticModelDefinitionsByChannel(channel) {
		if model != nil && model.ID == modelID {
			return cloneModelInfo(model)
		}
	}
	return nil
}

// GetMetaModels returns the standard Meta Muse model definitions.
func GetMetaModels() []*ModelInfo {
	return cloneModelInfos(getModels().Meta)
}

// LookupStaticModelInfo searches all static model definitions for a model by ID.
// Returns nil if no matching model is found.
func LookupStaticModelInfo(modelID string) *ModelInfo {
	if modelID == "" {
		return nil
	}

	data := getModels()
	allModels := [][]*ModelInfo{
		data.Claude,
		data.Gemini,
		data.Vertex,
		data.AIStudio,
		data.CodexPro,
		data.Kimi,
		data.Antigravity,
		data.DeepSeekWeb,
		data.XAI,
		data.Devin,
		staticDevinModels,
		GetXiaohuanxiongModels(),
		GetCodeArtsModels(),
		data.Meta,
	}
	for _, models := range allModels {
		for _, m := range models {
			if m != nil && m.ID == modelID {
				return cloneModelInfo(m)
			}
		}
	}

	return nil
}
