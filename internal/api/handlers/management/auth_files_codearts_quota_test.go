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

// TestGetCodeArtsQuotaReadsAFloatLedger drives the real upstream client against a
// server that reports credit amounts as floats.
//
// This is the shape a funded account receives: the same metric arrived as an
// integer while the package held no credits, and a strict int64 field rejected
// the float with "cannot unmarshal number 100.0", failing the entire balance
// read for every account that actually has credits.
func TestGetCodeArtsQuotaReadsAFloatLedger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != codeartsauth.PluginStatisticsPath {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{
			"metrics":[
				{"name":"usageTotalPackageCredit","package_credit_amount":5520.0,"package_credit_used":120.0,
				 "package_credit_remain":5400.0,"package_credit_expiring_amount":5400.0},
				{"name":"usageTokenChatMessages","usage_token_num":1000.0,"package_token_amount":10000000.0}
			],
			"package":{"spec_code":"codearts.agent.individual.basic","package_name_en":"Basic","is_credit_package":true}
		}`))
	}))
	defer upstream.Close()

	previous := newCodeArtsBalanceClient
	newCodeArtsBalanceClient = func(cfg *config.Config, proxyURL string) codeArtsBalanceFetcher {
		return codeartsauth.NewClientWithProxyURL(cfg, proxyURL).WithSnapHost(upstream.URL)
	}
	t.Cleanup(func() { newCodeArtsBalanceClient = previous })

	manager, auth := registerCodeArtsAuth(t)
	h := &Handler{authManager: manager}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v8/management/codearts-quota?auth_index="+auth.EnsureIndex(), nil)

	h.GetCodeArtsQuota(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var body CodeArtsQuota
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.TotalQuota != 5520 || body.TotalBalance != 5400 || body.UsedAmount != 120 {
		t.Fatalf("credit figures lost: %+v", body)
	}
	if body.ExpiringAmount != 5400 {
		t.Fatalf("expiring amount = %d, want 5400", body.ExpiringAmount)
	}
	if !body.CreditsKnown {
		t.Fatal("credits_known not set for a successful ledger read")
	}
	if body.Package != "Basic" || body.SpecCode != "codearts.agent.individual.basic" {
		t.Fatalf("package metadata lost: %+v", body)
	}
}
