package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	minimaxauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/minimax"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

var minimaxRefreshLead = 5 * time.Minute

// minimaxAuthSpec parameterizes the MiniMax MCode device-flow authenticator so
// the international (minimax.io) and mainland (minimax.cn) deployments share
// identical login logic.
type minimaxAuthSpec struct {
	provider    string
	label       string
	fileNamePre string
	region      minimaxauth.Region
}

var minimaxENAuthSpec = minimaxAuthSpec{
	provider:    "minimax",
	label:       "MiniMax Code (International)",
	fileNamePre: "minimax",
	region:      minimaxauth.RegionEN,
}

var minimaxCNAuthSpec = minimaxAuthSpec{
	provider:    "minimax-cn",
	label:       "MiniMax Code (China)",
	fileNamePre: "minimax-cn",
	region:      minimaxauth.RegionCN,
}

// MinimaxAuthenticator implements the MiniMax Code international
// (account.minimax.io) OAuth device authorization flow.
type MinimaxAuthenticator struct{}

// NewMinimaxAuthenticator constructs the international MiniMax authenticator.
func NewMinimaxAuthenticator() Authenticator { return &MinimaxAuthenticator{} }

// Provider returns the international provider key.
func (MinimaxAuthenticator) Provider() string { return "minimax" }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (MinimaxAuthenticator) RefreshLead() *time.Duration { return &minimaxRefreshLead }

// Login starts the MiniMax Code device flow and persists the resulting tokens.
func (a MinimaxAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	return minimaxLogin(ctx, cfg, opts, minimaxENAuthSpec)
}

// MinimaxCNAuthenticator implements the MiniMax Code mainland
// (account.minimax.cn) OAuth device authorization flow.
type MinimaxCNAuthenticator struct{}

// NewMinimaxCNAuthenticator constructs the mainland MiniMax authenticator.
func NewMinimaxCNAuthenticator() Authenticator { return &MinimaxCNAuthenticator{} }

// Provider returns the mainland provider key.
func (MinimaxCNAuthenticator) Provider() string { return "minimax-cn" }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (MinimaxCNAuthenticator) RefreshLead() *time.Duration { return &minimaxRefreshLead }

// Login starts the MiniMax Code mainland device flow and persists the tokens.
func (a MinimaxCNAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	return minimaxLogin(ctx, cfg, opts, minimaxCNAuthSpec)
}

func minimaxLogin(ctx context.Context, cfg *config.Config, opts *LoginOptions, spec minimaxAuthSpec) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts == nil {
		opts = &LoginOptions{}
	}
	client := minimaxauth.NewClient(cfg, spec.region)
	fmt.Printf("Starting %s authentication...\n", spec.label)
	device, errStart := client.StartDeviceAuthorization(ctx)
	if errStart != nil {
		return nil, errStart
	}
	authURL := device.AuthorizationURL()
	fmt.Printf("\nTo authenticate, please visit:\n%s\n", authURL)
	fmt.Printf("and enter the code: %s\n\n", device.UserCode)
	if !opts.NoBrowser && browser.IsAvailable() {
		if errOpen := browser.OpenURL(authURL); errOpen != nil {
			log.Warnf("Failed to open browser automatically: %v", errOpen)
		} else {
			fmt.Println("Browser opened automatically.")
		}
	}
	fmt.Println("Waiting for authorization...")
	token, errPoll := client.PollDeviceToken(ctx, device)
	if errPoll != nil {
		return nil, errPoll
	}
	fileName := fmt.Sprintf("%s-%d.json", spec.fileNamePre, time.Now().UnixMilli())
	return &coreauth.Auth{
		ID:       fileName,
		Provider: spec.provider,
		FileName: fileName,
		Label:    spec.label,
		Metadata: minimaxTokenMetadata(token, spec),
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"base_url":                 spec.region.InferenceBaseURL(),
			MinimaxRegionAttribute:     string(spec.region),
		},
	}, nil
}

// MinimaxRegionAttribute stores the account region on the auth record so the
// executor can resolve the correct inference base URL after restart.
const MinimaxRegionAttribute = "minimax_region"

func minimaxTokenMetadata(token *minimaxauth.TokenData, spec minimaxAuthSpec) map[string]any {
	metadata := map[string]any{
		"type":         spec.provider,
		"auth_kind":    "oauth",
		"access_token": token.AccessToken,
		"token_type":   token.TokenType,
		"expires_in":   token.ExpiresIn,
		"base_url":     spec.region.InferenceBaseURL(),
		"region":       string(spec.region),
		"timestamp":    time.Now().UnixMilli(),
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if !token.ExpiresAt.IsZero() {
		metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if token.AccountID != "" {
		metadata["account_id"] = token.AccountID
	}
	if token.Subject != "" {
		metadata["subject"] = token.Subject
	}
	return metadata
}
