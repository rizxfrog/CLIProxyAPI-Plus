package cliproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// lobsteraiServerModelRefreshScanInterval is how often accounts are rescanned
	// for an expired or unfetched catalogue. It bounds discovery latency; the
	// fetch itself only happens once the refresh interval has elapsed.
	lobsteraiServerModelRefreshScanInterval = time.Minute
	// lobsteraiServerModelRefreshInterval is how long a fetched catalogue is trusted.
	lobsteraiServerModelRefreshInterval = registry.ModelsRefreshInterval
	// lobsteraiServerModelRetryInterval bounds retries after a failed fetch.
	lobsteraiServerModelRetryInterval = 10 * time.Minute
	// lobsteraiServerModelCreatedAt is the registry entry creation timestamp.
	lobsteraiServerModelCreatedAt int64 = 1759276800
)

// lobsteraiServerCatalogueEntry records the last successful fetch for one account.
type lobsteraiServerCatalogueEntry struct {
	// expiresAt is when the fetched catalogue should be refreshed again.
	expiresAt time.Time
	// nextRetry gates retries after a failed fetch.
	nextRetry time.Time
}

var (
	lobsteraiServerNowFunc     = time.Now
	lobsteraiServerProbeWg     sync.WaitGroup
	lobsteraiServerCatalogueMu sync.RWMutex
	lobsteraiServerCatalogue   = make(map[string]lobsteraiServerCatalogueEntry)
	// lobsteraiServerProbeSlots bounds concurrent catalogue fetches.
	lobsteraiServerProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// lobsteraiServerCatalogueKey isolates catalogues by account and credential. The
// credential is part of the key because a rotated token can belong to a different
// account/plan.
func lobsteraiServerCatalogueKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	baseURL := ""
	credential := ""
	if auth.Attributes != nil {
		baseURL = strings.TrimRight(strings.TrimSpace(auth.Attributes["base_url"]), "/")
		credential = strings.TrimSpace(auth.Attributes["access_token"])
	}
	if credential == "" {
		credential = lobsteraiServerAttribute(auth, "access_token")
	}
	return fmt.Sprintf("%s|%s|%x", auth.ID, baseURL, sha256.Sum256([]byte(credential)))
}

// asyncProbeLobsteraiServerCatalogue launches a bounded, detached fetch for one
// account. The caller's context only gates admission: a request-scoped context
// would otherwise cancel the probe as soon as the registration call returns.
func (s *Service) asyncProbeLobsteraiServerCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
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
	snapshotKey := lobsteraiServerCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	lobsteraiServerProbeWg.Add(1)
	go func() {
		defer lobsteraiServerProbeWg.Done()
		lobsteraiServerProbeSlots <- struct{}{}
		defer func() { <-lobsteraiServerProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeLobsteraiServerCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeLobsteraiServerCatalogueProbe fetches the account's model catalogue and
// republishes the account's registered models when the fetch succeeds.
func (s *Service) executeLobsteraiServerCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	catalogue, err := s.fetchLobsteraiServerCatalogue(ctx, auth)
	if err != nil {
		if ctx.Err() == nil {
			log.Debugf("lobsterai-server: catalogue probe for %s failed: %v", auth.ID, err)
		}
		s.deferLobsteraiServerCatalogueRetry(snapshotKey)
		return
	}
	if len(catalogue) == 0 {
		// An empty catalogue is authoritative for the account, but revoking every
		// model would leave the credential unusable. Keep the fallback catalogue
		// and retry later instead.
		log.Debugf("lobsterai-server: catalogue for %s is empty; keeping fallback models", auth.ID)
		s.deferLobsteraiServerCatalogueRetry(snapshotKey)
		return
	}
	// The snapshot must still describe the live credential before publication.
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.lobsteraiServerModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	lobsteraiServerCatalogueMu.Lock()
	lobsteraiServerCatalogue[snapshotKey] = lobsteraiServerCatalogueEntry{
		expiresAt: lobsteraiServerNowFunc().Add(lobsteraiServerModelRefreshInterval),
	}
	lobsteraiServerCatalogueMu.Unlock()
}

// deferLobsteraiServerCatalogueRetry schedules a bounded retry after a failed or
// empty fetch.
func (s *Service) deferLobsteraiServerCatalogueRetry(snapshotKey string) {
	lobsteraiServerCatalogueMu.Lock()
	entry := lobsteraiServerCatalogue[snapshotKey]
	entry.nextRetry = lobsteraiServerNowFunc().Add(lobsteraiServerModelRetryInterval)
	lobsteraiServerCatalogue[snapshotKey] = entry
	lobsteraiServerCatalogueMu.Unlock()
}

// fetchLobsteraiServerCatalogue reads the account's model catalogue with the
// account access token.
func (s *Service) fetchLobsteraiServerCatalogue(ctx context.Context, auth *coreauth.Auth) ([]lobsteraiserverauth.Model, error) {
	if auth == nil {
		return nil, fmt.Errorf("lobsterai-server: auth is required")
	}
	accessToken := lobsteraiServerAttribute(auth, "access_token")
	if accessToken == "" {
		return nil, fmt.Errorf("lobsterai-server: no access token on auth %s", auth.ID)
	}
	backendURL := lobsteraiServerAttribute(auth, "backend_url")
	if backendURL == "" {
		backendURL = lobsteraiserverauth.DefaultBackendURL
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	client := lobsteraiserverauth.NewClientWithProxyURL(cfg, lobsteraiServerModelFetchProxyURL(s, auth), backendURL)
	return client.FetchModels(ctx, accessToken)
}

// lobsteraiServerModelsForCatalogue maps a fetched catalogue through the same
// exclusion, alias, plugin and prefix pipeline static registrations use, and
// carries each model's apiFormat so the executor can pick the upstream surface.
func (s *Service) lobsteraiServerModelsForCatalogue(auth *coreauth.Auth, catalogue []lobsteraiserverauth.Model) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue))
	for _, entry := range catalogue {
		info := lobsteraiServerCatalogueModelInfo(auth, entry)
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
	models = applyOAuthModelAliasForAuth(cfg, constant.LobsteraiServer, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(constant.LobsteraiServer, models, excluded)
	models = applyOAuthSettingsForAuth(cfg, constant.LobsteraiServer, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// lobsteraiServerCatalogueModelInfo builds the registry entry for one backend
// model. Per-model metadata is refined from the curated static catalogue when a
// matching entry exists.
func lobsteraiServerCatalogueModelInfo(auth *coreauth.Auth, entry lobsteraiserverauth.Model) *ModelInfo {
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return nil
	}
	var info *ModelInfo
	if auth != nil {
		if info = registry.LookupStaticModelInfoByChannel(id, constant.LobsteraiServer); info == nil {
			info = registry.LookupStaticModelInfo(id)
		}
	}
	if info == nil {
		info = &ModelInfo{Object: "model", Created: lobsteraiServerModelCreatedAt, OwnedBy: constant.LobsteraiServer}
	}
	clone := *info
	clone.ID = id
	clone.Type = constant.LobsteraiServer
	if clone.Object == "" {
		clone.Object = "model"
	}
	if clone.Created == 0 {
		clone.Created = lobsteraiServerModelCreatedAt
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
	// Publishing the declared levels is what makes a client's reasoning effort
	// reach the gateway: without model-level thinking metadata the shared pipeline
	// leaves `reasoning_effort` out of the upstream body entirely, so every
	// request silently ran at the backend default no matter what was asked for.
	//
	// The catalogue entry owns this metadata. The static registry is keyed by
	// model id across every provider, so a coincidentally equal id (this channel
	// and, say, the Kimi channel both ship "kimi-k3") would otherwise inherit a
	// level list the LobsterAI backend never declared.
	switch {
	case !entry.Thinking:
		// No thinking at all: an effort must not be forwarded.
		clone.Thinking = nil
	case len(entry.ReasoningLevels) == 0:
		// The backend reports thinking but exposes no choice, so there is no
		// vocabulary to validate against. An empty support record forwards exactly
		// what the client asked for without inventing either a restriction or an
		// allowance.
		clone.Thinking = &registry.ThinkingSupport{}
	default:
		zeroAllowed := false
		for _, level := range entry.ReasoningLevels {
			if strings.EqualFold(level, "off") || strings.EqualFold(level, "none") {
				zeroAllowed = true
				break
			}
		}
		clone.Thinking = &registry.ThinkingSupport{
			ZeroAllowed: zeroAllowed,
			Levels:      append([]string(nil), entry.ReasoningLevels...),
		}
	}
	return &clone
}

// lobsteraiServerAttribute reads a credential attribute, falling back to metadata.
func lobsteraiServerAttribute(auth *coreauth.Auth, keys ...string) string {
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

// lobsteraiServerModelFetchProxyURL resolves the proxy used for catalogue fetches.
func lobsteraiServerModelFetchProxyURL(s *Service, auth *coreauth.Auth) string {
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

// refreshLobsteraiServerCatalogues reschedules accounts whose catalogue expired
// or whose last fetch failed. It never blocks on the fetch itself.
func (s *Service) refreshLobsteraiServerCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := lobsteraiServerNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled || !strings.EqualFold(auth.Provider, constant.LobsteraiServer) {
			continue
		}
		snapshot := auth.Clone()
		key := lobsteraiServerCatalogueKey(snapshot)
		lobsteraiServerCatalogueMu.RLock()
		entry := lobsteraiServerCatalogue[key]
		lobsteraiServerCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeLobsteraiServerCatalogue(ctx, snapshot, constant.LobsteraiServer)
	}
}

// runLobsteraiServerModelRefresh periodically reschedules stale catalogues.
func (s *Service) runLobsteraiServerModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(lobsteraiServerModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshLobsteraiServerCatalogues(ctx)
			timer.Reset(lobsteraiServerModelRefreshScanInterval)
		}
	}
}
