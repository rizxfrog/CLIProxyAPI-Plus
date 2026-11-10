package schedule

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveDir normalizes the configured directory into an absolute path.
//
// A leading "~" expands to the user home directory and a relative path resolves
// against the working directory, matching how auth-dir and the kv-store path are
// treated. An empty value stays empty, which is how the feature stays disabled.
func ResolveDir(raw string) (string, error) {
	resolved := strings.TrimSpace(raw)
	if resolved == "" {
		return "", nil
	}
	if strings.HasPrefix(resolved, "~") {
		homeDir, errHome := os.UserHomeDir()
		if errHome != nil {
			return "", fmt.Errorf("schedule: resolve user home directory: %w", errHome)
		}
		remainder := strings.TrimLeft(strings.TrimPrefix(resolved, "~"), "/\\")
		if remainder == "" {
			resolved = filepath.Clean(homeDir)
		} else {
			normalized := strings.ReplaceAll(remainder, "\\", "/")
			resolved = filepath.Join(homeDir, filepath.FromSlash(normalized))
		}
	}
	resolved = filepath.Clean(resolved)
	if !filepath.IsAbs(resolved) {
		absPath, errAbs := filepath.Abs(resolved)
		if errAbs != nil {
			return "", fmt.Errorf("schedule: resolve absolute path: %w", errAbs)
		}
		resolved = filepath.Clean(absPath)
	}
	return resolved, nil
}
