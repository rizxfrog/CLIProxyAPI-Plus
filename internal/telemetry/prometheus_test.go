package telemetry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/usage"
)

func metric(t *testing.T, m *Metrics, name string, labels map[string]string) *dto.Metric {
	t.Helper()
	families, err := m.registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, sample := range family.Metric {
			match := len(sample.Label) == len(labels)
			for _, pair := range sample.Label {
				if labels[pair.GetName()] != pair.GetValue() {
					match = false
				}
			}
			if match {
				return sample
			}
		}
	}
	t.Fatalf("missing metric %s %v", name, labels)
	return nil
}

func TestHTTPMetrics(t *testing.T) {
	m := New()
	engine := gin.New()
	engine.Use(m.Middleware(), gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) { c.AbortWithStatus(500) }))
	engine.GET("/items/:id", func(c *gin.Context) { c.Status(201) })
	engine.GET("/panic", func(c *gin.Context) { panic("test") })
	engine.GET("/metrics", gin.WrapH(m.ProtectedHandler("secret")))
	for _, tc := range []struct {
		path, route string
		status      int
	}{{"/items/private-id", "/items/:id", 201}, {"/missing/private-id", "unmatched", 404}, {"/panic", "/panic", 500}} {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.status {
			t.Fatalf("status = %d", w.Code)
		}
		got := metric(t, m, "cliproxy_http_requests_total", map[string]string{"method": "GET", "route": tc.route, "status": fmt.Sprint(tc.status)}).GetCounter().GetValue()
		if got != 1 {
			t.Fatalf("count = %v", got)
		}
		if metric(t, m, "cliproxy_http_request_duration_seconds", map[string]string{"method": "GET", "route": tc.route}).GetHistogram().GetSampleCount() != 1 {
			t.Fatal("duration not recorded")
		}
	}
	if metric(t, m, "cliproxy_http_requests_in_flight", nil).GetGauge().GetValue() != 0 {
		t.Fatal("in-flight leaked")
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	engine.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("scrape status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), `route="/metrics"`) || strings.Contains(w.Body.String(), "private-id") {
		t.Fatal("scrape or raw path leaked into labels")
	}
}

func TestProtectedHandler(t *testing.T) {
	for _, tc := range []struct {
		token, auth string
		status      int
	}{
		{"secret", "", 401}, {"secret", "Bearer wrong", 401}, {"secret", "Basic secret", 401}, {"secret", "Bearer secret extra", 401}, {"", "Bearer secret", 401}, {"secret", "Bearer secret", 200}, {"secret", "bearer secret", 200},
	} {
		m := New()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/metrics", nil)
		req.Header.Set("Authorization", tc.auth)
		m.ProtectedHandler(tc.token).ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("status = %d, want %d", w.Code, tc.status)
		}
		if strings.Contains(w.Body.String(), "secret") {
			t.Fatal("credential leaked")
		}
	}
}

func TestUsageMetrics(t *testing.T) {
	m := New()
	r := usage.Record{Provider: "provider", Model: "model", Stream: true, TTFT: 2 * time.Second, APIKey: "never-export-this", Detail: usage.Detail{InputTokens: 10, OutputTokens: 20}}
	// Dispatch synchronously through the plugin interface to avoid timing assumptions.
	var plugin usage.Plugin = m
	plugin.HandleUsage(context.Background(), r)
	r.Failed = true
	r.Detail = usage.Detail{InputTokens: -1, OutputTokens: -1}
	plugin.HandleUsage(context.Background(), r)
	for _, outcome := range []string{"success", "failure"} {
		if metric(t, m, "cliproxy_upstream_requests_total", map[string]string{"provider": "provider", "model": "model", "outcome": outcome}).GetCounter().GetValue() != 1 {
			t.Fatal("missing usage count")
		}
	}
	for kind, want := range map[string]float64{"input": 10, "output": 20} {
		if metric(t, m, "cliproxy_tokens_total", map[string]string{"provider": "provider", "model": "model", "type": kind}).GetCounter().GetValue() != want {
			t.Fatal("wrong tokens")
		}
	}
	h := metric(t, m, "cliproxy_ttft_seconds", map[string]string{"provider": "provider", "model": "model"}).GetHistogram()
	if h.GetSampleCount() != 1 || h.GetSampleSum() != 2 {
		t.Fatal("wrong TTFT")
	}
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	if strings.Contains(w.Body.String(), r.APIKey) {
		t.Fatal("API key leaked")
	}
}

func TestBoundedLabels(t *testing.T) {
	m := New()
	for i := 0; i < maxUsageLabelPairs+10; i++ {
		m.HandleUsage(context.Background(), usage.Record{Provider: "p", Model: fmt.Sprint(i)})
	}
	if len(m.labels) != maxUsageLabelPairs {
		t.Fatal("unbounded label cache")
	}
	if metric(t, m, "cliproxy_upstream_requests_total", map[string]string{"provider": "other", "model": "other", "outcome": "success"}).GetCounter().GetValue() != 10 {
		t.Fatal("overflow not aggregated")
	}
	if p, model := m.usageLabels("p", strings.Repeat("x", 129)); p != "other" || model != "other" {
		t.Fatal("oversized label accepted")
	}
	if boundedMethod("ARBITRARY") != "OTHER" {
		t.Fatal("unbounded method")
	}
}

func TestConcurrentRequestsPreserveFlush(t *testing.T) {
	m := New()
	engine := gin.New()
	engine.Use(m.Middleware())
	entered := make(chan struct{}, 8)
	release := make(chan struct{})
	engine.GET("/stream", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = c.Writer.WriteString("data: test\n\n")
		c.Writer.Flush()
		entered <- struct{}{}
		<-release
	})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stream", nil))
			if !w.Flushed {
				t.Error("flush lost")
			}
		}()
	}
	for i := 0; i < 8; i++ {
		<-entered
	}
	active := metric(t, m, "cliproxy_http_requests_in_flight", nil).GetGauge().GetValue()
	close(release)
	wg.Wait()
	if active != 8 {
		t.Fatalf("active = %v", active)
	}
	if metric(t, m, "cliproxy_http_requests_in_flight", nil).GetGauge().GetValue() != 0 {
		t.Fatal("in-flight leaked")
	}
}

type usageCallback func(context.Context, usage.Record)

func (f usageCallback) HandleUsage(ctx context.Context, r usage.Record) { f(ctx, r) }

func TestUsageManagerDelivery(t *testing.T) {
	m := New()
	manager := usage.NewManager(16)
	defer manager.Stop()
	done := make(chan struct{})
	manager.Register(m)
	manager.Register(usageCallback(func(context.Context, usage.Record) { close(done) }))
	manager.Publish(context.Background(), usage.Record{Provider: "p", Model: "m"})
	<-done
	if metric(t, m, "cliproxy_upstream_requests_total", map[string]string{"provider": "p", "model": "m", "outcome": "success"}).GetCounter().GetValue() != 1 {
		t.Fatal("usage event not delivered")
	}
}

func TestIndependentRegistriesAndDefault(t *testing.T) {
	a, b := New(), New()
	a.HandleUsage(context.Background(), usage.Record{})
	if len(b.labels) != 0 {
		t.Fatal("registries share state")
	}
	if Default() != Default() {
		t.Fatal("default recreated")
	}
}
