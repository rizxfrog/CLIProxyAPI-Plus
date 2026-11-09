package auth

import (
	"strings"
	"testing"
	"time"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func TestLobsteraiServerTokenMetadataPersistsRefreshableCredential(t *testing.T) {
	token := &lobsteraiserverauth.TokenData{
		AccessToken:  "acc-1",
		RefreshToken: "ref-1",
		ExpiresIn:    3600,
		ExpiresAt:    time.Unix(1_800_000_000, 0),
	}
	user := &lobsteraiserverauth.User{ID: "u1", Email: "a@b.c", Name: "Ann"}
	metadata := lobsteraiServerTokenMetadata(token, user, nil, nil)

	if metadata["access_token"] != "acc-1" || metadata["refresh_token"] != "ref-1" {
		t.Fatalf("metadata missing tokens: %+v", metadata)
	}
	if metadata["base_url"] != lobsteraiserverauth.DefaultProxyBaseURL {
		t.Fatalf("base_url = %v", metadata["base_url"])
	}
	if metadata["backend_url"] != lobsteraiserverauth.DefaultBackendURL {
		t.Fatalf("backend_url = %v", metadata["backend_url"])
	}
	if metadata["expired"] != time.Unix(1_800_000_000, 0).UTC().Format(time.RFC3339) {
		t.Fatalf("expired = %v", metadata["expired"])
	}
	if metadata["user_id"] != "u1" || metadata["user_email"] != "a@b.c" {
		t.Fatalf("metadata missing account identity: %+v", metadata)
	}
}

func TestLobsteraiServerTokenMetadataKeepsUnknownCreditsAbsent(t *testing.T) {
	token := &lobsteraiserverauth.TokenData{AccessToken: "acc-1"}
	metadata := lobsteraiServerTokenMetadata(token, &lobsteraiserverauth.User{ID: "u1"}, &lobsteraiserverauth.Quota{
		SubscriptionStatus: "free",
	}, nil)
	if metadata["subscription_status"] != "free" {
		t.Fatalf("subscription_status = %v", metadata["subscription_status"])
	}
	if _, present := metadata["quota_remaining"]; present {
		t.Fatalf("quota_remaining must stay absent when the backend omits credits: %+v", metadata)
	}
}

func TestLobsteraiServerTokenMetadataReportsKnownCredits(t *testing.T) {
	metadata := lobsteraiServerTokenMetadata(
		&lobsteraiserverauth.TokenData{AccessToken: "acc-1"},
		&lobsteraiserverauth.User{ID: "u1"},
		&lobsteraiserverauth.Quota{Credits: 42, CreditsKnown: true},
		nil,
	)
	if metadata["quota_remaining"] != int64(42) {
		t.Fatalf("quota_remaining = %v", metadata["quota_remaining"])
	}
}

func TestLobsteraiServerAttributesPinBaseURLAndAccessToken(t *testing.T) {
	attributes := lobsteraiServerAttributes(&lobsteraiserverauth.TokenData{AccessToken: "acc-1"})
	if attributes[coreauth.AttributeAuthKind] != coreauth.AuthKindOAuth {
		t.Fatalf("auth kind = %q", attributes[coreauth.AttributeAuthKind])
	}
	if attributes["base_url"] != lobsteraiserverauth.DefaultProxyBaseURL {
		t.Fatalf("base_url = %q", attributes["base_url"])
	}
	if attributes["backend_url"] != lobsteraiserverauth.DefaultBackendURL {
		t.Fatalf("backend_url = %q", attributes["backend_url"])
	}
	if attributes["access_token"] != "acc-1" {
		t.Fatalf("access_token = %q", attributes["access_token"])
	}
}

func TestLobsteraiServerFileNameIsStableAndSafe(t *testing.T) {
	name := lobsteraiServerFileName(&lobsteraiserverauth.User{Email: "a/b@c.d"})
	if strings.ContainsAny(name, "/\\@ ") {
		t.Fatalf("file name is not filesystem safe: %q", name)
	}
	if !strings.HasPrefix(name, "lobsterai-server-") || !strings.HasSuffix(name, ".json") {
		t.Fatalf("file name = %q", name)
	}
	if name != lobsteraiServerFileName(&lobsteraiserverauth.User{Email: "a/b@c.d"}) {
		t.Fatal("file name must be deterministic for the same account")
	}
}

func TestLobsteraiServerLabelPrefersEmail(t *testing.T) {
	if got := lobsteraiServerLabel(&lobsteraiserverauth.User{Email: "a@b.c", ID: "u1"}); got != "a@b.c" {
		t.Fatalf("label = %q", got)
	}
	if got := lobsteraiServerLabel(nil); got != lobsteraiserverauth.Label {
		t.Fatalf("label = %q", got)
	}
}

func TestMergeLobsteraiServerUserFillsMissingFields(t *testing.T) {
	base := &lobsteraiserverauth.User{ID: "u1"}
	merged := mergeLobsteraiServerUser(base, &lobsteraiserverauth.User{ID: "ignored", Email: "a@b.c", Name: "Ann"})
	if merged.ID != "u1" {
		t.Fatalf("ID = %q, want the base identity preserved", merged.ID)
	}
	if merged.Email != "a@b.c" || merged.Name != "Ann" {
		t.Fatalf("merged = %+v", merged)
	}
}

func TestLobsteraiServerTokenMetadataRecordsTheEnterpriseIdentity(t *testing.T) {
	token := &lobsteraiserverauth.TokenData{AccessToken: "acc-1"}

	// A personal account persists no enterprise identity, which is what keeps the
	// two enterprise headers off every subsequent request.
	personal := lobsteraiServerTokenMetadata(token, &lobsteraiserverauth.User{ID: "u1"}, nil, nil)
	for _, key := range []string{"account_mode", "enterprise_id", "enterprise_name", "enterprise_role"} {
		if _, present := personal[key]; present {
			t.Fatalf("%s must stay absent for a personal account: %+v", key, personal)
		}
	}

	enterprise := lobsteraiServerTokenMetadata(token, &lobsteraiserverauth.User{ID: "u1"}, nil, &lobsteraiserverauth.EnterpriseContext{
		AccountMode:    lobsteraiserverauth.AccountModeEnterprise,
		EnterpriseID:   "1001",
		EnterpriseName: "Example Enterprise",
		Role:           "super_admin",
		MemberID:       "7",
	})
	if enterprise["account_mode"] != lobsteraiserverauth.AccountModeEnterprise {
		t.Fatalf("account_mode = %v", enterprise["account_mode"])
	}
	if enterprise["enterprise_id"] != "1001" {
		t.Fatalf("enterprise_id = %v", enterprise["enterprise_id"])
	}
	if enterprise["enterprise_name"] != "Example Enterprise" || enterprise["enterprise_role"] != "super_admin" {
		t.Fatalf("enterprise metadata = %+v", enterprise)
	}
	if enterprise["enterprise_member_id"] != "7" {
		t.Fatalf("enterprise_member_id = %v", enterprise["enterprise_member_id"])
	}

	// A context without an id must not be persisted: the proxy would then send the
	// account-mode header with an empty id.
	incomplete := lobsteraiServerTokenMetadata(token, &lobsteraiserverauth.User{ID: "u1"}, nil, &lobsteraiserverauth.EnterpriseContext{
		AccountMode: lobsteraiserverauth.AccountModeEnterprise,
	})
	if _, present := incomplete["enterprise_id"]; present {
		t.Fatalf("enterprise_id must stay absent: %+v", incomplete)
	}
}
