package executor

import (
	"context"
	"net/http"
	"strings"
	"time"

	xiaohuanxiongauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	xiaohuanxionght "github.com/router-for-me/CLIProxyAPI/v8/internal/thinking/provider/xiaohuanxiong"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	log "github.com/sirupsen/logrus"
)

// XiaohuanxiongExecutor talks to the SenseTime Xiaohuanxiong (商汤小浣熊 /
// Raccoon) OpenAI-compatible gateway.
//
// The gateway is a plain OpenAI chat-completions surface, so this executor
// embeds OpenAICompatExecutor and only adds what the gateway needs:
//
//   - bearer-token credentials taken from the OAuth access token, with refresh
//     support against /api/web/auth/v1/refresh
//   - the per-model-family deep-thinking dialect, applied to the final upstream
//     body because the shared thinking pipeline sees only the "openai" format
//   - the vendor correlation headers the upstream expects
type XiaohuanxiongExecutor struct {
	*OpenAICompatExecutor
}

// NewXiaohuanxiongExecutor constructs a Xiaohuanxiong executor.
func NewXiaohuanxiongExecutor(cfg *config.Config) *XiaohuanxiongExecutor {
	e := NewOpenAICompatExecutor(constant.Xiaohuanxiong, cfg)
	// The gateway speaks the OpenAI wire protocol, so the default translation
	// already produces the correct body shape. Only the thinking dialect needs
	// rewriting, which is done on the final upstream payload.
	e.outgoingTransforms = applyXiaohuanxiongOutgoingTransforms
	return &XiaohuanxiongExecutor{OpenAICompatExecutor: e}
}

// Identifier returns the executor identifier.
func (e *XiaohuanxiongExecutor) Identifier() string { return constant.Xiaohuanxiong }

// PrepareRequest injects Xiaohuanxiong credentials into an ad-hoc request.
func (e *XiaohuanxiongExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	return e.OpenAICompatExecutor.PrepareRequest(req, prepareXiaohuanxiongAuth(auth))
}

// HttpRequest executes an ad-hoc Xiaohuanxiong request with normalized credentials.
func (e *XiaohuanxiongExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	return e.OpenAICompatExecutor.HttpRequest(ctx, prepareXiaohuanxiongAuth(auth), req)
}

// Refresh rotates Xiaohuanxiong credentials using the stored refresh token.
//
// Xiaohuanxiong issues short-lived JWT access tokens alongside a refresh token.
// When only an opaque access token is present (manual paste path) there is
// nothing to refresh and the auth is returned unchanged, matching the upstream
// behavior where an expired token simply requires a new login.
func (e *XiaohuanxiongExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	if refreshed, handled, err := helps.RefreshAuthViaHome(ctx, e.cfg, auth); handled {
		return refreshed, err
	}
	if auth == nil {
		return nil, nil
	}

	refreshToken := xiaohuanxiongMetadataString(auth, "refresh_token")
	if refreshToken == "" {
		// Nothing to rotate; the access token is authoritative.
		return auth, nil
	}

	token, errRefresh := xiaohuanxiongauth.NewClientWithProxyURL(e.cfg, auth.ProxyURL).Refresh(ctx, refreshToken)
	if errRefresh != nil {
		return nil, errRefresh
	}

	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["type"] = constant.Xiaohuanxiong
	auth.Metadata["auth_kind"] = cliproxyauth.AuthKindOAuth
	auth.Metadata["access_token"] = token.AccessToken
	if strings.TrimSpace(token.RefreshToken) != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if expiry := xiaohuanxiongauth.ExpiryRFC3339(token.AccessToken); expiry != "" {
		auth.Metadata["expired"] = expiry
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)

	// The access token is read back from metadata by prepareXiaohuanxiongAuth, so
	// a stale api_key attribute must be dropped for the rotated token to apply.
	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes["api_key"] = token.AccessToken
	auth.Attributes[cliproxyauth.AttributeAuthKind] = cliproxyauth.AuthKindOAuth
	if strings.TrimSpace(auth.Attributes["base_url"]) == "" {
		auth.Attributes["base_url"] = xiaohuanxiongauth.LLMBaseURL
	}
	return auth, nil
}

// prepareXiaohuanxiongAuth normalizes credentials so the embedded
// OpenAI-compatible executor can issue the request.
//
// The executor reads base_url and api_key from attributes, while OAuth logins
// store the token under metadata. This bridges both and falls back to the
// documented gateway base URL.
func prepareXiaohuanxiongAuth(auth *cliproxyauth.Auth) *cliproxyauth.Auth {
	if auth == nil {
		return nil
	}
	prepared := auth.Clone()
	if prepared.Attributes == nil {
		prepared.Attributes = make(map[string]string)
	}

	if strings.TrimSpace(prepared.Attributes["base_url"]) == "" {
		baseURL := xiaohuanxiongMetadataString(prepared, "base_url")
		if baseURL == "" {
			baseURL = xiaohuanxiongauth.LLMBaseURL
		}
		prepared.Attributes["base_url"] = baseURL
	}

	if strings.TrimSpace(prepared.Attributes["api_key"]) == "" {
		prepared.Attributes["api_key"] = xiaohuanxiongMetadataString(prepared, "access_token")
	}

	return prepared
}

// xiaohuanxiongMetadataString reads a trimmed string value from auth metadata.
func xiaohuanxiongMetadataString(auth *cliproxyauth.Auth, key string) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	value, _ := auth.Metadata[key].(string)
	return strings.TrimSpace(value)
}

// applyXiaohuanxiongOutgoingTransforms rewrites the final upstream body.
//
// The shared thinking pipeline dispatches on the request format ("openai" for
// this gateway), so the gateway's per-model-family dialects must be applied
// here, on the payload that is actually sent.
func applyXiaohuanxiongOutgoingTransforms(_ context.Context, _ *cliproxyauth.Auth, baseModel string, _ cliproxyexecutor.Options, translated []byte) []byte {
	if len(translated) == 0 {
		return translated
	}
	updated, errApply := xiaohuanxionght.TranslateRequestThinking(translated, baseModel)
	if errApply != nil {
		log.Warnf("xiaohuanxiong executor: apply thinking dialect failed: %v", errApply)
		return translated
	}
	return updated
}
