package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	_ "github.com/router-for-me/CLIProxyAPI/v8/internal/translator"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestLobsteraiServerExecutorIdentifier(t *testing.T) {
	exec := NewLobsteraiServerExecutor(nil)
	if got := exec.Identifier(); got != constant.LobsteraiServer {
		t.Fatalf("Identifier() = %q, want %q", got, constant.LobsteraiServer)
	}
}

func TestLobsteraiServerChatURLIsTheOnlySurface(t *testing.T) {
	// /messages answers every model with 40300 and /responses is 404, so the
	// executor must target chat/completions for every credential.
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"base_url": "https://lobsterai-server.youdao.com/api/proxy/v1/"}}
	want := "https://lobsterai-server.youdao.com/api/proxy/v1/chat/completions"
	if got := lobsteraiServerChatURL(auth); got != want {
		t.Fatalf("chat URL = %q, want %q", got, want)
	}
	if got := lobsteraiServerChatURL(nil); got != lobsteraiserverauth.DefaultProxyBaseURL+"/chat/completions" {
		t.Fatalf("default chat URL = %q", got)
	}
}

func TestLobsteraiServerRequestToFormatIsAlwaysOpenAI(t *testing.T) {
	// The upstream only speaks OpenAI chat completions, so an Anthropic- or
	// Responses-shaped client is translated to OpenAI rather than passed through.
	exec := NewLobsteraiServerExecutor(nil)
	for _, model := range []string{"deepseek-v4-flash", "kimi-k3", "anything"} {
		if got := exec.RequestToFormat(cliproxyexecutor.Request{Model: model}, cliproxyexecutor.Options{}); got != sdktranslator.FormatOpenAI {
			t.Fatalf("RequestToFormat(%q) = %q, want openai", model, got)
		}
	}
}

func TestLobsteraiServerStreamErrorPayloadDetectsInBandFailure(t *testing.T) {
	// The gateway reports an unsupported/inaccessible model as an SSE event:error
	// frame inside HTTP 200; treating that as success returns an empty answer.
	status, message, ok := lobsteraiServerStreamErrorPayload([]byte(`event:error
data:{"type":"error","error":{"type":"proxy_error","message":"不支持的模型: x (allowed providers: LobsterAI)","code":40300}}`))
	if !ok {
		t.Fatal("in-band error not detected")
	}
	if status != 403 {
		t.Fatalf("status = %d, want 403", status)
	}
	if !strings.Contains(message, "不支持的模型") {
		t.Fatalf("message = %q", message)
	}

	if _, _, okEmpty := lobsteraiServerStreamErrorPayload([]byte(`{"choices":[{"delta":{"content":"hi"}}]}`)); okEmpty {
		t.Fatal("a normal chunk must not be read as an error")
	}
}

func TestLobsteraiServerAggregateStreamFolds(t *testing.T) {
	raw := []byte(`{"id":"a","object":"chat.completion.chunk","model":"deepseek-flash","choices":[{"index":0,"delta":{"role":"assistant","content":""}}]}
{"id":"a","object":"chat.completion.chunk","model":"deepseek-flash","choices":[{"index":0,"delta":{"content":"hello "}}]}
{"id":"a","object":"chat.completion.chunk","model":"deepseek-flash","choices":[{"index":0,"delta":{"content":"world"},"finish_reason":"stop"}]}`)
	out := lobsteraiServerAggregateStream(raw)
	if got := gjson.GetBytes(out, "object").String(); got != "chat.completion" {
		t.Fatalf("object = %q (payload=%s)", got, out)
	}
	if got := gjson.GetBytes(out, "choices.0.message.content").String(); got != "hello world" {
		t.Fatalf("content = %q", got)
	}
	if got := gjson.GetBytes(out, "choices.0.finish_reason").String(); got != "stop" {
		t.Fatalf("finish_reason = %q", got)
	}
}

func TestLobsteraiServerCredentialUsesAccessToken(t *testing.T) {
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"access_token": "acct-token"}}
	if got := lobsteraiServerCredential(auth); got != "acct-token" {
		t.Fatalf("credential = %q", got)
	}
	auth2 := &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "meta-token"}}
	if got := lobsteraiServerCredential(auth2); got != "meta-token" {
		t.Fatalf("credential fallback = %q", got)
	}
}

func TestLobsteraiServerExecuteOpenAIFormat(t *testing.T) {
	var upstreamURL, authHeader, capsHeader, versionHeader string
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		authHeader = req.Header.Get("Authorization")
		capsHeader = req.Header.Get("X-LobsterAI-Client-Capabilities")
		versionHeader = req.Header.Get("X-LobsterAI-Client-Version")
		_, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"chatcmpl-1","object":"chat.completion","model":"deepseek-v4-flash","choices":[{"index":0,"message":{"role":"assistant","content":"hello lobster"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":3,"total_tokens":8}}`,
			)),
		}, nil
	}))

	exec := NewLobsteraiServerExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Provider:   constant.LobsteraiServer,
		Attributes: map[string]string{"access_token": "acct-token", "base_url": lobsteraiserverauth.DefaultProxyBaseURL},
	}
	payload := []byte(`{"model":"deepseek-v4-flash","messages":[{"role":"user","content":"hi"}]}`)

	resp, errExec := exec.Execute(ctx, auth, cliproxyexecutor.Request{Model: "deepseek-v4-flash", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI})
	if errExec != nil {
		t.Fatalf("Execute() error = %v", errExec)
	}
	if upstreamURL != lobsteraiserverauth.DefaultProxyBaseURL+"/chat/completions" {
		t.Fatalf("upstreamURL = %q", upstreamURL)
	}
	if authHeader != "Bearer acct-token" {
		t.Fatalf("Authorization = %q", authHeader)
	}
	if !strings.Contains(capsHeader, "kimi-k3-agentic-v1") {
		t.Fatalf("capabilities header = %q", capsHeader)
	}
	if versionHeader != lobsteraiserverauth.ClientVersion {
		t.Fatalf("client version header = %q", versionHeader)
	}
	if got := gjson.GetBytes(resp.Payload, "choices.0.message.content").String(); got != "hello lobster" {
		t.Fatalf("response content = %q (payload=%s)", got, resp.Payload)
	}
}

func TestLobsteraiServerExecuteTranslatesAnthropicToChatSurface(t *testing.T) {
	// The gateway has no Anthropic surface (/messages -> 40300), so an
	// Anthropic-shaped client must be translated to OpenAI and posted to
	// chat/completions, then translated back.
	var upstreamURL string
	var upstreamBody []byte
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		upstreamURL = req.URL.String()
		upstreamBody, _ = io.ReadAll(req.Body)
		// Mirrors the live gateway frame shape exactly (object/choices/delta plus
		// the null "message" field and reasoning_content).
		chunk := func(delta, finish string) string {
			return "data:" + `{"id":"a","created":1,"model":"deepseek-v4-flash","object":"chat.completion.chunk","usage":null,"choices":[{"index":0,"finish_reason":` + finish + `,"delta":` + delta + `,"message":null}]}` + "\n\n"
		}
		sse := chunk(`{"role":"assistant","content":"","reasoning_content":""}`, "null") +
			chunk(`{"content":"hello","reasoning_content":""}`, "null") +
			chunk(`{}`, `"stop"`) +
			"data:[DONE]\n\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sse)),
		}, nil
	}))

	exec := NewLobsteraiServerExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Provider:   constant.LobsteraiServer,
		Attributes: map[string]string{"access_token": "acct-token", "base_url": lobsteraiserverauth.DefaultProxyBaseURL},
	}
	payload := []byte(`{"model":"deepseek-v4-flash","max_tokens":1024,"messages":[{"role":"user","content":"hi"}]}`)

	resp, errExec := exec.Execute(ctx, auth, cliproxyexecutor.Request{Model: "deepseek-v4-flash", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatClaude})
	if errExec != nil {
		t.Fatalf("Execute() error = %v", errExec)
	}
	if upstreamURL != lobsteraiserverauth.DefaultProxyBaseURL+"/chat/completions" {
		t.Fatalf("upstreamURL = %q, want the chat surface", upstreamURL)
	}
	// The upstream body must be OpenAI-shaped and always streaming.
	if got := gjson.GetBytes(upstreamBody, "model").String(); got != "deepseek-v4-flash" {
		t.Fatalf("upstream model = %q (body=%s)", got, upstreamBody)
	}
	if !gjson.GetBytes(upstreamBody, "stream").Bool() {
		t.Fatalf("upstream stream flag not set: %s", upstreamBody)
	}
	// The response is translated back to the client's Anthropic shape.
	if got := gjson.GetBytes(resp.Payload, "content.0.text").String(); got != "hello" {
		t.Fatalf("response content = %q (payload=%s)", got, resp.Payload)
	}
}

func TestLobsteraiServerExecutePreservesUpstreamStatus(t *testing.T) {
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Content-Type": []string{"application/json"}, "Retry-After": []string{"30"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"rate limited","type":"rate_limit_error"}}`)),
		}, nil
	}))

	exec := NewLobsteraiServerExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Provider:   constant.LobsteraiServer,
		Attributes: map[string]string{"access_token": "acct-token", "base_url": lobsteraiserverauth.DefaultProxyBaseURL},
	}
	payload := []byte(`{"model":"deepseek-v4-flash","messages":[{"role":"user","content":"hi"}]}`)

	_, errExec := exec.Execute(ctx, auth, cliproxyexecutor.Request{Model: "deepseek-v4-flash", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI})
	if errExec == nil {
		t.Fatal("expected an error for a 429 upstream response")
	}
	statusErr, ok := errExec.(interface{ StatusCode() int })
	if !ok {
		t.Fatalf("error %T does not expose a status code", errExec)
	}
	if got := statusErr.StatusCode(); got != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", got)
	}
}

func TestLobsteraiServerCountTokensEstimates(t *testing.T) {
	exec := NewLobsteraiServerExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{Provider: constant.LobsteraiServer, Attributes: map[string]string{"access_token": "acct-token"}}
	payload := []byte(`{"model":"deepseek-v4-flash","messages":[{"role":"user","content":"count these tokens please"}]}`)
	resp, errCount := exec.CountTokens(context.Background(), auth, cliproxyexecutor.Request{Model: "deepseek-v4-flash", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI})
	if errCount != nil {
		t.Fatalf("CountTokens() error = %v", errCount)
	}
	if len(resp.Payload) == 0 {
		t.Fatal("CountTokens() returned an empty payload")
	}
}

func TestLobsteraiServerEnterpriseHeadersFollowTheCredential(t *testing.T) {
	// The backend scopes an enterprise account to its pool from this header pair,
	// so the inference path must present it — and a personal account must send
	// neither header rather than an empty one.
	base := map[string]string{"access_token": "acct-token", "base_url": lobsteraiserverauth.DefaultProxyBaseURL}
	for name, tt := range map[string]struct {
		extra map[string]string
		mode  string
		id    string
	}{
		"personal":   {extra: nil},
		"enterprise": {extra: map[string]string{"account_mode": "enterprise", "enterprise_id": "1001"}, mode: "enterprise", id: "1001"},
		"mode only":  {extra: map[string]string{"account_mode": "enterprise"}},
		"id only":    {extra: map[string]string{"enterprise_id": "1001"}},
	} {
		attrs := map[string]string{}
		for key, value := range base {
			attrs[key] = value
		}
		for key, value := range tt.extra {
			attrs[key] = value
		}
		req := httptest.NewRequest(http.MethodPost, "http://upstream.invalid/chat/completions", nil)
		if errPrepare := (&LobsteraiServerExecutor{}).PrepareRequest(req, &cliproxyauth.Auth{Attributes: attrs}); errPrepare != nil {
			t.Fatalf("%s: PrepareRequest: %v", name, errPrepare)
		}
		if got := req.Header.Get(lobsteraiserverauth.HeaderAccountMode); got != tt.mode {
			t.Fatalf("%s: account mode header = %q, want %q", name, got, tt.mode)
		}
		if got := req.Header.Get(lobsteraiserverauth.HeaderEnterpriseID); got != tt.id {
			t.Fatalf("%s: enterprise id header = %q, want %q", name, got, tt.id)
		}
	}
}
