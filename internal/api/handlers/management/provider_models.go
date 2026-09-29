package management

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/util"
)

var nativeModelProviders = []string{"aistudio", "antigravity", "claude", "cline", "codearts", "codebuddy-ai", "codebuddy-cn", "codex", "deepseek-web", "devin", "gemini", "gemini-interactions", "kimi", "meta", "minimax", "minimax-cn", "qoder-ai", "qoder-cn", "qwen-web", "trae", "vertex", "xai", "xiaohuanxiong"}

type managedProviderModel struct {
	ID                     string              `json:"id"`
	DisplayName            string              `json:"display_name"`
	Source                 string              `json:"source"`
	Enabled                bool                `json:"enabled"`
	RegisteredAccountCount *int                `json:"registered_account_count"`
	InactiveReasons        []string            `json:"inactive_reasons"`
	Definition             *registry.ModelInfo `json:"definition"`
}

type managedProviderCatalog struct {
	Provider      string                 `json:"provider"`
	Override      config.ProviderModels  `json:"override"`
	Models        []managedProviderModel `json:"models"`
	ReadOnly      bool                   `json:"read_only"`
	RuntimeStatus string                 `json:"runtime_status"`
	CatalogScope  string                 `json:"catalog_scope"`
}

// providerCatalogsLocked reads saved intent and observed registry state separately.
// A void asynchronous reload hook cannot prove application; never label it applied.
func (h *Handler) providerCatalogsLocked() []managedProviderCatalog {
	cfg := h.cfg
	if cfg == nil {
		cfg = &config.Config{}
	}
	catalogs := map[string]map[string]*registry.ModelInfo{}
	sources := map[string]map[string]string{}
	counts := map[string]map[string]int{}
	add := func(provider string, models []*registry.ModelInfo, source string) {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" {
			return
		}
		if catalogs[provider] == nil {
			catalogs[provider] = map[string]*registry.ModelInfo{}
			sources[provider] = map[string]string{}
			counts[provider] = map[string]int{}
		}
		for _, model := range models {
			if model == nil {
				continue
			}
			id := model.MetadataModelID
			if id == "" {
				id = model.ID
			}
			if id == "" {
				continue
			}
			if _, ok := catalogs[provider][id]; !ok {
				clone := *model
				clone.ID = id
				catalogs[provider][id] = &clone
				sources[provider][id] = source
			}
		}
	}
	for _, provider := range nativeModelProviders {
		add(provider, registry.GetStaticModelDefinitionsByChannel(provider), "default")
	}
	for provider, models := range configuredProviderCatalog(cfg) {
		add(provider, models, "configured")
	}
	if h.pluginHost != nil {
		for provider, models := range h.pluginHost.ModelCatalog() {
			add(provider, models, "default")
		}
	}
	for provider := range cfg.ProviderModels {
		add(provider, nil, "")
	}
	if h.authManager != nil && !cfg.Home.Enabled {
		for _, auth := range h.authManager.List() {
			provider := strings.ToLower(strings.TrimSpace(auth.Provider))
			if v := strings.TrimSpace(auth.Attributes["provider_key"]); v != "" {
				provider = strings.ToLower(v)
			}
			if strings.TrimSpace(auth.Attributes["compat_name"]) != "" {
				provider = util.OpenAICompatibleProviderKey(provider)
			}
			// An auth whose provider resolves to empty cannot be attributed to any
			// catalog entry; skip it instead of writing to counts[""], which would
			// panic on a nil map because add() early-returns for an empty provider.
			if provider == "" {
				continue
			}
			add(provider, nil, "")
			if auth.Disabled {
				continue
			}
			seen := map[string]bool{}
			for _, model := range registry.GetGlobalRegistry().GetModelsForClient(auth.ID) {
				id := model.MetadataModelID
				if id == "" {
					id = model.ID
				}
				isCustom := false
				for _, custom := range cfg.ProviderModels[provider].Custom {
					if custom.ID == id {
						isCustom = true
						break
					}
				}
				if !isCustom {
					add(provider, []*registry.ModelInfo{model}, "runtime")
				}
				if !seen[id] {
					counts[provider][id]++
					seen[id] = true
				}
			}
		}
	}
	providers := make([]string, 0, len(catalogs))
	for provider := range catalogs {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	out := make([]managedProviderCatalog, 0, len(providers))
	for _, provider := range providers {
		overlay := cfg.ProviderModels[provider]
		if overlay.Disabled == nil {
			overlay.Disabled = []string{}
		}
		if overlay.Custom == nil {
			overlay.Custom = []config.ProviderModel{}
		}
		base := make([]*registry.ModelInfo, 0, len(catalogs[provider]))
		for _, model := range catalogs[provider] {
			base = append(base, model)
		}
		customIDs := map[string]bool{}
		for _, model := range overlay.Custom {
			customIDs[model.ID] = true
		}
		models := config.MergeProviderModels(base, provider, overlay, true)
		seen := map[string]bool{}
		for _, model := range models {
			seen[model.ID] = true
		}
		disabled := map[string]bool{}
		for _, id := range overlay.Disabled {
			disabled[id] = true
			if !seen[id] {
				models = append(models, &registry.ModelInfo{ID: id, DisplayName: id})
				sources[provider][id] = "unavailable"
			}
		}
		sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
		catalog := managedProviderCatalog{Provider: provider, Override: overlay, Models: []managedProviderModel{}, RuntimeStatus: "unverified", CatalogScope: "local"}
		if cfg.Home.Enabled {
			catalog.ReadOnly = true
			catalog.RuntimeStatus = "home-managed"
			catalog.CatalogScope = "local-preview"
		}
		for _, model := range models {
			source := sources[provider][model.ID]
			if customIDs[model.ID] {
				if source == "" {
					source = "custom"
				} else {
					source = "override"
				}
			}
			count := counts[provider][model.ID]
			row := managedProviderModel{ID: model.ID, DisplayName: model.DisplayName, Source: source, Enabled: !disabled[model.ID], Definition: model, RegisteredAccountCount: &count, InactiveReasons: []string{}}
			if cfg.Home.Enabled {
				row.RegisteredAccountCount = nil
				row.InactiveReasons = append(row.InactiveReasons, "home-managed")
			} else if disabled[model.ID] {
				row.InactiveReasons = append(row.InactiveReasons, "provider-disabled")
			} else if count == 0 {
				row.InactiveReasons = append(row.InactiveReasons, "not-registered")
			}
			catalog.Models = append(catalog.Models, row)
		}
		out = append(out, catalog)
	}
	return out
}

func (h *Handler) GetProviderModels(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"providers": h.providerCatalogsLocked()})
}

func (h *Handler) GetProviderModel(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, catalog := range h.providerCatalogsLocked() {
		if catalog.Provider == provider {
			c.JSON(http.StatusOK, catalog)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
}

func (h *Handler) PutProviderModel(c *gin.Context)    { h.saveProviderModel(c, false) }
func (h *Handler) DeleteProviderModel(c *gin.Context) { h.saveProviderModel(c, true) }

func (h *Handler) saveProviderModel(c *gin.Context, reset bool) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(503, gin.H{"error": "config unavailable"})
		return
	}
	if h.cfg.Home.Enabled {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"code": "provider_models_home_managed", "error": "Provider models are managed by Home; local overrides are read-only in Home mode."})
		return
	}
	var entry config.ProviderModels
	if !reset {
		decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
		decoder.DisallowUnknownFields()
		var body *config.ProviderModels
		if errDecode := decoder.Decode(&body); errDecode != nil || body == nil {
			h.mu.Unlock()
			c.JSON(400, gin.H{"error": "invalid provider model body"})
			return
		}
		if errEnd := decoder.Decode(new(any)); errEnd != io.EOF {
			h.mu.Unlock()
			c.JSON(400, gin.H{"error": "invalid trailing JSON"})
			return
		}
		entry = *body
	}
	normalized, errNormalize := config.NormalizeProviderModels(map[string]config.ProviderModels{provider: entry})
	if errNormalize != nil {
		h.mu.Unlock()
		c.JSON(400, gin.H{"error": errNormalize.Error()})
		return
	}
	previousConfig := h.cfg
	previous := previousConfig.ProviderModels
	next := make(map[string]config.ProviderModels, len(previous)+1)
	for key, value := range previous {
		next[key] = value
	}
	if reset {
		delete(next, provider)
	} else {
		next[provider] = normalized[provider]
	}
	h.cfg = previousConfig.CloneForRuntime()
	h.cfg.ProviderModels = next
	snapshot, saved := h.saveConfigAndSnapshotLocked(c)
	if !saved {
		h.cfg = previousConfig
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"status": "ok", "runtime_status": "pending", "provider": provider})
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
}
