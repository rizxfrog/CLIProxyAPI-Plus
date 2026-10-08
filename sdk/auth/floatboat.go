package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	floatboatauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/floatboat"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

var floatboatRefreshLead = 5 * time.Minute

// FloatboatAuthenticator implements the FloatBoat (aoe.chat "Agent OS") login
// flow. FloatBoat signs in through the product web and returns the
// authorization code on the "aoe://" deep link registered by its desktop app.
// CLIProxyAPI cannot own that scheme, so the flow opens the sign-in page and
// asks the user to paste the resulting callback URL back (the same manual-paste
// shape used by Cline and TRAE).
type FloatboatAuthenticator struct{}

// NewFloatboatAuthenticator constructs a FloatBoat authenticator.
func NewFloatboatAuthenticator() Authenticator { return &FloatboatAuthenticator{} }

// Provider returns the FloatBoat provider key.
func (FloatboatAuthenticator) Provider() string { return floatboatauth.Provider }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (FloatboatAuthenticator) RefreshLead() *time.Duration { return &floatboatRefreshLead }

// Login drives the FloatBoat manual-paste login flow and mints the inference key.
func (a FloatboatAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts == nil {
		opts = &LoginOptions{}
	}

	state, errState := misc.GenerateRandomState()
	if errState != nil {
		return nil, fmt.Errorf("floatboat state generation failed: %w", errState)
	}

	client := floatboatauth.NewClient(cfg)
	signInURL := client.BuildSignInURL(state)

	fmt.Println("Starting FloatBoat authentication...")
	fmt.Printf("Visit the following URL to sign in:\n%s\n", signInURL)
	if !opts.NoBrowser && browser.IsAvailable() {
		if errOpen := browser.OpenURL(signInURL); errOpen != nil {
			log.Warnf("Failed to open browser automatically: %v", errOpen)
		} else {
			fmt.Println("Browser opened automatically.")
		}
	}

	callbackURL, errCallback := floatboatResolveCallback(opts)
	if errCallback != nil {
		return nil, errCallback
	}

	code, returnedState, errParse := floatboatParseCallback(callbackURL)
	if errParse != nil {
		return nil, errParse
	}
	if returnedState != "" && returnedState != state {
		return nil, fmt.Errorf("floatboat: state mismatch (expected %s, got %s)", state, returnedState)
	}

	token, errExchange := client.ExchangeCode(ctx, code, returnedState, "")
	if errExchange != nil {
		return nil, errExchange
	}

	// The profile and inference key are best-effort: a credential that only
	// carries the access token can still mint the key on first use.
	profile, errProfile := client.FetchUserProfile(ctx, token.AccessToken)
	if errProfile != nil {
		log.Warnf("floatboat: user profile fetch failed: %v", errProfile)
		profile = &floatboatauth.UserProfile{}
	}
	apiKey, errAPIKey := client.FetchNewAPIKey(ctx, token.AccessToken)
	if errAPIKey != nil {
		log.Warnf("floatboat: newapi key fetch failed: %v", errAPIKey)
	}

	fileName := floatboatFileName(profile)
	metadata := floatboatTokenMetadata(token, profile, apiKey)
	attributes := map[string]string{
		coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
		"base_url":                 floatboatauth.DefaultInferenceBaseURL,
		"access_token":             token.AccessToken,
	}
	if apiKey != "" {
		attributes["api_key"] = apiKey
	}

	label := floatboatLabel(profile)
	fmt.Printf("FloatBoat authentication successful (%s)\n", label)

	return &coreauth.Auth{
		ID:         fileName,
		Provider:   floatboatauth.Provider,
		FileName:   fileName,
		Label:      label,
		Metadata:   metadata,
		Attributes: attributes,
	}, nil
}

// floatboatResolveCallback obtains the deep-link callback URL, preferring an
// explicit metadata value (management flow) over an interactive prompt.
func floatboatResolveCallback(opts *LoginOptions) (string, error) {
	if opts.Metadata != nil {
		for _, key := range []string{"callback_url", "redirect_url", "callback"} {
			if value := strings.TrimSpace(opts.Metadata[key]); value != "" {
				return value, nil
			}
		}
	}
	if opts.Prompt == nil {
		return "", fmt.Errorf("cliproxy auth: floatboat login requires the callback URL; run interactively or provide callback_url metadata")
	}
	fmt.Println("After signing in, FloatBoat redirects to an aoe:// address. Copy that full URL from the address bar.")
	callbackURL, errPrompt := opts.Prompt("Paste the FloatBoat callback URL: ")
	if errPrompt != nil {
		return "", errPrompt
	}
	if strings.TrimSpace(callbackURL) == "" {
		return "", fmt.Errorf("cliproxy auth: floatboat login cancelled")
	}
	return callbackURL, nil
}

// floatboatParseCallback extracts code/state from the pasted callback, which may
// be an aoe:// deep link or a plain URL carrying the same query parameters.
func floatboatParseCallback(callbackURL string) (string, string, error) {
	trimmed := strings.TrimSpace(callbackURL)
	if trimmed == "" {
		return "", "", fmt.Errorf("floatboat: callback URL is empty")
	}
	parsed, errParse := url.Parse(trimmed)
	if errParse != nil {
		return "", "", fmt.Errorf("floatboat: parse callback URL: %w", errParse)
	}
	query := parsed.Query()
	if errCode := strings.TrimSpace(query.Get("error")); errCode != "" {
		description := strings.TrimSpace(query.Get("error_description"))
		if description != "" {
			return "", "", fmt.Errorf("floatboat: authorization failed (%s: %s)", errCode, description)
		}
		return "", "", fmt.Errorf("floatboat: authorization failed (%s)", errCode)
	}
	code := strings.TrimSpace(query.Get("code"))
	if code == "" {
		return "", "", fmt.Errorf("floatboat: callback URL missing code")
	}
	return code, strings.TrimSpace(query.Get("state")), nil
}

func floatboatTokenMetadata(token *floatboatauth.TokenData, profile *floatboatauth.UserProfile, apiKey string) map[string]any {
	metadata := map[string]any{
		"type":         floatboatauth.Provider,
		"auth_kind":    "oauth",
		"access_token": token.AccessToken,
		"base_url":     floatboatauth.DefaultInferenceBaseURL,
		"backend_url":  floatboatauth.DefaultBackendURL,
		"timestamp":    time.Now().UnixMilli(),
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		metadata["expires_in"] = token.ExpiresIn
	}
	if token.RefreshExpiresIn > 0 {
		metadata["refresh_expires_in"] = token.RefreshExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(apiKey) != "" {
		metadata["api_key"] = apiKey
	}
	if profile != nil {
		if profile.ID != "" {
			metadata["user_id"] = profile.ID
		}
		if profile.Email != "" {
			metadata["email"] = profile.Email
			metadata["user_email"] = profile.Email
		}
		if profile.Name != "" {
			metadata["user_name"] = profile.Name
		}
		if profile.Image != "" {
			metadata["user_avatar"] = profile.Image
		}
		if profile.MembershipLevel != "" {
			metadata["membership_level"] = profile.MembershipLevel
		}
		metadata["has_active_subscription"] = profile.HasSubscription
		metadata["quota_remaining"] = profile.Credits
	}
	return metadata
}

func floatboatLabel(profile *floatboatauth.UserProfile) string {
	if profile != nil {
		if profile.Email != "" {
			return profile.Email
		}
		if profile.ID != "" {
			return profile.ID
		}
	}
	return floatboatauth.Label
}

// floatboatFileName derives a collision-resistant, filesystem-safe credential name.
func floatboatFileName(profile *floatboatauth.UserProfile) string {
	identity := ""
	if profile != nil {
		identity = strings.TrimSpace(profile.ID)
		if identity == "" {
			identity = strings.TrimSpace(profile.Email)
		}
	}
	if identity == "" {
		identity = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", "@", "_at_", " ", "_").Replace(identity)
	return fmt.Sprintf("floatboat-%s.json", safe)
}
