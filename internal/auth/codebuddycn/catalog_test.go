package codebuddycn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newCatalogClient points the client's auth anchors at a test server so the
// derived catalogue origin resolves there.
func newCatalogClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := NewClientWithProxyURL(nil, "")
	client.stateURL = server.URL + "/v2/plugin/auth/state"
	client.tokenURL = server.URL + "/v2/plugin/auth/token"
	client.refreshURL = server.URL + "/v2/plugin/auth/token/refresh"
	return client, server
}

// catalogServer answers the account probe and the configuration endpoint.
func catalogServer(t *testing.T, configBody string, sawUA *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sawUA != nil && r.URL.Path == ConfigPath {
			*sawUA = r.Header.Get("User-Agent")
		}
		switch r.URL.Path {
		case AccountPath:
			_, _ = w.Write([]byte(`{"code":0,"data":{"uid":"u1","uin":"100","nickname":"tester","pluginEnabled":true}}`))
		case ConfigPath:
			_, _ = w.Write([]byte(configBody))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	})
}

// TestFetchCatalogMapsTheCloudProductConfiguration pins the model mapping.
//
// The catalogue is one field of the whole cloud product configuration, so the
// mapping has to ignore agents/productFeatures and read only data.models.
func TestFetchCatalogMapsTheCloudProductConfiguration(t *testing.T) {
	var sawUA string
	client, server := newCatalogClient(t, catalogServer(t, `{"code":0,"data":{
		"models":[
			{"id":"glm-5.2","name":"GLM-5.2","maxInputTokens":1000000,"maxOutputTokens":50000,
			 "supportsImages":true,"supportsToolCall":true,
			 "reasoning":{"supportedEfforts":["low","medium","high"],"canDisableThinking":true},
			 "tags":["chat"]},
			{"id":"glm-5.0-turbo","name":"GLM-5.0 Turbo","maxInputTokens":200000},
			{"id":"","name":"nameless"},
			{"id":"img-gen","tags":["text-to-image"]}
		],
		"agents":[{"name":"cli","models":["glm-5.2"]}],
		"productFeatures":{"CommandsSmartMerge":true}
	}}`, &sawUA))
	defer server.Close()

	catalog, errFetch := client.FetchCatalog(context.Background(), "tok", ProductCLI, "")
	if errFetch != nil {
		t.Fatalf("fetch catalog: %v", errFetch)
	}
	if len(catalog.Models) != 3 {
		t.Fatalf("model count = %d, want 3 (blank id dropped): %+v", len(catalog.Models), catalog.Models)
	}
	first := catalog.Models[0]
	if first.ID != "glm-5.2" || first.Name != "GLM-5.2" {
		t.Fatalf("identity lost: %+v", first)
	}
	if first.ContextWindow != 1000000 || first.MaxOutputTokens != 50000 {
		t.Fatalf("limits lost: %+v", first)
	}
	if !first.SupportsImages || !first.SupportsToolCall {
		t.Fatalf("capabilities lost: %+v", first)
	}
	if len(first.ReasoningLevels) != 3 || !first.CanDisableThinking {
		t.Fatalf("reasoning lost: %+v", first)
	}
	// The CLI agent's ids are exposed separately: data.models is wider than what
	// the CLI advertises.
	if len(catalog.CLIModelIDs) != 1 || catalog.CLIModelIDs[0] != "glm-5.2" {
		t.Fatalf("cli agent ids = %+v", catalog.CLIModelIDs)
	}
	if len(catalog.AgentNames) != 1 || catalog.AgentNames[0] != "cli" {
		t.Fatalf("agent names = %+v", catalog.AgentNames)
	}
	// The User-Agent is what selects the catalogue server-side.
	if sawUA != "CLI/2.63.2 CodeBuddy/2.63.2" {
		t.Fatalf("User-Agent = %q", sawUA)
	}
}

// TestFetchCatalogUsesTheIDEProductLine pins that the product line reaches the
// User-Agent, because the gateway returns a different catalogue per product.
func TestFetchCatalogUsesTheIDEProductLine(t *testing.T) {
	var sawUA string
	client, server := newCatalogClient(t, catalogServer(t, `{"code":0,"data":{"models":[]}}`, &sawUA))
	defer server.Close()

	if _, errFetch := client.FetchCatalog(context.Background(), "tok", ProductIDE, ""); errFetch != nil {
		t.Fatalf("fetch catalog: %v", errFetch)
	}
	if sawUA != "WorkBuddy/5.3.14" {
		t.Fatalf("User-Agent = %q, want the IDE line", sawUA)
	}
}

// TestFetchCatalogRejectsANullCatalogue pins that a null model list is an error.
//
// The gateway answers HTTP 200 with models == null for a revoked token, which is
// identical to the anonymous response. Reporting it as an empty catalogue would
// silently revoke every model on the account.
func TestFetchCatalogRejectsANullCatalogue(t *testing.T) {
	client, server := newCatalogClient(t, catalogServer(t, `{"code":0,"data":{"models":null}}`, nil))
	defer server.Close()

	_, errFetch := client.FetchCatalog(context.Background(), "tok", ProductCLI, "")
	if errFetch == nil {
		t.Fatal("a null catalogue must not be reported as success")
	}
	if !strings.Contains(errFetch.Error(), "null") {
		t.Fatalf("message does not explain the null catalogue: %v", errFetch)
	}
}

// TestFetchCatalogProbesTheAccountFirst pins that a rejected token fails before
// the catalogue is read.
func TestFetchCatalogProbesTheAccountFirst(t *testing.T) {
	var configHit bool
	client, server := newCatalogClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ConfigPath {
			configHit = true
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"msg":"unauthorized"}`))
	}))
	defer server.Close()

	if _, errFetch := client.FetchCatalog(context.Background(), "dead", ProductCLI, ""); errFetch == nil {
		t.Fatal("a rejected token must fail")
	}
	if configHit {
		t.Fatal("the catalogue must not be requested for a rejected token")
	}
}

// TestFetchCatalogReportsAUserAgentRejection pins the 12403 envelope code.
func TestFetchCatalogReportsAUserAgentRejection(t *testing.T) {
	client, server := newCatalogClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == AccountPath {
			_, _ = w.Write([]byte(`{"code":0,"data":{"uid":"u1"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":12403,"msg":"check ua, get coding copilot version error"}`))
	}))
	defer server.Close()

	_, errFetch := client.FetchCatalog(context.Background(), "tok", ProductCLI, "")
	if errFetch == nil {
		t.Fatal("a UA rejection must fail")
	}
	if !strings.Contains(errFetch.Error(), "User-Agent") {
		t.Fatalf("message does not name the User-Agent: %v", errFetch)
	}
}

// TestFetchCatalogRequiresAToken pins the empty-token guard.
func TestFetchCatalogRequiresAToken(t *testing.T) {
	client := NewClientWithProxyURL(nil, "")
	if _, errFetch := client.FetchCatalog(context.Background(), "   ", ProductCLI, ""); errFetch == nil {
		t.Fatal("an empty token must fail without a request")
	}
}
