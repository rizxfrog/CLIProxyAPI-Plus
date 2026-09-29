package codearts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	log "github.com/sirupsen/logrus"
)

// TokenStorage persists CodeArts credentials to the auth directory.
//
// Unlike a bearer-token provider, CodeArts needs the whole Huawei Cloud
// credential bundle plus the DPoP key pair: STS rejects a refresh whose proof of
// possession does not use the key that performed the original login.
type TokenStorage struct {
	// AccessKey is the temporary Huawei Cloud access key id.
	AccessKey string `json:"access_key,omitempty"`
	// SecretKey is the temporary Huawei Cloud secret access key.
	SecretKey string `json:"secret_key,omitempty"`
	// SecurityToken is the temporary Huawei Cloud security token.
	SecurityToken string `json:"security_token,omitempty"`
	// RefreshToken rotates the credential bundle.
	RefreshToken string `json:"refresh_token,omitempty"`
	// Expired is the RFC3339 expiry of the credential bundle.
	Expired string `json:"expired,omitempty"`
	// DpopPrivateKey is the private JWK used to sign DPoP proofs.
	DpopPrivateKey string `json:"dpop_private_key,omitempty"`
	// DpopPublicKey is the public JWK embedded in the DPoP proof header.
	DpopPublicKey string `json:"dpop_public_key,omitempty"`
	// CodeVerifier is the PKCE verifier reused on refresh.
	CodeVerifier string `json:"code_verifier,omitempty"`
	// UserID is the resolved account id.
	UserID string `json:"user_id,omitempty"`
	// UserName is the resolved account name.
	UserName string `json:"user_name,omitempty"`
	// DomainID is the resolved account (domain) id.
	DomainID string `json:"domain_id,omitempty"`
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

	if errMkdir := os.MkdirAll(filepath.Dir(authFilePath), 0o700); errMkdir != nil {
		return fmt.Errorf("codearts: create auth directory: %w", errMkdir)
	}

	data, errMerge := misc.MergeMetadata(ts, ts.Metadata)
	if errMerge != nil {
		return fmt.Errorf("codearts: merge metadata: %w", errMerge)
	}

	file, errCreate := os.Create(authFilePath)
	if errCreate != nil {
		return fmt.Errorf("codearts: create token file: %w", errCreate)
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			log.Errorf("codearts token storage: close token file error: %v", errClose)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if errEncode := encoder.Encode(data); errEncode != nil {
		return fmt.Errorf("codearts: write token file: %w", errEncode)
	}
	return nil
}

// IsExpired reports whether the credential bundle is inside the refresh window.
func (ts *TokenStorage) IsExpired() bool {
	if ts == nil {
		return true
	}
	expiry := ts.Expiry()
	if expiry.IsZero() {
		// Without a recorded expiry the upstream is authoritative.
		return false
	}
	return time.Now().Add(RefreshWindowSeconds * time.Second).After(expiry)
}

// NeedsRefresh reports whether the credentials can and should be rotated.
func (ts *TokenStorage) NeedsRefresh() bool {
	if ts == nil {
		return false
	}
	if strings.TrimSpace(ts.RefreshToken) == "" {
		return false
	}
	return ts.IsExpired()
}

// Expiry parses the recorded credential expiry.
func (ts *TokenStorage) Expiry() time.Time {
	if ts == nil {
		return time.Time{}
	}
	raw := strings.TrimSpace(ts.Expired)
	if raw == "" {
		return time.Time{}
	}
	parsed, errParse := time.Parse(time.RFC3339, raw)
	if errParse != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

// Credentials returns the signed-request credential triple.
func (ts *TokenStorage) Credentials() Credentials {
	if ts == nil {
		return Credentials{}
	}
	return Credentials{
		AccessKey:     strings.TrimSpace(ts.AccessKey),
		SecretKey:     strings.TrimSpace(ts.SecretKey),
		SecurityToken: strings.TrimSpace(ts.SecurityToken),
	}
}

// DpopKeyPair returns the stored DPoP key pair, or nil when absent.
func (ts *TokenStorage) DpopKeyPair() *DpopKeyPair {
	if ts == nil {
		return nil
	}
	if strings.TrimSpace(ts.DpopPrivateKey) == "" {
		return nil
	}
	return &DpopKeyPair{
		PrivateKey: strings.TrimSpace(ts.DpopPrivateKey),
		PublicKey:  strings.TrimSpace(ts.DpopPublicKey),
	}
}
