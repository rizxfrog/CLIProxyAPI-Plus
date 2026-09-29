package synthesizer

import (
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// TestSynthesizeCodeArtsKeys pins the credential -> Auth mapping, including the
// persistence of the refresh token and the DPoP-independent fields.
func TestSynthesizeCodeArtsKeys(t *testing.T) {
	cfg := &config.Config{
		CodeArtsKey: []config.CodeArtsKey{
			{
				APIKey:        "AK1",
				SecretKey:     "SK1",
				SecurityToken: "ST1",
				RefreshToken:  "rt1",
				Prefix:        "ca",
				BaseURL:       "https://custom.example/api/v2",
				ProxyURL:      "socks5://proxy.example:1080",
				Priority:      7,
			},
			{
				// Missing secret key: must be skipped.
				APIKey: "AK2",
			},
			{
				APIKey:    "AK3",
				SecretKey: "SK3",
			},
		},
	}
	synth := NewConfigSynthesizer()
	out, errSynthesize := synth.Synthesize(&SynthesisContext{
		Config:      cfg,
		Now:         time.Now().UTC(),
		IDGenerator: NewStableIDGenerator(),
	})
	if errSynthesize != nil {
		t.Fatalf("synthesize: %v", errSynthesize)
	}

	var found []*coreauth.Auth
	for _, auth := range out {
		if auth.Provider == constant.CodeArts {
			found = append(found, auth)
		}
	}
	if len(found) != 2 {
		t.Fatalf("synthesized %d codearts auths, want 2", len(found))
	}

	first := found[0]
	if first.Attributes["api_key"] != "AK1" || first.Attributes["secret_key"] != "SK1" || first.Attributes["security_token"] != "ST1" {
		t.Fatalf("attributes = %+v", first.Attributes)
	}
	if first.Attributes["base_url"] != "https://custom.example/api/v2" {
		t.Fatalf("base_url = %q", first.Attributes["base_url"])
	}
	if first.Attributes["priority"] != "7" {
		t.Fatalf("priority = %q", first.Attributes["priority"])
	}
	if first.Prefix != "ca" {
		t.Fatalf("prefix = %q", first.Prefix)
	}
	if first.ProxyURL != "socks5://proxy.example:1080" {
		t.Fatalf("proxy = %q", first.ProxyURL)
	}
	if first.Metadata["refresh_token"] != "rt1" {
		t.Fatalf("refresh_token not preserved: %+v", first.Metadata)
	}

	// The second entry has no explicit base-url and must fall back to the default.
	second := found[1]
	if second.Attributes["api_key"] != "AK3" {
		t.Fatalf("unexpected second entry: %+v", second.Attributes)
	}
	if second.Attributes["base_url"] != codeArtsDefaultBaseURL {
		t.Fatalf("default base_url = %q, want %q", second.Attributes["base_url"], codeArtsDefaultBaseURL)
	}
	if _, hasRefresh := second.Metadata["refresh_token"]; hasRefresh {
		t.Fatal("an absent refresh token must not be recorded")
	}
}

// TestSynthesizeCodeArtsKeysSkipsEmpty guards the empty-config path.
func TestSynthesizeCodeArtsKeysSkipsEmpty(t *testing.T) {
	synth := NewConfigSynthesizer()
	out, errSynthesize := synth.Synthesize(&SynthesisContext{
		Config:      &config.Config{},
		Now:         time.Now().UTC(),
		IDGenerator: NewStableIDGenerator(),
	})
	if errSynthesize != nil {
		t.Fatalf("synthesize: %v", errSynthesize)
	}
	for _, auth := range out {
		if auth.Provider == constant.CodeArts {
			t.Fatalf("no codearts auth should be produced: %+v", auth)
		}
	}
}

// TestSynthesizeCodeArtsStableID pins that the generated id is deterministic for
// the same credential so restarts do not churn auth records.
func TestSynthesizeCodeArtsStableID(t *testing.T) {
	cfg := &config.Config{CodeArtsKey: []config.CodeArtsKey{{APIKey: "AK1", SecretKey: "SK1"}}}
	build := func() string {
		out, _ := NewConfigSynthesizer().Synthesize(&SynthesisContext{
			Config:      cfg,
			Now:         time.Now().UTC(),
			IDGenerator: NewStableIDGenerator(),
		})
		for _, auth := range out {
			if auth.Provider == constant.CodeArts {
				return auth.ID
			}
		}
		return ""
	}
	first := build()
	second := build()
	if first == "" || first != second {
		t.Fatalf("ids are not stable: %q vs %q", first, second)
	}
}
