package codebuddycn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Catalog endpoints. The CLI addresses both from the origin, never from the
// /v2 chat-gateway base URL a credential stores.
const (
	// ConfigPath returns the whole cloud product configuration. The model list
	// is one field of it, alongside agents, productFeatures and links.
	ConfigPath = "/v3/config"
	// AccountPath reports the signed-in account. It is the only reliable way to
	// tell "expired token" from "empty catalogue", because /v3/config answers
	// HTTP 200 with a null model list for a revoked token.
	AccountPath = "/v2/plugin/account"
)

// Product lines select which catalogue the gateway returns. The prefix, not the
// version, chooses the product: a missing or malformed User-Agent is rejected
// outright with code 12403.
const (
	// ProductCLI is the CLI catalogue, and the one the CLI itself advertises.
	ProductCLI = "cli"
	// ProductIDE is the WorkBuddy/IDE catalogue, which adds IDE-only models.
	ProductIDE = "ide"
)

// CatalogErrorCodeUA is the envelope code returned when the User-Agent is
// absent or unparseable.
const CatalogErrorCodeUA = 12403

// Version constants for the two product lines. The gateway only parses the
// version shape, so these track the CLI build the endpoints were confirmed
// against rather than needing to match the installed client.
const (
	// CLIVersion is the CodeBuddy CLI version used in the CLI User-Agent.
	CLIVersion = "2.63.2"
	// IDEVersions is the WorkBuddy/IDE version used in the IDE User-Agent.
	IDEVersions = "5.3.14"
)

// UserAgentFor builds the product-line User-Agent. The version is opaque to the
// gateway; any x.y.z parses.
func UserAgentFor(product, version string) string {
	trimmed := strings.TrimSpace(version)
	switch strings.ToLower(strings.TrimSpace(product)) {
	case ProductIDE:
		if trimmed == "" {
			trimmed = IDEVersions
		}
		return "WorkBuddy/" + trimmed
	default:
		if trimmed == "" {
			trimmed = CLIVersion
		}
		return "CLI/" + trimmed + " CodeBuddy/" + trimmed
	}
}

// CatalogModel is one chat entry of the cloud product catalogue.
type CatalogModel struct {
	// ID is the upstream model identifier used on the wire.
	ID string
	// Name is the display name.
	Name string
	// ContextWindow is maxInputTokens.
	ContextWindow int
	// MaxOutputTokens is maxOutputTokens.
	MaxOutputTokens int
	// SupportsImages marks multimodal models.
	SupportsImages bool
	// SupportsToolCall reports tool/function calling support.
	SupportsToolCall bool
	// ReasoningLevels lists the accepted thinking efforts, empty when the model
	// exposes no choice.
	ReasoningLevels []string
	// CanDisableThinking reports whether thinking can be turned off.
	CanDisableThinking bool
	// Tags are the upstream labels (used to detect generation models).
	Tags []string
}

// Catalog is the resolved model catalogue for one account.
type Catalog struct {
	// Models are the chat models, in upstream order.
	Models []CatalogModel
	// AgentNames lists the agents the catalogue advertises, for diagnostics.
	AgentNames []string
	// CLIModelIDs are the ids the CLI agent advertises. The full list is wider:
	// it also carries image/video generators and models only exposed to
	// internal sub-agents.
	CLIModelIDs []string
}

// FetchCatalog reads the account's cloud product catalogue.
//
// It calls /v2/plugin/account first because /v3/config cannot distinguish an
// expired token from an empty catalogue: a revoked credential answers HTTP 200
// with models == null, which is identical to the anonymous response.
func (c *Client) FetchCatalog(ctx context.Context, accessToken, product, version string) (*Catalog, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	token := strings.TrimSpace(accessToken)
	if token == "" {
		return nil, fmt.Errorf("codebuddy: no access token")
	}
	origin := c.catalogOrigin()

	if errAccount := c.checkAccount(ctx, origin, token, UserAgentFor(product, version)); errAccount != nil {
		return nil, errAccount
	}

	payload, errConfig := c.getEnvelope(ctx, origin, ConfigPath, token, product, version, nil)
	if errConfig != nil {
		return nil, errConfig
	}
	data, _ := payload["data"].(map[string]any)
	if data == nil {
		return nil, fmt.Errorf("codebuddy: configuration response has no data object")
	}
	rawModels, present := data["models"]
	if !present || rawModels == nil {
		// Documented as the revoked-credential shape. Reaching here means the
		// account probe passed and the catalogue was still null, so the account
		// is readable but the catalogue is genuinely unavailable.
		return nil, fmt.Errorf("codebuddy: catalogue is null for a valid account")
	}
	rows, ok := rawModels.([]any)
	if !ok {
		return nil, fmt.Errorf("codebuddy: configuration response has no model list")
	}

	catalog := &Catalog{
		Models:      make([]CatalogModel, 0, len(rows)),
		AgentNames:  agentNames(data),
		CLIModelIDs: agentModelIDs(data, ProductCLI),
	}
	for _, row := range rows {
		entry, ok := row.(map[string]any)
		if !ok {
			continue
		}
		model, ok := catalogModelFrom(entry)
		if !ok {
			continue
		}
		catalog.Models = append(catalog.Models, model)
	}
	return catalog, nil
}

// catalogOrigin returns scheme://host for the catalogue endpoints. Credentials
// store base_url with a /v2 suffix for the chat gateway; the catalogue is
// addressed from the origin, which is derived from the auth host the client was
// built with.
func (c *Client) catalogOrigin() string {
	if trimmed := strings.TrimSpace(c.catalogOriginOverride); trimmed != "" {
		return trimmed
	}
	// stateURL is the per-gateway anchor (CN or international), so deriving the
	// origin from it keeps both editions correct without a second setting.
	for _, candidate := range []string{c.stateURL, c.tokenURL} {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		if !strings.Contains(trimmed, "://") {
			trimmed = "https://" + trimmed
		}
		parsed, errParse := url.Parse(trimmed)
		if errParse != nil || parsed.Host == "" {
			continue
		}
		return parsed.Scheme + "://" + parsed.Host
	}
	if strings.TrimSpace(c.domain) != "" {
		return "https://" + c.domain
	}
	return "https://copilot.tencent.com"
}

// checkAccount probes credential validity.
func (c *Client) checkAccount(ctx context.Context, origin, token, userAgent string) error {
	payload, errEnvelope := c.getEnvelope(ctx, origin, AccountPath, token, "", "", map[string]string{
		"User-Agent": userAgent,
	})
	if errEnvelope != nil {
		return fmt.Errorf("codebuddy: credential check failed: %w", errEnvelope)
	}
	if data, _ := payload["data"].(map[string]any); data == nil {
		return fmt.Errorf("codebuddy: account response has no data object")
	}
	return nil
}

// getEnvelope performs an authenticated GET and validates the envelope code.
func (c *Client) getEnvelope(ctx context.Context, origin, path, token, product, version string, extraHeaders map[string]string) (map[string]any, error) {
	endpoint := strings.TrimRight(origin, "/") + path
	req, errReq := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if errReq != nil {
		return nil, fmt.Errorf("codebuddy: create request: %w", errReq)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Domain", c.domain)
	if product != "" {
		req.Header.Set("User-Agent", UserAgentFor(product, version))
	} else if strings.TrimSpace(req.Header.Get("User-Agent")) == "" {
		req.Header.Set("User-Agent", UserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}

	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("codebuddy: request failed: %w", errDo)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, errRead := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if errRead != nil {
		return nil, fmt.Errorf("codebuddy: read response: %w", errRead)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("codebuddy: gateway rejected the credential (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("codebuddy: gateway rejected the request (HTTP 400); check the User-Agent product line")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("codebuddy: unexpected HTTP status %d", resp.StatusCode)
	}

	var payload map[string]any
	if errUnmarshal := json.Unmarshal(body, &payload); errUnmarshal != nil {
		return nil, fmt.Errorf("codebuddy: invalid JSON response: %w", errUnmarshal)
	}
	if code, ok := envelopeCode(payload); ok && code != 0 {
		if code == CatalogErrorCodeUA {
			return nil, fmt.Errorf("codebuddy: gateway rejected the User-Agent (%d)", code)
		}
		return nil, fmt.Errorf("codebuddy: upstream business error %d", code)
	}
	return payload, nil
}

// envelopeCode reads the numeric envelope code, tolerating a string.
func envelopeCode(payload map[string]any) (int64, bool) {
	switch value := payload["code"].(type) {
	case float64:
		return int64(value), true
	case string:
		var parsed int64
		if _, errScan := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); errScan == nil {
			return parsed, true
		}
	}
	return 0, false
}

// agentNames lists the catalogue's agent names.
func agentNames(data map[string]any) []string {
	agents, _ := data["agents"].([]any)
	names := make([]string, 0, len(agents))
	for _, agent := range agents {
		entry, ok := agent.(map[string]any)
		if !ok {
			continue
		}
		if name, _ := entry["name"].(string); strings.TrimSpace(name) != "" {
			names = append(names, strings.TrimSpace(name))
		}
	}
	return names
}

// agentModelIDs returns the model ids advertised by one agent.
func agentModelIDs(data map[string]any, agentName string) []string {
	agents, _ := data["agents"].([]any)
	for _, agent := range agents {
		entry, ok := agent.(map[string]any)
		if !ok {
			continue
		}
		if name, _ := entry["name"].(string); !strings.EqualFold(strings.TrimSpace(name), agentName) {
			continue
		}
		raw, _ := entry["models"].([]any)
		ids := make([]string, 0, len(raw))
		for _, item := range raw {
			switch value := item.(type) {
			case string:
				if trimmed := strings.TrimSpace(value); trimmed != "" {
					ids = append(ids, trimmed)
				}
			case map[string]any:
				if id, _ := value["id"].(string); strings.TrimSpace(id) != "" {
					ids = append(ids, strings.TrimSpace(id))
				}
			}
		}
		return ids
	}
	return nil
}

// catalogModelFrom maps one upstream row, rejecting entries without an id.
func catalogModelFrom(row map[string]any) (CatalogModel, bool) {
	id, _ := row["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		return CatalogModel{}, false
	}
	model := CatalogModel{
		ID:               id,
		Name:             id,
		ContextWindow:    numericField(row["maxInputTokens"]),
		MaxOutputTokens:  numericField(row["maxOutputTokens"]),
		SupportsImages:   boolField(row, "supportsImages"),
		SupportsToolCall: boolField(row, "supportsToolCall"),
	}
	if name, _ := row["name"].(string); strings.TrimSpace(name) != "" {
		model.Name = strings.TrimSpace(name)
	}
	if reasoning, ok := row["reasoning"].(map[string]any); ok {
		if efforts, ok := reasoning["supportedEfforts"].([]any); ok {
			for _, effort := range efforts {
				if text, _ := effort.(string); strings.TrimSpace(text) != "" {
					model.ReasoningLevels = append(model.ReasoningLevels, strings.TrimSpace(text))
				}
			}
		}
		if len(model.ReasoningLevels) == 0 {
			if single, _ := reasoning["effort"].(string); strings.TrimSpace(single) != "" {
				model.ReasoningLevels = []string{strings.TrimSpace(single)}
			}
		}
		model.CanDisableThinking = boolField(reasoning, "canDisableThinking")
	}
	if tags, ok := row["tags"].([]any); ok {
		for _, tag := range tags {
			if text, _ := tag.(string); strings.TrimSpace(text) != "" {
				model.Tags = append(model.Tags, strings.TrimSpace(text))
			}
		}
	}
	return model, true
}

// numericField reads a non-negative number from an untyped JSON value.
func numericField(value any) int {
	switch typed := value.(type) {
	case float64:
		if typed > 0 {
			return int(typed)
		}
	case int:
		if typed > 0 {
			return typed
		}
	case int64:
		if typed > 0 {
			return int(typed)
		}
	}
	return 0
}

// boolField reads a boolean member of an untyped JSON object.
func boolField(row map[string]any, key string) bool {
	value, _ := row[key].(bool)
	return value
}

// NewTestClientAt builds a client whose auth anchors point at an explicit
// origin, so the catalogue path can be exercised against a test server.
func NewTestClientAt(origin string) *Client {
	trimmed := strings.TrimRight(strings.TrimSpace(origin), "/")
	return newClient(http.DefaultClient, trimmed+"/v2/plugin/auth/state",
		trimmed+"/v2/plugin/auth/token", trimmed+"/v2/plugin/auth/token/refresh", "copilot.tencent.com")
}

// WithOrigin overrides the origin the catalogue is read from.
//
// A credential already carries base_url (the chat gateway, e.g.
// .../v2), and the catalogue lives at that host's origin. Pinning it here also
// lets a credential target a staging or mirror deployment.
func (c *Client) WithOrigin(origin string) *Client {
	if c == nil {
		return nil
	}
	if trimmed := strings.TrimRight(strings.TrimSpace(origin), "/"); trimmed != "" {
		c.catalogOriginOverride = trimmed
	}
	return c
}
