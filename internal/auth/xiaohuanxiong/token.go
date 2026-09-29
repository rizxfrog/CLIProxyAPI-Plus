package xiaohuanxiong

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	log "github.com/sirupsen/logrus"
)

// TokenStorage persists Xiaohuanxiong credentials to the auth directory.
type TokenStorage struct {
	// AccessToken is the bearer token used for LLM and catalog requests.
	AccessToken string `json:"access_token"`
	// RefreshToken rotates the access token.
	RefreshToken string `json:"refresh_token,omitempty"`
	// Expired is the RFC3339 access-token expiry, derived from the JWT exp claim.
	Expired string `json:"expired,omitempty"`
	// OfficeIdentity is the optional office account identity.
	OfficeIdentity string `json:"office_identity,omitempty"`
	// OfficeOrgName is the optional organization display name.
	OfficeOrgName string `json:"office_org_name,omitempty"`
	// OfficeOrgRole is the optional organization role.
	OfficeOrgRole string `json:"office_org_role,omitempty"`
	// Type indicates the authentication provider type.
	Type string `json:"type"`

	// Metadata holds arbitrary key-value pairs injected via hooks. It is not
	// exported to JSON directly to allow flattening during serialization.
	Metadata map[string]any `json:"-"`
}

// SetMetadata injects metadata into the storage before saving.
func (ts *TokenStorage) SetMetadata(meta map[string]any) {
	ts.Metadata = meta
}

// SaveTokenToFile serializes the credentials to a JSON file.
func (ts *TokenStorage) SaveTokenToFile(authFilePath string) error {
	misc.LogSavingCredentials(authFilePath)
	ts.Type = Provider

	if err := os.MkdirAll(filepath.Dir(authFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	data, errMerge := misc.MergeMetadata(ts, ts.Metadata)
	if errMerge != nil {
		return fmt.Errorf("failed to merge metadata: %w", errMerge)
	}

	f, errCreate := os.Create(authFilePath)
	if errCreate != nil {
		return fmt.Errorf("failed to create token file: %v", errCreate)
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			log.Errorf("xiaohuanxiong token storage: close token file error: %v", errClose)
		}
	}()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if errEncode := encoder.Encode(data); errEncode != nil {
		return fmt.Errorf("failed to write token to file: %v", errEncode)
	}
	return nil
}

// IsExpired reports whether the access token is inside the refresh window.
//
// Xiaohuanxiong issues JWT access tokens, so expiry is read from the exp claim.
// When the token is opaque, or carries no exp, the token is treated as valid and
// the upstream surfaces the authoritative failure.
func (ts *TokenStorage) IsExpired() bool {
	if ts == nil {
		return true
	}
	expiry, ok := jwtExpiry(ts.AccessToken)
	if !ok {
		// Fall back to a recorded expiry when present.
		if strings.TrimSpace(ts.Expired) == "" {
			return false
		}
		parsed, errParse := time.Parse(time.RFC3339, strings.TrimSpace(ts.Expired))
		if errParse != nil {
			return true
		}
		expiry = parsed
	}
	return time.Now().Add(RefreshWindowSeconds * time.Second).After(expiry)
}

// NeedsRefresh reports whether the credentials can and should be refreshed.
func (ts *TokenStorage) NeedsRefresh() bool {
	if ts == nil {
		return false
	}
	if strings.TrimSpace(ts.RefreshToken) == "" {
		return false
	}
	return ts.IsExpired()
}

// jwtExpiry reads the exp claim from a JWT without verifying its signature.
// Xiaohuanxiong access tokens are JWTs; the signature is validated upstream.
func jwtExpiry(token string) (time.Time, bool) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return time.Time{}, false
	}
	payload := parts[1]
	// Restore base64url padding before decoding.
	if remainder := len(payload) % 4; remainder != 0 {
		payload += strings.Repeat("=", 4-remainder)
	}
	decoded, errDecode := base64.URLEncoding.DecodeString(payload)
	if errDecode != nil {
		// Some encoders emit standard base64 rather than base64url.
		decoded, errDecode = base64.StdEncoding.DecodeString(payload)
		if errDecode != nil {
			return time.Time{}, false
		}
	}
	var claims map[string]any
	if errUnmarshal := json.Unmarshal(decoded, &claims); errUnmarshal != nil {
		return time.Time{}, false
	}
	raw, ok := claims["exp"]
	if !ok {
		return time.Time{}, false
	}
	switch value := raw.(type) {
	case float64:
		if value <= 0 {
			return time.Time{}, false
		}
		return time.Unix(int64(value), 0).UTC(), true
	case json.Number:
		seconds, errInt := value.Int64()
		if errInt != nil || seconds <= 0 {
			return time.Time{}, false
		}
		return time.Unix(seconds, 0).UTC(), true
	default:
		return time.Time{}, false
	}
}

// ExpiryRFC3339 renders the access-token expiry, if the token carries one.
func ExpiryRFC3339(accessToken string) string {
	expiry, ok := jwtExpiry(accessToken)
	if !ok {
		return ""
	}
	return expiry.Format(time.RFC3339)
}
