package codearts

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// TestLedgerNumberAcceptsEveryObservedShape pins that an amount survives the
// JSON type the route happens to use.
//
// The same metric arrives as an integer for a credit-less account and as a float
// once the account holds credits, and a strict int64 field rejected the float
// with "cannot unmarshal number 100.0", failing the whole balance read.
func TestLedgerNumberAcceptsEveryObservedShape(t *testing.T) {
	cases := []struct {
		raw  string
		want int64
	}{
		{`0`, 0},
		{`100`, 100},
		{`100.0`, 100},
		{`100.75`, 100},
		{`1e7`, 10_000_000},
		{`1.0e7`, 10_000_000},
		{`"100"`, 100},
		{`"100.0"`, 100},
		{`" 42 "`, 42},
		{`null`, 0},
		{`-5`, -5},
	}
	for _, tc := range cases {
		var got ledgerNumber
		if err := json.Unmarshal([]byte(tc.raw), &got); err != nil {
			t.Errorf("unmarshal %q: %v", tc.raw, err)
			continue
		}
		if got.Int64() != tc.want {
			t.Errorf("unmarshal %q = %d, want %d", tc.raw, got.Int64(), tc.want)
		}
	}
}

// TestLedgerNumberRejectsNonNumericText pins that garbage is an error rather
// than a silent zero, so a broken payload cannot masquerade as "no credits".
func TestLedgerNumberRejectsNonNumericText(t *testing.T) {
	for _, raw := range []string{`"abc"`, `"12abc"`, `true`, `{}`, `[]`} {
		var got ledgerNumber
		if err := json.Unmarshal([]byte(raw), &got); err == nil {
			t.Errorf("unmarshal %q unexpectedly succeeded with %d", raw, got.Int64())
		}
	}
}

// TestFetchPluginStatisticsReadsAFloatLedger reproduces the live failure: an
// account holding credits reports the amounts as floats.
func TestFetchPluginStatisticsReadsAFloatLedger(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PluginStatisticsPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"metrics":[
				{"name":"usageTokenChatMessages","usage_token_num":1234.0,"package_token_amount":10000000.0},
				{"name":"usageTotalPackageCredit","package_credit_amount":100.0,"package_credit_used":0.0,
				 "package_credit_remain":100.0,"package_credit_expiring_amount":100.0}
			],
			"package":{"spec_code":"codearts.agent.individual.basic","status":"normal",
			           "package_name_en":"Basic","is_credit_package":true},
			"start_date":"2026-11-01","end_date":"2026-12-01"
		}`))
	}))
	defer server.Close()

	balance, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance != nil {
		t.Fatalf("balance: %v", errBalance)
	}
	if balance.TotalQuota != 100 || balance.TotalBalance != 100 || balance.ExpiringAmount != 100 {
		t.Fatalf("credit figures lost: %+v", balance)
	}
	if balance.UsedAmount != 0 {
		t.Fatalf("used = %d, want 0", balance.UsedAmount)
	}
	if balance.DailyTokenLimit != 10_000_000 || balance.DailyTokensUsed != 1234 {
		t.Fatalf("token figures lost: %+v", balance)
	}
	if !balance.CreditsKnown {
		t.Fatal("a successfully read ledger must report credits_known")
	}
	if balance.Package != "Basic" || balance.SpecCode != "codearts.agent.individual.basic" {
		t.Fatalf("package lost: %+v", balance)
	}
}

// TestFetchPluginStatisticsAcceptsMixedNumericTypes pins that one field's type
// does not depend on another's: real payloads mix integers and floats.
func TestFetchPluginStatisticsAcceptsMixedNumericTypes(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metrics":[{"name":"usageTotalPackageCredit",
			"package_credit_amount":0,"package_credit_remain":"100","package_credit_used":0.5}]}`))
	}))
	defer server.Close()

	balance, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance != nil {
		t.Fatalf("balance: %v", errBalance)
	}
	if balance.TotalBalance != 100 {
		t.Fatalf("string amount lost: %+v", balance)
	}
	if balance.TotalQuota != 0 {
		t.Fatalf("quota = %d, want 0", balance.TotalQuota)
	}
}
