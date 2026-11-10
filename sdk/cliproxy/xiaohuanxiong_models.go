package cliproxy

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	xiaohuanxiongauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// xiaohuanxiongModelRefreshScanInterval is how often accounts are rescanned
	// for an expired or unfetched catalogue.
	xiaohuanxiongModelRefreshScanInterval = time.Minute
	// xiaohuanxiongModelRefreshInterval is how long a fetched catalogue is trusted.
	xiaohuanxiongModelRefreshInterval = registry.ModelsRefreshInterval
	// xiaohuanxiongModelRetryInterval bounds retries after a failed fetch.
	xiaohuanxiongModelRetryInterval = 10 * time.Minute
	// xiaohuanxiongModelCreatedAt is the registry entry creation timestamp.
	xiaohuanxiongModelCreatedAt int64 = 1759276800
)

// xiaohuanxiongCatalogueEntry records the last successful fetch for one account.
type xiaohuanxiongCatalogueEntry struct {
	expiresAt time.Time
	nextRetry time.Time
}

var (
	xiaohuanxiongNowFunc     = time.Now
	xiaohuanxiongProbeWg     sync.WaitGroup
	xiaohuanxiongCatalogueMu sync.RWMutex
	xiaohuanxiongCatalogue   = make(map[string]xiaohuanxiongCatalogueEntry)
	// xiaohuanxiongProbeSlots bounds concurrent catalogue fetches.
	xiaohuanxiongProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// xiaohuanxiongCatalogueKey isolates catalogues by account and credential. The
// access token is part of the key because a new login can belong to a different
// account while keeping the same auth ID.
func xiaohuanxiongCatalogueKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	baseURL := ""
	session := ""
	if auth.Attributes != nil {
		baseURL = strings.TrimRight(strings.TrimSpace(auth.Attributes["base_url"]), "/")
		session = strings.TrimSpace(auth.Attributes["access_token"])
	}
	if session == "" {
		session = xiaohuanxiongCatalogueAttribute(auth, "access_token")
	}
	return fmt.Sprintf("%s|%s|%x", auth.ID, baseURL, sha256.Sum256([]byte(session)))
}

// asyncProbeXiaohuanxiongCatalogue launches a bounded detached fetch for one
// account. The caller's context only gates admission: a request-scoped context
// would cancel the probe as soon as the registration call returns.
func (s *Service) asyncProbeXiaohuanxiongCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
	if s == nil || auth == nil || auth.ID == "" || auth.Disabled {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil || s.antigravityHomeEnabled() {
		return
	}
	if pluginHostHasAuthModelProvider(s.pluginHost, auth.Provider) {
		return
	}
	snapshot := auth.Clone()
	snapshotKey := xiaohuanxiongCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	xiaohuanxiongProbeWg.Add(1)
	go func() {
		defer xiaohuanxiongProbeWg.Done()
		xiaohuanxiongProbeSlots <- struct{}{}
		defer func() { <-xiaohuanxiongProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeXiaohuanxiongCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeXiaohuanxiongCatalogueProbe fetches the account's model catalogue and
// republishes the account's registered models when the fetch succeeds.
func (s *Service) executeXiaohuanxiongCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	catalogue, errFetch := s.fetchXiaohuanxiongCatalogue(ctx, auth)
	if errFetch != nil {
		if ctx.Err() == nil {
			log.Debugf("xiaohuanxiong: catalogue probe for %s failed: %v", auth.ID, errFetch)
		}
		s.deferXiaohuanxiongCatalogueRetry(snapshotKey)
		return
	}
	if len(catalogue) == 0 {
		// An empty catalogue leaves the credential unusable if every fallback row
		// were revoked, so the fallback list stays published and a retry is
		// scheduled instead.
		log.Debugf("xiaohuanxiong: catalogue for %s is empty; keeping fallback models", auth.ID)
		s.deferXiaohuanxiongCatalogueRetry(snapshotKey)
		return
	}
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.xiaohuanxiongModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	xiaohuanxiongCatalogueMu.Lock()
	xiaohuanxiongCatalogue[snapshotKey] = xiaohuanxiongCatalogueEntry{
		expiresAt: xiaohuanxiongNowFunc().Add(xiaohuanxiongModelRefreshInterval),
	}
	xiaohuanxiongCatalogueMu.Unlock()
}

// deferXiaohuanxiongCatalogueRetry schedules a bounded retry after a failed or
// empty fetch.
func (s *Service) deferXiaohuanxiongCatalogueRetry(snapshotKey string) {
	xiaohuanxiongCatalogueMu.Lock()
	entry := xiaohuanxiongCatalogue[snapshotKey]
	entry.nextRetry = xiaohuanxiongNowFunc().Add(xiaohuanxiongModelRetryInterval)
	xiaohuanxiongCatalogue[snapshotKey] = entry
	xiaohuanxiongCatalogueMu.Unlock()
}

// xiaohuanxiongCatalogModel is one chat entry of the /model_catalog response.
type xiaohuanxiongCatalogModel struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	DisplayDescription string   `json:"display_description"`
	Visible            bool     `json:"visible"`
	Tags               []string `json:"tags"`
	AbilityLevel       int      `json:"ability_level"`
	BillingCategory    string   `json:"billing_category"`
	BillingMultiplier  float64  `json:"billing_multiplier"`
	BillingStatus      string   `json:"billing_status"`
	Params             struct {
		ContextWindow int `json:"context_window"`
		MaxTokens     int `json:"max_tokens"`
	} `json:"params"`
}

// xiaohuanxiongCatalogPayload is the /model_catalog envelope. Only the chat
// category is consumed: the gateway may publish other categories later.
type xiaohuanxiongCatalogPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Categories []struct {
			Type   string                      `json:"type"`
			Models []xiaohuanxiongCatalogModel `json:"models"`
		} `json:"categories"`
	} `json:"data"`
}

// fetchXiaohuanxiongCatalogue reads the account's chat model catalogue with its
// access token.
func (s *Service) fetchXiaohuanxiongCatalogue(ctx context.Context, auth *coreauth.Auth) ([]xiaohuanxiongCatalogModel, error) {
	if auth == nil {
		return nil, fmt.Errorf("xiaohuanxiong: auth is required")
	}
	session := xiaohuanxiongCatalogueAttribute(auth, "access_token")
	if session == "" {
		session = xiaohuanxiongCatalogueAttribute(auth, "api_key")
	}
	if session == "" {
		return nil, fmt.Errorf("xiaohuanxiong: no access token on auth %s", auth.ID)
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	client := xiaohuanxiongauth.NewClientWithBaseURL(cfg, xiaohuanxiongModelFetchProxyURL(s, auth), xiaohuanxiongCatalogueOrigin(auth))
	raw, errFetch := client.FetchModelCatalog(ctx, session)
	if errFetch != nil {
		return nil, errFetch
	}
	return parseXiaohuanxiongCatalogue(raw)
}

// parseXiaohuanxiongCatalogue returns the models of the chat category.
//
// The desktop client resolves its model list the same way: it requires
// code == 0, then takes the category whose type is "chat".
func parseXiaohuanxiongCatalogue(raw []byte) ([]xiaohuanxiongCatalogModel, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil, fmt.Errorf("xiaohuanxiong: model catalog is empty")
	}
	var payload xiaohuanxiongCatalogPayload
	if errUnmarshal := json.Unmarshal(raw, &payload); errUnmarshal != nil {
		return nil, fmt.Errorf("xiaohuanxiong: decode model catalog: %w", errUnmarshal)
	}
	if payload.Code != 0 {
		return nil, fmt.Errorf("xiaohuanxiong: model catalog code %d: %s", payload.Code, payload.Message)
	}
	for _, category := range payload.Data.Categories {
		if !strings.EqualFold(strings.TrimSpace(category.Type), "chat") {
			continue
		}
		models := make([]xiaohuanxiongCatalogModel, 0, len(category.Models))
		for _, model := range category.Models {
			if strings.TrimSpace(model.Name) == "" {
				continue
			}
			models = append(models, model)
		}
		return models, nil
	}
	return nil, nil
}

// xiaohuanxiongModelsForCatalogue maps a fetched catalogue through the same
// exclusion, alias, plugin and prefix pipeline the static registration uses.
func (s *Service) xiaohuanxiongModelsForCatalogue(auth *coreauth.Auth, catalogue []xiaohuanxiongCatalogModel) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue))
	for _, entry := range catalogue {
		info := xiaohuanxiongCatalogueModelInfo(entry)
		if info == nil {
			continue
		}
		models = append(models, info)
	}
	excluded := cfg.OAuthExcludedModels[coreauth.OAuthModelAliasChannel(auth.Provider, auth.AuthKind())]
	if value := strings.TrimSpace(auth.Attributes["excluded_models"]); value != "" {
		excluded = strings.Split(value, ",")
	}
	models = applyExcludedModels(models, excluded)
	models = applyOAuthModelAliasForAuth(cfg, constant.Xiaohuanxiong, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(constant.Xiaohuanxiong, models, excluded)
	models = applyOAuthSettingsForAuth(cfg, constant.Xiaohuanxiong, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// xiaohuanxiongCatalogueModelInfo builds the registry entry for one catalogue
// row, refining per-model metadata from the curated fallback catalogue.
func xiaohuanxiongCatalogueModelInfo(entry xiaohuanxiongCatalogModel) *ModelInfo {
	id := strings.TrimSpace(entry.Name)
	if id == "" {
		return nil
	}
	info := registry.LookupStaticModelInfoByChannel(id, constant.Xiaohuanxiong)
	if info == nil {
		info = registry.LookupStaticModelInfo(id)
	}
	if info == nil {
		info = &ModelInfo{Object: "model", Created: xiaohuanxiongModelCreatedAt, OwnedBy: "sensetime"}
	}
	clone := *info
	clone.ID = id
	clone.Type = constant.Xiaohuanxiong
	if clone.Object == "" {
		clone.Object = "model"
	}
	if clone.Created == 0 {
		clone.Created = xiaohuanxiongModelCreatedAt
	}
	// The gateway's display_description is the operator-facing sentence; the
	// fallback catalog stores the same normalized name, so it wins when present.
	if strings.TrimSpace(clone.DisplayName) == "" {
		clone.DisplayName = firstNonEmptyString(entry.Description, entry.DisplayDescription)
	}
	if clone.DisplayName == "" {
		clone.DisplayName = id
	}
	if entry.Params.ContextWindow > 0 {
		clone.ContextLength = entry.Params.ContextWindow
	}
	// The tag list is the gateway's own capability statement; "vision" is the
	// only tag that changes the wire contract.
	if xiaohuanxiongTagPresent(entry.Tags, "vision") {
		clone.SupportedInputModalities = []string{"text", "image"}
	}
	return &clone
}

// xiaohuanxiongTagPresent reports whether a catalogue tag is present.
func xiaohuanxiongTagPresent(tags []string, want string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), want) {
			return true
		}
	}
	return false
}

// firstNonEmptyString returns the first trimmed non-empty value.
func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// xiaohuanxiongCatalogueOrigin derives the web origin the catalogue is read
// from. A credential stores the full LLM gateway base URL, while the client
// appends that path itself, so the LLMPath suffix is stripped back off.
func xiaohuanxiongCatalogueOrigin(auth *coreauth.Auth) string {
	base := strings.TrimRight(xiaohuanxiongCatalogueAttribute(auth, "base_url"), "/")
	if base == "" {
		return xiaohuanxiongauth.BaseURL
	}
	if origin := strings.TrimSuffix(base, xiaohuanxiongauth.LLMPath); origin != base && origin != "" {
		return origin
	}
	return xiaohuanxiongauth.BaseURL
}

// xiaohuanxiongCatalogueAttribute reads a credential attribute, falling back to
// metadata.
func xiaohuanxiongCatalogueAttribute(auth *coreauth.Auth, key string) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if value := strings.TrimSpace(auth.Attributes[key]); value != "" {
			return value
		}
	}
	if auth.Metadata != nil {
		if value, ok := auth.Metadata[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// xiaohuanxiongModelFetchProxyURL resolves the proxy used for catalogue fetches.
func xiaohuanxiongModelFetchProxyURL(s *Service, auth *coreauth.Auth) string {
	if auth != nil {
		if proxyURL := strings.TrimSpace(auth.ProxyURL); proxyURL != "" {
			return proxyURL
		}
	}
	if s != nil {
		s.cfgMu.RLock()
		defer s.cfgMu.RUnlock()
		if s.cfg != nil {
			return strings.TrimSpace(s.cfg.ProxyURL)
		}
	}
	return ""
}

// refreshXiaohuanxiongCatalogues reschedules accounts whose catalogue expired
// or whose last fetch failed. It never blocks on the fetch itself.
func (s *Service) refreshXiaohuanxiongCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := xiaohuanxiongNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled || !strings.EqualFold(auth.Provider, constant.Xiaohuanxiong) {
			continue
		}
		snapshot := auth.Clone()
		key := xiaohuanxiongCatalogueKey(snapshot)
		xiaohuanxiongCatalogueMu.RLock()
		entry := xiaohuanxiongCatalogue[key]
		xiaohuanxiongCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeXiaohuanxiongCatalogue(ctx, snapshot, constant.Xiaohuanxiong)
	}
}

// runXiaohuanxiongModelRefresh periodically reschedules stale catalogues.
func (s *Service) runXiaohuanxiongModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(xiaohuanxiongModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshXiaohuanxiongCatalogues(ctx)
			timer.Reset(xiaohuanxiongModelRefreshScanInterval)
		}
	}
}
