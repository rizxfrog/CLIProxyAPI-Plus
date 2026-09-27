package cline

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessTokenWithPrefix(t *testing.T) {
	cases := map[string]string{
		"abc":          "abc",
		"workos:abc":   "abc",
		"  workos:x  ": "x",
		"sk-123":       "sk-123",
		"sk_123":       "sk_123",
		"":             "",
	}
	for in, want := range cases {
		if got := AccessTokenWithPrefix(in); got != want {
			t.Errorf("AccessTokenWithPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAuthorizationHeader(t *testing.T) {
	if got := AuthorizationHeader("tok"); got != "Bearer tok" {
		t.Fatalf("AuthorizationHeader = %q", got)
	}
	if got := AuthorizationHeader("workos:raw"); got != "Bearer raw" {
		t.Fatalf("AuthorizationHeader strips legacy prefix = %q", got)
	}
	if got := AuthorizationHeader("sk-abc"); got != "Bearer sk-abc" {
		t.Fatalf("AuthorizationHeader BYOK = %q", got)
	}
	if got := AuthorizationHeader(""); got != "" {
		t.Fatalf("AuthorizationHeader empty = %q", got)
	}
}

func TestBuildLoginURL(t *testing.T) {
	client := NewClient(nil)
	login, err := client.BuildLoginURL()
	if err != nil {
		t.Fatalf("BuildLoginURL error = %v", err)
	}
	if !strings.HasPrefix(login.URL, "https://api.cline.bot/api/v1/auth/authorize?") {
		t.Fatalf("unexpected authorize URL: %s", login.URL)
	}
	for _, want := range []string{"client_type=extension", "callback_url=", "redirect_uri=", "state=" + login.State} {
		if !strings.Contains(login.URL, want) {
			t.Errorf("URL missing %q: %s", want, login.URL)
		}
	}
}

func TestParseCallbackDecodesEmbeddedBundle(t *testing.T) {
	bundle, _ := json.Marshal(map[string]any{
		"accessToken":  "workos-access",
		"refreshToken": "refresh-1",
		"email":        "user@example.com",
		"firstName":    "Ada",
		"lastName":     "Lovelace",
		"expiresAt":    "2030-01-02T03:04:05Z",
	})
	code := base64.StdEncoding.EncodeToString(bundle)
	client := NewClient(nil)

	token, err := client.ParseCallback(code)
	if err != nil {
		t.Fatalf("ParseCallback error = %v", err)
	}
	if token.AccessToken != "workos-access" || token.RefreshToken != "refresh-1" {
		t.Fatalf("tokens = %+v", token)
	}
	if token.Email != "user@example.com" {
		t.Fatalf("email = %q", token.Email)
	}
	if token.ExpiresAt == 0 {
		t.Fatalf("expiresAt not parsed: %+v", token)
	}
	if got := token.AccountLabel(); got != "Ada Lovelace" {
		t.Fatalf("AccountLabel = %q", got)
	}
}

func TestParseCallbackFromFullURL(t *testing.T) {
	bundle, _ := json.Marshal(map[string]any{"accessToken": "workos-access", "email": "u@e.com"})
	code := base64.StdEncoding.EncodeToString(bundle)
	client := NewClient(nil)
	callback := "http://127.0.0.1:18080/callback?code=" + code + "&state=abc"

	token, err := client.ParseCallback(callback)
	if err != nil {
		t.Fatalf("ParseCallback error = %v", err)
	}
	if token.AccessToken != "workos-access" {
		t.Fatalf("token = %+v", token)
	}
}

func TestParseCallbackEmpty(t *testing.T) {
	client := NewClient(nil)
	if _, err := client.ParseCallback("   "); err == nil {
		t.Fatal("expected error for empty callback")
	}
}

func TestExchangeTokenParsesCamelCaseEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != RefreshPath {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["grantType"] != "refresh_token" || body["clientType"] != "extension" {
			t.Errorf("unexpected refresh body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"data":{"accessToken":"new-acc","refreshToken":"new-ref","expiresAt":"2031-05-06T07:08:09Z"}}`))
	}))
	defer server.Close()

	client := NewClient(nil)
	client.apiHost = server.URL
	token, err := client.ExchangeToken(t.Context(), "old-ref")
	if err != nil {
		t.Fatalf("ExchangeToken error = %v", err)
	}
	if token.AccessToken != "new-acc" || token.RefreshToken != "new-ref" {
		t.Fatalf("token = %+v", token)
	}
	if token.ExpiresAt == 0 {
		t.Fatalf("expiresAt not parsed: %+v", token)
	}
}

func TestExchangeTokenKeepsOldRefreshWhenOmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"accessToken":"acc-only"}`))
	}))
	defer server.Close()

	client := NewClient(nil)
	client.apiHost = server.URL
	token, err := client.ExchangeToken(t.Context(), "keep-me")
	if err != nil {
		t.Fatalf("ExchangeToken error = %v", err)
	}
	if token.RefreshToken != "keep-me" {
		t.Fatalf("refresh token = %q, want keep-me", token.RefreshToken)
	}
}

func TestExchangeTokenHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.apiHost = server.URL
	if _, err := client.ExchangeToken(t.Context(), "bad"); err == nil {
		t.Fatal("expected error for non-2xx refresh response")
	}
}
