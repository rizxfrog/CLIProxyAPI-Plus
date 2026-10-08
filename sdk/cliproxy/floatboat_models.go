package cliproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	floatboatauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/floatboat"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// floatboatModelRefreshScanInterval is how often accounts are rescanned for
	// an expired or unfetched catalogue. It bounds discovery latency; the fetch
	// itself only happens once floatboatModelRefreshInterval has elapsed.
	floatboatModelRefreshScanInterval = time.Minute
	// floatboatModelRefreshInterval is how long a fetched catalogue is trusted.
	floatboatModelRefreshInterval = registry.ModelsRefreshInterval
	// floatboatModelRetryInterval bounds retries after a failed fetch. Gateway
	// catalogues change rarely, so a short fixed delay is sufficient.
	floatboatModelRetryInterval = 10 * time.Minute
)

// floatboatCatalogueEntry records the last successful fetch for one account.
type floatboatCatalogueEntry struct {
	// expiresAt is when the fetched catalogue should be refreshed again.
	expiresAt time.Time
	// nextRetry gates retries after a failed fetch.
	nextRetry time.Time
}

var (
	floatboatNowFunc     = time.Now
	floatboatProbeWg     sync.WaitGroup
	floatboatCatalogueMu sync.RWMutex
	floatboatCatalogue   = make(map[string]floatboatCatalogueEntry)
	// floatboatProbeSlots bounds concurrent catalogue fetches.
	floatboatProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// floatboatCatalogueKey isolates catalogues by account and route. The credential
// is part of the key because a rotated key can belong to a different plan group.
func floatboatCatalogueKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	baseURL := ""
	credential := ""
	if auth.Attributes != nil {
		baseURL = strings.TrimRight(strings.TrimSpace(auth.Attributes["base_url"]), "/")
		credential = strings.TrimSpace(auth.Attributes["api_key"])
		if credential == "" {
			credential = strings.TrimSpace(auth.Attributes["access_token"])
		}
	}
	if credential == "" && auth.Metadata != nil {
		if key, ok := auth.Metadata["api_key"].(string); ok {
			credential = strings.TrimSpace(key)
		}
	}
	return fmt.Sprintf("%s|%s|%x", auth.ID, baseURL, sha256.Sum256([]byte(credential)))
}

// asyncProbeFloatboatCatalogues launches a bounded, detached fetch for one
// account. The caller's context only gates admission: a request-scoped context
// would otherwise cancel the probe as soon as the registration call returns.
func (s *Service) asyncProbeFloatboatCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
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
	snapshotKey := floatboatCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	floatboatProbeWg.Add(1)
	go func() {
		defer floatboatProbeWg.Done()
		floatboatProbeSlots <- struct{}{}
		defer func() { <-floatboatProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeFloatboatCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeFloatboatCatalogueProbe fetches the account's gateway catalogue and
// republishes the account's registered models when the fetch succeeds.
func (s *Service) executeFloatboatCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	catalogue, err := s.fetchFloatboatCatalogue(ctx, auth)
	if err != nil {
		if ctx.Err() == nil {
			log.Debugf("floatboat: catalogue probe for %s failed: %v", auth.ID, err)
		}
		floatboatCatalogueMu.Lock()
		entry := floatboatCatalogue[snapshotKey]
		entry.nextRetry = floatboatNowFunc().Add(floatboatModelRetryInterval)
		floatboatCatalogue[snapshotKey] = entry
		floatboatCatalogueMu.Unlock()
		return
	}
	if len(catalogue.Entries) == 0 {
		// An empty catalogue is authoritative for the account, but revoking every
		// model would leave the credential unusable. Keep the fallback catalogue
		// and retry later instead.
		log.Debugf("floatboat: catalogue for %s is empty; keeping fallback models", auth.ID)
		floatboatCatalogueMu.Lock()
		entry := floatboatCatalogue[snapshotKey]
		entry.nextRetry = floatboatNowFunc().Add(floatboatModelRetryInterval)
		floatboatCatalogue[snapshotKey] = entry
		floatboatCatalogueMu.Unlock()
		return
	}
	// The snapshot must still describe the live credential before publication.
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.floatboatModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	floatboatCatalogueMu.Lock()
	floatboatCatalogue[snapshotKey] = floatboatCatalogueEntry{
		expiresAt: floatboatNowFunc().Add(floatboatModelRefreshInterval),
	}
	floatboatCatalogueMu.Unlock()
}

// fetchFloatboatCatalogue reads the gateway catalogue with the credential's
// inference key. base_url is the gateway origin; on the real product it is
// newapi.aoe.chat rather than the account backend.
func (s *Service) fetchFloatboatCatalogue(ctx context.Context, auth *coreauth.Auth) (*floatboatauth.Catalogue, error) {
	if auth == nil {
		return nil, fmt.Errorf("floatboat: auth is required")
	}
	baseURL := floatboatAttribute(auth, "base_url", "inference_base_url")
	if baseURL == "" {
		baseURL = floatboatauth.DefaultInferenceBaseURL
	}
	apiKey := floatboatAttribute(auth, "api_key")
	if apiKey == "" {
		apiKey = floatboatAttribute(auth, "access_token")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("floatboat: no inference credential on auth %s", auth.ID)
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	client := floatboatauth.NewClientWithProxyURL(cfg, floatboatModelFetchProxyURL(s, auth), baseURL)
	return client.FetchCatalogue(ctx, baseURL, apiKey)
}

// floatboatModelsForCatalogue maps a fetched catalogue through the same
// exclusion, alias, plugin and prefix pipeline static registrations use.
func (s *Service) floatboatModelsForCatalogue(auth *coreauth.Auth, catalogue *floatboatauth.Catalogue) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue.Entries))
	for _, entry := range catalogue.Entries {
		info := floatboatCatalogueModelInfo(auth, entry)
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
	models = applyOAuthModelAliasForAuth(cfg, constant.Floatboat, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(constant.Floatboat, models)
	models = applyOAuthSettingsForAuth(cfg, constant.Floatboat, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// floatboatCatalogueModelInfo builds the registry entry for one gateway model.
// Per-model metadata (context window, thinking support) is not published by the
// gateway, so the curated family defaults are reused as a refinement layer.
func floatboatCatalogueModelInfo(auth *coreauth.Auth, entry floatboatauth.CatalogueEntry) *ModelInfo {
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return nil
	}
	var info *ModelInfo
	if auth != nil {
		if info = registry.LookupStaticModelInfoByChannel(id, constant.Floatboat); info == nil {
			info = registry.LookupStaticModelInfo(id)
		}
	}
	if info == nil {
		info = &ModelInfo{Object: "model", Created: floatboatModelCreatedAt, OwnedBy: "floatboat"}
	}
	clone := *info
	clone.ID = id
	clone.Type = constant.Floatboat
	if clone.Object == "" {
		clone.Object = "model"
	}
	if clone.Created == 0 {
		clone.Created = floatboatModelCreatedAt
	}
	if display := strings.TrimSpace(entry.DisplayName); display != "" {
		clone.DisplayName = display
	}
	if clone.DisplayName == "" {
		clone.DisplayName = id
	}
	return &clone
}

// floatboatModelCreatedAt is the gateway catalogue's own creation timestamp.
const floatboatModelCreatedAt int64 = 1759276800

// floatboatAttribute reads a credential attribute, falling back to metadata.
func floatboatAttribute(auth *coreauth.Auth, keys ...string) string {
	if auth == nil {
		return ""
	}
	for _, key := range keys {
		if auth.Attributes != nil {
			if value := strings.TrimSpace(auth.Attributes[key]); value != "" {
				return value
			}
		}
		if auth.Metadata != nil {
			if value, ok := auth.Metadata[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

// floatboatModelFetchProxyURL resolves the proxy used for catalogue fetches.
func floatboatModelFetchProxyURL(s *Service, auth *coreauth.Auth) string {
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

// refreshFloatboatCatalogues reschedules accounts whose catalogue expired or
// whose last fetch failed. It never blocks on the fetch itself.
func (s *Service) refreshFloatboatCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := floatboatNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled || !strings.EqualFold(auth.Provider, constant.Floatboat) {
			continue
		}
		snapshot := auth.Clone()
		key := floatboatCatalogueKey(snapshot)
		floatboatCatalogueMu.RLock()
		entry := floatboatCatalogue[key]
		floatboatCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeFloatboatCatalogue(ctx, snapshot, constant.Floatboat)
	}
}

// runFloatboatModelRefresh periodically reschedules stale floatboat catalogues.
func (s *Service) runFloatboatModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(floatboatModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshFloatboatCatalogues(ctx)
			timer.Reset(floatboatModelRefreshScanInterval)
		}
	}
}
