package trae

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	log "github.com/sirupsen/logrus"
)

// SOLO agent gateway constants recovered from TRAE_CN-Setup-x64.exe 3.4.17
// (product.json bootConfig.agent.trae and app/modules/ai-agent/ai_agent.dll
// route table).
const (
	// AppID is the desktop IDE application id sent as X-App-Id.
	AppID = "6eefa01c-1036-4c7e-9ca5-d891f63bfcd8"

	// IdeVersionCode is the desktop IDE build code sent as X-Ide-Version-Code.
	IdeVersionCode = "20260901"

	// DetailParamPath lists the models a chat function may call. This is the
	// only endpoint that returns the config_name values the SOLO chat channel
	// accepts, together with per-model context window and display metadata.
	//
	// Verified against the live gateway: /api/ide/v1/model_list also exists but
	// requires type=chat|builder and serves IDE editor models (name=seed_m8, …)
	// which the solo_work_lite channel rejects with 4001 "param is invalid".
	DetailParamPath = "/api/ide/v1/get_detail_param"

	// FunctionSoloWorkLite is the SOLO free-chat function whose model set this
	// provider serves.
	FunctionSoloWorkLite = "solo_work_lite"
)

// modelRequestTimeout bounds credential-authenticated catalogue reads. It is
// not applied to inference.
const modelRequestTimeout = 30 * time.Second

// AgentHeaders carries the desktop identity the SOLO gateway authenticates.
type AgentHeaders struct {
	AccessToken string
	UID         string
	MachineID   string
	DeviceID    string
}

// ApplyAgentHeaders sets the auth and content-negotiation headers shared by the
// SOLO agent endpoints.
func ApplyAgentHeaders(req *http.Request, id AgentHeaders, stream bool) {
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("User-Agent", "Trae/"+IdeVersion)
	if token := strings.TrimSpace(id.AccessToken); token != "" {
		req.Header.Set("Authorization", "Cloud-IDE-JWT "+token)
		req.Header.Set("X-Cloudide-Token", token)
		req.Header.Set("X-Ide-Token", token)
	}
	if uid := strings.TrimSpace(id.UID); uid != "" {
		req.Header.Set("X-Uid", uid)
	}
}

// ApplyDesktopHeaders sets the desktop-IDE identification headers.
func ApplyDesktopHeaders(req *http.Request, id AgentHeaders) {
	req.Header.Set("X-App-Id", AppID)
	req.Header.Set("X-App-Version", "default")
	req.Header.Set("X-Ide-Version", IdeVersion)
	req.Header.Set("X-Ide-Version-Code", IdeVersionCode)
	req.Header.Set("X-App-Version-Code", IdeVersionCode)
	req.Header.Set("X-Ide-Version-Type", "stable")
	req.Header.Set("X-Device-Type", "windows")
	req.Header.Set("X-OS-Version", "Windows 11 Pro")
	req.Header.Set("X-Device-Brand", "83DG")
	req.Header.Set("Request-Traffic-Type", "prod")
	if machineID := strings.TrimSpace(id.MachineID); machineID != "" {
		req.Header.Set("X-Machine-Id", machineID)
	}
	if deviceID := strings.TrimSpace(id.DeviceID); deviceID != "" {
		req.Header.Set("X-Device-Id", deviceID)
	}
}

// Model is one routable model from the gateway catalogue. ID is the upstream
// config_name the chat endpoint expects.
type Model struct {
	ID           string
	DisplayName  string
	ContextLimit int
	MaxTokens    int
	Multimodal   bool
}

// ModelClient reads the account's model catalogue from the SOLO agent gateway.
type ModelClient struct {
	httpClient *http.Client
}

// NewModelClient creates a proxy-aware catalogue client.
func NewModelClient(cfg *config.Config, proxyURL string) *ModelClient {
	httpClient := &http.Client{Timeout: modelRequestTimeout}
	var sdkCfg config.SDKConfig
	if cfg != nil {
		sdkCfg = cfg.SDKConfig
	}
	if trimmed := strings.TrimSpace(proxyURL); trimmed != "" {
		sdkCfg.ProxyURL = trimmed
	}
	httpClient = util.SetProxy(&sdkCfg, httpClient)
	return &ModelClient{httpClient: httpClient}
}

// FetchResult is the parsed catalogue plus the payload it was parsed from. The
// raw body is kept so callers can emit a one-shot schema report against the
// real upstream response, which is the only way to confirm field names.
type FetchResult struct {
	Models []Model
	Raw    []byte
}

// detailParamRequest is the get_detail_param body. need_prompt=false keeps the
// response to model metadata: with prompts enabled the gateway inlines the
// encrypted prompt set, which this provider never uses.
type detailParamRequest struct {
	Function   string `json:"function"`
	NeedPrompt bool   `json:"need_prompt"`
}

// FetchCatalogue reads the catalogue authorized for this credential.
func (c *ModelClient) FetchCatalogue(ctx context.Context, baseURL string, id AgentHeaders) (*FetchResult, error) {
	if strings.TrimSpace(id.AccessToken) == "" {
		return nil, fmt.Errorf("trae: access token is required")
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = AgentHost
	}
	payload, errMarshal := json.Marshal(detailParamRequest{Function: FunctionSoloWorkLite, NeedPrompt: false})
	if errMarshal != nil {
		return nil, errMarshal
	}
	request, errRequest := http.NewRequestWithContext(ctx, http.MethodPost, base+DetailParamPath, bytes.NewReader(payload))
	if errRequest != nil {
		return nil, errRequest
	}
	ApplyAgentHeaders(request, id, false)
	ApplyDesktopHeaders(request, id)
	resp, errDo := c.httpClient.Do(request)
	if errDo != nil {
		return nil, fmt.Errorf("trae: model list request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("trae: close model list body: %v", errClose)
		}
	}()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if errRead != nil {
		return nil, fmt.Errorf("trae: read model list: %w", errRead)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("trae: model list HTTP %d: %s", resp.StatusCode, summarizeBody(raw))
	}
	models, errParse := ParseDetailParam(raw)
	if errParse != nil {
		return nil, errParse
	}
	return &FetchResult{Models: models, Raw: raw}, nil
}

// ParseDetailParam decodes a get_detail_param response.
//
// Live payload shape (solo_work_lite, verified against the gateway):
//
//	{"code":0,"config_info_list":[{"config_name":"glm-5.2",
//	   "context_window_tokens":{"dev":200000},"config_switch":true,
//	   "is_invisible_to_user":false,"display_config":{"display_name":"GLM-5.2",
//	   "multimodal":false,"model_capability":"reasoning_model","max_mode":false}}]}
//
// config_name is the routing key the SOLO chat channel accepts; display_name and
// window metadata live under display_config / context_window_tokens.
func ParseDetailParam(raw []byte) ([]Model, error) {
	var envelope struct {
		Code           json.Number       `json:"code"`
		Message        string            `json:"message"`
		ConfigInfoList []json.RawMessage `json:"config_info_list"`
	}
	if errUnmarshal := json.Unmarshal(bytes.TrimSpace(raw), &envelope); errUnmarshal != nil {
		return nil, fmt.Errorf("trae: decode detail param: %w", errUnmarshal)
	}
	if code := strings.TrimSpace(envelope.Code.String()); code != "" && code != "0" {
		return nil, fmt.Errorf("trae: detail param error (code %s): %s", code, strings.TrimSpace(envelope.Message))
	}
	models := make([]Model, 0, len(envelope.ConfigInfoList))
	seen := make(map[string]struct{}, len(envelope.ConfigInfoList))
	for _, entry := range envelope.ConfigInfoList {
		model, ok := parseConfigInfo(entry)
		if !ok {
			continue
		}
		if _, dup := seen[model.ID]; dup {
			continue
		}
		seen[model.ID] = struct{}{}
		models = append(models, model)
	}
	return models, nil
}

// parseConfigInfo maps one config_info_list entry.
func parseConfigInfo(raw json.RawMessage) (Model, bool) {
	var fields map[string]any
	if errUnmarshal := json.Unmarshal(raw, &fields); errUnmarshal != nil || len(fields) == 0 {
		return Model{}, false
	}
	// config_switch=false disables the entry; is_invisible_to_user hides internal
	// helper models (computer_use_subagent, custom_model_placeholder, …) that the
	// desktop client drops from its picker.
	if truthy(fields["is_invisible_to_user"]) {
		return Model{}, false
	}
	if flag, ok := fields["config_switch"].(bool); ok && !flag {
		return Model{}, false
	}
	id := firstString(fields, "config_name")
	if id == "" {
		return Model{}, false
	}
	model := Model{ID: id, DisplayName: id}
	display, _ := fields["display_config"].(map[string]any)
	if nested := firstString(display, "display_name"); nested != "" {
		model.DisplayName = nested
	}
	model.ContextLimit = firstContextLimit(fields)
	model.MaxTokens = firstInt(fields, "max_tokens")
	model.Multimodal = truthy(display["multimodal"])
	return model, true
}

// firstContextLimit reads the context window from the catalogue row. The desktop
// client renders it as a nested object and prefers "max" when Max mode is on
// (`context_window_tokens?.max ?? context_window_tokens?.dev`), so the widest
// published value is the honest ContextLength. Live rows publish dev only, so
// both variants are read.
func firstContextLimit(fields map[string]any) int {
	for _, key := range []string{"context_window_tokens", "context_window_size"} {
		if value := nestedInt(fields[key], "max", "dev", "default"); value > 0 {
			return value
		}
	}
	// max_tokens is the prompt budget when no window object is published.
	return firstInt(fields, "max_tokens")
}

// nestedInt accepts either a plain number or an object keyed by variant.
func nestedInt(value any, keys ...string) int {
	switch typed := value.(type) {
	case float64:
		if typed > 0 {
			return int(typed)
		}
	case map[string]any:
		for _, key := range keys {
			if nested, ok := typed[key].(float64); ok && nested > 0 {
				return int(nested)
			}
		}
	}
	return 0
}

func firstString(fields map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := fields[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func firstInt(fields map[string]any, keys ...string) int {
	for _, key := range keys {
		if value, ok := fields[key].(float64); ok && value > 0 {
			return int(value)
		}
	}
	return 0
}

func truthy(value any) bool {
	flag, ok := value.(bool)
	return ok && flag
}

func summarizeBody(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if len(text) > 512 {
		text = text[:512]
	}
	return text
}
