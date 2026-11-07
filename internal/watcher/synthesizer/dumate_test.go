package synthesizer

import (
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// TestFileSynthesizerDumateCredential pins the auth record the Dumate executor
// depends on: an API-key-shaped credential with the loopback base URL and the
// X-Dumate-* runtime headers built from the exported credential fields.
func TestFileSynthesizerDumateCredential(t *testing.T) {
	raw := `{
		"type": "dumate",
		"base_url": "http://127.0.0.1:52194/api/qianfanproxy/v1",
		"api_key": "nokey",
		"account_id": "91fc4f6327264f38a269db24427804e0",
		"device_id": "1|windows.deadbeef",
		"inapp_key": "3a7d79475dff98880ebce69d59450ee89c3c225742ed8d3b5bb712a4f3ee9e87"
	}`
	auths, err := SynthesizeAuthFile(&SynthesisContext{Config: &config.Config{}, Now: time.Unix(1, 0)}, "dumate.json", []byte(raw))
	if err != nil || len(auths) != 1 {
		t.Fatalf("SynthesizeAuthFile() = %d auths, %v", len(auths), err)
	}
	auth := auths[0]
	if auth.Provider != constant.Dumate {
		t.Fatalf("provider = %q, want %q", auth.Provider, constant.Dumate)
	}
	// The file synthesizer records auth_kind=oauth for every file credential;
	// Dumate does not have a refresh lead registered, so the auto-refresh loop
	// never schedules it.
	if got := auth.Attributes[coreauth.AttributeAuthKind]; got != coreauth.AuthKindOAuth {
		t.Fatalf("auth_kind = %q, want %q", got, coreauth.AuthKindOAuth)
	}
	if got := auth.Attributes["base_url"]; got != "http://127.0.0.1:52194/api/qianfanproxy/v1" {
		t.Fatalf("base_url = %q", got)
	}
	if got := auth.Attributes["api_key"]; got != "nokey" {
		t.Fatalf("api_key = %q", got)
	}
	for name, want := range map[string]string{
		"header:X-Dumate-Account-Id": "91fc4f6327264f38a269db24427804e0",
		"header:X-Dumate-Device-Id":  "1|windows.deadbeef",
		"header:X-Dumate-Inapp-Key":  "3a7d79475dff98880ebce69d59450ee89c3c225742ed8d3b5bb712a4f3ee9e87",
	} {
		if got := auth.Attributes[name]; got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

// TestFileSynthesizerDumatePortFallback verifies a credential may carry only the
// backend port instead of the full base URL.
func TestFileSynthesizerDumatePortFallback(t *testing.T) {
	raw := `{"type":"dumate","port":52194,"account_id":"acct"}`
	auths, err := SynthesizeAuthFile(&SynthesisContext{Config: &config.Config{}, Now: time.Unix(1, 0)}, "dumate.json", []byte(raw))
	if err != nil || len(auths) != 1 {
		t.Fatalf("SynthesizeAuthFile() = %d auths, %v", len(auths), err)
	}
	if got := auths[0].Attributes["base_url"]; got != "http://127.0.0.1:52194/api/qianfanproxy/v1" {
		t.Fatalf("base_url = %q", got)
	}
	// The placeholder bearer token is applied when the file omits one.
	if got := auths[0].Attributes["api_key"]; got != "nokey" {
		t.Fatalf("api_key = %q", got)
	}
}

// TestFileSynthesizerDumateExplicitHeadersWin verifies an operator-supplied
// headers map overrides the generated X-Dumate-* headers.
func TestFileSynthesizerDumateExplicitHeadersWin(t *testing.T) {
	raw := `{"type":"dumate","port":52194,"inapp_key":"generated","headers":{"X-Dumate-Inapp-Key":"override"}}`
	auths, err := SynthesizeAuthFile(&SynthesisContext{Config: &config.Config{}, Now: time.Unix(1, 0)}, "dumate.json", []byte(raw))
	if err != nil || len(auths) != 1 {
		t.Fatalf("SynthesizeAuthFile() = %d auths, %v", len(auths), err)
	}
	if got := auths[0].Attributes["header:X-Dumate-Inapp-Key"]; got != "override" {
		t.Fatalf("header:X-Dumate-Inapp-Key = %q, want override", got)
	}
}

// TestFileSynthesizerDumateWithoutEndpoint pins that a credential missing both
// base_url and port still loads; the executor then fails with a clear
// "missing provider baseURL" error instead of a silent empty success.
func TestFileSynthesizerDumateWithoutEndpoint(t *testing.T) {
	raw := `{"type":"dumate","account_id":"acct"}`
	auths, err := SynthesizeAuthFile(&SynthesisContext{Config: &config.Config{}, Now: time.Unix(1, 0)}, "dumate.json", []byte(raw))
	if err != nil || len(auths) != 1 {
		t.Fatalf("SynthesizeAuthFile() = %d auths, %v", len(auths), err)
	}
	if got := auths[0].Attributes["base_url"]; got != "" {
		t.Fatalf("base_url = %q, want empty", got)
	}
}
