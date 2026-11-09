package checkin

import (
	headertemplate "github.com/router-for-me/CLIProxyAPI/v8/internal/checkin/template"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

// builtinProviders is the single source of truth for the daily check-in
// protocols included in this build.
//
// The advertised capability list and the adapter set both derive from it, so a
// provider can never be listed in the management panel without an
// implementation, or implemented without being listed.
//
// The configuration allowlist in internal/config is a separate copy on purpose:
// that package cannot import this one without a cycle, and a typo in a config
// file must still fail loudly. internal/checkin's own test pins the two
// together.
var builtinProviders = []struct {
	id    string
	name  string
	build func() Adapter
}{
	{"codearts", "CodeArts", NewCodeArtsAdapter},
	{"codebuddy-cn", "WorkBuddy (CodeBuddy CN)", NewCodeBuddyAdapter},
	{"dumate", "Dumate", NewDumateAdapter},
	{"lobsterai-server", "LobsterAI (NetEase/Youdao)", NewLobsterAIAdapter},
	{"trae", "Trae", NewTraeAdapter},
}

// Capabilities lists only the daily check-in protocols included in this build.
func Capabilities() []Capability {
	result := make([]Capability, 0, len(builtinProviders))
	for _, provider := range builtinProviders {
		defaults := (config.CheckinProviderConfig{}).Effective()
		defaults.Headers = map[string]string{}
		result = append(result, Capability{ID: provider.id, Name: provider.name, Defaults: defaults, ProtectedHeaders: headertemplate.ProtectedHeaders(provider.id)})
	}
	return result
}

// BuiltinAdapters returns one adapter per advertised capability, in the same
// order as Capabilities.
func BuiltinAdapters() []Adapter {
	result := make([]Adapter, 0, len(builtinProviders))
	for _, provider := range builtinProviders {
		result = append(result, provider.build())
	}
	return result
}
