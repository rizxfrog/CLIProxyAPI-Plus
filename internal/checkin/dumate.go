package checkin

import (
	"context"
	"errors"
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

// ExternalDependency fingerprints the console session the check-in depends on.
//
// The cookie usually comes from the local client's own store, not from the
// credential file, so a re-login in that client changes no metadata and would
// leave a blocked account blocked until the next day. Folding a fingerprint of
// the resolved cookie into the repair hint lets the scheduler notice the new
// session and retry within the same window. The fingerprint never leaves the
// scheduler, and a cookie is already a secret the credential carries.
func (dumateAdapter) ExternalDependency(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	cookie, err := dumateauth.ResolveConsoleCookie(auth.Metadata)
	if err != nil || cookie == "" {
		return "unavailable"
	}
	return Fingerprint(cookie)
}

func (dumateAdapter) CheckIn(ctx context.Context, cfg *config.Config, auth *coreauth.Auth, _ http.Header, day string) Result {
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), "dumate") {
		return Result{Outcome: OutcomeBlocked, Message: "dumate: account unavailable"}
	}
	cookie, errCookie := dumateauth.ResolveConsoleCookie(auth.Metadata)
	if errCookie != nil {
		return Result{Outcome: OutcomeBlocked, Message: "credential_unavailable", Warning: "dumate_console_cookie_unavailable"}
	}

	client := dumateauth.NewConsoleClient(helps.NewProxyAwareHTTPClient(ctx, cfg, auth, 0))
	info, err := client.FetchLoginBonusInfo(ctx, cookie)
	if err != nil {
		return dumateFailure(err)
	}
	// HasIssued is the console's own "already claimed today" flag; the day list
	// covers a timezone disagreement between the scheduler and the console.
	if info.HasIssued || info.SignedInOn(day) {
		return Result{Outcome: OutcomeAlreadyDone, Message: "Today's check-in is already claimed", Date: day}
	}

	claimed, err := client.ClaimLoginBonus(ctx, cookie)
	if err != nil {
		return dumateClaimFailure(ctx, client, cookie, day, err)
	}
	if !claimed {
		// The server answered success but did not grant, which means the day was
		// already accounted for upstream.
		return Result{Outcome: OutcomeAlreadyDone, Message: "Today's check-in is already claimed", Date: day}
	}
	return Result{Outcome: OutcomeSuccess, Message: "Daily check-in claimed", Date: day}
}

// dumateClaimFailure decides whether a rejected claim means the session is dead.
//
// A claim refuses with the same "login expired" text when today's bonus was
// already taken in another window, so the refusal alone cannot distinguish an
// expired session from a lost race. Re-reading the sign-in state settles it: a
// day that is now marked claimed means the reward exists and needs no retry,
// while an unchanged state leaves re-login as the only way forward.
func dumateClaimFailure(ctx context.Context, client *dumateauth.ConsoleClient, cookie, day string, claimErr error) Result {
	var consoleErr *dumateauth.ConsoleError
	if !errors.As(claimErr, &consoleErr) || !consoleErr.SessionExpired() {
		return dumateFailure(claimErr)
	}
	if ctx.Err() == nil {
		if info, err := client.FetchLoginBonusInfo(ctx, cookie); err == nil && (info.HasIssued || info.SignedInOn(day)) {
			return Result{Outcome: OutcomeAlreadyDone, Message: "Today's check-in is already claimed", Date: day}
		}
	}
	return dumateFailure(claimErr)
}

// dumateFailure maps a console error onto an outcome without publishing the
// upstream message, which carries account and session detail.
//
// A rejected session is blocked rather than failed: no retry can succeed until
// the operator signs in again, and the scheduler re-plans a blocked account as
// soon as its credential changes, so a refreshed console cookie is picked up the
// same day instead of waiting for tomorrow's window.
func dumateFailure(err error) Result {
	var consoleErr *dumateauth.ConsoleError
	if errors.As(err, &consoleErr) && consoleErr.SessionExpired() {
		return Result{Outcome: OutcomeBlocked, Message: "credential_unavailable", Warning: "dumate_console_session_expired"}
	}
	return Result{Outcome: OutcomeFailed, Message: "dumate_checkin_failed"}
}
