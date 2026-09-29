package api

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	sdkaccess "github.com/router-for-me/CLIProxyAPI/v8/sdk/access"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func TestMetricsRoute(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		enabled              bool
		token, authorization string
		status               int
	}{
		{"disabled", false, "secret", "Bearer secret", 404},
		{"missing startup token", true, "", "", 404},
		{"unauthenticated", true, "secret", "", 401},
		{"incorrect token", true, "secret", "Bearer wrong", 401},
		{"authenticated", true, "secret", "Bearer secret", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CLIPROXY_METRICS_TOKEN", tc.token)
			dir := t.TempDir()
			cfg := &config.Config{AuthDir: dir, Debug: true, CommercialMode: true, MetricsEnabled: tc.enabled}
			server := NewServer(cfg, auth.NewManager(nil, nil, nil), sdkaccess.NewManager(), filepath.Join(dir, "config.yaml"))
			req := httptest.NewRequest("GET", "/metrics", nil)
			req.Header.Set("Authorization", tc.authorization)
			w := httptest.NewRecorder()
			server.engine.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.status == 200 && !strings.Contains(w.Body.String(), "go_goroutines") {
				t.Fatal("missing runtime metrics")
			}
		})
	}
}
