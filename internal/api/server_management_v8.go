package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/api/handlers/management"
	apimiddleware "github.com/router-for-me/CLIProxyAPI/v8/internal/api/middleware"
)

// registerManagementV8Routes defines the v8 contract independently of v0.
// Configuration paths mirror the v8 YAML tree; operational routes use its groups.
func (s *Server) registerManagementV8Routes() {
	const prefix = "/v8/management"
	antiBot := apimiddleware.ManagementAntiBotMiddleware(s.cfg.RemoteManagement.AntiBot, nil)
	s.engine.GET(prefix+"/oauth/callback", s.managementAvailabilityMiddleware(), antiBot, s.mgmt.GetOAuthCallback)
	s.engine.POST(prefix+"/oauth/callback", s.managementAvailabilityMiddleware(), antiBot, s.mgmt.PostOAuthCallback)

	v8 := s.engine.Group(prefix)
	v8.Use(s.managementAvailabilityMiddleware(), antiBot, s.mgmt.Middleware(), func(c *gin.Context) {
		c.Set(management.ConfigV8ContextKey, true)
	})
	v8.GET("/config", s.mgmt.ConfigV8)
	v8.PUT("/config", s.mgmt.ConfigV8)
	v8.PATCH("/config", s.mgmt.ConfigV8)
	v8.GET("/config.yaml", s.mgmt.ConfigV8)
	v8.PUT("/config.yaml", s.mgmt.ConfigV8)
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		v8.Handle(method, "/config/*path", s.mgmt.ConfigV8)
	}

	v8.GET("/server/latest-version", s.mgmt.GetLatestVersion)
	v8.POST("/requests/api-call", s.mgmt.APICall)
	v8.POST("/routing/cooldown/reset", s.mgmt.ResetQuota)
	v8.GET("/routing/model-definitions/:channel", s.mgmt.GetStaticModelDefinitions)

	// Provider model management. The management panel addresses the v8 contract
	// exclusively, so these must exist here and not only under v0.
	v8.GET("/provider-models", s.mgmt.GetProviderModels)
	v8.GET("/provider-models/:provider", s.mgmt.GetProviderModel)
	v8.PUT("/provider-models/:provider", s.mgmt.PutProviderModel)
	v8.DELETE("/provider-models/:provider", s.mgmt.DeleteProviderModel)

	// Per-provider quota reads. Same reasoning as above: the v8 contract serves
	// the panel, so each provider route is registered here as well as under v0.
	v8.GET("/floatboat-quota", s.mgmt.GetFloatboatQuota)
	v8.GET("/codearts-quota", s.mgmt.GetCodeArtsQuota)
	v8.GET("/qoder-cn-quota", s.mgmt.GetQoderCNQuota)
	v8.GET("/qoder-ai-quota", s.mgmt.GetQoderCNQuota)
	v8.GET("/minimax-quota", s.mgmt.GetMinimaxQuota)

	v8.GET("/observability/logs", s.mgmt.GetLogs)
	v8.DELETE("/observability/logs", s.mgmt.DeleteLogs)
	v8.GET("/observability/logs/errors", s.mgmt.GetRequestErrorLogs)
	v8.GET("/observability/logs/errors/:name", s.mgmt.DownloadRequestErrorLog)
	v8.GET("/observability/logs/requests/:id", s.mgmt.GetRequestLogByID)
	v8.GET("/observability/usage/api-keys", s.mgmt.GetAPIKeyUsage)
	v8.GET("/observability/usage/queue", s.mgmt.GetUsageQueue)

	v8.GET("/credentials", s.mgmt.ListAuthFiles)
	v8.POST("/credentials", s.mgmt.UploadAuthFile)
	v8.DELETE("/credentials", s.mgmt.DeleteAuthFile)
	v8.GET("/credentials/models", s.mgmt.GetAuthFileModels)
	v8.GET("/credentials/download", s.mgmt.DownloadAuthFile)
	v8.PATCH("/credentials/status", s.mgmt.PatchAuthFileStatus)
	v8.PATCH("/credentials/fields", s.mgmt.PatchAuthFileFields)
	v8.POST("/credentials/refresh", s.mgmt.RefreshAuthFiles)
	v8.POST("/oauth/import", s.mgmt.ImportOAuthV8)
	v8.GET("/oauth/auth-url", s.mgmt.StartOAuthV8)
	v8.GET("/oauth/status", s.mgmt.GetAuthStatus)
	v8.DELETE("/oauth/session", s.mgmt.CancelAuthSession)
	v8.POST("/trae-auth-callback", s.mgmt.PostTraeAuthCallback)
	v8.POST("/cline-auth-callback", s.mgmt.PostClineAuthCallback)
	v8.POST("/xiaohuanxiong-auth-callback", s.mgmt.PostXiaohuanxiongAuthCallback)
	v8.POST("/codearts-auth-callback", s.mgmt.PostCodeArtsAuthCallback)

	v8.GET("/plugins", s.mgmt.ListPlugins)
	v8.DELETE("/plugins/:id", s.mgmt.DeletePlugin)
	v8.GET("/plugins/store", s.mgmt.ListPluginStore)
	v8.POST("/plugins/store/:id/install", s.mgmt.InstallPluginFromStore)
	v8.GET("/plugins/:id/quota", s.mgmt.GetPluginQuota)
	v8.POST("/plugins/:id/quota", s.mgmt.FetchPluginQuota)
	v8.DELETE("/plugins/:id/quota", s.mgmt.ResetPluginQuota)
}
