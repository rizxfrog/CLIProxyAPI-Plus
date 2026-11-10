package cliproxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codebuddycn"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

// TestCodeBuddyCatalogueModelInfoCarriesLimits pins the per-model mapping.
func TestCodeBuddyCatalogueModelInfoCarriesLimits(t *testing.T) {
	info := codebuddyCatalogueModelInfo(constant.CodeBuddyCN, codebuddycn.CatalogModel{
		ID:                 "glm-5.2",
		Name:               "GLM-5.2",
		ContextWindow:      1_000_000,
		MaxOutputTokens:    50_000,
		SupportsImages:     true,
		ReasoningLevels:    []string{"low", "high"},
		CanDisableThinking: true,
	})
	if info == nil {
		t.Fatal("model info is nil")
	}
	if info.ID != "glm-5.2" || info.DisplayName != "GLM-5.2" {
		t.Fatalf("identity lost: %+v", info)
	}
	if info.ContextLength != 1_000_000 || info.MaxCompletionTokens != 50_000 {
		t.Fatalf("limits lost: %+v", info)
	}
	if info.Type != constant.CodeBuddyCN || info.OwnedBy != "codebuddy" {
		t.Fatalf("attribution lost: %+v", info)
	}
	if len(info.SupportedInputModalities) != 2 {
		t.Fatalf("image capability lost: %+v", info.SupportedInputModalities)
	}
	if info.Thinking == nil || len(info.Thinking.Levels) != 2 || !info.Thinking.ZeroAllowed {
		t.Fatalf("reasoning lost: %+v", info.Thinking)
	}
}

// TestCodeBuddyCatalogueModelInfoSkipsBlankID pins that a malformed row cannot
// register a nameless model.
func TestCodeBuddyCatalogueModelInfoSkipsBlankID(t *testing.T) {
	if info := codebuddyCatalogueModelInfo(constant.CodeBuddyCN, codebuddycn.CatalogModel{ID: "   "}); info != nil {
		t.Fatalf("blank id produced %+v", info)
	}
}

// TestCodeBuddyCatalogueModelInfoOmitsUndeclaredThinking pins that a model with
// no declared efforts carries no thinking record, so no restriction is invented.
func TestCodeBuddyCatalogueModelInfoOmitsUndeclaredThinking(t *testing.T) {
	info := codebuddyCatalogueModelInfo(constant.CodeBuddyCN, codebuddycn.CatalogModel{ID: "plain", Name: "Plain"})
	if info == nil {
		t.Fatal("model info is nil")
	}
	if info.Thinking != nil {
		t.Fatalf("undeclared thinking must stay unconstrained: %+v", info.Thinking)
	}
}

// TestCodeBuddyCatalogueProbeEndToEnd pins discovery against a gateway serving
// /v3/config: the mapped catalogue must reach the global registry, which is what
// the model list endpoint and selection read.
func TestCodeBuddyCatalogueProbeEndToEnd(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/plugin/account":
			_, _ = w.Write([]byte(`{"code":0,"data":{"uid":"u1","pluginEnabled":true}}`))
		case "/v3/config":
			_, _ = w.Write([]byte(`{"code":0,"data":{
				"models":[
					{"id":"live-model","name":"Live Model","maxInputTokens":500000,
					 "maxOutputTokens":32000,"supportsToolCall":true,
					 "reasoning":{"supportedEfforts":["low","medium","high"]}}
				],
				"agents":[{"name":"cli","models":["live-model"]}]
			}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	reg := registry.GetGlobalRegistry()
	auth := &coreauth.Auth{
		ID:       "codebuddy-catalogue-e2e",
		Provider: constant.CodeBuddyCN,
		Attributes: map[string]string{
			"auth_kind":    coreauth.AuthKindOAuth,
			"access_token": "tok",
		},
		Metadata: map[string]any{"access_token": "tok"},
	}
	service := &Service{cfg: &config.Config{}}
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	// The catalogue is addressed from the auth host, so point a client at the
	// test server and keep the service pipeline under test.
	client := codebuddycn.NewTestClientAt(upstream.URL)
	catalogue, errFetch := client.FetchCatalog(t.Context(), "tok", codebuddycn.ProductCLI, "")
	if errFetch != nil {
		t.Fatalf("fetch catalogue: %v", errFetch)
	}
	if len(catalogue.Models) != 1 || catalogue.Models[0].ID != "live-model" {
		t.Fatalf("catalogue = %+v", catalogue.Models)
	}
	models := service.codebuddyModelsForCatalogue(auth, catalogue)
	service.registerResolvedModelsForAuth(auth, constant.CodeBuddyCN, models)

	info := registry.LookupModelInfo("live-model", constant.CodeBuddyCN)
	if info == nil {
		t.Fatalf("live-model was not published; got %d models", len(models))
	}
	if info.ContextLength != 500000 || info.MaxCompletionTokens != 32000 {
		t.Fatalf("live limits not published: %+v", info)
	}
	if info.Thinking == nil || len(info.Thinking.Levels) != 3 {
		t.Fatalf("thinking levels not published: %+v", info.Thinking)
	}
}
