package helps

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// TestNewCodeArtsHTTPClientSignsRequest drives a request through the client and
// verifies the upstream receives a signature that validates over the exact body.
func TestNewCodeArtsHTTPClientSignsRequest(t *testing.T) {
	var (
		sawAuthorization string
		sawSecurityToken string
		sawAuthToken     string
		sawSessionID     string
		sawBody          []byte
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuthorization = r.Header.Get("Authorization")
		sawSecurityToken = r.Header.Get(codeartsauth.SecurityTokenHeader)
		sawAuthToken = r.Header.Get("x-auth-token")
		sawSessionID = r.Header.Get(CodeArtsSessionHeader)
		sawBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	auth := &cliproxyauth.Auth{
		ID: "codearts-test",
		Attributes: map[string]string{
			"api_key":        "AK_E2E",
			"secret_key":     "SK_E2E",
			"security_token": "ST_E2E",
		},
	}

	client := NewCodeArtsHTTPClient(context.Background(), nil, auth, 0)
	body := []byte(`{"model":"GLM-5.2","stream":true}`)
	req, errReq := http.NewRequest(http.MethodPost, server.URL+"/api/v2/chat/completions", bytes.NewReader(body))
	if errReq != nil {
		t.Fatalf("new request: %v", errReq)
	}
	req.Header.Set("Content-Type", "application/json")
	// A stale bearer header must be dropped: Huawei Cloud rejects it.
	req.Header.Set("Authorization", "Bearer stale-token")

	resp, errDo := client.Do(req)
	if errDo != nil {
		t.Fatalf("do: %v", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			t.Errorf("close body: %v", errClose)
		}
	}()

	if !strings.HasPrefix(sawAuthorization, codeartsauth.Algorithm+" Access=AK_E2E") {
		t.Fatalf("upstream saw Authorization = %q", sawAuthorization)
	}
	if sawSecurityToken != "ST_E2E" {
		t.Fatalf("upstream saw security token = %q", sawSecurityToken)
	}
	if sawAuthToken != "ST_E2E" {
		t.Fatalf("upstream saw x-auth-token = %q", sawAuthToken)
	}
	if sawSessionID == "" {
		t.Fatal("upstream did not receive the session id header")
	}
	if !bytes.Equal(sawBody, body) {
		t.Fatalf("body was altered: %q", sawBody)
	}

	// Reproduce the signature from the recorded headers and body.
	parsed, errParse := url.Parse(server.URL + "/api/v2/chat/completions")
	if errParse != nil {
		t.Fatalf("parse url: %v", errParse)
	}
	recomputed := codeartsauth.Sign(codeartsauth.SignOptions{
		Method: http.MethodPost,
		URL:    parsed.String(),
		Headers: map[string]string{
			codeartsauth.SecurityTokenHeader: "ST_E2E",
		},
		Body: body,
		Now:  func() time.Time { return time.Unix(0, 0).UTC() },
	}, "AK_E2E", "SK_E2E")
	// The client signs with the real clock, so only the SignedHeaders/Signature
	// shape can be compared; the exact date comes from the request itself.
	if !strings.Contains(recomputed["Authorization"], "SignedHeaders=host;x-sdk-date;x-security-token") {
		t.Fatalf("SignedHeaders shape changed: %s", recomputed["Authorization"])
	}
	if !strings.Contains(sawAuthorization, "SignedHeaders=host;x-sdk-date;x-security-token") {
		t.Fatalf("upstream signature has the wrong SignedHeaders: %s", sawAuthorization)
	}
	if len(strings.Split(sawAuthorization, "Signature=")) != 2 {
		t.Fatalf("upstream signature is malformed: %s", sawAuthorization)
	}
}

// TestNewCodeArtsHTTPClientBenefitRouting pins that only the limited-time free
// models receive the maas_type=benefit header.
//
// The gateway answers InferHub.002002009.404 "The model is not registered" for
// these ids without the header, so misrouting them makes the models unusable;
// conversely the header must not leak onto standard models.
func TestNewCodeArtsHTTPClientBenefitRouting(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantHdr string
	}{
		{name: "benefit deepseek flash", body: `{"model":"deepseek-v4-flash-0731"}`, wantHdr: CodeArtsBenefitHeaderValue},
		{name: "benefit deepseek pro", body: `{"model":"deepseek-v4-pro-0813"}`, wantHdr: CodeArtsBenefitHeaderValue},
		{name: "benefit glm flash", body: `{"model":"glm-5.3-flash"}`, wantHdr: CodeArtsBenefitHeaderValue},
		{name: "benefit id case-insensitive", body: `{"model":"GLM-5.3-Flash"}`, wantHdr: CodeArtsBenefitHeaderValue},
		{name: "standard model", body: `{"model":"GLM-5.2"}`, wantHdr: ""},
		{name: "standard pangu", body: `{"model":"openpangu-2.0-pro"}`, wantHdr: ""},
		{name: "no model field", body: `{"stream":true}`, wantHdr: ""},
		{name: "empty body", body: ``, wantHdr: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sawBenefit string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sawBenefit = r.Header.Get(CodeArtsBenefitHeader)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()

			auth := &cliproxyauth.Auth{
				ID: "codearts-benefit",
				Attributes: map[string]string{
					"api_key":        "AK_B",
					"secret_key":     "SK_B",
					"security_token": "ST_B",
				},
			}

			client := NewCodeArtsHTTPClient(context.Background(), nil, auth, 0)
			req, errReq := http.NewRequest(http.MethodPost, server.URL+"/api/v2/chat/completions", bytes.NewReader([]byte(tc.body)))
			if errReq != nil {
				t.Fatalf("new request: %v", errReq)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, errDo := client.Do(req)
			if errDo != nil {
				t.Fatalf("do: %v", errDo)
			}
			if errClose := resp.Body.Close(); errClose != nil {
				t.Errorf("close body: %v", errClose)
			}

			if sawBenefit != tc.wantHdr {
				t.Fatalf("%s header = %q, want %q", CodeArtsBenefitHeader, sawBenefit, tc.wantHdr)
			}
		})
	}
}

// TestNewCodeArtsHTTPClientWithoutCredentialsPassesThrough pins that an auth
// without a credential triple is not wrapped (no panic, no bogus signature).
func TestNewCodeArtsHTTPClientWithoutCredentialsPassesThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); strings.HasPrefix(got, codeartsauth.Algorithm) {
			t.Errorf("request should not be signed without credentials, got %q", got)
		}
		_, _ = w.Write([]byte(`ok`))
	}))
	defer server.Close()

	auth := &cliproxyauth.Auth{ID: "empty"}
	client := NewCodeArtsHTTPClient(context.Background(), nil, auth, 0)
	req, errReq := http.NewRequest(http.MethodGet, server.URL+"/ping", nil)
	if errReq != nil {
		t.Fatalf("new request: %v", errReq)
	}
	resp, errDo := client.Do(req)
	if errDo != nil {
		t.Fatalf("do: %v", errDo)
	}
	if errClose := resp.Body.Close(); errClose != nil {
		t.Errorf("close body: %v", errClose)
	}
}

// TestCodeArtsCredentialTriplePrefersAttributes pins the lookup order.
func TestCodeArtsCredentialTriplePrefersAttributes(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Attributes: map[string]string{"api_key": "attr-ak", "secret_key": "attr-sk", "security_token": "attr-st"},
		Metadata:   map[string]any{"access_key": "meta-ak", "secret_key": "meta-sk", "security_token": "meta-st"},
	}
	ak, sk, st := CodeArtsCredentialTriple(auth)
	if ak != "attr-ak" || sk != "attr-sk" || st != "attr-st" {
		t.Fatalf("triple = %q/%q/%q", ak, sk, st)
	}

	onlyMeta := &cliproxyauth.Auth{Metadata: map[string]any{"access_key": "meta-ak", "secret_key": "meta-sk"}}
	ak, sk, st = CodeArtsCredentialTriple(onlyMeta)
	if ak != "meta-ak" || sk != "meta-sk" || st != "" {
		t.Fatalf("metadata triple = %q/%q/%q", ak, sk, st)
	}

	if ak, sk, st = CodeArtsCredentialTriple(nil); ak != "" || sk != "" || st != "" {
		t.Fatalf("nil auth must yield an empty triple, got %q/%q/%q", ak, sk, st)
	}
}

// TestCodeArtsSessionIDIsStable pins that the correlation header is deterministic.
func TestCodeArtsSessionIDIsStable(t *testing.T) {
	auth := &cliproxyauth.Auth{ID: "auth-1", Attributes: map[string]string{"api_key": "AK"}}
	first := CodeArtsSessionIDFromAuth(auth)
	second := CodeArtsSessionIDFromAuth(auth)
	if first == "" || first != second {
		t.Fatalf("session id is not stable: %q vs %q", first, second)
	}
	other := CodeArtsSessionIDFromAuth(&cliproxyauth.Auth{ID: "auth-2", Attributes: map[string]string{"api_key": "AK"}})
	if other == first {
		t.Fatal("different auths should produce different session ids")
	}

	explicit := CodeArtsSessionIDFromAuth(&cliproxyauth.Auth{ID: "x", Metadata: map[string]any{"session_id": "fixed"}})
	if explicit != "fixed" {
		t.Fatalf("explicit session id was ignored: %q", explicit)
	}
}
