package cliproxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	xiaohuanxiongauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

// TestParseXiaohuanxiongCatalogueKeepsChatCategoryOnly pins that only the chat
// category is consumed and that a non-zero envelope code is an error.
func TestParseXiaohuanxiongCatalogueKeepsChatCategoryOnly(t *testing.T) {
	raw := []byte(`{"code":0,"message":"success","data":{"categories":[
		{"type":"image","models":[{"name":"sd-xl"}]},
		{"type":"chat","models":[
			{"name":"sn-glm-5-3","description":"GLM 5.3","visible":true,
			 "ability_level":2,"tags":["general","reasoning"],
			 "params":{"context_window":1000000,"max_tokens":100000}},
			{"name":"","visible":true},
			{"name":"sn-kimi-k3","tags":["vision"],"params":{"context_window":256000}}
		]}
	]}}`)
	models, errParse := parseXiaohuanxiongCatalogue(raw)
	if errParse != nil {
		t.Fatalf("parse: %v", errParse)
	}
	if len(models) != 2 {
		t.Fatalf("want 2 chat models, got %d: %+v", len(models), models)
	}
	if models[0].Name != "sn-glm-5-3" || models[0].Params.ContextWindow != 1000000 {
		t.Fatalf("first row lost: %+v", models[0])
	}
	if models[1].Name != "sn-kimi-k3" || models[1].Params.ContextWindow != 256000 {
		t.Fatalf("second row lost: %+v", models[1])
	}
}

// TestParseXiaohuanxiongCatalogueRejectsNonZeroCode pins that a failed envelope
// is surfaced instead of being read as an empty catalogue.
func TestParseXiaohuanxiongCatalogueRejectsNonZeroCode(t *testing.T) {
	raw := []byte(`{"code":200001,"message":"authorization_empty_error","data":{}}`)
	if _, errParse := parseXiaohuanxiongCatalogue(raw); errParse == nil {
		t.Fatal("non-zero code must fail")
	}
}

// TestParseXiaohuanxiongCatalogueEmptyCategoryIsNotAnError pins that a chat
// category with no rows yields no models rather than a decode failure; the probe
// keeps the fallback catalogue published in that case.
func TestParseXiaohuanxiongCatalogueEmptyCategoryIsNotAnError(t *testing.T) {
	models, errParse := parseXiaohuanxiongCatalogue([]byte(`{"code":0,"data":{"categories":[{"type":"chat","models":[]}]}}`))
	if errParse != nil {
		t.Fatalf("parse: %v", errParse)
	}
	if len(models) != 0 {
		t.Fatalf("want no models, got %+v", models)
	}
}

// TestXiaohuanxiongCatalogueModelInfoCarriesContext pins the per-model mapping.
func TestXiaohuanxiongCatalogueModelInfoCarriesContext(t *testing.T) {
	entry := xiaohuanxiongCatalogModel{Name: "sn-glm-5-3", Description: "GLM 5.3"}
	entry.Params.ContextWindow = 1000000
	entry.Tags = []string{"general", "reasoning", "vision"}

	info := xiaohuanxiongCatalogueModelInfo(entry)
	if info == nil {
		t.Fatal("model info is nil")
	}
	if info.ID != "sn-glm-5-3" || info.DisplayName != "GLM 5.3" {
		t.Fatalf("identity lost: %+v", info)
	}
	if info.Type != constant.Xiaohuanxiong || info.OwnedBy != "sensetime" {
		t.Fatalf("attribution lost: %+v", info)
	}
	if info.ContextLength != 1000000 {
		t.Fatalf("context lost: %+v", info.ContextLength)
	}
	if len(info.SupportedInputModalities) != 2 {
		t.Fatalf("image capability lost: %+v", info.SupportedInputModalities)
	}
}

// TestXiaohuanxiongCatalogueModelInfoOmitsImageWithoutVisionTag pins that a
// text-only row does not advertise image input.
func TestXiaohuanxiongCatalogueModelInfoOmitsImageWithoutVisionTag(t *testing.T) {
	info := xiaohuanxiongCatalogueModelInfo(xiaohuanxiongCatalogModel{Name: "raccoon-405a1c", Tags: []string{"general", "fast"}})
	if info == nil {
		t.Fatal("model info is nil")
	}
	if xiaohuanxiongTagPresent(info.SupportedInputModalities, "image") {
		t.Fatalf("image must not be advertised: %+v", info.SupportedInputModalities)
	}
}

// TestXiaohuanxiongCatalogueModelInfoSkipsBlankName pins that a malformed row
// cannot register a nameless model.
func TestXiaohuanxiongCatalogueModelInfoSkipsBlankName(t *testing.T) {
	if info := xiaohuanxiongCatalogueModelInfo(xiaohuanxiongCatalogModel{Name: "   "}); info != nil {
		t.Fatalf("blank name produced %+v", info)
	}
}

// TestXiaohuanxiongCatalogueOriginStripsLLMPath pins the origin derivation. A
// credential stores the full gateway base URL while the client appends that path
// again, so keeping the path would request /api/web/llm/v2/api/web/llm/v2/...
func TestXiaohuanxiongCatalogueOriginStripsLLMPath(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		want    string
	}{
		{"production", "https://xiaohuanxiong.com/api/web/llm/v2", "https://xiaohuanxiong.com"},
		{"trailing slash", "https://xiaohuanxiong.com/api/web/llm/v2/", "https://xiaohuanxiong.com"},
		{"self-hosted", "https://self-hosted.example/api/web/llm/v2", "https://self-hosted.example"},
		{"missing attribute", "", xiaohuanxiongauth.BaseURL},
		{"unrelated path", "https://other.example/proxy", xiaohuanxiongauth.BaseURL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &coreauth.Auth{Attributes: map[string]string{"base_url": tc.baseURL}}
			if got := xiaohuanxiongCatalogueOrigin(auth); got != tc.want {
				t.Fatalf("origin = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestXiaohuanxiongCatalogueKeySeparatesSessions pins that a rotated token on the
// same auth ID is treated as a different catalogue, so a stale entry cannot mask
// a fresh login.
func TestXiaohuanxiongCatalogueKeySeparatesSessions(t *testing.T) {
	base := &coreauth.Auth{ID: "acct", Attributes: map[string]string{
		"base_url":     "https://xiaohuanxiong.com/api/web/llm/v2",
		"access_token": "token-a",
	}}
	rotated := base.Clone()
	rotated.Attributes["access_token"] = "token-b"

	keyBase := xiaohuanxiongCatalogueKey(base)
	if keyBase == "" {
		t.Fatal("key is empty")
	}
	if keyBase == xiaohuanxiongCatalogueKey(rotated) {
		t.Fatal("a rotated session must change the catalogue key")
	}
}

// TestXiaohuanxiongCatalogueProbeEndToEnd pins discovery against a gateway
// serving /model_catalog: the mapped catalogue must reach the global registry,
// which is what the model list endpoint and selection read.
func TestXiaohuanxiongCatalogueProbeEndToEnd(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/web/llm/v2/model_catalog" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"categories":[
			{"type":"chat","default_model":"live-default","models":[
				{"name":"live-xh-model","description":"Live Model","visible":true,
				 "ability_level":3,"tags":["general","code","vision"],
				 "params":{"context_window":1000000,"max_tokens":100000}},
				{"name":"live-default","visible":false,"tags":["auto"],
				 "params":{"context_window":256000}}
			]}
		]}}`))
	}))
	defer upstream.Close()

	reg := registry.GetGlobalRegistry()
	auth := &coreauth.Auth{
		ID:       "xiaohuanxiong-catalogue-e2e",
		Provider: constant.Xiaohuanxiong,
		Attributes: map[string]string{
			"auth_kind":    coreauth.AuthKindOAuth,
			"access_token": "tok",
			"base_url":     upstream.URL + "/api/web/llm/v2",
		},
		Metadata: map[string]any{"access_token": "tok", "base_url": upstream.URL + "/api/web/llm/v2"},
	}
	service := &Service{cfg: &config.Config{}}
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	catalogue, errFetch := service.fetchXiaohuanxiongCatalogue(t.Context(), auth)
	if errFetch != nil {
		t.Fatalf("fetch catalogue: %v", errFetch)
	}
	if len(catalogue) != 2 || catalogue[0].Name != "live-xh-model" {
		t.Fatalf("catalogue = %+v", catalogue)
	}
	models := service.xiaohuanxiongModelsForCatalogue(auth, catalogue)
	service.registerResolvedModelsForAuth(auth, constant.Xiaohuanxiong, models)

	info := registry.LookupModelInfo("live-xh-model", constant.Xiaohuanxiong)
	if info == nil {
		t.Fatalf("live-xh-model was not published; got %d models", len(models))
	}
	if info.ContextLength != 1000000 {
		t.Fatalf("live context not published: %+v", info.ContextLength)
	}
	// The hidden default row is routable, so it must be published too.
	if hidden := registry.LookupModelInfo("live-default", constant.Xiaohuanxiong); hidden == nil {
		t.Fatal("hidden default model was dropped; it is still routable upstream")
	}
}

// TestXiaohuanxiongCatalogueProbeKeepsFallbackOnFailure pins that a failed fetch
// schedules a retry without revoking the fallback catalogue.
func TestXiaohuanxiongCatalogueProbeKeepsFallbackOnFailure(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer upstream.Close()

	auth := &coreauth.Auth{
		ID:       "xiaohuanxiong-probe-failure",
		Provider: constant.Xiaohuanxiong,
		Attributes: map[string]string{
			"access_token": "tok",
			"base_url":     upstream.URL + "/api/web/llm/v2",
		},
	}
	service := &Service{cfg: &config.Config{}}
	if _, errFetch := service.fetchXiaohuanxiongCatalogue(t.Context(), auth); errFetch == nil {
		t.Fatal("a 500 must fail the fetch")
	}

	key := xiaohuanxiongCatalogueKey(auth)
	t.Cleanup(func() {
		xiaohuanxiongCatalogueMu.Lock()
		delete(xiaohuanxiongCatalogue, key)
		xiaohuanxiongCatalogueMu.Unlock()
	})
	service.deferXiaohuanxiongCatalogueRetry(key)
	xiaohuanxiongCatalogueMu.RLock()
	entry := xiaohuanxiongCatalogue[key]
	xiaohuanxiongCatalogueMu.RUnlock()
	if entry.nextRetry.IsZero() {
		t.Fatal("a failed fetch must schedule a retry")
	}
}
