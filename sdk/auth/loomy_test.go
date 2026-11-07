package auth

import (
	"context"
	"strings"
	"testing"

	loomyauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/loomy"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
)

// TestLoomyAuthenticatorIdentity pins the provider key and refresh lead.
func TestLoomyAuthenticatorIdentity(t *testing.T) {
	authenticator := NewLoomyAuthenticator()
	if got := authenticator.Provider(); got != constant.Loomy {
		t.Fatalf("Provider() = %q", got)
	}
	// No proactive refresh: probing cannot advance a fixed-window session, so a
	// lead would re-probe the same credential forever.
	if lead := authenticator.RefreshLead(); lead != nil {
		t.Fatalf("RefreshLead() = %v, want nil", *lead)
	}
}

// TestLoomyAuthenticatorRefusesWithoutPrompt verifies the SMS flow reports the
// missing prerequisite instead of starting a flow that cannot complete.
func TestLoomyAuthenticatorRefusesWithoutPrompt(t *testing.T) {
	authenticator := NewLoomyAuthenticator()
	_, errLogin := authenticator.Login(context.Background(), &config.Config{}, &LoginOptions{})
	if errLogin == nil || !strings.Contains(errLogin.Error(), "interactive prompt") {
		t.Fatalf("Login() error = %v", errLogin)
	}
}

// TestLoomySessionRecordCarriesDerivedExpiry verifies the persisted shape: the
// session, the derived 14-day expiry, the account identity and the OAuth kind.
func TestLoomySessionRecordCarriesDerivedExpiry(t *testing.T) {
	token := &loomyauth.TokenData{Session: "abc123", UserID: "100000000000000001", Phone: "13800000000", Nickname: "unit"}
	record := LoomySessionRecord(nil, token, "https://gateway.example/api/v1/")
	if record == nil {
		t.Fatal("record is nil")
	}
	if record.Provider != constant.Loomy {
		t.Fatalf("provider = %q", record.Provider)
	}
	if record.Label != "13800000000" {
		t.Fatalf("label = %q", record.Label)
	}
	if got := record.Metadata["access_token"]; got != "abc123" {
		t.Fatalf("access_token = %v", got)
	}
	if got := record.Metadata["base_url"]; got != "https://gateway.example/api/v1" {
		t.Fatalf("base_url = %v", got)
	}
	if got := record.Metadata["auth_kind"]; got != "oauth" {
		t.Fatalf("auth_kind = %v", got)
	}
	if expired, _ := record.Metadata["expired"].(string); expired == "" {
		t.Fatal("expired is missing")
	}
	if _, hasRefresh := record.Metadata["refresh_token"]; hasRefresh {
		t.Fatal("a refresh token must not be recorded: Loomy has no refresh endpoint")
	}
	if record.Attributes["access_token"] != "abc123" || record.Attributes["base_url"] != "https://gateway.example/api/v1" {
		t.Fatalf("attributes = %+v", record.Attributes)
	}
}

func TestLoomySessionRecordUsesPrefixWhenForced(t *testing.T) {
	token := &loomyauth.TokenData{Session: "abc123", UserID: "100000000000000001"}
	cfg := &config.Config{}
	cfg.ForceModelPrefix = true
	record := LoomySessionRecord(cfg, token, "")
	if record.Prefix != constant.Loomy || record.Metadata["prefix"] != constant.Loomy {
		t.Fatalf("forced prefix was not persisted: prefix=%q metadata=%v", record.Prefix, record.Metadata["prefix"])
	}
	plain := LoomySessionRecord(&config.Config{}, token, "")
	if plain.Prefix != "" {
		t.Fatalf("unexpected prefix without forced mode: %q", plain.Prefix)
	}
	if _, exists := plain.Metadata["prefix"]; exists {
		t.Fatal("unexpected persisted prefix without forced mode")
	}
}
