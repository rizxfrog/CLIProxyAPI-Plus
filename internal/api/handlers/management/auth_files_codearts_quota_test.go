package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// fakeCodeArtsBalance is a deterministic stand-in for the balance client.
type fakeCodeArtsBalance struct {
	balance   *codeartsauth.TokenBalance
	err       error
	gotAK     string
	gotSK     string
	gotST     string
	callCount int
}

func (f *fakeCodeArtsBalance) FetchTokenBalance(_ context.Context, ak, sk, st string) (*codeartsauth.TokenBalance, error) {
	f.callCount++
	f.gotAK, f.gotSK, f.gotST = ak, sk, st
	if f.err != nil {
		return nil, f.err
	}
	return f.balance, nil
}

func withFakeCodeArtsBalance(t *testing.T, fake *fakeCodeArtsBalance) {
	t.Helper()
	previous := newCodeArtsBalanceClient
	newCodeArtsBalanceClient = func(_ *config.Config, _ string) codeArtsBalanceFetcher { return fake }
	t.Cleanup(func() { newCodeArtsBalanceClient = previous })
}

// registerCodeArtsAuth builds a manager carrying one CodeArts credential whose
// security token is still valid (expires one hour out), so the quota handler
// never tries to refresh against STS.
func registerCodeArtsAuth(t *testing.T) (*coreauth.Manager, *coreauth.Auth) {
	t.Helper()
	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{
		ID:       "codearts:oauth:test",
		Provider: "codearts",
		Metadata: map[string]any{
			"access_key":     "AK_TEST",
			"secret_key":     "SK_TEST",
			"security_token": "ST_TEST",
			"expired":        time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		},
	}
	if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
		t.Fatalf("register codearts auth: %v", errRegister)
	}
	return manager, auth
}

func TestGetCodeArtsQuotaReturnsBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeCodeArtsBalance{balance: &codeartsauth.TokenBalance{
		Channel:         "codearts",
		TotalQuota:      10000000,
		TotalBalance:    9998929,
		UsedAmount:      1071,
		DailyTokenLimit: 10000000,
		DailyTokensUsed: 1071,
	}}
	withFakeCodeArtsBalance(t, fake)

	manager, auth := registerCodeArtsAuth(t)
	h := &Handler{authManager: manager}
	authIndex := auth.EnsureIndex()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/codearts-quota?auth_index="+authIndex, nil)

	h.GetCodeArtsQuota(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if fake.gotAK != "AK_TEST" || fake.gotSK != "SK_TEST" || fake.gotST != "ST_TEST" {
		t.Fatalf("balance fetch got credentials (%q,%q,%q)", fake.gotAK, fake.gotSK, fake.gotST)
	}
	var body CodeArtsQuota
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.TotalQuota != 10000000 || body.TotalBalance != 9998929 || body.UsedAmount != 1071 {
		t.Fatalf("unexpected quota body: %+v", body)
	}
}

func TestGetCodeArtsQuotaRequiresAuthIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{authManager: coreauth.NewManager(nil, nil, nil)}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/codearts-quota", nil)

	h.GetCodeArtsQuota(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}

func TestGetCodeArtsQuotaRejectsUnknownAuthIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{authManager: coreauth.NewManager(nil, nil, nil)}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/codearts-quota?auth_index=nope", nil)

	h.GetCodeArtsQuota(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestCodeArtsCredentialExpired(t *testing.T) {
	future := &coreauth.Auth{Metadata: map[string]any{"expired": time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}}
	if codeartsCredentialExpired(future) {
		t.Fatal("future expiry should not be treated as expired")
	}
	past := &coreauth.Auth{Metadata: map[string]any{"expired": time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)}}
	if !codeartsCredentialExpired(past) {
		t.Fatal("past expiry should be treated as expired")
	}
	empty := &coreauth.Auth{Metadata: map[string]any{}}
	if codeartsCredentialExpired(empty) {
		t.Fatal("missing expiry should defer to the upstream, not force a refresh")
	}
}
