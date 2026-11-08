package checkin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// lobsterAIBackend serves the activity surface the adapter drives. It is keyed
// on the credential's backend_url attribute, so the adapter reaches the test
// server without a transport override.
type lobsterAIBackend struct {
	slot    map[string]any
	context map[string]any
	claim   map[string]any
	// claims counts POSTs so a test can prove an already-claimed day was not
	// re-posted.
	claims atomic.Int32
	// sawClaimBody captures the action payload for assertions.
	sawClaimBody map[string]any
}

func newLobsterAIBackend(t *testing.T, backend *lobsterAIBackend) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
			t.Errorf("missing bearer authorization: %q", got)
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/client-activities/slot":
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "success", "data": backend.slot})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/context"):
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "success", "data": backend.context})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/actions/"):
			backend.claims.Add(1)
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			backend.sawClaimBody = body
			_ = json.NewEncoder(w).Encode(backend.claim)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func lobsterAIAccount(backendURL string) *coreauth.Auth {
	return &coreauth.Auth{
		ID:       "lobsterai-1",
		Provider: "lobsterai-server",
		Metadata: map[string]any{"user_id": "123318"},
		Attributes: map[string]string{
			"access_token": "acc-1",
			"backend_url":  backendURL,
		},
	}
}

func lobsterAIOpenSlot() map[string]any {
	return map[string]any{
		"slotState": "available",
		"activity": map[string]any{
			"activityCode":   "daily-check-in-evergreen-prod-20260814",
			"configRevision": float64(7),
		},
	}
}

func lobsterAIClaimedContext() map[string]any {
	return map[string]any{
		"state":   map[string]any{"claimedToday": true, "claimedDays": float64(3)},
		"actions": []any{"check_in"},
	}
}

func lobsterAIClaimableContext() map[string]any {
	return map[string]any{
		"state":   map[string]any{"claimedToday": false, "completed": false, "rewardCredits": float64(100)},
		"actions": []any{"check_in"},
	}
}

func TestLobsterAICheckinSkipsAnAlreadyClaimedDay(t *testing.T) {
	backend := &lobsterAIBackend{
		slot:    lobsterAIOpenSlot(),
		context: lobsterAIClaimedContext(),
		claim:   map[string]any{"code": 0, "data": map[string]any{"result": map[string]any{"creditsGranted": 100.0}}},
	}
	server := newLobsterAIBackend(t, backend)
	got := lobsterAIAdapter{}.CheckIn(context.Background(), &config.Config{}, lobsterAIAccount(server.URL), nil, "2026-10-15")
	if got.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q (%s)", got.Outcome, got.Message)
	}
	if backend.claims.Load() != 0 {
		t.Fatalf("a claimed day must not be re-posted; claims = %d", backend.claims.Load())
	}
}

func TestLobsterAICheckinClaimsWhenTheDayIsOpen(t *testing.T) {
	backend := &lobsterAIBackend{
		slot:    lobsterAIOpenSlot(),
		context: lobsterAIClaimableContext(),
		claim:   map[string]any{"code": 0, "data": map[string]any{"result": map[string]any{"creditsGranted": 100.0, "periodKey": "2026-10-15"}}},
	}
	server := newLobsterAIBackend(t, backend)
	got := lobsterAIAdapter{}.CheckIn(context.Background(), &config.Config{}, lobsterAIAccount(server.URL), nil, "2026-10-15")
	if got.Outcome != OutcomeSuccess {
		t.Fatalf("outcome = %q (%s)", got.Outcome, got.Message)
	}
	if backend.claims.Load() != 1 {
		t.Fatalf("claims = %d", backend.claims.Load())
	}
	if backend.sawClaimBody["configRevision"] != float64(7) {
		t.Fatalf("configRevision = %v", backend.sawClaimBody["configRevision"])
	}
	if _, ok := backend.sawClaimBody["idempotencyKey"].(string); !ok {
		t.Fatalf("idempotencyKey missing: %v", backend.sawClaimBody)
	}
}

func TestLobsterAICheckinReportsADuplicateClaimAsAlreadyDone(t *testing.T) {
	backend := &lobsterAIBackend{
		slot:    lobsterAIOpenSlot(),
		context: lobsterAIClaimableContext(),
		claim:   map[string]any{"code": lobsteraiserver.ActivityAlreadyClaimedCode, "message": "已领取"},
	}
	server := newLobsterAIBackend(t, backend)
	got := lobsterAIAdapter{}.CheckIn(context.Background(), &config.Config{}, lobsterAIAccount(server.URL), nil, "2026-10-15")
	if got.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q (%s)", got.Outcome, got.Message)
	}
}

func TestLobsterAICheckinReportsACompletedPeriodAsAlreadyDone(t *testing.T) {
	backend := &lobsterAIBackend{
		slot:    lobsterAIOpenSlot(),
		context: map[string]any{"state": map[string]any{"completed": true}, "actions": []any{"check_in"}},
		claim:   map[string]any{"code": 0, "data": map[string]any{}},
	}
	server := newLobsterAIBackend(t, backend)
	got := lobsterAIAdapter{}.CheckIn(context.Background(), &config.Config{}, lobsterAIAccount(server.URL), nil, "2026-10-15")
	if got.Outcome != OutcomeAlreadyDone {
		t.Fatalf("outcome = %q (%s)", got.Outcome, got.Message)
	}
	if backend.claims.Load() != 0 {
		t.Fatalf("a completed period must not be re-posted; claims = %d", backend.claims.Load())
	}
}

func TestLobsterAICheckinReportsAnEmptySlotAsUnavailable(t *testing.T) {
	backend := &lobsterAIBackend{
		slot:    map[string]any{"slotState": "empty"},
		context: map[string]any{},
		claim:   map[string]any{"code": 0},
	}
	server := newLobsterAIBackend(t, backend)
	got := lobsterAIAdapter{}.CheckIn(context.Background(), &config.Config{}, lobsterAIAccount(server.URL), nil, "2026-10-15")
	if got.Outcome != OutcomeUnavailable {
		t.Fatalf("outcome = %q (%s)", got.Outcome, got.Message)
	}
	if backend.claims.Load() != 0 {
		t.Fatalf("claims = %d", backend.claims.Load())
	}
}

func TestLobsterAICheckinReportsAnUpstreamBusinessError(t *testing.T) {
	backend := &lobsterAIBackend{
		slot:    lobsterAIOpenSlot(),
		context: lobsterAIClaimableContext(),
		claim:   map[string]any{"code": 51100, "message": "活动未开始"},
	}
	server := newLobsterAIBackend(t, backend)
	got := lobsterAIAdapter{}.CheckIn(context.Background(), &config.Config{}, lobsterAIAccount(server.URL), nil, "2026-10-15")
	if got.Outcome != OutcomeFailed {
		t.Fatalf("outcome = %q (%s)", got.Outcome, got.Message)
	}
	if !strings.Contains(got.Message, "51100") {
		t.Fatalf("message must carry the backend code: %q", got.Message)
	}
}

func TestLobsterAICheckinBlocksOnMissingCredentials(t *testing.T) {
	adapter := lobsterAIAdapter{}
	if got := adapter.CheckIn(context.Background(), &config.Config{}, nil, nil, "2026-10-15"); got.Outcome != OutcomeBlocked {
		t.Fatalf("nil auth outcome = %q", got.Outcome)
	}
	noToken := &coreauth.Auth{ID: "x", Provider: "lobsterai-server", Metadata: map[string]any{}}
	if got := adapter.CheckIn(context.Background(), &config.Config{}, noToken, nil, "2026-10-15"); got.Outcome != OutcomeBlocked {
		t.Fatalf("tokenless outcome = %q", got.Outcome)
	}
	foreign := &coreauth.Auth{ID: "y", Provider: "trae", Metadata: map[string]any{}}
	if got := adapter.CheckIn(context.Background(), &config.Config{}, foreign, nil, "2026-10-15"); got.Outcome != OutcomeBlocked {
		t.Fatalf("foreign provider outcome = %q", got.Outcome)
	}
}

func TestLobsterAIIdentityUsesTheAccountID(t *testing.T) {
	adapter := lobsterAIAdapter{}
	if got := adapter.Identity(lobsterAIAccount("http://127.0.0.1:1")); got != "123318" {
		t.Fatalf("identity = %q", got)
	}
	if got := adapter.Identity(nil); got != "" {
		t.Fatalf("nil identity = %q", got)
	}
}

func TestLobsterAIIsAdvertisedAsACapability(t *testing.T) {
	found := false
	for _, capability := range Capabilities() {
		if capability.ID != "lobsterai-server" {
			continue
		}
		found = true
		if capability.Name == "" {
			t.Fatal("capability has no display name")
		}
	}
	if !found {
		t.Fatal("lobsterai-server is missing from Capabilities()")
	}
}
