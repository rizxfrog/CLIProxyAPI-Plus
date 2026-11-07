// Package dumate holds the credential shape and request-signing contract for
// the Baidu Dumate (百度搭子) desktop client.
//
// Dumate has no OAuth authorization server and no device/user token endpoint.
// The desktop client spawns a local Go backend (dumate-main-server) that binds
// to 127.0.0.1 on an ephemeral port and exposes the sign-in session over an
// OpenAI-compatible surface at
//
//	http://127.0.0.1:<port>/api/qianfanproxy/v1/chat/completions
//
// The backend authorizes each call with three headers plus a placeholder bearer
// token, and forwards the request to https://dumate-svc.baidu.com. The
// authoritative credential is therefore a locally exported JSON file, not a
// token this process can obtain or refresh:
//
//   - X-Dumate-Account-Id  signed-in account id (stable per account)
//   - X-Dumate-Device-Id   per-installation device fingerprint
//   - X-Dumate-Inapp-Key   per-process in-app key; regenerated on every client
//     start, so an exported credential file is invalidated by a client restart
//
// There is no refresh counterpart: an expired or rotated credential requires
// re-exporting the file from the running client.
package dumate

import (
	"strconv"
	"strings"
)

const (
	// InferencePathSuffix is the path prefix the local backend mounts the
	// OpenAI-compatible inference surface on. A configured base_url must
	// include it because the shared executor appends /chat/completions.
	InferencePathSuffix = "/api/qianfanproxy/v1"

	// PlaceholderAPIKey is the bearer token the desktop client itself sends.
	// Authorization is carried by the X-Dumate-* headers, but the backend still
	// requires a non-empty Authorization header.
	PlaceholderAPIKey = "nokey"

	// DefaultHost is the loopback address the local backend binds to.
	DefaultHost = "127.0.0.1"

	// HeaderAccountID carries the signed-in account id.
	HeaderAccountID = "X-Dumate-Account-Id"
	// HeaderDeviceID carries the per-installation device fingerprint.
	HeaderDeviceID = "X-Dumate-Device-Id"
	// HeaderInappKey carries the per-process in-app key.
	HeaderInappKey = "X-Dumate-Inapp-Key"
)

// Metadata keys accepted in an exported credential file.
const (
	FieldBaseURL   = "base_url"
	FieldAPIKey    = "api_key"
	FieldPort      = "port"
	FieldAccountID = "account_id"
	FieldDeviceID  = "device_id"
	FieldInappKey  = "inapp_key"
)

// MetadataString returns a trimmed string field from an auth metadata map.
func MetadataString(metadata map[string]any, key string) string {
	if len(metadata) == 0 {
		return ""
	}
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}

// MetadataPort renders a port value that may be a JSON number or string.
func MetadataPort(metadata map[string]any) string {
	if len(metadata) == 0 {
		return ""
	}
	switch value := metadata[FieldPort].(type) {
	case string:
		return strings.TrimSpace(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	default:
		return ""
	}
}

// ResolveBaseURL returns the inference base URL for an exported credential.
//
// base_url wins when present; otherwise a base URL is built from port. The
// returned value always includes InferencePathSuffix so the shared executor's
// /chat/completions append lands on the backend's OpenAI surface.
func ResolveBaseURL(metadata map[string]any) string {
	base := strings.TrimRight(MetadataString(metadata, FieldBaseURL), "/")
	if base != "" {
		if strings.HasSuffix(base, InferencePathSuffix) {
			return base
		}
		return base + InferencePathSuffix
	}
	port := MetadataPort(metadata)
	if port == "" {
		return ""
	}
	return "http://" + DefaultHost + ":" + port + InferencePathSuffix
}

// ResolveAPIKey returns the bearer token to send, defaulting to the placeholder
// the desktop client uses when the credential file omits one.
func ResolveAPIKey(metadata map[string]any) string {
	if key := MetadataString(metadata, FieldAPIKey); key != "" {
		return key
	}
	return PlaceholderAPIKey
}

// RuntimeHeaders builds the X-Dumate-* header set from the exported credential
// fields. Empty fields are omitted so a partially filled file fails upstream
// with the backend's own error instead of sending blank credentials.
func RuntimeHeaders(metadata map[string]any) map[string]string {
	headers := make(map[string]string, 3)
	if value := MetadataString(metadata, FieldAccountID); value != "" {
		headers[HeaderAccountID] = value
	}
	if value := MetadataString(metadata, FieldDeviceID); value != "" {
		headers[HeaderDeviceID] = value
	}
	if value := MetadataString(metadata, FieldInappKey); value != "" {
		headers[HeaderInappKey] = value
	}
	if len(headers) == 0 {
		return nil
	}
	return headers
}
