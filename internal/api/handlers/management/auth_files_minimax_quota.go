package management

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	minimaxauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/minimax"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// MinimaxQuotaWindow is one rate-limit window on the console card.
//
// The MiniMax coding plan meters a short rolling window (5 hours) and a weekly
// window. RemainingPercent is the remaining allowance when the upstream reports
// a bounded window; Unlimited marks a window the plan does not cap.
type MinimaxQuotaWindow struct {
	// Kind is "five_hour" or "weekly".
	Kind string `json:"kind"`
	// RemainingPercent is 0-100, omitted when Unlimited or unknown.
	RemainingPercent *int `json:"remaining_percent,omitempty"`
	// ResetAtMs is the window reset instant (epoch ms), 0 when unknown.
	ResetAtMs int64 `json:"reset_at_ms,omitempty"`
	// Unlimited reports an upstream "no limit" window.
	Unlimited bool `json:"unlimited"`
}

// MinimaxQuota is the coding-plan quota of one MiniMax Code credential, shaped
// for the management console quota page.
//
// The payload is assembled from the signed matrix reads the native client uses
// (account identity, membership) plus the open-platform coding-plan probe. A
// credential without an active coding plan legitimately carries no windows and
// reports not_subscribed.
type MinimaxQuota struct {
	// Plan is the token-plan tier label, when the account is subscribed.
	Plan string `json:"plan,omitempty"`
	// HasTokenPlan reports an explicit subscription flag when upstream sends one.
	HasTokenPlan *bool `json:"has_token_plan,omitempty"`
	// NotSubscribed marks an explicit has_token_plan == false.
	NotSubscribed bool `json:"not_subscribed"`
	// ExpiresAtMs is the plan expiry (epoch ms), 0 when unknown.
	ExpiresAtMs int64 `json:"expires_at_ms,omitempty"`
	// CreditBalance is the account credit balance string, when reported.
	CreditBalance string `json:"credit_balance,omitempty"`
	// Windows are the coding-plan rate-limit windows (5-hour then weekly).
	Windows []MinimaxQuotaWindow `json:"windows"`
}

// GetMinimaxQuota returns the coding-plan quota for one MiniMax Code credential,
// identified by auth_index. A 401/403 from the gateway triggers a single refresh
// through the stored refresh token, and the rotated credential is persisted so
// subsequent inference calls reuse it.
func (h *Handler) GetMinimaxQuota(c *gin.Context) {
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		writeQuotaError(c, http.StatusBadRequest, "auth_index is required")
		return
	}
	auth := h.authByIndex(authIndex)
	provider := ""
	if auth != nil {
		provider = strings.ToLower(strings.TrimSpace(auth.Provider))
	}
	if provider != constant.Minimax && provider != constant.MinimaxCN {
		writeQuotaError(c, http.StatusNotFound, "minimax credential not found")
		return
	}
	accessToken := minimaxAccessToken(auth)
	if accessToken == "" {
		writeQuotaError(c, http.StatusBadRequest, "minimax credential has no access token")
		return
	}

	region := minimaxRegionFromAuth(provider, auth)
	client := minimaxauth.NewQuotaClient(h.cfg, region, auth.ProxyURL)
	if matrixOrigin, openOrigin := minimaxQuotaOriginOverride(region); matrixOrigin != "" && openOrigin != "" {
		client = client.WithOrigins(matrixOrigin, openOrigin)
	}

	quota, errFetch := h.fetchMinimaxQuota(c.Request.Context(), client, auth, accessToken)
	if errFetch != nil {
		statusCode := http.StatusBadGateway
		if minimaxauth.IsAuthError(errFetch) {
			statusCode = http.StatusUnauthorized
		}
		writeQuotaError(c, statusCode, errFetch.Error())
		return
	}
	c.JSON(http.StatusOK, buildMinimaxQuota(quota))
}

// fetchMinimaxQuota reads the quota, refreshing the credential and retrying once
// when the gateway rejects the stored token.
func (h *Handler) fetchMinimaxQuota(ctx context.Context, client *minimaxauth.Client, auth *coreauth.Auth, accessToken string) (*minimaxauth.AccountQuota, error) {
	quota, errFetch := client.FetchAccountQuota(ctx, accessToken)
	if errFetch == nil || !minimaxauth.IsAuthError(errFetch) {
		return quota, errFetch
	}
	refreshToken := minimaxMetaString(auth, "refresh_token")
	if refreshToken == "" {
		return nil, errFetch
	}
	token, errRefresh := client.Refresh(ctx, refreshToken)
	if errRefresh != nil {
		// Surface the original rejection: the stale token is the root cause.
		return nil, errFetch
	}
	h.persistMinimaxToken(ctx, auth, token)
	return client.FetchAccountQuota(ctx, token.AccessToken)
}

// minimaxQuotaOriginOverride returns a test-injected origin pair (matrix, open
// platform) for the region, or empty strings in production. It is a variable so
// tests can point the handler at a stub server.
var minimaxQuotaOriginOverride = func(minimaxauth.Region) (string, string) { return "", "" }

// buildMinimaxQuota maps the auth-layer snapshot to the console payload.
func buildMinimaxQuota(quota *minimaxauth.AccountQuota) MinimaxQuota {
	payload := MinimaxQuota{HasTokenPlan: quota.Membership.HasTokenPlan, NotSubscribed: quota.NotSubscribed}
	payload.Plan = strings.TrimSpace(quota.Membership.Tier)
	payload.ExpiresAtMs = quota.Membership.ExpiresAtMs
	payload.CreditBalance = strings.TrimSpace(quota.Membership.CreditBalance)
	if snapshot := quota.Snapshot; snapshot != nil {
		payload.Windows = append(payload.Windows, MinimaxQuotaWindow{
			Kind:             "five_hour",
			RemainingPercent: snapshot.FiveHour.RemainingPercent,
			ResetAtMs:        snapshot.FiveHour.ResetAtMs,
			Unlimited:        snapshot.FiveHour.Unlimited,
		})
		payload.Windows = append(payload.Windows, MinimaxQuotaWindow{
			Kind:             "weekly",
			RemainingPercent: snapshot.Weekly.RemainingPercent,
			ResetAtMs:        snapshot.Weekly.ResetAtMs,
			Unlimited:        snapshot.Weekly.Unlimited,
		})
	}
	return payload
}

// minimaxRegionFromAuth resolves the account region from the stored metadata,
// falling back to the provider key.
func minimaxRegionFromAuth(provider string, auth *coreauth.Auth) minimaxauth.Region {
	if auth != nil {
		if raw := minimaxMetaString(auth, "region"); raw != "" {
			return minimaxauth.NormalizeRegion(raw)
		}
		if auth.Attributes != nil {
			if raw := strings.TrimSpace(auth.Attributes["minimax_region"]); raw != "" {
				return minimaxauth.NormalizeRegion(raw)
			}
		}
	}
	if provider == constant.MinimaxCN {
		return minimaxauth.RegionCN
	}
	return minimaxauth.RegionEN
}

// minimaxAccessToken reads the bearer token from metadata, falling back to the
// api_key attribute the synthesizer seeds for inference.
func minimaxAccessToken(auth *coreauth.Auth) string {
	if token := minimaxMetaString(auth, "access_token"); token != "" {
		return token
	}
	if auth == nil {
		return ""
	}
	return strings.TrimSpace(auth.Attributes["api_key"])
}

// persistMinimaxToken writes a rotated token pair into the auth record and asks
// the auth manager to persist it.
func (h *Handler) persistMinimaxToken(ctx context.Context, auth *coreauth.Auth, token *minimaxauth.TokenData) {
	if auth == nil || token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["access_token"] = token.AccessToken
	if trimmed := strings.TrimSpace(token.RefreshToken); trimmed != "" {
		auth.Metadata["refresh_token"] = trimmed
	}
	if trimmed := strings.TrimSpace(token.TokenType); trimmed != "" {
		auth.Metadata["token_type"] = trimmed
	}
	if token.ExpiresIn > 0 {
		auth.Metadata["expires_in"] = token.ExpiresIn
	}
	if !token.ExpiresAt.IsZero() {
		auth.Metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)

	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes["api_key"] = token.AccessToken

	if h.authManager != nil {
		if _, errUpdate := h.authManager.Update(ctx, auth); errUpdate != nil {
			// A persistence failure is not fatal for this read; the rotated token
			// is still used for the retry.
			_ = errUpdate
		}
	}
}

// minimaxMetaString reads a trimmed string value from auth metadata.
func minimaxMetaString(auth *coreauth.Auth, key string) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	value, _ := auth.Metadata[key].(string)
	return strings.TrimSpace(value)
}
