package schedule

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type capturedRequest struct {
	method      string
	path        string
	contentType string
	auth        string
	host        string
	body        string
}

func TestExecutorSendsStoredRequestAndDropsConnectionHeaders(t *testing.T) {
	received := make(chan capturedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 0)
		if r.Body != nil {
			read, _ := io.ReadAll(r.Body)
			body = read
		}
		received <- capturedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			contentType: r.Header.Get("Content-Type"),
			auth:        r.Header.Get("Authorization"),
			host:        r.Host,
			body:        string(body),
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	entry := newExecutor().run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{
			Method: "POST",
			URL:    server.URL + "/hook",
			Headers: map[string]string{
				"Authorization":  "Bearer secret",
				"Host":           "attacker.example.com",
				"Content-Length": "999",
				"Connection":     "close",
			},
			Body:      `{"a":1}`,
			TimeoutMS: 5000,
		},
	})
	if entry.Status != StatusSuccess {
		t.Fatalf("Status = %q (%s), want success", entry.Status, entry.Error)
	}
	if entry.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", entry.StatusCode)
	}
	if entry.ResponseSnippet != "hello" {
		t.Fatalf("ResponseSnippet = %q", entry.ResponseSnippet)
	}
	if entry.DurationMS < 0 {
		t.Fatalf("DurationMS = %d", entry.DurationMS)
	}
	got := <-received
	if got.method != "POST" || got.path != "/hook" {
		t.Fatalf("server saw %s %s", got.method, got.path)
	}
	if got.auth != "Bearer secret" {
		t.Fatalf("Authorization = %q", got.auth)
	}
	if got.body != `{"a":1}` {
		t.Fatalf("body = %q", got.body)
	}
	// A stored Host must never rewrite the connection target, and the transport
	// owns Content-Length.
	if strings.Contains(got.host, "attacker") {
		t.Fatalf("stored Host header was applied: %q", got.host)
	}
	if got.host == "" {
		t.Fatal("server saw no Host header at all")
	}
}

func TestExecutorEncodesFormFields(t *testing.T) {
	received := make(chan capturedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		read, _ := io.ReadAll(r.Body)
		received <- capturedRequest{
			method:      r.Method,
			contentType: r.Header.Get("Content-Type"),
			body:        string(read),
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	entry := newExecutor().run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{
			Method: "POST",
			URL:    server.URL,
			Form: []FormField{
				{Name: "a b", Value: "1&2"},
				{Name: "c", Value: "3"},
			},
			TimeoutMS: 5000,
		},
	})
	if entry.Status != StatusSuccess {
		t.Fatalf("Status = %q (%s)", entry.Status, entry.Error)
	}
	got := <-received
	if got.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q", got.contentType)
	}
	if got.body != "a+b=1%262&c=3" {
		t.Fatalf("body = %q", got.body)
	}
}

// A 4xx or 5xx is a completed request that failed, not a transport error.
func TestExecutorReportsHTTPFailureWithStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	entry := newExecutor().run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{Method: "GET", URL: server.URL, TimeoutMS: 5000},
	})
	if entry.Status != StatusFailed {
		t.Fatalf("Status = %q, want failed", entry.Status)
	}
	if entry.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d", entry.StatusCode)
	}
	if !strings.Contains(entry.Error, "502") {
		t.Fatalf("Error = %q", entry.Error)
	}
}

// curl does not follow redirects by default, and neither does a stored task.
func TestExecutorFollowsRedirectsOnlyWhenAsked(t *testing.T) {
	var target *httptest.Server
	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("final"))
			return
		}
		http.Redirect(w, r, target.URL+"/final", http.StatusFound)
	}))
	defer target.Close()

	executor := newExecutor()
	plain := executor.run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{Method: "GET", URL: target.URL, TimeoutMS: 5000},
	})
	if plain.StatusCode != http.StatusFound {
		t.Fatalf("plain StatusCode = %d, want 302", plain.StatusCode)
	}
	followed := executor.run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{Method: "GET", URL: target.URL, TimeoutMS: 5000, FollowRedirects: true},
	})
	if followed.StatusCode != http.StatusOK || followed.ResponseSnippet != "final" {
		t.Fatalf("followed = %d %q", followed.StatusCode, followed.ResponseSnippet)
	}
}

// A hung upstream must not occupy a worker forever, and the failure must name the
// timeout without echoing the URL (which may carry a token in its query).
func TestExecutorTimesOutAndOmitsCredentialsFromTheError(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	defer close(release)

	entry := newExecutor().run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{
			Method:    "GET",
			URL:       server.URL + "/hook?token=super-secret",
			TimeoutMS: 50,
		},
	})
	if entry.Status != StatusFailed {
		t.Fatalf("Status = %q, want failed", entry.Status)
	}
	if entry.Error != "request timed out" {
		t.Fatalf("Error = %q, want a timeout message", entry.Error)
	}
	if strings.Contains(entry.Error, "super-secret") {
		t.Fatalf("Error leaked the query string: %q", entry.Error)
	}
	if entry.StatusCode != 0 {
		t.Fatalf("StatusCode = %d, want 0", entry.StatusCode)
	}
}

func TestExecutorTruncatesLargeResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 100*1024))
	}))
	defer server.Close()
	entry := newExecutor().run(context.Background(), Task{
		ID: "t", Name: "t", Enabled: true,
		Request: Request{Method: "GET", URL: server.URL, TimeoutMS: 5000},
	})
	if entry.Status != StatusSuccess {
		t.Fatalf("Status = %q", entry.Status)
	}
	if len(entry.ResponseSnippet) != responseSnippetBytes {
		t.Fatalf("snippet length = %d, want %d", len(entry.ResponseSnippet), responseSnippetBytes)
	}
}
