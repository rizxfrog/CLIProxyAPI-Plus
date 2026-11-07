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
			dumateauth.FieldConsoleCookie: "BDUSS=test-cookie",
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
		if r.Header.Get("Cookie") != "BDUSS=test-cookie" {
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
		_, _ = w.Write([]byte(`{"logId":"log-2","code":1001,"message":"登录凭证已过期","success":false}`))
	})

	result := dumateAdapter{}.CheckIn(ctx, &config.Config{}, dumateAccount(), nil, "2026-11-08")
	if result.Outcome != OutcomeFailed {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	if strings.Contains(result.Message, "secret") {
		t.Fatalf("message leaked upstream detail: %q", result.Message)
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
