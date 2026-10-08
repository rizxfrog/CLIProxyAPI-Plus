package executor

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	floatboatauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/floatboat"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	_ "github.com/router-for-me/CLIProxyAPI/v8/internal/translator"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestFloatboatExecutorIdentifierAndFormat(t *testing.T) {
	exec := NewFloatboatExecutor(nil)
	if got := exec.Identifier(); got != constant.Floatboat {
		t.Fatalf("Identifier() = %q, want %q", got, constant.Floatboat)
	}
	for _, source := range []sdktranslator.Format{sdktranslator.FormatClaude, sdktranslator.FormatOpenAI, sdktranslator.FormatOpenAIResponse} {
		if got := exec.RequestToFormat(cliproxyexecutor.Request{}, cliproxyexecutor.Options{SourceFormat: source}); got != sdktranslator.FormatClaude {
			t.Fatalf("RequestToFormat(%q) = %q, want claude", source, got)
		}
	}
}

func TestFloatboatMessagesURLDefaultAndOverride(t *testing.T) {
	if got := floatboatMessagesURL(nil); got != "https://newapi.aoe.chat/v1/messages" {
		t.Fatalf("default messages URL = %q", got)
	}
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"base_url": "https://gateway.example.com/"}}
	if got := floatboatMessagesURL(auth); got != "https://gateway.example.com/v1/messages" {
		t.Fatalf("override messages URL = %q", got)
	}
}

func TestFloatboatCredentialPrefersApiKey(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Attributes: map[string]string{"api_key": "fbk_live_key", "access_token": "acct-token"},
		Metadata:   map[string]any{"api_key": "meta-key"},
	}
	if got := floatboatCredential(auth); got != "fbk_live_key" {
		t.Fatalf("credential = %q, want the minted api_key", got)
	}
	// Falls back to the access token when the minted key is absent.
	auth2 := &cliproxyauth.Auth{Attributes: map[string]string{"access_token": "acct-token"}}
	if got := floatboatCredential(auth2); got != "acct-token" {
		t.Fatalf("credential fallback = %q, want access_token", got)
	}
}

func TestFloatboatNonStreamingClaudePassthrough(t *testing.T) {
	var upstreamURL, authHeader, versionHeader string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		authHeader = req.Header.Get("Authorization")
		versionHeader = req.Header.Get("Anthropic-Version")
		_, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"hello float"}],"stop_reason":"end_turn","usage":{"input_tokens":5,"output_tokens":3}}`,
			)),
		}, nil
	}))

	exec := NewFloatboatExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Provider:   constant.Floatboat,
		Attributes: map[string]string{"api_key": "fbk_live_key"},
	}
	payload := []byte(`{"model":"claude-sonnet-4-6","max_tokens":1024,"messages":[{"role":"user","content":"hi"}]}`)

	resp, err := exec.Execute(ctx, auth, cliproxyexecutor.Request{Model: "claude-sonnet-4-6", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if upstreamURL != "https://newapi.aoe.chat/v1/messages" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if authHeader != "Bearer fbk_live_key" {
		t.Fatalf("Authorization = %q", authHeader)
	}
	if versionHeader != "2023-06-01" {
		t.Fatalf("Anthropic-Version = %q", versionHeader)
	}
	if got := gjson.GetBytes(resp.Payload, "content.0.text").String(); got != "hello float" {
		t.Fatalf("response content = %q (payload=%s)", got, resp.Payload)
	}
}

func TestFloatboatNonStreamingOpenAITranslation(t *testing.T) {
	var upstreamURL string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		_, _ = io.ReadAll(req.Body)
		sse := "event: message_start\n" +
			`data: {"type":"message_start","message":{"id":"msg_2","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[],"usage":{"input_tokens":4,"output_tokens":0}}}` + "\n\n" +
			"event: content_block_delta\n" +
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"openai float"}}` + "\n\n" +
			"event: message_delta\n" +
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":4,"output_tokens":2}}` + "\n\n" +
			"event: message_stop\n" +
			`data: {"type":"message_stop"}` + "\n\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sse)),
		}, nil
	}))

	exec := NewFloatboatExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"api_key": "fbk_live_key"}}
	payload := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`)

	resp, err := exec.Execute(ctx, auth, cliproxyexecutor.Request{Model: "claude-sonnet-4-6", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if upstreamURL != "https://newapi.aoe.chat/v1/messages" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if got := gjson.GetBytes(resp.Payload, "choices.0.message.content").String(); got != "openai float" {
		t.Fatalf("response content = %q (payload=%s)", got, resp.Payload)
	}
}

func TestFloatboatStreamingClaude(t *testing.T) {
	var upstreamURL, acceptHeader string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		acceptHeader = req.Header.Get("Accept")
		_, _ = io.ReadAll(req.Body)
		sse := "event: message_start\n" +
			`data: {"type":"message_start","message":{"id":"msg_3","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[],"usage":{"input_tokens":3,"output_tokens":0}}}` + "\n\n" +
			"event: content_block_delta\n" +
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hi"}}` + "\n\n" +
			"event: message_stop\n" +
			`data: {"type":"message_stop"}` + "\n\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sse)),
		}, nil
	}))

	exec := NewFloatboatExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"api_key": "fbk_live_key"}}
	payload := []byte(`{"model":"claude-sonnet-4-6","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"hi"}]}`)

	result, err := exec.ExecuteStream(ctx, auth, cliproxyexecutor.Request{Model: "claude-sonnet-4-6", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude, Stream: true})
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if upstreamURL != "https://newapi.aoe.chat/v1/messages" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if acceptHeader != "text/event-stream" {
		t.Fatalf("Accept = %q", acceptHeader)
	}
	var combined strings.Builder
	for chunk := range result.Chunks {
		if chunk.Err != nil {
			t.Fatalf("stream chunk error = %v", chunk.Err)
		}
		combined.Write(chunk.Payload)
	}
	if !strings.Contains(combined.String(), "content_block_delta") {
		t.Fatalf("stream output missing content_block_delta: %s", combined.String())
	}
}

func TestFloatboatUpstreamErrorStatusPreserved(t *testing.T) {
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`)),
		}, nil
	}))

	exec := NewFloatboatExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"api_key": "fbk_live_key"}}
	payload := []byte(`{"model":"claude-sonnet-4-6","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`)

	_, err := exec.Execute(ctx, auth, cliproxyexecutor.Request{Model: "claude-sonnet-4-6", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	status, ok := err.(interface{ StatusCode() int })
	if !ok || status.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("error status = %v, want 429", err)
	}
}

func TestFloatboatCountTokensEstimates(t *testing.T) {
	exec := NewFloatboatExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"api_key": "fbk_live_key"}}
	payload := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hello world"}]}`)

	resp, err := exec.CountTokens(context.Background(), auth, cliproxyexecutor.Request{Model: "claude-sonnet-4-6", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude})
	if err != nil {
		t.Fatalf("CountTokens() error = %v", err)
	}
	if got := gjson.GetBytes(resp.Payload, "input_tokens").Int(); got <= 0 {
		t.Fatalf("token count = %d, want > 0 (payload=%s)", got, resp.Payload)
	}
}

func TestFloatboatRefreshWithoutRefreshTokenIsNoop(t *testing.T) {
	exec := NewFloatboatExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "acct"}}
	refreshed, err := exec.Refresh(context.Background(), auth)
	if err != nil {
		t.Fatalf("Refresh without refresh token: %v", err)
	}
	if refreshed != auth {
		t.Fatal("Refresh without refresh token should return the same auth")
	}
}

func TestFloatboatPrepareRequestInjectsBearer(t *testing.T) {
	req, errReq := http.NewRequest(http.MethodPost, "https://newapi.aoe.chat/v1/messages", nil)
	if errReq != nil {
		t.Fatalf("NewRequest: %v", errReq)
	}
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"api_key": "fbk_live_key"}}
	exec := NewFloatboatExecutor(nil)
	if err := exec.PrepareRequest(req, auth); err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer fbk_live_key" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != floatboatauth.UserAgent {
		t.Fatalf("User-Agent = %q", got)
	}
}
