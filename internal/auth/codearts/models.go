package codearts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Model is one entry of the CodeArts model catalog.
//
// The upstream catalog carries both the platform builtin models
// (/v1/model/builtin) and the CodeAgent model list
// (/v1/agent-center/agents/detail); both are normalized into this shape.
type Model struct {
	// ID is the upstream model identifier used on the wire.
	ID string `json:"model_id"`
	// Name is the display name.
	Name string `json:"model_name"`
	// Description is the localized model description.
	Description string `json:"model_desc,omitempty"`
	// Provider is the upstream provider key (always "inferhub-provider").
	Provider string `json:"provider,omitempty"`
	// ModelType distinguishes "custom" models from platform ones.
	ModelType string `json:"model_type,omitempty"`
	// Category is the upstream model_category (图片理解 marks multimodal models).
	Category string `json:"model_category,omitempty"`
	// ContextWindow is the total context window (limit.context upstream).
	ContextWindow int `json:"context_window,omitempty"`
	// InputContextWindow is the input context window (limit.input upstream).
	InputContextWindow int `json:"input_context_window,omitempty"`
	// OutputContextWindow is the output window (limit.output upstream). It is the
	// value sent as max_tokens on chat requests.
	OutputContextWindow int `json:"output_context_window,omitempty"`
	// MaxTokens is the upstream max_tokens fallback for OutputContextWindow.
	MaxTokens int `json:"max_tokens,omitempty"`
	// InputLength is the upstream input_length field.
	InputLength int `json:"input_length,omitempty"`
	// TruncateLength is the upstream truncate_length field.
	TruncateLength int `json:"truncate_length,omitempty"`
	// SupportsImages marks multimodal models.
	SupportsImages bool `json:"supports_images,omitempty"`
	// ThinkLevel is the default thinking strength (low/medium/high/xhigh/max).
	ThinkLevel string `json:"think_level,omitempty"`
	// IsLuxury marks priority (尊享) models.
	IsLuxury bool `json:"is_luxury,omitempty"`
	// EnableQueue marks models that queue instead of failing.
	EnableQueue bool `json:"enable_queue,omitempty"`
	// DisplayEnabled reports whether the catalog entry should be surfaced.
	DisplayEnabled *bool `json:"display_enabled,omitempty"`
	// ModelMode is the upstream model_mode. The value "fastmode" marks a fast
	// model, which the desktop client mirrors into the `x-mode-type` request
	// header. It lives at the model's top level, not inside model_parameters,
	// so expanding the parameter block drops it unless carried explicitly.
	ModelMode string `json:"model_mode,omitempty"`
	// IsFreeBenefit marks the limited-time free models from gateway/config.
	IsFreeBenefit bool `json:"is_free_benefit,omitempty"`
	// BenefitChannel is the maas_type value the model must be requested with.
	// It is "benefit" for the free pool and "education" for the education
	// entitlement; empty means the model needs no entitlement routing.
	BenefitChannel string `json:"maas_type,omitempty"`
	// BaseURL overrides the LLM gateway for free-benefit models.
	BaseURL string `json:"base_url,omitempty"`
}

// builtinModelsResponse mirrors GET /v1/model/builtin.
type builtinModelsResponse struct {
	BuiltinModels []json.RawMessage `json:"builtinModels"`
	ErrorCode     string            `json:"error_code"`
	ErrorMsg      string            `json:"error_msg"`
}

// FetchBuiltinModels lists the platform model catalog.
func (c *Client) FetchBuiltinModels(ctx context.Context, creds Credentials, language string) ([]*Model, error) {
	endpoint := strings.TrimRight(c.snapHost, "/") + BuiltinModelsPath
	raw, errFetch := c.signedGetJSON(ctx, endpoint, creds, map[string]string{
		"Agent-Type": AgentTypePromptCenter,
		"X-Language": defaultLanguage(language),
	})
	if errFetch != nil {
		return nil, errFetch
	}

	var parsed builtinModelsResponse
	if errUnmarshal := json.Unmarshal(raw, &parsed); errUnmarshal != nil {
		return nil, fmt.Errorf("codearts: invalid builtin model response: %w", errUnmarshal)
	}
	if parsed.ErrorCode != "" && parsed.ErrorCode != "0000" {
		return nil, &Error{StatusCode: http.StatusOK, Code: parsed.ErrorCode, Message: parsed.ErrorMsg}
	}
	if len(parsed.BuiltinModels) == 0 {
		return nil, fmt.Errorf("codearts: builtin model response has no models")
	}

	out := make([]*Model, 0, len(parsed.BuiltinModels))
	for _, entry := range parsed.BuiltinModels {
		var model Model
		if errUnmarshal := json.Unmarshal(entry, &model); errUnmarshal != nil {
			continue
		}
		normalized, ok := normalizeModel(&model)
		if !ok {
			continue
		}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("codearts: no usable builtin models")
	}
	return out, nil
}

// FetchAgentModels lists the models attached to the primary CodeAgent.
//
// The desktop client resolves the CodeAgent id from
// /v1/agent-center/agents/useragents and then reads its model list from
// /v1/agent-center/agents/detail.
func (c *Client) FetchAgentModels(ctx context.Context, creds Credentials, language string) ([]*Model, error) {
	base := strings.TrimRight(c.snapHost, "/")
	agentsEndpoint := base + "/v1/agent-center/agents/useragents?offset=0&limit=100&is_primary_agent=true"
	rawAgents, errAgents := c.signedGetJSON(ctx, agentsEndpoint, creds, map[string]string{
		"Agent-Type": "AgentCenter",
		"X-Language": defaultLanguage(language),
	})
	if errAgents != nil {
		return nil, errAgents
	}

	var agents struct {
		Agents []struct {
			AgentID string `json:"agent_id"`
			Name    string `json:"agent_name"`
			Alias   struct {
				ZhCN string `json:"alias_zh_cn"`
			} `json:"alias"`
			ShowInIDE bool `json:"show_in_ide"`
		} `json:"agents"`
	}
	if errUnmarshal := json.Unmarshal(rawAgents, &agents); errUnmarshal != nil {
		return nil, fmt.Errorf("codearts: invalid agent list response: %w", errUnmarshal)
	}

	agentID := ""
	for _, agent := range agents.Agents {
		if strings.EqualFold(strings.TrimSpace(agent.Name), "CodeAgent") && agent.ShowInIDE {
			agentID = strings.TrimSpace(agent.AgentID)
			break
		}
	}
	if agentID == "" && len(agents.Agents) > 0 {
		agentID = strings.TrimSpace(agents.Agents[0].AgentID)
	}
	if agentID == "" {
		return nil, fmt.Errorf("codearts: no primary CodeAgent found")
	}

	detailEndpoint := base + "/v1/agent-center/agents/detail?agent_id=" + agentID
	rawDetail, errDetail := c.signedGetJSON(ctx, detailEndpoint, creds, map[string]string{
		"Agent-Type": "AgentCenter",
		"X-Language": defaultLanguage(language),
	})
	if errDetail != nil {
		return nil, errDetail
	}

	var detail struct {
		GPTs struct {
			Models []struct {
				ModelName       string          `json:"model_name"`
				ModelType       string          `json:"model_type"`
				Credit          json.RawMessage `json:"credit"`
				UsageTokenNum   int             `json:"usage_token_num"`
				ModelQuotaNum   int             `json:"model_quota_num"`
				ThinkLevel      string          `json:"think_level"`
				ModelMode       string          `json:"model_mode"`
				ModelParameters struct {
					ModelID             string `json:"model_id"`
					ContextWindow       int    `json:"context_window"`
					InputContextWindow  int    `json:"input_context_window"`
					OutputContextWindow int    `json:"output_context_window"`
					MaxTokens           int    `json:"max_tokens"`
					InputLength         int    `json:"input_length"`
					SupportsImages      bool   `json:"supports_images"`
				} `json:"model_parameters"`
			} `json:"models"`
		} `json:"gpts"`
	}
	if errUnmarshal := json.Unmarshal(rawDetail, &detail); errUnmarshal != nil {
		return nil, fmt.Errorf("codearts: invalid agent detail response: %w", errUnmarshal)
	}

	out := make([]*Model, 0, len(detail.GPTs.Models))
	seen := make(map[string]struct{}, len(detail.GPTs.Models))
	for _, entry := range detail.GPTs.Models {
		model := &Model{
			ID:                  strings.TrimSpace(entry.ModelParameters.ModelID),
			Name:                strings.TrimSpace(entry.ModelName),
			ContextWindow:       entry.ModelParameters.ContextWindow,
			InputContextWindow:  entry.ModelParameters.InputContextWindow,
			OutputContextWindow: entry.ModelParameters.OutputContextWindow,
			MaxTokens:           entry.ModelParameters.MaxTokens,
			InputLength:         entry.ModelParameters.InputLength,
			SupportsImages:      entry.ModelParameters.SupportsImages,
			ThinkLevel:          strings.TrimSpace(entry.ThinkLevel),
			ModelType:           strings.TrimSpace(entry.ModelType),
			ModelMode:           strings.TrimSpace(entry.ModelMode),
		}
		normalized, ok := normalizeModel(model)
		if !ok {
			continue
		}
		// model_mode selects the fast channel for this model. Record it while the
		// catalogue is in hand: the routing decision is made later from the
		// request body alone, with no catalogue available.
		if fastID := normalized.ID; strings.EqualFold(normalized.ModelMode, FastModelMode) {
			RegisterFastModel(fastID)
		}
		if _, exists := seen[normalized.ID]; exists {
			continue
		}
		seen[normalized.ID] = struct{}{}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("codearts: agent has no usable models")
	}
	return out, nil
}

// FetchFreeBenefitModels lists the limited-time free models from the developer
// gateway. It returns an empty slice (no error) when the program is inactive.
func (c *Client) FetchFreeBenefitModels(ctx context.Context, creds Credentials) ([]*Model, error) {
	endpoint := strings.TrimRight(c.benefitURL, "/") + GatewayConfigPath
	raw, errFetch := c.signedGetJSON(ctx, endpoint, creds, map[string]string{"X-Language": "zh-cn"})
	if errFetch != nil {
		return nil, errFetch
	}

	var parsed struct {
		ErrorCode string `json:"error_code"`
		Result    struct {
			BaseURL string `json:"base_url"`
			Models  []struct {
				ModelID        string `json:"model_id"`
				ModelName      string `json:"model_name"`
				Description    string `json:"description"`
				ModelDesc      string `json:"model_desc"`
				ContextWindow  int    `json:"context_window"`
				MaxTokens      int    `json:"max_tokens"`
				TruncateLength int    `json:"truncate_length"`
			} `json:"models"`
		} `json:"result"`
	}
	if errUnmarshal := json.Unmarshal(raw, &parsed); errUnmarshal != nil {
		return nil, fmt.Errorf("codearts: invalid gateway/config response: %w", errUnmarshal)
	}
	if parsed.ErrorCode != "0000" {
		return nil, nil
	}

	baseURL := strings.TrimSpace(parsed.Result.BaseURL)
	out := make([]*Model, 0, len(parsed.Result.Models))
	for _, entry := range parsed.Result.Models {
		id := firstNonEmpty(entry.ModelID, entry.ModelName)
		if id == "" {
			continue
		}
		name := firstNonEmpty(entry.ModelName, entry.ModelID)
		// The gateway lists the pool but not the label a request must carry, so
		// record the route now: without this a benefit model published after this
		// binary shipped would be requested as a normal model and rejected.
		RegisterBenefitRoute(id, BenefitChannelBenefit)
		out = append(out, &Model{
			ID:                  id,
			Name:                name,
			Description:         firstNonEmpty(entry.ModelDesc, entry.Description),
			Provider:            "inferhub-provider",
			ContextWindow:       entry.ContextWindow,
			OutputContextWindow: entry.MaxTokens,
			MaxTokens:           entry.MaxTokens,
			TruncateLength:      entry.TruncateLength,
			IsFreeBenefit:       true,
			BenefitChannel:      BenefitChannelBenefit,
			BaseURL:             baseURL,
		})
	}
	return out, nil
}

// FetchModels gathers the full catalog the desktop client presents: the
// CodeAgent model list, or the platform builtin list as a fallback, merged with
// any limited-time free models.
func (c *Client) FetchModels(ctx context.Context, creds Credentials, language string) ([]*Model, error) {
	models, errModels := c.FetchAgentModels(ctx, creds, language)
	if errModels != nil || len(models) == 0 {
		models, errModels = c.FetchBuiltinModels(ctx, creds, language)
		if errModels != nil {
			return nil, errModels
		}
	}
	if free, errFree := c.FetchFreeBenefitModels(ctx, creds); errFree == nil && len(free) > 0 {
		seen := make(map[string]struct{}, len(models))
		for _, model := range models {
			seen[model.ID] = struct{}{}
		}
		for _, model := range free {
			if _, exists := seen[model.ID]; exists {
				continue
			}
			seen[model.ID] = struct{}{}
			models = append(models, model)
		}
	}
	return models, nil
}

// excludedModelNames mirrors the desktop client's name blacklist. These entries
// exist upstream but must not be offered as selectable models: the client keeps
// them out of the model picker on purpose, so surfacing them here would offer
// something the vendor UI never lets a user choose.
var excludedModelNames = map[string]struct{}{
	"glm-5.1":       {},
	"qwen3-vl-235b": {},
	"maas-glm-4.7":  {},
}

// isThirdPartyModelType reports whether the entry is a third-party hosted model.
// The desktop client drops these before the catalogue ever reaches the kernel.
func isThirdPartyModelType(modelType string) bool {
	return strings.EqualFold(strings.TrimSpace(modelType), "third")
}

// normalizeModel fills defaults and rejects unusable entries.
func normalizeModel(model *Model) (*Model, bool) {
	if model == nil {
		return nil, false
	}
	id := firstNonEmpty(model.ID, model.Name)
	if id == "" {
		return nil, false
	}
	model.ID = id
	if strings.TrimSpace(model.Name) == "" {
		model.Name = id
	}
	// Two filters the desktop client applies before publishing a catalogue:
	// third-party hosted models are never routed through the kernel, and a small
	// name blacklist is hidden from the model picker.
	if isThirdPartyModelType(model.ModelType) {
		return nil, false
	}
	if _, excluded := excludedModelNames[strings.ToLower(strings.TrimSpace(model.Name))]; excluded {
		return nil, false
	}
	if strings.TrimSpace(model.Provider) == "" {
		model.Provider = "inferhub-provider"
	}
	if model.DisplayEnabled != nil && !*model.DisplayEnabled {
		return nil, false
	}
	if model.OutputContextWindow == 0 {
		model.OutputContextWindow = model.MaxTokens
	}
	if model.ContextWindow == 0 {
		model.ContextWindow = model.InputContextWindow
	}
	return model, true
}

// signedGetJSON performs a signed GET and returns the decoded body.
func (c *Client) signedGetJSON(ctx context.Context, endpoint string, creds Credentials, extraHeaders map[string]string) ([]byte, error) {
	headers := map[string]string{"Accept": "application/json"}
	for key, value := range extraHeaders {
		if c.checkinHeaders != nil || strings.TrimSpace(value) != "" {
			headers[key] = value
		}
	}
	req, errReq := c.signedRequest(ctx, http.MethodGet, endpoint, nil, creds, headers)
	if errReq != nil {
		return nil, errReq
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("codearts: request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			return
		}
	}()
	data, errRead := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if errRead != nil {
		return nil, fmt.Errorf("codearts: read response: %w", errRead)
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}
	return data, nil
}

// signedPostJSON performs a signed JSON POST and returns the decoded body.
func (c *Client) signedPostJSON(ctx context.Context, endpoint string, creds Credentials, payload []byte, extraHeaders map[string]string) ([]byte, error) {
	headers := map[string]string{"Content-Type": "application/json", "Accept": "application/json"}
	for key, value := range extraHeaders {
		if c.checkinHeaders != nil || strings.TrimSpace(value) != "" {
			headers[key] = value
		}
	}
	req, errReq := c.signedRequest(ctx, http.MethodPost, endpoint, payload, creds, headers)
	if errReq != nil {
		return nil, errReq
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("codearts: request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			return
		}
	}()
	data, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, fmt.Errorf("codearts: read response: %w", errRead)
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}
	return data, nil
}

// defaultLanguage normalizes the catalog language header.
func defaultLanguage(language string) string {
	if strings.EqualFold(strings.TrimSpace(language), "en-us") {
		return "en-us"
	}
	return "zh-cn"
}
