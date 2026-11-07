package management

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/airjelly"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	sdkauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
)

type airJellyPending struct {
	verifier string
	expires  time.Time
}

var airJellyFlows = struct {
	sync.Mutex
	pending map[string]airJellyPending
}{pending: make(map[string]airJellyPending)}

func (h *Handler) StartAirJellyOAuth(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	verifier, err := airjelly.RandomURLSafe(48)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not start OAuth"})
		return
	}
	state, err := airjelly.RandomURLSafe(24)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not start OAuth"})
		return
	}
	airJellyFlows.Lock()
	for id, pending := range airJellyFlows.pending {
		if time.Now().After(pending.expires) {
			delete(airJellyFlows.pending, id)
		}
	}
	airJellyFlows.pending[state] = airJellyPending{verifier: verifier, expires: time.Now().Add(10 * time.Minute)}
	airJellyFlows.Unlock()
	c.JSON(200, gin.H{"state": state, "url": airjelly.GoogleURL(verifier), "callback": airjelly.Callback})
}

func airJellyHTTPClient(ctx context.Context, h *Handler) *http.Client {
	return helps.NewProxyAwareHTTPClient(ctx, h.cfg, nil, 0)
}

func (h *Handler) CompleteAirJellyOAuth(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var body struct {
		State       string `json:"state"`
		CallbackURL string `json:"callback_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	parsed, err := url.Parse(strings.TrimSpace(body.CallbackURL))
	if err != nil || parsed.Hostname() != "localhost" || parsed.Port() != "19877" || parsed.Scheme != "http" {
		c.JSON(400, gin.H{"error": "expected the AirJelly localhost:19877 callback URL"})
		return
	}
	code := parsed.Query().Get("code")
	if code == "" {
		c.JSON(400, gin.H{"error": "callback has no authorization code"})
		return
	}
	airJellyFlows.Lock()
	flow, ok := airJellyFlows.pending[body.State]
	delete(airJellyFlows.pending, body.State)
	airJellyFlows.Unlock()
	if !ok || time.Now().After(flow.expires) {
		c.JSON(400, gin.H{"error": "OAuth flow expired; start again"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	session, err := airjelly.ExchangeCode(ctx, airJellyHTTPClient(ctx, h), code, flow.verifier)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	h.saveAirJellySession(c, session)
}

func (h *Handler) LoginAirJellyPassword(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	if strings.TrimSpace(body.Email) == "" || body.Password == "" {
		c.JSON(400, gin.H{"error": "email and password are required"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	session, err := airjelly.LoginPassword(ctx, airJellyHTTPClient(ctx, h), strings.TrimSpace(body.Email), body.Password)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	h.saveAirJellySession(c, session)
}

func (h *Handler) SignUpAirJelly(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Email) == "" || body.Password == "" {
		c.JSON(400, gin.H{"error": "email and password are required"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	session, needsConfirmation, err := airjelly.SignUp(ctx, airJellyHTTPClient(ctx, h), strings.TrimSpace(body.Email), body.Password)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	if needsConfirmation {
		c.JSON(200, gin.H{"status": "confirmation_required", "provider": airjelly.Provider})
		return
	}
	h.saveAirJellySession(c, session)
}

func (h *Handler) VerifyAirJellySignUp(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Email) == "" || strings.TrimSpace(body.Code) == "" {
		c.JSON(400, gin.H{"error": "email and verification code are required"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	session, err := airjelly.VerifySignUpOTP(ctx, airJellyHTTPClient(ctx, h), strings.TrimSpace(body.Email), strings.TrimSpace(body.Code))
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	h.saveAirJellySession(c, session)
}

func (h *Handler) ResendAirJellySignUpCode(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var body struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Email) == "" {
		c.JSON(400, gin.H{"error": "email is required"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	if err := airjelly.ResendSignUpOTP(ctx, airJellyHTTPClient(ctx, h), strings.TrimSpace(body.Email)); err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "confirmation_required", "provider": airjelly.Provider})
}

func (h *Handler) saveAirJellySession(c *gin.Context, session *airjelly.Session) {
	record := sdkauth.AirJellySessionRecord(session)
	if record == nil {
		c.JSON(502, gin.H{"error": "invalid session"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	profileErr := airjelly.EnsureProfile(ctx, airJellyHTTPClient(ctx, h), session)
	if _, err := h.saveTokenRecord(c.Request.Context(), record); err != nil {
		c.JSON(500, gin.H{"error": "could not save AirJelly authentication"})
		return
	}
	response := gin.H{"status": "ok", "provider": airjelly.Provider, "id": record.FileName, "label": record.Label}
	if profileErr != nil {
		response["profile_warning"] = "profile initialization needs a retry in the AirJelly client"
	}
	c.JSON(200, response)
}

func (h *Handler) GetAirJellyQuota(c *gin.Context) {
	index := strings.TrimSpace(c.Query("auth_index"))
	if index == "" {
		writeQuotaError(c, 400, "auth_index is required")
		return
	}
	auth := h.authByIndex(index)
	if auth == nil || auth.Provider != airjelly.Provider {
		writeQuotaError(c, 404, "airjelly credential not found")
		return
	}
	token := auth.Attributes["access_token"]
	if token == "" {
		if v, ok := auth.Metadata["access_token"].(string); ok {
			token = v
		}
	}
	if token == "" {
		writeQuotaError(c, 400, "airjelly credential has no access token")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	client := helps.NewProxyAwareHTTPClient(ctx, h.cfg, auth, 30*time.Second)
	quota, err := airjelly.ReadQuota(ctx, client, token)
	if err != nil {
		writeQuotaError(c, 502, err.Error())
		return
	}
	c.JSON(200, quota)
}
