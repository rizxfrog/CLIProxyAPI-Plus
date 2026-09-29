package executor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	// codeArtsDefaultMaxTokens bounds the output when a model declares no limit.
	codeArtsDefaultMaxTokens = 32000

	// codeArtsMaxOutputTokens is the hard ceiling the gateway enforces on
	// max_tokens. It is validated server-side (RANGE_VALIDATOR) and is
	// independent of the per-model values advertised by the model catalog:
	// every CodeArts model, including those advertising 131072 or 393216,
	// rejects max_tokens > 65536 with HTTP 400. Requests above this ceiling
	// must be clamped or they fail before reaching the model.
	codeArtsMaxOutputTokens = 65536
)

// CodeArtsExecutor talks to the Huawei Cloud CodeArts (CodeArts Work) LLM
// gateway.
//
// The gateway is a plain OpenAI chat-completions surface (POST /api/v2/chat/completions),
// so this executor embeds OpenAICompatExecutor and only adds what CodeArts
// needs:
//
//   - Huawei Cloud SDK-HMAC-SHA256 signing with a temporary AK/SK credential
//     triple. Signing is installed as a RoundTripper so it always covers the
//     final request bytes.
//   - credential rotation against Huawei Cloud STS (OAuth2 + PKCE + DPoP)
//   - the CodeArts thinking dialect (enable_thinking / tool_stream)
type CodeArtsExecutor struct {
	*OpenAICompatExecutor
}

// NewCodeArtsExecutor constructs a CodeArts executor.
func NewCodeArtsExecutor(cfg *config.Config) *CodeArtsExecutor {
	e := NewOpenAICompatExecutor(constant.CodeArts, cfg)
	e.httpClientFactory = helps.NewCodeArtsHTTPClient
	e.outgoingTransforms = applyCodeArtsOutgoingTransforms
	return &CodeArtsExecutor{OpenAICompatExecutor: e}
}

// Identifier returns the executor identifier.
func (e *CodeArtsExecutor) Identifier() string { return constant.CodeArts }

// PrepareRequest normalizes the credential attributes. The actual Huawei Cloud
// signature is applied by the client returned from NewCodeArtsHTTPClient, so an
// ad-hoc request must be executed through HttpRequest/HttpClient rather than
// plain http.Client.
func (e *CodeArtsExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	if req == nil {
		return nil
	}
	prepared := prepareCodeArtsAuth(auth)
	if prepared == nil {
		return codeArtsStatusError{code: http.StatusUnauthorized, msg: "codearts executor: auth is nil"}
	}
	return e.OpenAICompatExecutor.PrepareRequest(req, prepared)
}

// HttpRequest signs and executes an ad-hoc CodeArts request.
func (e *CodeArtsExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("codearts executor: request is nil")
	}
	prepared := prepareCodeArtsAuth(auth)
	if prepared == nil {
		return nil, codeArtsStatusError{code: http.StatusUnauthorized, msg: "codearts executor: auth is nil"}
	}
	return e.OpenAICompatExecutor.HttpRequest(ctx, prepared, req)
}

// Execute runs a non-streaming chat completion.
func (e *CodeArtsExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	prepared, errPrepare := e.refreshIfNeeded(ctx, auth)
	if errPrepare != nil {
		return cliproxyexecutor.Response{}, errPrepare
	}
	return e.OpenAICompatExecutor.Execute(ctx, prepared, req, opts)
}

// ExecuteStream runs a streaming chat completion.
func (e *CodeArtsExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	prepared, errPrepare := e.refreshIfNeeded(ctx, auth)
	if errPrepare != nil {
		return nil, errPrepare
	}
	return e.OpenAICompatExecutor.ExecuteStream(ctx, prepared, req, opts)
}

// RequestToFormat reports the upstream request format.
func (e *CodeArtsExecutor) RequestToFormat(_ cliproxyexecutor.Request, _ cliproxyexecutor.Options) sdktranslator.Format {
	return sdktranslator.FormatOpenAI
}

// refreshIfNeeded rotates the temporary credentials when they are inside the
// refresh window, then returns the auth the shared executor should use.
func (e *CodeArtsExecutor) refreshIfNeeded(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	prepared := prepareCodeArtsAuth(auth)
	if prepared == nil {
		return nil, codeArtsStatusError{code: http.StatusUnauthorized, msg: "codearts executor: missing credentials"}
	}
	if !codeArtsCredentialsExpired(prepared) {
		return prepared, nil
	}
	refreshed, errRefresh := e.Refresh(ctx, prepared)
	if errRefresh != nil {
		// A failed rotation is not fatal: the current credential may still be
		// valid and the upstream is authoritative.
		log.Warnf("codearts executor: credential refresh failed: %v", errRefresh)
		return prepared, nil
	}
	if refreshed == nil {
		return prepared, nil
	}
	return prepareCodeArtsAuth(refreshed), nil
}

// Refresh rotates the temporary Huawei Cloud credentials.
//
// Refresh needs the stored refresh token, the PKCE verifier and the original
// DPoP key pair, because Huawei Cloud STS re-validates proof of possession. When
// any part is missing the credential cannot be rotated and the auth is returned
// unchanged so the upstream surfaces the expiry to the operator.
func (e *CodeArtsExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	if refreshed, handled, err := helps.RefreshAuthViaHome(ctx, e.cfg, auth); handled {
		return refreshed, err
	}
	if auth == nil {
		return nil, nil
	}

	refreshToken := codeArtsMetadataString(auth, "refresh_token")
	if refreshToken == "" {
		return auth, nil
	}
	keyPair := codeArtsDpopKeyPair(auth)
	if keyPair == nil {
		return nil, fmt.Errorf("codearts: cannot refresh credentials without the original DPoP key pair")
	}

	client := codeartsauth.NewClientWithProxyURL(e.cfg, auth.ProxyURL)
	token, errRefresh := client.Refresh(ctx, refreshToken, codeArtsMetadataString(auth, "code_verifier"), keyPair)
	if errRefresh != nil {
		return nil, errRefresh
	}

	updated := applyCodeArtsToken(auth.Clone(), token)
	name := codeArtsMetadataString(auth, "user_name")
	if name == "" {
		name = token.AccessKey
	}
	updated.Label = firstNonEmptyString(updated.Label, "codearts-"+name)
	return updated, nil
}

// prepareCodeArtsAuth bridges the stored credential into the attributes the
// embedded OpenAI-compatible executor reads (api_key/base_url), keeping the
// CodeArts-specific fields the signing transport needs.
func prepareCodeArtsAuth(auth *cliproxyauth.Auth) *cliproxyauth.Auth {
	if auth == nil {
		return nil
	}
	prepared := auth.Clone()
	if prepared.Attributes == nil {
		prepared.Attributes = make(map[string]string)
	}
	if strings.TrimSpace(prepared.Attributes["base_url"]) == "" {
		baseURL := firstNonEmptyString(codeArtsMetadataString(prepared, "base_url"), codeartsauth.InferHubBaseURL)
		if !strings.Contains(baseURL, "://") {
			baseURL = "https://" + baseURL
		}
		prepared.Attributes["base_url"] = baseURL
	} else if !strings.Contains(prepared.Attributes["base_url"], "://") {
		prepared.Attributes["base_url"] = "https://" + prepared.Attributes["base_url"]
	}
	if strings.TrimSpace(prepared.Attributes["api_key"]) == "" {
		prepared.Attributes["api_key"] = firstNonEmptyString(
			codeArtsMetadataString(prepared, "access_key"),
			codeArtsMetadataString(prepared, "api_key"),
		)
	}
	if strings.TrimSpace(prepared.Attributes["secret_key"]) == "" {
		prepared.Attributes["secret_key"] = codeArtsMetadataString(prepared, "secret_key")
	}
	if strings.TrimSpace(prepared.Attributes["security_token"]) == "" {
		prepared.Attributes["security_token"] = codeArtsMetadataString(prepared, "security_token")
	}
	// The OpenAI-compatible executor sets a bearer Authorization header that
	// Huawei Cloud does not accept; the signing transport removes it, but keep a
	// placeholder so the credential check passes.
	if prepared.Attributes["api_key"] == "" {
		prepared.Attributes["api_key"] = "codearts-signed"
	}
	return prepared
}

// codeArtsMetadataString reads a trimmed string value from auth metadata.
func codeArtsMetadataString(auth *cliproxyauth.Auth, key string) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	value, _ := auth.Metadata[key].(string)
	return strings.TrimSpace(value)
}

// codeArtsDpopKeyPair reconstructs the DPoP key pair from auth metadata.
func codeArtsDpopKeyPair(auth *cliproxyauth.Auth) *codeartsauth.DpopKeyPair {
	privateKey := codeArtsMetadataString(auth, "dpop_private_key")
	if privateKey == "" {
		return nil
	}
	return &codeartsauth.DpopKeyPair{
		PrivateKey: privateKey,
		PublicKey:  codeArtsMetadataString(auth, "dpop_public_key"),
	}
}

// codeArtsCredentialsExpired reports whether the credential bundle is inside the
// refresh window.
func codeArtsCredentialsExpired(auth *cliproxyauth.Auth) bool {
	if auth == nil {
		return false
	}
	if codeArtsMetadataString(auth, "refresh_token") == "" {
		return false
	}
	if codeArtsDpopKeyPair(auth) == nil {
		return false
	}
	raw := codeArtsMetadataString(auth, "expired")
	if raw == "" {
		return false
	}
	expiry, errParse := time.Parse(time.RFC3339, raw)
	if errParse != nil {
		return false
	}
	return time.Now().Add(codeartsauth.RefreshWindowSeconds * time.Second).After(expiry)
}

// applyCodeArtsToken writes a rotated credential bundle back onto an auth.
func applyCodeArtsToken(auth *cliproxyauth.Auth, token *codeartsauth.TokenData) *cliproxyauth.Auth {
	if auth == nil || token == nil {
		return auth
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["type"] = constant.CodeArts
	auth.Metadata["auth_kind"] = cliproxyauth.AuthKindOAuth
	auth.Metadata["access_key"] = token.AccessKey
	auth.Metadata["secret_key"] = token.SecretKey
	auth.Metadata["security_token"] = token.SecurityToken
	if token.RefreshToken != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if token.CodeVerifier != "" {
		auth.Metadata["code_verifier"] = token.CodeVerifier
	}
	if token.DpopKeyPair != nil {
		auth.Metadata["dpop_private_key"] = token.DpopKeyPair.PrivateKey
		auth.Metadata["dpop_public_key"] = token.DpopKeyPair.PublicKey
	}
	if !token.ExpiresAt.IsZero() {
		auth.Metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)

	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes["api_key"] = token.AccessKey
	auth.Attributes["secret_key"] = token.SecretKey
	auth.Attributes["security_token"] = token.SecurityToken
	auth.Attributes[cliproxyauth.AttributeAuthKind] = cliproxyauth.AuthKindOAuth
	if strings.TrimSpace(auth.Attributes["base_url"]) == "" {
		auth.Attributes["base_url"] = codeartsauth.InferHubBaseURL
	}
	return auth
}

// applyCodeArtsOutgoingTransforms rewrites the final upstream body with the
// CodeArts dialect.
//
// The desktop kernel always sends tool_stream=true, expresses thinking through
// enable_thinking, and never sends reasoning_effort.
func applyCodeArtsOutgoingTransforms(_ context.Context, _ *cliproxyauth.Auth, baseModel string, _ cliproxyexecutor.Options, translated []byte) []byte {
	if len(translated) == 0 || !gjson.ValidBytes(translated) {
		return translated
	}
	// The gateway requires a bounded output. Honor the client's value when it is
	// within the server ceiling, else fall back to the catalog/default. Any
	// value above the ceiling is rejected outright, so it is always clamped.
	limit := codeArtsDefaultMaxTokens
	if info := registry.LookupModelInfo(baseModel, constant.CodeArts); info != nil && info.MaxCompletionTokens > 0 {
		limit = info.MaxCompletionTokens
	}
	if requested := gjson.GetBytes(translated, "max_tokens"); requested.Exists() && requested.Int() > 0 {
		limit = int(requested.Int())
	}
	if limit > codeArtsMaxOutputTokens {
		limit = codeArtsMaxOutputTokens
	}
	if updated, errSet := sjson.SetBytes(translated, "max_tokens", limit); errSet == nil {
		translated = updated
	}

	// The gateway streams tool-call deltas only in streaming mode, so mirror the
	// client's stream flag instead of forcing it on for non-streaming requests.
	toolStream := gjson.GetBytes(translated, "stream").Bool()
	if updated, errSet := sjson.SetBytes(translated, "tool_stream", toolStream); errSet == nil {
		translated = updated
	}

	if effort := gjson.GetBytes(translated, "reasoning_effort"); effort.Exists() {
		enabled := !strings.EqualFold(strings.TrimSpace(effort.String()), string(thinking.LevelNone))
		if updated, errSet := sjson.SetBytes(translated, "enable_thinking", enabled); errSet == nil {
			translated = updated
		}
	}
	// reasoning_effort is not part of the CodeArts wire format.
	if stripped, errDelete := sjson.DeleteBytes(translated, "reasoning_effort"); errDelete == nil {
		translated = stripped
	}
	return translated
}

// codeArtsStatusError is the executor-local error type used before an upstream
// response body exists.
type codeArtsStatusError struct {
	code int
	msg  string
}

func (e codeArtsStatusError) Error() string {
	return fmt.Sprintf("codearts executor: status %d: %s", e.code, e.msg)
}

// firstNonEmptyString returns the first trimmed non-empty value.
func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
