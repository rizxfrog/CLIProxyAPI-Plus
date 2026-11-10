package codearts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient builds a client pointed at a test server.
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := NewClient(nil)
	client.stsHost = server.URL
	client.portalHost = server.URL
	client.snapHost = server.URL
	client.benefitURL = server.URL
	return client, server
}

// signedHeaderPresent asserts the request carried a Huawei Cloud signature.
func signedHeaderPresent(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Authorization"), Algorithm+" Access=")
}

// TestBuildAuthorizationURL pins the portal authorize parameters.
func TestBuildAuthorizationURL(t *testing.T) {
	client := NewClient(nil)
	url, errBuild := client.BuildAuthorizationURL(strings.Repeat("ab", 64), "10000", "ticket123")
	if errBuild != nil {
		t.Fatalf("build url: %v", errBuild)
	}
	for _, fragment := range []string{
		"/portal/authorize?",
		"client_id=codearts",
		"uri_scheme=codearts",
		"port=10000",
		"code_challenge_method=SHA-256",
		"ticket_id=ticket123",
		"plugin-name=snap_AIIDE",
		"plugin-version=5.1.0",
		"theme=2",
		"locale=zh-cn",
	} {
		if !strings.Contains(url, fragment) {
			t.Errorf("authorize url missing %q:\n%s", fragment, url)
		}
	}
}

// TestExchangeCodeParsesCredentials pins the STS response mapping.
func TestExchangeCodeParsesCredentials(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != TokenPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !signedHeaderPresent(r) && r.Header.Get("DPoP") == "" {
			// the exchange is DPoP-authenticated, not HMAC signed
			t.Errorf("exchange request is missing the DPoP proof")
		}
		if r.Header.Get("DPoP") == "" {
			t.Errorf("DPoP header is required")
		}
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %q", r.Form.Get("grant_type"))
		}
		if r.Form.Get("redirect_uri") != CallbackURL("10000") {
			t.Errorf("redirect_uri = %q", r.Form.Get("redirect_uri"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"refresh_token": "rt-1",
			"credentials": {
				"access_key_id": "AK123",
				"secret_access_key": "SK456",
				"security_token": "ST789",
				"expiration": "2030-01-02T03:04:05Z"
			}
		}`))
	}))
	defer server.Close()

	keyPair, errKey := GenerateDpopKeyPair()
	if errKey != nil {
		t.Fatalf("generate key pair: %v", errKey)
	}
	token, errExchange := client.ExchangeCode(context.Background(), "the-code", "verifier", keyPair, "10000")
	if errExchange != nil {
		t.Fatalf("exchange: %v", errExchange)
	}
	if token.AccessKey != "AK123" || token.SecretKey != "SK456" || token.SecurityToken != "ST789" {
		t.Fatalf("unexpected credentials: %+v", token)
	}
	if token.RefreshToken != "rt-1" {
		t.Fatalf("refresh token = %q", token.RefreshToken)
	}
	if token.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z") != "2030-01-02T03:04:05Z" {
		t.Fatalf("expiry = %s", token.ExpiresAt)
	}
	if token.CodeVerifier != "verifier" {
		t.Fatalf("code verifier = %q", token.CodeVerifier)
	}
}

// TestExchangeCodeSurfacesErrors pins the STS error mapping.
func TestExchangeCodeSurfacesErrors(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid authorization code"}}`))
	}))
	defer server.Close()

	keyPair, _ := GenerateDpopKeyPair()
	_, errExchange := client.ExchangeCode(context.Background(), "bad", "verifier", keyPair, "10000")
	if errExchange == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(errExchange.Error(), "invalid authorization code") {
		t.Fatalf("unexpected error: %v", errExchange)
	}
}

// TestExchangeCodeRequiresDpopKey pins that a missing key is rejected locally.
func TestExchangeCodeRequiresDpopKey(t *testing.T) {
	client := NewClient(nil)
	if _, errExchange := client.ExchangeCode(context.Background(), "code", "verifier", nil, "10000"); errExchange == nil {
		t.Fatal("expected an error when the DPoP key pair is missing")
	}
}

// TestRefreshRequiresDpopKey pins that refresh cannot silently proceed.
func TestRefreshRequiresDpopKey(t *testing.T) {
	client := NewClient(nil)
	if _, errRefresh := client.Refresh(context.Background(), "rt", "verifier", nil); errRefresh == nil {
		t.Fatal("expected an error when the DPoP key pair is missing")
	}
}

// TestGetUserInfoParsesPrincipal pins the caller-identity mapping and that the
// request is HMAC signed.
func TestGetUserInfoParsesPrincipal(t *testing.T) {
	var sawSignature bool
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != CallerIdentityPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		sawSignature = signedHeaderPresent(r)
		if r.Header.Get(SecurityTokenHeader) != "ST789" {
			t.Errorf("missing security token header")
		}
		_, _ = w.Write([]byte(`{
			"principal_urn": "iam::1234567890:user/alice",
			"principal_id": "pid-1",
			"account_id": "acct-9",
			"station_type": "HC"
		}`))
	}))
	defer server.Close()

	info, errInfo := client.GetUserInfo(context.Background(), "AK123", "SK456", "ST789")
	if errInfo != nil {
		t.Fatalf("get user info: %v", errInfo)
	}
	if !sawSignature {
		t.Fatal("caller-identity request was not HMAC signed")
	}
	if info.UserName != "alice" {
		t.Fatalf("user name = %q, want alice", info.UserName)
	}
	if info.UserID != "pid-1" || info.DomainID != "acct-9" {
		t.Fatalf("unexpected identity: %+v", info)
	}
}

// TestUrnLeafMatchesDesktopClient pins the principal URN parsing rule.
//
// The desktop client takes the text after the last ':' and then, when a '/'
// remains, the text after the FIRST '/' — so an agency URN keeps its inner
// slash. Matching that exactly avoids diverging on delegated (agency) accounts.
func TestUrnLeafMatchesDesktopClient(t *testing.T) {
	cases := map[string]string{
		"iam::0123456789:user/alice":           "alice",
		"iam::0123456789:agency/team/deployer": "team/deployer",
		"plain-name":                           "plain-name",
		"":                                     "",
	}
	for urn, want := range cases {
		if got := urnLeaf(urn); got != want {
			t.Errorf("urnLeaf(%q) = %q, want %q", urn, got, want)
		}
	}
}

// TestFetchBuiltinModelsParsesCatalog pins the builtin catalog mapping.
func TestFetchBuiltinModelsParsesCatalog(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != BuiltinModelsPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !signedHeaderPresent(r) {
			t.Error("catalog request was not HMAC signed")
		}
		if r.Header.Get("Agent-Type") != AgentTypePromptCenter {
			t.Errorf("Agent-Type = %q", r.Header.Get("Agent-Type"))
		}
		_, _ = w.Write([]byte(`{"builtinModels":[
			{"model_id":"GLM-5.2","model_name":"GLM-5.2","context_window":202752,"input_context_window":196608,"output_context_window":131072,"model_category":"文本","supports_images":false},
			{"model_id":"Qwen3-VL-235B","model_name":"Qwen3-VL-235B","model_category":"图片理解","supports_images":true},
			{"model_id":"vision-1","model_name":"Vision-1","model_category":"图片理解","supports_images":true},
			{"model_id":"third-1","model_name":"Third One","model_type":"third"},
			{"model_id":"","model_name":""}
		]}`))
	}))
	defer server.Close()

	models, errModels := client.FetchBuiltinModels(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"}, "zh-cn")
	if errModels != nil {
		t.Fatalf("fetch models: %v", errModels)
	}
	// The name blacklist drops Qwen3-VL-235B and the third-party type is
	// filtered, leaving the text model, the vision model, and nothing else.
	if len(models) != 2 {
		t.Fatalf("model count = %d, want 2", len(models))
	}
	if models[0].ID != "GLM-5.2" {
		t.Fatalf("model id = %q", models[0].ID)
	}
	if models[0].ContextWindow != 202752 || models[0].InputContextWindow != 196608 || models[0].OutputContextWindow != 131072 {
		t.Fatalf("unexpected limits: %+v", models[0])
	}
	if models[0].Provider != "inferhub-provider" {
		t.Fatalf("provider = %q, want inferhub-provider", models[0].Provider)
	}
	if !models[1].SupportsImages {
		t.Fatal("multimodal flag was not preserved")
	}
	for _, model := range models {
		if model.ID == "Qwen3-VL-235B" || model.ID == "third-1" {
			t.Fatalf("%s should have been filtered out", model.ID)
		}
	}
}

// TestFetchAgentModelsResolvesCodeAgent pins the two-step agent lookup.
func TestFetchAgentModelsResolvesCodeAgent(t *testing.T) {
	var detailAgent string
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent-center/agents/useragents":
			_, _ = w.Write([]byte(`{"agents":[
				{"agent_id":"other","agent_name":"Other","show_in_ide":true},
				{"agent_id":"code-agent-1","agent_name":"CodeAgent","alias":{"alias_zh_cn":"智能体"},"show_in_ide":true}
			]}`))
		case "/v1/agent-center/agents/detail":
			detailAgent = r.URL.Query().Get("agent_id")
			_, _ = w.Write([]byte(`{"gpts":{"models":[{
				"model_name":"GLM-5.2","model_type":"builtin","think_level":"high",
				"model_parameters":{"model_id":"GLM-5.2","context_window":202752,"output_context_window":131072,"supports_images":false}
			}]}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	models, errModels := client.FetchAgentModels(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"}, "zh-cn")
	if errModels != nil {
		t.Fatalf("fetch agent models: %v", errModels)
	}
	if detailAgent != "code-agent-1" {
		t.Fatalf("detail agent_id = %q, want code-agent-1", detailAgent)
	}
	if len(models) != 1 || models[0].ID != "GLM-5.2" {
		t.Fatalf("unexpected models: %+v", models)
	}
	if models[0].ThinkLevel != "high" {
		t.Fatalf("think level = %q", models[0].ThinkLevel)
	}
}

// TestFetchFreeBenefitModels pins the gateway/config mapping and that a disabled
// program returns an empty list without an error.
func TestFetchFreeBenefitModels(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error_code":"0000","result":{
			"base_url":"https://free.example.com/api/v2",
			"models":[{"model_id":"Free-1","model_name":"Free One","context_window":128000,"max_tokens":8192}]
		}}`))
	}))
	defer server.Close()

	models, errModels := client.FetchFreeBenefitModels(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errModels != nil {
		t.Fatalf("free models: %v", errModels)
	}
	if len(models) != 1 {
		t.Fatalf("model count = %d", len(models))
	}
	if !models[0].IsFreeBenefit || models[0].OutputContextWindow != 8192 {
		t.Fatalf("unexpected free model: %+v", models[0])
	}
	if models[0].BaseURL != "https://free.example.com/api/v2" {
		t.Fatalf("free base url = %q", models[0].BaseURL)
	}

	disabled, server2 := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error_code":"9999","result":{}}`))
	}))
	defer server2.Close()
	list, errList := disabled.FetchFreeBenefitModels(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errList != nil || len(list) != 0 {
		t.Fatalf("disabled program should yield no models: %v %v", list, errList)
	}
}

// TestListActivitiesAndDailyCheckIn pins the welfare read + claim + confirm flow.
func TestListActivitiesAndDailyCheckIn(t *testing.T) {
	var claimBody map[string]any
	var confirmBody map[string]any
	var confirmed bool
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case WelfareDeliveryPath:
			if r.URL.Query().Get("channel") != "DESKTOP" {
				t.Errorf("channel = %q", r.URL.Query().Get("channel"))
			}
			// Mirrors the production payload: campaignId is a NUMBER and the
			// daily check-in is typed USER_LOGIN with extra.triggerEvent
			// user.login. A student activity is present but not claimable.
			_, _ = w.Write([]byte(`{"code":0,"data":{"items":[
				{"campaignId":2,"type":"STUDENT_CERTIFIED","status":null,"claimable":false,"benefitAmount":4000},
				{"campaignId":1,"type":"USER_LOGIN","title":"每日签到领1000 积分","status":"ELIGIBLE","claimable":true,"benefitAmount":1000,"extra":{"triggerEvent":"user.login"}}
			]}}`))
		case WelfareClaimPath:
			_ = json.NewDecoder(r.Body).Decode(&claimBody)
			_, _ = w.Write([]byte(`{"code":0,"data":{"userBenefitId":"ub-1"}}`))
		case WelfareConfirmPath:
			confirmed = true
			_ = json.NewDecoder(r.Body).Decode(&confirmBody)
			_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	activities, errActivities := client.ListActivities(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errActivities != nil {
		t.Fatalf("list activities: %v", errActivities)
	}
	if len(activities) != 2 {
		t.Fatalf("activity count = %d", len(activities))
	}
	// The numeric campaignId must survive as a string, otherwise confirm's
	// campaignId field would be rejected upstream.
	if activities[1].CampaignID != "1" {
		t.Fatalf("numeric campaignId decoded as %q, want \"1\"", activities[1].CampaignID)
	}
	if !activities[1].ClaimableNow() || activities[0].ClaimableNow() {
		t.Fatal("claimable detection is wrong: only the daily check-in (USER_LOGIN/user.login) is claimable")
	}

	result, errCheckIn := client.DailyCheckIn(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errCheckIn != nil {
		t.Fatalf("daily check-in: %v", errCheckIn)
	}
	if !result.Claimed || result.CampaignID != "1" {
		t.Fatalf("unexpected check-in result: %+v", result)
	}
	if result.BenefitAmount != 1000 {
		t.Fatalf("benefit amount = %d, want 1000", result.BenefitAmount)
	}
	if claimBody["campaignId"] != "1" || claimBody["channel"] != "DESKTOP" {
		t.Fatalf("unexpected claim body: %+v", claimBody)
	}
	if idempotentKey, _ := claimBody["idempotentKey"].(string); !strings.HasPrefix(idempotentKey, "claim_1_") {
		t.Fatalf("idempotent key = %q", idempotentKey)
	}
	if !confirmed {
		t.Fatal("the claim was not confirmed")
	}
	// confirm mirrors the desktop client: the body carries ONLY userBenefitId,
	// never campaignId.
	if confirmBody["userBenefitId"] != "ub-1" || len(confirmBody) != 1 {
		t.Fatalf("unexpected confirm body: %+v", confirmBody)
	}
	if result.ConfirmError != "" {
		t.Fatalf("unexpected confirm error: %q", result.ConfirmError)
	}
}

// TestConfirmFailureDoesNotFailCheckIn pins the best-effort semantics of
// confirm: credits are granted by the claim, the desktop client ignores
// confirm failures, and a failing confirm must neither fail the check-in nor
// trigger the simple-benefit fallback (which would double-report the reward).
func TestConfirmFailureDoesNotFailCheckIn(t *testing.T) {
	confirmCalls := 0
	fallbackCalls := 0
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case WelfareDeliveryPath:
			_, _ = w.Write([]byte(`{"code":0,"data":{"items":[
				{"campaignId":1,"type":"USER_LOGIN","title":"每日签到领1000 积分","status":"ELIGIBLE","claimable":true,"benefitAmount":1000,"extra":{"triggerEvent":"user.login"}}
			]}}`))
		case WelfareClaimPath:
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":42}}`))
		case WelfareConfirmPath:
			confirmCalls++
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error_code":"HDN.1000","error_msg":"userBenefitId : unknown exception"}`))
		case BenefitClaimPath:
			fallbackCalls++
			_, _ = w.Write([]byte(`{"error_code":"0000","result":{}}`))
		}
	}))
	defer server.Close()

	result, err := client.DailyCheckIn(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if err != nil {
		t.Fatalf("check-in must succeed even when confirm fails: %v", err)
	}
	if !result.Claimed || result.CampaignID != "1" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.UserBenefitID != "42" {
		t.Fatalf("numeric claim id decoded as %q, want \"42\"", result.UserBenefitID)
	}
	if result.ConfirmError == "" {
		t.Fatal("confirm failure should be surfaced as ConfirmError")
	}
	if confirmCalls != 1 {
		t.Fatalf("confirm calls = %d, want 1", confirmCalls)
	}
	if fallbackCalls != 0 {
		t.Fatal("confirm failure must not trigger the simple-benefit fallback")
	}
}

// TestClaimableNowRecognizesProductionDailyActivity pins the real-world shape:
// the daily check-in is USER_LOGIN with triggerEvent user.login, so matching on
// DAILY_CLAIM alone would never claim it.
func TestClaimableNowRecognizesProductionDailyActivity(t *testing.T) {
	daily := Activity{
		CampaignID:   "1",
		Type:         normalizeActivityType("USER_LOGIN"),
		TriggerEvent: "user.login",
		Title:        "每日签到领1000 积分",
		Claimable:    true,
		Status:       WelfareEligible,
	}
	if !daily.ClaimableNow() {
		t.Fatal("production daily check-in must be claimable")
	}

	// Already claimed today.
	claimed := daily
	claimed.Status = WelfareClaimed
	if claimed.ClaimableNow() {
		t.Fatal("a claimed daily check-in must not be claimable")
	}

	// Not claimable upstream.
	notClaimable := daily
	notClaimable.Claimable = false
	if notClaimable.ClaimableNow() {
		t.Fatal("claimable=false must not be claimable")
	}

	// A non-daily activity must never be picked up.
	student := Activity{Type: "student_certify", Claimable: true, Status: WelfareEligible}
	if student.ClaimableNow() {
		t.Fatal("non-daily activities must not be claimable as check-in")
	}
}

// TestDailyCheckInReportsAlreadyClaimed pins the no-op path.
func TestDailyCheckInReportsAlreadyClaimed(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"items":[
			{"campaignId":"daily-1","type":"DAILY_CLAIM","status":"CONFIRMED","claimable":true,"benefitAmount":1000}
		]}}`))
	}))
	defer server.Close()

	result, errCheckIn := client.DailyCheckIn(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errCheckIn != nil {
		t.Fatalf("daily check-in: %v", errCheckIn)
	}
	if result.Claimed || !result.AlreadyClaimed {
		t.Fatalf("unexpected result: %+v", result)
	}
}

// TestDailyCheckInFallsBackToBenefitAPI pins the developer-gateway fallback.
func TestDailyCheckInFallsBackToBenefitAPI(t *testing.T) {
	var hitBenefit bool
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case WelfareDeliveryPath:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`not found`))
		case BenefitClaimPath:
			hitBenefit = true
			_, _ = w.Write([]byte(`{"error_code":"0000","result":{}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	result, errCheckIn := client.DailyCheckIn(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errCheckIn != nil {
		t.Fatalf("daily check-in: %v", errCheckIn)
	}
	if !result.Claimed || !hitBenefit {
		t.Fatalf("fallback was not used: %+v hit=%v", result, hitBenefit)
	}
}

// TestBalanceUnwrapsResult pins the balance envelope mapping.
func TestBalanceUnwrapsResult(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != BalancePath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"error_code":"0000","result":{"remaining":42}}`))
	}))
	defer server.Close()

	raw, errBalance := client.Balance(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"})
	if errBalance != nil {
		t.Fatalf("balance: %v", errBalance)
	}
	if !strings.Contains(string(raw), `"remaining":42`) {
		t.Fatalf("unexpected balance payload: %s", raw)
	}
}

// TestGetAuthenticationByTicketParsesCredentials pins the ticket fallback shape.
func TestGetAuthenticationByTicketParsesCredentials(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/snap-manager"+TicketPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("ticket_id") != "tid" || r.URL.Query().Get("secret") != "sec" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		if r.Header.Get("plugin-name") != PluginName {
			t.Errorf("plugin-name = %q", r.Header.Get("plugin-name"))
		}
		_, _ = w.Write([]byte(`{
			"user_id":"u1","user_name":"alice","domain_id":"d1",
			"credential":{"access":"AK","secret":"SK","securitytoken":"ST","expires_at":"2030-01-02T03:04:05Z"}
		}`))
	}))
	defer server.Close()

	token, errTicket := client.GetAuthenticationByTicket(context.Background(), "tid", "sec")
	if errTicket != nil {
		t.Fatalf("ticket: %v", errTicket)
	}
	if token.AccessKey != "AK" || token.SecretKey != "SK" || token.SecurityToken != "ST" {
		t.Fatalf("unexpected credentials: %+v", token)
	}
}

// TestParseTrackerRejectsUnknownActivityType keeps normalization explicit.
func TestNormalizeActivityType(t *testing.T) {
	cases := map[string]string{
		"DAILY_CLAIM":       "daily_claim",
		"daily_claim":       "daily_claim",
		"USER_LOGIN":        "login",
		"INVITE_USER":       "invite",
		"STUDENT_CERTIFIED": "student_certify",
		"weird":             "weird",
	}
	for input, want := range cases {
		if got := normalizeActivityType(input); got != want {
			t.Errorf("normalizeActivityType(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestIsFreeBenefitModel pins the free-model routing classification.
//
// The limited-time free models are rejected with InferHub.002002009.404 unless
// the request carries maas_type=benefit, so this predicate gates a header that
// decides whether those models work at all.
func TestIsFreeBenefitModel(t *testing.T) {
	benefit := []string{
		"deepseek-v4-flash-0731",
		"deepseek-v4-pro-0813",
		"glm-5.3-flash",
		"  glm-5.3-flash  ",
		"GLM-5.3-Flash",
	}
	for _, id := range benefit {
		if !IsFreeBenefitModel(id) {
			t.Errorf("IsFreeBenefitModel(%q) = false, want true", id)
		}
	}
	standard := []string{
		"GLM-5.2",
		"GLM-5.1",
		"glm-5.2-sft-harmony",
		"openpangu-2.0-pro",
		"openpangu-2.0-flash",
		"Qwen3-VL-235B",
		// Serving the benefit header for a standard model is harmless but
		// unnecessary, so unrelated ids must not be classified as benefit.
		"deepseek-v4-pro",
		"glm-5.2",
		"",
	}
	for _, id := range standard {
		if IsFreeBenefitModel(id) {
			t.Errorf("IsFreeBenefitModel(%q) = true, want false", id)
		}
	}
}

// TestFetchTokenBalanceExplainsDisabledBenefitGateway pins the diagnostic path
// for the developer gateway's "invalid channel, group not found" envelope. The
// gateway answers HTTP 200 for every endpoint once the account's benefit program
// is off, so the balance read must say so instead of surfacing a bare 4000.
//
// The statistics ledger is the primary read, so the handler fails it first to
// force the gateway fallback this test is about.
func TestFetchTokenBalanceExplainsDisabledBenefitGateway(t *testing.T) {
	var gatewayHit, statusHit bool
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PluginStatisticsPath:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error_code":"APIG.0101","error_msg":"not published"}`))
		case BalancePath:
			gatewayHit = true
			_, _ = w.Write([]byte(`{"error_code":"4000","error_msg":"invalid channel, group not found","result":null}`))
		case BenefitGatewayConfigPath:
			statusHit = true
			_, _ = w.Write([]byte(`{"enabled":false,"is_examination_domain":false,"examination_model_enable":false}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	_, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance == nil {
		t.Fatal("expected an error")
	}
	if !gatewayHit || !statusHit {
		t.Fatalf("gateway=%v status=%v, want both reads", gatewayHit, statusHit)
	}
	if !IsBenefitGatewayDisabled(errBalance) {
		t.Fatalf("error is not typed as a disabled gateway: %v", errBalance)
	}
	var disabled *BenefitGatewayDisabledError
	if !errors.As(errBalance, &disabled) {
		t.Fatalf("errors.As failed: %v", errBalance)
	}
	if disabled.Enabled == nil || *disabled.Enabled {
		t.Fatalf("switch was not folded in: %+v", disabled)
	}
	if !strings.Contains(errBalance.Error(), "disabled") {
		t.Fatalf("message does not explain the state: %v", errBalance)
	}
}

// TestFetchTokenBalanceKeepsOtherEnvelopeErrors pins that unrelated upstream
// codes still surface as plain API errors.
func TestFetchTokenBalanceKeepsOtherEnvelopeErrors(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PluginStatisticsPath:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error_code":"APIG.0101","error_msg":"not published"}`))
		case BenefitGatewayConfigPath:
			t.Errorf("the switch must not be read for an unrelated code")
		default:
			_, _ = w.Write([]byte(`{"error_code":"1001","error_msg":"quota not provisioned","result":null}`))
		}
	}))
	defer server.Close()

	_, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance == nil {
		t.Fatal("expected an error")
	}
	if IsBenefitGatewayDisabled(errBalance) {
		t.Fatalf("unrelated code was misclassified: %v", errBalance)
	}
	var apiErr *Error
	if !errors.As(errBalance, &apiErr) || apiErr.Code != "1001" {
		t.Fatalf("unexpected error: %v", errBalance)
	}
}

// TestDisabledBenefitGatewayMessageWithoutSwitch covers the case where the
// switch itself cannot be read: the message must still name the upstream reason
// instead of pretending the program state is known.
func TestDisabledBenefitGatewayMessageWithoutSwitch(t *testing.T) {
	disabled := &BenefitGatewayDisabledError{Code: ErrCodeNoChannelGroup, Message: "invalid channel, group not found"}
	message := disabled.Error()
	if strings.Contains(message, "disabled") && !strings.Contains(message, "channel group") {
		t.Fatalf("message must not claim the program is off without evidence: %q", message)
	}
	if !strings.Contains(message, ErrCodeNoChannelGroup) {
		t.Fatalf("message must carry the upstream code: %q", message)
	}
	if IsBenefitGatewayDisabled(nil) {
		t.Fatal("nil must not be classified as disabled")
	}
}

// TestFetchTokenBalancePrefersPluginStatistics pins the primary credit read: the
// 0.2.8 desktop client replaced the developer-gateway balance call with
// snap-manager/v1/statistics/plugin, and that is what the account menu renders.
// The gateway must not be consulted when this read succeeds.
func TestFetchTokenBalancePrefersPluginStatistics(t *testing.T) {
	var gatewayHit bool
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PluginStatisticsPath:
			if r.Header.Get("Agent-Type") != "" {
				t.Errorf("snap-manager statistics takes no Agent-Type, got %q", r.Header.Get("Agent-Type"))
			}
			if !strings.HasPrefix(r.Header.Get("Authorization"), Algorithm+" Access=") {
				t.Errorf("statistics request was not signed")
			}
			_, _ = w.Write([]byte(`{
				"metrics":[
					{"name":"usageTokenChatMessages","usage_token_num":120,"package_token_amount":1000},
					{"name":"usageTotalPackageCredit","package_credit_amount":500,"package_credit_used":120,
					 "package_credit_remain":380,"package_credit_expiring_amount":40}
				],
				"package":{"spec_code":"codearts.agent.individual.trial","status":"normal",
					"package_name_en":"Trial","is_token_package":false,"is_credit_package":true},
				"start_date":"2026-10-16","end_date":"2026-11-17"}`))
		case BalancePath:
			gatewayHit = true
			_, _ = w.Write([]byte(`{"error_code":"0000","result":{"total_quota":1}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	balance, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance != nil {
		t.Fatalf("balance: %v", errBalance)
	}
	if gatewayHit {
		t.Error("the developer gateway must not be called when the ledger answers")
	}
	if !balance.CreditsKnown {
		t.Error("credits_known must be set for a ledger read")
	}
	if balance.TotalQuota != 500 || balance.UsedAmount != 120 || balance.TotalBalance != 380 {
		t.Fatalf("credit figures not mapped: %+v", balance)
	}
	if balance.ExpiringAmount != 40 {
		t.Fatalf("expiring amount not mapped: %d", balance.ExpiringAmount)
	}
	if balance.DailyTokensUsed != 120 || balance.DailyTokenLimit != 1000 {
		t.Fatalf("token figures not mapped: %+v", balance)
	}
	if balance.Package != "Trial" || balance.SpecCode != "codearts.agent.individual.trial" {
		t.Fatalf("package identity not mapped: %+v", balance)
	}
}

// TestFetchTokenBalanceReportsAZeroLedgerAsKnown pins that an all-zero package
// still reads as a successful, known balance. "Trial with no credits" and
// "could not read the ledger" must not collapse into the same answer.
func TestFetchTokenBalanceReportsAZeroLedgerAsKnown(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"metrics":[
				{"name":"usageTokenChatMessages","usage_token_num":0,"package_token_amount":0},
				{"name":"usageTotalPackageCredit","package_credit_amount":0,"package_credit_used":0,
				 "package_credit_remain":0,"package_credit_expiring_amount":0}
			],
			"package":{"spec_code":"codearts.agent.individual.trial","status":"normal","package_name_en":"Trial"}}`))
	}))
	defer server.Close()

	balance, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance != nil {
		t.Fatalf("balance: %v", errBalance)
	}
	if !balance.CreditsKnown {
		t.Fatal("an all-zero ledger is still a known balance")
	}
	if balance.TotalQuota != 0 || balance.TotalBalance != 0 {
		t.Fatalf("unexpected figures: %+v", balance)
	}
}

// TestFetchTokenBalanceFallsBackToTheGateway pins that the developer gateway is
// still consulted when the ledger cannot be read, so an account whose benefit
// program is on keeps a working read.
func TestFetchTokenBalanceFallsBackToTheGateway(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == PluginStatisticsPath {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error_code":"5000","error_msg":"internal error"}`))
			return
		}
		_, _ = w.Write([]byte(`{"error_code":"0000","result":{"channel":"codearts","total_quota":900,"total_balance":700,"used_amount":200}}`))
	}))
	defer server.Close()

	balance, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance != nil {
		t.Fatalf("balance: %v", errBalance)
	}
	if balance.TotalQuota != 900 || balance.TotalBalance != 700 {
		t.Fatalf("unexpected fallback figures: %+v", balance)
	}
	if balance.CreditsKnown {
		t.Error("the gateway path does not prove the ledger was read")
	}
}

// TestBenefitRouteLearnedFromCatalogue pins that a benefit model the catalogue
// discovers is routable without a code change.
//
// The fixed seed list only covers models known at build time. A pool model
// published later would otherwise be requested without maas_type and rejected
// with InferHub.002002009.404, which reads like a broken model id rather than a
// missing routing header.
func TestBenefitRouteLearnedFromCatalogue(t *testing.T) {
	const lateModel = "deepseek-v5-turbo-0101"
	if _, known := BenefitRouteFor(lateModel); known {
		t.Fatalf("%s is unexpectedly already routable", lateModel)
	}

	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error_code":"0000","result":{
			"base_url":"https://free.example.com/api/v2",
			"models":[{"model_id":"` + lateModel + `","model_name":"` + lateModel + `"}]
		}}`))
	}))
	defer server.Close()

	if _, errFetch := client.FetchFreeBenefitModels(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"}); errFetch != nil {
		t.Fatalf("fetch free models: %v", errFetch)
	}

	channel, known := BenefitRouteFor(lateModel)
	if !known {
		t.Fatalf("%s was not registered by the catalogue fetch", lateModel)
	}
	if channel != BenefitChannelBenefit {
		t.Fatalf("channel = %q, want %q", channel, BenefitChannelBenefit)
	}
	if !IsFreeBenefitModel(lateModel) {
		t.Fatalf("%s should classify as a free-benefit model", lateModel)
	}
}

// TestEducationChannelIsNotReportedAsFreePool pins that the two entitlements stay
// distinct: a caller that asks "is this the free pool" must not be told yes for
// an education model, because they are billed and gated differently.
func TestEducationChannelIsNotReportedAsFreePool(t *testing.T) {
	const educationModel = "edu-only-model"
	RegisterBenefitRoute(educationModel, BenefitChannelEducation)

	channel, known := BenefitRouteFor(educationModel)
	if !known || channel != BenefitChannelEducation {
		t.Fatalf("route = (%q, %v), want (%q, true)", channel, known, BenefitChannelEducation)
	}
	if IsFreeBenefitModel(educationModel) {
		t.Fatal("an education model must not be reported as part of the free pool")
	}
}

// TestFastModelRouteIsLearnedFromAgentDetail pins that model_mode is carried from
// the agent detail top level.
//
// model_mode sits beside model_parameters, not inside it, so a naive expansion of
// the parameter block drops it. Without it the fast-channel header is never sent.
func TestFastModelRouteIsLearnedFromAgentDetail(t *testing.T) {
	const fastID = "GLM-5.2-FAST"
	if IsFastModel(fastID) {
		t.Fatalf("%s is unexpectedly already a fast model", fastID)
	}

	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent-center/agents/useragents":
			_, _ = w.Write([]byte(`{"agents":[{"agent_id":"code-agent-1","agent_name":"CodeAgent","alias":{"alias_zh_cn":"智能体"},"show_in_ide":true}]}`))
		case "/v1/agent-center/agents/detail":
			_, _ = w.Write([]byte(`{"gpts":{"models":[{
				"model_name":"` + fastID + `","model_type":"builtin","model_mode":"fastmode",
				"model_parameters":{"model_id":"` + fastID + `","context_window":202752}
			},{
				"model_name":"GLM-5.2","model_type":"builtin",
				"model_parameters":{"model_id":"GLM-5.2","context_window":202752}
			}]}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	if _, errModels := client.FetchAgentModels(context.Background(), Credentials{AccessKey: "AK", SecretKey: "SK"}, "zh-cn"); errModels != nil {
		t.Fatalf("fetch agent models: %v", errModels)
	}

	if !IsFastModel(fastID) {
		t.Fatalf("%s should be registered as a fast model", fastID)
	}
	if IsFastModel("GLM-5.2") {
		t.Fatal("a model without model_mode must not be registered as fast")
	}
}

// TestTerminalRefreshErrors pins that a spent STS refresh token is classified as
// terminal rather than retryable.
//
// Huawei Cloud rotates the refresh token on every exchange, so a replay answers
// STS5.1806 "the refresh token has been used". The shared refresh loop retries
// non-terminal errors indefinitely, which would leave the account reporting a
// retryable failure forever instead of asking for a fresh sign-in.
func TestTerminalRefreshErrors(t *testing.T) {
	cases := []struct {
		code     string
		terminal bool
	}{
		{ErrCodeRefreshTokenUsed, true},
		{ErrCodeRefreshTokenInvalid, true},
		{ErrCodeRefreshTokenExpired, true},
		// A shape or transport complaint is not terminal: the same credential
		// may refresh successfully once the cause is fixed.
		{ErrCodeNoChannelGroup, false},
		{"STS5.1804", false},
		{"", false},
	}
	for _, tc := range cases {
		err := &Error{StatusCode: http.StatusBadRequest, Code: tc.code, Message: "upstream said no"}
		if got := err.IsTerminalRefreshError(); got != tc.terminal {
			t.Errorf("IsTerminalRefreshError(%q) = %v, want %v", tc.code, got, tc.terminal)
		}
	}
	if (&Error{}).IsTerminalRefreshError() {
		t.Error("an empty error must not be terminal")
	}
	var nilErr *Error
	if nilErr.IsTerminalRefreshError() {
		t.Error("a nil error must not be terminal")
	}
}

// TestRotationUnavailableIsTerminal pins that a credential missing its DPoP key
// pair is not retried forever.
//
// Huawei Cloud STS re-validates proof of possession on every exchange, so a
// credential without the original key can never refresh. Retrying it leaves the
// account reporting a retryable error indefinitely instead of asking for a fresh
// sign-in.
func TestRotationUnavailableIsTerminal(t *testing.T) {
	err := &RotationUnavailableError{Reason: "the original DPoP key pair"}
	if !err.IsTerminalRefreshError() {
		t.Fatal("a missing DPoP key pair must be terminal")
	}
	if !strings.Contains(err.Error(), "DPoP") {
		t.Fatalf("message does not name the missing piece: %q", err.Error())
	}
	var nilErr *RotationUnavailableError
	if !nilErr.IsTerminalRefreshError() {
		t.Fatal("even a nil rotation error reports no retry can help")
	}

	// Refresh must produce that type rather than an untyped error.
	client := NewClient(nil)
	if _, errEmpty := client.Refresh(context.Background(), "  ", "verifier", &DpopKeyPair{}); errEmpty == nil {
		t.Fatal("an empty refresh token must fail")
	} else {
		var typed *RotationUnavailableError
		if !errors.As(errEmpty, &typed) {
			t.Fatalf("empty refresh token error is untyped: %T", errEmpty)
		}
	}
	if _, errKey := client.Refresh(context.Background(), "token", "verifier", nil); errKey == nil {
		t.Fatal("a missing key pair must fail")
	} else {
		var typed *RotationUnavailableError
		if !errors.As(errKey, &typed) {
			t.Fatalf("missing key pair error is untyped: %T", errKey)
		}
	}
}

// TestFetchTokenBalanceNamesTheLedgerWhenBothReadsFail pins that the message
// reports the surface that broke, not the last one tried.
//
// Answering with the gateway's "benefit program disabled" says nothing about
// why the authoritative ledger failed, and a switched-off gateway is the
// expected state for most accounts, so it would send the operator to the wrong
// subsystem.
func TestFetchTokenBalanceNamesTheLedgerWhenBothReadsFail(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PluginStatisticsPath:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error_code":"5000","error_msg":"internal error"}`))
		case BenefitGatewayConfigPath:
			_, _ = w.Write([]byte(`{"enabled":false}`))
		default:
			_, _ = w.Write([]byte(`{"error_code":"4000","error_msg":"invalid channel, group not found","result":null}`))
		}
	}))
	defer server.Close()

	_, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance == nil {
		t.Fatal("expected an error")
	}
	message := errBalance.Error()
	if !strings.Contains(message, PluginStatisticsPath) {
		t.Fatalf("message does not name the failed ledger read: %q", message)
	}
	if !strings.Contains(message, "internal error") {
		t.Fatalf("message does not carry the ledger's own reason: %q", message)
	}
	// The gateway state stays visible as context, but must not be the headline.
	if strings.HasPrefix(message, "codearts: the developer gateway is off") {
		t.Fatalf("gateway state displaced the real cause: %q", message)
	}

	var readErr *BalanceReadError
	if !errors.As(errBalance, &readErr) {
		t.Fatalf("error is not a BalanceReadError: %T", errBalance)
	}
	if !readErr.FallbackDisabled() {
		t.Fatal("the gateway's switched-off state was not recorded")
	}
	// The gateway verdict stays queryable for callers that need it.
	if !IsBenefitGatewayDisabled(errBalance) {
		t.Fatal("the disabled-gateway verdict is no longer reachable")
	}
}

// TestFetchTokenBalanceLedgerOnlyFailureStandalone pins that a ledger failure
// with a working gateway still reports the ledger.
func TestFetchTokenBalanceLedgerOnlyFailureStandalone(t *testing.T) {
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == PluginStatisticsPath {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error_code":"APIG.0301","error_msg":"Incorrect IAM authentication information"}`))
			return
		}
		_, _ = w.Write([]byte(`{"error_code":"1001","error_msg":"quota not provisioned","result":null}`))
	}))
	defer server.Close()

	_, errBalance := client.FetchTokenBalance(context.Background(), "AK", "SK", "ST")
	if errBalance == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(errBalance.Error(), PluginStatisticsPath) {
		t.Fatalf("message does not name the ledger: %q", errBalance.Error())
	}
	// The status mapping must follow the ledger's HTTP status, not the gateway's
	// always-200 envelope.
	var apiErr *Error
	if !errors.As(errBalance, &apiErr) {
		t.Fatalf("no typed upstream error: %T", errBalance)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want the ledger's 403", apiErr.StatusCode)
	}
}
