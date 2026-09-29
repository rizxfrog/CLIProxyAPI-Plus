package auth

import (
	"bytes"
	"strings"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// systemPromptOverrideMarker tags injected prompt sections so retries and
// failover never stack duplicate injections on the same payload.
const systemPromptOverrideMarker = "[CLIProxyAPI-system-prompt-override]"

// requestFormatClass buckets the wire formats seen at the conductor into the
// three system-prompt containers the injector knows how to edit:
// "openai" (messages[].role=system), "claude" (top-level system), and
// "gemini" (systemInstruction). Anything else is left untouched.
func requestFormatClass(toFormat sdktranslator.Format) string {
	switch toFormat.String() {
	case "openai", "openai-response", "responses", "codex", "interactions":
		return "openai"
	case "claude":
		return "claude"
	case "gemini", "antigravity":
		return "gemini"
	default:
		return ""
	}
}

// applySystemPromptOverride applies the configured system prompt override to
// the request payload when the selected provider/model matches the override
// rules. It runs after credential selection (provider is known) and before
// protocol translation inside the executor, so the payload is still in the
// inbound protocol shape and one implementation covers every executor.
//
// Pipeline per matching request:
//  1. tool description replacements (tools[].description and format variants)
//  2. system prompt replacements (find→replace on client system text)
//  3. prompt section append (inline prompt, or prompt-file when prompt is empty)
func applySystemPromptOverride(provider string, req cliproxyexecutor.Request, opts cliproxyexecutor.Options, cfg systemPromptOverrideRules) (cliproxyexecutor.Request, cliproxyexecutor.Options) {
	if !cfg.Enabled {
		return req, opts
	}
	payload := req.Payload
	if len(payload) == 0 {
		payload = opts.OriginalRequest
	}
	if len(payload) == 0 || payload[0] != '{' {
		return req, opts
	}
	if bytes.Contains(payload, []byte(systemPromptOverrideMarker)) {
		// Already injected on a previous attempt (retry/failover): keep the
		// existing injection even if the rule text changed mid-flight.
		return req, opts
	}
	if !systemPromptOverrideMatches(cfg, provider, req.Model) {
		return req, opts
	}
	formatClass := requestFormatClass(opts.SourceFormat)
	if formatClass == "" {
		return req, opts
	}

	section := resolvePromptText(cfg)
	var injected []byte
	var ok bool
	switch formatClass {
	case "openai":
		// Chat-shaped payloads carry messages[]; Responses-shaped payloads
		// (openai-response / codex entry) carry input/instructions instead.
		if gjson.GetBytes(payload, "messages").IsArray() {
			injected, ok = injectOpenAISystemPrompt(payload, section, cfg.Replacements)
		} else {
			injected, ok = injectResponsesSystemPrompt(payload, section, cfg.Replacements)
		}
	case "claude":
		injected, ok = injectClaudeSystemPrompt(payload, section, cfg.Replacements)
	case "gemini":
		injected, ok = injectGeminiSystemPrompt(payload, section, cfg.Replacements)
	}
	if len(cfg.ToolReplacements) > 0 {
		if toolPatched, okTool := applyToolDescriptionReplacements(injected, cfg.ToolReplacements); okTool {
			injected = toolPatched
			ok = true
		}
	}
	if !ok {
		return req, opts
	}
	req.Payload = injected
	opts.OriginalRequest = injected
	return req, opts
}

// applyReplacements applies find→replace rules in order. Rules with an empty
// Find are ignored. Returns the (possibly unchanged) text.
func applyReplacements(text string, rules []internalconfig.PromptReplacementRule) string {
	if text == "" || len(rules) == 0 {
		return text
	}
	for _, rule := range rules {
		if rule.Find == "" {
			continue
		}
		text = strings.ReplaceAll(text, rule.Find, rule.Replace)
	}
	return text
}

// applyToolDescriptionReplacements rewrites tool descriptions across the
// OpenAI (tools[].description, tools[].function.description) and Gemini
// (tools[].functionDeclarations[].description) layouts.
func applyToolDescriptionReplacements(payload []byte, rules []internalconfig.PromptReplacementRule) ([]byte, bool) {
	if len(payload) == 0 || len(rules) == 0 {
		return payload, false
	}
	tools := gjson.GetBytes(payload, "tools")
	if !tools.Exists() || !tools.IsArray() {
		return payload, false
	}
	out := payload
	changed := false
	for idx, tool := range tools.Array() {
		// Layout 1: tools[i].description (OpenAI hosted tools, Gemini tools).
		if desc := tool.Get("description"); desc.Exists() && desc.Type == gjson.String {
			if next := applyReplacements(desc.String(), rules); next != desc.String() {
				if updated, errSet := sjson.SetBytes(out, "tools."+itoa(idx)+".description", next); errSet == nil {
					out = updated
					changed = true
				}
			}
		}
		// Layout 2: tools[i].function.description (OpenAI function tools).
		if fn := tool.Get("function"); fn.Exists() && fn.IsObject() {
			if desc := fn.Get("description"); desc.Exists() && desc.Type == gjson.String {
				if next := applyReplacements(desc.String(), rules); next != desc.String() {
					if updated, errSet := sjson.SetBytes(out, "tools."+itoa(idx)+".function.description", next); errSet == nil {
						out = updated
						changed = true
					}
				}
			}
		}
		// Layout 3: tools[i].functionDeclarations[].description (Gemini).
		if decls := tool.Get("functionDeclarations"); decls.Exists() && decls.IsArray() {
			for declIdx, decl := range decls.Array() {
				if desc := decl.Get("description"); desc.Exists() && desc.Type == gjson.String {
					if next := applyReplacements(desc.String(), rules); next != desc.String() {
						path := "tools." + itoa(idx) + ".functionDeclarations." + itoa(declIdx) + ".description"
						if updated, errSet := sjson.SetBytes(out, path, next); errSet == nil {
							out = updated
							changed = true
						}
					}
				}
			}
		}
	}
	return out, changed
}

// systemPromptOverrideMatches reports whether provider and model match the
// exclusion-first rule set: excluded-providers wins over providers, and an
// empty providers list means "all providers".
func systemPromptOverrideMatches(cfg systemPromptOverrideRules, provider, model string) bool {
	if !cfg.Enabled {
		return false
	}
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return false
	}
	if containsTrimmedLower(cfg.ExcludedProviders, provider) {
		return false
	}
	if len(cfg.Providers) > 0 && !containsTrimmedLower(cfg.Providers, provider) {
		return false
	}
	if len(cfg.Models) > 0 {
		matched := false
		for _, pattern := range cfg.Models {
			if matchSystemPromptOverrideModel(pattern, model) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func containsTrimmedLower(list []string, value string) bool {
	for _, item := range list {
		if strings.TrimSpace(strings.ToLower(item)) == value {
			return true
		}
	}
	return false
}

// matchSystemPromptOverrideModel performs '*' wildcard matching on model names,
// consistent with the payload-rules matcher.
func matchSystemPromptOverrideModel(pattern, model string) bool {
	pattern = strings.TrimSpace(pattern)
	model = strings.TrimSpace(model)
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	pi, si := 0, 0
	starIdx, matchIdx := -1, 0
	for si < len(model) {
		if pi < len(pattern) && pattern[pi] == model[si] {
			pi++
			si++
			continue
		}
		if pi < len(pattern) && pattern[pi] == '*' {
			starIdx = pi
			matchIdx = si
			pi++
			continue
		}
		if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			si = matchIdx
			continue
		}
		return false
	}
	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}
	return pi == len(pattern)
}

// injectOpenAISystemPrompt appends the section to the last system message when
// one exists, or inserts a leading system message otherwise. Appending to the
// trailing system message keeps the client's own system content in front of the
// injected section, preserving upstream prompt-cache prefixes.
func injectOpenAISystemPrompt(payload []byte, section string, replacements []internalconfig.PromptReplacementRule) ([]byte, bool) {
	original := payload
	messages := gjson.GetBytes(payload, "messages")
	if !messages.Exists() || !messages.IsArray() {
		return payload, false
	}
	// Apply system-prompt replacements to every system message first.
	if len(replacements) > 0 {
		for idx, msg := range messages.Array() {
			if msg.Get("role").String() != "system" {
				continue
			}
			content := msg.Get("content")
			if !content.Exists() || content.Type != gjson.String {
				continue
			}
			if next := applyReplacements(content.String(), replacements); next != content.String() {
				if updated, errSet := sjson.SetBytes(payload, "messages."+itoa(idx)+".content", next); errSet == nil {
					payload = updated
				}
			}
		}
		messages = gjson.GetBytes(payload, "messages")
	}
	if section == "" {
		// No section to append: report success only when replacements changed something.
		return payload, !bytes.Equal(payload, original)
	}
	lastSystem := -1
	for idx, msg := range messages.Array() {
		if msg.Get("role").String() == "system" {
			lastSystem = idx
		}
	}
	sectionJSON, errQuote := sjson.SetBytes([]byte(`{"role":"system","content":""}`), "content", section)
	if errQuote != nil {
		return payload, false
	}
	if lastSystem >= 0 {
		path := "messages." + itoa(lastSystem) + ".content"
		current := gjson.GetBytes(payload, path)
		if current.IsArray() {
			// Multimodal content arrays cannot be text-appended safely; insert
			// a separate system message right after the last text system one.
			updated, errSet := sjson.SetRawBytes(payload, "messages.-1", sectionJSON)
			if errSet != nil {
				return payload, false
			}
			return moveLastMessageBeforeIndex(updated, lastSystem+1), true
		}
		merged, errSet := sjson.SetBytes(payload, path, current.String()+"\n\n"+section)
		if errSet != nil {
			return payload, false
		}
		return merged, true
	}
	updated, errSet := sjson.SetRawBytes(payload, "messages.-1", sectionJSON)
	if errSet != nil {
		return payload, false
	}
	return moveLastMessageToFirst(updated), true
}

// moveLastMessageToFirst rotates the messages array so the element appended at
// the end (the injected system message) becomes messages[0].
func moveLastMessageToFirst(payload []byte) []byte {
	messages := gjson.GetBytes(payload, "messages")
	if !messages.IsArray() {
		return payload
	}
	arr := messages.Array()
	if len(arr) < 2 {
		return payload
	}
	reordered := make([]string, 0, len(arr))
	reordered = append(reordered, arr[len(arr)-1].Raw)
	for _, msg := range arr[:len(arr)-1] {
		reordered = append(reordered, msg.Raw)
	}
	updated, errSet := sjson.SetRawBytes(payload, "messages", []byte("["+strings.Join(reordered, ",")+"]"))
	if errSet != nil {
		return payload
	}
	return updated
}

// moveLastMessageBeforeIndex rotates the messages array so the element appended
// at the end lands at position targetIdx.
func moveLastMessageBeforeIndex(payload []byte, targetIdx int) []byte {
	messages := gjson.GetBytes(payload, "messages")
	if !messages.IsArray() {
		return payload
	}
	arr := messages.Array()
	if targetIdx <= 0 || targetIdx >= len(arr) {
		return payload
	}
	reordered := make([]string, 0, len(arr))
	for _, msg := range arr[:targetIdx] {
		reordered = append(reordered, msg.Raw)
	}
	reordered = append(reordered, arr[len(arr)-1].Raw)
	for _, msg := range arr[targetIdx : len(arr)-1] {
		reordered = append(reordered, msg.Raw)
	}
	updated, errSet := sjson.SetRawBytes(payload, "messages", []byte("["+strings.Join(reordered, ",")+"]"))
	if errSet != nil {
		return payload
	}
	return updated
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// injectResponsesSystemPrompt appends the section to the instructions field of
// an OpenAI Responses-format payload. Responses has no messages array; the
// system prompt lives in instructions.
func injectResponsesSystemPrompt(payload []byte, section string, replacements []internalconfig.PromptReplacementRule) ([]byte, bool) {
	original := payload
	if !gjson.GetBytes(payload, "input").Exists() && !gjson.GetBytes(payload, "instructions").Exists() {
		return payload, false
	}
	// Apply system-prompt replacements to existing instructions first.
	if len(replacements) > 0 {
		if current := gjson.GetBytes(payload, "instructions"); current.Exists() && current.Type == gjson.String {
			if next := applyReplacements(current.String(), replacements); next != current.String() {
				if updated, errSet := sjson.SetBytes(payload, "instructions", next); errSet == nil {
					payload = updated
				}
			}
		}
	}
	if section == "" {
		// No section to append: report success only when replacements changed something.
		return payload, !bytes.Equal(payload, original)
	}
	current := gjson.GetBytes(payload, "instructions")
	if !current.Exists() {
		updated, errSet := sjson.SetBytes(payload, "instructions", section)
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	updated, errSet := sjson.SetBytes(payload, "instructions", current.String()+"\n\n"+section)
	if errSet != nil {
		return payload, false
	}
	return updated, true
}

// injectClaudeSystemPrompt appends the section to the top-level system field.
// The Claude system field may be a string or an array of content blocks.
func injectClaudeSystemPrompt(payload []byte, section string, replacements []internalconfig.PromptReplacementRule) ([]byte, bool) {
	original := payload
	system := gjson.GetBytes(payload, "system")
	// Apply system-prompt replacements to the existing system text/blocks first.
	if len(replacements) > 0 && system.Exists() {
		if system.Type == gjson.String {
			if next := applyReplacements(system.String(), replacements); next != system.String() {
				if updated, errSet := sjson.SetBytes(payload, "system", next); errSet == nil {
					payload = updated
					system = gjson.GetBytes(payload, "system")
				}
			}
		} else if system.IsArray() {
			for idx, block := range system.Array() {
				text := block.Get("text")
				if !text.Exists() || text.Type != gjson.String {
					continue
				}
				if next := applyReplacements(text.String(), replacements); next != text.String() {
					if updated, errSet := sjson.SetBytes(payload, "system."+itoa(idx)+".text", next); errSet == nil {
						payload = updated
					}
				}
			}
			system = gjson.GetBytes(payload, "system")
		}
	}
	if section == "" {
		// No section to append: report success only when replacements changed something.
		return payload, !bytes.Equal(payload, original)
	}
	if !system.Exists() {
		updated, errSet := sjson.SetBytes(payload, "system", section)
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	if system.Type == gjson.String {
		updated, errSet := sjson.SetBytes(payload, "system", system.String()+"\n\n"+section)
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	if system.IsArray() {
		block := map[string]any{"type": "text", "text": section}
		updated, errSet := sjson.SetBytes(payload, "system.-1", block)
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	return payload, false
}

// injectGeminiSystemPrompt appends the section to systemInstruction.parts.
func injectGeminiSystemPrompt(payload []byte, section string, replacements []internalconfig.PromptReplacementRule) ([]byte, bool) {
	original := payload
	sys := gjson.GetBytes(payload, "systemInstruction")
	// Apply system-prompt replacements to existing parts first.
	if len(replacements) > 0 && sys.Exists() {
		parts := sys.Get("parts")
		if parts.Exists() && parts.IsArray() {
			for idx, part := range parts.Array() {
				text := part.Get("text")
				if !text.Exists() || text.Type != gjson.String {
					continue
				}
				if next := applyReplacements(text.String(), replacements); next != text.String() {
					if updated, errSet := sjson.SetBytes(payload, "systemInstruction.parts."+itoa(idx)+".text", next); errSet == nil {
						payload = updated
					}
				}
			}
			sys = gjson.GetBytes(payload, "systemInstruction")
		} else if sys.IsArray() {
			for idx, part := range sys.Array() {
				text := part.Get("text")
				if !text.Exists() || text.Type != gjson.String {
					continue
				}
				if next := applyReplacements(text.String(), replacements); next != text.String() {
					if updated, errSet := sjson.SetBytes(payload, "systemInstruction."+itoa(idx)+".text", next); errSet == nil {
						payload = updated
					}
				}
			}
			sys = gjson.GetBytes(payload, "systemInstruction")
		}
	}
	if section == "" {
		// No section to append: report success only when replacements changed something.
		return payload, !bytes.Equal(payload, original)
	}
	if !sys.Exists() {
		instruction := map[string]any{
			"parts": []map[string]any{{"text": section}},
		}
		updated, errSet := sjson.SetBytes(payload, "systemInstruction", instruction)
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	if sys.IsArray() {
		// Raw parts array: append a text part.
		part := map[string]any{"text": section}
		updated, errSet := sjson.SetBytes(payload, "systemInstruction.-1", part)
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	if sys.IsObject() {
		parts := sys.Get("parts")
		if parts.Exists() && parts.IsArray() {
			part := map[string]any{"text": section}
			updated, errSet := sjson.SetBytes(payload, "systemInstruction.parts.-1", part)
			if errSet != nil {
				return payload, false
			}
			return updated, true
		}
		updated, errSet := sjson.SetBytes(payload, "systemInstruction.parts", []map[string]any{{"text": section}})
		if errSet != nil {
			return payload, false
		}
		return updated, true
	}
	return payload, false
}
