package openai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	runtimeexecutor "github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/api/handlers"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	"github.com/tidwall/gjson"
)

type qwenImagesTransport func(*http.Request) (*http.Response, error)

func (f qwenImagesTransport) RoundTripperFor(*coreauth.Auth) http.RoundTripper { return f }

func (f qwenImagesTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestQwenWebImagesRouteUsesAccountScheduler(t *testing.T) {
	const authID = "qwen-web-images-route-test"
	reg := registry.GetGlobalRegistry()
	reg.RegisterClient(authID, "qwen-web", registry.GetQwenWebModels())
	defer reg.UnregisterClient(authID)
	manager := coreauth.NewManager(nil, nil, nil)
	manager.RegisterExecutor(runtimeexecutor.NewQwenWebExecutor(nil))
	if _, err := manager.Register(context.Background(), &coreauth.Auth{ID: authID, Provider: "qwen-web", Status: coreauth.StatusActive, Metadata: map[string]any{"access_token": "test-session"}}); err != nil {
		t.Fatal(err)
	}
	handler := NewOpenAIAPIHandler(handlers.NewBaseAPIHandlers(&config.SDKConfig{}, manager))
	calls := 0
	rt := qwenImagesTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("Authorization") != "Bearer test-session" {
			t.Error("account token not injected")
		}
		body := `{"success":true}`
		kind := "application/json"
		switch r.URL.Path {
		case "/api/v2/chats/new":
			body = `{"success":true,"data":{"id":"chat"}}`
		case "/api/v2/chat/completions":
			kind = "text/event-stream"
			body = "data: {\"response.created\":{\"chat_id\":\"chat\",\"response_id\":\"resp\"}}\n\n" + "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"https://cdn.qwenlm.ai/output/x/t2i/y.png?key=z\",\"phase\":\"image_gen\",\"status\":\"typing\"}}]}\n\n" + "data: {\"choices\":[{\"delta\":{\"content\":\"\",\"role\":\"assistant\",\"status\":\"finished\",\"phase\":\"image_gen\"}}]}\n\n"
		case "/api/v2/chats/chat":
		default:
			t.Errorf("unexpected URL: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {kind}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	manager.SetRoundTripperProvider(rt)
	wrapped := func(c *gin.Context) {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "cliproxy.roundtripper", rt))
		handler.ImagesGenerations(c)
	}
	response := performImagesEndpointRequest(t, "/v1/images/generations", "application/json", strings.NewReader(`{"model":"qwen-web-image","prompt":"draw a cat"}`), wrapped)
	if response.Code != 200 {
		t.Fatalf("route failed: %d %s", response.Code, response.Body.String())
	}
	if gjson.GetBytes(response.Body.Bytes(), "data.0.url").String() != "https://cdn.qwenlm.ai/output/x/t2i/y.png?key=z" {
		t.Fatalf("wrong image output %s", response.Body.String())
	}
	if calls != 3 {
		t.Fatalf("expected authenticated create/generate/delete; calls=%d", calls)
	}
}
