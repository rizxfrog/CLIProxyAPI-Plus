package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	clineauth "github.com/router-for-me/CLIProxyAPI/v7/internal/auth/cline"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/util"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

var clineRefreshLead = 24 * time.Hour

// ClineAuthenticator implements the Cline (cline.bot) manual-paste OAuth login
// flow. Cline's authorization-code callback carries the credential bundle as a
// base64-encoded document rather than a bare code, and the desktop flow forces a
// 127.0.0.1 callback_url, so remote CLIProxyAPI deployments cannot receive the
// callback automatically. The user signs in, copies the full callback URL from
// the browser address bar, and pastes it back.
type ClineAuthenticator struct{}

// NewClineAuthenticator constructs a Cline authenticator.
func NewClineAuthenticator() Authenticator { return &ClineAuthenticator{} }

// Provider returns the Cline provider key.
func (ClineAuthenticator) Provider() string { return "cline" }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (ClineAuthenticator) RefreshLead() *time.Duration { return &clineRefreshLead }

// Login drives the Cline manual-paste login flow.
func (a ClineAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts == nil {
		opts = &LoginOptions{}
	}

	client := clineauth.NewClient(cfg)
	loginURL, errBuild := client.BuildLoginURL()
	if errBuild != nil {
		return nil, errBuild
	}

	fmt.Println("Starting Cline authentication...")
	if !opts.NoBrowser && browser.IsAvailable() {
		if errOpen := browser.OpenURL(loginURL.URL); errOpen != nil {
			log.Warnf("Failed to open browser automatically: %v", errOpen)
			fmt.Printf("Please open this URL manually:\n%s\n", loginURL.URL)
		} else {
			fmt.Println("Browser opened automatically.")
		}
	} else {
		util.PrintSSHTunnelInstructions(18080)
		fmt.Printf("Please open this URL to continue authentication:\n%s\n", loginURL.URL)
	}

	fmt.Println("After signing in, Cline redirects to a 127.0.0.1 address. Copy that full URL from the address bar.")
	if opts.Prompt == nil {
		return nil, fmt.Errorf("cliproxy auth: cline login requires an interactive prompt to paste the callback URL")
	}
	callbackURL, errPrompt := opts.Prompt("Paste the Cline callback URL: ")
	if errPrompt != nil {
		return nil, errPrompt
	}
	if strings.TrimSpace(callbackURL) == "" {
		return nil, fmt.Errorf("cliproxy auth: cline login cancelled")
	}

	token, errParse := client.ParseCallback(callbackURL)
	if errParse != nil {
		return nil, errParse
	}

	label := token.AccountLabel()
	fileName := clineFileName(token)
	metadata := map[string]any{
		"type":          "cline",
		"auth_kind":     "oauth",
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expires_at":    token.ExpiresAt,
		"email":         token.Email,
		"timestamp":     time.Now().UnixMilli(),
	}
	if strings.TrimSpace(token.RefreshToken) == "" {
		delete(metadata, "refresh_token")
	}
	if token.ExpiresAt <= 0 {
		delete(metadata, "expires_at")
	}
	if label == "" {
		label = "Cline"
	}

	return &coreauth.Auth{
		ID:       fileName,
		Provider: a.Provider(),
		FileName: fileName,
		Label:    label,
		Metadata: metadata,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"api_key":                  token.AccessToken,
			"base_url":                 clineauth.BaseURL,
		},
	}, nil
}

// clineFileName derives a stable, filesystem-safe credential file name.
func clineFileName(token *clineauth.TokenData) string {
	identity := strings.TrimSpace(token.Email)
	if identity == "" {
		identity = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	identity = strings.NewReplacer("/", "_", "\\", "_", "@", "_at_", " ", "_").Replace(identity)
	return fmt.Sprintf("cline-%s.json", identity)
}
