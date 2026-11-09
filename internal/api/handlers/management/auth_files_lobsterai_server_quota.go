package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// LobsteraiServerQuota is the read-only credit/plan view of one LobsterAI
// credential, shaped for the management console quota page.
//
// The numbers come from the account backend's own reads
// (GET /api/user/quota, GET /api/user/profile-summary and the client-activities
// slot/context). They are deliberately kept separate from:
//
//   - scheduler cooldown, which is local routing state, not an account balance;
//   - usage accounting, which counts tokens for a single request.
//
// Two credit reads are reported side by side on purpose. Credits is the
// free-tier bucket that GET /api/user/quota exposes and that the client's
// low-credit purchase gate watches. TotalCredits is the account total from
// GET /api/user/profile-summary, which is what the client's own account menu
// renders (profileSummary.totalCreditsRemaining / creditItems). Reporting only
// the former understates any account that also holds activity or daily-login
// grants, so neither read is allowed to stand in for the other.
//
// The daily check-in activity is reported as observed state only here. Claiming
// it is a separate, opt-in action that only the check-in scheduler performs when
// an operator explicitly enables the provider for this credential; this endpoint
// never claims anything. An unknown value is reported as absent
// (CreditsKnown=false / TotalCreditsKnown=false / ActivityAvailable=false)
// rather than as zero or unlimited.
type LobsteraiServerQuota struct {
	// SubscriptionStatus is the raw plan/status string (e.g. "free", "active").
	SubscriptionStatus string `json:"subscription_status,omitempty"`
	// Credits is the remaining free-tier credit balance when the backend reports one.
	Credits int64 `json:"credits"`
	// CreditsKnown reports whether Credits was present; false means unknown, not 0.
	CreditsKnown bool `json:"credits_known"`
	// PurchasedCredits is a separate credit bucket when the plan carries one.
	PurchasedCredits int64 `json:"purchased_credits,omitempty"`

	// TotalCredits is the account's own credit total from profile-summary.
	TotalCredits float64 `json:"total_credits"`
	// TotalCreditsKnown reports whether TotalCredits was present.
	TotalCreditsKnown bool `json:"total_credits_known"`
	// CreditPackages is the per-bucket breakdown behind TotalCredits.
	CreditPackages []LobsteraiServerCreditPackage `json:"credit_packages,omitempty"`

	// ActivityAvailable reports whether a check-in/credit activity slot is open.
	ActivityAvailable bool `json:"activity_available"`
	// ActivityCode identifies the open activity (e.g. "daily_check_in").
	ActivityCode string `json:"activity_code,omitempty"`
	// ActivityTemplate is the activity's template when the backend reports one.
	ActivityTemplate string `json:"activity_template,omitempty"`
	// ActivityPlacement is the slot placement the read targeted.
	ActivityPlacement string `json:"activity_placement,omitempty"`
	// ActivityContext is the raw read-only activity context payload, if any.
	ActivityContext map[string]any `json:"activity_context,omitempty"`
	// ClaimedToday mirrors the activity's own "already claimed today" flag.
	ClaimedToday bool `json:"claimed_today,omitempty"`
	// ClaimedDays and ClaimedCredits are the activity's own running counters.
	ClaimedDays    int64   `json:"claimed_days,omitempty"`
	ClaimedCredits float64 `json:"claimed_credits,omitempty"`

	// AccountMode is the credential's account mode ("enterprise" once the account
	// joined an enterprise). Informational: it explains which enterprise headers
	// the proxy sends on the account's behalf, and is absent for a personal
	// account, which sends none.
	AccountMode string `json:"account_mode,omitempty"`
}

// LobsteraiServerCreditPackage is one credit bucket behind the account total.
type LobsteraiServerCreditPackage struct {
	// Type is the machine group code (e.g. "subscription", "bonus").
	Type string `json:"type,omitempty"`
	// Label is the localized bucket name the client shows.
	Label string `json:"label,omitempty"`
	// LabelEn is the bucket's English name when the backend provides one.
	LabelEn string `json:"label_en,omitempty"`
	// CreditsRemaining is the remaining amount in this bucket.
	CreditsRemaining float64 `json:"credits_remaining"`
	// ExpiresAt is the backend's own expiry string, empty when it never expires.
	ExpiresAt string `json:"expires_at,omitempty"`
}

// GetLobsteraiServerQuota returns the read-only credit/plan view for one
// LobsterAI credential, identified by auth_index. It never performs an activity
// action (check-in / claim) on the account's behalf.
func (h *Handler) GetLobsteraiServerQuota(c *gin.Context) {
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		writeQuotaError(c, http.StatusBadRequest, "auth_index is required")
		return
	}
	auth := h.authByIndex(authIndex)
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), constant.LobsteraiServer) {
		writeQuotaError(c, http.StatusNotFound, "lobsterai-server credential not found")
		return
	}
	accessToken := lobsteraiServerQuotaAccessToken(auth)
	if accessToken == "" {
		writeQuotaError(c, http.StatusBadRequest, "lobsterai-server credential has no access token")
		return
	}

	backendURL := lobsteraiServerQuotaBaseURL(auth)
	client := lobsteraiserverauth.NewClientWithProxyURL(h.cfg, auth.ProxyURL, backendURL)
	// The backend scopes an enterprise account to its pool from these headers, so
	// the management read must present the same identity as the client's requests.
	client.SetKeyfromIdentity("", "", lobsteraiServerQuotaField(auth, "client_version"), "")
	client.SetEnterpriseContext(
		lobsteraiServerQuotaField(auth, "account_mode"),
		lobsteraiServerQuotaField(auth, "enterprise_id"),
	)
	client.SetActivityPlatform(lobsteraiServerQuotaField(auth, "activity_platform"))

	quota, errQuota := client.FetchQuota(c.Request.Context(), accessToken)
	if errQuota != nil {
		writeQuotaError(c, http.StatusBadGateway, errQuota.Error())
		return
	}

	out := LobsteraiServerQuota{
		SubscriptionStatus: quota.SubscriptionStatus,
		Credits:            quota.Credits,
		CreditsKnown:       quota.CreditsKnown,
		PurchasedCredits:   quota.PurchasedCredits,
		AccountMode:        lobsteraiServerQuotaField(auth, "account_mode"),
	}
	// The credit summary is best-effort like the activity read below: a failure
	// there must not hide the free-tier balance, and an absent total stays
	// explicitly unknown instead of becoming a zero.
	if summary, errSummary := client.FetchCreditSummary(c.Request.Context(), accessToken); errSummary == nil && summary != nil {
		out.TotalCredits = summary.TotalCreditsRemaining
		out.TotalCreditsKnown = summary.TotalKnown
		if len(summary.Packages) > 0 {
			packages := make([]LobsteraiServerCreditPackage, 0, len(summary.Packages))
			for _, item := range summary.Packages {
				packages = append(packages, LobsteraiServerCreditPackage{
					Type:             item.Type,
					Label:            item.Label,
					LabelEn:          item.LabelEn,
					CreditsRemaining: item.CreditsRemaining,
					ExpiresAt:        item.ExpiresAt,
				})
			}
			out.CreditPackages = packages
		}
	}
	// The activity snapshot is best-effort: an unavailable activity must not
	// fail the whole quota read.
	if snapshot, errActivity := client.FetchActivitySnapshot(
		c.Request.Context(),
		accessToken,
		lobsteraiserverauth.ActivityPlacementDesktopSidebar,
		lobsteraiserverauth.ActivityContainerAPIVersionDailyCheckIn,
	); errActivity == nil && snapshot != nil {
		out.ActivityPlacement = snapshot.Placement
		out.ActivityAvailable = snapshot.Available()
		out.ActivityCode = snapshot.ActivityCode
		out.ActivityContext = snapshot.Context
		out.ClaimedToday = snapshot.ClaimedToday()
		out.ClaimedDays = snapshot.ClaimedDays()
		out.ClaimedCredits = snapshot.ClaimedCredits()
		if snapshot.Context != nil {
			out.ActivityTemplate = lobsteraiServerQuotaString(snapshot.Context, "template", "activityTemplate")
		}
	}
	c.JSON(http.StatusOK, out)
}

// lobsteraiServerQuotaAccessToken resolves the account access token that
// authorizes the quota and activity reads.
func lobsteraiServerQuotaAccessToken(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if token := strings.TrimSpace(auth.Attributes["access_token"]); token != "" {
			return token
		}
	}
	if auth.Metadata != nil {
		if token, ok := auth.Metadata["access_token"].(string); ok {
			return strings.TrimSpace(token)
		}
	}
	return ""
}

// lobsteraiServerQuotaBaseURL resolves the account backend origin.
func lobsteraiServerQuotaBaseURL(auth *coreauth.Auth) string {
	if auth != nil {
		for _, source := range []map[string]string{auth.Attributes} {
			if source == nil {
				continue
			}
			if raw := strings.TrimRight(strings.TrimSpace(source["backend_url"]), "/"); raw != "" {
				return raw
			}
		}
	}
	return lobsteraiserverauth.DefaultBackendURL
}

// lobsteraiServerQuotaField reads a trimmed credential field from attributes,
// then metadata.
func lobsteraiServerQuotaField(auth *coreauth.Auth, keys ...string) string {
	if auth == nil {
		return ""
	}
	for _, key := range keys {
		if auth.Attributes != nil {
			if value := strings.TrimSpace(auth.Attributes[key]); value != "" {
				return value
			}
		}
		if auth.Metadata != nil {
			if value, ok := auth.Metadata[key].(string); ok {
				if trimmed := strings.TrimSpace(value); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}

// lobsteraiServerQuotaString reads the first non-empty string field.
func lobsteraiServerQuotaString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}
