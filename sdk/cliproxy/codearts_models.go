package cliproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
)

const (
	// codeartsModelRefreshScanInterval is how often accounts are rescanned for
	// an expired or unfetched catalogue. It bounds discovery latency; the fetch
	// itself only happens once the refresh interval has elapsed.
	codeartsModelRefreshScanInterval = time.Minute
	// codeartsModelRefreshInterval is how long a fetched catalogue is trusted.
	codeartsModelRefreshInterval = registry.ModelsRefreshInterval
	// codeartsModelRetryInterval bounds retries after a failed fetch.
	codeartsModelRetryInterval = 10 * time.Minute
	// codeartsModelCreatedAt is the registry entry creation timestamp.
	codeartsModelCreatedAt int64 = 1759276800
	// codeartsAgentTypePromptCenter is the Agent-Type the snap engine expects on
	// catalog calls. It is repeated here because the executor package owns the
	// canonical constant and importing it would drag the runtime into sdk.
	codeartsAgentTypePromptCenter = "PromptCenter"
)

// codeartsCatalogueEntry records the last successful fetch for one account.
type codeartsCatalogueEntry struct {
	// expiresAt is when the fetched catalogue should be refreshed again.
	expiresAt time.Time
	// nextRetry gates retries after a failed fetch.
	nextRetry time.Time
}

var (
	codeartsNowFunc     = time.Now
	codeartsProbeWg     sync.WaitGroup
	codeartsCatalogueMu sync.RWMutex
	codeartsCatalogue   = make(map[string]codeartsCatalogueEntry)
	// codeartsProbeSlots bounds concurrent catalogue fetches.
	codeartsProbeSlots = make(chan struct{}, modelRegistrationMaxWorkersPerCategory)
)

// codeartsCatalogueKey isolates catalogues by account and credential. The
// credential is part of the key because a rotated STS triple can belong to a
// different account or plan.
func codeartsCatalogueKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	credential := ""
	if auth.Attributes != nil {
		credential = strings.TrimSpace(auth.Attributes["api_key"])
	}
	if credential == "" {
		credential = codeartsAttribute(auth, "access_key", "api_key")
	}
	return fmt.Sprintf("%s|%x", auth.ID, sha256.Sum256([]byte(credential)))
}

// asyncProbeCodeArtsCatalogue launches a bounded, detached fetch for one
// account. The caller's context only gates admission: a request-scoped context
// would otherwise cancel the probe as soon as the registration call returns.
func (s *Service) asyncProbeCodeArtsCatalogue(ctx context.Context, auth *coreauth.Auth, providerKey string) {
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
	snapshotKey := codeartsCatalogueKey(snapshot)
	regEpoch := snapshot.RegistrationEpoch
	s.cfgMu.RLock()
	lifetime := s.antigravityContext
	s.cfgMu.RUnlock()
	if lifetime != nil && lifetime.Err() != nil {
		return
	}
	codeartsProbeWg.Add(1)
	go func() {
		defer codeartsProbeWg.Done()
		codeartsProbeSlots <- struct{}{}
		defer func() { <-codeartsProbeSlots }()
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if lifetime != nil {
			stop := context.AfterFunc(lifetime, cancel)
			defer stop()
			if lifetime.Err() != nil {
				return
			}
		}
		s.executeCodeArtsCatalogueProbe(probeCtx, snapshot, providerKey, snapshotKey, regEpoch)
	}()
}

// executeCodeArtsCatalogueProbe fetches the account's model catalogue and
// republishes the account's registered models when the fetch succeeds.
func (s *Service) executeCodeArtsCatalogueProbe(ctx context.Context, auth *coreauth.Auth, providerKey, snapshotKey string, regEpoch uint64) {
	// A credential whose config entry lists models explicitly is pinned to that
	// list: the operator chose it, so a discovered catalogue must not replace it.
	if entry := s.resolveConfigCodeArtsKey(auth); entry != nil && len(entry.Models) > 0 {
		return
	}
	catalogue, err := s.fetchCodeArtsCatalogue(ctx, auth)
	if err != nil {
		if ctx.Err() == nil {
			log.Debugf("codearts: catalogue probe for %s failed: %v", auth.ID, err)
		}
		s.deferCodeArtsCatalogueRetry(snapshotKey)
		return
	}
	if len(catalogue) == 0 {
		// An empty catalogue is authoritative for the account, but revoking every
		// model would leave the credential unusable. Keep the fallback catalogue
		// and retry later instead.
		log.Debugf("codearts: catalogue for %s is empty; keeping fallback models", auth.ID)
		s.deferCodeArtsCatalogueRetry(snapshotKey)
		return
	}
	// The snapshot must still describe the live credential before publication.
	if s.coreManager != nil {
		current, exists := s.coreManager.GetByID(auth.ID)
		if !exists || current == nil || current.Disabled || current.Provider != auth.Provider || current.RegistrationEpoch != regEpoch {
			return
		}
	}
	models := s.codeartsModelsForCatalogue(auth, catalogue)
	if len(models) == 0 {
		return
	}
	s.registerResolvedModelsForAuth(auth, providerKey, models)
	codeartsCatalogueMu.Lock()
	codeartsCatalogue[snapshotKey] = codeartsCatalogueEntry{
		expiresAt: codeartsNowFunc().Add(codeartsModelRefreshInterval),
	}
	codeartsCatalogueMu.Unlock()
}

// deferCodeArtsCatalogueRetry schedules a bounded retry after a failed or empty
// fetch.
func (s *Service) deferCodeArtsCatalogueRetry(snapshotKey string) {
	codeartsCatalogueMu.Lock()
	entry := codeartsCatalogue[snapshotKey]
	entry.nextRetry = codeartsNowFunc().Add(codeartsModelRetryInterval)
	codeartsCatalogue[snapshotKey] = entry
	codeartsCatalogueMu.Unlock()
}

// fetchCodeArtsCatalogue reads the account's model catalogue with a signed
// request. The STS triple is rotated first when it has entered the refresh
// window, because an expired token cannot sign anything.
func (s *Service) fetchCodeArtsCatalogue(ctx context.Context, auth *coreauth.Auth) ([]*codeartsauth.Model, error) {
	if auth == nil {
		return nil, fmt.Errorf("codearts: auth is required")
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	creds, rotated, errCreds := codeartsCatalogueCredentials(ctx, cfg, auth)
	if errCreds != nil {
		return nil, errCreds
	}
	if rotated != nil {
		s.persistCodeArtsRotatedCredentials(ctx, auth, rotated)
	}
	client := codeartsauth.NewClientWithProxyURL(cfg, auth.ProxyURL)
	// A credential may pin a different snap engine origin (a regional mirror or
	// a test server); the compiled-in default is only the vendor's primary host.
	client.WithSnapHost(codeartsAttribute(auth, "snap_engine_url"))
	client.WithBenefitURL(codeartsAttribute(auth, "benefit_api_url"))
	language := codeartsAttribute(auth, "llm_language")
	return client.FetchModels(ctx, creds, language)
}

// codeartsCatalogueCredentials returns a usable signed-request triple, rotating
// the stored one when it is inside the refresh window.
//
// Rotation needs the refresh token, the PKCE verifier and the original DPoP key
// pair: Huawei Cloud STS re-validates proof of possession, so a credential
// missing the key cannot be rotated and is returned as-is (the upstream then
// reports expiry, which is the honest answer).
func codeartsCatalogueCredentials(ctx context.Context, cfg *config.Config, auth *coreauth.Auth) (codeartsauth.Credentials, *codeartsauth.TokenData, error) {
	creds := codeartsauth.Credentials{
		AccessKey:     codeartsAttribute(auth, "access_key", "api_key"),
		SecretKey:     codeartsAttribute(auth, "secret_key"),
		SecurityToken: codeartsAttribute(auth, "security_token"),
	}
	if !codeartsCatalogueCredentialExpired(auth) {
		if creds.AccessKey == "" || creds.SecretKey == "" {
			return creds, nil, fmt.Errorf("codearts: credential %s is incomplete", auth.ID)
		}
		return creds, nil, nil
	}
	refreshToken := codeartsAttribute(auth, "refresh_token")
	privateKey := codeartsAttribute(auth, "dpop_private_key")
	if refreshToken == "" || privateKey == "" {
		// Nothing to rotate with; hand back what is stored and let the upstream
		// decide whether it still works.
		return creds, nil, nil
	}
	keyPair := &codeartsauth.DpopKeyPair{
		PrivateKey: privateKey,
		PublicKey:  codeartsAttribute(auth, "dpop_public_key"),
	}
	client := codeartsauth.NewClientWithProxyURL(cfg, auth.ProxyURL)
	token, errRefresh := client.Refresh(ctx, refreshToken, codeartsAttribute(auth, "code_verifier"), keyPair)
	if errRefresh != nil {
		return codeartsauth.Credentials{}, nil, errRefresh
	}
	return codeartsauth.Credentials{
		AccessKey:     token.AccessKey,
		SecretKey:     token.SecretKey,
		SecurityToken: token.SecurityToken,
	}, token, nil
}

// codeartsCatalogueCredentialExpired reports whether the stored security token
// is inside the refresh window. A missing expiry is treated as valid so the
// upstream stays authoritative.
func codeartsCatalogueCredentialExpired(auth *coreauth.Auth) bool {
	raw := codeartsAttribute(auth, "expired")
	if raw == "" {
		return false
	}
	expiry, errParse := time.Parse(time.RFC3339, raw)
	if errParse != nil || expiry.IsZero() {
		return false
	}
	return time.Now().Add(codeartsauth.RefreshWindowSeconds * time.Second).After(expiry)
}

// persistCodeArtsRotatedCredentials writes a rotated triple back onto the auth
// record and asks the manager to save it, mirroring the executor.
func (s *Service) persistCodeArtsRotatedCredentials(ctx context.Context, auth *coreauth.Auth, token *codeartsauth.TokenData) {
	if auth == nil || token == nil {
		return
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["access_key"] = token.AccessKey
	auth.Metadata["secret_key"] = token.SecretKey
	auth.Metadata["security_token"] = token.SecurityToken
	if token.RefreshToken != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if !token.ExpiresAt.IsZero() {
		auth.Metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)
	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes["api_key"] = token.AccessKey
	auth.Attributes["secret_key"] = token.SecretKey
	auth.Attributes["security_token"] = token.SecurityToken
	if s != nil && s.coreManager != nil {
		if _, errUpdate := s.coreManager.Update(ctx, auth); errUpdate != nil {
			log.Debugf("codearts: persisting rotated credential for %s failed: %v", auth.ID, errUpdate)
		}
	}
}

// codeartsModelsForCatalogue maps a fetched catalogue through the same
// exclusion, alias, plugin and prefix pipeline static registrations use.
func (s *Service) codeartsModelsForCatalogue(auth *coreauth.Auth, catalogue []*codeartsauth.Model) []*ModelInfo {
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	if cfg == nil {
		cfg = &config.Config{}
	}
	models := make([]*ModelInfo, 0, len(catalogue))
	for _, entry := range catalogue {
		info := codeartsCatalogueModelInfo(entry)
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
	models = applyOAuthModelAliasForAuth(cfg, constant.CodeArts, auth.AuthKind(), auth.Attributes, models)
	models = s.appendPluginModels(constant.CodeArts, models, excluded)
	models = applyOAuthSettingsForAuth(cfg, constant.CodeArts, auth.AuthKind(), models)
	return applyModelPrefixes(models, auth.Prefix, cfg.ForceModelPrefix)
}

// codeartsCatalogueModelInfo builds the registry entry for one backend model.
//
// Metadata comes from the live catalogue, not the static registry: the static
// list is keyed by model id across every provider, so a coincidentally equal id
// would otherwise inherit limits and capabilities CodeArts never declared.
func codeartsCatalogueModelInfo(entry *codeartsauth.Model) *ModelInfo {
	if entry == nil {
		return nil
	}
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		return nil
	}
	info := &ModelInfo{
		Object:      "model",
		Created:     codeartsModelCreatedAt,
		OwnedBy:     constant.CodeArts,
		Type:        constant.CodeArts,
		ID:          id,
		DisplayName: strings.TrimSpace(entry.Name),
	}
	if info.DisplayName == "" {
		info.DisplayName = id
	}
	if description := strings.TrimSpace(entry.Description); description != "" {
		info.Description = description
	}
	if entry.ContextWindow > 0 {
		info.ContextLength = entry.ContextWindow
	}
	if entry.OutputContextWindow > 0 {
		info.MaxCompletionTokens = entry.OutputContextWindow
	}
	if entry.SupportsImages {
		info.SupportedInputModalities = []string{"TEXT", "IMAGE"}
	}
	// The catalogue owns thinking metadata. Declaring the level list is what
	// lets a client's reasoning effort survive validation; leaving it nil means
	// an effort is not forwarded at all.
	if level := strings.TrimSpace(entry.ThinkLevel); level != "" {
		info.Thinking = &registry.ThinkingSupport{
			ZeroAllowed: true,
			Levels:      []string{"low", "medium", "high"},
		}
	}
	return info
}

// codeartsAttribute reads a credential attribute, falling back to metadata.
func codeartsAttribute(auth *coreauth.Auth, keys ...string) string {
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

// refreshCodeArtsCatalogues reschedules accounts whose catalogue expired or
// whose last fetch failed. It never blocks on the fetch itself.
func (s *Service) refreshCodeArtsCatalogues(ctx context.Context) {
	if s == nil || s.coreManager == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := codeartsNowFunc()
	for _, auth := range s.coreManager.List() {
		if ctx.Err() != nil {
			return
		}
		if auth == nil || auth.Disabled || !strings.EqualFold(auth.Provider, constant.CodeArts) {
			continue
		}
		snapshot := auth.Clone()
		key := codeartsCatalogueKey(snapshot)
		codeartsCatalogueMu.RLock()
		entry := codeartsCatalogue[key]
		codeartsCatalogueMu.RUnlock()
		if now.Before(entry.expiresAt) || now.Before(entry.nextRetry) {
			continue
		}
		s.asyncProbeCodeArtsCatalogue(ctx, snapshot, constant.CodeArts)
	}
}

// runCodeArtsModelRefresh periodically reschedules stale catalogues.
func (s *Service) runCodeArtsModelRefresh(ctx context.Context) {
	if s == nil || s.antigravityHomeEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(codeartsModelRefreshScanInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.refreshCodeArtsCatalogues(ctx)
			timer.Reset(codeartsModelRefreshScanInterval)
		}
	}
}
