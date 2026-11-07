package management

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	dumateauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/dumate"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
)

// dumateQuotaResponse is the points wallet published to the management panel.
//
// The console groups the wallet into bucket families. Only the two families any
// observed account has produced are summed here; the console's other package
// types (for example purchased top-ups) have no sample yet, so they are not
// guessed at. AvailablePoints stays the console's own total, so a family that
// starts appearing shows up as a visible difference instead of being folded
// into a wrong bucket.
type dumateQuotaResponse struct {
	Provider   string `json:"provider"`
	ObservedAt string `json:"observed_at"`
	// AvailablePoints, UsedPoints and the two family totals are decimal strings
	// passed through in the console's own units (points, 2 decimals).
	AvailablePoints    string `json:"available_points"`
	UsedPoints         string `json:"used_points"`
	IsSubscribed       bool   `json:"is_subscribed"`
	Throttled          bool   `json:"throttled"`
	ThrottleReason     string `json:"throttle_reason"`
	ThrottleType       string `json:"throttle_type"`
	SubscriptionPoints string `json:"subscription_points"`
	BonusPoints        string `json:"bonus_points"`
	// ExpiresAt is the soonest future expiry across active buckets, in Unix
	// seconds; 0 when no bucket reports one.
	ExpiresAt int64 `json:"expires_at"`
}

// GetDumateQuota returns the points wallet for one Dumate credential.
//
// The wallet lives on console.bce.baidu.com and is authorized by the console
// cookie stored in the credential, not by the local backend credential used for
// inference. A file exported without the cookie reports a clear error instead
// of a fabricated zero balance.
func (h *Handler) GetDumateQuota(c *gin.Context) {
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		writeQuotaError(c, http.StatusBadRequest, "auth_index is required")
		return
	}
	auth := h.authByIndex(authIndex)
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), constant.Dumate) {
		writeQuotaError(c, http.StatusNotFound, "dumate credential not found")
		return
	}
	cookie, errCookie := dumateauth.ResolveConsoleCookie(auth.Metadata)
	if errCookie != nil {
		writeQuotaError(c, http.StatusBadRequest, errCookie.Error())
		return
	}

	// The upstream origin is fixed by the provider client. Credential metadata
	// must not be able to redirect a privileged management read to another host.
	client := dumateauth.NewConsoleClient(helps.NewProxyAwareHTTPClient(c.Request.Context(), h.cfg, auth, 0))
	overview, err := client.FetchPointsOverview(c.Request.Context(), cookie, "")
	if err != nil {
		status := http.StatusBadGateway
		var upstreamErr *dumateauth.ConsoleError
		if errors.As(err, &upstreamErr) &&
			(upstreamErr.StatusCode == http.StatusUnauthorized || upstreamErr.StatusCode == http.StatusForbidden) {
			status = upstreamErr.StatusCode
		}
		writeQuotaError(c, status, err.Error())
		return
	}

	c.JSON(http.StatusOK, buildDumateQuotaResponse(overview))
}

func buildDumateQuotaResponse(overview *dumateauth.PointsOverview) dumateQuotaResponse {
	return dumateQuotaResponse{
		Provider:           constant.Dumate,
		ObservedAt:         time.Now().UTC().Format(time.RFC3339),
		AvailablePoints:    overview.TotalPoints,
		UsedPoints:         overview.UsedPoints,
		IsSubscribed:       overview.IsSubscribed,
		Throttled:          overview.ModelThrottleInfo.Throttled,
		ThrottleReason:     overview.ModelThrottleInfo.Reason,
		ThrottleType:       overview.ModelThrottleInfo.ThrottleType,
		SubscriptionPoints: remainingPoints(overview.Subscription),
		BonusPoints:        remainingPoints(overview.Incremental),
		ExpiresAt:          soonestExpiry(overview.Subscription, overview.Incremental),
	}
}

// remainingPoints sums total-used across active buckets, formatted to 2 decimals.
func remainingPoints(packages []dumateauth.QuotaPackage) string {
	total := 0.0
	for _, entry := range packages {
		if !strings.EqualFold(strings.TrimSpace(entry.Status), "active") {
			continue
		}
		total += parsePoints(entry.TotalPoints) - parsePoints(entry.UsedPoints)
	}
	return strconv.FormatFloat(total, 'f', 2, 64)
}

// soonestExpiry returns the earliest future points expiry across active buckets.
func soonestExpiry(buckets ...[]dumateauth.QuotaPackage) int64 {
	now := time.Now().Unix()
	var soonest int64
	for _, packages := range buckets {
		for _, entry := range packages {
			if !strings.EqualFold(strings.TrimSpace(entry.Status), "active") {
				continue
			}
			expiry := entry.PointsExpireDate
			if expiry == 0 {
				expiry = entry.ExpireDate
			}
			if expiry <= now {
				continue
			}
			if soonest == 0 || expiry < soonest {
				soonest = expiry
			}
		}
	}
	return soonest
}

// parsePoints reads a decimal point string; malformed values count as zero so a
// single odd bucket cannot fail the whole read.
func parsePoints(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}
