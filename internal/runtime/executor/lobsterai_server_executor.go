// Package executor provides per-provider runtime executors.
//
// This file implements the LobsterAI (NetEase/Youdao) managed-account executor.
// LobsterAI authenticates a user's subscription against its own backend
// (https://lobsterai-server.youdao.com) and forwards inference to that origin's
// managed gateway
//
//	{base}/api/proxy/v1/chat/completions
//
// with `Authorization: Bearer <access_token>` plus the client-identification
// headers the desktop app always sends.
//
// Three measured upstream behaviours shape this executor (all confirmed against
// the live gateway):
//
//   - Only the OpenAI Chat Completions surface exists. /messages answers every
//     model with code 40300 ("不支持的模型") and /responses is 404, so the
//     Anthropic- and Responses-shaped frontends are served by translating to and
//     from OpenAI rather than by a passthrough.
//   - The gateway always answers with SSE, even for stream:false. Execute drives
//     the streaming path and folds the chunks back into a chat.completion.
//   - Application-level failures (unsupported model 40300, inaccessible model
//     40301) arrive as an SSE `event:error` frame inside an HTTP 200, so the
//     stream is inspected for an in-band error before it is published.
//
// The provider identity ("lobsterai-server") stays separate from the wire
// format, so CLIProxyAPI reuses the existing OpenAI translators.
package executor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	log "github.com/sirupsen/logrus"
)

// compile-time assertion that the executor satisfies the full ProviderExecutor
// contract.
var _ cliproxyauth.ProviderExecutor = (*LobsteraiServerExecutor)(nil)

// LobsteraiServerExecutor proxies requests to the LobsterAI managed gateway.
type LobsteraiServerExecutor struct {
	cfg *config.Config
}

// NewLobsteraiServerExecutor creates the LobsterAI executor.
func NewLobsteraiServerExecutor(cfg *config.Config) *LobsteraiServerExecutor {
	return &LobsteraiServerExecutor{cfg: cfg}
}

// Identifier returns the executor/provider identifier.
func (e *LobsteraiServerExecutor) Identifier() string { return constant.LobsteraiServer }

// RequestToFormat reports the upstream wire format. The gateway implements only
// OpenAI Chat Completions, so every request is translated to OpenAI regardless
// of the model.
func (e *LobsteraiServerExecutor) RequestToFormat(_ cliproxyexecutor.Request, _ cliproxyexecutor.Options) sdktranslator.Format {
	return sdktranslator.FormatOpenAI
}

// PrepareRequest injects LobsterAI credentials into an outgoing HTTP request.
func (e *LobsteraiServerExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	if req == nil {
		return nil
	}
	token := lobsteraiServerCredential(auth)
	if strings.TrimSpace(token) != "" {
		req.Header.Del("x-api-key")
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-LobsterAI-Client-Capabilities", lobsteraiserverauth.ClientCapabilities)
	req.Header.Set("X-LobsterAI-Client-Version", lobsteraiServerClientVersion(auth))
	req.Header.Set("User-Agent", lobsteraiserverauth.UserAgent)
	lobsteraiServerApplyEnterpriseHeaders(req, auth)
	var attrs map[string]string
	if auth != nil {
		attrs = auth.Attributes
	}
	util.ApplyCustomHeadersFromAttrs(req, attrs)
	return nil
}

// HttpRequest injects LobsterAI credentials into the request and executes it.
func (e *LobsteraiServerExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("lobsterai-server executor: request is nil")
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

// Execute performs a non-streaming request.
//
// The gateway only implements SSE, so a non-streaming answer is produced from
// the streamed body. How depends on the client's shape: an OpenAI client gets
// the chunks folded into a chat.completion, while a client that speaks another
// protocol (Anthropic Messages, OpenAI Responses) gets the whole SSE body
// translated into its own non-stream payload.
func (e *LobsteraiServerExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (resp cliproxyexecutor.Response, err error) {
	responseFormat := cliproxyexecutor.ResponseFormatOrSource(opts)
	if responseFormat == sdktranslator.FormatOpenAI {
		return e.executeFoldOpenAI(ctx, auth, req, opts)
	}
	return e.executeNonStream(ctx, auth, req, opts, responseFormat)
}

// executeFoldOpenAI streams and folds the OpenAI chunks into a chat.completion.
func (e *LobsteraiServerExecutor) executeFoldOpenAI(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
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
	payload := lobsteraiServerAggregateStream(buffer.Bytes())
	if len(bytes.TrimSpace(payload)) == 0 {
		return cliproxyexecutor.Response{}, statusErr{code: http.StatusBadGateway, msg: "lobsterai-server: empty upstream response"}
	}
	return cliproxyexecutor.Response{Payload: payload, Headers: streamResult.Headers}, nil
}

// executeNonStream translates the gateway's streamed body into the client's
// non-stream format.
//
// The body is folded into an OpenAI chat.completion first, then the shared
// non-stream translator converts it into the client's shape (Anthropic Messages,
// OpenAI Responses, or a passthrough for OpenAI). Feeding the raw SSE frames to
// the translator is not possible: its non-stream entry point expects a bare JSON
// object, while the stream entry point only accepts a single "data:" line.
func (e *LobsteraiServerExecutor) executeNonStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options, responseFormat sdktranslator.Format) (resp cliproxyexecutor.Response, err error) {
	ctx = helps.EnsureSessionContext(ctx, opts, req.Payload)
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	to := sdktranslator.FormatOpenAI

	reporter := helps.NewExecutorUsageReporter(ctx, e, baseModel, auth)
	defer reporter.TrackFailure(ctx, &err)

	from := opts.SourceFormat
	originalPayloadSource := req.Payload
	if len(opts.OriginalRequest) > 0 {
		originalPayloadSource = opts.OriginalRequest
	}
	originalPayload := originalPayloadSource

	isCompat := helps.APIKeyModelIsCompat(req)
	originalTranslated, body := helps.TranslateRequestPairWithAPIKeyModelCompatibility(ctx, opts.Headers, e.cfg, from, to, baseModel, originalPayload, req.Payload, false, isCompat)
	body = helps.SetStringIfDifferent(body, "model", baseModel)
	body, err = helps.ApplyRequestThinking(body, req, opts, from.String(), to.String(), e.Identifier())
	if err != nil {
		return resp, err
	}
	body = helps.ApplyPayloadConfigWithRequest(e.cfg, baseModel, to.String(), from.String(), "", body, originalTranslated, helps.PayloadRequestedModel(opts, req.Model), helps.PayloadRequestPath(opts), opts.Headers)
	body = lobsteraiServerForceStream(body)

	url := lobsteraiServerChatURL(auth)
	httpReq, errReq := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if errReq != nil {
		return resp, errReq
	}
	if errHeaders := e.applyHeaders(httpReq, auth, true); errHeaders != nil {
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
	defer func() {
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("lobsterai-server executor: close response body error: %v", errClose)
		}
	}()
	helps.RecordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())
	data, errRead := io.ReadAll(io.LimitReader(httpResp.Body, 64<<20))
	if errRead != nil {
		helps.RecordAPIResponseError(ctx, e.cfg, errRead)
		return resp, errRead
	}
	helps.AppendAPIResponseChunk(ctx, e.cfg, data)
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return resp, lobsteraiServerUpstreamError(httpResp.StatusCode, httpResp.Header, data)
	}
	// The gateway reports an unsupported/inaccessible model inside HTTP 200.
	if status, message, isError := lobsteraiServerStreamErrorPayload(data); isError {
		return resp, statusErr{code: status, msg: message}
	}

	completion := lobsteraiServerSSEToCompletion(data)
	if len(completion) == 0 {
		// A plain JSON body (an ad-hoc or future non-stream response) is already
		// a chat.completion and passes through.
		if trimmed := bytes.TrimSpace(data); bytes.HasPrefix(trimmed, []byte("{")) && json.Valid(trimmed) {
			completion = trimmed
		}
	}
	if len(completion) == 0 {
		return resp, statusErr{code: http.StatusBadGateway, msg: "lobsterai-server: empty upstream response"}
	}
	reporter.ObserveResponseModel(completion)
	reporter.Publish(ctx, helps.ParseOpenAIUsage(completion))

	var param any
	out := sdktranslator.TranslateNonStream(ctx, to, responseFormat, req.Model, opts.OriginalRequest, body, completion, &param)
	if responseFormat == sdktranslator.FormatOpenAIResponse {
		out = helps.EnsureResponsesUsageDetails(out)
	}
	return cliproxyexecutor.Response{Payload: out, Headers: httpResp.Header.Clone()}, nil
}

// ExecuteStream performs a streaming request against the managed gateway.
func (e *LobsteraiServerExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (_ *cliproxyexecutor.StreamResult, err error) {
	ctx = helps.EnsureSessionContext(ctx, opts, req.Payload)
	if opts.Alt == "responses/compact" {
		return nil, statusErr{code: http.StatusNotImplemented, msg: "/responses/compact not supported"}
	}
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	// The gateway only speaks OpenAI Chat Completions, so every frontend is
	// translated to that format and the upstream call is always streamed.
	to := sdktranslator.FormatOpenAI

	reporter := helps.NewExecutorUsageReporter(ctx, e, baseModel, auth)
	defer reporter.TrackFailure(ctx, &err)

	from := opts.SourceFormat
	responseFormat := cliproxyexecutor.ResponseFormatOrSource(opts)

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
	body = lobsteraiServerForceStream(body)

	url := lobsteraiServerChatURL(auth)
	httpReq, errReq := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if errReq != nil {
		return nil, errReq
	}
	if errHeaders := e.applyHeaders(httpReq, auth, true); errHeaders != nil {
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
			log.Errorf("lobsterai-server executor: close response body error: %v", errClose)
		}
		helps.AppendAPIResponseChunk(ctx, e.cfg, b)
		helps.LogWithRequestID(ctx).Debugf("request error, error status: %d, error message: %s", httpResp.StatusCode, helps.SummarizeErrorBody(httpResp.Header.Get("Content-Type"), b))
		return nil, lobsteraiServerUpstreamError(httpResp.StatusCode, httpResp.Header, b)
	}

	out := make(chan cliproxyexecutor.StreamChunk, 1)
	go func() {
		defer close(out)
		defer func() {
			if errClose := httpResp.Body.Close(); errClose != nil {
				log.Errorf("lobsterai-server executor: close response body error: %v", errClose)
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
			streamUsage.ObserveOpenAIStream(line)
			// The gateway reports an unsupported (40300) or inaccessible (40301)
			// model as an event:error frame inside HTTP 200. Publishing that as a
			// normal chunk would surface an empty completion, so the error wins.
			if status, message, isError := lobsteraiServerStreamErrorPayload(line); isError {
				streamUsage.PublishFailure(ctx, reporter, statusErr{code: status, msg: message})
				select {
				case out <- cliproxyexecutor.StreamChunk{Err: statusErr{code: status, msg: message}}:
				case <-ctx.Done():
				}
				return
			}
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

// CountTokens estimates input tokens for the request with the OpenAI estimator,
// so token counting stays available without an upstream round trip.
func (e *LobsteraiServerExecutor) CountTokens(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	to := sdktranslator.FormatOpenAI

	from := opts.SourceFormat
	responseFormat := cliproxyexecutor.ResponseFormatOrSource(opts)
	stream := from != to
	body := helps.TranslateRequestWithAPIKeyModelCompatibility(ctx, opts.Headers, e.cfg, from, to, baseModel, req.Payload, stream, helps.APIKeyModelIsCompat(req))
	var errThinking error
	body, errThinking = helps.ApplyRequestThinking(body, req, opts, from.String(), to.String(), e.Identifier())
	if errThinking != nil {
		return cliproxyexecutor.Response{}, errThinking
	}

	enc, errEnc := helps.TokenizerForModel(baseModel)
	if errEnc != nil {
		return cliproxyexecutor.Response{}, fmt.Errorf("lobsterai-server executor: tokenizer init failed: %w", errEnc)
	}
	count, errCount := helps.CountOpenAIChatTokens(enc, body)
	if errCount != nil {
		return cliproxyexecutor.Response{}, fmt.Errorf("lobsterai-server executor: token counting failed: %w", errCount)
	}
	out := sdktranslator.TranslateTokenCount(ctx, to, responseFormat, count, helps.BuildOpenAIUsageJSON(count))
	return cliproxyexecutor.Response{Payload: out}, nil
}

// Refresh rotates the LobsterAI credential against /api/auth/refresh. A refresh
// response that omits a replacement refresh token keeps the stored one.
func (e *LobsteraiServerExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	log.Debugf("lobsterai-server executor: refresh called")
	if refreshed, handled, errHome := helps.RefreshAuthViaHome(ctx, e.cfg, auth); handled {
		return refreshed, errHome
	}
	if auth == nil {
		return nil, fmt.Errorf("lobsterai-server executor: auth is nil")
	}
	refreshToken := lobsteraiServerAttributeString(auth, "refresh_token")
	if refreshToken == "" {
		// Nothing to rotate; the access token is authoritative until it expires.
		return auth, nil
	}
	client := lobsteraiserverauth.NewClientWithProxyURL(e.cfg, auth.ProxyURL, lobsteraiServerBackendURL(auth))
	token, errRefresh := client.Refresh(ctx, refreshToken)
	if errRefresh != nil {
		return nil, errRefresh
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["type"] = lobsteraiserverauth.Provider
	auth.Metadata["auth_kind"] = cliproxyauth.AuthKindOAuth
	auth.Metadata["access_token"] = token.AccessToken
	if strings.TrimSpace(token.RefreshToken) != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		auth.Metadata["expires_in"] = token.ExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		auth.Metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)
	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes[cliproxyauth.AttributeAuthKind] = cliproxyauth.AuthKindOAuth
	auth.Attributes["access_token"] = token.AccessToken
	if strings.TrimSpace(auth.Attributes["base_url"]) == "" {
		auth.Attributes["base_url"] = lobsteraiserverauth.DefaultProxyBaseURL
	}
	return auth, nil
}

// applyHeaders sets the LobsterAI credentials and client-identification headers.
func (e *LobsteraiServerExecutor) applyHeaders(httpReq *http.Request, auth *cliproxyauth.Auth, stream bool) error {
	if httpReq == nil {
		return nil
	}
	token := lobsteraiServerCredential(auth)
	if strings.TrimSpace(token) != "" {
		httpReq.Header.Del("x-api-key")
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LobsterAI-Client-Capabilities", lobsteraiserverauth.ClientCapabilities)
	httpReq.Header.Set("X-LobsterAI-Client-Version", lobsteraiServerClientVersion(auth))
	httpReq.Header.Set("User-Agent", lobsteraiserverauth.UserAgent)
	lobsteraiServerApplyEnterpriseHeaders(httpReq, auth)
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

func (e *LobsteraiServerExecutor) recordRequest(ctx context.Context, url string, headers http.Header, body []byte, auth *cliproxyauth.Auth) {
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
