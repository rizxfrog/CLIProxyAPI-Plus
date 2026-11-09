package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAutoCheckinV8RoundTripPreservesDynamicMaps(t *testing.T) {
	data := []byte(`automation:
  checkin:
    providers:
      codebuddy-cn:
        enabled: true
        window:
          start: "0:10"
          end: "24:00"
        headers:
          UID: '${meta.uid}'
          X-Mixed-Case: 'prefix-$${literal}-${meta.uid}'
      codearts: {}
      trae: {}
`)
	if err := ValidateV8Config(data); err != nil {
		t.Fatal(err)
	}
	before, err := ParseConfigBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	normalized, _, err := NormalizeConfigLayout(data, true)
	if err != nil {
		t.Fatal(err)
	}
	after, err := ParseConfigBytes(normalized)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.AutoCheckin, after.AutoCheckin) {
		t.Fatalf("migration changed dynamic maps: %s", normalized)
	}
	configPath := filepath.Join(t.TempDir(), "checkin.yaml")
	if err := os.WriteFile(configPath, normalized, 0600); err != nil {
		t.Fatal(err)
	}
	if err := SaveConfigPreserveComments(configPath, after); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateV8Config(saved); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.AutoCheckin, loaded.AutoCheckin) {
		t.Fatalf("save/load changed templates: %s", saved)
	}
	effective := loaded.AutoCheckin.Providers["codebuddy-cn"].Effective()
	if !effective.Enabled || effective.Timezone != "Asia/Shanghai" || effective.Window.Start != "00:10" || effective.Window.End != "24:00" {
		t.Fatalf("effective settings: %+v", effective)
	}
	stored := loaded.AutoCheckin.Providers["codebuddy-cn"]
	if stored.Timezone != "" || stored.Window.Start != "0:10" {
		t.Fatalf("effective defaults mutated stored configuration: %+v", stored)
	}
	clone := loaded.CloneForRuntime()
	cloneProvider := clone.AutoCheckin.Providers["codebuddy-cn"]
	cloneProvider.Headers["UID"] = "changed"
	clone.AutoCheckin.Providers["codebuddy-cn"] = cloneProvider
	if loaded.AutoCheckin.Providers["codebuddy-cn"].Headers["UID"] != "${meta.uid}" {
		t.Fatal("runtime snapshot shares header maps")
	}
	// A full provider replacement can remove a header, not merge stale keys back.
	replacement := loaded.AutoCheckin.Providers["codebuddy-cn"]
	replacement.Enabled = false
	delete(replacement.Headers, "UID")
	loaded.AutoCheckin.Providers["codebuddy-cn"] = replacement
	if err := SaveConfigPreserveComments(configPath, loaded); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := reloaded.AutoCheckin.Providers["codebuddy-cn"].Headers["UID"]; found {
		t.Fatal("deleted header survived save")
	}
	if reloaded.AutoCheckin.Providers["codebuddy-cn"].Enabled {
		t.Fatal("disabled provider re-enabled by save merge")
	}
}

func TestAutoCheckinRejectsInvalidConfiguration(t *testing.T) {
	cases := []struct{ name, body string }{
		{"unsupported provider", "unknown: {}"},
		{"wrong provider case", "Trae: {}"},
		{"bad timezone", "trae: {timezone: Mars/Unknown}"},
		{"local timezone", "trae: {timezone: Local}"},
		{"equal window", `trae: {window: {start: "12:00", end: "12:00"}}`},
		{"cross midnight", `trae: {window: {start: "23:00", end: "01:00"}}`},
		{"24 hour start", `trae: {window: {start: "24:00", end: "24:00"}}`},
		{"invalid minute", `trae: {window: {start: "01:60"}}`},
		{"unknown field", `trae: {script: run}`},
		{"boolean coercion", `trae: {enabled: "true"}`},
		{"numeric header", `trae: {headers: {UID: 123}}`},
		{"duplicate headers", `trae: {headers: {UID: one, uid: two}}`},
		{"protected header", `trae: {headers: {Cloud-IDE-JWT: secret}}`},
		{"invalid template", `trae: {headers: {UID: "${meta..uid}"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte("automation:\n  checkin:\n    providers:\n      " + tc.body + "\n")
			if err := ValidateV8Config(data); err == nil {
				t.Fatal("v8 accepted invalid configuration")
			}
			if _, err := ParseConfigBytes(data); err == nil {
				t.Fatal("parse accepted invalid configuration")
			}
			path := filepath.Join(t.TempDir(), "checkin.yaml")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadConfig(path); err == nil {
				t.Fatal("load accepted invalid configuration")
			}
		})
	}
}

func TestAutoCheckinOmittedDefaultsRemainOmitted(t *testing.T) {
	cfg, err := ParseConfigBytes([]byte("automation: {checkin: {providers: {trae: {}}}}\n"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := yaml.Marshal(cfg.AutoCheckin)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "timezone") || strings.Contains(string(data), "enabled") || strings.Contains(string(data), "window") {
		t.Fatalf("omitted defaults materialized: %s", data)
	}
	if cfg.AutoCheckin.Providers["trae"].Effective().Enabled {
		t.Fatal("provider enabled by default")
	}
	// Metadata requirements are per-account and must not fail configuration validation.
	cfg.AutoCheckin.Providers["trae"] = CheckinProviderConfig{Headers: map[string]string{"UID": "${meta.uid}"}}
	if err := cfg.AutoCheckin.Validate(); err != nil {
		t.Fatal(err)
	}
}

// TestAutoCheckinNamesTheOffendingTimezone pins that the timezone error carries
// the rejected value. A bare "invalid timezone" leaves the operator unable to
// tell which of several provider cards holds the bad value, and a lowercase or
// misspelled zone is the common case the panel cannot catch on its own.
func TestAutoCheckinNamesTheOffendingTimezone(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"unknown zone", "providers: {trae: {timezone: Mars/Unknown}}", `"Mars/Unknown"`},
		{"lowercase zone", "providers: {codearts: {timezone: asia/shanghai}}", `"asia/shanghai"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tc.body), &node); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			var cfg AutoCheckinConfig
			if err := node.Decode(&cfg); err != nil {
				t.Fatalf("decode: %v", err)
			}
			errValidate := cfg.Validate()
			if errValidate == nil {
				t.Fatal("expected a validation error")
			}
			message := errValidate.Error()
			if !strings.Contains(message, tc.want) {
				t.Fatalf("message does not name the rejected zone: %q", message)
			}
			if !strings.Contains(message, "Asia/Shanghai") {
				t.Fatalf("message does not show an example: %q", message)
			}
		})
	}
}
