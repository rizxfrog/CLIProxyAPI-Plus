package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/airjelly"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

type AirJellyAuthenticator struct{}

func NewAirJellyAuthenticator() Authenticator  { return AirJellyAuthenticator{} }
func (AirJellyAuthenticator) Provider() string { return airjelly.Provider }
func (AirJellyAuthenticator) RefreshLead() *time.Duration {
	d := 5 * time.Minute
	return &d
}

// Browser sign-in is handled by management's PKCE endpoints. The CLI can use a
// signed-in desktop session through the credential import endpoint.
func (AirJellyAuthenticator) Login(context.Context, *config.Config, *LoginOptions) (*coreauth.Auth, error) {
	return nil, errors.New("airjelly: sign in through the management OAuth page")
}

func AirJellySessionRecord(session *airjelly.Session) *coreauth.Auth {
	if session == nil {
		return nil
	}
	id := strings.TrimSpace(session.User.ID)
	digest := sha256.Sum256([]byte(id))
	name := fmt.Sprintf("airjelly-%s.json", hex.EncodeToString(digest[:16]))
	meta := map[string]any{
		"type":          airjelly.Provider,
		"auth_kind":     coreauth.AuthKindOAuth,
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"expired":       time.Unix(session.ExpiresAt, 0).UTC().Format(time.RFC3339),
		"userid":        id,
		"email":         session.User.Email,
		"base_url":      airjelly.Gateway,
		"prefix":        airjelly.Provider,
	}
	return &coreauth.Auth{
		ID: name, FileName: name, Provider: airjelly.Provider, Prefix: airjelly.Provider,
		Label:    session.User.Email,
		Metadata: meta,
		Attributes: map[string]string{
			coreauth.AttributeAuthKind: coreauth.AuthKindOAuth,
			"access_token":             session.AccessToken,
			"refresh_token":            session.RefreshToken,
			"base_url":                 airjelly.Gateway,
		},
	}
}
