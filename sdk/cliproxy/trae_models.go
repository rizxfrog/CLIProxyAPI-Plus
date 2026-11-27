package cliproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	traeauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/trae"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// traeModelRefreshScanInterval is how often accounts are rescanned for an
	// expired or unfetched catalogue.
	traeModelRefreshScanInterval = time.Minute
	// traeModelRefreshInterval is how long a fetched catalogue is trusted.
	traeModelRefreshInterval = registry.ModelsRefreshInterval
	// traeModelRetryInterval bounds retries after a failed fetch.
	traeModelRetryInterval = 10 * time.Minute
	// traeModelCreatedAt is the registry entry creation timestamp.
	traeModelCreatedAt int64 = 1767225600
)

// traeCatalogueEntry records the last successful fetch for one account.
type traeCatalogueEntry struct {
	expiresAt time.Time
	nextRetry time.Time
}

var (
	traeNowFunc     = time.Now
	traeProbeWg     sync.WaitGroup
	traeCatalogueMu sync.RWMutex
	traeCatalogue   = make(map[string]traeCatalogueEntry)
	// traeProbeSlots bounds concurrent catalogue fetches.
	traeProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// traeCatalogueKey isolates catalogues by account and credential. The token is
// part of the key because a refreshed token can belong to another account.
func traeCatalogueKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	baseURL := ""
	if auth.Attributes != nil {
		baseURL = strings.TrimRight(strings.TrimSpace(auth.Attributes["base_url"]), "/")
	}
	if baseURL == "" {
		baseURL = strings.TrimRight(traeAttribute(auth, "base_url"), "/")
	}
	credential := traeCredentialValue(auth, "access_token", "api_key")
	return fmt.Sprintf("%s|%s|%x", auth.ID, baseURL, sha256.Sum256([]byte(credential)))
}

// asyncProbeTraeCatalogue launches a bounded, detached catalogue fetch for one
// account. The caller's context only gates admission: a request-scoped context
// would cancel the probe as soon as the registration call returns.
func (s *Service) asyncProbeTraeCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
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
	snapshotKey := traeCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	traeProbeWg.Add(1)
	go func() {
		defer traeProbeWg.Done()
		traeProbeSlots <- struct{}{}
		defer func() { <-traeProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeTraeCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeTraeCatalogueProbe fetches the account's catalogue and republishes the
// account's registered models when the fetch succeeds.
func (s *Service) executeTraeCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	fetched, errFetch := s.fetchTraeCatalogue(ctx, auth)
	if errFetch != nil {
		if ctx.Err() == nil {
			log.Debugf("trae: catalogue probe for %s failed: %v", auth.ID, errFetch)
		}
		s.deferTraeCatalogueRetry(snapshotKey)
		return
	}
	s.reportTraeCatalogueSchema(auth, snapshotKey, fetched.Raw)
	catalogue := fetched.Models
	if len(catalogue) == 0 {
		// An empty catalogue is authoritative for the account, but revoking every
		// model would leave the credential unusable. Keep the current set and
		// retry later instead.
		log.Debugf("trae: catalogue for %s is empty; keeping current models", auth.ID)
		s.deferTraeCatalogueRetry(snapshotKey)
		return
	}
	// The snapshot must still describe the live credential before publication.
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.traeModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	traeCatalogueMu.Lock()
	traeCatalogue[snapshotKey] = traeCatalogueEntry{
		expiresAt: traeNowFunc().Add(traeModelRefreshInterval),
	}
	traeCatalogueMu.Unlock()
}

// traeSchemaReported tracks accounts whose catalogue schema has already been
// logged. The report is a one-shot diagnostic: it exists to confirm the inferred
// field names against the first real upstream payload, not to log every refresh.
var (
	traeSchemaReportedMu sync.Mutex
	traeSchemaReported   = make(map[string]struct{})
)

// reportTraeCatalogueSchema logs a value-free description of the upstream
// payload once per account credential. Field names on the live gateway were
// inferred from the desktop client DTOs, so the first real response is the only
// evidence that confirms them. Values are never logged: a live row carries
// account credentials and encrypted prompt blobs.
func (s *Service) reportTraeCatalogueSchema(auth *coreauth.Auth, snapshotKey string, raw []byte) {
	if auth == nil || len(raw) == 0 {
		return
	}
	traeSchemaReportedMu.Lock()
	if _, done := traeSchemaReported[snapshotKey]; done {
		traeSchemaReportedMu.Unlock()
		return
	}
	traeSchemaReported[snapshotKey] = struct{}{}
	traeSchemaReportedMu.Unlock()

	log.Infof("trae: catalogue schema for %s (one-shot, values redacted): %s", auth.ID, traeauth.DescribeModelList(raw))
}

// deferTraeCatalogueRetry schedules a bounded retry after a failed or empty
// fetch.
func (s *Service) deferTraeCatalogueRetry(snapshotKey string) {
	traeCatalogueMu.Lock()
	entry := traeCatalogue[snapshotKey]
	entry.nextRetry = traeNowFunc().Add(traeModelRetryInterval)
	traeCatalogue[snapshotKey] = entry
	traeCatalogueMu.Unlock()
}

// fetchTraeCatalogue reads the account's model catalogue from the SOLO agent
// gateway using the credential's desktop identity. The raw payload is returned
// alongside the parsed rows so the caller can emit a schema report.
func (s *Service) fetchTraeCatalogue(ctx context.Context, auth *coreauth.Auth) (*traeauth.FetchResult, error) {
	if auth == nil {
		return nil, fmt.Errorf("trae: auth is required")
	}
	headers := traeAgentHeaders(auth)
	if headers.AccessToken == "" {
		return nil, fmt.Errorf("trae: no access token on auth %s", auth.ID)
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	client := traeauth.NewModelClient(cfg, traeModelFetchProxyURL(s, auth))
	return client.FetchCatalogue(ctx, traeAttribute(auth, "base_url"), headers)
}

// traeAgentHeaders resolves the desktop identity used to authenticate the
// catalogue request. Mutable runtime fields in metadata (rotated access token,
// refreshed identity) take precedence over the immutable attributes, mirroring
// the executor's credential resolution.
func traeAgentHeaders(auth *coreauth.Auth) traeauth.AgentHeaders {
	return traeauth.AgentHeaders{
		AccessToken: traeCredentialValue(auth, "access_token", "api_key"),
		UID:         traeCredentialValue(auth, "uid"),
		MachineID:   traeCredentialValue(auth, "machine_id"),
		DeviceID:    traeCredentialValue(auth, "device_id"),
	}
}

// traeCredentialValue reads a credential field from metadata first, then from
// attributes.
func traeCredentialValue(auth *coreauth.Auth, keys ...string) string {
	if auth == nil {
		return ""
	}
	for _, key := range keys {
		if auth.Metadata != nil {
			if value, ok := auth.Metadata[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
		if auth.Attributes != nil {
			if value := strings.TrimSpace(auth.Attributes[key]); value != "" {
				return value
			}
		}
	}
	return ""
}

// traeModelsForCatalogue maps a fetched catalogue through the same exclusion,
// alias, plugin and prefix pipeline the static registration uses.
func (s *Service) traeModelsForCatalogue(auth *coreauth.Auth, catalogue []traeauth.Model) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue))
	for _, entry := range catalogue {
		info := traeCatalogueModelInfo(entry)
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
	models = applyOAuthModelAliasForAuth(cfg, constant.Trae, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(constant.Trae, models, excluded)
	models = applyOAuthSettingsForAuth(cfg, constant.Trae, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// traeCatalogueModelInfo builds the registry entry for one catalogue row,
// refining per-model metadata from the curated static catalogue when present.
func traeCatalogueModelInfo(entry traeauth.Model) *ModelInfo {
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return nil
	}
	info := registry.LookupStaticModelInfoByChannel(id, constant.Trae)
	if info == nil {
		info = registry.LookupStaticModelInfo(id)
	}
	if info == nil {
		info = &ModelInfo{Object: "model", Created: traeModelCreatedAt, OwnedBy: constant.Trae}
	}
	clone := *info
	clone.ID = id
	clone.Type = constant.Trae
	if clone.Object == "" {
		clone.Object = "model"
	}
	if clone.Created == 0 {
		clone.Created = traeModelCreatedAt
	}
	if display := strings.TrimSpace(entry.DisplayName); display != "" {
		clone.DisplayName = display
	}
	if clone.DisplayName == "" {
		clone.DisplayName = id
	}
	if entry.ContextLimit > 0 {
		clone.ContextLength = entry.ContextLimit
	}
	if entry.MaxTokens > 0 {
		clone.MaxCompletionTokens = entry.MaxTokens
	}
	if entry.Multimodal && len(clone.SupportedInputModalities) == 0 {
		clone.SupportedInputModalities = []string{"text", "image"}
	}
	return &clone
}

// traeAttribute reads a credential attribute, falling back to metadata. It is
// used for route settings where the immutable attribute wins.
func traeAttribute(auth *coreauth.Auth, keys ...string) string {
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

// traeModelFetchProxyURL resolves the proxy used for catalogue fetches.
func traeModelFetchProxyURL(s *Service, auth *coreauth.Auth) string {
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

// refreshTraeCatalogues reschedules accounts whose catalogue expired or whose
// last fetch failed. It never blocks on the fetch itself.
func (s *Service) refreshTraeCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := traeNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled || !strings.EqualFold(auth.Provider, constant.Trae) {
			continue
		}
		snapshot := auth.Clone()
		key := traeCatalogueKey(snapshot)
		traeCatalogueMu.RLock()
		entry := traeCatalogue[key]
		traeCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeTraeCatalogue(ctx, snapshot, constant.Trae)
	}
}

// runTraeModelRefresh periodically reschedules stale trae catalogues.
func (s *Service) runTraeModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(traeModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshTraeCatalogues(ctx)
			timer.Reset(traeModelRefreshScanInterval)
		}
	}
}
