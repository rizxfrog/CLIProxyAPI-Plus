package cliproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	codebuddycn "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codebuddycn"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// codebuddyModelRefreshScanInterval is how often accounts are rescanned for
	// an expired or unfetched catalogue.
	codebuddyModelRefreshScanInterval = time.Minute
	// codebuddyModelRefreshInterval is how long a fetched catalogue is trusted.
	codebuddyModelRefreshInterval = registry.ModelsRefreshInterval
	// codebuddyModelRetryInterval bounds retries after a failed fetch.
	codebuddyModelRetryInterval = 10 * time.Minute
	// codebuddyModelCreatedAt is the registry entry creation timestamp.
	codebuddyModelCreatedAt int64 = 1756684800
)

// codebuddyCatalogueEntry records the last successful fetch for one account.
type codebuddyCatalogueEntry struct {
	// expiresAt is when the fetched catalogue should be refreshed again.
	expiresAt time.Time
	// nextRetry gates retries after a failed fetch.
	nextRetry time.Time
}

var (
	codebuddyNowFunc     = time.Now
	codebuddyProbeWg     sync.WaitGroup
	codebuddyCatalogueMu sync.RWMutex
	codebuddyCatalogue   = make(map[string]codebuddyCatalogueEntry)
	// codebuddyProbeSlots bounds concurrent catalogue fetches.
	codebuddyProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// codebuddyCatalogueKey isolates catalogues by account and credential.
func codebuddyCatalogueKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	token := codebuddyAttribute(auth, "access_token")
	return fmt.Sprintf("%s|%x", auth.ID, sha256.Sum256([]byte(token)))
}

// asyncProbeCodeBuddyCatalogue launches a bounded, detached fetch for one
// account, for either CodeBuddy edition.
func (s *Service) asyncProbeCodeBuddyCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
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
	snapshotKey := codebuddyCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	codebuddyProbeWg.Add(1)
	go func() {
		defer codebuddyProbeWg.Done()
		codebuddyProbeSlots <- struct{}{}
		defer func() { <-codebuddyProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeCodeBuddyCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeCodeBuddyCatalogueProbe fetches the account's catalogue and republishes
// its models when the fetch succeeds.
func (s *Service) executeCodeBuddyCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	// An operator-provided model list is authoritative: a discovered catalogue
	// must not replace an explicit choice.
	if s.codeBuddyConfigPinsModels(auth) {
		return
	}
	catalogue, err := s.fetchCodeBuddyCatalogue(ctx, auth)
	if err != nil {
		if ctx.Err() == nil {
			log.Debugf("codebuddy: catalogue probe for %s failed: %v", auth.ID, err)
		}
		s.deferCodeBuddyCatalogueRetry(snapshotKey)
		return
	}
	if len(catalogue.Models) == 0 {
		// An empty catalogue would revoke every model; keep the fallback and
		// retry later instead.
		log.Debugf("codebuddy: catalogue for %s is empty; keeping fallback models", auth.ID)
		s.deferCodeBuddyCatalogueRetry(snapshotKey)
		return
	}
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.codebuddyModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	codebuddyCatalogueMu.Lock()
	codebuddyCatalogue[snapshotKey] = codebuddyCatalogueEntry{
		expiresAt: codebuddyNowFunc().Add(codebuddyModelRefreshInterval),
	}
	codebuddyCatalogueMu.Unlock()
}

// deferCodeBuddyCatalogueRetry schedules a bounded retry.
func (s *Service) deferCodeBuddyCatalogueRetry(snapshotKey string) {
	codebuddyCatalogueMu.Lock()
	entry := codebuddyCatalogue[snapshotKey]
	entry.nextRetry = codebuddyNowFunc().Add(codebuddyModelRetryInterval)
	codebuddyCatalogue[snapshotKey] = entry
	codebuddyCatalogueMu.Unlock()
}

// fetchCodeBuddyCatalogue reads the account's cloud product catalogue.
func (s *Service) fetchCodeBuddyCatalogue(ctx context.Context, auth *coreauth.Auth) (*codebuddycn.Catalog, error) {
	if auth == nil {
		return nil, fmt.Errorf("codebuddy: auth is required")
	}
	token := codebuddyAttribute(auth, "access_token")
	if token == "" {
		return nil, fmt.Errorf("codebuddy: no access token on auth %s", auth.ID)
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	// The two editions share one client implementation and differ only in host.
	var client *codebuddycn.Client
	if strings.EqualFold(auth.Provider, constant.CodeBuddyAI) {
		client = codebuddycn.NewAIClientWithProxyURL(cfg, auth.ProxyURL)
	} else {
		client = codebuddycn.NewClientWithProxyURL(cfg, auth.ProxyURL)
	}
	// A credential may pin a different origin (a regional mirror or a test
	// server); the compiled-in host is only the vendor default.
	client.WithOrigin(codebuddyAttribute(auth, "catalog_origin"))
	product := codebuddyAttribute(auth, "product")
	return client.FetchCatalog(ctx, token, product, codebuddyAttribute(auth, "product_version"))
}

// codebuddyModelsForCatalogue maps the catalogue through the shared exclusion,
// alias, plugin and prefix pipeline.
func (s *Service) codebuddyModelsForCatalogue(auth *coreauth.Auth, catalogue *codebuddycn.Catalog) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue.Models))
	for _, entry := range catalogue.Models {
		info := codebuddyCatalogueModelInfo(auth.Provider, entry)
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
	models = applyOAuthModelAliasForAuth(cfg, auth.Provider, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(auth.Provider, models, excluded)
	models = applyOAuthSettingsForAuth(cfg, auth.Provider, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// codebuddyCatalogueModelInfo builds the registry entry for one catalogue model.
//
// Metadata comes from the live catalogue rather than the static registry, which
// is keyed by model id across every provider and would otherwise lend this
// channel limits it never declared.
func codebuddyCatalogueModelInfo(provider string, entry codebuddycn.CatalogModel) *ModelInfo {
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return nil
	}
	info := &ModelInfo{
		Object:      "model",
		Created:     codebuddyModelCreatedAt,
		OwnedBy:     "codebuddy",
		Type:        provider,
		ID:          id,
		DisplayName: id,
	}
	if name := strings.TrimSpace(entry.Name); name != "" {
		info.DisplayName = name
	}
	if entry.ContextWindow > 0 {
		info.ContextLength = entry.ContextWindow
	}
	if entry.MaxOutputTokens > 0 {
		info.MaxCompletionTokens = entry.MaxOutputTokens
	}
	if entry.SupportsImages {
		info.SupportedInputModalities = []string{"TEXT", "IMAGE"}
	}
	// Publishing the declared efforts is what lets a client's reasoning choice
	// survive validation; an absent record would drop it from the request.
	if len(entry.ReasoningLevels) > 0 {
		info.Thinking = &registry.ThinkingSupport{
			ZeroAllowed: entry.CanDisableThinking,
			Levels:      append([]string(nil), entry.ReasoningLevels...),
		}
	}
	return info
}

// codebuddyAttribute reads a credential attribute, falling back to metadata.
func codebuddyAttribute(auth *coreauth.Auth, keys ...string) string {
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

// codeBuddyConfigPinsModels reports whether the credential's config entry lists
// its own models, for either edition.
func (s *Service) codeBuddyConfigPinsModels(auth *coreauth.Auth) bool {
	if s == nil || auth == nil {
		return false
	}
	if strings.EqualFold(auth.Provider, constant.CodeBuddyAI) {
		entry := s.resolveConfigCodeBuddyAIKey(auth)
		return entry != nil && len(entry.Models) > 0
	}
	entry := s.resolveConfigCodeBuddyCNKey(auth)
	return entry != nil && len(entry.Models) > 0
}

// refreshCodeBuddyCatalogues reschedules accounts whose catalogue expired or
// whose last fetch failed.
func (s *Service) refreshCodeBuddyCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := codebuddyNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled {
			continue
		}
		if !strings.EqualFold(auth.Provider, constant.CodeBuddyCN) && !strings.EqualFold(auth.Provider, constant.CodeBuddyAI) {
			continue
		}
		snapshot := auth.Clone()
		key := codebuddyCatalogueKey(snapshot)
		codebuddyCatalogueMu.RLock()
		entry := codebuddyCatalogue[key]
		codebuddyCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeCodeBuddyCatalogue(ctx, snapshot, strings.ToLower(auth.Provider))
	}
}

// runCodeBuddyModelRefresh periodically reschedules stale catalogues.
func (s *Service) runCodeBuddyModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(codebuddyModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshCodeBuddyCatalogues(ctx)
			timer.Reset(codebuddyModelRefreshScanInterval)
		}
	}
}
