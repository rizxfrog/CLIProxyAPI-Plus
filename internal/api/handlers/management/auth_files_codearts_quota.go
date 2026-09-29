package management

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	codeartsauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/codearts"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// CodeArtsQuota is the account credit balance returned by the developer
// gateway, shaped for the management console quota page.
type CodeArtsQuota struct {
	Channel           string `json:"channel"`
	TotalQuota        int64  `json:"total_quota"`
	TotalBalance      int64  `json:"total_balance"`
	UsedAmount        int64  `json:"used_amount"`
	DailyTokenLimit   int64  `json:"daily_token_limit"`
	DailyTokensUsed   int64  `json:"daily_tokens_used"`
	MonthlyTokenLimit int64  `json:"monthly_token_limit"`
	MonthlyTokensUsed int64  `json:"monthly_tokens_used"`
	ExpireTime        int64  `json:"expire_time"`
}

// GetCodeArtsQuota returns the signed-in account's credit balance for one
// CodeArts credential, identified by auth_index. The temporary AK/SK triple is
// rotated through STS first when it is inside the refresh window, because an
// expired security token cannot produce a valid signature.
func (h *Handler) GetCodeArtsQuota(c *gin.Context) {
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_index is required"})
		return
	}
	auth := h.authByIndex(authIndex)
	if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), "codearts") {
		c.JSON(http.StatusNotFound, gin.H{"error": "codearts credential not found"})
		return
	}

	creds, errCreds := h.codeartsFreshCredentials(c, auth)
	if errCreds != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": errCreds.Error()})
		return
	}

	client := newCodeArtsBalanceClient(h.cfg, auth.ProxyURL)
	balance, errBalance := client.FetchTokenBalance(c.Request.Context(), creds.AccessKey, creds.SecretKey, creds.SecurityToken)
	if errBalance != nil {
		status := http.StatusBadGateway
		if apiErr, ok := errBalance.(*codeartsauth.Error); ok && apiErr.StatusCode >= 400 {
			status = apiErr.StatusCode
		}
		c.JSON(status, gin.H{"error": errBalance.Error()})
		return
	}

	c.JSON(http.StatusOK, CodeArtsQuota{
		Channel:           balance.Channel,
		TotalQuota:        balance.TotalQuota,
		TotalBalance:      balance.TotalBalance,
		UsedAmount:        balance.UsedAmount,
		DailyTokenLimit:   balance.DailyTokenLimit,
		DailyTokensUsed:   balance.DailyTokensUsed,
		MonthlyTokenLimit: balance.MonthlyTokenLimit,
		MonthlyTokensUsed: balance.MonthlyTokensUsed,
		ExpireTime:        balance.ExpireTime,
	})
}

// codeartsFreshCredentials returns a credential triple that is valid for
// signing, rotating it through STS when the stored security token has entered
// the refresh window. The refreshed triple is persisted through the auth
// manager so subsequent inference calls reuse it.
func (h *Handler) codeartsFreshCredentials(c *gin.Context, auth *coreauth.Auth) (codeartsauth.Credentials, error) {
	read := func(a *coreauth.Auth) codeartsauth.Credentials {
		return codeartsauth.Credentials{
			AccessKey:     codeArtsMetaString(a, "access_key"),
			SecretKey:     codeArtsMetaString(a, "secret_key"),
			SecurityToken: codeArtsMetaString(a, "security_token"),
		}
	}

	if !codeartsCredentialExpired(auth) {
		creds := read(auth)
		if creds.AccessKey != "" && creds.SecretKey != "" {
			return creds, nil
		}
	}

	refreshToken := codeArtsMetaString(auth, "refresh_token")
	if refreshToken == "" {
		// Without a refresh token the stored triple is all we have; surface it
		// and let the upstream report expiry if it is already invalid.
		creds := read(auth)
		if creds.AccessKey == "" || creds.SecretKey == "" {
			return creds, errCodeArtsCredential("codearts credential is incomplete")
		}
		return creds, nil
	}
	privateKey := codeArtsMetaString(auth, "dpop_private_key")
	if privateKey == "" {
		return codeartsauth.Credentials{}, errCodeArtsCredential("codearts: cannot refresh without the original DPoP key pair")
	}
	keyPair := &codeartsauth.DpopKeyPair{PrivateKey: privateKey, PublicKey: codeArtsMetaString(auth, "dpop_public_key")}

	client := codeartsauth.NewClientWithProxyURL(h.cfg, auth.ProxyURL)
	token, errRefresh := client.Refresh(c.Request.Context(), refreshToken, codeArtsMetaString(auth, "code_verifier"), keyPair)
	if errRefresh != nil {
		return codeartsauth.Credentials{}, errRefresh
	}
	h.persistCodeArtsToken(c.Request.Context(), auth, token)
	return codeartsauth.Credentials{
		AccessKey:     token.AccessKey,
		SecretKey:     token.SecretKey,
		SecurityToken: token.SecurityToken,
	}, nil
}

// codeartsCredentialExpired reports whether the stored security token is inside
// the refresh window (or has no recorded expiry, in which case it is treated as
// still usable so the upstream stays authoritative).
func codeartsCredentialExpired(auth *coreauth.Auth) bool {
	expiredRaw := codeArtsMetaString(auth, "expired")
	if expiredRaw == "" {
		return false
	}
	expiry, errParse := time.Parse(time.RFC3339, expiredRaw)
	if errParse != nil || expiry.IsZero() {
		return false
	}
	return time.Now().Add(codeartsauth.RefreshWindowSeconds * time.Second).After(expiry)
}

// persistCodeArtsToken writes a refreshed triple back into the auth record and
// asks the auth manager to save it, mirroring what the executor does after a
// transparent refresh.
func (h *Handler) persistCodeArtsToken(ctx context.Context, auth *coreauth.Auth, token *codeartsauth.TokenData) {
	if auth == nil || token == nil {
		return
	}
	if auth.Metadata == nil {
		auth.Metadata = make(map[string]any)
	}
	auth.Metadata["access_key"] = token.AccessKey
	auth.Metadata["secret_key"] = token.SecretKey
	auth.Metadata["security_token"] = token.SecurityToken
	if token.RefreshToken != "" {
		auth.Metadata["refresh_token"] = token.RefreshToken
	}
	if !token.ExpiresAt.IsZero() {
		auth.Metadata["expired"] = token.ExpiresAt.UTC().Format(time.RFC3339)
	}
	auth.Metadata["last_refresh"] = time.Now().UTC().Format(time.RFC3339)

	if auth.Attributes == nil {
		auth.Attributes = make(map[string]string)
	}
	auth.Attributes["api_key"] = token.AccessKey
	auth.Attributes["secret_key"] = token.SecretKey
	auth.Attributes["security_token"] = token.SecurityToken

	if h.authManager != nil {
		if _, errUpdate := h.authManager.Update(ctx, auth); errUpdate != nil {
			// A persistence failure is not fatal for this read; the refreshed
			// triple is still returned to the caller.
			_ = errUpdate
		}
	}
}

// codeArtsMetaString reads a trimmed string value from auth metadata.
func codeArtsMetaString(auth *coreauth.Auth, key string) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	value, _ := auth.Metadata[key].(string)
	return strings.TrimSpace(value)
}

// errCodeArtsCredential is a simple credential error.
type errCodeArtsCredential string

func (e errCodeArtsCredential) Error() string { return string(e) }

// codeArtsBalanceFetcher abstracts the gateway call so tests can substitute a
// fake without touching the network.
type codeArtsBalanceFetcher interface {
	FetchTokenBalance(ctx context.Context, accessKey, secretKey, securityToken string) (*codeartsauth.TokenBalance, error)
}

var newCodeArtsBalanceClient = func(cfg *config.Config, proxyURL string) codeArtsBalanceFetcher {
	return codeartsauth.NewClientWithProxyURL(cfg, proxyURL)
}
