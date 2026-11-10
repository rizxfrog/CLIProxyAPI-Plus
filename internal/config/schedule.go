package config

import "strings"

// ScheduleConfig configures the scheduled-request automation.
//
// Definitions live outside the configuration file: this only names the directory
// that holds them, one JSON document per task group. An empty Dir disables the
// feature, so the scheduler never starts and no file is read.
type ScheduleConfig struct {
	// Dir is the directory holding one JSON document per task group. A leading
	// "~" expands to the user home directory; a relative path resolves against
	// the working directory.
	Dir string `yaml:"dir,omitempty" json:"dir"`
}

// EffectiveDir returns the trimmed directory, or an empty string when the
// feature is not configured.
func (c ScheduleConfig) EffectiveDir() string {
	if c.Dir == "" {
		return ""
	}
	return strings.TrimSpace(c.Dir)
}

// Enabled reports whether a directory is configured.
func (c ScheduleConfig) Enabled() bool { return c.EffectiveDir() != "" }
