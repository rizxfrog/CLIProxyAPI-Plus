package executor

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	clineauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/cline"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

func TestPrepareClineAuthSendsBareToken(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: "cline",
		Metadata: map[string]any{"access_token": "raw-token"},
	}
	prepared := prepareClineAuth(auth)
	if prepared == auth {
		t.Fatal("prepareClineAuth returned the original auth")
	}
	if got := prepared.Attributes["api_key"]; got != "raw-token" {
		t.Fatalf("api_key = %q, want bare token (no workos: prefix)", got)
	}
	if got := prepared.Attributes["base_url"]; got != clineauth.BaseURL {
		t.Fatalf("base_url = %q, want %q", got, clineauth.BaseURL)
	}
	if got := prepared.Attributes["header:X-CLIENT-TYPE"]; got != "cline" {
		t.Fatalf("X-CLIENT-TYPE = %q", got)
	}
	if got := prepared.Attributes["header:HTTP-Referer"]; got != "https://cline.bot" {
		t.Fatalf("HTTP-Referer = %q", got)
	}
	// The original auth must not be mutated.
	if _, ok := auth.Attributes["api_key"]; ok {
		t.Fatal("original auth attributes were mutated")
	}
}

func TestPrepareClineAuthKeepsBYOKKey(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider:   "cline",
		Attributes: map[string]string{"api_key": "sk-byok-123"},
	}
	prepared := prepareClineAuth(auth)
	if got := prepared.Attributes["api_key"]; got != "sk-byok-123" {
		t.Fatalf("api_key = %q, want unprefixed BYOK key", got)
	}
}

func TestPrepareClineAuthRespectsCustomHeaders(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider:   "cline",
		Attributes: map[string]string{"header:X-CLIENT-TYPE": "custom"},
	}
	prepared := prepareClineAuth(auth)
	if got := prepared.Attributes["header:X-CLIENT-TYPE"]; got != "custom" {
		t.Fatalf("custom X-CLIENT-TYPE overridden: %q", got)
	}
}

func TestPrepareClineAuthNil(t *testing.T) {
	if got := prepareClineAuth(nil); got != nil {
		t.Fatalf("prepareClineAuth(nil) = %v, want nil", got)
	}
}

func TestClinePrepareRequestInjectsBareAuthorization(t *testing.T) {
	executor := NewClineExecutor(nil)
	req, err := http.NewRequest(http.MethodPost, "https://example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = executor.PrepareRequest(req, &cliproxyauth.Auth{Metadata: map[string]any{"access_token": "raw"}})
	if err != nil {
		t.Fatalf("PrepareRequest error = %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer raw" {
		t.Fatalf("Authorization = %q, want Bearer raw (no workos: prefix)", got)
	}
	if got := req.Header.Get("HTTP-Referer"); got != "https://cline.bot" {
		t.Fatalf("HTTP-Referer = %q", got)
	}
}

func TestClineExecutorIdentifier(t *testing.T) {
	if got := NewClineExecutor(nil).Identifier(); got != "cline" {
		t.Fatalf("Identifier = %q, want cline", got)
	}
}

func TestApplyClineOutgoingTransformsForcesStream(t *testing.T) {
	body := []byte(`{"model":"anthropic/claude-opus-4.8","stream":false,"messages":[]}`)
	out := applyClineOutgoingTransforms(nil, nil, "anthropic/claude-opus-4.8", cliproxyexecutor.Options{}, body)
	if !gjson.GetBytes(out, "stream").Bool() {
		t.Fatalf("stream not forced true: %s", string(out))
	}
}

func TestUnwrapClineEnvelope(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantRaw bool
	}{
		{"plain chat.completion", `{"id":"x","choices":[]}`, `{"id":"x","choices":[]}`, true},
		{"wrapped", `{"success":true,"data":{"id":"x","choices":[]}}`, `{"id":"x","choices":[]}`, true},
		{"wrapped failure flag", `{"success":false,"data":{"id":"x"}}`, `{"success":false,"data":{"id":"x"}}`, true},
		{"non-json", `not json`, `not json`, true},
	}
	for _, tc := range cases {
		got := string(unwrapClineEnvelope([]byte(tc.in)))
		if got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestClineAccessTokenHasNoPrefix(t *testing.T) {
	if got := clineauth.AccessTokenWithPrefix("plain-token"); got != "plain-token" {
		t.Errorf("AccessTokenWithPrefix bare token = %q, want %q", got, "plain-token")
	}
	if got := clineauth.AccessTokenWithPrefix("workos:tok"); got != "tok" {
		t.Errorf("AccessTokenWithPrefix legacy prefix = %q, want %q", got, "tok")
	}
}
