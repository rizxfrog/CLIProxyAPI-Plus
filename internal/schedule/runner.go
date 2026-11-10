package schedule

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// responseSnippetBytes bounds how much of a response body is kept for display.
// A scheduled task's response is evidence, not a payload to be replayed.
const responseSnippetBytes = 2 * 1024

// ignoredHeaders are dropped from a stored header set. They either describe one
// specific connection (Host, Content-Length, Connection) or are managed by the
// transport, and injecting them from a stored document would corrupt the request
// rather than reproduce it.
var ignoredHeaders = map[string]struct{}{
	"host":              {},
	"content-length":    {},
	"connection":        {},
	"keep-alive":        {},
	"transfer-encoding": {},
	"upgrade":           {},
	"te":                {},
	"trailer":           {},
}

// executor performs task requests.
//
// A per-request timeout is set here even though upstream LLM traffic in this
// codebase is deliberately timeout-free after a connection is established: a
// scheduled task is a discrete operator-configured HTTP call with an explicit
// timeout_ms, not a request being proxied on behalf of a client, and without a
// bound one unreachable host would occupy a worker forever.
type executor struct {
	follow   *http.Client
	noFollow *http.Client
}

func newExecutor() *executor {
	return &executor{
		follow: &http.Client{},
		noFollow: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// buildRequest turns a stored request into an outgoing one.
func buildRequest(ctx context.Context, request Request) (*http.Request, error) {
	var body io.Reader
	contentType := ""
	switch {
	case len(request.Form) > 0:
		values := url.Values{}
		for _, field := range request.Form {
			values.Add(field.Name, field.Value)
		}
		body = strings.NewReader(values.Encode())
		contentType = "application/x-www-form-urlencoded"
	case request.Body != "":
		body = strings.NewReader(request.Body)
	}
	outgoing, errNew := http.NewRequestWithContext(ctx, request.Method, request.URL, body)
	if errNew != nil {
		return nil, fmt.Errorf("schedule: build request: %w", errNew)
	}
	for name, value := range request.Headers {
		if _, ignored := ignoredHeaders[strings.ToLower(strings.TrimSpace(name))]; ignored {
			continue
		}
		outgoing.Header.Set(name, value)
	}
	if contentType != "" && outgoing.Header.Get("Content-Type") == "" {
		outgoing.Header.Set("Content-Type", contentType)
	}
	return outgoing, nil
}

// run performs one task's request. It never returns an error: every outcome,
// including a transport failure, is recorded as a run entry so the page can show
// what happened instead of only that something did.
func (e *executor) run(ctx context.Context, task Task) RunEntry {
	entry := RunEntry{StartedAt: time.Now().UTC(), Status: StatusFailed}
	timeout := time.Duration(task.Request.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = time.Duration(DefaultTimeoutMS) * time.Millisecond
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	finish := func() RunEntry {
		entry.FinishedAt = time.Now().UTC()
		entry.DurationMS = entry.FinishedAt.Sub(entry.StartedAt).Milliseconds()
		return entry
	}

	outgoing, errBuild := buildRequest(runCtx, task.Request)
	if errBuild != nil {
		entry.Error = errBuild.Error()
		return finish()
	}
	client := e.noFollow
	if task.Request.FollowRedirects {
		client = e.follow
	}
	response, errDo := client.Do(outgoing)
	if errDo != nil {
		entry.Error = describeTransportError(errDo)
		return finish()
	}
	snippet, _ := io.ReadAll(io.LimitReader(response.Body, responseSnippetBytes))
	if errClose := response.Body.Close(); errClose != nil {
		// A close error after a complete read cannot change the verdict.
		entry.Error = strings.TrimSpace(entry.Error)
	}
	entry.StatusCode = response.StatusCode
	entry.ResponseSnippet = string(snippet)
	if response.StatusCode >= http.StatusBadRequest {
		entry.Status = StatusFailed
		entry.Error = fmt.Sprintf("HTTP %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	} else {
		entry.Status = StatusSuccess
	}
	return finish()
}

// describeTransportError reports a transport failure without leaking credentials:
// a *url.Error stringifies the full URL, which may carry a token in its query.
func describeTransportError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		if errors.Is(urlErr.Err, context.DeadlineExceeded) {
			return "request timed out"
		}
		if errors.Is(urlErr.Err, context.Canceled) {
			return "request canceled"
		}
		return urlErr.Err.Error()
	}
	if err == nil {
		return ""
	}
	return err.Error()
}
