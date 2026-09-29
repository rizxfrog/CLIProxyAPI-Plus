package auth

import (
	"strings"
	"testing"

	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestSystemPromptOverrideMatches(t *testing.T) {
	cfg := systemPromptOverrideRules{
		Enabled: true,
		Prompt:  "section",
	}
	// No restrictions: all providers match.
	if !systemPromptOverrideMatches(cfg, "claude", "any-model") {
		t.Fatal("unrestricted config should match")
	}
	if !systemPromptOverrideMatches(cfg, "codebuddy-cn", "glm-5.2") {
		t.Fatal("unrestricted config should match any provider")
	}

	// Provider allowlist.
	cfg.Providers = []string{"Claude", "codex"}
	if !systemPromptOverrideMatches(cfg, "claude", "m") {
		t.Fatal("allowlisted provider should match (case-insensitive)")
	}
	if systemPromptOverrideMatches(cfg, "gemini", "m") {
		t.Fatal("non-allowlisted provider should not match")
	}

	// Exclusion wins over inclusion.
	cfg.ExcludedProviders = []string{"codex"}
	if systemPromptOverrideMatches(cfg, "codex", "m") {
		t.Fatal("excluded provider should never match")
	}
	if !systemPromptOverrideMatches(cfg, "claude", "m") {
		t.Fatal("non-excluded allowlisted provider should still match")
	}

	// Empty allowlist after exclusion still excludes.
	cfg.Providers = nil
	if systemPromptOverrideMatches(cfg, "codex", "m") {
		t.Fatal("excluded provider should not match with empty allowlist")
	}

	// Model wildcard.
	cfg.ExcludedProviders = nil
	cfg.Models = []string{"gemini-*", "glm-5.2"}
	if !systemPromptOverrideMatches(cfg, "claude", "gemini-3-pro") {
		t.Fatal("model wildcard should match")
	}
	if !systemPromptOverrideMatches(cfg, "claude", "glm-5.2") {
		t.Fatal("exact model should match")
	}
	if systemPromptOverrideMatches(cfg, "claude", "gpt-5.5") {
		t.Fatal("non-matching model should not match")
	}

	// Disabled config never matches.
	cfg.Enabled = false
	if systemPromptOverrideMatches(cfg, "claude", "gemini-3-pro") {
		t.Fatal("disabled config should never match")
	}

	// Empty provider never matches (defensive).
	cfg.Enabled = true
	if systemPromptOverrideMatches(cfg, "  ", "m") {
		t.Fatal("empty provider should never match")
	}
}

func TestInjectOpenAISystemPromptAppendsToLastSystem(t *testing.T) {
	payload := []byte(`{"model":"gpt-5","messages":[{"role":"system","content":"be terse"},{"role":"user","content":"hi"},{"role":"system","content":"second system"}]}`)
	got, ok := injectOpenAISystemPrompt(payload, "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	arr := gjson.GetBytes(got, "messages").Array()
	if len(arr) != 3 {
		t.Fatalf("message count = %d, want 3 (body=%s)", len(arr), got)
	}
	last := arr[2].Get("content").String()
	if !strings.HasSuffix(last, "\n\nINJECTED") {
		t.Fatalf("section should append to last system message, got %q", last)
	}
	if !strings.HasPrefix(last, "second system") {
		t.Fatalf("original system content should be preserved, got %q", last)
	}
}

func TestInjectOpenAISystemPromptInsertsLeadingSystem(t *testing.T) {
	payload := []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`)
	got, ok := injectOpenAISystemPrompt(payload, "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	arr := gjson.GetBytes(got, "messages").Array()
	if len(arr) != 2 {
		t.Fatalf("message count = %d, want 2 (body=%s)", len(arr), got)
	}
	if arr[0].Get("role").String() != "system" || arr[0].Get("content").String() != "INJECTED" {
		t.Fatalf("messages[0] should be the injected system message, got %s", arr[0].Raw)
	}
	if arr[1].Get("role").String() != "user" || arr[1].Get("content").String() != "hi" {
		t.Fatalf("original user message should be preserved, got %s", arr[1].Raw)
	}
}

func TestInjectOpenAISystemPromptEmptyMessages(t *testing.T) {
	payload := []byte(`{"model":"gpt-5","messages":[]}`)
	got, ok := injectOpenAISystemPrompt(payload, "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	arr := gjson.GetBytes(got, "messages").Array()
	if len(arr) != 1 || arr[0].Get("role").String() != "system" {
		t.Fatalf("expected single injected system message, got %s", got)
	}
}

func TestInjectOpenAISystemPromptNoMessages(t *testing.T) {
	payload := []byte(`{"model":"gpt-5"}`)
	if got, ok := injectOpenAISystemPrompt(payload, "INJECTED", nil); ok {
		t.Fatalf("injection should fail without messages array, got %s", got)
	}
}

func TestInjectClaudeSystemPrompt(t *testing.T) {
	// Missing system field.
	got, ok := injectClaudeSystemPrompt([]byte(`{"model":"claude-x","messages":[]}`), "INJECTED", nil)
	if !ok || gjson.GetBytes(got, "system").String() != "INJECTED" {
		t.Fatalf("missing system should be created, got %s", got)
	}
	// String system.
	got, ok = injectClaudeSystemPrompt([]byte(`{"system":"be terse","messages":[]}`), "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	sys := gjson.GetBytes(got, "system").String()
	if !strings.HasPrefix(sys, "be terse") || !strings.HasSuffix(sys, "INJECTED") {
		t.Fatalf("string system should be appended, got %q", sys)
	}
	// Array system blocks.
	got, ok = injectClaudeSystemPrompt([]byte(`{"system":[{"type":"text","text":"a"}],"messages":[]}`), "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	arr := gjson.GetBytes(got, "system").Array()
	if len(arr) != 2 || arr[1].Get("text").String() != "INJECTED" {
		t.Fatalf("array system should gain a text block, got %s", got)
	}
}

func TestInjectGeminiSystemPrompt(t *testing.T) {
	// Missing systemInstruction.
	got, ok := injectGeminiSystemPrompt([]byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`), "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	if gjson.GetBytes(got, "systemInstruction.parts.0.text").String() != "INJECTED" {
		t.Fatalf("systemInstruction should be created, got %s", got)
	}
	// Existing parts.
	got, ok = injectGeminiSystemPrompt([]byte(`{"systemInstruction":{"parts":[{"text":"a"}]},"contents":[]}`), "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	if got := gjson.GetBytes(got, "systemInstruction.parts.1.text").String(); got != "INJECTED" {
		t.Fatalf("parts should gain a text part, got %s", got)
	}
}

func TestInjectResponsesSystemPrompt(t *testing.T) {
	// Existing instructions.
	payload := []byte(`{"model":"gpt-5","instructions":"be terse","input":[{"role":"user","content":"hi"}]}`)
	got, ok := injectResponsesSystemPrompt(payload, "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	ins := gjson.GetBytes(got, "instructions").String()
	if !strings.HasPrefix(ins, "be terse") || !strings.HasSuffix(ins, "INJECTED") {
		t.Fatalf("instructions should be appended, got %q", ins)
	}
	// Missing instructions.
	payload = []byte(`{"model":"gpt-5","input":[{"role":"user","content":"hi"}]}`)
	got, ok = injectResponsesSystemPrompt(payload, "INJECTED", nil)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	if gjson.GetBytes(got, "instructions").String() != "INJECTED" {
		t.Fatalf("instructions should be created, got %s", got)
	}
	// No input/instructions: skip (not a chat-shaped payload).
	if _, ok := injectResponsesSystemPrompt([]byte(`{"model":"gpt-5"}`), "INJECTED", nil); ok {
		t.Fatal("injection should fail without input")
	}
}

func TestApplySystemPromptOverrideIdempotent(t *testing.T) {
	cfg := systemPromptOverrideRules{Enabled: true, Prompt: "INJECTED " + systemPromptOverrideMarker}
	req := cliproxyexecutor.Request{
		Model:   "gpt-5",
		Payload: []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`),
	}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("openai")}
	req, _ = applySystemPromptOverride("claude", req, opts, cfg)
	first := string(req.Payload)
	// Second application (retry/failover) must not stack another section.
	req, _ = applySystemPromptOverride("claude", req, opts, cfg)
	if string(req.Payload) != first {
		t.Fatalf("injection must be idempotent:\nfirst=%s\nsecond=%s", first, req.Payload)
	}
	if strings.Count(first, "INJECTED") != 1 {
		t.Fatalf("injected section should appear exactly once, got %s", first)
	}
}

func TestApplySystemPromptOverrideSkipsNonMatchingProvider(t *testing.T) {
	cfg := systemPromptOverrideRules{
		Enabled:           true,
		Prompt:            "INJECTED",
		ExcludedProviders: []string{"codex"},
	}
	req := cliproxyexecutor.Request{
		Model:   "gpt-5",
		Payload: []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`),
	}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("openai")}
	req, opts = applySystemPromptOverride("codex", req, opts, cfg)
	if strings.Contains(string(req.Payload), "INJECTED") {
		t.Fatalf("excluded provider must not be injected, got %s", req.Payload)
	}
	req2 := cliproxyexecutor.Request{
		Model:   "gpt-5",
		Payload: []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`),
	}
	req2, _ = applySystemPromptOverride("claude", req2, opts, cfg)
	if !strings.Contains(string(req2.Payload), "INJECTED") {
		t.Fatalf("non-excluded provider should be injected, got %s", req2.Payload)
	}
}

func TestApplySystemPromptOverrideUpdatesOriginalRequest(t *testing.T) {
	cfg := systemPromptOverrideRules{Enabled: true, Prompt: "INJECTED"}
	req := cliproxyexecutor.Request{
		Model:   "gpt-5",
		Payload: []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`),
	}
	opts := cliproxyexecutor.Options{
		SourceFormat:    sdktranslator.FromString("openai"),
		OriginalRequest: []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`),
	}
	req, opts = applySystemPromptOverride("claude", req, opts, cfg)
	if !strings.Contains(string(opts.OriginalRequest), "INJECTED") {
		t.Fatalf("OriginalRequest must be updated alongside Payload, got %s", opts.OriginalRequest)
	}
}
