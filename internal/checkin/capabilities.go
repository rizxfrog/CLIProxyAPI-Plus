package checkin

import (
	headertemplate "github.com/router-for-me/CLIProxyAPI/v8/internal/checkin/template"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

// Capabilities lists only the daily check-in protocols included in this build.
func Capabilities() []Capability {
	providers := []struct{ id, name string }{
		{"codearts", "CodeArts"},
		{"codebuddy-cn", "WorkBuddy (CodeBuddy CN)"},
		{"dumate", "Dumate"},
		{"lobsterai-server", "LobsterAI (NetEase/Youdao)"},
		{"trae", "Trae"},
	}
	result := make([]Capability, 0, len(providers))
	for _, provider := range providers {
		defaults := (config.CheckinProviderConfig{}).Effective()
		defaults.Headers = map[string]string{}
		result = append(result, Capability{ID: provider.id, Name: provider.name, Defaults: defaults, ProtectedHeaders: headertemplate.ProtectedHeaders(provider.id)})
	}
	return result
}
