package checkin

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

type externalStubAdapter struct {
	mu   sync.Mutex
	dep  string
	fail bool
}

func (a *externalStubAdapter) Provider() string { return "dumate" }
func (a *externalStubAdapter) ExternalDependency(*coreauth.Auth) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.dep
}
func (a *externalStubAdapter) CheckIn(context.Context, *config.Config, *coreauth.Auth, http.Header, string) Result {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.fail {
		return Result{Outcome: OutcomeBlocked, Message: "credential_unavailable"}
	}
	return Result{Outcome: OutcomeSuccess, Message: "ok"}
}

func TestSchedulerResumesBlockedAccountWhenAnExternalDependencyChanges(t *testing.T) {
	adapter := &externalStubAdapter{dep: "old-session", fail: true}
	account := &coreauth.Auth{ID: "dumate-ext", Provider: "dumate", RegistrationEpoch: 1, Metadata: map[string]any{"account_id": "a"}}
	source := &testSource{accounts: map[string]*coreauth.Auth{account.ID: account}}
	clock := &testClock{now: testTime("2026-10-15T12:00:00+08:00")}
	service := newService(source, clock, func(n int64) int64 { return n / 2 }, adapter)
	service.ctx = context.Background()
	service.ApplyConfig(&config.Config{AutoCheckin: config.AutoCheckinConfig{Providers: map[string]config.CheckinProviderConfig{"dumate": {Enabled: true, Timezone: "Asia/Shanghai", Window: config.CheckinWindow{Start: "00:10", End: "03:10"}}}}})
	service.reconcile()
	service.execute(dequeue(t, service))
	if status := onlyStatus(t, service, "2026-10-15"); status.Phase != "blocked" {
		t.Fatalf("expected blocked: %+v", status)
	}
	service.reconcile()
	if len(service.jobs) != 0 {
		t.Fatal("unchanged dependency must not retry")
	}
	adapter.mu.Lock()
	adapter.dep = "new-session"
	adapter.fail = false
	adapter.mu.Unlock()
	service.reconcile()
	service.execute(dequeue(t, service))
	if status := onlyStatus(t, service, "2026-10-15"); !status.Completed {
		t.Fatalf("rotated session did not resume the blocked checkin: %+v", status)
	}
}
