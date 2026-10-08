package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	floatboatauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/floatboat"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// FloatboatQuota is the measured allowance of one FloatBoat credential, shaped
// for the management console quota page.
//
// The numbers come from the inference gateway's OpenAI-compatible billing reads
// (/v1/dashboard/billing/subscription and /v1/dashboard/billing/usage), which are
// the only measured allowance the product exposes. They are deliberately kept
// separate from:
//
//   - the login-time snapshot fields on the credential (has_active_subscription,
//     quota_remaining), which are static at sign-in and not refreshed; and
//   - scheduler cooldown, which is local routing state, not an account balance.
//
// An unknown value is reported as absent (RemainingKnown=false) rather than as
// zero or unlimited, because the two reads are independent and either may fail.
type FloatboatQuota struct {
	// Currency is the denomination of the amounts below.
	Currency string `json:"currency"`
	// SoftLimit, HardLimit and Used are the gateway's own amounts.
	SoftLimit float64 `json:"soft_limit"`
	HardLimit float64 `json:"hard_limit"`
	Used      float64 `json:"used"`
	// Remaining is HardLimit-Used, valid only when RemainingKnown is true.
	Remaining      float64 `json:"remaining"`
	RemainingKnown bool    `json:"remaining_known"`
	// HasPaymentMethod mirrors the gateway's flag.
	HasPaymentMethod bool `json:"has_payment_method"`
	// HasActiveSubscription is the login-time snapshot, surfaced as-is.
	HasActiveSubscription bool `json:"has_active_subscription"`
	// Groups are the gateway entitlement groups the inference key belongs to.
	Groups []string `json:"groups,omitempty"`
}

// GetFloatboatQuota returns the measured allowance for one FloatBoat credential,
// identified by auth_index.
func (h *Handler) GetFloatboatQuota(c *gin.Context) {
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		writeQuotaError(c, http.StatusBadRequest, "auth_index is required")
		return
	}
	auth := h.authByIndex(authIndex)
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), constant.Floatboat) {
		writeQuotaError(c, http.StatusNotFound, "floatboat credential not found")
		return
	}
	baseURL, apiKey := floatboatQuotaCredential(auth)
	if apiKey == "" {
		writeQuotaError(c, http.StatusBadRequest, "floatboat credential has no inference key")
		return
	}

	client := floatboatauth.NewClientWithProxyURL(h.cfg, auth.ProxyURL, baseURL)
	billing, errBilling := client.FetchBilling(c.Request.Context(), baseURL, apiKey)
	if errBilling != nil {
		writeQuotaError(c, http.StatusBadGateway, errBilling.Error())
		return
	}

	quota := FloatboatQuota{
		Currency:              billing.Currency,
		SoftLimit:             billing.SoftLimitUSD,
		HardLimit:             billing.HardLimitUSD,
		Used:                  billing.TotalUsageUSD,
		HasPaymentMethod:      billing.HasPaymentMethod,
		HasActiveSubscription: floatboatQuotaBool(auth, "has_active_subscription"),
	}
	if remaining, ok := billing.Remaining(); ok {
		quota.Remaining = remaining
		quota.RemainingKnown = true
	}
	// Entitlement groups come from the same catalogue read the model probe uses;
	// they explain which plan the key routes against.
	if catalogue, errCatalogue := client.FetchCatalogue(c.Request.Context(), baseURL, apiKey); errCatalogue == nil {
		quota.Groups = catalogue.Groups
	}

	c.JSON(http.StatusOK, quota)
}

// floatboatQuotaCredential resolves the gateway origin and inference key that
// authorize the billing reads.
func floatboatQuotaCredential(auth *coreauth.Auth) (baseURL, apiKey string) {
	if auth == nil {
		return floatboatauth.DefaultInferenceBaseURL, ""
	}
	if auth.Attributes != nil {
		baseURL = strings.TrimRight(strings.TrimSpace(auth.Attributes["base_url"]), "/")
		if baseURL == "" {
			baseURL = strings.TrimRight(strings.TrimSpace(auth.Attributes["inference_base_url"]), "/")
		}
		apiKey = strings.TrimSpace(auth.Attributes["api_key"])
		if apiKey == "" {
			apiKey = strings.TrimSpace(auth.Attributes["access_token"])
		}
	}
	if baseURL == "" {
		baseURL = floatboatauth.DefaultInferenceBaseURL
	}
	if apiKey == "" && auth.Metadata != nil {
		if key, ok := auth.Metadata["api_key"].(string); ok {
			apiKey = strings.TrimSpace(key)
		}
	}
	return baseURL, apiKey
}

// floatboatQuotaBool reads a boolean flag from credential metadata.
func floatboatQuotaBool(auth *coreauth.Auth, key string) bool {
	if auth == nil || auth.Metadata == nil {
		return false
	}
	value, _ := auth.Metadata[key].(bool)
	return value
}
