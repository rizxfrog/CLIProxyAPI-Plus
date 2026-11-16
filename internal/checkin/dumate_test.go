package checkin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	dumateauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/dumate"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// dumateConsoleTransport redirects the fixed console origin to a test server and
// fails the test if the adapter ever calls another host.
type dumateConsoleTransport struct {
	target *url.URL
	base   http.RoundTripper
	t      *testing.T
}

func (tr dumateConsoleTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	wantHost := strings.TrimPrefix(dumateauth.ConsoleBaseURL, "https://")
	if req.URL.Scheme != "https" || req.URL.Host != wantHost {
		tr.t.Errorf("unexpected business host %s://%s", req.URL.Scheme, req.URL.Host)
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme, clone.URL.Host = tr.target.Scheme, tr.target.Host
	clone.Host = tr.target.Host
	return tr.base.RoundTrip(clone)
}

func dumateCheckinContext(t *testing.T, handler http.HandlerFunc) context.Context {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	target, errParse := url.Parse(server.URL)
	if errParse != nil {
		t.Fatalf("parse test server url: %v", errParse)
	}
	return context.WithValue(context.Background(), "cliproxy.roundtripper", http.RoundTripper(dumateConsoleTransport{target, server.Client().Transport, t}))
}

func dumateAccount() *coreauth.Auth {
	return &coreauth.Auth{
		ID:       "dumate-1",
		Provider: "dumate",
		Metadata: map[string]any{
			dumateauth.FieldAccountID:     "acct-1",
			dumateauth.FieldConsoleCookie: `BDUSS=test-cookie; bce-user-info="2026-11-07T10:56:10Z|csrf-sig"`,
		},
	}
}

func envelope(result string) string {
	return `{"logId":"log-1","code":0,"message":"","success":true,"result":` + result + `}`
}

func TestDumateCheckinSkipsAnAlreadyClaimedDay(t *testing.T) {
	day := "2026-11-07"
	claims := 0
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != `BDUSS=test-cookie; bce-user-info="2026-11-07T10:56:10Z|csrf-sig"` {
			t.Errorf("cookie = %q", r.Header.Get("Cookie"))
		}
		if r.Method != http.MethodGet || r.URL.Path != "/api/dumate/points/loginBonusInfo" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(envelope(`{"totalTimes":1,"totalPoints":500,"hasIssued":true,"signInDays":["` + day + `"]}`)))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, day)
	if result.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q (%s)", result.Outcome, result.Message)
	}
	if result.Date != day {
		t.Fatalf("date = %q", result.Date)
	}
	if claims != 0 {
		t.Fatalf("claims = %d, want 0", claims)
	}
}

func TestDumateCheckinClaimsWhenTheServerHasNotIssuedToday(t *testing.T) {
	day := "2026-11-08"
	var claimMethod, claimPath string
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/dumate/points/loginBonusInfo":
			_, _ = w.Write([]byte(envelope(`{"totalTimes":1,"totalPoints":500,"hasIssued":false,"signInDays":["2026-11-07"]}`)))
		case "/api/dumate/points/loginBonus":
			claimMethod, claimPath = r.Method, r.URL.Path
			_, _ = w.Write([]byte(envelope(`true`)))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, day)
	if result.Outcome != OutcomeSuccess {
		t.Fatalf("outcome = %q (%s)", result.Outcome, result.Message)
	}
	if claimMethod != http.MethodPost || claimPath != "/api/dumate/points/loginBonus" {
		t.Fatalf("claim request = %s %s", claimMethod, claimPath)
	}
}

func TestDumateCheckinTreatsADayMismatchAsAlreadyDone(t *testing.T) {
	// The scheduler's calendar day and the console's own day can disagree by a
	// timezone; the console flag is then the only proof, and claiming again must
	// not be attempted.
	claims := 0
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/dumate/points/loginBonus" {
			claims++
		}
		_, _ = w.Write([]byte(envelope(`{"hasIssued":true,"signInDays":[]}`)))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, "2026-11-08")
	if result.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	if claims != 0 {
		t.Fatalf("claims = %d, want 0", claims)
	}
}

func TestDumateCheckinReportsARefusedClaim(t *testing.T) {
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/dumate/points/loginBonusInfo" {
			_, _ = w.Write([]byte(envelope(`{"hasIssued":false,"signInDays":[]}`)))
			return
		}
		_, _ = w.Write([]byte(envelope(`false`)))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, "2026-11-08")
	// Upstream accepted the call but granted nothing: the day is already
	// accounted for, which is not a failure and must not be retried blindly.
	if result.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q", result.Outcome)
	}
}

func TestDumateCheckinReportsAnUpstreamFailure(t *testing.T) {
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"logId":"log-2","code":1001,"message":"upstream boom","success":false}`))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, "2026-11-08")
	if result.Outcome != OutcomeFailed {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	if result.Message != "dumate_checkin_failed" {
		t.Fatalf("message = %q", result.Message)
	}
	if strings.Contains(result.Message, "secret") {
		t.Fatalf("message leaked upstream detail: %q", result.Message)
	}
}

// An expired console session answers HTTP 200 with an object-shaped message, so
// the classification must survive both the envelope shape and the safe-code
// filter that used to collapse it into a generic failure.
func TestDumateCheckinBlocksAnExpiredConsoleSession(t *testing.T) {
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"message":{"global":"登录凭证已过期，请重新登录"}}`))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, "2026-11-08")
	if result.Outcome != OutcomeBlocked {
		t.Fatalf("outcome = %q (%s)", result.Outcome, result.Message)
	}
	// The scheduler re-plans an account blocked on "credential_unavailable" once
	// its credential changes, so the code is what makes a re-login recover.
	if result.Message != "credential_unavailable" {
		t.Fatalf("message = %q", result.Message)
	}
	if result.Warning != "dumate_console_session_expired" {
		t.Fatalf("warning = %q", result.Warning)
	}
	if strings.Contains(result.Warning, "登录") {
		t.Fatalf("warning leaked upstream text: %q", result.Warning)
	}
}

func TestDumateCheckinBlocksAnExpiredClaim(t *testing.T) {
	// The read endpoint can still answer from a cached session while the claim
	// rejects it; the claim is where the expiry usually shows up.
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/dumate/points/loginBonusInfo" {
			_, _ = w.Write([]byte(envelope(`{"hasIssued":false,"signInDays":[]}`)))
			return
		}
		_, _ = w.Write([]byte(`{"success":false,"message":{"global":"登录凭证已过期，请重新登录"}}`))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, "2026-11-08")
	if result.Outcome != OutcomeBlocked || result.Message != "credential_unavailable" {
		t.Fatalf("outcome = %q (%s)", result.Outcome, result.Message)
	}
	if result.Warning != "dumate_console_session_expired" {
		t.Fatalf("warning = %q", result.Warning)
	}
}

// A claim refuses with the same expiry text when another window already took
// today's bonus, so the refusal alone must not be reported as a dead session.
func TestDumateCheckinTreatsALosingClaimRaceAsAlreadyDone(t *testing.T) {
	day := "2026-11-08"
	infoReads := 0
	ctx := dumateCheckinContext(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/dumate/points/loginBonusInfo":
			infoReads++
			if infoReads == 1 {
				_, _ = w.Write([]byte(envelope(`{"hasIssued":false,"signInDays":["2026-11-07"]}`)))
				return
			}
			_, _ = w.Write([]byte(envelope(`{"hasIssued":true,"signInDays":["2026-11-07","` + day + `"]}`)))
		default:
			_, _ = w.Write([]byte(`{"success":false,"message":{"global":"登录凭证已过期，请重新登录"}}`))
		}
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, day)
	if result.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q (%s)", result.Outcome, result.Message)
	}
	if result.Warning != "" {
		t.Fatalf("warning = %q", result.Warning)
	}
}

// The console cookie normally lives in the local client's store, so a re-login
// changes no credential metadata. The dependency fingerprint is what lets the
// scheduler notice the new session and retry the same day.
func TestDumateExternalDependencyTracksTheConsoleSession(t *testing.T) {
	adapter := dumateAdapter{}
	first := adapter.ExternalDependency(dumateAccount())
	if first == "" {
		t.Fatal("dependency is empty")
	}
	if first != adapter.ExternalDependency(dumateAccount()) {
		t.Fatal("dependency is not stable for an unchanged credential")
	}
	rotated := dumateAccount()
	rotated.Metadata[dumateauth.FieldConsoleCookie] = "BDUSS=rotated-cookie"
	if adapter.ExternalDependency(rotated) == first {
		t.Fatal("a re-exported cookie must change the dependency")
	}
	if strings.Contains(first, "test-cookie") {
		t.Fatal("dependency leaked the cookie value")
	}
	if got := adapter.ExternalDependency(nil); got != "" {
		t.Fatalf("nil dependency = %q", got)
	}
}

func TestDumateCheckinRejectsForeignAccounts(t *testing.T) {
	adapter := dumateAdapter{}
	if got := adapter.Provider(); got != "dumate" {
		t.Fatalf("provider = %q", got)
	}
	if got := adapter.CheckIn(context.Background(), &config.Config{}, nil, nil, "2026-11-08"); got.Outcome != OutcomeBlocked {
		t.Fatalf("nil auth outcome = %q", got.Outcome)
	}
	foreign := &coreauth.Auth{ID: "x", Provider: "trae", Metadata: map[string]any{}}
	if got := adapter.CheckIn(context.Background(), &config.Config{}, foreign, nil, "2026-11-08"); got.Outcome != OutcomeBlocked {
		t.Fatalf("foreign provider outcome = %q", got.Outcome)
	}
}

func TestDumateIdentityUsesTheAccountID(t *testing.T) {
	adapter := dumateAdapter{}
	if got := adapter.Identity(dumateAccount()); got != "acct-1" {
		t.Fatalf("identity = %q", got)
	}
	if got := adapter.Identity(nil); got != "" {
		t.Fatalf("nil identity = %q", got)
	}
}

func TestDumateIsAdvertisedAsACapability(t *testing.T) {
	found := false
	for _, capability := range Capabilities() {
		if capability.ID != "dumate" {
			continue
		}
		found = true
		if capability.Name == "" {
			t.Fatal("capability has no display name")
		}
		// The adapter builds its own Cookie header, so the operator must not be
		// asked to template one.
		for _, protected := range capability.ProtectedHeaders {
			if strings.EqualFold(protected, "Cookie") {
				t.Fatal("Cookie must stay configurable, not protected")
			}
		}
	}
	if !found {
		t.Fatal("dumate is missing from Capabilities()")
	}
}
