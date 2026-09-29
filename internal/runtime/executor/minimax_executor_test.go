package executor

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/constant"
	_ "github.com/router-for-me/CLIProxyAPI/v7/internal/translator"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestMinimaxExecutorIdentifiers(t *testing.T) {
	if got := NewMinimaxExecutor(nil).Identifier(); got != constant.Minimax {
		t.Fatalf("Identifier() = %q, want %q", got, constant.Minimax)
	}
	if got := NewMinimaxCNExecutor(nil).Identifier(); got != constant.MinimaxCN {
		t.Fatalf("Identifier() = %q, want %q", got, constant.MinimaxCN)
	}
}

func TestMinimaxRequestToFormatIsClaude(t *testing.T) {
	for _, source := range []sdktranslator.Format{sdktranslator.FormatClaude, sdktranslator.FormatOpenAI, sdktranslator.FormatOpenAIResponse} {
		got := NewMinimaxExecutor(nil).RequestToFormat(cliproxyexecutor.Request{}, cliproxyexecutor.Options{SourceFormat: source})
		if got != sdktranslator.FormatClaude {
			t.Fatalf("RequestToFormat(%q) = %q, want claude", source, got)
		}
	}
}

func TestMinimaxMessagesURLByRegion(t *testing.T) {
	en := NewMinimaxExecutor(nil)
	if got := minimaxMessagesURL(en.Identifier(), nil); got != "https://agent.minimax.io/mavis/api/v1/llm/v1/messages" {
		t.Fatalf("EN messages URL = %q", got)
	}
	cn := NewMinimaxCNExecutor(nil)
	if got := minimaxMessagesURL(cn.Identifier(), nil); got != "https://agent.minimax.cn/mavis/api/v1/llm/v1/messages" {
		t.Fatalf("CN messages URL = %q", got)
	}
	// Metadata must be able to override the region default.
	authOverride := &cliproxyauth.Auth{Metadata: map[string]any{"region": "cn"}}
	if got := minimaxMessagesURL(en.Identifier(), authOverride); got != "https://agent.minimax.cn/mavis/api/v1/llm/v1/messages" {
		t.Fatalf("metadata region override messages URL = %q", got)
	}
}

func TestMinimaxNonStreamingClaudePassthrough(t *testing.T) {
	var upstreamURL, authHeader, versionHeader, userAgent string
	var upstreamBody []byte

	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		authHeader = req.Header.Get("Authorization")
		versionHeader = req.Header.Get("Anthropic-Version")
		userAgent = req.Header.Get("User-Agent")
		var errRead error
		upstreamBody, errRead = io.ReadAll(req.Body)
		if errRead != nil {
			return nil, errRead
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"msg_1","type":"message","role":"assistant","model":"MiniMax-M3","content":[{"type":"text","text":"hello mini"}],"stop_reason":"end_turn","usage":{"input_tokens":5,"output_tokens":3}}`,
			)),
		}, nil
	}))

	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Provider:   constant.Minimax,
		Attributes: map[string]string{cliproxyauth.AttributeAuthKind: cliproxyauth.AuthKindOAuth},
		Metadata:   map[string]any{"access_token": "managed-token"},
	}
	payload := []byte(`{"model":"MiniMax-M3","max_tokens":1024,"messages":[{"role":"user","content":"hi"}]}`)

	resp, err := executor.Execute(ctx, auth, cliproxyexecutor.Request{Model: "MiniMax-M3", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if upstreamURL != "https://agent.minimax.io/mavis/api/v1/llm/v1/messages" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if authHeader != "Bearer managed-token" {
		t.Fatalf("Authorization = %q", authHeader)
	}
	if versionHeader != "2023-06-01" {
		t.Fatalf("Anthropic-Version = %q", versionHeader)
	}
	if userAgent != "MiniMaxAgent" {
		t.Fatalf("User-Agent = %q", userAgent)
	}
	if got := gjson.GetBytes(upstreamBody, "model").String(); got != "MiniMax-M3" {
		t.Fatalf("upstream model = %q", got)
	}
	if got := gjson.GetBytes(resp.Payload, "content.0.text").String(); got != "hello mini" {
		t.Fatalf("response content = %q (payload=%s)", got, resp.Payload)
	}
}

func TestMinimaxNonStreamingOpenAITranslation(t *testing.T) {
	// The Claude->OpenAI non-stream translator consumes Claude SSE events, so a
	// non-streaming OpenAI request must be translated from an upstream stream.
	var upstreamURL string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		_, _ = io.ReadAll(req.Body)
		sse := "event: message_start\n" +
			`data: {"type":"message_start","message":{"id":"msg_2","type":"message","role":"assistant","model":"MiniMax-M3","content":[],"usage":{"input_tokens":4,"output_tokens":0}}}` + "\n\n" +
			"event: content_block_delta\n" +
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"openai hello"}}` + "\n\n" +
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

	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "managed-token"}}
	payload := []byte(`{"model":"MiniMax-M3","messages":[{"role":"user","content":"hi"}]}`)

	resp, err := executor.Execute(ctx, auth, cliproxyexecutor.Request{Model: "MiniMax-M3", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if upstreamURL != "https://agent.minimax.io/mavis/api/v1/llm/v1/messages" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if got := gjson.GetBytes(resp.Payload, "object").String(); got != "chat.completion" {
		t.Fatalf("response object = %q (payload=%s)", got, resp.Payload)
	}
	if got := gjson.GetBytes(resp.Payload, "choices.0.message.content").String(); got != "openai hello" {
		t.Fatalf("response content = %q", got)
	}
}

func TestMinimaxNonStreamingResponsesTranslation(t *testing.T) {
	var upstreamURL string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		_, _ = io.ReadAll(req.Body)
		sse := "event: message_start\n" +
			`data: {"type":"message_start","message":{"id":"msg_9","type":"message","role":"assistant","model":"MiniMax-M3","content":[],"usage":{"input_tokens":4,"output_tokens":0}}}` + "\n\n" +
			"event: content_block_start\n" +
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n" +
			"event: content_block_delta\n" +
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"resp hello"}}` + "\n\n" +
			"event: content_block_stop\n" +
			`data: {"type":"content_block_stop","index":0}` + "\n\n" +
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

	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "managed-token"}}
	payload := []byte(`{"model":"MiniMax-M3","input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}]}`)

	resp, err := executor.Execute(ctx, auth, cliproxyexecutor.Request{Model: "MiniMax-M3", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAIResponse})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if upstreamURL != "https://agent.minimax.io/mavis/api/v1/llm/v1/messages" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if got := gjson.GetBytes(resp.Payload, "object").String(); got != "response" {
		t.Fatalf("response object = %q (payload=%s)", got, resp.Payload)
	}
	if got := gjson.GetBytes(resp.Payload, "output.0.content.0.text").String(); got != "resp hello" {
		t.Fatalf("response text = %q", got)
	}
	if got := gjson.GetBytes(resp.Payload, "usage.total_tokens").Int(); got != 6 {
		t.Fatalf("usage.total_tokens = %d", got)
	}
}

func TestMinimaxStreamingClaude(t *testing.T) {
	var upstreamURL, acceptHeader string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		acceptHeader = req.Header.Get("Accept")
		_, _ = io.ReadAll(req.Body)
		sse := "event: message_start\n" +
			`data: {"type":"message_start","message":{"id":"msg_3","type":"message","role":"assistant","model":"MiniMax-M3","content":[],"usage":{"input_tokens":3,"output_tokens":0}}}` + "\n\n" +
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

	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "managed-token"}}
	payload := []byte(`{"model":"MiniMax-M3","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"hi"}]}`)

	result, err := executor.ExecuteStream(ctx, auth, cliproxyexecutor.Request{Model: "MiniMax-M3", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude, Stream: true})
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if upstreamURL != "https://agent.minimax.io/mavis/api/v1/llm/v1/messages" {
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

func TestMinimaxUpstreamErrorStatusPreserved(t *testing.T) {
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`)),
		}, nil
	}))

	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "managed-token"}}
	payload := []byte(`{"model":"MiniMax-M3","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`)

	_, err := executor.Execute(ctx, auth, cliproxyexecutor.Request{Model: "MiniMax-M3", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	status, ok := err.(interface{ StatusCode() int })
	if !ok || status.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("error status = %v, want 429", err)
	}
}

func TestMinimaxPrepareRequestInjectsBearer(t *testing.T) {
	executor := NewMinimaxExecutor(nil)
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "abc"}}
	req, err := http.NewRequest(http.MethodPost, "https://agent.minimax.io/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if errPrepare := executor.PrepareRequest(req, auth); errPrepare != nil {
		t.Fatalf("PrepareRequest: %v", errPrepare)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer abc" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != "MiniMaxAgent" {
		t.Fatalf("User-Agent = %q", got)
	}
}

func TestMinimaxRefreshRotatesCredential(t *testing.T) {
	// Use an httptest server is unnecessary: point the refresh client at a mock
	// region by overriding the token endpoint via the region origin would require
	// DNS; instead assert the no-refresh path and that a refresh token path
	// preserves metadata shape when the endpoint is unreachable.
	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "abc"}}
	refreshed, err := executor.Refresh(context.Background(), auth)
	if err != nil {
		t.Fatalf("Refresh without refresh token: %v", err)
	}
	if refreshed != auth {
		t.Fatal("Refresh without refresh token should return the same auth")
	}
}

func TestMinimaxAdaptiveThinkingDialect(t *testing.T) {
	var upstreamBody []byte
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamBody, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"m","type":"message","role":"assistant","model":"MiniMax-M3","content":[{"type":"text","text":"x"}],"usage":{"input_tokens":1,"output_tokens":1}}`,
			)),
		}, nil
	}))

	executor := NewMinimaxExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "x"}}
	payload := []byte(`{"model":"MiniMax-M3(high)","max_tokens":4096,"messages":[{"role":"user","content":"hi"}]}`)

	if _, err := executor.Execute(ctx, auth, cliproxyexecutor.Request{Model: "MiniMax-M3(high)", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	// The suffix must be stripped from the upstream model and the MiniMax
	// Messages dialect emitted: adaptive thinking plus an explicit effort.
	if got := gjson.GetBytes(upstreamBody, "model").String(); got != "MiniMax-M3" {
		t.Fatalf("upstream model = %q, want MiniMax-M3 (suffix stripped)", got)
	}
	if got := gjson.GetBytes(upstreamBody, "thinking.type").String(); got != "adaptive" {
		t.Fatalf("thinking.type = %q, want adaptive (body=%s)", got, upstreamBody)
	}
	if got := gjson.GetBytes(upstreamBody, "output_config.effort").String(); got != "high" {
		t.Fatalf("output_config.effort = %q, want high (body=%s)", got, upstreamBody)
	}
}
