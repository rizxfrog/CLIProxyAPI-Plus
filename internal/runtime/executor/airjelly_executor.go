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

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/airjelly"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
)

var _ cliproxyauth.ProviderExecutor = (*AirJellyExecutor)(nil)

type AirJellyExecutor struct{ cfg *config.Config }

func NewAirJellyExecutor(cfg *config.Config) *AirJellyExecutor { return &AirJellyExecutor{cfg: cfg} }
func (e *AirJellyExecutor) Identifier() string                 { return airjelly.Provider }
func (e *AirJellyExecutor) RequestToFormat(_ cliproxyexecutor.Request, _ cliproxyexecutor.Options) sdktranslator.Format {
	return sdktranslator.FormatOpenAI
}

func airJellyToken(auth *cliproxyauth.Auth, key string) string {
	if auth == nil {
		return ""
	}
	if v := strings.TrimSpace(auth.Attributes[key]); v != "" {
		return v
	}
	if v, ok := auth.Metadata[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func (e *AirJellyExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	if req == nil {
		return nil
	}
	return airjelly.SignRequest(req, airJellyToken(auth, "access_token"))
}

func (e *AirJellyExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("airjelly: nil request")
	}
	if ctx == nil {
		ctx = req.Context()
	}
	req = req.WithContext(ctx)
	if err := e.PrepareRequest(req, auth); err != nil {
		return nil, err
	}
	return helps.NewProxyAwareHTTPClient(ctx, e.cfg, auth, 0).Do(req)
}

func (e *AirJellyExecutor) requestBody(ctx context.Context, req cliproxyexecutor.Request, opts cliproxyexecutor.Options, stream bool) ([]byte, error) {
	baseModel := thinking.ParseSuffix(req.Model).ModelName
	if baseModel != airjelly.Model {
		return nil, statusErr{code: 404, msg: "airjelly: model unavailable"}
	}
	from, to := opts.SourceFormat, sdktranslator.FormatOpenAI
	original := req.Payload
	if len(opts.OriginalRequest) > 0 {
		original = opts.OriginalRequest
	}
	translated, body := helps.TranslateRequestPairWithAPIKeyModelCompatibility(ctx, opts.Headers, e.cfg, from, to, baseModel, original, req.Payload, stream, helps.APIKeyModelIsCompat(req))
	body = helps.SetStringIfDifferent(body, "model", baseModel)
	var err error
	body, err = helps.ApplyRequestThinking(body, req, opts, from.String(), to.String(), e.Identifier())
	if err != nil {
		return nil, err
	}
	body = helps.ApplyPayloadConfigWithRequest(e.cfg, baseModel, to.String(), from.String(), "", body, translated, helps.PayloadRequestedModel(opts, req.Model), helps.PayloadRequestPath(opts), opts.Headers)
	// A payload override must never route this credential to a paid tier.
	body = helps.SetStringIfDifferent(body, "model", airjelly.Model)
	body = helps.SetBoolIfDifferent(body, "stream", stream)
	return body, nil
}

func (e *AirJellyExecutor) call(ctx context.Context, auth *cliproxyauth.Auth, body []byte, stream bool) (*http.Response, error) {
	url := airjelly.Gateway + "/api/v1/llm/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	if err = e.PrepareRequest(req, auth); err != nil {
		return nil, err
	}
	resp, err := helps.NewProxyAwareHTTPClient(ctx, e.cfg, auth, 0).Do(req)
	if err != nil {
		helps.RecordAPIResponseError(ctx, e.cfg, err)
		return nil, err
	}
	helps.RecordAPIResponseMetadata(ctx, e.cfg, resp.StatusCode, resp.Header.Clone())
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		helps.AppendAPIResponseChunk(ctx, e.cfg, b)
		return nil, statusErr{code: resp.StatusCode, msg: "airjelly: " + string(bytes.TrimSpace(b))}
	}
	return resp, nil
}

func (e *AirJellyExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (out cliproxyexecutor.Response, err error) {
	ctx = helps.EnsureSessionContext(ctx, opts, req.Payload)
	body, err := e.requestBody(ctx, req, opts, false)
	if err != nil {
		return out, err
	}
	model := thinking.ParseSuffix(req.Model).ModelName
	reporter := helps.NewExecutorUsageReporter(ctx, e, model, auth)
	defer reporter.TrackFailure(ctx, &err)
	resp, err := e.call(ctx, auth, body, false)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return out, err
	}
	helps.AppendAPIResponseChunk(ctx, e.cfg, b)
	if !json.Valid(b) || len(bytes.TrimSpace(b)) == 0 {
		return out, statusErr{code: 502, msg: "airjelly: invalid completion response"}
	}
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	_ = json.Unmarshal(b, &envelope)
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		return out, statusErr{code: 502, msg: "airjelly: " + string(envelope.Error)}
	}
	reporter.ObserveResponseModel(b)
	reporter.Publish(ctx, helps.ParseOpenAIUsage(b))
	format := cliproxyexecutor.ResponseFormatOrSource(opts)
	var param any
	translated := sdktranslator.TranslateNonStream(ctx, sdktranslator.FormatOpenAI, format, req.Model, opts.OriginalRequest, body, b, &param)
	if format == sdktranslator.FormatOpenAIResponse {
		translated = helps.EnsureResponsesUsageDetails(translated)
	}
	return cliproxyexecutor.Response{Payload: translated, Headers: resp.Header.Clone()}, nil
}

func (e *AirJellyExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (_ *cliproxyexecutor.StreamResult, err error) {
	ctx = helps.EnsureSessionContext(ctx, opts, req.Payload)
	body, err := e.requestBody(ctx, req, opts, true)
	if err != nil {
		return nil, err
	}
	model := thinking.ParseSuffix(req.Model).ModelName
	reporter := helps.NewExecutorUsageReporter(ctx, e, model, auth)
	defer reporter.TrackFailure(ctx, &err)
	resp, err := e.call(ctx, auth, body, true)
	if err != nil {
		return nil, err
	}
	out := make(chan cliproxyexecutor.StreamChunk, 1)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(nil, 52_428_800)
		var param any
		var usage helps.StreamUsageBuffer
		defer usage.Publish(ctx, reporter)
		format := cliproxyexecutor.ResponseFormatOrSource(opts)
		for scanner.Scan() {
			line := bytes.Clone(scanner.Bytes())
			helps.AppendAPIResponseChunk(ctx, e.cfg, line)
			reporter.ObserveResponseModel(line)
			usage.ObserveOpenAIStream(line)
			if bytes.HasPrefix(line, []byte("data: ")) {
				var event struct {
					Error json.RawMessage `json:"error"`
				}
				if json.Unmarshal(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data: "))), &event) == nil && len(event.Error) > 0 && string(event.Error) != "null" {
					e := statusErr{code: 502, msg: "airjelly: " + string(event.Error)}
					usage.PublishFailure(ctx, reporter, e)
					select {
					case out <- cliproxyexecutor.StreamChunk{Err: e}:
					case <-ctx.Done():
					}
					return
				}
			}
			chunks := sdktranslator.TranslateStream(ctx, sdktranslator.FormatOpenAI, format, req.Model, opts.OriginalRequest, body, line, &param)
			for _, chunk := range chunks {
				select {
				case out <- cliproxyexecutor.StreamChunk{Payload: chunk}:
				case <-ctx.Done():
					return
				}
			}
		}
		if scanErr := scanner.Err(); scanErr != nil {
			usage.PublishFailure(ctx, reporter, scanErr)
			select {
			case out <- cliproxyexecutor.StreamChunk{Err: scanErr}:
			case <-ctx.Done():
			}
		}
	}()
	return &cliproxyexecutor.StreamResult{Headers: resp.Header.Clone(), Chunks: out}, nil
}

func (e *AirJellyExecutor) CountTokens(ctx context.Context, _ *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	body, err := e.requestBody(ctx, req, opts, false)
	if err != nil {
		return cliproxyexecutor.Response{}, err
	}
	enc, err := helps.TokenizerForModel(airjelly.Model)
	if err != nil {
		return cliproxyexecutor.Response{}, err
	}
	count, err := helps.CountOpenAIChatTokens(enc, body)
	if err != nil {
		return cliproxyexecutor.Response{}, err
	}
	return cliproxyexecutor.Response{Payload: sdktranslator.TranslateTokenCount(ctx, sdktranslator.FormatOpenAI, cliproxyexecutor.ResponseFormatOrSource(opts), count, helps.BuildOpenAIUsageJSON(count))}, nil
}

func (e *AirJellyExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	if refreshed, handled, err := helps.RefreshAuthViaHome(ctx, e.cfg, auth); handled {
		return refreshed, err
	}
	if auth == nil {
		return nil, fmt.Errorf("airjelly: nil auth")
	}
	refresh := airJellyToken(auth, "refresh_token")
	if refresh == "" {
		return nil, fmt.Errorf("airjelly: refresh token missing")
	}
	refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	client := helps.NewProxyAwareHTTPClient(refreshCtx, e.cfg, auth, 30*time.Second)
	session, err := airjelly.Refresh(refreshCtx, client, refresh)
	if err != nil {
		return nil, err
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Metadata["access_token"] = session.AccessToken
	auth.Metadata["refresh_token"] = session.RefreshToken
	auth.Metadata["expired"] = time.Unix(session.ExpiresAt, 0).UTC().Format(time.RFC3339)
	auth.Attributes["access_token"] = session.AccessToken
	auth.Attributes["refresh_token"] = session.RefreshToken
	return auth, nil
}
