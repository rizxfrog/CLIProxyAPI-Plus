// Package executor provides per-provider runtime executors.
//
// This file implements the MiniMax Code (MCode) managed-account executor. The
// MiniMax Code client authenticates a user's subscription with an OAuth 2.0
// device-code grant and then speaks the Anthropic Messages wire protocol to the
// managed agent backend at
//
//	https://agent.minimax.io/mavis/api/v1/llm/v1/messages   (international)
//	https://agent.minimax.cn/mavis/api/v1/llm/v1/messages   (mainland China)
//
// with `Authorization: Bearer <oauth access token>`. CLIProxyAPI therefore keeps
// the provider identity (`minimax` / `minimax-cn`) separate from the wire format
// (`claude`, i.e. Anthropic Messages) and reuses the existing Claude translators.
package executor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	minimaxauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/minimax"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	log "github.com/sirupsen/logrus"
)

// MinimaxProviderEN and MinimaxProviderCN are the provider keys for the
// international and mainland MiniMax Code managed accounts.
const (
	MinimaxProviderEN = "minimax"
	MinimaxProviderCN = "minimax-cn"
)

// minimaxRegionFromProvider maps a provider key to its regional deployment.
func minimaxRegionFromProvider(provider string) minimaxauth.Region {
	if strings.EqualFold(strings.TrimSpace(provider), MinimaxProviderCN) ||
		strings.EqualFold(strings.TrimSpace(provider), "minimax.ai") {
		return minimaxauth.RegionCN
	}
	return minimaxauth.RegionEN
}

// MinimaxExecutor proxies requests to a MiniMax Code managed-account backend
// using Anthropic Messages semantics.
type MinimaxExecutor struct {
	cfg      *config.Config
	provider string
}

// NewMinimaxExecutor creates the international MiniMax Code executor.
func NewMinimaxExecutor(cfg *config.Config) *MinimaxExecutor {
	return &MinimaxExecutor{cfg: cfg, provider: MinimaxProviderEN}
}

// NewMinimaxCNExecutor creates the mainland China MiniMax Code executor.
func NewMinimaxCNExecutor(cfg *config.Config) *MinimaxExecutor {
	return &MinimaxExecutor{cfg: cfg, provider: MinimaxProviderCN}
}

// Identifier returns the executor/provider identifier.
func (e *MinimaxExecutor) Identifier() string {
	if e == nil || strings.TrimSpace(e.provider) == "" {
		return MinimaxProviderEN
	}
	return e.provider
}

// RequestToFormat reports the upstream wire format after auth selection. The
// MiniMax managed backend is Anthropic-Messages compatible.
func (e *MinimaxExecutor) RequestToFormat(_ cliproxyexecutor.Request, _ cliproxyexecutor.Options) sdktranslator.Format {
	return sdktranslator.FormatClaude
}

// PrepareRequest injects MiniMax credentials into the outgoing HTTP request.
func (e *MinimaxExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	if req == nil {
		return nil
	}
	token := minimaxAccessToken(auth)
	if strings.TrimSpace(token) != "" {
		req.Header.Del("x-api-key")
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", minimaxauth.UserAgent)
	var attrs map[string]string
	if auth != nil {
		attrs = auth.Attributes
	}
	util.ApplyCustomHeadersFromAttrs(req, attrs)
	return nil
}

// HttpRequest injects MiniMax credentials into the request and executes it.
func (e *MinimaxExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("minimax executor: request is nil")
	}
	if ctx == nil {
		ctx = req.Context()
	}
	httpReq := req.WithContext(ctx)
	if err := e.PrepareRequest(httpReq, auth); err != nil {
		return nil, err
	}
	httpClient := helps.NewProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	return httpClient.Do(httpReq)
}

// minimaxBaseURL resolves the managed backend base URL for the credential.
func minimaxBaseURL(provider string, auth *cliproxyauth.Auth) string {
	region := minimaxRegionFromProvider(provider)
	if auth != nil {
		if auth.Metadata != nil {
			if raw, ok := auth.Metadata["region"].(string); ok && strings.TrimSpace(raw) != "" {
				region = minimaxauth.NormalizeRegion(raw)
			}
		}
		if auth.Attributes != nil {
			if raw := strings.TrimSpace(auth.Attributes["minimax_region"]); raw != "" {
				region = minimaxauth.NormalizeRegion(raw)
			}
		}
	}
	if auth != nil {
		if auth.Attributes != nil {
			if raw := strings.TrimRight(strings.TrimSpace(auth.Attributes["base_url"]), "/"); raw != "" {
				return raw
			}
		}
		if auth.Metadata != nil {
			if raw, ok := auth.Metadata["base_url"].(string); ok && strings.TrimSpace(raw) != "" {
				return strings.TrimRight(strings.TrimSpace(raw), "/")
			}
		}
	}
	return region.InferenceBaseURL()
}

// minimaxMessagesURL builds the Anthropic Messages endpoint URL.
func minimaxMessagesURL(provider string, auth *cliproxyauth.Auth) string {
	return minimaxBaseURL(provider, auth) + "/v1/messages"
}

// minimaxCountTokensURL builds the Anthropic count_tokens endpoint URL.
func minimaxCountTokensURL(provider string, auth *cliproxyauth.Auth) string {
	return minimaxBaseURL(provider, auth) + "/v1/messages/count_tokens"
}

func minimaxAccessToken(auth *cliproxyauth.Auth) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if token := strings.TrimSpace(auth.Attributes["api_key"]); token != "" && !strings.HasPrefix(token, "sk-xxx") {
			return token
		}
		if token := strings.TrimSpace(auth.Attributes["access_token"]); token != "" {
			return token
		}
	}
	if auth.Metadata != nil {
		if token, ok := auth.Metadata["access_token"].(string); ok && strings.TrimSpace(token) != "" {
			return strings.TrimSpace(token)
		}
	}
	return ""
}

func minimaxMetadataString(auth *cliproxyauth.Auth, key string) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	value, _ := auth.Metadata[key].(string)
	return strings.TrimSpace(value)
}

// Execute performs a non-streaming Anthropic Messages request against the
// managed MiniMax backend.
func (e *MinimaxExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (resp cliproxyexecutor.Response, err error) {
	ctx = helps.EnsureSessionContext(ctx, opts, req.Payload)
	if opts.Alt == "responses/compact" {
		return resp, statusErr{code: http.StatusNotImplemented, msg: "/responses/compact not supported"}
	}
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	token := minimaxAccessToken(auth)

	reporter := helps.NewExecutorUsageReporter(ctx, e, baseModel, auth)
	defer reporter.TrackFailure(ctx, &err)

	from := opts.SourceFormat
	responseFormat := cliproxyexecutor.ResponseFormatOrSource(opts)
	to := sdktranslator.FormatClaude

	originalPayloadSource := req.Payload
	if len(opts.OriginalRequest) > 0 {
		originalPayloadSource = opts.OriginalRequest
	}
	originalPayload := originalPayloadSource

	// The managed backend is Anthropic Messages, so translation targets `claude`.
	// Non-claude callers need upstream streaming to translate the event stream.
	upstreamStream := responseFormat != to
	isCompat := helps.APIKeyModelIsCompat(req)
	originalTranslated, body := helps.TranslateRequestPairWithAPIKeyModelCompatibility(ctx, opts.Headers, e.cfg, from, to, baseModel, originalPayload, req.Payload, upstreamStream, isCompat)
	body = helps.SetStringIfDifferent(body, "model", baseModel)
	body, err = helps.ApplyRequestThinking(body, req, opts, from.String(), to.String(), e.Identifier())
	if err != nil {
		return resp, err
	}
	requestedModel := helps.PayloadRequestedModel(opts, req.Model)
	requestPath := helps.PayloadRequestPath(opts)
	body = helps.ApplyPayloadConfigWithRequest(e.cfg, baseModel, to.String(), from.String(), "", body, originalTranslated, requestedModel, requestPath, opts.Headers)
	body = ensureModelMaxTokens(body, baseModel)
	body = disableThinkingIfToolChoiceForced(body)
	if upstreamStream {
		body = helps.SetBoolIfDifferent(body, "stream", true)
	}

	url := minimaxMessagesURL(e.Identifier(), auth)
	httpReq, errReq := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if errReq != nil {
		return resp, errReq
	}
	if errHeaders := e.applyHeaders(httpReq, auth, token, upstreamStream); errHeaders != nil {
		return resp, errHeaders
	}
	e.recordRequest(ctx, url, httpReq.Header.Clone(), body, auth)

	httpClient := helps.NewProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	httpClient = reporter.TrackHTTPClient(httpClient)
	httpResp, errDo := httpClient.Do(httpReq)
	if errDo != nil {
		helps.RecordAPIResponseError(ctx, e.cfg, errDo)
		return resp, errDo
	}
	helps.RecordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(httpResp.Body, 8<<20))
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("minimax executor: close response body error: %v", errClose)
		}
		helps.AppendAPIResponseChunk(ctx, e.cfg, b)
		helps.LogWithRequestID(ctx).Debugf("request error, error status: %d, error message: %s", httpResp.StatusCode, helps.SummarizeErrorBody(httpResp.Header.Get("Content-Type"), b))
		return resp, classifyClaudeUpstreamErrorWithCooling(httpResp.StatusCode, httpResp.Header, b, false)
	}
	data, errRead := io.ReadAll(io.LimitReader(httpResp.Body, 64<<20))
	if errClose := httpResp.Body.Close(); errClose != nil {
		log.Errorf("minimax executor: close response body error: %v", errClose)
	}
	if errRead != nil {
		helps.RecordAPIResponseError(ctx, e.cfg, errRead)
		return resp, errRead
	}
	helps.AppendAPIResponseChunk(ctx, e.cfg, data)
	reporter.ObserveResponseModel(data)
	reporter.Publish(ctx, helps.ParseClaudeUsage(data))
	var param any
	out := sdktranslator.TranslateNonStream(ctx, to, responseFormat, req.Model, opts.OriginalRequest, body, data, &param)
	if responseFormat == sdktranslator.FormatOpenAIResponse {
		out = helps.EnsureResponsesUsageDetails(out)
	}
	return cliproxyexecutor.Response{Payload: out, Headers: httpResp.Header.Clone()}, nil
}

// ExecuteStream performs a streaming Anthropic Messages request against the
// managed MiniMax backend.
func (e *MinimaxExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (_ *cliproxyexecutor.StreamResult, err error) {
	ctx = helps.EnsureSessionContext(ctx, opts, req.Payload)
	if opts.Alt == "responses/compact" {
		return nil, statusErr{code: http.StatusNotImplemented, msg: "/responses/compact not supported"}
	}
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	token := minimaxAccessToken(auth)

	reporter := helps.NewExecutorUsageReporter(ctx, e, baseModel, auth)
	defer reporter.TrackFailure(ctx, &err)

	from := opts.SourceFormat
	responseFormat := cliproxyexecutor.ResponseFormatOrSource(opts)
	to := sdktranslator.FormatClaude

	originalPayloadSource := req.Payload
	if len(opts.OriginalRequest) > 0 {
		originalPayloadSource = opts.OriginalRequest
	}
	originalPayload := originalPayloadSource

	isCompat := helps.APIKeyModelIsCompat(req)
	originalTranslated, body := helps.TranslateRequestPairWithAPIKeyModelCompatibility(ctx, opts.Headers, e.cfg, from, to, baseModel, originalPayload, req.Payload, true, isCompat)
	body = helps.SetStringIfDifferent(body, "model", baseModel)
	body, err = helps.ApplyRequestThinking(body, req, opts, from.String(), to.String(), e.Identifier())
	if err != nil {
		return nil, err
	}
	requestedModel := helps.PayloadRequestedModel(opts, req.Model)
	requestPath := helps.PayloadRequestPath(opts)
	body = helps.ApplyPayloadConfigWithRequest(e.cfg, baseModel, to.String(), from.String(), "", body, originalTranslated, requestedModel, requestPath, opts.Headers)
	body = ensureModelMaxTokens(body, baseModel)
	body = disableThinkingIfToolChoiceForced(body)
	body = helps.SetBoolIfDifferent(body, "stream", true)

	url := minimaxMessagesURL(e.Identifier(), auth)
	httpReq, errReq := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if errReq != nil {
		return nil, errReq
	}
	if errHeaders := e.applyHeaders(httpReq, auth, token, true); errHeaders != nil {
		return nil, errHeaders
	}
	e.recordRequest(ctx, url, httpReq.Header.Clone(), body, auth)

	httpClient := helps.NewProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	httpClient = reporter.TrackHTTPClient(httpClient)
	httpResp, errDo := httpClient.Do(httpReq)
	if errDo != nil {
		helps.RecordAPIResponseError(ctx, e.cfg, errDo)
		return nil, errDo
	}
	helps.RecordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(httpResp.Body, 8<<20))
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("minimax executor: close response body error: %v", errClose)
		}
		helps.AppendAPIResponseChunk(ctx, e.cfg, b)
		helps.LogWithRequestID(ctx).Debugf("request error, error status: %d, error message: %s", httpResp.StatusCode, helps.SummarizeErrorBody(httpResp.Header.Get("Content-Type"), b))
		return nil, classifyClaudeUpstreamErrorWithCooling(httpResp.StatusCode, httpResp.Header, b, false)
	}

	out := make(chan cliproxyexecutor.StreamChunk, 1)
	go func() {
		defer close(out)
		defer func() {
			if errClose := httpResp.Body.Close(); errClose != nil {
				log.Errorf("minimax executor: close response body error: %v", errClose)
			}
		}()
		scanner := bufio.NewScanner(httpResp.Body)
		scanner.Buffer(nil, 52_428_800)
		var param any
		var streamUsage helps.StreamUsageBuffer
		defer streamUsage.Publish(ctx, reporter)
		for scanner.Scan() {
			line := scanner.Bytes()
			helps.AppendAPIResponseChunk(ctx, e.cfg, line)
			reporter.ObserveResponseModel(line)
			streamUsage.ObserveClaudeStream(line)
			chunks := sdktranslator.TranslateStream(ctx, to, responseFormat, req.Model, opts.OriginalRequest, body, bytes.Clone(line), &param)
			if responseFormat == sdktranslator.FormatOpenAIResponse {
				for i := range chunks {
					chunks[i] = helps.EnsureResponsesUsageDetails(chunks[i])
				}
			}
			for i := range chunks {
				select {
				case out <- cliproxyexecutor.StreamChunk{Payload: chunks[i]}:
				case <-ctx.Done():
					return
				}
			}
		}
		if errScan := scanner.Err(); errScan != nil {
			helps.RecordAPIResponseError(ctx, e.cfg, errScan)
			streamUsage.PublishFailure(ctx, reporter, errScan)
			select {
			case out <- cliproxyexecutor.StreamChunk{Err: errScan}:
			case <-ctx.Done():
			}
		}
	}()
	return &cliproxyexecutor.StreamResult{Headers: httpResp.Header.Clone(), Chunks: out}, nil
}

// CountTokens estimates input tokens for the request. The managed backend does
// implement /v1/messages/count_tokens, but the shared O200kBase estimator keeps
// token counting available even when the account lacks the endpoint.
func (e *MinimaxExecutor) CountTokens(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	from := opts.SourceFormat
	responseFormat := cliproxyexecutor.ResponseFormatOrSource(opts)
	to := sdktranslator.FormatClaude
	stream := from != to
	body := helps.TranslateRequestWithAPIKeyModelCompatibility(ctx, opts.Headers, e.cfg, from, to, baseModel, req.Payload, stream, helps.APIKeyModelIsCompat(req))
	var errThinking error
	body, errThinking = helps.ApplyRequestThinking(body, req, opts, from.String(), to.String(), e.Identifier())
	if errThinking != nil {
		return cliproxyexecutor.Response{}, errThinking
	}
	count, errCount := helps.CountClaudeInputTokens(body)
	if errCount != nil {
		return cliproxyexecutor.Response{}, fmt.Errorf("minimax executor: token counting failed: %w", errCount)
	}
	usageJSON := []byte(fmt.Sprintf(`{"input_tokens":%d}`, count))
	out := sdktranslator.TranslateTokenCount(ctx, to, responseFormat, count, usageJSON)
	return cliproxyexecutor.Response{Payload: out}, nil
}

// Refresh rotates the MiniMax OAuth credential using the stored refresh token.
func (e *MinimaxExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	log.Debugf("minimax executor: refresh called")
	if refreshed, handled, errHome := helps.RefreshAuthViaHome(ctx, e.cfg, auth); handled {
		return refreshed, errHome
	}
	if auth == nil {
		return nil, fmt.Errorf("minimax executor: auth is nil")
	}
	refreshToken := minimaxMetadataString(auth, "refresh_token")
	if refreshToken == "" {
		// Nothing to rotate; the access token is authoritative until it expires.
		return auth, nil
	}
	region := minimaxRegionFromProvider(e.Identifier())
	if raw := minimaxMetadataString(auth, "region"); raw != "" {
		region = minimaxauth.NormalizeRegion(raw)
	}
	token, errRefresh := minimaxauth.NewClientWithProxyURL(e.cfg, region, auth.ProxyURL).Refresh(ctx, refreshToken)
	if errRefresh != nil {
		return nil, errRefresh
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["type"] = e.Identifier()
	auth.Metadata["auth_kind"] = cliproxyauth.AuthKindOAuth
	auth.Metadata["access_token"] = token.AccessToken
	if strings.TrimSpace(token.RefreshToken) != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if strings.TrimSpace(token.TokenType) != "" {
		auth.Metadata["token_type"] = token.TokenType
	}
	if token.ExpiresIn > 0 {
		auth.Metadata["expires_in"] = token.ExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		auth.Metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if token.AccountID != "" {
		auth.Metadata["account_id"] = token.AccountID
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)
	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes[cliproxyauth.AttributeAuthKind] = cliproxyauth.AuthKindOAuth
	if strings.TrimSpace(auth.Attributes["base_url"]) == "" {
		auth.Attributes["base_url"] = region.InferenceBaseURL()
	}
	return auth, nil
}

// applyHeaders sets the Anthropic Messages and MiniMax credentials on the request.
func (e *MinimaxExecutor) applyHeaders(httpReq *http.Request, auth *cliproxyauth.Auth, token string, stream bool) error {
	if httpReq == nil {
		return nil
	}
	if strings.TrimSpace(token) != "" {
		httpReq.Header.Del("x-api-key")
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")
	httpReq.Header.Set("User-Agent", minimaxauth.UserAgent)
	accept := "application/json"
	if stream {
		accept = "text/event-stream"
	}
	httpReq.Header.Set("Accept", accept)
	var attrs map[string]string
	if auth != nil {
		attrs = auth.Attributes
	}
	util.ApplyCustomHeadersFromAttrs(httpReq, attrs)
	return nil
}

func (e *MinimaxExecutor) recordRequest(ctx context.Context, url string, headers http.Header, body []byte, auth *cliproxyauth.Auth) {
	var authID, authLabel, authType, authValue string
	if auth != nil {
		authID = auth.ID
		authLabel = auth.Label
		authType, authValue = auth.AccountInfo()
	}
	helps.RecordAPIRequest(ctx, e.cfg, helps.UpstreamRequestLog{
		URL:       url,
		Method:    http.MethodPost,
		Headers:   headers,
		Body:      body,
		Provider:  e.Identifier(),
		AuthID:    authID,
		AuthLabel: authLabel,
		AuthType:  authType,
		AuthValue: authValue,
	})
}

var _ cliproxyauth.ProviderExecutor = (*MinimaxExecutor)(nil)
