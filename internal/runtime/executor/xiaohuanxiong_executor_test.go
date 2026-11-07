package executor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	xiaohuanxiongauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestXiaohuanxiongExecutorUsesOAuthMetadataForChat(t *testing.T) {
	for _, stream := range []bool{false, true} {
		name := "nonstream"
		if stream {
			name = "stream"
		}
		t.Run(name, func(t *testing.T) {
			calls := 0
			ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if got := req.URL.String(); got != xiaohuanxiongauth.LLMBaseURL+"/chat/completions" {
					t.Errorf("upstream URL = %q", got)
				}
				if got := req.Header.Get("Authorization"); got != "Bearer oauth-token" {
					t.Errorf("Authorization = %q", got)
				}
				body := `{"id":"chatcmpl-1","object":"chat.completion","model":"glm-5-3","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`
				contentType := "application/json"
				if stream {
					body = "data: " + `{"id":"chatcmpl-1","object":"chat.completion.chunk","model":"glm-5-3","choices":[{"index":0,"delta":{"content":"ok"},"finish_reason":null}]}` + "\n\ndata: [DONE]\n\n"
					contentType = "text/event-stream"
				}
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			}))
			auth := &cliproxyauth.Auth{Provider: constant.Xiaohuanxiong, Metadata: map[string]any{"access_token": "oauth-token"}}
			req := cliproxyexecutor.Request{Model: "glm-5-3", Payload: []byte(`{"model":"glm-5-3","messages":[{"role":"user","content":"hi"}]}`)}
			opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI, Stream: stream}
			exec := NewXiaohuanxiongExecutor(&config.Config{})
			if stream {
				result, err := exec.ExecuteStream(ctx, auth, req, opts)
				if err != nil {
					t.Fatalf("ExecuteStream() error = %v", err)
				}
				for chunk := range result.Chunks {
					if chunk.Err != nil {
						t.Fatalf("stream chunk error = %v", chunk.Err)
					}
				}
			} else if _, err := exec.Execute(ctx, auth, req, opts); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if calls != 1 {
				t.Fatalf("upstream requests = %d, want 1", calls)
			}
			if auth.Attributes != nil {
				t.Fatal("executor mutated source auth")
			}
		})
	}
}

// TestXiaohuanxiongExecutorIdentifier pins the provider key so the conductor's
// registration switch keeps matching.
func TestXiaohuanxiongExecutorIdentifier(t *testing.T) {
	e := NewXiaohuanxiongExecutor(nil)
	if got := e.Identifier(); got != constant.Xiaohuanxiong {
		t.Fatalf("Identifier() = %q, want %q", got, constant.Xiaohuanxiong)
	}
}

// TestPrepareXiaohuanxiongAuthDerivesDefaults verifies credentials are bridged
// from OAuth metadata into the attributes the embedded executor reads.
func TestPrepareXiaohuanxiongAuthDerivesDefaults(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: constant.Xiaohuanxiong,
		Metadata: map[string]any{"access_token": "token-from-metadata"},
	}
	prepared := prepareXiaohuanxiongAuth(auth)
	if prepared == nil {
		t.Fatal("prepared auth is nil")
	}
	if got := prepared.Attributes["api_key"]; got != "token-from-metadata" {
		t.Fatalf("api_key = %q", got)
	}
	if got := prepared.Attributes["base_url"]; got != xiaohuanxiongauth.LLMBaseURL {
		t.Fatalf("base_url = %q, want %q", got, xiaohuanxiongauth.LLMBaseURL)
	}
	// The original auth must not be mutated.
	if auth.Attributes != nil && auth.Attributes["api_key"] != "" {
		t.Fatal("prepareXiaohuanxiongAuth mutated the source auth")
	}
}

// TestPrepareXiaohuanxiongAuthPrefersExplicitAttributes keeps operator overrides.
func TestPrepareXiaohuanxiongAuthPrefersExplicitAttributes(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: constant.Xiaohuanxiong,
		Attributes: map[string]string{
			"api_key":  "explicit-key",
			"base_url": "https://self-hosted.example/api/web/llm/v2",
		},
		Metadata: map[string]any{"access_token": "metadata-key"},
	}
	prepared := prepareXiaohuanxiongAuth(auth)
	if got := prepared.Attributes["api_key"]; got != "explicit-key" {
		t.Fatalf("api_key = %q, want explicit-key", got)
	}
	if got := prepared.Attributes["base_url"]; got != "https://self-hosted.example/api/web/llm/v2" {
		t.Fatalf("base_url = %q", got)
	}
}

// TestPrepareXiaohuanxiongAuthNil guards the nil auth path.
func TestPrepareXiaohuanxiongAuthNil(t *testing.T) {
	if got := prepareXiaohuanxiongAuth(nil); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

// TestXiaohuanxiongRefreshWithoutRefreshToken is the manual-paste path: an
// access token alone cannot rotate, and that must not be an error.
func TestXiaohuanxiongRefreshWithoutRefreshToken(t *testing.T) {
	e := NewXiaohuanxiongExecutor(nil)
	auth := &cliproxyauth.Auth{
		Provider: constant.Xiaohuanxiong,
		Metadata: map[string]any{"access_token": "opaque"},
	}
	refreshed, err := e.Refresh(context.Background(), auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshed == nil || refreshed.Metadata["access_token"] != "opaque" {
		t.Fatalf("auth should pass through unchanged, got %+v", refreshed)
	}
}

// TestApplyXiaohuanxiongOutgoingTransformsRewritesDialect verifies the executor
// hook applies the model family dialect to the final upstream body.
func TestApplyXiaohuanxiongOutgoingTransformsRewritesDialect(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		body     string
		wantPath string
		want     string
	}{
		{
			name:     "glm-5-3 disabled effort becomes the low tier",
			model:    "glm-5-3",
			body:     `{"model":"glm-5-3","reasoning_effort":"none"}`,
			wantPath: "reasoning_effort",
			want:     "low",
		},
		{
			name:     "deepseek moves effort into extra_body.thinking",
			model:    "deepseek-v4-pro-0813",
			body:     `{"model":"deepseek-v4-pro-0813","reasoning_effort":"high"}`,
			wantPath: "extra_body.thinking.type",
			want:     "enabled",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := applyXiaohuanxiongOutgoingTransforms(context.Background(), nil, tc.model, cliproxyexecutor.Options{}, []byte(tc.body))
			var decoded map[string]any
			if errUnmarshal := json.Unmarshal(out, &decoded); errUnmarshal != nil {
				t.Fatalf("result is not valid JSON: %s", out)
			}
			if got := gjson.GetBytes(out, tc.wantPath).String(); got != tc.want {
				t.Fatalf("%s = %q, want %q (body=%s)", tc.wantPath, got, tc.want, out)
			}
		})
	}
}

// TestApplyXiaohuanxiongOutgoingTransformsEmptyBody guards the empty payload path.
func TestApplyXiaohuanxiongOutgoingTransformsEmptyBody(t *testing.T) {
	if out := applyXiaohuanxiongOutgoingTransforms(context.Background(), nil, "glm-5-3", cliproxyexecutor.Options{}, nil); len(out) != 0 {
		t.Fatalf("expected empty passthrough, got %q", out)
	}
}
