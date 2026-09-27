package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func providerModelsTestRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/provider-models", h.GetProviderModels)
	r.GET("/provider-models/:provider", h.GetProviderModel)
	r.PUT("/provider-models/:provider", h.PutProviderModel)
	r.DELETE("/provider-models/:provider", h.DeleteProviderModel)
	return r
}
func providerModelsRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}
func TestProviderModelsCatalogWithoutCredentials(t *testing.T) {
	h := &Handler{cfg: &config.Config{ProviderModels: map[string]config.ProviderModels{"trae": {Disabled: []string{"glm-5.2", "retired"}, Custom: []config.ProviderModel{{ID: "glm-5.4"}}}}}}
	r := providerModelsTestRouter(h)
	w := providerModelsRequest(r, "GET", "/provider-models/trae", "")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var catalog managedProviderCatalog
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.ReadOnly || catalog.RuntimeStatus != "unverified" {
		t.Fatalf("incorrect status: %#v", catalog)
	}
	rows := map[string]managedProviderModel{}
	for _, row := range catalog.Models {
		rows[row.ID] = row
		if row.RegisteredAccountCount == nil || *row.RegisteredAccountCount != 0 {
			t.Fatalf("fabricated count: %#v", row)
		}
	}
	if rows["glm-5.2"].Enabled || rows["glm-5.2"].Source != "default" || !rows["glm-5.4"].Enabled || rows["glm-5.4"].Source != "custom" || rows["retired"].Source != "unavailable" {
		t.Fatalf("bad catalog: %#v", rows)
	}
	if w := providerModelsRequest(r, "GET", "/provider-models", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"provider":"cline"`) {
		t.Fatalf("incomplete collection: %s", w.Body.String())
	}
}
func TestProviderModelsManagementSaveValidationReset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("port: 8317\n"), 0600); err != nil {
		t.Fatal(err)
	}
	applied := make(chan *config.Config, 2)
	h := &Handler{cfg: &config.Config{}, configFilePath: path, configReloadHook: func(_ context.Context, cfg *config.Config) { applied <- cfg }}
	r := providerModelsTestRouter(h)
	for _, body := range []string{`null`, `{"custom":[{"id":"a"},{"id":"a"}]}`, `{"disabled":["*"]}`, `{"unknown":true}`, `{} {}`, `{"custom":[{"id":"new","context-length":-1}]}`} {
		if w := providerModelsRequest(r, "PUT", "/provider-models/trae", body); w.Code != 400 {
			t.Fatalf("accepted %s: %d %s", body, w.Code, w.Body.String())
		}
	}
	if len(h.cfg.ProviderModels) != 0 {
		t.Fatal("invalid request mutated config")
	}
	w := providerModelsRequest(r, "PUT", "/provider-models/trae", `{"disabled":["glm-5.2"],"custom":[{"id":"glm-5.4","display-name":"GLM 5.4"}]}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"runtime_status":"pending"`) {
		t.Fatalf("save failed: %s", w.Body.String())
	}
	if cfg := <-applied; len(cfg.ProviderModels["trae"].Custom) != 1 {
		t.Fatal("hook omitted overlays")
	}
	loaded, err := config.LoadConfig(path)
	if err != nil || len(loaded.ProviderModels["trae"].Disabled) != 1 {
		t.Fatalf("save not persisted: %v", err)
	}
	w = providerModelsRequest(r, "DELETE", "/provider-models/trae", "")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	<-applied
	loaded, err = config.LoadConfig(path)
	if err != nil || len(loaded.ProviderModels) != 0 {
		t.Fatalf("reset not persisted: %v", err)
	}
	h.mu.Lock()
	h.configFilePath = filepath.Join(t.TempDir(), "absent.yaml")
	h.mu.Unlock()
	if w := providerModelsRequest(r, "PUT", "/provider-models/trae", `{"disabled":["old"]}`); w.Code != 500 {
		t.Fatal("expected persistence failure")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.cfg.ProviderModels) != 0 {
		t.Fatal("failed save mutated config")
	}
}
func TestProviderModelsHomeReadOnly(t *testing.T) {
	cfg := &config.Config{ProviderModels: map[string]config.ProviderModels{"trae": {Custom: []config.ProviderModel{{ID: "new"}}}}}
	cfg.Home.Enabled = true
	path := filepath.Join(t.TempDir(), "config.yaml")
	before := []byte("port: 8317\n")
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	h := &Handler{cfg: cfg, configFilePath: path}
	r := providerModelsTestRouter(h)
	w := providerModelsRequest(r, "GET", "/provider-models/trae", "")
	var catalog managedProviderCatalog
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if !catalog.ReadOnly || catalog.RuntimeStatus != "home-managed" || catalog.CatalogScope != "local-preview" {
		t.Fatalf("misleading Home catalog: %#v", catalog)
	}
	for _, row := range catalog.Models {
		if row.RegisteredAccountCount != nil {
			t.Fatal("fabricated Home count")
		}
	}
	for _, method := range []string{"PUT", "DELETE"} {
		w := providerModelsRequest(r, method, "/provider-models/trae", `{}`)
		if w.Code != 409 || !strings.Contains(w.Body.String(), "provider_models_home_managed") {
			t.Fatalf("Home write accepted: %s", w.Body.String())
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) || len(cfg.ProviderModels["trae"].Custom) != 1 {
		t.Fatal("Home request modified saved intent")
	}
}
func TestProviderModelsCatalogSkipsAuthWithoutProvider(t *testing.T) {
	// An auth whose provider resolves to empty must not crash the catalog: it
	// cannot be attributed to any provider, and add() early-returns for an
	// empty provider, so an unguarded counts[provider] write would panic.
	reg := registry.GetGlobalRegistry()
	const clientID = "provider-models-empty-provider"
	reg.RegisterClient(clientID, "someprovider", []*registry.ModelInfo{{ID: "m1"}})
	t.Cleanup(func() { reg.UnregisterClient(clientID) })

	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{ID: clientID, Provider: "", Status: coreauth.StatusActive}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}

	h := &Handler{cfg: &config.Config{}, authManager: manager}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("catalog panicked on auth without provider: %v", r)
		}
	}()
	if catalogs := h.providerCatalogsLocked(); len(catalogs) == 0 {
		t.Fatal("expected native provider catalogs")
	}
}

func TestProviderModelsCountsCanonicalModelsNotAliases(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{ID: t.Name(), Provider: "trae"}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
	reg := registry.GetGlobalRegistry()
	reg.RegisterClient(auth.ID, "trae", []*registry.ModelInfo{{ID: "alias", MetadataModelID: "glm-5.2"}, {ID: "prefix/alias", MetadataModelID: "glm-5.2"}})
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })
	h := &Handler{cfg: &config.Config{}, authManager: manager}
	w := providerModelsRequest(providerModelsTestRouter(h), "GET", "/provider-models/trae", "")
	var catalog managedProviderCatalog
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	for _, row := range catalog.Models {
		if row.ID == "glm-5.2" {
			if row.RegisteredAccountCount == nil || *row.RegisteredAccountCount != 1 {
				t.Fatalf("incorrect count: %#v", row)
			}
			return
		}
	}
	t.Fatal("canonical model missing")
}
