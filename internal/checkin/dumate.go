package checkin

import (
	"context"
	"net/http"
	"strings"

	dumateauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/dumate"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// dumateAdapter claims the Dumate daily sign-in bonus.
//
// The reward lives on the Baidu Cloud console and is authorized by the console
// cookie stored with the credential. The upstream claim has no idempotency key,
// so the adapter reads the sign-in state first and only claims when the server
// does not already count today; a claim that still reports "not claimed" is
// reported as already done rather than as a failure.
type dumateAdapter struct{}

// NewDumateAdapter builds the Dumate daily sign-in adapter.
func NewDumateAdapter() Adapter { return dumateAdapter{} }

func (dumateAdapter) Provider() string { return "dumate" }

// Identity keys the daily observation on the account id, so re-exporting the
// same account with a rotated in-app key keeps one schedule entry.
func (dumateAdapter) Identity(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	return dumateauth.MetadataString(auth.Metadata, dumateauth.FieldAccountID)
}

// RequiredMetadata lists the credential fields addressable from a check-in
// request header template. Dumate builds its own Cookie header, so this only
// exposes the fields an operator could reference deliberately.
func (dumateAdapter) RequiredMetadata() []string {
	return []string{"account_id", "console_cookie"}
}

func (dumateAdapter) CheckIn(ctx context.Context, cfg *config.Config, auth *coreauth.Auth, _ http.Header, day string) Result {
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), "dumate") {
		return Result{Outcome: OutcomeBlocked, Message: "dumate: account unavailable"}
	}
	cookie, errCookie := dumateauth.ResolveConsoleCookie(auth.Metadata)
	if errCookie != nil {
		return Result{Outcome: OutcomeBlocked, Message: "credential_unavailable", Warning: errCookie.Error()}
	}

	client := dumateauth.NewConsoleClient(helps.NewProxyAwareHTTPClient(ctx, cfg, auth, 0))
	info, err := client.FetchLoginBonusInfo(ctx, cookie)
	if err != nil {
		return Result{Outcome: OutcomeFailed, Message: err.Error()}
	}
	// HasIssued is the console's own "already claimed today" flag; the day list
	// covers a timezone disagreement between the scheduler and the console.
	if info.HasIssued || info.SignedInOn(day) {
		return Result{Outcome: OutcomeAlreadyDone, Message: "Today's check-in is already claimed", Date: day}
	}

	claimed, err := client.ClaimLoginBonus(ctx, cookie)
	if err != nil {
		return Result{Outcome: OutcomeFailed, Message: err.Error()}
	}
	if !claimed {
		// The server answered success but did not grant, which means the day was
		// already accounted for upstream.
		return Result{Outcome: OutcomeAlreadyDone, Message: "Today's check-in is already claimed", Date: day}
	}
	return Result{Outcome: OutcomeSuccess, Message: "Daily check-in claimed", Date: day}
}
