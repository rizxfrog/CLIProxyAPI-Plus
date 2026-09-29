package executor

import (
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

func TestCodeBuddyAIExecutorIdentifier(t *testing.T) {
	executor := NewCodeBuddyAIExecutor(nil)
	if got := executor.Identifier(); got != constant.CodeBuddyAI {
		t.Fatalf("Identifier() = %q, want %q", got, constant.CodeBuddyAI)
	}
}

func TestPrepareCodeBuddyAIAuthUsesOAuthTokenAndDefaults(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: constant.CodeBuddyAI,
		Metadata: map[string]any{
			"access_token": "cbai-oauth",
		},
		Attributes: map[string]string{
			"header:X-IDE-Name": "custom-client",
		},
	}
	prepared := prepareCodeBuddyAuth(auth, codeBuddyAIAuthDefaults)
	if prepared == auth {
		t.Fatal("prepareCodeBuddyAuth returned original auth")
	}
	if got := prepared.Attributes["api_key"]; got != "cbai-oauth" {
		t.Fatalf("api_key = %q", got)
	}
	if got := prepared.Attributes["base_url"]; got != "https://www.codebuddy.ai/v2" {
		t.Fatalf("base_url = %q", got)
	}
	if got := prepared.Attributes["header:X-IDE-Name"]; got != "custom-client" {
		t.Fatalf("custom X-IDE-Name = %q", got)
	}
	if got := prepared.Attributes["header:X-Product"]; got != "SaaS" {
		t.Fatalf("X-Product = %q", got)
	}
	if _, ok := auth.Attributes["api_key"]; ok {
		t.Fatal("original auth was mutated")
	}
}

func TestApplyCodeBuddyAIOutgoingTransformsForcesStream(t *testing.T) {
	got := applyCodeBuddyAIOutgoingTransforms(nil, nil, "gpt-5.5", cliproxyexecutor.Options{}, []byte(`{"model":"gpt-5.5"}`))
	if !gjson.GetBytes(got, "stream").Bool() {
		t.Fatalf("stream should be forced true, got body=%s", got)
	}
}

func TestInjectCodeBuddyAISystemPromptPrependsWhenMissing(t *testing.T) {
	body := []byte(`{"model":"gpt-5.5","messages":[{"role":"user","content":"hi"}]}`)
	got := injectCodeBuddyAISystemPrompt(body, "gpt-5.5")

	msgs := gjson.GetBytes(got, "messages")
	if !msgs.IsArray() {
		t.Fatalf("messages = %s", msgs.Raw)
	}
	arr := msgs.Array()
	if len(arr) != 2 {
		t.Fatalf("message count = %d, want 2 (body=%s)", len(arr), got)
	}
	if role := arr[0].Get("role").String(); role != "system" {
		t.Fatalf("messages[0].role = %q, want system (body=%s)", role, got)
	}
	if arr[1].Get("role").String() != "user" || arr[1].Get("content").String() != "hi" {
		t.Fatalf("original user message was not preserved: %s", got)
	}
	content := arr[0].Get("content").String()
	if !strings.HasPrefix(content, "You are CodeBuddy Code.") {
		t.Fatalf("system prompt does not look like the CodeBuddy CLI prompt: %.80q", content)
	}
	if !strings.Contains(content, "The exact model ID is gpt-5.5.") {
		t.Fatalf("system prompt missing model id: body=%s", got)
	}
	if strings.Contains(content, "{{CODEBUDDY_AI_") {
		t.Fatalf("unrendered placeholder left in system prompt: %s", got)
	}
}

func TestInjectCodeBuddyAISystemPromptKeepsExistingSystem(t *testing.T) {
	body := []byte(`{"model":"gpt-5.5","messages":[{"role":"system","content":"be terse"},{"role":"user","content":"hi"}]}`)
	got := injectCodeBuddyAISystemPrompt(body, "gpt-5.5")
	arr := gjson.GetBytes(got, "messages").Array()
	if len(arr) != 2 {
		t.Fatalf("message count = %d, want 2 (body=%s)", len(arr), got)
	}
	if arr[0].Get("content").String() != "be terse" {
		t.Fatalf("existing system prompt was replaced: %s", got)
	}
}

func TestInjectCodeBuddyAISystemPromptIgnoresAbsentMessages(t *testing.T) {
	body := `{"model":"gpt-5.5"}`
	got := injectCodeBuddyAISystemPrompt([]byte(body), "gpt-5.5")
	if string(got) != body {
		t.Fatalf("body changed unexpectedly: in=%s out=%s", body, got)
	}
}

func TestInjectCodeBuddyAISystemPromptFillsEmptyMessages(t *testing.T) {
	got := injectCodeBuddyAISystemPrompt([]byte(`{"model":"gpt-5.5","messages":[]}`), "gpt-5.5")
	arr := gjson.GetBytes(got, "messages").Array()
	if len(arr) != 1 || arr[0].Get("role").String() != "system" {
		t.Fatalf("expected single injected system message, got body=%s", got)
	}
}

func TestApplyCodeBuddyAIOutgoingTransformsInjectsSystem(t *testing.T) {
	got := applyCodeBuddyAIOutgoingTransforms(nil, nil, "glm-5.2", cliproxyexecutor.Options{},
		[]byte(`{"model":"glm-5.2","messages":[{"role":"user","content":"hi"}]}`))
	arr := gjson.GetBytes(got, "messages").Array()
	if len(arr) != 2 || arr[0].Get("role").String() != "system" {
		t.Fatalf("expected injected system message, got body=%s", got)
	}
	if !gjson.GetBytes(got, "stream").Bool() {
		t.Fatalf("stream should be forced true, got body=%s", got)
	}
}
