// Package telemetry exposes process-wide Prometheus metrics without retaining request secrets.
package telemetry

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/usage"
)

// Metrics owns an isolated registry. Production uses Default; tests use New.
// Usage label pairs are bounded to prevent untrusted model names exhausting memory.
type Metrics struct {
	labelsMu sync.Mutex
	labels   map[[2]string]struct{}
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
	upstream *prometheus.CounterVec
	tokens   *prometheus.CounterVec
	ttft     *prometheus.HistogramVec
}

var _ usage.Plugin = (*Metrics)(nil)

const maxUsageLabelPairs = 256

var defaultMetrics = sync.OnceValue(func() *Metrics {
	m := New()
	usage.RegisterNamedPlugin("builtin-prometheus", m)
	return m
})

// Default registers exactly one process-wide consumer on the global usage manager.
// Server recreation preserves counters and does not accumulate duplicate plugins.
func Default() *Metrics { return defaultMetrics() }

// New creates metrics without registering a usage consumer.
func New() *Metrics {
	m := &Metrics{registry: prometheus.NewRegistry(), labels: make(map[[2]string]struct{})}
	m.requests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cliproxy_http_requests_total", Help: "Completed client HTTP requests."}, []string{"method", "route", "status"})
	m.duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "cliproxy_http_request_duration_seconds", Help: "Client HTTP request duration including streaming.", Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600}}, []string{"method", "route"})
	m.inFlight = prometheus.NewGauge(prometheus.GaugeOpts{Name: "cliproxy_http_requests_in_flight", Help: "Currently active client HTTP requests."})
	m.upstream = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cliproxy_upstream_requests_total", Help: "Provider usage events by outcome."}, []string{"provider", "model", "outcome"})
	m.tokens = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cliproxy_tokens_total", Help: "Reported token usage."}, []string{"provider", "model", "type"})
	m.ttft = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "cliproxy_ttft_seconds", Help: "Reported time to first token for successful streaming requests.", Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60}}, []string{"provider", "model"})
	m.registry.MustRegister(m.requests, m.duration, m.inFlight, m.upstream, m.tokens, m.ttft, collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return m
}

// ProtectedHandler requires a dedicated bearer token, including for local scrapes.
// An empty token fails closed. The caller must keep this handler out of request logs.
func (m *Metrics) ProtectedHandler(token string) http.Handler {
	handler := m.Handler()
	expected := sha256.Sum256([]byte(token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if token == "" || len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		actual := sha256.Sum256([]byte(parts[1]))
		if subtle.ConstantTimeCompare(expected[:], actual[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// Handler exposes the registry; production routes must use ProtectedHandler.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware must be installed before recovery to observe recovered HTTP statuses.
// It never wraps the response writer, preserving streaming and websocket behavior.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		start := time.Now()
		m.inFlight.Inc()
		defer m.inFlight.Dec()
		c.Next()
		method := boundedMethod(c.Request.Method)
		m.requests.WithLabelValues(method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.duration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	}
}

// HandleUsage counts emitted records, not necessarily every physical upstream attempt.
func (m *Metrics) HandleUsage(_ context.Context, r usage.Record) {
	provider, model := m.usageLabels(r.Provider, r.Model)
	outcome := "success"
	if r.Failed {
		outcome = "failure"
	}
	m.upstream.WithLabelValues(provider, model, outcome).Inc()
	if r.Detail.InputTokens > 0 {
		m.tokens.WithLabelValues(provider, model, "input").Add(float64(r.Detail.InputTokens))
	}
	if r.Detail.OutputTokens > 0 {
		m.tokens.WithLabelValues(provider, model, "output").Add(float64(r.Detail.OutputTokens))
	}
	if r.Stream && !r.Failed && r.TTFT > 0 {
		m.ttft.WithLabelValues(provider, model).Observe(r.TTFT.Seconds())
	}
}

func (m *Metrics) usageLabels(provider, model string) (string, string) {
	if len(provider) > 128 || len(model) > 128 {
		return "other", "other"
	}
	provider, model = label(provider), label(model)
	key := [2]string{provider, model}
	m.labelsMu.Lock()
	defer m.labelsMu.Unlock()
	if _, ok := m.labels[key]; ok {
		return provider, model
	}
	if len(m.labels) >= maxUsageLabelPairs {
		return "other", "other"
	}
	m.labels[key] = struct{}{}
	return provider, model
}

func boundedMethod(method string) string {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "TRACE":
		return method
	default:
		return "OTHER"
	}
}

func label(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}
