package management

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/qwenweb"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// LoginQwenWeb acquires and persists a web session without keeping passwords or
// unrelated cookies. The management middleware supplies authentication.
func (h *Handler) LoginQwenWeb(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 70*1024)
	var input qwenweb.LoginInput
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(400, gin.H{"error": "invalid Web login request"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	client := qwenweb.Client{HTTP: helps.NewProxyAwareHTTPClient(ctx, h.cfg, nil, 0)}
	user, err := client.Login(ctx, input)
	input = qwenweb.LoginInput{}
	if err != nil {
		code := 502
		if status, ok := err.(interface{ StatusCode() int }); ok {
			code = status.StatusCode()
		}
		// Upstream 401 must not be confused with a management-key 401 by the UI.
		if code == 401 {
			code = http.StatusUnprocessableEntity
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	digest := sha256.Sum256([]byte(user.ID))
	name := "qwen-web-" + hex.EncodeToString(digest[:16]) + ".json"
	metadata := map[string]any{
		"type": qwenweb.Provider, "access_token": user.Token, "email": user.Email, "account_id": user.ID,
	}
	// The web frontend authenticates chat requests with the token Cookie, so it
	// is stored alongside the token. No password is ever persisted.
	if user.Cookie != "" {
		metadata["cookie"] = user.Cookie
	}
	record := &coreauth.Auth{ID: name, FileName: name, Provider: qwenweb.Provider, Label: "Qwen Web", Metadata: metadata}
	if _, errSave := h.saveTokenRecord(ctx, record); errSave != nil {
		c.JSON(500, gin.H{"error": "could not save Qwen Web authentication"})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "provider": qwenweb.Provider, "id": name})
}
