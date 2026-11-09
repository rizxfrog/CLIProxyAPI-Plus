package checkin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// lobsterAIAdapter claims the LobsterAI daily check-in reward.
//
// The reward is a backend activity the desktop client drives, so this adapter
// runs only for a credential whose operator explicitly enabled the provider in
// the check-in settings; nothing else in CLIProxyAPI claims it (the quota
// endpoint reports the activity state read-only).
//
// The adapter reads the activity state before acting, so a schedule that fires
// twice in one day cannot double-claim: the activity's own claimedToday flag and
// the terminal duplicate-claim code both map to already_done rather than to a
// failure.
type lobsterAIAdapter struct{}

// NewLobsterAIAdapter builds the LobsterAI daily check-in adapter.
func NewLobsterAIAdapter() Adapter { return lobsterAIAdapter{} }

func (lobsterAIAdapter) Provider() string { return constant.LobsteraiServer }

// Identity keys the daily observation on the account id, so re-exporting the
// same account with a rotated access token keeps one schedule entry.
func (lobsterAIAdapter) Identity(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	for _, key := range []string{"user_id", "uid", "email"} {
		if value := lobsterAIAttribute(auth, key); value != "" {
			return value
		}
	}
	return ""
}

// RequiredMetadata lists the credential fields addressable from a check-in
// request header template.
func (lobsterAIAdapter) RequiredMetadata() []string {
	return []string{"access_token", "user_id", "backend_url", "client_version", "account_mode", "enterprise_id", "activity_platform"}
}

func (lobsterAIAdapter) CheckIn(ctx context.Context, cfg *config.Config, auth *coreauth.Auth, _ http.Header, day string) Result {
	result := lobsterAICheckIn(ctx, cfg, auth, day)
	if result.Warning != "" {
		return result
	}
	if warning := lobsterAIPlatformWarning(auth); warning != "" {
		result.Warning = warning
	}
	return result
}

// lobsterAIPlatformWarning reports a host the backend cannot be told about.
//
// The activity surface only recognizes the desktop client's own process.platform
// values (win32/darwin/linux) and answers an empty slot for anything else. On any
// other host the adapter falls back to "linux", which is a guess: it is reported
// rather than hidden, and `activity_platform` overrides it.
func lobsterAIPlatformWarning(auth *coreauth.Auth) string {
	return lobsterAIPlatformWarningFor(runtime.GOOS, auth)
}

// lobsterAIPlatformWarningFor is the pure form of lobsterAIPlatformWarning.
func lobsterAIPlatformWarningFor(goos string, auth *coreauth.Auth) string {
	if lobsterAIAttribute(auth, "activity_platform") != "" {
		return ""
	}
	if lobsteraiserver.ActivityPlatformFor(goos) != "" {
		return ""
	}
	return fmt.Sprintf("activity platform is unknown for %s; sent linux instead (set activity_platform to override)", goos)
}

// lobsterAICheckIn performs the check-in protocol.
func lobsterAICheckIn(ctx context.Context, cfg *config.Config, auth *coreauth.Auth, day string) Result {
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), constant.LobsteraiServer) {
		return Result{Outcome: OutcomeBlocked, Message: "lobsterai-server: account unavailable"}
	}
	accessToken := lobsterAIAttribute(auth, "access_token")
	if accessToken == "" {
		return Result{Outcome: OutcomeBlocked, Message: "credential_unavailable"}
	}
	client := lobsteraiserver.NewClientWithProxyURL(cfg, auth.ProxyURL, lobsterAIAttribute(auth, "backend_url"))
	client.SetKeyfromIdentity("", "", lobsterAIAttribute(auth, "client_version"), "")
	// An enterprise account must present its identity on every request, and the
	// host's own GOOS is not always an environment the backend recognizes, so both
	// are taken from the credential when it carries them.
	client.SetEnterpriseContext(lobsterAIAttribute(auth, "account_mode"), lobsterAIAttribute(auth, "enterprise_id"))
	client.SetActivityPlatform(lobsterAIAttribute(auth, "activity_platform"))

	snapshot, errSnapshot := client.FetchActivitySnapshot(
		ctx,
		accessToken,
		lobsteraiserver.ActivityPlacementDesktopSidebar,
		lobsteraiserver.ActivityContainerAPIVersionDailyCheckIn,
	)
	if errSnapshot != nil {
		return Result{Outcome: OutcomeFailed, Message: errSnapshot.Error()}
	}
	if !snapshot.Available() {
		return Result{Outcome: OutcomeUnavailable, Message: "No check-in activity is open for this account", Date: day}
	}
	// completed covers a finished one-time activity, claimedToday covers the
	// daily one. Both mean there is nothing left to grant today.
	if snapshot.Completed() || snapshot.ClaimedToday() {
		return Result{
			Outcome: OutcomeAlreadyDone,
			Message: lobsterAIAlreadyDoneMessage(snapshot),
			Date:    day,
		}
	}
	actionID := lobsterAIAction(snapshot)
	if actionID == "" {
		return Result{Outcome: OutcomeUnavailable, Message: "The check-in activity exposes no action", Date: day}
	}
	claim, errClaim := client.ClaimActivity(ctx, accessToken, lobsteraiserver.ActivityClaimRequest{
		ActivityCode:   snapshot.ActivityCode,
		ActionID:       actionID,
		ConfigRevision: snapshot.ConfigRevision,
	})
	if errors.Is(errClaim, lobsteraiserver.ErrActivityAlreadyClaimed) {
		return Result{Outcome: OutcomeAlreadyDone, Message: lobsterAIAlreadyDoneMessage(snapshot), Date: day}
	}
	if errClaim != nil {
		return Result{Outcome: OutcomeFailed, Message: errClaim.Error()}
	}
	message := "Daily check-in claimed"
	if claim.GrantedKnown && claim.Granted > 0 {
		message = fmt.Sprintf("Daily check-in claimed (%g credits)", claim.Granted)
	}
	if days := snapshot.ClaimedDays(); days > 0 {
		message = fmt.Sprintf("%s; %d-day streak", message, days)
	}
	return Result{Outcome: OutcomeSuccess, Message: message, Date: day}
}

// lobsterAIAlreadyDoneMessage reports the activity's own counters so the operator
// can tell a working schedule from a misconfigured one.
func lobsterAIAlreadyDoneMessage(snapshot *lobsteraiserver.ActivitySnapshot) string {
	message := "Today's check-in is already claimed"
	if days := snapshot.ClaimedDays(); days > 0 {
		message = fmt.Sprintf("%s; %d days claimed so far", message, days)
	}
	return message
}

// lobsterAIAction prefers the daily check-in action, falling back to the only
// action the activity exposes: the one-time credit reward names it "claim".
func lobsterAIAction(snapshot *lobsteraiserver.ActivitySnapshot) string {
	actions := snapshot.Actions()
	for _, action := range actions {
		if action == lobsteraiserver.ActivityActionCheckIn {
			return action
		}
	}
	if len(actions) > 0 {
		return actions[0]
	}
	return ""
}

// lobsterAIAttribute reads a trimmed credential field from attributes, then
// metadata.
func lobsterAIAttribute(auth *coreauth.Auth, key string) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if value := strings.TrimSpace(auth.Attributes[key]); value != "" {
			return value
		}
	}
	if auth.Metadata != nil {
		if value, ok := auth.Metadata[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
