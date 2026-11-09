package checkin

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

// TestAdvertisedCapabilitiesAreAcceptedByConfigValidation pins the two provider
// lists together.
//
// The config allowlist cannot import this package without a cycle, so it is a
// separate copy — and it silently drifted: dumate and lobsterai-server were
// implemented, advertised, and listed in the management panel, then refused with
// "automation.checkin: unsupported provider" the moment an operator tried to
// enable them.
func TestAdvertisedCapabilitiesAreAcceptedByConfigValidation(t *testing.T) {
	for _, capability := range Capabilities() {
		cfg := config.AutoCheckinConfig{Providers: map[string]config.CheckinProviderConfig{
			capability.ID: {Enabled: true},
		}}
		if errValidate := cfg.Validate(); errValidate != nil {
			t.Fatalf("capability %q is advertised but rejected by config validation: %v", capability.ID, errValidate)
		}
	}
	// The allowlist must still reject a provider this build does not implement,
	// or a typo in a config file would enable nothing and say nothing.
	unknown := config.AutoCheckinConfig{Providers: map[string]config.CheckinProviderConfig{
		"definitely-not-a-provider": {Enabled: true},
	}}
	if errValidate := unknown.Validate(); errValidate == nil {
		t.Fatal("an unknown provider must be rejected")
	}
}

// TestCapabilitiesAndAdaptersMatch pins the advertised list to the adapter set:
// a capability without an adapter would schedule work that can never run, and an
// adapter without a capability could never be enabled by an operator.
func TestCapabilitiesAndAdaptersMatch(t *testing.T) {
	adapters := BuiltinAdapters()
	capabilities := Capabilities()
	if len(adapters) != len(capabilities) {
		t.Fatalf("adapters=%d capabilities=%d", len(adapters), len(capabilities))
	}
	seen := make(map[string]bool, len(capabilities))
	for i, capability := range capabilities {
		if got := adapters[i].Provider(); got != capability.ID {
			t.Fatalf("adapter[%d].Provider() = %q, want %q", i, got, capability.ID)
		}
		if capability.Name == "" {
			t.Fatalf("capability %q has no display name", capability.ID)
		}
		if seen[capability.ID] {
			t.Fatalf("capability %q listed twice", capability.ID)
		}
		seen[capability.ID] = true
		// Every adapter builds its own request, so a capability must not expose a
		// header template the operator can tamper with accidentally: the base list
		// always protects the credential-bearing headers.
		if len(capability.ProtectedHeaders) == 0 {
			t.Fatalf("capability %q protects no headers", capability.ID)
		}
	}
}
