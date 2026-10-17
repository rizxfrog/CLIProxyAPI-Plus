package management

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func TestXiaohuanxiongQuotaTokenPrefersCurrentMetadata(t *testing.T) {
	auth := &coreauth.Auth{
		Metadata:   map[string]any{"access_token": "new-token"},
		Attributes: map[string]string{"api_key": "old-token"},
	}
	if got := xiaohuanxiongQuotaToken(auth); got != "new-token" {
		t.Fatalf("token = %q", got)
	}
}

func TestGetXiaohuanxiongQuotaRequiresAuthIndex(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v8/management/xiaohuanxiong-quota", nil)
	(&Handler{}).GetXiaohuanxiongQuota(ctx)
	if ctx.Writer.Status() != http.StatusBadRequest {
		t.Fatalf("status = %d", ctx.Writer.Status())
	}
}
