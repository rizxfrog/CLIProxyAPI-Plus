package auth

import (
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/airjelly"
)

func TestAirJellySessionRecord(t *testing.T) {
	s := &airjelly.Session{AccessToken: "access", RefreshToken: "refresh", ExpiresAt: 2_000_000_000}
	s.User.ID = "account-id"
	s.User.Email = "user@example.test"
	record := AirJellySessionRecord(s)
	if record.Provider != airjelly.Provider || record.Prefix != airjelly.Provider || record.Metadata["prefix"] != airjelly.Provider {
		t.Fatal("AirJelly account prefix missing")
	}
	if !strings.HasPrefix(record.FileName, "airjelly-") || strings.Contains(record.FileName, "account-id") {
		t.Fatal("filename must be stable and opaque")
	}
	if record.Attributes["refresh_token"] != "refresh" {
		t.Fatal("refresh credential missing")
	}
}
