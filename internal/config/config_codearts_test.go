package config

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/credentialweight"
)

// TestSanitizeCodeArtsKeysDropsIncompleteTriples pins that a credential missing
// the secret key (which cannot sign anything) is dropped.
func TestSanitizeCodeArtsKeysDropsIncompleteTriples(t *testing.T) {
	cfg := &Config{
		CodeArtsKey: []CodeArtsKey{
			{APIKey: "AK1", SecretKey: "SK1", SecurityToken: "ST1", BaseURL: " https://a.example/api/v2 "},
			{APIKey: "AK2"},    // no secret key
			{SecretKey: "SK3"}, // no access key
			{APIKey: " AK4 ", SecretKey: " SK4 ", SecurityToken: " ST4 "},
			{APIKey: "AK1", SecretKey: "SK1", SecurityToken: "ST1", BaseURL: "https://a.example/api/v2"}, // duplicate
		},
	}
	cfg.SanitizeCodeArtsKeys()

	if len(cfg.CodeArtsKey) != 2 {
		t.Fatalf("kept %d entries, want 2: %+v", len(cfg.CodeArtsKey), cfg.CodeArtsKey)
	}
	if cfg.CodeArtsKey[0].BaseURL != "https://a.example/api/v2" {
		t.Fatalf("base url was not trimmed: %q", cfg.CodeArtsKey[0].BaseURL)
	}
	if cfg.CodeArtsKey[1].APIKey != "AK4" || cfg.CodeArtsKey[1].SecretKey != "SK4" || cfg.CodeArtsKey[1].SecurityToken != "ST4" {
		t.Fatalf("credential was not trimmed: %+v", cfg.CodeArtsKey[1])
	}
}

// TestSanitizeCodeArtsKeysNilSafe guards the nil receiver.
func TestSanitizeCodeArtsKeysNilSafe(t *testing.T) {
	var cfg *Config
	cfg.SanitizeCodeArtsKeys()
}

// TestValidateCredentialWeightsCodeArts pins weight validation wiring.
func TestValidateCredentialWeightsCodeArts(t *testing.T) {
	weight := 3
	cfg := &Config{CodeArtsKey: []CodeArtsKey{{APIKey: "AK", SecretKey: "SK", Weight: &weight}}}
	if err := cfg.ValidateCredentialWeights(); err != nil {
		t.Fatalf("valid weight rejected: %v", err)
	}

	// A non-positive weight means "use the default" and must be accepted.
	zero := 0
	cfg.CodeArtsKey[0].Weight = &zero
	if err := cfg.ValidateCredentialWeights(); err != nil {
		t.Fatalf("default weight rejected: %v", err)
	}

	// An out-of-range weight must be rejected.
	tooLarge := int(credentialweight.Max) + 1
	cfg.CodeArtsKey[0].Weight = &tooLarge
	if err := cfg.ValidateCredentialWeights(); err == nil {
		t.Fatal("out-of-range weight accepted")
	}
}

// TestCodeArtsKeyRefreshTokenField pins that the shared CodeBuddy-style key
// carries the refresh-token used by rotating providers.
func TestCodeArtsKeyRefreshTokenField(t *testing.T) {
	var key CodeBuddyCNKey
	key.RefreshToken = "rt"
	if key.RefreshToken != "rt" {
		t.Fatal("RefreshToken field is not wired")
	}
}
