package trae

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cataloguePayload mirrors the live get_detail_param response: config_info_list
// rows carry config_name as the routing key, with display metadata nested under
// display_config and the window under context_window_tokens.
const cataloguePayload = `{
  "code": 0,
  "message": "",
  "allow_tenant_user_add_model": true,
  "config_info_list": [
    {"config_name":"glm-5.2","config_switch":true,"is_invisible_to_user":false,"usage":"chat_completion",
     "context_window_tokens":{"dev":200000},
     "display_config":{"display_name":"GLM-5.2","multimodal":false,"model_capability":"reasoning_model"}},
    {"config_name":"kimi-k3","config_switch":true,"is_invisible_to_user":false,
     "context_window_tokens":{"dev":200000,"max":256000},
     "display_config":{"display_name":"Kimi-K3","multimodal":true}},
    {"config_name":"glm-5.2","config_switch":true,"display_config":{"display_name":"duplicate"}},
    {"config_name":"computer_use_subagent","is_invisible_to_user":true,"config_switch":true},
    {"config_name":"custom_model_placeholder","config_switch":true,"display_config":{"display_name":""}},
    {"config_name":"disabled-model","config_switch":false},
    {"display_config":{"display_name":"no routing key"}}
  ]
}`

func TestParseDetailParamKeepsRoutableRowsOnly(t *testing.T) {
	models, errParse := ParseDetailParam([]byte(cataloguePayload))
	if errParse != nil {
		t.Fatalf("ParseDetailParam: %v", errParse)
	}
	if len(models) != 3 {
		t.Fatalf("models = %+v, want 3 routable rows", models)
	}
	if models[0].ID != "glm-5.2" || models[0].DisplayName != "GLM-5.2" {
		t.Fatalf("first model = %+v", models[0])
	}
	if models[0].ContextLimit != 200000 || models[0].Multimodal {
		t.Fatalf("glm-5.2 metadata = %+v", models[0])
	}
	// max wins over dev when both are published.
	if models[1].ID != "kimi-k3" || models[1].ContextLimit != 256000 || !models[1].Multimodal {
		t.Fatalf("kimi-k3 metadata = %+v", models[1])
	}
	// An empty display_name falls back to the routing key.
	if models[2].ID != "custom_model_placeholder" || models[2].DisplayName != "custom_model_placeholder" {
		t.Fatalf("placeholder fallback = %+v", models[2])
	}
}

func TestParseDetailParamRejectsGatewayErrorCode(t *testing.T) {
	_, errParse := ParseDetailParam([]byte(`{"code":1001,"message":"not able to authenticate you","config_info_list":[]}`))
	if errParse == nil {
		t.Fatal("expected a non-zero gateway code to be an error")
	}
}

// TestFetchCatalogueSendsDesktopIdentity pins the request the gateway accepts:
// GET /api/ide/v1/model_list with the Cloud-IDE-JWT auth triplet and the
// desktop identification headers.
func TestFetchCatalogueSendsDesktopIdentity(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	var gotHeader http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod, gotHeader = r.URL.Path, r.Method, r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = w.Write([]byte(`{"code":0,"config_info_list":[{"config_name":"glm-5.3","config_switch":true,"display_config":{"display_name":"GLM-5.3"}}]}`))
	}))
	defer server.Close()

	client := NewModelClient(nil, "")
	fetched, errFetch := client.FetchCatalogue(context.Background(), server.URL, AgentHeaders{
		AccessToken: "jwt-token",
		UID:         "3273661518457721",
		MachineID:   "machine-1",
		DeviceID:    "device-1",
	})
	if errFetch != nil {
		t.Fatalf("FetchCatalogue: %v", errFetch)
	}
	if len(fetched.Models) != 1 || fetched.Models[0].ID != "glm-5.3" {
		t.Fatalf("models = %+v", fetched.Models)
	}
	if !strings.Contains(string(fetched.Raw), "config_info_list") {
		t.Fatalf("raw payload not preserved: %q", fetched.Raw)
	}
	if gotMethod != http.MethodPost || gotPath != DetailParamPath {
		t.Fatalf("request = %s %s, want POST %s", gotMethod, gotPath, DetailParamPath)
	}
	if !strings.Contains(gotBody, `"function":"solo_work_lite"`) || !strings.Contains(gotBody, `"need_prompt":false`) {
		t.Fatalf("request body = %s, want solo_work_lite with need_prompt=false", gotBody)
	}
	for header, want := range map[string]string{
		"Authorization":        "Cloud-IDE-JWT jwt-token",
		"X-Ide-Token":          "jwt-token",
		"X-Uid":                "3273661518457721",
		"X-App-Id":             AppID,
		"X-Ide-Version":        IdeVersion,
		"X-Ide-Version-Code":   IdeVersionCode,
		"X-Machine-Id":         "machine-1",
		"X-Device-Id":          "device-1",
		"Request-Traffic-Type": "prod",
	} {
		if got := gotHeader.Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestFetchCatalogueSurfacesHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":1001,"message":"unauthenticated"}`))
	}))
	defer server.Close()

	client := NewModelClient(nil, "")
	if _, errFetch := client.FetchCatalogue(context.Background(), server.URL, AgentHeaders{AccessToken: "expired"}); errFetch == nil {
		t.Fatal("expected a 401 to be an error")
	}
}

// TestParseDetailParamAcceptsFlatContextWindow keeps older payloads working: the
// window may arrive as a plain number instead of the {dev,max} object the
// desktop client renders.
func TestParseDetailParamAcceptsFlatContextWindow(t *testing.T) {
	models, errParse := ParseDetailParam([]byte(`{"code":0,"config_info_list":[
		{"config_name":"glm-5.2","config_switch":true,"context_window_tokens":180000},
		{"config_name":"kimi-k3","config_switch":true,"max_tokens":150000}
	]}`))
	if errParse != nil {
		t.Fatalf("ParseDetailParam: %v", errParse)
	}
	if len(models) != 2 {
		t.Fatalf("models = %+v", models)
	}
	if models[0].ContextLimit != 180000 {
		t.Fatalf("flat context_window_tokens = %d, want 180000", models[0].ContextLimit)
	}
	if models[1].ContextLimit != 150000 {
		t.Fatalf("max_tokens fallback = %d, want 150000", models[1].ContextLimit)
	}
}
