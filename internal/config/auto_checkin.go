package config

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // Keep IANA timezones available in minimal deployments.

	headertemplate "github.com/router-for-me/CLIProxyAPI/v8/internal/checkin/template"
	"gopkg.in/yaml.v3"
)

// AutoCheckinConfig contains only explicit provider choices; omitted providers stay disabled.
type AutoCheckinConfig struct {
	Providers map[string]CheckinProviderConfig `yaml:"providers,omitempty" json:"providers"`
}

// CheckinProviderConfig configures a single built-in daily check-in adapter.
type CheckinProviderConfig struct {
	Enabled  bool              `yaml:"enabled,omitempty" json:"enabled"`
	Timezone string            `yaml:"timezone,omitempty" json:"timezone"`
	Window   CheckinWindow     `yaml:"window,omitempty" json:"window"`
	Headers  map[string]string `yaml:"headers,omitempty" json:"headers"`
}

// CheckinWindow is a same-calendar-day, half-open window in the configured timezone.
type CheckinWindow struct {
	Start string `yaml:"start,omitempty" json:"start"`
	End   string `yaml:"end,omitempty" json:"end"`
}

// Effective returns runtime defaults without modifying the stored configuration.
func (c CheckinProviderConfig) Effective() CheckinProviderConfig {
	c.Timezone = strings.TrimSpace(c.Timezone)
	if c.Timezone == "" {
		c.Timezone = "Asia/Shanghai"
	}
	c.Window.Start = normalizeCheckinTime(c.Window.Start, "00:10", false)
	c.Window.End = normalizeCheckinTime(c.Window.End, "03:10", true)
	headers := make(map[string]string, len(c.Headers))
	for name, value := range c.Headers {
		headers[http.CanonicalHeaderKey(name)] = value
	}
	c.Headers = headers
	return c
}

func normalizeCheckinTime(value, fallback string, allowEnd bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if minutes, err := checkinMinutes(value, allowEnd); err == nil {
		return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
	}
	return value
}

func checkinMinutes(value string, allowEnd bool) (int, error) {
	hour, minute, ok := strings.Cut(value, ":")
	if !ok || len(hour) < 1 || len(hour) > 2 || len(minute) != 2 {
		return 0, fmt.Errorf("time must use H:MM or HH:MM")
	}
	for _, part := range []string{hour, minute} {
		for _, c := range part {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("time must use H:MM or HH:MM")
			}
		}
	}
	h, _ := strconv.Atoi(hour)
	m, _ := strconv.Atoi(minute)
	if h == 24 && m == 0 && allowEnd {
		return 24 * 60, nil
	}
	if h > 23 || m > 59 {
		return 0, fmt.Errorf("invalid time")
	}
	return h*60 + m, nil
}

// Validate checks semantic settings and compiles templates without accessing credentials.
func (c AutoCheckinConfig) Validate() error {
	providers := make([]string, 0, len(c.Providers))
	for provider := range c.Providers {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	for _, provider := range providers {
		switch provider {
		case "codearts", "codebuddy-cn", "dumate", "lobsterai-server", "trae":
		default:
			return fmt.Errorf("automation.checkin: unsupported provider %q", provider)
		}
		// Validate raw names before canonicalization can merge case-insensitive duplicates.
		if _, err := headertemplate.Compile(c.Providers[provider].Headers, provider); err != nil {
			return fmt.Errorf("automation.checkin.providers.%s: %w", provider, err)
		}
		effective := c.Providers[provider].Effective()
		// Local is process-dependent, not an IANA timezone.
		if effective.Timezone == "Local" {
			return fmt.Errorf("automation.checkin.providers.%s: timezone must be an IANA name", provider)
		}
		if _, err := time.LoadLocation(effective.Timezone); err != nil {
			// Name the value: an operator who typed a lowercase or misspelled
			// zone otherwise only learns that *something* is invalid.
			return fmt.Errorf("automation.checkin.providers.%s: invalid timezone %q (expected an IANA name such as Asia/Shanghai)", provider, effective.Timezone)
		}
		start, err := checkinMinutes(effective.Window.Start, false)
		if err != nil {
			return fmt.Errorf("automation.checkin.providers.%s.window.start: %w", provider, err)
		}
		end, err := checkinMinutes(effective.Window.End, true)
		if err != nil {
			return fmt.Errorf("automation.checkin.providers.%s.window.end: %w", provider, err)
		}
		if start >= end {
			return fmt.Errorf("automation.checkin.providers.%s.window: start must precede end within the same day", provider)
		}
	}
	return nil
}

// UnmarshalYAML keeps dynamic provider/header maps while rejecting coercions and unknown fields.
func (c *AutoCheckinConfig) UnmarshalYAML(node *yaml.Node) error {
	if err := validateCheckinShape(node, "checkin"); err != nil {
		return err
	}
	type plain AutoCheckinConfig
	var decoded plain
	if err := node.Decode(&decoded); err != nil {
		return err
	}
	*c = AutoCheckinConfig(decoded)
	return nil
}

func validateCheckinShape(node *yaml.Node, kind string) error {
	if node.Tag == "!!null" {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("automation.checkin %s must be a mapping", kind)
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Tag != "!!str" {
			return fmt.Errorf("automation.checkin %s keys must be strings", kind)
		}
		child := ""
		switch kind {
		case "checkin":
			if key.Value == "providers" {
				child = "providers"
			}
		case "providers":
			child = "provider"
		case "provider":
			switch key.Value {
			case "enabled":
				child = "!!bool"
			case "timezone":
				child = "!!str"
			case "window":
				child = "window"
			case "headers":
				child = "headers"
			}
		case "window":
			if key.Value == "start" || key.Value == "end" {
				child = "!!str"
			}
		case "headers":
			child = "!!str"
		}
		if child == "" {
			return fmt.Errorf("automation.checkin: unknown %s field %q", kind, key.Value)
		}
		if strings.HasPrefix(child, "!!") {
			if value.Kind != yaml.ScalarNode || value.Tag != child {
				return fmt.Errorf("automation.checkin %s.%s has invalid type", kind, key.Value)
			}
		} else if err := validateCheckinShape(value, child); err != nil {
			return err
		}
	}
	return nil
}
