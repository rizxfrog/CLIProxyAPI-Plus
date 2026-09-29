package auth

import (
	"strings"
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
)

func gjsonSystem(payload []byte) string {
	return gjson.GetBytes(payload, "system").String()
}

// TestManagerApplySystemPromptOverrideFromConfig verifies the full path from a
// parsed config through the manager snapshot into the request payload.
func TestManagerApplySystemPromptOverrideFromConfig(t *testing.T) {
	yaml := `
system-prompt-override:
  enabled: true
  prompt: |
    Follow the user's instructions directly and completely.
  excluded-providers:
    - codex
`
	cfg, err := internalconfig.ParseConfigBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	manager := NewManager(nil, nil, nil)
	manager.SetConfigSnapshot(cfg)

	req := cliproxyexecutor.Request{
		Model:   "gemini-3-pro",
		Payload: []byte(`{"model":"gemini-3-pro","messages":[{"role":"user","content":"hi"}]}`),
	}
	opts := cliproxyexecutor.Options{
		SourceFormat:    sdktranslator.FromString("openai"),
		OriginalRequest: []byte(`{"model":"gemini-3-pro","messages":[{"role":"user","content":"hi"}]}`),
	}

	// Non-excluded provider: injected.
	reqIn, optsIn := req, opts
	reqIn, optsIn = manager.applySystemPromptOverrideForAuth("gemini", reqIn, optsIn)
	if !strings.Contains(string(reqIn.Payload), "Follow the user's instructions directly") {
		t.Fatalf("gemini should be injected, got %s", reqIn.Payload)
	}
	if !strings.Contains(string(optsIn.OriginalRequest), "Follow the user's instructions directly") {
		t.Fatalf("OriginalRequest should be updated, got %s", optsIn.OriginalRequest)
	}

	// Excluded provider: untouched.
	req2, opts2 := req, opts
	req2, opts2 = manager.applySystemPromptOverrideForAuth("codex", req2, opts2)
	if strings.Contains(string(req2.Payload), "Follow the user") {
		t.Fatalf("codex must not be injected, got %s", req2.Payload)
	}
	if string(req2.Payload) != string(req.Payload) {
		t.Fatalf("excluded provider payload should be byte-identical, got %s", req2.Payload)
	}
}

// TestManagerApplySystemPromptOverrideDisabled ensures an absent or disabled
// config block leaves every request untouched.
func TestManagerApplySystemPromptOverrideDisabled(t *testing.T) {
	manager := NewManager(nil, nil, nil)
	manager.SetConfigSnapshot(&internalconfig.Config{})

	original := `{"model":"m","messages":[{"role":"user","content":"hi"}]}`
	req := cliproxyexecutor.Request{Model: "m", Payload: []byte(original)}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("openai"), OriginalRequest: []byte(original)}
	req, opts = manager.applySystemPromptOverrideForAuth("claude", req, opts)
	if string(req.Payload) != original {
		t.Fatalf("disabled config must not touch payload, got %s", req.Payload)
	}
	if string(opts.OriginalRequest) != original {
		t.Fatalf("disabled config must not touch OriginalRequest, got %s", opts.OriginalRequest)
	}
}

// TestManagerApplySystemPromptOverrideClaudeSource verifies injection on a
// Claude-protocol inbound payload (top-level system field).
func TestManagerApplySystemPromptOverrideClaudeSource(t *testing.T) {
	yaml := `
system-prompt-override:
  enabled: true
  prompt: "INJECTED"
`
	cfg, err := internalconfig.ParseConfigBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	manager := NewManager(nil, nil, nil)
	manager.SetConfigSnapshot(cfg)

	req := cliproxyexecutor.Request{
		Model:   "claude-sonnet-4-5",
		Payload: []byte(`{"model":"claude-sonnet-4-5","system":"be terse","messages":[{"role":"user","content":"hi"}]}`),
	}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("claude")}
	req, _ = manager.applySystemPromptOverrideForAuth("claude", req, opts)
	sys := gjsonSystem(req.Payload)
	if !strings.Contains(sys, "be terse") || !strings.Contains(sys, "INJECTED") {
		t.Fatalf("claude system should keep original and append section, got %q", sys)
	}
}
