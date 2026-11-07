package management

import (
	"testing"
	"time"

	dumateauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/dumate"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
)

func packageOf(status, total, used string) dumateauth.QuotaPackage {
	return dumateauth.QuotaPackage{
		Status:      status,
		TotalPoints: total,
		UsedPoints:  used,
	}
}

func TestDumateRemainingPointsCountsActiveBucketsOnly(t *testing.T) {
	remaining := remainingPoints([]dumateauth.QuotaPackage{
		packageOf("active", "5000.00", "6.09"),
		packageOf("expired", "100.00", "0.00"),
		packageOf("active", "20.00", "0.00"),
	})
	if remaining != "5013.91" {
		t.Fatalf("remainingPoints() = %q, want 5013.91", remaining)
	}
	if got := remainingPoints(nil); got != "0.00" {
		t.Fatalf("remainingPoints(nil) = %q", got)
	}
	// An unreadable value must not poison the whole bucket family.
	if got := remainingPoints([]dumateauth.QuotaPackage{packageOf("active", "", "x")}); got != "0.00" {
		t.Fatalf("malformed bucket = %q", got)
	}
}

func TestDumateSoonestExpiry(t *testing.T) {
	now := time.Now().Unix()
	subscription := []dumateauth.QuotaPackage{{
		Status: "active", ExpireDate: now + 90_000, PointsExpireDate: now + 90_000,
	}}
	incremental := []dumateauth.QuotaPackage{
		{Status: "active", ExpireDate: now + 30_000},
		{Status: "expired", ExpireDate: now + 5},
		{Status: "active", ExpireDate: now - 10},
	}
	if got := soonestExpiry(subscription, incremental); got != now+30_000 {
		t.Fatalf("soonestExpiry() = %d, want %d", got, now+30_000)
	}
	if got := soonestExpiry(nil, nil); got != 0 {
		t.Fatalf("soonestExpiry(empty) = %d, want 0", got)
	}
	// An inactive bucket must not contribute an expiry even when it is the
	// nearest one in time.
	if got := soonestExpiry(incremental[:1]); got != now+30_000 {
		t.Fatalf("soonestExpiry(single) = %d", got)
	}
}

func TestBuildDumateQuotaResponse(t *testing.T) {
	overview := &dumateauth.PointsOverview{
		IsSubscribed: true,
		UsedPoints:   "6.09",
		TotalPoints:  "5520.00",
		Subscription: []dumateauth.QuotaPackage{packageOf("active", "5000.00", "6.09")},
		Incremental: []dumateauth.QuotaPackage{
			packageOf("active", "20.00", "0.00"),
			packageOf("active", "500.00", "0.00"),
		},
		ModelThrottleInfo: dumateauth.QuotaThrottle{Throttled: true, Reason: "daily_limit", ThrottleType: "soft"},
	}

	response := buildDumateQuotaResponse(overview)
	if response.Provider != constant.Dumate {
		t.Fatalf("provider = %q", response.Provider)
	}
	// Decimal strings are passed through in the console's own units.
	if response.AvailablePoints != "5520.00" || response.UsedPoints != "6.09" {
		t.Fatalf("totals = %+v", response)
	}
	if response.SubscriptionPoints != "4993.91" || response.BonusPoints != "520.00" {
		t.Fatalf("families = %+v", response)
	}
	if !response.IsSubscribed || !response.Throttled || response.ThrottleReason != "daily_limit" {
		t.Fatalf("state = %+v", response)
	}
	if response.ObservedAt == "" {
		t.Fatal("observed_at is required so the panel can show staleness")
	}
}
