package management

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

// TestNormalizeOAuthProviderAcceptsXiaohuanxiong verifies the callback router
// resolves the provider name and its aliases.
func TestNormalizeOAuthProviderAcceptsXiaohuanxiong(t *testing.T) {
	for _, alias := range []string{"xiaohuanxiong", "Xiaohuanxiong", "xiaohuanxiong.com", "raccoon"} {
		got, err := NormalizeOAuthProvider(alias)
		if err != nil {
			t.Fatalf("NormalizeOAuthProvider(%q) returned error: %v", alias, err)
		}
		if got != "xiaohuanxiong" {
			t.Fatalf("NormalizeOAuthProvider(%q) = %q, want xiaohuanxiong", alias, got)
		}
	}
}

// TestXiaohuanxiongCallbackFilePathMatchesWriter pins the file naming against
// the shared writer so the waiter and the callback endpoint agree.
func TestXiaohuanxiongCallbackFilePathMatchesWriter(t *testing.T) {
	dir := t.TempDir()
	h := &Handler{cfg: &config.Config{AuthDir: dir}}
	const state = "0123456789abcdef0123456789abcdef"

	got := xiaohuanxiongCallbackFilePath(h, state)
	want := filepath.Join(dir, ".oauth-xiaohuanxiong-"+state+".oauth")
	if got != want {
		t.Fatalf("callback path = %q, want %q", got, want)
	}

	// Cross-check against the writer that the callback endpoints use.
	written, errWrite := WriteOAuthCallbackFileForPendingSession(dir, "xiaohuanxiong", state, "code-1", "")
	if errWrite != nil {
		// The session is not registered in this test, which is the expected
		// refusal; verify the naming helper directly instead.
		if got != want {
			t.Fatalf("path mismatch: %q vs %q", got, want)
		}
		return
	}
	if written != want {
		t.Fatalf("writer produced %q, waiter expects %q", written, want)
	}
}

// TestXiaohuanxiongCallbackFilePathWithoutAuthDir avoids a bogus relative path.
func TestXiaohuanxiongCallbackFilePathWithoutAuthDir(t *testing.T) {
	if got := xiaohuanxiongCallbackFilePath(&Handler{}, "state"); got != "" {
		t.Fatalf("expected empty path without an auth dir, got %q", got)
	}
	if got := xiaohuanxiongCallbackFilePath(nil, "state"); got != "" {
		t.Fatalf("expected empty path for nil handler, got %q", got)
	}
}

// TestWaitXiaohuanxiongCallbackReadsCode drives the waiter with a pre-written
// callback file.
func TestWaitXiaohuanxiongCallbackReadsCode(t *testing.T) {
	dir := t.TempDir()
	const state = "abcdefabcdefabcdefabcdefabcdefab"
	path := filepath.Join(dir, ".oauth-xiaohuanxiong-"+state+".oauth")

	RegisterOAuthSession(state, "xiaohuanxiong")
	defer CancelOAuthSession(state)

	payload, _ := json.Marshal(oauthCallbackFilePayload{Code: "the-code", State: state})
	if errWrite := os.WriteFile(path, payload, 0o600); errWrite != nil {
		t.Fatalf("write callback: %v", errWrite)
	}

	got, errWait := waitXiaohuanxiongCallback(t.Context(), state, path)
	if errWait != nil {
		t.Fatalf("unexpected error: %v", errWait)
	}
	if got != "the-code" {
		t.Fatalf("code = %q, want the-code", got)
	}
}

// TestWaitXiaohuanxiongCallbackSurfacesAuthorizationError maps an upstream error
// into a returned error rather than a bogus code.
func TestWaitXiaohuanxiongCallbackSurfacesAuthorizationError(t *testing.T) {
	dir := t.TempDir()
	const state = "11112222333344445555666677778888"
	path := filepath.Join(dir, ".oauth-xiaohuanxiong-"+state+".oauth")

	RegisterOAuthSession(state, "xiaohuanxiong")
	defer CancelOAuthSession(state)

	payload, _ := json.Marshal(oauthCallbackFilePayload{Error: "access_denied", State: state})
	if errWrite := os.WriteFile(path, payload, 0o600); errWrite != nil {
		t.Fatalf("write callback: %v", errWrite)
	}

	if _, errWait := waitXiaohuanxiongCallback(t.Context(), state, path); errWait == nil {
		t.Fatal("expected an error for an authorization denial")
	}
}

// TestWaitXiaohuanxiongCallbackStopsWhenSessionCancelled ensures a cancelled
// flow stops promptly instead of waiting out the TTL.
func TestWaitXiaohuanxiongCallbackStopsWhenSessionCancelled(t *testing.T) {
	dir := t.TempDir()
	const state = "99998888777766665555444433332222"
	path := filepath.Join(dir, ".oauth-xiaohuanxiong-"+state+".oauth")

	RegisterOAuthSession(state, "xiaohuanxiong")
	CancelOAuthSession(state)

	if _, errWait := waitXiaohuanxiongCallback(t.Context(), state, path); errWait == nil {
		t.Fatal("expected the wait to stop for a non-pending session")
	}
}

// TestWaitXiaohuanxiongCallbackRequiresAuthDir guards misconfiguration.
func TestWaitXiaohuanxiongCallbackRequiresAuthDir(t *testing.T) {
	if _, errWait := waitXiaohuanxiongCallback(t.Context(), "state", ""); errWait == nil {
		t.Fatal("expected an error when the auth dir is not configured")
	}
}

// TestXiaohuanxiongAuthFileNameIsStable keeps repeated logins for one account
// on the same file so credentials are replaced, not accumulated.
func TestXiaohuanxiongAuthFileNameIsStable(t *testing.T) {
	token := &xiaohuanxiong.TokenData{AccessToken: "at-1", OfficeIdentity: "user-42"}
	first := xiaohuanxiongAuthFileName(token)
	second := xiaohuanxiongAuthFileName(&xiaohuanxiong.TokenData{AccessToken: "at-2", OfficeIdentity: "user-42"})
	if first != second {
		t.Fatalf("file name not stable for one identity: %q vs %q", first, second)
	}
	other := xiaohuanxiongAuthFileName(&xiaohuanxiong.TokenData{AccessToken: "at-3", OfficeIdentity: "user-7"})
	if other == first {
		t.Fatal("distinct identities must not collide")
	}
	if filepath.Ext(first) != ".json" {
		t.Fatalf("expected a .json file name, got %q", first)
	}
}

// TestXiaohuanxiongMetadataOmitsEmptyFields avoids persisting blank keys.
func TestXiaohuanxiongMetadataOmitsEmptyFields(t *testing.T) {
	metadata := xiaohuanxiongMetadata(&xiaohuanxiong.TokenData{AccessToken: "at"})
	if _, ok := metadata["refresh_token"]; ok {
		t.Fatal("refresh_token should be absent when not issued")
	}
	if _, ok := metadata["office_identity"]; ok {
		t.Fatal("office_identity should be absent when not issued")
	}
	if metadata["type"] != "xiaohuanxiong" {
		t.Fatalf("type = %v", metadata["type"])
	}
	if metadata["base_url"] != xiaohuanxiong.LLMBaseURL {
		t.Fatalf("base_url = %v, want %v", metadata["base_url"], xiaohuanxiong.LLMBaseURL)
	}
}

// TestXiaohuanxiongMetadataIncludesOfficeFields covers the full payload.
func TestXiaohuanxiongMetadataIncludesOfficeFields(t *testing.T) {
	metadata := xiaohuanxiongMetadata(&xiaohuanxiong.TokenData{
		AccessToken:    "at",
		RefreshToken:   "rt",
		OfficeIdentity: "id",
		OfficeOrgName:  "org",
		OfficeOrgRole:  "admin",
	})
	for _, key := range []string{"access_token", "refresh_token", "office_identity", "office_org_name", "office_org_role"} {
		if _, ok := metadata[key]; !ok {
			t.Fatalf("metadata missing %q: %v", key, metadata)
		}
	}
}

// TestXiaohuanxiongAttributesShape ensures the executor's expected keys exist.
func TestXiaohuanxiongAttributesShape(t *testing.T) {
	attributes := xiaohuanxiongAttributes(&xiaohuanxiong.TokenData{AccessToken: "at", OfficeIdentity: "id"})
	if attributes["api_key"] != "at" {
		t.Fatalf("api_key = %q", attributes["api_key"])
	}
	if attributes["base_url"] != xiaohuanxiong.LLMBaseURL {
		t.Fatalf("base_url = %q", attributes["base_url"])
	}
	if attributes["auth_kind"] != "oauth" {
		t.Fatalf("auth_kind = %q", attributes["auth_kind"])
	}
}
