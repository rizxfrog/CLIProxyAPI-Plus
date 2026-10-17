package management

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	xiaohuanxiongauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/xiaohuanxiong"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// GetXiaohuanxiongQuota returns the current point balance for one credential.
func (h *Handler) GetXiaohuanxiongQuota(c *gin.Context) {
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		writeQuotaError(c, http.StatusBadRequest, "auth_index is required")
		return
	}
	auth := h.authByIndex(authIndex)
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), constant.Xiaohuanxiong) {
		writeQuotaError(c, http.StatusNotFound, "xiaohuanxiong credential not found")
		return
	}
	accessToken := xiaohuanxiongQuotaToken(auth)
	if accessToken == "" {
		writeQuotaError(c, http.StatusBadRequest, "xiaohuanxiong credential has no access token")
		return
	}
	// The upstream is fixed by the provider client. Credential metadata must not
	// be allowed to redirect a privileged management request to another host.
	client := xiaohuanxiongauth.NewClientWithProxyURL(h.cfg, auth.ProxyURL)
	balance, err := client.FetchPointsBalance(c.Request.Context(), accessToken)
	if upstreamErr, ok := err.(*xiaohuanxiongauth.Error); ok && upstreamErr.StatusCode == http.StatusUnauthorized {
		if fresh, errRefresh := h.refreshXiaohuanxiongQuotaToken(c.Request.Context(), client, auth); errRefresh == nil {
			balance, err = client.FetchPointsBalance(c.Request.Context(), fresh)
		}
	}
	if err != nil {
		status := http.StatusBadGateway
		if upstreamErr, ok := err.(*xiaohuanxiongauth.Error); ok &&
			(upstreamErr.StatusCode == http.StatusUnauthorized || upstreamErr.StatusCode == http.StatusForbidden) {
			status = upstreamErr.StatusCode
		}
		writeQuotaError(c, status, err.Error())
		return
	}
	c.JSON(http.StatusOK, balance)
}

// A quota read may be the first account operation after token expiry. Rotate
// the saved token once, then retry the balance read with the fresh credential.
func (h *Handler) refreshXiaohuanxiongQuotaToken(ctx context.Context, client *xiaohuanxiongauth.Client, auth *coreauth.Auth) (string, error) {
	refreshToken, _ := auth.Metadata["refresh_token"].(string)
	token, err := client.Refresh(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	updated := auth.Clone()
	if updated.Metadata == nil {
		updated.Metadata = make(map[string]any)
	}
	updated.Metadata["access_token"] = token.AccessToken
	if token.RefreshToken != "" {
		updated.Metadata["refresh_token"] = token.RefreshToken
	}
	if expiry := xiaohuanxiongauth.ExpiryRFC3339(token.AccessToken); expiry != "" {
		updated.Metadata["expired"] = expiry
	}
	updated.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)
	if updated.Attributes == nil {
		updated.Attributes = make(map[string]string)
	}
	updated.Attributes["api_key"] = token.AccessToken
	if h.authManager != nil {
		_, _ = h.authManager.Update(ctx, updated)
	}
	return token.AccessToken, nil
}

func xiaohuanxiongQuotaToken(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	if auth.Metadata != nil {
		if token, ok := auth.Metadata["access_token"].(string); ok && strings.TrimSpace(token) != "" {
			return strings.TrimSpace(token)
		}
	}
	return strings.TrimSpace(auth.Attributes["api_key"])
}
