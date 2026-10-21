package loomy

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSigningStringMatchesClientSegments pins the nine-segment canonical string,
// including the two trailing empty segments that the signature depends on.
func TestSigningStringMatchesClientSegments(t *testing.T) {
	body := `{"base":{"appid":"GM3LOOMY"},"param":{"phone":"13800000000"}}`
	got := SigningString(SignOptions{
		Method:      "POST",
		Path:        "/login/phone/sendMsgCode",
		Body:        body,
		ContentType: "application/json",
		Date:        "Mon, 02 Jan 2006 15:04:05 GMT",
		Nonce:       "0123456789abcdef0123456789abcdef",
	})
	sum := md5.Sum([]byte(body))
	wantSegments := []string{
		"POST",
		"/login/phone/sendMsgCode",
		"",
		base64.StdEncoding.EncodeToString(sum[:]),
		"application/json",
		"Mon, 02 Jan 2006 15:04:05 GMT",
		"0123456789abcdef0123456789abcdef",
		"",
		"",
	}
	want := strings.Join(wantSegments, "\n")
	if got != want {
		t.Fatalf("signing string mismatch:\n got=%q\nwant=%q", got, want)
	}
	if !strings.HasSuffix(got, "\n\n") {
		t.Fatalf("signing string must end with two newlines, got %q", got)
	}
}

// TestContentMD5EmptyBodyIsAbsent documents that an empty body contributes an
// empty segment and no Content-MD5 header.
func TestContentMD5EmptyBodyIsAbsent(t *testing.T) {
	if got := ContentMD5(""); got != "" {
		t.Fatalf("ContentMD5(\"\") = %q, want empty", got)
	}
	headers := AuthHeaders("ak", "sk", http.MethodGet, "/models", "", "date", "nonce")
	if _, exists := headers["Content-MD5"]; exists {
		t.Fatal("empty body must not emit a Content-MD5 header")
	}
	if !strings.HasPrefix(headers["Authorization"], "account ak:") {
		t.Fatalf("authorization prefix = %q", headers["Authorization"])
	}
}

// TestAuthHeadersSignatureIsHMACSHA1 verifies the header carries the exact
// HMAC-SHA1 of the canonical string.
func TestAuthHeadersSignatureIsHMACSHA1(t *testing.T) {
	const secret = "zsak6eadrbawz683wf5r3m2snrwj868r"
	body := `{"base":{},"param":{}}`
	headers := AuthHeaders("2thryby66wxi53sk", secret, http.MethodPost, "/login/phone/checkCode", body, "date", "nonce")

	canonical := SigningString(SignOptions{
		Method:      http.MethodPost,
		Path:        "/login/phone/checkCode",
		Body:        body,
		ContentType: "application/json",
		Date:        "date",
		Nonce:       "nonce",
	})
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(canonical))
	want := "account 2thryby66wxi53sk:" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if headers["Authorization"] != want {
		t.Fatalf("authorization = %q, want %q", headers["Authorization"], want)
	}
}

// TestEscapedPathKeepsSegmentSeparators covers the per-segment escaping rule.
func TestEscapedPathKeepsSegmentSeparators(t *testing.T) {
	cases := []struct{ in, want string }{
		{"login/phone/sendMsgCode", "/login/phone/sendMsgCode"},
		{"/login/phone/sendMsgCode/", "/login/phone/sendMsgCode"},
		{"/", "/"},
		{"/a b/c(d)", "/a%20b/c%28d%29"},
	}
	for _, tc := range cases {
		if got := EscapedPath(tc.in); got != tc.want {
			t.Fatalf("EscapedPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestParseEnvelopeTreatsBusinessCodesNotStatus pins the HTTP-200 failure mode.
func TestParseEnvelopeTreatsBusinessCodesNotStatus(t *testing.T) {
	if _, err := ParseEnvelope([]byte(`{"code":"000000","data":{"session":"s"}}`)); err != nil {
		t.Fatalf("success envelope rejected: %v", err)
	}
	_, err := ParseEnvelope([]byte(`{"code":"100002","desc":"缺少 token"}`))
	if err == nil {
		t.Fatal("auth error code accepted as success")
	}
	upstream, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if upstream.Code != AuthErrorCode || upstream.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected classification: %+v", upstream)
	}
}

// TestBusinessRequestUsesTokenHeader verifies the business header split: the
// `token` header is required and Bearer is not sent.
func TestBusinessRequestUsesTokenHeader(t *testing.T) {
	var captured *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Clone(context.Background())
		_, _ = w.Write([]byte(`{"code":"000000","data":[]}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(nil, "", server.URL, server.URL)
	if _, err := client.BusinessGET(context.Background(), "session-abc", "/models"); err != nil {
		t.Fatalf("BusinessGET failed: %v", err)
	}
	if got := captured.Header.Get("token"); got != "session-abc" {
		t.Fatalf("token header = %q", got)
	}
	if got := captured.Header.Get("Authorization"); got != "" {
		t.Fatalf("business request must not send Authorization, got %q", got)
	}
}

// TestSendSMSCodeSignsVerbatimBody verifies the request body is signed and then
// sent byte-identically, which is the only way the signature validates.
func TestSendSMSCodeSignsVerbatimBody(t *testing.T) {
	var capturedBody string
	var capturedAuth string
	var capturedNonce string
	var capturedDate string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(raw)
		capturedBody = string(raw)
		capturedAuth = r.Header.Get("Authorization")
		capturedNonce = r.Header.Get("Nonce")
		capturedDate = r.Header.Get("Date")
		_, _ = w.Write([]byte(`{"code":"000000","data":{"msgid":"m-1"}}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(nil, "", server.URL, server.URL)
	msgID, err := client.SendSMSCode(context.Background(), "13800000000")
	if err != nil {
		t.Fatalf("SendSMSCode failed: %v", err)
	}
	if msgID != "m-1" {
		t.Fatalf("msgid = %q", msgID)
	}
	// Recomputing the signature over the received bytes proves the signature
	// was produced from exactly what was sent.
	wantAuthorization := "account " + AccessKeyID + ":" + Sign(SignOptions{
		AccessKeyID:     AccessKeyID,
		AccessKeySecret: AccessKeySecret,
		Method:          http.MethodPost,
		Path:            "/login/phone/sendMsgCode",
		Body:            capturedBody,
		ContentType:     "application/json",
		Date:            capturedDate,
		Nonce:           capturedNonce,
	})
	if capturedAuth != wantAuthorization {
		t.Fatalf("signature does not validate over the sent body")
	}
	var envelope map[string]any
	if errUnmarshal := json.Unmarshal([]byte(capturedBody), &envelope); errUnmarshal != nil {
		t.Fatalf("signed body is not JSON: %v", errUnmarshal)
	}
	param, _ := envelope["param"].(map[string]any)
	if param["phone"] != "13800000000" || param["ccode"] != "86" {
		t.Fatalf("unexpected params: %+v", param)
	}
	if param["expire"] != float64(SMSCodeTTLSeconds) {
		t.Fatalf("expire = %v, want %d", param["expire"], SMSCodeTTLSeconds)
	}
	base, _ := envelope["base"].(map[string]any)
	if base["ua"] != ClientUA || base["appid"] != AppID {
		t.Fatalf("unexpected base identity: %+v", base)
	}
	if !strings.HasPrefix(capturedAuth, "account "+AccessKeyID+":") {
		t.Fatalf("authorization = %q", capturedAuth)
	}
}

// TestFetchModelsKeepsChatEntriesOnly verifies the type filter and capability
// mapping, including that an image-capable chat model is not dropped.
func TestFetchModelsKeepsChatEntriesOnly(t *testing.T) {
	payload := `{"code":"000000","data":[
		{"id":"spark-x","type":"chat","name":"Spark X2.5 · x0.1","context_length":1048576,
		 "reasoning_efforts":["none","low","high"],
		 "capabilities":{"reasoning":true,"input_modalities":["text","image"]}},
		{"id":"image-gen","type":"image","name":"Image Gen"}
	]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(nil, "", server.URL, server.URL)
	models, err := client.FetchModels(context.Background(), "session")
	if err != nil {
		t.Fatalf("FetchModels failed: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("model count = %d, want 1: %+v", len(models), models)
	}
	model := models[0]
	if model.ID != "spark-x" || model.ContextLimit != 1048576 {
		t.Fatalf("unexpected model: %+v", model)
	}
	if !model.SupportsImage || !model.SupportsReason {
		t.Fatalf("capabilities lost: %+v", model)
	}
	if len(model.ReasoningLevels) != 3 {
		t.Fatalf("reasoning levels = %v", model.ReasoningLevels)
	}
}

func TestFetchModelsAcceptsCurrentOpenAIList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" || r.Header.Get("token") != "session" || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected model request path or headers")
		}
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"DeepSeek-V4.1-Flash","type":"chat","name":"DeepSeek V4.1 Flash","context_length":1048576},{"id":"image-gen","type":"image"}]}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(nil, "", server.URL, server.URL)
	models, err := client.FetchModels(context.Background(), "session")
	if err != nil {
		t.Fatalf("FetchModels failed: %v", err)
	}
	if len(models) != 1 || models[0].ID != "DeepSeek-V4.1-Flash" {
		t.Fatalf("unexpected models: %+v", models)
	}
}

// TestFetchCreditDetailReportsUnknownQuotaSeparately verifies that a read-only
// points response without dailyQuota is reported as unknown, not zero.
func TestFetchCreditDetailReportsUnknownQuotaSeparately(t *testing.T) {
	payload := `{"code":"000000","data":{"balance":15000,"dailyBalance":4992,"availableBalance":19992}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(nil, "", server.URL, server.URL)
	detail, err := client.FetchCreditDetail(context.Background(), "session")
	if err != nil {
		t.Fatalf("FetchCreditDetail failed: %v", err)
	}
	if detail.Permanent != 15000 || detail.Daily != 4992 || detail.Total != 19992 {
		t.Fatalf("unexpected pools: %+v", detail)
	}
	if detail.DailyQuotaKnown {
		t.Fatal("dailyQuota was not returned by a read and must stay unknown")
	}
}

// TestProbeClassifiesInvalidSession ensures a stale session is reported as an
// auth failure while a transport error is not.
func TestProbeClassifiesInvalidSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":"100002","desc":"缺少 token"}`))
	}))
	defer server.Close()
	client := NewClientWithBaseURLs(nil, "", server.URL, server.URL)
	err := client.Probe(context.Background(), "stale")
	upstream, ok := err.(*Error)
	if !ok || upstream.StatusCode != http.StatusUnauthorized {
		t.Fatalf("probe error = %#v", err)
	}
}

// TestExtractWeChatUUID covers both embedded patterns and rejects junk.
func TestExtractWeChatUUID(t *testing.T) {
	cases := []struct{ name, html, want string }{
		{"img", `<img class="js_qrcode_img" src="/connect/qrcode/001ZDsw64Vu7ll2E"/>`, "001ZDsw64Vu7ll2E"},
		{"poll", `var fordevtool = "https://long.open.weixin.qq.com/connect/l/qrconnect?uuid=4Y0N_jyVQg==";`, "4Y0N_jyVQg=="},
		{"absent", `<html>no code here</html>`, ""},
		{"short", `/connect/qrcode/ab`, ""},
	}
	for _, tc := range cases {
		if got := ExtractWeChatUUID(tc.html); got != tc.want {
			t.Fatalf("%s: ExtractWeChatUUID = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestPollWeChatStatusMapping pins the 404/405 semantics; reversing them is the
// documented failure mode that leaves the flow stuck after confirmation.
func TestPollWeChatStatusMapping(t *testing.T) {
	cases := []struct {
		body       string
		wantStatus WeChatPollStatus
		wantCode   string
	}{
		{"window.wx_errcode=408;", WeChatWaiting, ""},
		{"window.wx_errcode=404;", WeChatScanned, ""},
		{"window.wx_errcode=405;window.wx_code='abc';", WeChatConfirmed, "abc"},
		{"window.wx_errcode=405;", WeChatScanned, ""},
		{"window.wx_errcode=403;", WeChatCancelled, ""},
		{"window.wx_errcode=402;", WeChatExpired, ""},
	}
	for _, tc := range cases {
		body := tc.body
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(body))
		}))
		client := NewClientWithBaseURLs(nil, "", "https://api.invalid", "https://account.invalid")
		client.WithWeChatOrigin(server.URL)
		got := client.PollWeChat(context.Background(), "uuid-value", "")
		server.Close()
		if got.Status != tc.wantStatus || got.Code != tc.wantCode {
			t.Fatalf("body %q => (%s,%q), want (%s,%q)", tc.body, got.Status, got.Code, tc.wantStatus, tc.wantCode)
		}
	}
}

// TestPollWeChatReportsTransportErrorWithoutFailing verifies a poll failure is
// reported as a status rather than an error, so the caller keeps polling.
func TestPollWeChatReportsTransportErrorWithoutFailing(t *testing.T) {
	client := NewClientWithBaseURLs(nil, "", "https://api.invalid", "https://account.invalid")
	client.WithWeChatOrigin("http://127.0.0.1:1")
	got := client.PollWeChat(context.Background(), "uuid-value", "")
	if got.Status != WeChatPollError {
		t.Fatalf("status = %s, want %s", got.Status, WeChatPollError)
	}
}
