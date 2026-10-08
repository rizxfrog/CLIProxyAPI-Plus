package floatboat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// cataloguePayload mirrors the gateway's real /api/pricing shape: a key-scoped
// auto_groups list plus a per-model cross-account enable_groups union.
const cataloguePayload = `{
  "auto_groups": ["default"],
  "group_ratio": {"default": 0.00075, "vip": 0.00075},
  "success": true,
  "data": [
    {"model_name":"deepseek-flash","model_ratio":0.4,"enable_groups":["default","vip"],"supported_endpoint_types":["anthropic","openai","gemini"]},
    {"model_name":"gemini-3.8-flash","model_ratio":0.375,"enable_groups":["default"],"supported_endpoint_types":["gemini","openai"]},
    {"model_name":"glm-5.2","model_ratio":0.725,"enable_groups":["default","vip"],"supported_endpoint_types":["anthropic","openai"]},
    {"model_name":"grok-latest","model_ratio":3,"enable_groups":["vip"],"supported_endpoint_types":["openai"]},
    {"model_name":"claude-opus-4-6","model_ratio":5,"enable_groups":["default"],"supported_endpoint_types":["anthropic"]},
    {"model_name":"jev-latest","model_ratio":0.021,"enable_groups":["default"],"supported_endpoint_types":["typesafe"]},
    {"model_name":"chirp-3-hd","model_ratio":15,"enable_groups":["default"],"supported_endpoint_types":["audio-speech"]},
    {"model_name":"deepseek-flash","model_ratio":0.4,"enable_groups":["default"],"supported_endpoint_types":["openai"]}
  ]
}`

// TestParseCatalogueFiltersEntitlementAndCapability pins the two independent
// filters: a model is routable only when one of the key's own groups intersects
// the model's published groups AND the model advertises a chat endpoint family.
// Regression guard for the observed gateway behaviour where enable_groups lists
// vip-only models that the base key is rejected from at request time.
func TestParseCatalogueFiltersEntitlementAndCapability(t *testing.T) {
	catalogue, err := parseCatalogue([]byte(cataloguePayload))
	if err != nil {
		t.Fatalf("parseCatalogue: %v", err)
	}
	if len(catalogue.Groups) != 1 || catalogue.Groups[0] != "default" {
		t.Fatalf("groups = %v, want [default]", catalogue.Groups)
	}
	got := map[string]CatalogueEntry{}
	for _, entry := range catalogue.Entries {
		got[entry.ID] = entry
	}
	// Entitled + chat-capable.
	for _, want := range []string{"deepseek-flash", "gemini-3.8-flash", "glm-5.2", "claude-opus-4-6"} {
		if _, ok := got[want]; !ok {
			t.Errorf("catalogue missing routable model %q", want)
		}
	}
	// VIP-only: advertised for a group the key does not own.
	if _, ok := got["grok-latest"]; ok {
		t.Error("catalogue must not include a model outside the key's auto_groups")
	}
	// Entitled but not a chat family this provider can drive.
	for _, bad := range []string{"jev-latest", "chirp-3-hd"} {
		if _, ok := got[bad]; ok {
			t.Errorf("catalogue must exclude non-chat model %q", bad)
		}
	}
	// Duplicate model_name entries collapse to one.
	count := 0
	for _, entry := range catalogue.Entries {
		if entry.ID == "deepseek-flash" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("deepseek-flash appeared %d times, want 1", count)
	}
}

// TestParseCatalogueKeepsAllEntriesWithoutAutoGroups covers a gateway that does
// not report the key's groups: entitlement cannot be decided, so capability
// alone filters the list rather than dropping every model.
func TestParseCatalogueKeepsAllEntriesWithoutAutoGroups(t *testing.T) {
	catalogue, err := parseCatalogue([]byte(`{"success":true,"data":[
		{"model_name":"deepseek-flash","enable_groups":["vip"],"supported_endpoint_types":["openai"]},
		{"model_name":"jev-latest","enable_groups":["vip"],"supported_endpoint_types":["typesafe"]}
	]}`))
	if err != nil {
		t.Fatalf("parseCatalogue: %v", err)
	}
	if len(catalogue.Groups) != 0 {
		t.Fatalf("groups = %v, want none", catalogue.Groups)
	}
	if len(catalogue.Entries) != 1 || catalogue.Entries[0].ID != "deepseek-flash" {
		t.Fatalf("entries = %+v, want only the chat-capable model", catalogue.Entries)
	}
}

// TestParseCatalogueRejectsMalformed checks a response without the data array
// fails loudly instead of silently registering nothing.
func TestParseCatalogueRejectsMalformed(t *testing.T) {
	if _, err := parseCatalogue([]byte(`{"success":true}`)); err == nil {
		t.Fatal("expected an error for a payload without a data array")
	}
	if _, err := parseCatalogue([]byte(`{"code":40101,"message":"invalid key"}`)); err == nil {
		t.Fatal("expected a non-zero envelope code to surface")
	}
}

// TestFetchCatalogueUsesInferenceKey hits the gateway with the minted api_key
// and the pricing path, and confirms the parse is wired end to end.
func TestFetchCatalogueUsesInferenceKey(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		_, _ = w.Write([]byte(cataloguePayload))
	}))
	defer server.Close()

	client := NewClientWithProxyURL(nil, "", server.URL)
	catalogue, err := client.FetchCatalogue(context.Background(), server.URL, "sk-test-key")
	if err != nil {
		t.Fatalf("FetchCatalogue: %v", err)
	}
	if gotPath != DefaultPricingPath {
		t.Fatalf("path = %q, want %q", gotPath, DefaultPricingPath)
	}
	if gotAuth != "Bearer sk-test-key" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if len(catalogue.Entries) == 0 {
		t.Fatal("no entries parsed")
	}
	if _, err := client.FetchCatalogue(context.Background(), server.URL, "  "); err == nil {
		t.Fatal("expected an error when the api key is empty")
	}
}

// TestFetchBillingComputesRemaining pins the measured allowance: remaining is
// derived only when both gateway reads succeed, and an unknown value is never
// reported as zero or unlimited.
func TestFetchBillingComputesRemaining(t *testing.T) {
	usageOK := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case DefaultBillingSubscriptionPath:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "billing_subscription", "has_payment_method": true,
				"soft_limit_usd": 0.8, "hard_limit_usd": 0.8, "system_hard_limit_usd": 0.8,
			})
		case DefaultBillingUsagePath:
			if !usageOK {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "total_usage": 0.4})
		}
	}))
	defer server.Close()

	client := NewClientWithProxyURL(nil, "", server.URL)
	billing, err := client.FetchBilling(context.Background(), server.URL, "sk-test-key")
	if err != nil {
		t.Fatalf("FetchBilling: %v", err)
	}
	if !billing.HasPaymentMethod || billing.HardLimitUSD != 0.8 || billing.TotalUsageUSD != 0.4 {
		t.Fatalf("billing = %+v", billing)
	}
	remaining, ok := billing.Remaining()
	if !ok || remaining < 0.399 || remaining > 0.401 {
		t.Fatalf("remaining = %v (ok=%v), want ~0.4", remaining, ok)
	}

	// A failed usage read must not be reported as a zero balance.
	usageOK = false
	partial, errPartial := client.FetchBilling(context.Background(), server.URL, "sk-test-key")
	if errPartial != nil {
		t.Fatalf("FetchBilling (partial): %v", errPartial)
	}
	if _, ok := partial.Remaining(); ok {
		t.Fatal("remaining must be unknown when the usage read failed")
	}
	if !partial.SubscriptionOK || partial.UsageOK {
		t.Fatalf("partial ok flags = sub:%v usage:%v", partial.SubscriptionOK, partial.UsageOK)
	}
}
