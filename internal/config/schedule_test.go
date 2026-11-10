package config

import (
	"strings"
	"testing"
)

// The v8 layout is what operators write and what the management API addresses:
// automation.schedule.dir. It must parse, and it must survive the layout
// normalisation that an explicit v8 write performs.
func TestScheduleV8LayoutParsesAndNormalizes(t *testing.T) {
	data := []byte(`automation:
  schedule:
    dir: ./schedules
`)
	if errValidate := ValidateV8Config(data); errValidate != nil {
		t.Fatalf("ValidateV8Config: %v", errValidate)
	}
	cfg, errParse := ParseConfigBytes(data)
	if errParse != nil {
		t.Fatalf("ParseConfigBytes: %v", errParse)
	}
	if cfg.Schedule.Dir != "./schedules" {
		t.Fatalf("Schedule.Dir = %q, want ./schedules", cfg.Schedule.Dir)
	}
	if !cfg.Schedule.Enabled() {
		t.Fatal("Schedule.Enabled() = false, want true")
	}

	normalized, _, errNormalize := NormalizeConfigLayout(data, true)
	if errNormalize != nil {
		t.Fatalf("NormalizeConfigLayout: %v", errNormalize)
	}
	if !strings.Contains(string(normalized), "automation:") {
		t.Fatalf("normalized layout lost the v8 grouping:\n%s", normalized)
	}
	if strings.Contains(string(normalized), "\nschedule:") {
		t.Fatalf("normalized layout kept a top-level schedule key:\n%s", normalized)
	}
	roundTripped, errRoundTrip := ParseConfigBytes(normalized)
	if errRoundTrip != nil {
		t.Fatalf("ParseConfigBytes(normalized): %v", errRoundTrip)
	}
	if roundTripped.Schedule.Dir != cfg.Schedule.Dir {
		t.Fatalf("round trip lost the directory: %q != %q", roundTripped.Schedule.Dir, cfg.Schedule.Dir)
	}
}

// The legacy top-level key is what the struct marshals to, so it must keep
// working for hand-edited files that predate the v8 layout.
func TestScheduleLegacyLayoutParses(t *testing.T) {
	cfg, errParse := ParseConfigBytes([]byte("schedule:\n  dir: /var/lib/schedules\n"))
	if errParse != nil {
		t.Fatalf("ParseConfigBytes: %v", errParse)
	}
	if cfg.Schedule.Dir != "/var/lib/schedules" {
		t.Fatalf("Schedule.Dir = %q, want /var/lib/schedules", cfg.Schedule.Dir)
	}
}

// An absent block must leave the feature off rather than defaulting to a
// directory nobody configured.
func TestScheduleAbsentMeansDisabled(t *testing.T) {
	cfg, errParse := ParseConfigBytes([]byte("port: 8317\n"))
	if errParse != nil {
		t.Fatalf("ParseConfigBytes: %v", errParse)
	}
	if cfg.Schedule.Enabled() {
		t.Fatalf("Schedule.Enabled() = true with dir %q, want false", cfg.Schedule.Dir)
	}
	if cfg.Schedule.EffectiveDir() != "" {
		t.Fatalf("EffectiveDir() = %q, want empty", cfg.Schedule.EffectiveDir())
	}
}
