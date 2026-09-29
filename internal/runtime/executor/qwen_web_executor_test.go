package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	execution "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	translator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	"github.com/tidwall/gjson"
)

type qwenWebTestTransport func(*http.Request) (*http.Response, error)

func (f qwenWebTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func qwenWebTestContext(t *testing.T, stream string) (context.Context, *[]string, *sync.Mutex) {
	t.Helper()
	calls := []string{}
	mu := &sync.Mutex{}
	rt := qwenWebTestTransport(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		calls = append(calls, r.Method+" "+r.URL.Path)
		mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer test-session" {
			t.Error("missing auth")
		}
		body := `{"success":true}`
		contentType := "application/json"
		switch r.URL.Path {
		case "/api/v2/chats/new":
			body = `{"success":true,"data":{"id":"chat-1"}}`
		case "/api/v2/chat/completions":
			raw, _ := io.ReadAll(r.Body)
			if gjson.GetBytes(raw, "model").String() != "qwen3.7-plus" || gjson.GetBytes(raw, "messages.0.feature_config.function_calling").Bool() {
				t.Errorf("wrong payload: %s", raw)
			}
			contentType = "text/event-stream"
			body = stream
		case "/api/v2/chats/chat-1":
			if r.Method == http.MethodGet {
				body = `{"success":true,"data":{"chat":{"messages":[{"role":"assistant","content":"https://cdn.qwenlm.ai/output/x/t2i/generated.png?key=z","phase":"image_gen"}]}}}`
			}
		default:
			t.Error("unexpected path " + r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	return context.WithValue(context.Background(), "cliproxy.roundtripper", rt), &calls, mu
}
func TestQwenWebExecutorChatAndStream(t *testing.T) {
	for _, streaming := range []bool{false, true} {
		t.Run(fmt.Sprint(streaming), func(t *testing.T) {
			ctx, calls, mu := qwenWebTestContext(t, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\",\"phase\":\"answer\",\"status\":\"finished\"}}]}\n\n")
			e := NewQwenWebExecutor(nil)
			auth := &coreauth.Auth{Metadata: map[string]any{"access_token": "test-session"}}
			body := []byte(`{"messages":[{"role":"user","content":"hello"}]}`)
			req := execution.Request{Model: "qwen-web-chat", Payload: body}
			opts := execution.Options{SourceFormat: translator.FormatOpenAI, ResponseFormat: translator.FormatOpenAI, OriginalRequest: body, Stream: streaming}
			if streaming {
				result, err := e.ExecuteStream(ctx, auth, req, opts)
				if err != nil {
					t.Fatal(err)
				}
				output := ""
				for chunk := range result.Chunks {
					if chunk.Err != nil {
						t.Fatal(chunk.Err)
					}
					output += string(chunk.Payload)
				}
				if !strings.Contains(output, "Hello") || !strings.Contains(output, "stop") {
					t.Fatalf("bad stream %s", output)
				}
			} else {
				result, err := e.Execute(ctx, auth, req, opts)
				if err != nil {
					t.Fatal(err)
				}
				if gjson.GetBytes(result.Payload, "choices.0.message.content").String() != "Hello" {
					t.Fatalf("bad reply %s", result.Payload)
				}
			}
			mu.Lock()
			defer mu.Unlock()
			if len(*calls) != 3 || (*calls)[2] != "DELETE /api/v2/chats/chat-1" {
				t.Fatalf("cleanup absent: %v", *calls)
			}
		})
	}
}
func TestQwenWebExecutorImagesDetailFallback(t *testing.T) {
	ctx, calls, mu := qwenWebTestContext(t, "data: [DONE]\n\n")
	e := NewQwenWebExecutor(nil)
	auth := &coreauth.Auth{Metadata: map[string]any{"access_token": "test-session"}}
	body := []byte(`{"prompt":"a cat","response_format":"url"}`)
	result, err := e.Execute(ctx, auth, execution.Request{Model: "qwen-web-image", Payload: body}, execution.Options{SourceFormat: translator.Format("openai-image")})
	if err != nil {
		t.Fatal(err)
	}
	if gjson.GetBytes(result.Payload, "data.0.url").String() != "https://cdn.qwenlm.ai/output/x/t2i/generated.png?key=z" {
		t.Fatalf("bad images: %s", result.Payload)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(*calls) != 4 {
		t.Fatalf("paths: %v", *calls)
	}
}
func TestQwenWebExecutorRejectsUnsupportedBeforeNetwork(t *testing.T) {
	for _, tc := range []struct{ body, model, format string }{
		{`{"tools":[{"type":"function"}]}`, "qwen-web-chat", "openai"},
		{`{"messages":[{"role":"user","content":[{"type":"image_url"}]}]}`, "qwen-web-chat", "openai"},
		{`{"prompt":"x","response_format":"b64_json"}`, "qwen-web-image", "openai-image"},
		{`{"prompt":"x","stream":true}`, "qwen-web-image", "openai-image"},
	} {
		ctx, calls, _ := qwenWebTestContext(t, "")
		_, err := NewQwenWebExecutor(nil).Execute(ctx, nil, execution.Request{Model: tc.model, Payload: []byte(tc.body)}, execution.Options{SourceFormat: translator.Format(tc.format)})
		if err == nil || len(*calls) != 0 {
			t.Fatalf("not rejected locally: %s err=%v calls=%v", tc.body, err, *calls)
		}
	}
}
func TestQwenWebExecutorRedactsNetworkErrorAndRejectsForeignHost(t *testing.T) {
	rt := qwenWebTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("test-session secret proxy password")
	})
	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", rt)
	auth := &coreauth.Auth{Metadata: map[string]any{"access_token": "test-session"}}
	for _, target := range []string{"https://chat.qwen.ai/api/v1/auths/", "https://example.com/"} {
		req, _ := http.NewRequest(http.MethodGet, target, nil)
		_, err := NewQwenWebExecutor(nil).HttpRequest(ctx, auth, req)
		if err == nil || strings.Contains(err.Error(), "test-session") {
			t.Fatalf("unsafe error: %v", err)
		}
	}
}

func TestQwenWebRefreshSlidesSession(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/v1/auths/" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Cookie") != "token=old" {
			t.Errorf("cookie = %q", r.Header.Get("Cookie"))
		}
		// The server re-issues the session Cookie with a later expiry.
		http.SetCookie(w, &http.Cookie{Name: "token", Value: "renewed", HttpOnly: true, Path: "/"})
		fmt.Fprint(w, `{"id":"a","role":"user","token":"renewed"}`)
	}))
	defer server.Close()
	e := NewQwenWebExecutor(nil)
	e.baseURL = server.URL
	auth := &coreauth.Auth{ID: "a", Provider: "qwen-web", Metadata: map[string]any{
		"access_token": "old", "cookie": "token=old",
	}}
	updated, err := e.Refresh(context.Background(), auth)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
	if updated.Metadata["cookie"] != "token=renewed" {
		t.Fatalf("cookie = %v", updated.Metadata["cookie"])
	}
	if updated.Metadata["access_token"] != "renewed" {
		t.Fatalf("access_token = %v", updated.Metadata["access_token"])
	}
	// The caller's auth must not be mutated; the manager persists the returned copy.
	if auth.Metadata["cookie"] != "token=old" {
		t.Fatal("original auth mutated")
	}
	if e.RefreshLead() == nil || *e.RefreshLead() != 7*24*time.Hour {
		t.Fatalf("RefreshLead = %v", e.RefreshLead())
	}
}

func TestQwenWebRefreshRejectsUnrenewedSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No token Cookie: the session was not recognized.
		fmt.Fprint(w, `{"id":"a","role":"user"}`)
	}))
	defer server.Close()
	e := NewQwenWebExecutor(nil)
	e.baseURL = server.URL
	_, err := e.Refresh(context.Background(), &coreauth.Auth{Metadata: map[string]any{"cookie": "token=old"}})
	if err == nil {
		t.Fatal("expected an error when the session is not renewed")
	}
}

func TestQwenWebRefreshDoesNotRewriteWhenUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "token", Value: "same", Path: "/"})
		fmt.Fprint(w, `{"id":"a","role":"user","token":"same"}`)
	}))
	defer server.Close()
	e := NewQwenWebExecutor(nil)
	e.baseURL = server.URL
	auth := &coreauth.Auth{Metadata: map[string]any{"access_token": "same", "cookie": "token=same"}}
	updated, err := e.Refresh(context.Background(), auth)
	if err != nil {
		t.Fatal(err)
	}
	// Values are identical, so downstream persistence is a no-op diff.
	if updated.Metadata["cookie"] != "token=same" || updated.Metadata["access_token"] != "same" {
		t.Fatalf("unexpected values: %v", updated.Metadata)
	}
}

func TestQwenWebRefreshRejectsForeignHost(t *testing.T) {
	e := NewQwenWebExecutor(nil)
	// No test override: only chat.qwen.ai is allowed.
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/api/v1/auths/", nil)
	if err := e.PrepareRequest(req, &coreauth.Auth{Metadata: map[string]any{"cookie": "token=x"}}); err == nil {
		t.Fatal("foreign host accepted")
	}
}
