package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/qwenweb"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	execution "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/usage"
	translator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// QwenWebExecutor uses isolated temporary upstream chats. It deliberately does
// not emulate tool calling or share web conversation state between API callers.
type QwenWebExecutor struct {
	cfg *config.Config
	// baseURL overrides the upstream host in tests; empty means the real one.
	baseURL string
}

func NewQwenWebExecutor(cfg *config.Config) *QwenWebExecutor { return &QwenWebExecutor{cfg: cfg} }

// base returns the upstream base URL, limited to a test override or Qwen itself.
func (e *QwenWebExecutor) base() string {
	if e != nil && e.baseURL != "" {
		return e.baseURL
	}
	return qwenweb.BaseURL
}
func (e *QwenWebExecutor) Identifier() string { return qwenweb.Provider }
func (e *QwenWebExecutor) RequestToFormat(_ execution.Request, opts execution.Options) translator.Format {
	if opts.SourceFormat.String() == "openai-image" {
		return opts.SourceFormat
	}
	return translator.FormatOpenAI
}
func (e *QwenWebExecutor) PrepareRequest(req *http.Request, auth *coreauth.Auth) error {
	if req == nil {
		return fmt.Errorf("qwen-web: missing request")
	}
	// Never attach the session to another origin (including management APICall
	// URLs). The loopback override exists only for tests.
	if !e.allowsRequestURL(req.URL) {
		return &qwenweb.Error{Code: 400, Message: "Qwen Web credentials can only be used with chat.qwen.ai"}
	}
	cookie, token := "", ""
	if auth != nil {
		cookie, _ = auth.Metadata["cookie"].(string)
		token, _ = auth.Metadata["access_token"].(string)
	}
	cookie = strings.TrimSpace(cookie)
	token = strings.TrimSpace(token)
	if cookie == "" && token == "" {
		return &qwenweb.Error{Code: 401, Message: "Qwen Web session is missing; sign in again"}
	}
	if cookie != "" && strings.ContainsAny(cookie, "\r\n") {
		return &qwenweb.Error{Code: 401, Message: "Qwen Web session cookie is invalid; sign in again"}
	}
	if token != "" && strings.ContainsAny(token, "\r\n\t ") {
		return &qwenweb.Error{Code: 401, Message: "Qwen Web session is invalid; sign in again"}
	}
	req.Header = qwenweb.Headers(token, cookie)
	return nil
}
func (e *QwenWebExecutor) allowsRequestURL(u *url.URL) bool {
	if u == nil {
		return false
	}
	if e != nil && e.baseURL != "" {
		return u.String() == e.baseURL || strings.HasPrefix(u.String(), e.baseURL)
	}
	return u.Scheme == "https" && u.Host == "chat.qwen.ai"
}

func (e *QwenWebExecutor) HttpRequest(ctx context.Context, auth *coreauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("qwen-web: missing request")
	}
	req = req.Clone(ctx)
	if err := e.PrepareRequest(req, auth); err != nil {
		return nil, err
	}
	client := helps.NewProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return nil, &qwenweb.Error{Code: 502, Message: "Qwen Web connection failed"}
	}
	return resp, nil
}

// RefreshLead starts renewal a week before the 30-day session expires.
func (e *QwenWebExecutor) RefreshLead() *time.Duration {
	lead := 7 * 24 * time.Hour
	return &lead
}

// Refresh slides the Qwen Web session forward.
//
// Qwen issues no refresh_token and exposes no dedicated refresh endpoint:
// POST /api/v2/auths/refresh returns "not found" exactly like an arbitrary
// unknown path. Instead the server re-issues the session token Cookie on
// ordinary requests, and each re-issue sets exp = issue time + 30 days. Calling
// the session endpoint is therefore enough to slide an active session forward,
// matching what the real web frontend gets for free on every request.
//
// This cannot resurrect an expired session; once the deadline passes Qwen
// returns 401 and the account must sign in again.
func (e *QwenWebExecutor) Refresh(ctx context.Context, auth *coreauth.Auth) (*coreauth.Auth, error) {
	if auth == nil {
		return nil, fmt.Errorf("qwen-web: missing auth")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	updated := auth.Clone()
	if updated.Metadata == nil {
		updated.Metadata = map[string]any{}
	}
	resp, err := e.request(ctx, updated, http.MethodGet, "/api/v1/auths/", nil)
	if err != nil {
		return nil, err
	}
	defer closeQwenWebBody(resp.Body)
	if _, errRead := io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20)); errRead != nil {
		return nil, &qwenweb.Error{Code: 502, Message: "qwen-web: could not read session response"}
	}
	// The server re-issues the token on every call. Persist it only when it
	// actually changed, so routine refreshes avoid rewriting the auth file.
	if renewed := qwenweb.SessionCookie(resp.Cookies()); renewed != "" {
		current, _ := updated.Metadata["cookie"].(string)
		if renewed != current {
			updated.Metadata["cookie"] = renewed
		}
		// Keep access_token in sync so the bearer fallback and the JWT-derived
		// expiry stay accurate.
		if token := qwenweb.CookieValue(renewed); token != "" {
			if cur, _ := updated.Metadata["access_token"].(string); cur != token {
				updated.Metadata["access_token"] = token
			}
		}
	} else {
		// No new Cookie means the session was not recognized; surface that
		// instead of silently reporting success.
		return nil, &qwenweb.Error{Code: 401, Message: "qwen-web: session was not renewed; sign in again"}
	}
	return updated, nil
}
func (e *QwenWebExecutor) CountTokens(context.Context, *coreauth.Auth, execution.Request, execution.Options) (execution.Response, error) {
	return execution.Response{}, &qwenweb.Error{Code: 501, Message: "Qwen Web token counting is not supported"}
}
func (e *QwenWebExecutor) request(ctx context.Context, auth *coreauth.Auth, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, &qwenweb.Error{Code: 400, Message: "invalid Qwen Web request"}
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, e.base()+path, reader)
	if err != nil {
		return nil, &qwenweb.Error{Code: 400, Message: "invalid Qwen Web request"}
	}
	resp, err := e.HttpRequest(ctx, auth, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		closeQwenWebBody(resp.Body)
		code := resp.StatusCode
		if code < 400 {
			code = 502
		}
		return nil, &qwenweb.Error{Code: code, Message: fmt.Sprintf("Qwen Web rejected the request (HTTP %d)", resp.StatusCode)}
	}
	return resp, nil
}
func closeQwenWebBody(body io.Closer) {
	if err := body.Close(); err != nil {
		log.Debug("qwen-web: response close failed")
	}
}
func (e *QwenWebExecutor) deleteChat(ctx context.Context, auth *coreauth.Auth, id string) {
	if resp, err := e.request(ctx, auth, http.MethodDelete, "/api/v2/chats/"+url.PathEscape(id), nil); err == nil {
		closeQwenWebBody(resp.Body)
	} else {
		log.Debug("qwen-web: temporary chat cleanup failed")
	}
}

// prepare checks unsupported features before creating any upstream conversation.
func (e *QwenWebExecutor) prepare(ctx context.Context, req execution.Request, opts execution.Options) ([]byte, string, bool, string, error) {
	for _, raw := range [][]byte{req.Payload, opts.OriginalRequest} {
		if err := helps.QwenWebRejectUnsupported(raw); err != nil {
			return nil, "", false, "", err
		}
	}
	image := opts.SourceFormat.String() == "openai-image"
	if image {
		if req.Model != "qwen-web-image" {
			return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "use qwen-web-image for image generation"}
		}
		body := req.Payload
		prompt := strings.TrimSpace(gjson.GetBytes(body, "prompt").String())
		if prompt == "" {
			return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "image prompt is required"}
		}
		if opts.Stream || gjson.GetBytes(body, "stream").Bool() {
			return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "Qwen Web images do not support streaming"}
		}
		if n := gjson.GetBytes(body, "n"); n.Exists() && n.Int() != 1 {
			return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "Qwen Web supports n=1 only"}
		}
		format := gjson.GetBytes(body, "response_format").String()
		if format != "" && format != "url" {
			return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "Qwen Web images support response_format=url only"}
		}
		for _, key := range []string{"image", "images", "mask", "quality", "background", "output_format", "output_compression", "partial_images"} {
			if gjson.GetBytes(body, key).Exists() {
				return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "unsupported Qwen Web image option: " + key}
			}
		}
		ratio := "1:1"
		switch gjson.GetBytes(body, "size").String() {
		case "", "auto", "1024x1024", "1:1":
		case "1536x1024", "3:2":
			ratio = "3:2"
		case "1024x1536", "2:3":
			ratio = "2:3"
		case "16:9", "9:16":
			ratio = gjson.GetBytes(body, "size").String()
		default:
			return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "unsupported Qwen Web image size"}
		}
		if opts.Metadata != nil {
			if path, _ := opts.Metadata[execution.RequestPathMetadataKey].(string); path != "" && !strings.HasSuffix(path, "/images/generations") {
				return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "Qwen Web supports image generation, not editing"}
			}
		}
		return body, prompt, true, ratio, nil
	}
	if req.Model != "qwen-web-chat" {
		return nil, "", false, "", &qwenweb.Error{Code: 400, Message: "use qwen-web-chat for text chat"}
	}
	body := helps.TranslateRequestWithCodexMultiAgentV2(ctx, opts.Headers, e.cfg, opts.SourceFormat, translator.FormatOpenAI, req.Model, bytes.Clone(req.Payload), opts.Stream)
	if err := helps.QwenWebRejectUnsupported(body); err != nil {
		return nil, "", false, "", err
	}
	prompt, err := helps.QwenWebPrompt(body)
	return body, prompt, false, "", err
}
func (e *QwenWebExecutor) start(ctx context.Context, auth *coreauth.Auth, prompt string, image bool, ratio string) (string, *http.Response, error) {
	kind := "t2t"
	if image {
		kind = "t2i"
	}
	resp, err := e.request(ctx, auth, http.MethodPost, "/api/v2/chats/new", map[string]any{"title": "API chat", "models": []string{helps.QwenWebUpstreamModel}, "chat_mode": "normal", "chat_type": kind, "timestamp": time.Now().Unix()})
	if err != nil {
		return "", nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	closeQwenWebBody(resp.Body)
	if err != nil {
		return "", nil, &qwenweb.Error{Code: 502, Message: "could not read Qwen Web chat creation"}
	}
	id := gjson.GetBytes(raw, "data.id").String()
	if !gjson.GetBytes(raw, "success").Bool() || id == "" {
		return "", nil, &qwenweb.Error{Code: 502, Message: "Qwen Web did not create a conversation"}
	}
	resp, err = e.request(ctx, auth, http.MethodPost, "/api/v2/chat/completions?chat_id="+url.QueryEscape(id), helps.QwenWebPayload(id, prompt, image, ratio))
	if err != nil {
		e.deleteChat(ctx, auth, id)
		return "", nil, err
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		closeQwenWebBody(resp.Body)
		e.deleteChat(ctx, auth, id)
		return "", nil, &qwenweb.Error{Code: 502, Message: "Qwen Web did not return an event stream"}
	}
	return id, resp, nil
}
func (e *QwenWebExecutor) Execute(ctx context.Context, auth *coreauth.Auth, req execution.Request, opts execution.Options) (result execution.Response, err error) {
	reporter := helps.NewExecutorUsageReporter(ctx, e, req.Model, auth)
	defer reporter.TrackFailure(ctx, &err)
	body, prompt, image, ratio, err := e.prepare(ctx, req, opts)
	if err != nil {
		return result, err
	}
	id, resp, err := e.start(ctx, auth, prompt, image, ratio)
	if err != nil {
		return result, err
	}
	defer e.deleteChat(ctx, auth, id)
	defer closeQwenWebBody(resp.Body)
	var text, reasoning strings.Builder
	finishReason := "stop"
	urls := []string{}
	seen := map[string]bool{}
	appendURLs := func(raw []byte) {
		for _, u := range helps.QwenWebImages(raw) {
			if !seen[u] {
				seen[u] = true
				urls = append(urls, u)
			}
		}
	}
	err = helps.ConsumeQwenWebSSE(resp.Body, func(delta helps.QwenWebDelta) error {
		if delta.FinishReason == "length" {
			finishReason = "length"
		}
		reporter.ObserveTokenEvent(delta.Text != "" || delta.Reasoning != "")
		if text.Len()+reasoning.Len()+len(delta.Text)+len(delta.Reasoning) > 16<<20 {
			return &qwenweb.Error{Code: 502, Message: "Qwen Web response exceeds limit"}
		}
		text.WriteString(delta.Text)
		reasoning.WriteString(delta.Reasoning)
		if image {
			appendURLs(delta.Raw)
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if image {
		if len(urls) == 0 {
			detail, errDetail := e.request(ctx, auth, http.MethodGet, "/api/v2/chats/"+url.PathEscape(id), nil)
			if errDetail != nil {
				return result, errDetail
			}
			raw, errRead := io.ReadAll(io.LimitReader(detail.Body, 4<<20))
			closeQwenWebBody(detail.Body)
			if errRead != nil {
				return result, &qwenweb.Error{Code: 502, Message: "could not read Qwen Web image result"}
			}
			for _, u := range helps.QwenWebDetailImages(raw) {
				if !seen[u] {
					seen[u] = true
					urls = append(urls, u)
				}
			}
		}
		if len(urls) == 0 {
			return result, &qwenweb.Error{Code: 502, Message: "Qwen Web returned no image URL"}
		}
		data := []map[string]string{}
		for _, u := range urls {
			data = append(data, map[string]string{"url": u})
		}
		raw, _ := json.Marshal(map[string]any{"created": time.Now().Unix(), "data": data})
		reporter.Publish(ctx, usage.Detail{})
		return execution.Response{Payload: raw}, nil
	}
	if text.Len() == 0 && reasoning.Len() == 0 {
		return result, &qwenweb.Error{Code: 502, Message: "Qwen Web returned an empty answer"}
	}
	message := map[string]any{"role": "assistant", "content": text.String()}
	if reasoning.Len() > 0 {
		message["reasoning_content"] = reasoning.String()
	}
	raw, _ := json.Marshal(map[string]any{"id": "chatcmpl-" + uuid.NewString(), "object": "chat.completion", "created": time.Now().Unix(), "model": req.Model, "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": finishReason}}})
	var param any
	raw = translator.TranslateNonStream(ctx, translator.FormatOpenAI, execution.ResponseFormatOrSource(opts), req.Model, opts.OriginalRequest, body, raw, &param)
	reporter.Publish(ctx, usage.Detail{})
	return execution.Response{Payload: raw}, nil
}
func (e *QwenWebExecutor) ExecuteStream(ctx context.Context, auth *coreauth.Auth, req execution.Request, opts execution.Options) (*execution.StreamResult, error) {
	reporter := helps.NewExecutorUsageReporter(ctx, e, req.Model, auth)
	body, prompt, image, ratio, err := e.prepare(ctx, req, opts)
	if err != nil {
		reporter.PublishFailure(ctx, err)
		return nil, err
	}
	if image {
		return nil, &qwenweb.Error{Code: 400, Message: "Qwen Web images do not support streaming"}
	}
	id, resp, err := e.start(ctx, auth, prompt, false, ratio)
	if err != nil {
		reporter.PublishFailure(ctx, err)
		return nil, err
	}
	out := make(chan execution.StreamChunk)
	go func() {
		defer close(out)
		defer e.deleteChat(ctx, auth, id)
		defer closeQwenWebBody(resp.Body)
		var param any
		completionID := "chatcmpl-" + uuid.NewString()
		created := time.Now().Unix()
		visible := false
		finishReason := "stop"
		send := func(payload []byte, sendErr error) error {
			select {
			case out <- execution.StreamChunk{Payload: payload, Err: sendErr}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		emit := func(delta map[string]any, finish any) error {
			raw, _ := json.Marshal(map[string]any{"id": completionID, "object": "chat.completion.chunk", "created": created, "model": req.Model, "choices": []any{map[string]any{"index": 0, "delta": delta, "finish_reason": finish}}})
			raw = append([]byte("data: "), raw...)
			for _, chunk := range translator.TranslateStream(ctx, translator.FormatOpenAI, execution.ResponseFormatOrSource(opts), req.Model, opts.OriginalRequest, body, raw, &param) {
				if errSend := send(chunk, nil); errSend != nil {
					return errSend
				}
			}
			return nil
		}
		errStream := helps.ConsumeQwenWebSSE(resp.Body, func(delta helps.QwenWebDelta) error {
			if delta.FinishReason == "length" {
				finishReason = "length"
			}
			if delta.Text == "" && delta.Reasoning == "" {
				return nil
			}
			reporter.ObserveTokenEvent(true)
			data := map[string]any{}
			if !visible {
				data["role"] = "assistant"
				visible = true
			}
			if delta.Text != "" {
				data["content"] = delta.Text
			}
			if delta.Reasoning != "" {
				data["reasoning_content"] = delta.Reasoning
			}
			return emit(data, nil)
		})
		if errStream == nil && !visible {
			errStream = &qwenweb.Error{Code: 502, Message: "Qwen Web returned an empty answer"}
		}
		if errStream == nil {
			errStream = emit(map[string]any{}, finishReason)
		}
		if errStream != nil {
			reporter.PublishFailure(ctx, errStream)
			_ = send(nil, errStream)
			return
		}
		for _, chunk := range translator.TranslateStream(ctx, translator.FormatOpenAI, execution.ResponseFormatOrSource(opts), req.Model, opts.OriginalRequest, body, []byte("data: [DONE]"), &param) {
			if errSend := send(chunk, nil); errSend != nil {
				reporter.PublishFailure(ctx, errSend)
				return
			}
		}
		reporter.Publish(ctx, usage.Detail{})
	}()
	return &execution.StreamResult{Chunks: out}, nil
}
