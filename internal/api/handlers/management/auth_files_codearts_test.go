package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// fakeCodeArtsService is a deterministic stand-in for the OAuth client.
type fakeCodeArtsService struct {
	token       *codeartsauth.TokenData
	userInfo    *codeartsauth.UserInfo
	authURL     string
	lastCode    string
	lastVerify  string
	exchangeErr error
}

func (f *fakeCodeArtsService) BuildAuthorizationURL(verifier, port, ticketID string) (string, error) {
	f.lastVerify = verifier
	return f.authURL, nil
}

func (f *fakeCodeArtsService) ExchangeCode(_ context.Context, code, verifier string, _ *codeartsauth.DpopKeyPair, _ string) (*codeartsauth.TokenData, error) {
	f.lastCode = code
	f.lastVerify = verifier
	if f.exchangeErr != nil {
		return nil, f.exchangeErr
	}
	return f.token, nil
}

func (f *fakeCodeArtsService) GetUserInfo(_ context.Context, _, _, _ string) (*codeartsauth.UserInfo, error) {
	return f.userInfo, nil
}

// withFakeCodeArtsService swaps the package-level factory for one test.
func withFakeCodeArtsService(t *testing.T, fake *fakeCodeArtsService) {
	t.Helper()
	previous := newCodeArtsOAuthService
	newCodeArtsOAuthService = func(cfg *config.Config) codeartsOAuthService { return fake }
	t.Cleanup(func() { newCodeArtsOAuthService = previous })
}

func newCodeArtsTestHandler(t *testing.T) (*Handler, string) {
	t.Helper()
	authDir := t.TempDir()
	handler := &Handler{cfg: &config.Config{AuthDir: authDir}}
	return handler, authDir
}

// TestRequestCodeArtsTokenReturnsAuthorizationURL pins the start endpoint.
func TestRequestCodeArtsTokenReturnsAuthorizationURL(t *testing.T) {
	handler, _ := newCodeArtsTestHandler(t)
	fake := &fakeCodeArtsService{authURL: "https://codearts.huaweicloud.com/portal/authorize?x=1"}
	withFakeCodeArtsService(t, fake)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/codearts-auth-url", nil)

	handler.RequestCodeArtsToken(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if errUnmarshal := json.Unmarshal(recorder.Body.Bytes(), &payload); errUnmarshal != nil {
		t.Fatalf("unmarshal response: %v", errUnmarshal)
	}
	if payload["url"] != "https://codearts.huaweicloud.com/portal/authorize?x=1" {
		t.Fatalf("url = %v", payload["url"])
	}
	state, _ := payload["state"].(string)
	if strings.TrimSpace(state) == "" {
		t.Fatal("state is missing")
	}
	if !IsOAuthSessionPending(state, constant.CodeArts) {
		t.Fatal("the oauth session was not registered as pending")
	}
	if got, _ := payload["callback"].(string); !strings.Contains(got, "/oauth/callback") {
		t.Fatalf("callback = %q", got)
	}

	// A login context must have been stored for the pasted-callback path.
	if login := loadCodeArtsLoginContext(state); login == nil {
		t.Fatal("login context was not stored")
	} else if login.keyPair == nil || login.verifier == "" {
		t.Fatalf("login context is incomplete: %+v", login)
	}
	CompleteOAuthSession(state)
}

// TestPostCodeArtsAuthCallbackCompletesLogin pins the pasted-URL completion path.
func TestPostCodeArtsAuthCallbackCompletesLogin(t *testing.T) {
	handler, authDir := newCodeArtsTestHandler(t)
	fake := &fakeCodeArtsService{
		authURL:  "https://example.com/authorize",
		token:    &codeartsauth.TokenData{AccessKey: "AK", SecretKey: "SK", SecurityToken: "ST", RefreshToken: "rt", CodeVerifier: "v"},
		userInfo: &codeartsauth.UserInfo{UserID: "u1", UserName: "alice", DomainID: "d1"},
	}
	fake.token.DpopKeyPair = &codeartsauth.DpopKeyPair{PrivateKey: `{"kty":"EC"}`, PublicKey: `{"kty":"EC"}`}
	withFakeCodeArtsService(t, fake)

	// Start the flow to register the session + login context.
	startRecorder := httptest.NewRecorder()
	startCtx, _ := gin.CreateTestContext(startRecorder)
	startCtx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/codearts-auth-url", nil)
	handler.RequestCodeArtsToken(startCtx)

	var start map[string]any
	_ = json.Unmarshal(startRecorder.Body.Bytes(), &start)
	state, _ := start["state"].(string)

	body := `{"state":"` + state + `","redirect_url":"http://127.0.0.1:10000/oauth/callback?code=THE_CODE"}`
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v0/management/codearts-auth-callback", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.PostCodeArtsAuthCallback(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if fake.lastCode != "THE_CODE" {
		t.Fatalf("exchange received code %q, want THE_CODE", fake.lastCode)
	}
	if IsOAuthSessionPending(state, constant.CodeArts) {
		t.Fatal("the oauth session should be complete")
	}
	if loadCodeArtsLoginContext(state) != nil {
		t.Fatal("the login context should be forgotten after completion")
	}

	// The credential record must have been persisted to the auth directory.
	entries, errRead := os.ReadDir(authDir)
	if errRead != nil {
		t.Fatalf("read auth dir: %v", errRead)
	}
	var authFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			authFiles = append(authFiles, filepath.Join(authDir, entry.Name()))
		}
	}
	if len(authFiles) != 1 {
		t.Fatalf("persisted %d auth files, want 1: %v", len(authFiles), authFiles)
	}
	data, errReadFile := os.ReadFile(authFiles[0])
	if errReadFile != nil {
		t.Fatalf("read auth file: %v", errReadFile)
	}
	var record map[string]any
	if errUnmarshal := json.Unmarshal(data, &record); errUnmarshal != nil {
		t.Fatalf("unmarshal auth file: %v", errUnmarshal)
	}
	if record["type"] != constant.CodeArts {
		t.Fatalf("type = %v", record["type"])
	}
	if record["access_key"] != "AK" || record["secret_key"] != "SK" || record["security_token"] != "ST" {
		t.Fatalf("credentials not persisted: %+v", record)
	}
	if record["refresh_token"] != "rt" {
		t.Fatalf("refresh token not persisted: %+v", record)
	}
	if record["dpop_private_key"] == nil {
		t.Fatal("the DPoP key pair must be persisted, otherwise refresh can never work")
	}
}

// TestPostCodeArtsAuthCallbackRejectsUnknownState pins state validation.
func TestPostCodeArtsAuthCallbackRejectsUnknownState(t *testing.T) {
	handler, _ := newCodeArtsTestHandler(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v0/management/codearts-auth-callback",
		strings.NewReader(`{"state":"unknown","code":"x"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PostCodeArtsAuthCallback(ctx)
	if recorder.Code == http.StatusOK {
		t.Fatalf("unknown state must be rejected, got %d", recorder.Code)
	}
}

// TestParseCodeArtsCallback pins the callback parsing variants.
func TestParseCodeArtsCallback(t *testing.T) {
	code, errParse := parseCodeArtsCallback("http://127.0.0.1:10000/oauth/callback?code=abc&foo=bar")
	if errParse != nil || code != "abc" {
		t.Fatalf("query form: code=%q err=%v", code, errParse)
	}
	code, errParse = parseCodeArtsCallback("http://127.0.0.1:10000/oauth/callback#code=frag")
	if errParse != nil || code != "frag" {
		t.Fatalf("fragment form: code=%q err=%v", code, errParse)
	}
	code, errParse = parseCodeArtsCallback("bare-authorization-code")
	if errParse != nil || code != "bare-authorization-code" {
		t.Fatalf("bare code: code=%q err=%v", code, errParse)
	}
	if _, errParse = parseCodeArtsCallback("http://127.0.0.1:10000/oauth/callback?other=1"); errParse == nil {
		t.Fatal("expected an error for a callback without a code")
	}
	if _, errParse = parseCodeArtsCallback(""); errParse == nil {
		t.Fatal("expected an error for an empty callback")
	}
}

// TestCodeArtsAuthFileNameIsStable pins that a repeated login overwrites.
func TestCodeArtsAuthFileNameIsStable(t *testing.T) {
	token := &codeartsauth.TokenData{AccessKey: "AK"}
	info := &codeartsauth.UserInfo{UserID: "u1", UserName: "alice"}
	first := codeArtsAuthFileName(token, info)
	second := codeArtsAuthFileName(token, info)
	if first != second {
		t.Fatalf("file name is not stable: %q vs %q", first, second)
	}
	if !strings.HasPrefix(first, "codearts-") || !strings.HasSuffix(first, ".json") {
		t.Fatalf("unexpected file name: %q", first)
	}
	// A different account must produce a different file.
	if codeArtsAuthFileName(token, &codeartsauth.UserInfo{UserID: "u2"}) == first {
		t.Fatal("different accounts must not share a file name")
	}
}

// TestCodeArtsAttributesCarryTriple pins the attribute mapping.
func TestCodeArtsAttributesCarryTriple(t *testing.T) {
	attributes := codeArtsAttributes(
		&codeartsauth.TokenData{AccessKey: "AK", SecretKey: "SK", SecurityToken: "ST"},
		&codeartsauth.UserInfo{UserName: "alice"},
	)
	if attributes["api_key"] != "AK" || attributes["secret_key"] != "SK" || attributes["security_token"] != "ST" {
		t.Fatalf("attributes = %+v", attributes)
	}
	if attributes[coreauth.AttributeAuthKind] != coreauth.AuthKindOAuth {
		t.Fatalf("auth kind = %q", attributes[coreauth.AttributeAuthKind])
	}
	if attributes["user_name"] != "alice" {
		t.Fatalf("user name = %q", attributes["user_name"])
	}
}
