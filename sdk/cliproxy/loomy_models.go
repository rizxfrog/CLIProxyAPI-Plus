package cliproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	loomyauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/loomy"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// loomyModelRefreshScanInterval is how often accounts are rescanned for an
	// expired or unfetched catalogue.
	loomyModelRefreshScanInterval = time.Minute
	// loomyModelRefreshInterval is how long a fetched catalogue is trusted.
	loomyModelRefreshInterval = registry.ModelsRefreshInterval
	// loomyModelRetryInterval bounds retries after a failed fetch.
	loomyModelRetryInterval = 10 * time.Minute
	// loomyModelCreatedAt is the registry entry creation timestamp.
	loomyModelCreatedAt int64 = 1759276800
)

// loomyCatalogueEntry records the last successful fetch for one account.
type loomyCatalogueEntry struct {
	expiresAt time.Time
	nextRetry time.Time
}

var (
	loomyNowFunc     = time.Now
	loomyProbeWg     sync.WaitGroup
	loomyCatalogueMu sync.RWMutex
	loomyCatalogue   = make(map[string]loomyCatalogueEntry)
	// loomyProbeSlots bounds concurrent catalogue fetches.
	loomyProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// loomyCatalogueKey isolates catalogues by account and credential. The session
// is part of the key because a new login can belong to a different account.
func loomyCatalogueKey(auth *coreauth.Auth) string {
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
		session = loomyAttribute(auth, "access_token")
	}
	return fmt.Sprintf("%s|%s|%x", auth.ID, baseURL, sha256.Sum256([]byte(session)))
}

// asyncProbeLoomyCatalogue launches a bounded detached fetch for one account.
// The caller's context only gates admission: a request-scoped context would
// cancel the probe as soon as the registration call returns.
func (s *Service) asyncProbeLoomyCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
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
	snapshotKey := loomyCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	loomyProbeWg.Add(1)
	go func() {
		defer loomyProbeWg.Done()
		loomyProbeSlots <- struct{}{}
		defer func() { <-loomyProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeLoomyCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeLoomyCatalogueProbe fetches the account's model catalogue and
// republishes the account's registered models when the fetch succeeds.
func (s *Service) executeLoomyCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	catalogue, errFetch := s.fetchLoomyCatalogue(ctx, auth)
	if errFetch != nil {
		if ctx.Err() == nil {
			log.Debugf("loomy: catalogue probe for %s failed: %v", auth.ID, errFetch)
		}
		s.deferLoomyCatalogueRetry(snapshotKey)
		return
	}
	if len(catalogue) == 0 {
		// An empty catalogue leaves the credential unusable if every fallback row
		// were revoked, so the fallback list stays published and a retry is
		// scheduled instead.
		log.Debugf("loomy: catalogue for %s is empty; keeping fallback models", auth.ID)
		s.deferLoomyCatalogueRetry(snapshotKey)
		return
	}
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.loomyModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	loomyCatalogueMu.Lock()
	loomyCatalogue[snapshotKey] = loomyCatalogueEntry{
		expiresAt: loomyNowFunc().Add(loomyModelRefreshInterval),
	}
	loomyCatalogueMu.Unlock()
}

// deferLoomyCatalogueRetry schedules a bounded retry after a failed or empty
// fetch.
func (s *Service) deferLoomyCatalogueRetry(snapshotKey string) {
	loomyCatalogueMu.Lock()
	entry := loomyCatalogue[snapshotKey]
	entry.nextRetry = loomyNowFunc().Add(loomyModelRetryInterval)
	loomyCatalogue[snapshotKey] = entry
	loomyCatalogueMu.Unlock()
}

// fetchLoomyCatalogue reads the account's model catalogue with its session.
func (s *Service) fetchLoomyCatalogue(ctx context.Context, auth *coreauth.Auth) ([]loomyauth.Model, error) {
	if auth == nil {
		return nil, fmt.Errorf("loomy: auth is required")
	}
	session := loomyAttribute(auth, "access_token")
	if session == "" {
		return nil, fmt.Errorf("loomy: no session on auth %s", auth.ID)
	}
	baseURL := loomyAttribute(auth, "base_url")
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	client := loomyauth.NewClientWithBaseURLs(cfg, loomyModelFetchProxyURL(s, auth), baseURL, "")
	return client.FetchModels(ctx, session)
}

// loomyModelsForCatalogue maps a fetched catalogue through the same exclusion,
// alias, plugin and prefix pipeline the static registration uses.
func (s *Service) loomyModelsForCatalogue(auth *coreauth.Auth, catalogue []loomyauth.Model) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue))
	for _, entry := range catalogue {
		info := loomyCatalogueModelInfo(entry)
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
	models = applyOAuthModelAliasForAuth(cfg, constant.Loomy, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(constant.Loomy, models, excluded)
	models = applyOAuthSettingsForAuth(cfg, constant.Loomy, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// loomyCatalogueModelInfo builds the registry entry for one catalogue row,
// refining per-model metadata from the curated fallback catalogue.
func loomyCatalogueModelInfo(entry loomyauth.Model) *ModelInfo {
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return nil
	}
	info := registry.LookupStaticModelInfoByChannel(id, constant.Loomy)
	if info == nil {
		info = registry.LookupStaticModelInfo(id)
	}
	if info == nil {
		info = &ModelInfo{Object: "model", Created: loomyModelCreatedAt, OwnedBy: "iflytek"}
	}
	clone := *info
	clone.ID = id
	clone.Type = constant.Loomy
	if clone.Object == "" {
		clone.Object = "model"
	}
	if clone.Created == 0 {
		clone.Created = loomyModelCreatedAt
	}
	// The gateway's display name carries the points multiplier in its text; the
	// fallback catalog stores the same normalized name, so it wins when present.
	if strings.TrimSpace(clone.DisplayName) == "" {
		clone.DisplayName = entry.DisplayName
	}
	if clone.DisplayName == "" {
		clone.DisplayName = id
	}
	if entry.ContextLimit > 0 {
		clone.ContextLength = entry.ContextLimit
	}
	if len(entry.ReasoningLevels) > 0 {
		zeroAllowed := false
		for _, level := range entry.ReasoningLevels {
			if strings.EqualFold(level, "none") {
				zeroAllowed = true
				break
			}
		}
		clone.Thinking = &registry.ThinkingSupport{ZeroAllowed: zeroAllowed, Levels: append([]string(nil), entry.ReasoningLevels...)}
	}
	if entry.SupportsImage {
		clone.SupportedInputModalities = []string{"text", "image"}
	}
	return &clone
}

// loomyAttribute reads a credential attribute, falling back to metadata.
func loomyAttribute(auth *coreauth.Auth, key string) string {
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

// loomyModelFetchProxyURL resolves the proxy used for catalogue fetches.
func loomyModelFetchProxyURL(s *Service, auth *coreauth.Auth) string {
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

// refreshLoomyCatalogues reschedules accounts whose catalogue expired or whose
// last fetch failed. It never blocks on the fetch itself.
func (s *Service) refreshLoomyCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := loomyNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled || !strings.EqualFold(auth.Provider, constant.Loomy) {
			continue
		}
		snapshot := auth.Clone()
		key := loomyCatalogueKey(snapshot)
		loomyCatalogueMu.RLock()
		entry := loomyCatalogue[key]
		loomyCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeLoomyCatalogue(ctx, snapshot, constant.Loomy)
	}
}

// runLoomyModelRefresh periodically reschedules stale catalogues.
func (s *Service) runLoomyModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(loomyModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshLoomyCatalogues(ctx)
			timer.Reset(loomyModelRefreshScanInterval)
		}
	}
}
