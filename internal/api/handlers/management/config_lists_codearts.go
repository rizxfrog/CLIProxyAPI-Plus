package management

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

// normalizeCodeArtsKey trims and normalizes one CodeArts credential.
func normalizeCodeArtsKey(entry *config.CodeArtsKey) {
	if entry == nil {
		return
	}
	entry.APIKey = strings.TrimSpace(entry.APIKey)
	entry.SecretKey = strings.TrimSpace(entry.SecretKey)
	entry.SecurityToken = strings.TrimSpace(entry.SecurityToken)
	entry.RefreshToken = strings.TrimSpace(entry.RefreshToken)
	entry.Prefix = strings.TrimSpace(entry.Prefix)
	entry.BaseURL = strings.TrimSpace(entry.BaseURL)
	entry.ProxyURL = strings.TrimSpace(entry.ProxyURL)
	entry.Headers = config.NormalizeHeaders(entry.Headers)
	entry.ExcludedModels = config.NormalizeExcludedModels(entry.ExcludedModels)
	if len(entry.Models) == 0 {
		return
	}
	out := entry.Models[:0]
	seen := make(map[string]struct{}, len(entry.Models))
	for i := range entry.Models {
		model := entry.Models[i]
		model.Name = strings.TrimSpace(model.Name)
		model.Alias = strings.TrimSpace(model.Alias)
		if model.Name == "" && model.Alias == "" {
			continue
		}
		key := model.Name + "|" + model.Alias
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, model)
	}
	entry.Models = out
}

// codeArtsKeyPatch is the PATCH body for one CodeArts credential. It carries the
// extra secret-key/security-token/refresh-token fields the provider needs.
type codeArtsKeyPatch struct {
	APIKey         *string                 `json:"api-key"`
	SecretKey      *string                 `json:"secret-key"`
	SecurityToken  *string                 `json:"security-token"`
	RefreshToken   *string                 `json:"refresh-token"`
	Priority       *int                    `json:"priority"`
	Weight         json.RawMessage         `json:"weight"`
	Prefix         *string                 `json:"prefix"`
	BaseURL        *string                 `json:"base-url"`
	ProxyURL       *string                 `json:"proxy-url"`
	Models         *[]config.CodeArtsModel `json:"models"`
	Headers        *map[string]string      `json:"headers"`
	ExcludedModels *[]string               `json:"excluded-models"`
	DisableCooling json.RawMessage         `json:"disable-cooling"`
}

// codearts-api-key: []CodeArtsKey
func (h *Handler) GetCodeArtsKeys(c *gin.Context) {
	c.JSON(200, gin.H{"codearts-api-key": h.codeArtsKeysWithAuthIndex()})
}

func (h *Handler) PutCodeArtsKeys(c *gin.Context) {
	data, errRead := c.GetRawData()
	if errRead != nil {
		c.JSON(400, gin.H{"error": "failed to read body"})
		return
	}
	var arr []config.CodeArtsKey
	if errUnmarshal := json.Unmarshal(data, &arr); errUnmarshal != nil {
		var obj struct {
			Items []config.CodeArtsKey `json:"items"`
		}
		if errObject := json.Unmarshal(data, &obj); errObject != nil || len(obj.Items) == 0 {
			c.JSON(400, gin.H{"error": "invalid body"})
			return
		}
		arr = obj.Items
	}
	filtered := make([]config.CodeArtsKey, 0, len(arr))
	for i := range arr {
		entry := arr[i]
		normalizeCodeArtsKey(&entry)
		if entry.APIKey == "" || entry.SecretKey == "" {
			continue
		}
		if rejectInvalidCredentialWeight(c, fmt.Sprintf("codearts-api-key[%d].weight", i), entry.Weight) {
			return
		}
		filtered = append(filtered, entry)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cfg.CodeArtsKey = filtered
	h.cfg.SanitizeCodeArtsKeys()
	h.persistLocked(c)
}

func (h *Handler) PatchCodeArtsKey(c *gin.Context) {
	var body struct {
		Index *int              `json:"index"`
		Match *string           `json:"match"`
		Value *codeArtsKeyPatch `json:"value"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil || body.Value == nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	entries := h.cfg.CodeArtsKey
	targetIndex := -1
	if body.Index != nil && *body.Index >= 0 && *body.Index < len(entries) {
		targetIndex = *body.Index
	}
	if targetIndex == -1 && body.Match != nil {
		match := strings.TrimSpace(*body.Match)
		for i := range entries {
			if entries[i].APIKey == match {
				targetIndex = i
				break
			}
		}
	}
	if targetIndex == -1 {
		c.JSON(404, gin.H{"error": "item not found"})
		return
	}

	entry := entries[targetIndex]
	if body.Value.APIKey != nil {
		entry.APIKey = strings.TrimSpace(*body.Value.APIKey)
	}
	if body.Value.SecretKey != nil {
		entry.SecretKey = strings.TrimSpace(*body.Value.SecretKey)
	}
	if body.Value.SecurityToken != nil {
		entry.SecurityToken = strings.TrimSpace(*body.Value.SecurityToken)
	}
	if body.Value.RefreshToken != nil {
		entry.RefreshToken = strings.TrimSpace(*body.Value.RefreshToken)
	}
	if body.Value.Priority != nil {
		entry.Priority = *body.Value.Priority
	}
	if body.Value.Prefix != nil {
		entry.Prefix = strings.TrimSpace(*body.Value.Prefix)
	}
	if body.Value.BaseURL != nil {
		entry.BaseURL = strings.TrimSpace(*body.Value.BaseURL)
	}
	if body.Value.ProxyURL != nil {
		entry.ProxyURL = strings.TrimSpace(*body.Value.ProxyURL)
	}
	if body.Value.Models != nil {
		entry.Models = *body.Value.Models
	}
	if body.Value.Headers != nil {
		entry.Headers = *body.Value.Headers
	}
	if body.Value.ExcludedModels != nil {
		entry.ExcludedModels = *body.Value.ExcludedModels
	}
	if len(body.Value.Weight) > 0 {
		weight, errWeight := parseCredentialWeightPatch(body.Value.Weight)
		if errWeight != nil {
			c.JSON(400, gin.H{"error": errWeight.Error()})
			return
		}
		entry.Weight = weight
	}
	if len(body.Value.DisableCooling) > 0 {
		if string(body.Value.DisableCooling) == "null" {
			entry.DisableCooling = nil
		} else {
			var disableCooling bool
			if errUnmarshal := json.Unmarshal(body.Value.DisableCooling, &disableCooling); errUnmarshal != nil {
				c.JSON(400, gin.H{"error": "invalid disable-cooling"})
				return
			}
			entry.DisableCooling = &disableCooling
		}
	}
	normalizeCodeArtsKey(&entry)
	if rejectInvalidCredentialWeight(c, "codearts-api-key.weight", entry.Weight) {
		return
	}
	entries[targetIndex] = entry
	h.cfg.CodeArtsKey = entries
	h.cfg.SanitizeCodeArtsKeys()
	h.persistLocked(c)
}

func (h *Handler) DeleteCodeArtsKey(c *gin.Context) {
	apiKey := strings.TrimSpace(c.Query("api-key"))
	baseURL := strings.TrimSpace(c.Query("base-url"))
	if apiKey == "" && baseURL == "" {
		c.JSON(400, gin.H{"error": "api-key or base-url is required"})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	kept := make([]config.CodeArtsKey, 0, len(h.cfg.CodeArtsKey))
	removed := 0
	for _, entry := range h.cfg.CodeArtsKey {
		matches := false
		if apiKey != "" {
			matches = strings.EqualFold(strings.TrimSpace(entry.APIKey), apiKey)
			if matches && baseURL != "" {
				matches = strings.EqualFold(strings.TrimSpace(entry.BaseURL), baseURL)
			}
		} else {
			matches = strings.EqualFold(strings.TrimSpace(entry.BaseURL), baseURL)
		}
		if matches {
			removed++
			continue
		}
		kept = append(kept, entry)
	}
	if removed == 0 {
		c.JSON(404, gin.H{"error": "item not found"})
		return
	}
	h.cfg.CodeArtsKey = kept
	h.persistLocked(c)
}
