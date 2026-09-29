package executor

import (
	"context"
	"encoding/json"
	"testing"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	"github.com/tidwall/gjson"
)

// TestCodeArtsExecutorIdentifier pins the provider key so the conductor's
// registration switch keeps matching.
func TestCodeArtsExecutorIdentifier(t *testing.T) {
	e := NewCodeArtsExecutor(nil)
	if got := e.Identifier(); got != constant.CodeArts {
		t.Fatalf("Identifier() = %q, want %q", got, constant.CodeArts)
	}
}

// TestPrepareCodeArtsAuthBridgesMetadata verifies the credential triple is
// bridged from metadata into the attributes the shared executor and the signing
// transport read.
func TestPrepareCodeArtsAuthBridgesMetadata(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: constant.CodeArts,
		Metadata: map[string]any{
			"access_key":     "AK-from-metadata",
			"secret_key":     "SK-from-metadata",
			"security_token": "ST-from-metadata",
		},
	}
	prepared := prepareCodeArtsAuth(auth)
	if prepared == nil {
		t.Fatal("prepared auth is nil")
	}
	if got := prepared.Attributes["api_key"]; got != "AK-from-metadata" {
		t.Fatalf("api_key = %q", got)
	}
	if got := prepared.Attributes["secret_key"]; got != "SK-from-metadata" {
		t.Fatalf("secret_key = %q", got)
	}
	if got := prepared.Attributes["security_token"]; got != "ST-from-metadata" {
		t.Fatalf("security_token = %q", got)
	}
	if got := prepared.Attributes["base_url"]; got != codeartsauth.InferHubBaseURL {
		t.Fatalf("base_url = %q, want %q", got, codeartsauth.InferHubBaseURL)
	}
	// The source auth must not be mutated.
	if auth.Attributes != nil && auth.Attributes["api_key"] != "" {
		t.Fatal("prepareCodeArtsAuth mutated the source auth")
	}
}

// TestPrepareCodeArtsAuthPrefersAttributes keeps operator overrides.
func TestPrepareCodeArtsAuthPrefersAttributes(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: constant.CodeArts,
		Attributes: map[string]string{
			"api_key":        "explicit-ak",
			"secret_key":     "explicit-sk",
			"security_token": "explicit-st",
			"base_url":       "https://self-hosted.example/api/v2",
		},
		Metadata: map[string]any{"access_key": "metadata-ak"},
	}
	prepared := prepareCodeArtsAuth(auth)
	if prepared.Attributes["api_key"] != "explicit-ak" {
		t.Fatalf("api_key = %q", prepared.Attributes["api_key"])
	}
	if prepared.Attributes["base_url"] != "https://self-hosted.example/api/v2" {
		t.Fatalf("base_url = %q", prepared.Attributes["base_url"])
	}
}

// TestPrepareCodeArtsAuthAddsScheme tolerates a scheme-less base URL.
func TestPrepareCodeArtsAuthAddsScheme(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider:   constant.CodeArts,
		Attributes: map[string]string{"base_url": "snap-access.example/api/v2"},
	}
	prepared := prepareCodeArtsAuth(auth)
	if prepared.Attributes["base_url"] != "https://snap-access.example/api/v2" {
		t.Fatalf("base_url = %q", prepared.Attributes["base_url"])
	}
}

// TestPrepareCodeArtsAuthNil guards the nil auth path.
func TestPrepareCodeArtsAuthNil(t *testing.T) {
	if got := prepareCodeArtsAuth(nil); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

// TestCodeArtsRefreshWithoutRefreshToken is the manual-paste path: a static
// AK/SK triple cannot rotate, and that must not be an error.
func TestCodeArtsRefreshWithoutRefreshToken(t *testing.T) {
	e := NewCodeArtsExecutor(nil)
	auth := &cliproxyauth.Auth{
		Provider: constant.CodeArts,
		Metadata: map[string]any{"access_key": "AK", "secret_key": "SK"},
	}
	refreshed, err := e.Refresh(context.Background(), auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshed != auth {
		t.Fatal("auth should pass through unchanged when there is nothing to rotate")
	}
}

// TestCodeArtsRefreshRequiresDpopKey pins that rotation cannot silently proceed
// without the key pair Huawei Cloud STS needs.
func TestCodeArtsRefreshRequiresDpopKey(t *testing.T) {
	e := NewCodeArtsExecutor(nil)
	auth := &cliproxyauth.Auth{
		Provider: constant.CodeArts,
		Metadata: map[string]any{
			"access_key":    "AK",
			"secret_key":    "SK",
			"refresh_token": "rt",
		},
	}
	if _, err := e.Refresh(context.Background(), auth); err == nil {
		t.Fatal("expected an error when the DPoP key pair is missing")
	}
}

// TestApplyCodeArtsOutgoingTransforms pins the CodeArts wire dialect:
// tool_stream mirrors the stream flag, enable_thinking is derived from
// reasoning_effort, max_tokens is bounded, and reasoning_effort never leaks.
func TestApplyCodeArtsOutgoingTransforms(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantThinking  bool
		wantToolPool  bool
		wantMaxTokens int
	}{
		{
			name:          "streaming high effort enables thinking and tool_stream",
			body:          `{"model":"GLM-5.2","stream":true,"reasoning_effort":"high","max_tokens":1024}`,
			wantThinking:  true,
			wantToolPool:  true,
			wantMaxTokens: 1024,
		},
		{
			name:          "streaming none effort disables thinking",
			body:          `{"model":"GLM-5.2","stream":true,"reasoning_effort":"none","max_tokens":1024}`,
			wantThinking:  false,
			wantToolPool:  true,
			wantMaxTokens: 1024,
		},
		{
			name:          "absent max_tokens is filled from the catalog",
			body:          `{"model":"GLM-5.2","stream":true}`,
			wantThinking:  false,
			wantToolPool:  true,
			wantMaxTokens: codeArtsMaxOutputTokens,
		},
		{
			name:          "non-streaming requests do not ask for tool streaming",
			body:          `{"model":"GLM-5.2","stream":false,"max_tokens":512}`,
			wantThinking:  false,
			wantToolPool:  false,
			wantMaxTokens: 512,
		},
		{
			// The gateway rejects anything above 65536 with a RANGE_VALIDATOR
			// 400, so an over-limit request must be clamped rather than failed.
			name:          "over-limit max_tokens is clamped to the gateway ceiling",
			body:          `{"model":"GLM-5.2","stream":false,"max_tokens":200000}`,
			wantThinking:  false,
			wantToolPool:  false,
			wantMaxTokens: codeArtsMaxOutputTokens,
		},
		{
			name:          "limit exactly at the ceiling is preserved",
			body:          `{"model":"GLM-5.2","stream":false,"max_tokens":65536}`,
			wantThinking:  false,
			wantToolPool:  false,
			wantMaxTokens: codeArtsMaxOutputTokens,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := applyCodeArtsOutgoingTransforms(context.Background(), nil, "GLM-5.2", cliproxyexecutor.Options{}, []byte(tc.body))
			var decoded map[string]any
			if errUnmarshal := json.Unmarshal(out, &decoded); errUnmarshal != nil {
				t.Fatalf("result is not valid JSON: %s", out)
			}
			if got := gjson.GetBytes(out, "tool_stream").Bool(); got != tc.wantToolPool {
				t.Fatalf("tool_stream = %v, want %v", got, tc.wantToolPool)
			}
			if gjson.GetBytes(out, "reasoning_effort").Exists() {
				t.Fatalf("reasoning_effort leaked to the upstream body: %s", out)
			}
			gotMax := gjson.GetBytes(out, "max_tokens").Int()
			if gotMax != int64(tc.wantMaxTokens) {
				t.Fatalf("max_tokens = %d, want %d (body=%s)", gotMax, tc.wantMaxTokens, out)
			}
			if gotMax > codeArtsMaxOutputTokens {
				t.Fatalf("max_tokens %d exceeds the gateway ceiling %d: %s", gotMax, codeArtsMaxOutputTokens, out)
			}
			if effort := gjson.GetBytes([]byte(tc.body), "reasoning_effort"); effort.Exists() {
				if got := gjson.GetBytes(out, "enable_thinking").Bool(); got != tc.wantThinking {
					t.Fatalf("enable_thinking = %v, want %v (body=%s)", got, tc.wantThinking, out)
				}
			}
		})
	}
}

// TestApplyCodeArtsOutgoingTransformsEmptyBody guards the empty payload path.
func TestApplyCodeArtsOutgoingTransformsEmptyBody(t *testing.T) {
	if out := applyCodeArtsOutgoingTransforms(context.Background(), nil, "GLM-5.2", cliproxyexecutor.Options{}, nil); len(out) != 0 {
		t.Fatalf("expected empty passthrough, got %q", out)
	}
	if out := applyCodeArtsOutgoingTransforms(context.Background(), nil, "GLM-5.2", cliproxyexecutor.Options{}, []byte("not json")); string(out) != "not json" {
		t.Fatalf("expected invalid JSON passthrough, got %q", out)
	}
}

// TestCodeArtsCredentialsExpired pins the refresh-window decision.
func TestCodeArtsCredentialsExpired(t *testing.T) {
	keyPair, errKey := codeartsauth.GenerateDpopKeyPair()
	if errKey != nil {
		t.Fatalf("generate key pair: %v", errKey)
	}
	base := func(expired string) *cliproxyauth.Auth {
		return &cliproxyauth.Auth{
			Provider: constant.CodeArts,
			Metadata: map[string]any{
				"access_key":       "AK",
				"secret_key":       "SK",
				"refresh_token":    "rt",
				"expired":          expired,
				"dpop_private_key": keyPair.PrivateKey,
				"dpop_public_key":  keyPair.PublicKey,
			},
		}
	}
	if codeArtsCredentialsExpired(base("2099-01-01T00:00:00Z")) {
		t.Fatal("a far-future credential must not be considered expired")
	}
	if !codeArtsCredentialsExpired(base("2020-01-01T00:00:00Z")) {
		t.Fatal("a past credential must be considered expired")
	}
	if codeArtsCredentialsExpired(base("")) {
		t.Fatal("an unknown expiry must not force a refresh")
	}
}

// TestCodeArtsCredentialsExpiredWithoutKeyPair pins that a missing DPoP key
// blocks the automatic refresh path (it would fail upstream anyway).
func TestCodeArtsCredentialsExpiredWithoutKeyPair(t *testing.T) {
	auth := &cliproxyauth.Auth{
		Provider: constant.CodeArts,
		Metadata: map[string]any{
			"access_key":    "AK",
			"secret_key":    "SK",
			"refresh_token": "rt",
			"expired":       "2020-01-01T00:00:00Z",
		},
	}
	if codeArtsCredentialsExpired(auth) {
		t.Fatal("without a DPoP key pair the credential cannot be refreshed automatically")
	}
}

// TestApplyCodeArtsTokenWritesBack pins the rotated credential mapping.
func TestApplyCodeArtsTokenWritesBack(t *testing.T) {
	e := NewCodeArtsExecutor(nil)
	auth := &cliproxyauth.Auth{Provider: constant.CodeArts, Metadata: map[string]any{}, Attributes: map[string]string{}}
	refreshed := applyCodeArtsToken(auth, &codeartsauth.TokenData{
		AccessKey:     "AK2",
		SecretKey:     "SK2",
		SecurityToken: "ST2",
		RefreshToken:  "rt2",
		CodeVerifier:  "verifier2",
		DpopKeyPair:   &codeartsauth.DpopKeyPair{PrivateKey: "priv", PublicKey: "pub"},
	})
	if refreshed.Metadata["access_key"] != "AK2" || refreshed.Metadata["security_token"] != "ST2" {
		t.Fatalf("metadata not updated: %+v", refreshed.Metadata)
	}
	if refreshed.Attributes["secret_key"] != "SK2" {
		t.Fatalf("attributes not updated: %+v", refreshed.Attributes)
	}
	if refreshed.Metadata["dpop_private_key"] != "priv" {
		t.Fatal("the DPoP key pair must persist across a rotation")
	}
	if refreshed.Attributes["base_url"] != codeartsauth.InferHubBaseURL {
		t.Fatalf("base_url = %q", refreshed.Attributes["base_url"])
	}
	_ = e
}
