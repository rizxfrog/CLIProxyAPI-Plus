package synthesizer

import (
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
)

func TestConfigSynthesizerCodeBuddyAIKey(t *testing.T) {
	auths, err := NewConfigSynthesizer().Synthesize(&SynthesisContext{
		Config: &config.Config{CodeBuddyAIKey: []config.CodeBuddyAIKey{{
			APIKey: "cbai-key", Prefix: "cba", ProxyURL: "http://proxy.example",
		}}},
		Now: time.Unix(1, 0), IDGenerator: NewStableIDGenerator(),
	})
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if len(auths) != 1 {
		t.Fatalf("auth count = %d, want 1", len(auths))
	}
	auth := auths[0]
	if auth.Provider != constant.CodeBuddyAI || auth.Prefix != "cba" {
		t.Fatalf("auth = %#v", auth)
	}
	if auth.Attributes["base_url"] != "https://www.codebuddy.ai/v2" {
		t.Fatalf("base_url = %q", auth.Attributes["base_url"])
	}
	if auth.Attributes["api_key"] != "cbai-key" {
		t.Fatalf("api_key = %q", auth.Attributes["api_key"])
	}
}

func TestConfigSynthesizerCodeBuddyAIKeyBaseURLOverride(t *testing.T) {
	auths, err := NewConfigSynthesizer().Synthesize(&SynthesisContext{
		Config: &config.Config{CodeBuddyAIKey: []config.CodeBuddyAIKey{{
			APIKey: "cbai-key", BaseURL: "https://staging-codebuddy.tencent.com/v2",
		}}},
		Now: time.Unix(1, 0), IDGenerator: NewStableIDGenerator(),
	})
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if len(auths) != 1 {
		t.Fatalf("auth count = %d, want 1", len(auths))
	}
	if got := auths[0].Attributes["base_url"]; got != "https://staging-codebuddy.tencent.com/v2" {
		t.Fatalf("base_url = %q", got)
	}
}
