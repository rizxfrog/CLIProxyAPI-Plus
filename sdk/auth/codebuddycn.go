package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codebuddycn"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

var codeBuddyCNRefreshLead = 5 * time.Minute

// codeBuddyAuthSpec parameterizes the CodeBuddy browser-polling authenticator so
// the CN and international gateways share identical login logic.
type codeBuddyAuthSpec struct {
	provider    string
	label       string
	fileNamePre string
	baseURL     string
	newClient   func(cfg *config.Config) *codebuddycn.Client
}

var codeBuddyCNAuthSpec = codeBuddyAuthSpec{
	provider:    "codebuddy-cn",
	label:       "CodeBuddy CN",
	fileNamePre: "codebuddy-cn",
	baseURL:     codebuddycn.APIBaseURL,
	newClient:   codebuddycn.NewClient,
}

var codeBuddyAIAuthSpec = codeBuddyAuthSpec{
	provider:    "codebuddy-ai",
	label:       "CodeBuddy AI",
	fileNamePre: "codebuddy-ai",
	baseURL:     codebuddycn.AIBaseURL,
	newClient:   codebuddycn.NewAIClient,
}

// CodeBuddyCNAuthenticator implements CodeBuddy CN's browser polling flow.
type CodeBuddyCNAuthenticator struct{}

// NewCodeBuddyCNAuthenticator constructs a CodeBuddy CN authenticator.
func NewCodeBuddyCNAuthenticator() Authenticator { return &CodeBuddyCNAuthenticator{} }

// Provider returns the CodeBuddy CN provider key.
func (CodeBuddyCNAuthenticator) Provider() string { return "codebuddy-cn" }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (CodeBuddyCNAuthenticator) RefreshLead() *time.Duration { return &codeBuddyCNRefreshLead }

// Login starts browser authorization and waits for CodeBuddy to issue tokens.
func (a CodeBuddyCNAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	return codeBuddyLogin(ctx, cfg, opts, codeBuddyCNAuthSpec)
}

// CodeBuddyAIAuthenticator implements the international CodeBuddy AI browser
// polling flow. The protocol is identical to CodeBuddy CN; only the gateway host
// and the account's authentication domain differ.
type CodeBuddyAIAuthenticator struct{}

// NewCodeBuddyAIAuthenticator constructs a CodeBuddy AI authenticator.
func NewCodeBuddyAIAuthenticator() Authenticator { return &CodeBuddyAIAuthenticator{} }

// Provider returns the CodeBuddy AI provider key.
func (CodeBuddyAIAuthenticator) Provider() string { return "codebuddy-ai" }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (CodeBuddyAIAuthenticator) RefreshLead() *time.Duration { return &codeBuddyCNRefreshLead }

// Login starts browser authorization and waits for CodeBuddy AI to issue tokens.
func (a CodeBuddyAIAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	return codeBuddyLogin(ctx, cfg, opts, codeBuddyAIAuthSpec)
}

func codeBuddyLogin(ctx context.Context, cfg *config.Config, opts *LoginOptions, spec codeBuddyAuthSpec) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts == nil {
		opts = &LoginOptions{}
	}
	client := spec.newClient(cfg)
	fmt.Printf("Starting %s authentication...\n", spec.label)
	device, err := client.StartDeviceFlow(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nTo authenticate, please visit:\n%s\n\n", device.AuthURL)
	if !opts.NoBrowser && browser.IsAvailable() {
		if errOpen := browser.OpenURL(device.AuthURL); errOpen != nil {
			log.Warnf("Failed to open browser automatically: %v", errOpen)
		} else {
			fmt.Println("Browser opened automatically.")
		}
	}
	fmt.Println("Waiting for authorization...")
	token, err := client.WaitForAuthorization(ctx, device)
	if err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("%s-%d.json", spec.fileNamePre, time.Now().UnixMilli())
	metadata := tokenMetadata(token, spec)
	return &coreauth.Auth{
		ID:       fileName,
		Provider: spec.provider,
		FileName: fileName,
		Label:    spec.label,
		Metadata: metadata,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"base_url":                 spec.baseURL,
		},
	}, nil
}

func tokenMetadata(token *codebuddycn.TokenData, spec codeBuddyAuthSpec) map[string]any {
	metadata := map[string]any{
		"type":          spec.provider,
		"auth_kind":     "oauth",
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"token_type":    token.TokenType,
		"expires_in":    token.ExpiresIn,
		"base_url":      spec.baseURL,
		"timestamp":     time.Now().UnixMilli(),
	}
	if !token.ExpiresAt.IsZero() {
		metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(token.RefreshToken) == "" {
		delete(metadata, "refresh_token")
	}
	return metadata
}
