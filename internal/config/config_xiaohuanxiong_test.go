package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/credentialweight"
)

// TestXiaohuanxiongKeyAcceptsRefreshToken is a regression guard: the key type
// must expose refresh-token so automatic rotation is possible. An earlier
// iteration aliased the CodeBuddy key shape, which silently dropped the field.
func TestXiaohuanxiongKeyAcceptsRefreshToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
port: 8317
auth-dir: "` + dir + `/auths"
xiaohuanxiong-api-key:
  - api-key: "access-token-value"
    refresh-token: "refresh-token-value"
    weight: 2
    prefix: "xhx"
    base-url: "https://xiaohuanxiong.com/api/web/llm/v2"
    excluded-models:
      - "kimi-k3"
    models:
      - name: "glm-5-3"
        alias: "glm-5-3"
`
	if errWrite := os.WriteFile(path, []byte(content), 0o600); errWrite != nil {
		t.Fatalf("write config: %v", errWrite)
	}

	cfg, errLoad := LoadConfig(path)
	if errLoad != nil {
		t.Fatalf("LoadConfig: %v", errLoad)
	}
	if len(cfg.XiaohuanxiongKey) != 1 {
		t.Fatalf("expected 1 key entry, got %d (sanitizer may have dropped it)", len(cfg.XiaohuanxiongKey))
	}

	entry := cfg.XiaohuanxiongKey[0]
	if entry.APIKey != "access-token-value" {
		t.Fatalf("api-key = %q", entry.APIKey)
	}
	if entry.RefreshToken != "refresh-token-value" {
		t.Fatalf("refresh-token = %q, want refresh-token-value", entry.RefreshToken)
	}
	if entry.Weight == nil || *entry.Weight != 2 {
		t.Fatalf("weight = %v, want 2", entry.Weight)
	}
	if entry.Prefix != "xhx" {
		t.Fatalf("prefix = %q, want xhx", entry.Prefix)
	}
	if len(entry.ExcludedModels) != 1 || entry.ExcludedModels[0] != "kimi-k3" {
		t.Fatalf("excluded-models = %v", entry.ExcludedModels)
	}
	if len(entry.Models) != 1 || entry.Models[0].Name != "glm-5-3" {
		t.Fatalf("models = %v", entry.Models)
	}
}

// TestXiaohuanxiongKeyDropsEmptyToken matches the shared sanitizer contract.
func TestXiaohuanxiongKeyDropsEmptyToken(t *testing.T) {
	cfg := &Config{
		XiaohuanxiongKey: []XiaohuanxiongKey{
			{APIKey: "   "},
			{APIKey: "valid", RefreshToken: " rt "},
			{APIKey: "valid", RefreshToken: "dup"},
		},
	}
	cfg.SanitizeXiaohuanxiongKeys()
	if len(cfg.XiaohuanxiongKey) != 1 {
		t.Fatalf("expected 1 entry after sanitize, got %d", len(cfg.XiaohuanxiongKey))
	}
	if cfg.XiaohuanxiongKey[0].RefreshToken != "rt" {
		t.Fatalf("refresh token not trimmed: %q", cfg.XiaohuanxiongKey[0].RefreshToken)
	}
}

// TestXiaohuanxiongKeyWeightValidation verifies the key family is wired into the
// shared credential-weight validator.
//
// The contract is: non-positive weights are valid and normalize to zero (the
// credential is excluded from weighted routing), while values above the maximum
// are rejected. Asserting the over-max path proves the family is registered,
// because an unregistered family would never be checked at all.
func TestXiaohuanxiongKeyWeightValidation(t *testing.T) {
	overMax := int(credentialweight.Max + 1)
	cfg := &Config{
		XiaohuanxiongKey: []XiaohuanxiongKey{{APIKey: "tok", Weight: &overMax}},
	}
	errValidate := cfg.ValidateCredentialWeights()
	if errValidate == nil {
		t.Fatal("expected an over-max weight to be rejected for xiaohuanxiong-api-key")
	}

	// A normal weight must pass.
	normal := 3
	ok := &Config{XiaohuanxiongKey: []XiaohuanxiongKey{{APIKey: "tok", Weight: &normal}}}
	if errOK := ok.ValidateCredentialWeights(); errOK != nil {
		t.Fatalf("unexpected error for a valid weight: %v", errOK)
	}
}
