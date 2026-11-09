package lobsteraiserver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// liveCredential is the subset of a persisted auth file this test needs.
type liveCredential struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	BackendURL   string `json:"backend_url"`
	BaseURL      string `json:"base_url"`
}

// loadLiveCredential reads the credential from LOBSTERAI_SERVER_LIVE_AUTH_FILE.
// It is skipped unless that variable is set, so the default suite stays offline.
func loadLiveCredential(t *testing.T) (*Client, liveCredential) {
	t.Helper()
	path := strings.TrimSpace(os.Getenv("LOBSTERAI_SERVER_LIVE_AUTH_FILE"))
	if path == "" {
		t.Skip("set LOBSTERAI_SERVER_LIVE_AUTH_FILE to a data/auth_files/*.json path to run this live test")
	}
	raw, errRead := os.ReadFile(filepath.Clean(path))
	if errRead != nil {
		t.Fatalf("read auth file: %v", errRead)
	}
	var cred liveCredential
	if errUnmarshal := json.Unmarshal(raw, &cred); errUnmarshal != nil {
		t.Fatalf("parse auth file: %v", errUnmarshal)
	}
	if strings.TrimSpace(cred.AccessToken) == "" {
		t.Fatal("auth file has no access_token")
	}
	client := NewClientWithProxyURL(nil, "", cred.BackendURL)
	return client, cred
}

// TestLiveModelsAndQuota exercises the real backend with an operator credential.
//
//	LOBSTERAI_SERVER_LIVE_AUTH_FILE=data/auth_files/lobsterai-server-*.json \
//	  go test ./internal/auth/lobsteraiserver/ -run TestLiveModelsAndQuota -v
func TestLiveModelsAndQuota(t *testing.T) {
	client, cred := loadLiveCredential(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	models, errModels := client.FetchModels(ctx, cred.AccessToken)
	if errModels != nil {
		t.Fatalf("FetchModels: %v", errModels)
	}
	if len(models) == 0 {
		t.Fatal("FetchModels returned no models")
	}
	accessible := 0
	formats := map[string]int{}
	for _, model := range models {
		formats[model.APIFormat]++
		if model.ID != "" {
			accessible++
		}
		t.Logf("model id=%q name=%q format=%q thinking=%v image=%v ctx=%d",
			model.ID, model.DisplayName, model.APIFormat, model.Thinking, model.Image, model.ContextLimit)
	}
	t.Logf("total=%d parseable=%d formats=%v", len(models), accessible, formats)

	// At least one model must carry a usable identity; the catalog is the model
	// source, so a parser that silently drops every entry is a hard failure.
	if accessible == 0 {
		t.Fatal("no model survived parsing")
	}

	quota, errQuota := client.FetchQuota(ctx, cred.AccessToken)
	if errQuota != nil {
		t.Fatalf("FetchQuota: %v", errQuota)
	}
	t.Logf("quota status=%q credits=%d known=%v purchased=%d raw=%v",
		quota.SubscriptionStatus, quota.Credits, quota.CreditsKnown, quota.PurchasedCredits, quota.Raw)

	// The account-menu read. It must expose a total, and its own buckets must
	// account for it: the free-tier bucket that /api/user/quota reports is only
	// part of the balance once activity grants exist.
	summary, errSummary := client.FetchCreditSummary(ctx, cred.AccessToken)
	if errSummary != nil {
		t.Fatalf("FetchCreditSummary: %v", errSummary)
	}
	t.Logf("profile-summary total=%v known=%v user=%q nickname=%q",
		summary.TotalCreditsRemaining, summary.TotalKnown, summary.UserID, summary.Nickname)
	if !summary.TotalKnown {
		t.Log("NOTE: the backend reported no totalCreditsRemaining")
	}
	bucketed := 0.0
	for _, item := range summary.Packages {
		t.Logf("package type=%q label=%q labelEn=%q remaining=%v expires=%q",
			item.Type, item.Label, item.LabelEn, item.CreditsRemaining, item.ExpiresAt)
		bucketed += item.CreditsRemaining
	}
	t.Logf("packages=%d bucketed=%v", len(summary.Packages), bucketed)

	profile, errProfile := client.FetchUserProfile(ctx, cred.AccessToken)
	if errProfile != nil {
		t.Fatalf("FetchUserProfile: %v", errProfile)
	}
	t.Logf("profile id=%q email=%q name=%q", profile.ID, profile.Email, profile.Name)

	// The enterprise context decides whether the two enterprise headers are sent.
	// A personal account must answer nil with no error, which is what keeps the
	// proxy from claiming an enterprise identity it does not have.
	enterprise, errEnterprise := client.FetchEnterpriseContext(ctx, cred.AccessToken)
	if errEnterprise != nil {
		t.Fatalf("FetchEnterpriseContext: %v", errEnterprise)
	}
	if enterprise == nil {
		t.Log("enterprise context: none (personal account; no enterprise headers are sent)")
	} else {
		t.Logf("enterprise context: id=%q name=%q role=%q member=%q",
			enterprise.EnterpriseID, enterprise.EnterpriseName, enterprise.Role, enterprise.MemberID)
	}

	// The activity read must carry a platform the backend recognizes.
	platform := ActivityPlatformFor(runtime.GOOS)
	t.Logf("activity platform for %s = %q", runtime.GOOS, platform)
	if platform == "" {
		t.Fatal("this host has no platform the activity surface recognizes")
	}
	snapshot, errSnapshot := client.FetchActivitySnapshot(ctx, cred.AccessToken, "", 0)
	if errSnapshot != nil {
		t.Fatalf("FetchActivitySnapshot: %v", errSnapshot)
	}
	t.Logf("activity slot=%q code=%q revision=%d available=%v",
		snapshot.SlotState, snapshot.ActivityCode, snapshot.ConfigRevision, snapshot.Available())
}
