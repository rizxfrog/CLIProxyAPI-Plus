package management

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/tidwall/gjson"
)

type webLoginTestTransport func(*http.Request) (*http.Response, error)

func (f webLoginTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestWebLoginPersistsOnlySessionAndActivatesHook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	h := NewHandler(&config.Config{AuthDir: dir}, "", nil)
	var saved *coreauth.Auth
	h.SetPostAuthPersistHook(func(_ context.Context, a *coreauth.Auth) error { saved = a; return nil })
	rt := webLoginTestTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "chat.qwen.ai" || r.URL.Path != "/api/v1/auths/" || r.Header.Get("Cookie") != "qwen_token=fake-session" {
			t.Errorf("unexpected upstream request: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"id":"test-account","role":"user","email":"test@example.com","token":"fake-session"}`))}, nil
	})
	request := httptest.NewRequest("POST", "/v0/management/web-login/qwen-web", strings.NewReader(`{"method":"cookie","cookie":"qwen_token=fake-session; analytics=not-to-be-saved"}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(context.WithValue(request.Context(), "cliproxy.roundtripper", rt))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = request
	h.LoginQwenWeb(c)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "fake-session") || strings.Contains(w.Body.String(), "not-to-be-saved") {
		t.Fatal("credential leaked in response")
	}
	if saved == nil || saved.Provider != "qwen-web" || saved.Status != coreauth.StatusActive {
		t.Fatalf("account not synthesized for activation: %#v", saved)
	}
	filename := gjson.GetBytes(w.Body.Bytes(), "id").String()
	raw, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		t.Fatal(err)
	}
	if gjson.GetBytes(raw, "access_token").String() != "fake-session" || gjson.GetBytes(raw, "type").String() != "qwen-web" {
		t.Fatalf("bad auth record: %s", raw)
	}
	// The session cookie is intentionally persisted; the web frontend needs it.
	if gjson.GetBytes(raw, "cookie").String() == "" {
		t.Fatalf("session cookie not persisted: %s", raw)
	}
	for _, field := range []string{"password", "analytics"} {
		if gjson.GetBytes(raw, field).Exists() {
			t.Fatalf("unexpected persisted field %s", field)
		}
	}
	if strings.Contains(string(raw), "not-to-be-saved") {
		t.Fatal("unrelated cookies stored")
	}
	if strings.Contains(string(raw), "analytics") {
		t.Fatal("unrelated cookies stored")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("response may be cached")
	}
}
func TestWebLoginUpstream401DoesNotLogOutManagementClient(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: t.TempDir()}, "", nil)
	rt := webLoginTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 401, Body: io.NopCloser(strings.NewReader(`{"detail":"fake-session"}`)), Header: http.Header{}}, nil
	})
	req := httptest.NewRequest("POST", "/v0/management/web-login/qwen-web", strings.NewReader(`{"method":"token","token":"fake-session"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "cliproxy.roundtripper", rt))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	h.LoginQwenWeb(c)
	if w.Code != 422 || strings.Contains(w.Body.String(), "fake-session") {
		t.Fatalf("unsafe login rejection: %d %s", w.Code, w.Body.String())
	}
}
