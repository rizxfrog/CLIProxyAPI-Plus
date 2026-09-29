package synthesizer

import (
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
)

// TestConfigSynthesizerXiaohuanxiongKey verifies the auth record the executor
// depends on: provider key, default gateway URL, and a bearer api_key.
func TestConfigSynthesizerXiaohuanxiongKey(t *testing.T) {
	auths, err := NewConfigSynthesizer().Synthesize(&SynthesisContext{
		Config: &config.Config{XiaohuanxiongKey: []config.XiaohuanxiongKey{{
			APIKey: "access-token", Prefix: "xhx", ProxyURL: "http://proxy.example",
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
	if auth.Provider != constant.Xiaohuanxiong {
		t.Fatalf("provider = %q, want %q", auth.Provider, constant.Xiaohuanxiong)
	}
	if auth.Prefix != "xhx" {
		t.Fatalf("prefix = %q, want xhx", auth.Prefix)
	}
	if auth.Attributes["base_url"] != xiaohuanxiongDefaultBaseURL {
		t.Fatalf("base_url = %q, want %q", auth.Attributes["base_url"], xiaohuanxiongDefaultBaseURL)
	}
	if auth.Attributes["api_key"] != "access-token" {
		t.Fatalf("api_key = %q", auth.Attributes["api_key"])
	}
}

// TestConfigSynthesizerXiaohuanxiongRefreshTokenIsPropagated is a regression
// guard: without the refresh token the executor cannot rotate the access token,
// so a configured refresh-token must reach the auth metadata.
func TestConfigSynthesizerXiaohuanxiongRefreshTokenIsPropagated(t *testing.T) {
	auths, err := NewConfigSynthesizer().Synthesize(&SynthesisContext{
		Config: &config.Config{XiaohuanxiongKey: []config.XiaohuanxiongKey{{
			APIKey: "access-token", RefreshToken: "refresh-token",
		}}},
		Now: time.Unix(1, 0), IDGenerator: NewStableIDGenerator(),
	})
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if len(auths) != 1 {
		t.Fatalf("auth count = %d, want 1", len(auths))
	}
	got, _ := auths[0].Metadata["refresh_token"].(string)
	if got != "refresh-token" {
		t.Fatalf("metadata refresh_token = %q, want refresh-token (rotation would be impossible)", got)
	}
}

// TestConfigSynthesizerXiaohuanxiongBaseURLOverride allows self-hosted gateways.
func TestConfigSynthesizerXiaohuanxiongBaseURLOverride(t *testing.T) {
	auths, err := NewConfigSynthesizer().Synthesize(&SynthesisContext{
		Config: &config.Config{XiaohuanxiongKey: []config.XiaohuanxiongKey{{
			APIKey: "access-token", BaseURL: "https://self-hosted.example/api/web/llm/v2",
		}}},
		Now: time.Unix(1, 0), IDGenerator: NewStableIDGenerator(),
	})
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if len(auths) != 1 {
		t.Fatalf("auth count = %d, want 1", len(auths))
	}
	if got := auths[0].Attributes["base_url"]; got != "https://self-hosted.example/api/web/llm/v2" {
		t.Fatalf("base_url = %q", got)
	}
}
