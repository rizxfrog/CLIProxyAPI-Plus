// Package executor provides per-provider runtime executors.
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/sjson"

	clineauth "github.com/router-for-me/CLIProxyAPI/v7/internal/auth/cline"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/buildinfo"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/runtime/executor/helps"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

// ClineBaseURL is the Cline (cline.bot) OpenAI-compatible gateway
// chat-completions endpoint. The executor derives its base URL from the
// synthesized auth attributes; this constant documents the canonical endpoint.
const ClineBaseURL = clineauth.BaseURL + "/chat/completions"

// ClineExecutor talks to the Cline (cline.bot) OpenAI-compatible gateway.
//
// Two upstream requirements shape this executor:
//
//   - Cline's API only implements streaming (streamText); a non-streaming
//     request returns an empty body. The executor therefore forces upstream
//     streaming and folds the SSE chunks back into a chat.completion for
//     stream:false clients.
//   - The upstream bearer token is sent as a plain WorkOS access token
//     ("Bearer <token>"). Cline rejects the "workos:" prefix that older
//     integrations assumed; the prefix is NOT added here.
//   - Non-streaming chat responses are wrapped in a {"success":true,"data":{...}}
//     envelope; Execute unwraps it so stream:false clients receive a plain
//     chat.completion.
type ClineExecutor struct {
	*OpenAICompatExecutor
}

// Compile-time assertion that ClineExecutor satisfies the full ProviderExecutor
// contract (including the inherited CountTokens / HttpRequest methods).
var _ cliproxyauth.ProviderExecutor = (*ClineExecutor)(nil)

// NewClineExecutor constructs a Cline executor.
func NewClineExecutor(cfg *config.Config) *ClineExecutor {
	e := NewOpenAICompatExecutor(constant.Cline, cfg)
	e.outgoingTransforms = applyClineOutgoingTransforms
	return &ClineExecutor{
		OpenAICompatExecutor: e,
	}
}

// Identifier returns the executor identifier.
func (e *ClineExecutor) Identifier() string { return constant.Cline }

// Execute drives a streamed upstream request and folds the OpenAI chunks into a
// single chat.completion, because Cline's API only implements streaming.
func (e *ClineExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	opts.Stream = true
	streamResult, errStream := e.ExecuteStream(ctx, auth, req, opts)
	if errStream != nil {
		return cliproxyexecutor.Response{}, errStream
	}
	if streamResult == nil {
		return cliproxyexecutor.Response{}, nil
	}
	var buffer bytes.Buffer
	for chunk := range streamResult.Chunks {
		if chunk.Err != nil {
			return cliproxyexecutor.Response{}, chunk.Err
		}
		if len(chunk.Payload) > 0 {
			_, _ = buffer.Write(chunk.Payload)
			_, _ = buffer.Write([]byte("\n"))
		}
	}
	return cliproxyexecutor.Response{
		Payload: unwrapClineEnvelope(aggregateCodeBuddyCNChunks(buffer.Bytes())),
		Headers: streamResult.Headers,
	}, nil
}

// ExecuteStream forces streaming and delegates to the OpenAI-compatible executor
// with normalized Cline credentials.
func (e *ClineExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	opts.Stream = true
	return e.OpenAICompatExecutor.ExecuteStream(ctx, prepareClineAuth(auth), req, opts)
}

// PrepareRequest injects Cline OAuth or API-key credentials into ad-hoc requests.
func (e *ClineExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	return e.OpenAICompatExecutor.PrepareRequest(req, prepareClineAuth(auth))
}

// HttpRequest executes an ad-hoc Cline request with normalized credentials.
func (e *ClineExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	return e.OpenAICompatExecutor.HttpRequest(ctx, prepareClineAuth(auth), req)
}

// Refresh rotates Cline OAuth credentials using the stored refresh token.
func (e *ClineExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	return refreshClineAuth(ctx, e.OpenAICompatExecutor, auth)
}

// prepareClineAuth normalizes Cline credentials: it ensures the base URL is set,
// resolves the access token from metadata, and injects Cline's
// client-identification headers. Cline authenticates with a plain WorkOS access
// token (Bearer <token>); the legacy "workos:" prefix is NOT applied because the
// current API rejects it. BYOK api keys (sk-...) are passed through as-is.
func prepareClineAuth(auth *cliproxyauth.Auth) *cliproxyauth.Auth {
	if auth == nil {
		return nil
	}
	prepared := auth.Clone()
	if prepared.Attributes == nil {
		prepared.Attributes = make(map[string]string)
	}
	if strings.TrimSpace(prepared.Attributes["base_url"]) == "" {
		baseURL := codeBuddyCNMetadataString(prepared, "base_url")
		if baseURL == "" {
			baseURL = clineauth.BaseURL
		}
		prepared.Attributes["base_url"] = baseURL
	}
	// Resolve the raw token from attributes then metadata. Cline authenticates
	// with the bare access token; no "workos:" prefix is added.
	rawToken := strings.TrimSpace(prepared.Attributes["api_key"])
	if rawToken == "" {
		rawToken = codeBuddyCNMetadataString(prepared, "access_token")
	}
	if rawToken != "" {
		prepared.Attributes["api_key"] = rawToken
	}
	for name, value := range clineDefaultHeaders() {
		if !codeBuddyCNHasCustomHeader(prepared.Attributes, name) {
			prepared.Attributes["header:"+name] = value
		}
	}
	return prepared
}

// refreshClineAuth rotates a Cline credential using the OAuth refresh endpoint.
func refreshClineAuth(ctx context.Context, base *OpenAICompatExecutor, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	if refreshed, handled, err := helps.RefreshAuthViaHome(ctx, base.cfg, auth); handled {
		return refreshed, err
	}
	if auth == nil {
		return nil, fmt.Errorf("cline executor: auth is nil")
	}
	refreshToken := codeBuddyCNMetadataString(auth, "refresh_token")
	if refreshToken == "" {
		return auth, nil
	}
	client := clineauth.NewClientWithProxyURL(base.cfg, auth.ProxyURL)
	token, errRefresh := client.ExchangeToken(ctx, refreshToken)
	if errRefresh != nil {
		return nil, errRefresh
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["type"] = constant.Cline
	auth.Metadata["auth_kind"] = cliproxyauth.AuthKindOAuth
	auth.Metadata["access_token"] = token.AccessToken
	if strings.TrimSpace(token.RefreshToken) != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if token.ExpiresAt > 0 {
		auth.Metadata["expires_at"] = token.ExpiresAt
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)
	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes[cliproxyauth.AttributeAuthKind] = cliproxyauth.AuthKindOAuth
	auth.Attributes["api_key"] = token.AccessToken
	if strings.TrimSpace(auth.Attributes["base_url"]) == "" {
		auth.Attributes["base_url"] = clineauth.BaseURL
	}
	return auth, nil
}

// applyClineOutgoingTransforms forces streaming on the final upstream body.
// Cline rejects non-streaming chat requests with an empty response.
func applyClineOutgoingTransforms(ctx context.Context, auth *cliproxyauth.Auth, baseModel string, opts cliproxyexecutor.Options, translated []byte) []byte {
	if len(translated) == 0 {
		return translated
	}
	body, _ := sjson.SetBytes(translated, "stream", true)
	return body
}

// unwrapClineEnvelope strips Cline's {"success":true,"data":{...}} response
// envelope, returning the inner OpenAI chat.completion (or chunk stream) payload
// unchanged when it is not wrapped. Cline applies this envelope to non-streaming
// chat responses; CLIProxyAPI forces streaming, but ad-hoc and fallback paths
// may observe it, so callers should unwrap defensively.
func unwrapClineEnvelope(payload []byte) []byte {
	trimmed := bytes.TrimSpace(payload)
	if !bytes.HasPrefix(trimmed, []byte("{")) {
		return payload
	}
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(trimmed, &envelope); err != nil {
		return payload
	}
	if envelope.Data == nil || !envelope.Success {
		return payload
	}
	return envelope.Data
}

// clineDefaultHeaders returns the Cline client-identification headers required
// by the upstream, mirroring the extension/CLI request shape.
func clineDefaultHeaders() map[string]string {
	version := strings.TrimSpace(buildinfo.Version)
	if version == "" || version == "dev" {
		version = "1.0.0"
	}
	return map[string]string{
		"HTTP-Referer":       "https://cline.bot",
		"X-Title":            "Cline",
		"User-Agent":         "Cline/" + version,
		"X-CLIENT-TYPE":      "cline",
		"X-CLIENT-VERSION":   version,
		"X-PLATFORM":         "linux",
		"X-PLATFORM-VERSION": "unknown",
		"X-CORE-VERSION":     version,
		"X-IS-MULTIROOT":     "false",
	}
}
