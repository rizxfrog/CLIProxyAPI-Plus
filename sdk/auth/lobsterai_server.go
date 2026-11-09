package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// lobsteraiServerRefreshLead refreshes shortly before expiry. The backend issues
// short-lived access tokens, so a proactive lead keeps a request from paying for
// the refresh.
var lobsteraiServerRefreshLead = 5 * time.Minute

// LobsteraiServerAuthenticator implements the LobsterAI (NetEase/Youdao) login
// flow.
//
// LobsterAI signs in through the product backend and returns the authorization
// code to a loopback callback (http://127.0.0.1:<port>/auth/callback). This
// authenticator serves that loopback callback and also accepts a pasted callback
// URL, so the flow works for local CLI users and for remote/headless operators
// whose browser cannot reach the loopback port.
type LobsteraiServerAuthenticator struct {
	// CallbackPort pins the loopback port (0 selects an ephemeral port).
	CallbackPort int
}

// NewLobsteraiServerAuthenticator constructs a LobsterAI authenticator.
func NewLobsteraiServerAuthenticator() Authenticator { return &LobsteraiServerAuthenticator{} }

// Provider returns the LobsterAI provider key.
func (LobsteraiServerAuthenticator) Provider() string { return lobsteraiserverauth.Provider }

// RefreshLead instructs the runtime to refresh shortly before expiry.
func (LobsteraiServerAuthenticator) RefreshLead() *time.Duration { return &lobsteraiServerRefreshLead }

// Login drives the LobsterAI loopback login flow and persists the credential.
func (a LobsteraiServerAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
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
		return nil, fmt.Errorf("lobsterai-server state generation failed: %w", errState)
	}

	callbackPort := a.CallbackPort
	if opts.CallbackPort > 0 {
		callbackPort = opts.CallbackPort
	}

	client := lobsteraiserverauth.NewClient(cfg)

	// Start the loopback callback server first so the redirect URI carries the
	// real bound port.
	oauthServer := lobsteraiserverauth.NewOAuthServer(callbackPort)
	boundPort, errStart := oauthServer.Start()
	if errStart != nil {
		return nil, fmt.Errorf("lobsterai-server callback server failed: %w", errStart)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if stopErr := oauthServer.Stop(stopCtx); stopErr != nil {
			log.Warnf("lobsterai-server oauth callback server stop error: %v", stopErr)
		}
	}()

	redirectURI := oauthServer.RedirectURI()
	loginURL := client.BuildLoginURL(state, redirectURI)

	fmt.Println("Starting LobsterAI authentication...")
	fmt.Printf("Visit the following URL to sign in:\n%s\n", loginURL)
	if !opts.NoBrowser {
		if !browser.IsAvailable() {
			log.Warn("No browser available; please open the URL manually")
			util.PrintSSHTunnelInstructions(boundPort)
		} else if errOpen := browser.OpenURL(loginURL); errOpen != nil {
			log.Warnf("Failed to open browser automatically: %v", errOpen)
			util.PrintSSHTunnelInstructions(boundPort)
		} else {
			fmt.Println("Browser opened automatically.")
		}
	} else {
		util.PrintSSHTunnelInstructions(boundPort)
	}

	fmt.Println("Waiting for LobsterAI authentication callback...")

	callbackCh := make(chan *lobsteraiserverauth.OAuthResult, 1)
	callbackErrCh := make(chan error, 1)
	go func() {
		result, errWait := oauthServer.WaitForCallback(5 * time.Minute)
		if errWait != nil {
			callbackErrCh <- errWait
			return
		}
		callbackCh <- result
	}()

	// A pasted callback URL is the escape hatch for a remote operator whose
	// browser cannot reach the loopback listener.
	var promptTimer *time.Timer
	var promptC <-chan time.Time
	var manualInputCh <-chan string
	var manualInputErrCh <-chan error
	if opts.Prompt != nil {
		promptTimer = time.NewTimer(15 * time.Second)
		promptC = promptTimer.C
		defer promptTimer.Stop()
	}

	var result *lobsteraiserverauth.OAuthResult
	var manualDescription string

waitForCallback:
	for {
		select {
		case result = <-callbackCh:
			break waitForCallback
		case errWait := <-callbackErrCh:
			return nil, errWait
		case <-promptC:
			promptC = nil
			if promptTimer != nil {
				promptTimer.Stop()
			}
			select {
			case result = <-callbackCh:
				break waitForCallback
			case errWait := <-callbackErrCh:
				return nil, errWait
			default:
			}
			manualInputCh, manualInputErrCh = misc.AsyncPrompt(opts.Prompt, "Paste the LobsterAI callback URL (or press Enter to keep waiting): ")
			continue
		case input := <-manualInputCh:
			manualInputCh = nil
			manualInputErrCh = nil
			parsed, errParse := misc.ParseOAuthCallback(input)
			if errParse != nil {
				fmt.Printf("Invalid callback URL: %v\n", errParse)
				manualInputCh, manualInputErrCh = misc.AsyncPrompt(opts.Prompt, "Paste the LobsterAI callback URL (or press Enter to keep waiting): ")
				continue
			}
			if parsed == nil {
				manualInputCh, manualInputErrCh = misc.AsyncPrompt(opts.Prompt, "Paste the LobsterAI callback URL (or press Enter to keep waiting): ")
				continue
			}
			manualDescription = parsed.ErrorDescription
			result = &lobsteraiserverauth.OAuthResult{
				Code:  parsed.Code,
				State: parsed.State,
				Error: parsed.Error,
			}
			break waitForCallback
		case errManual := <-manualInputErrCh:
			return nil, errManual
		}
	}

	if result.Error != "" {
		if manualDescription != "" {
			return nil, fmt.Errorf("lobsterai-server authorization failed (%s: %s)", result.Error, manualDescription)
		}
		return nil, fmt.Errorf("lobsterai-server authorization failed (%s)", result.Error)
	}
	if result.State != "" && result.State != state {
		return nil, fmt.Errorf("lobsterai-server: state mismatch (expected %s, got %s)", state, result.State)
	}

	token, user, quota, errExchange := client.ExchangeCode(ctx, result.Code, result.State)
	if errExchange != nil {
		return nil, errExchange
	}

	// The exchange already carries the identity; a profile read is best-effort
	// and only refines it.
	if profile, errProfile := client.FetchUserProfile(ctx, token.AccessToken); errProfile == nil {
		if profile != nil {
			user = mergeLobsteraiServerUser(user, profile)
		}
	} else {
		log.Warnf("lobsterai-server: user profile fetch failed: %v", errProfile)
	}

	// An enterprise account must send its identity on every request, so the
	// context is read once at login and persisted with the credential. Failure is
	// not fatal: the account then behaves as personal, which is the correct
	// default for the common case.
	enterprise, errEnterprise := client.FetchEnterpriseContext(ctx, token.AccessToken)
	if errEnterprise != nil {
		log.Warnf("lobsterai-server: enterprise context fetch failed: %v", errEnterprise)
	}

	fileName := lobsteraiServerFileName(user)
	fmt.Printf("LobsterAI authentication successful (%s)\n", lobsteraiServerLabel(user))

	return &coreauth.Auth{
		ID:         fileName,
		Provider:   lobsteraiserverauth.Provider,
		FileName:   fileName,
		Label:      lobsteraiServerLabel(user),
		Metadata:   lobsteraiServerTokenMetadata(token, user, quota, enterprise),
		Attributes: lobsteraiServerAttributes(token),
	}, nil
}

// lobsteraiServerAttributes builds the runtime attributes the executor and
// synthesizer read back. base_url is the OpenAI-compatible proxy surface the
// executor posts to.
func lobsteraiServerAttributes(token *lobsteraiserverauth.TokenData) map[string]string {
	return map[string]string{
		coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
		"base_url":                 lobsteraiserverauth.DefaultProxyBaseURL,
		"backend_url":              lobsteraiserverauth.DefaultBackendURL,
		"access_token":             token.AccessToken,
	}
}

// lobsteraiServerTokenMetadata builds the persisted credential metadata. The
// schema matches what internal/watcher/synthesizer rehydrates on restart.
func lobsteraiServerTokenMetadata(token *lobsteraiserverauth.TokenData, user *lobsteraiserverauth.User, quota *lobsteraiserverauth.Quota, enterprise *lobsteraiserverauth.EnterpriseContext) map[string]any {
	metadata := map[string]any{
		"type":           lobsteraiserverauth.Provider,
		"auth_kind":      "oauth",
		"access_token":   token.AccessToken,
		"base_url":       lobsteraiserverauth.DefaultProxyBaseURL,
		"backend_url":    lobsteraiserverauth.DefaultBackendURL,
		"client_version": lobsteraiserverauth.ClientVersion,
		"timestamp":      time.Now().UnixMilli(),
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		metadata["refresh_token"] = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		metadata["expires_in"] = token.ExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	applyLobsteraiServerUserMetadata(metadata, user)
	applyLobsteraiServerQuotaMetadata(metadata, quota)
	applyLobsteraiServerEnterpriseMetadata(metadata, enterprise)
	return metadata
}

// applyLobsteraiServerEnterpriseMetadata records the joined-enterprise identity.
//
// A personal account carries no such fields, which is what makes the proxy omit
// the two enterprise headers: the absence of a persisted context is the signal,
// not an empty string that could be mistaken for an id.
func applyLobsteraiServerEnterpriseMetadata(metadata map[string]any, enterprise *lobsteraiserverauth.EnterpriseContext) {
	if metadata == nil || enterprise == nil || enterprise.EnterpriseID == "" {
		return
	}
	metadata["account_mode"] = lobsteraiserverauth.AccountModeEnterprise
	metadata["enterprise_id"] = enterprise.EnterpriseID
	if enterprise.EnterpriseName != "" {
		metadata["enterprise_name"] = enterprise.EnterpriseName
	}
	if enterprise.Role != "" {
		metadata["enterprise_role"] = enterprise.Role
	}
	if enterprise.MemberID != "" {
		metadata["enterprise_member_id"] = enterprise.MemberID
	}
}

// applyLobsteraiServerUserMetadata adds the account identity fields.
func applyLobsteraiServerUserMetadata(metadata map[string]any, user *lobsteraiserverauth.User) {
	if metadata == nil || user == nil {
		return
	}
	if user.ID != "" {
		metadata["user_id"] = user.ID
	}
	if user.Email != "" {
		metadata["email"] = user.Email
		metadata["user_email"] = user.Email
	}
	if user.Name != "" {
		metadata["user_name"] = user.Name
	}
	if user.Avatar != "" {
		metadata["user_avatar"] = user.Avatar
	}
}

// applyLobsteraiServerQuotaMetadata adds the login-time credit snapshot. Absent
// values stay absent rather than being reported as zero or unlimited.
func applyLobsteraiServerQuotaMetadata(metadata map[string]any, quota *lobsteraiserverauth.Quota) {
	if metadata == nil || quota == nil {
		return
	}
	if quota.SubscriptionStatus != "" {
		metadata["subscription_status"] = quota.SubscriptionStatus
	}
	if quota.CreditsKnown {
		metadata["quota_remaining"] = quota.Credits
	}
	if quota.PurchasedCredits > 0 {
		metadata["purchased_credits"] = quota.PurchasedCredits
	}
}

// mergeLobsteraiServerUser fills missing identity fields from the profile read.
func mergeLobsteraiServerUser(base, extra *lobsteraiserverauth.User) *lobsteraiserverauth.User {
	if base == nil {
		return extra
	}
	if extra == nil {
		return base
	}
	if base.ID == "" {
		base.ID = extra.ID
	}
	if base.Email == "" {
		base.Email = extra.Email
	}
	if base.Name == "" {
		base.Name = extra.Name
	}
	if base.Avatar == "" {
		base.Avatar = extra.Avatar
	}
	if base.Nickname == "" {
		base.Nickname = extra.Nickname
	}
	return base
}

// lobsteraiServerLabel derives a display label for the credential.
func lobsteraiServerLabel(user *lobsteraiserverauth.User) string {
	if user != nil {
		if user.Email != "" {
			return user.Email
		}
		if user.Nickname != "" {
			return user.Nickname
		}
		if user.Name != "" {
			return user.Name
		}
		if user.ID != "" {
			return user.ID
		}
	}
	return lobsteraiserverauth.Label
}

// lobsteraiServerFileName derives a collision-resistant, filesystem-safe
// credential name.
func lobsteraiServerFileName(user *lobsteraiserverauth.User) string {
	identity := ""
	if user != nil {
		identity = strings.TrimSpace(user.ID)
		if identity == "" {
			identity = strings.TrimSpace(user.Email)
		}
	}
	if identity == "" {
		identity = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", "@", "_at_", " ", "_", ":", "_").Replace(identity)
	return fmt.Sprintf("lobsterai-server-%s.json", safe)
}
