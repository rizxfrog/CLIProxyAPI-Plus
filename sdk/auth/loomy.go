package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	loomyauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/loomy"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// LoomyAuthenticator implements the iFlytek Loomy phone SMS login flow.
//
// The flow is code-based rather than URL-based: there is no browser page to
// open. Login therefore always prompts for a phone number, then for the SMS
// code that the account service delivers. The management API exposes the same
// two steps separately so a web panel can render its own form.
type LoomyAuthenticator struct{}

// NewLoomyAuthenticator constructs a Loomy authenticator.
func NewLoomyAuthenticator() Authenticator { return &LoomyAuthenticator{} }

// Provider returns the Loomy provider key.
func (LoomyAuthenticator) Provider() string { return constant.Loomy }

// RefreshLead reports that no proactive refresh is scheduled.
//
// Loomy has no refresh endpoint: the session is issued for a fixed 14-day window
// and only a new login replaces it. Returning nil keeps the auto-refresh loop
// from probing the same credential forever — a lead would fall due immediately
// after each successful probe, because a probe cannot advance the expiry. The
// recorded expiry still makes the selector skip an expired credential, and an
// explicit refresh request still runs the liveness probe.
func (LoomyAuthenticator) RefreshLead() *time.Duration { return nil }

// Login drives the SMS flow and persists the resulting session.
func (a LoomyAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts == nil {
		opts = &LoginOptions{}
	}
	if opts.Prompt == nil {
		return nil, fmt.Errorf("loomy login requires an interactive prompt for the phone number and SMS code")
	}

	client := loomyauth.NewClient(cfg)

	fmt.Println("Starting Loomy (iFlytek) authentication...")
	phone, errPhone := opts.Prompt("Phone number (11 digits, without country code): ")
	if errPhone != nil {
		return nil, fmt.Errorf("loomy login: phone prompt failed: %w", errPhone)
	}
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("loomy login: phone number is required")
	}

	msgID, errCode := client.SendSMSCode(ctx, phone)
	if errCode != nil {
		return nil, fmt.Errorf("loomy login: send SMS code failed: %w", errCode)
	}
	fmt.Println("SMS code sent.")

	code, errPrompt := opts.Prompt("SMS verification code: ")
	if errPrompt != nil {
		return nil, fmt.Errorf("loomy login: code prompt failed: %w", errPrompt)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("loomy login: verification code is required")
	}

	token, errLogin := client.LoginWithSMSCode(ctx, phone, code, msgID)
	if errLogin != nil {
		return nil, fmt.Errorf("loomy login: SMS login failed: %w", errLogin)
	}
	// Best-effort daily allowance initialisation, matching the desktop client.
	// A failure here must not fail the login: the credential is already valid.
	if _, errClaim := client.ClaimDailyQuota(ctx, token.Session); errClaim != nil {
		log.Warnf("loomy: daily quota initialisation failed (login unaffected): %v", errClaim)
	}

	return loomyAuthRecord(cfg, token), nil
}

// loomyAuthRecord builds the persisted record for a successful login.
func loomyAuthRecord(cfg *config.Config, token *loomyauth.TokenData) *coreauth.Auth {
	return loomyAuthRecordWithBase(cfg, token, loomyauth.APIBase)
}

// loomyAuthRecordWithBase builds a record against an explicit business origin.
// It exists so self-hosted deployments and tests can target another host.
func loomyAuthRecordWithBase(cfg *config.Config, token *loomyauth.TokenData, baseURL string) *coreauth.Auth {
	now := time.Now()
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = loomyauth.APIBase
	}
	fileName := fmt.Sprintf("loomy-%d.json", now.UnixMilli())
	expired := loomyauth.SessionExpiry(now).UTC().Format(time.RFC3339)
	metadata := map[string]any{
		"type":         constant.Loomy,
		"auth_kind":    coreauth.AuthKindOAuth,
		"access_token": token.Session,
		"userid":       token.UserID,
		"base_url":     base,
		"expired":      expired,
		"timestamp":    now.UnixMilli(),
	}
	// Forced model prefixes are an access rule. Persist the provider prefix so
	// newly logged-in accounts can actually serve their advertised models.
	prefix := ""
	if cfg != nil && cfg.ForceModelPrefix {
		prefix = constant.Loomy
		metadata["prefix"] = prefix
	}
	if token.Phone != "" {
		metadata["phone"] = token.Phone
	}
	if token.Nickname != "" {
		metadata["nickname"] = token.Nickname
	}
	label := token.Phone
	if label == "" {
		label = token.UserID
	}
	return &coreauth.Auth{
		ID:       fileName,
		Provider: constant.Loomy,
		FileName: fileName,
		Label:    label,
		Prefix:   prefix,
		Metadata: metadata,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"access_token":             token.Session,
			"base_url":                 base,
		},
	}
}

// LoomySessionRecord is exported for the management login handlers, which run
// the SMS steps themselves and then persist through the shared store.
func LoomySessionRecord(cfg *config.Config, token *loomyauth.TokenData, baseURL string) *coreauth.Auth {
	return loomyAuthRecordWithBase(cfg, token, baseURL)
}
