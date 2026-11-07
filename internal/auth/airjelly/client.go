package airjelly

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	Provider  = "airjelly"
	Model     = "airjelly-lite"
	Gateway   = "https://airjelly-server-v2.up.railway.app"
	Supabase  = "https://cmxoaabrzxxchikydykw.supabase.co"
	Callback  = "http://localhost:19877"
	UserAgent = "AirJelly/2.0.10 (win32)"
)

type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

type Quota struct {
	Plan               string  `json:"plan"`
	Status             string  `json:"status"`
	DailyRemaining     float64 `json:"dailyRemaining"`
	DailyUsed          float64 `json:"dailyUsed"`
	DailyLimit         float64 `json:"dailyLimit"`
	MonthlyRemaining   float64 `json:"monthlyRemaining"`
	MonthlyLimit       float64 `json:"monthlyLimit"`
	MonthlyPeriodEnd   string  `json:"monthlyPeriodEnd"`
	PermanentRemaining float64 `json:"permanentRemaining"`
	TotalRemaining     float64 `json:"totalRemaining"`
}

func RandomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GoogleURL(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	v := url.Values{}
	v.Set("provider", "google")
	v.Set("redirect_to", Callback)
	v.Set("code_challenge", base64.RawURLEncoding.EncodeToString(sum[:]))
	v.Set("code_challenge_method", "s256")
	return Supabase + "/auth/v1/authorize?" + v.Encode()
}

func token(ctx context.Context, client *http.Client, grant string, payload any) (*Session, error) {
	key := strings.TrimSpace(os.Getenv("AIRJELLY_SUPABASE_ANON_KEY"))
	if key == "" {
		return nil, errors.New("AIRJELLY_SUPABASE_ANON_KEY is required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, Supabase+"/auth/v1/token?grant_type="+url.QueryEscape(grant), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", key)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("airjelly auth: HTTP %d: %s", resp.StatusCode, truncate(b))
	}
	var s Session
	if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.AccessToken == "" || s.RefreshToken == "" || s.User.ID == "" {
		return nil, errors.New("airjelly auth: incomplete session")
	}
	if s.ExpiresAt == 0 && s.ExpiresIn > 0 {
		s.ExpiresAt = time.Now().Unix() + s.ExpiresIn
	}
	return &s, nil
}

func LoginPassword(ctx context.Context, client *http.Client, email, password string) (*Session, error) {
	return token(ctx, client, "password", map[string]string{"email": email, "password": password})
}

func ExchangeCode(ctx context.Context, client *http.Client, code, verifier string) (*Session, error) {
	return token(ctx, client, "pkce", map[string]string{"auth_code": code, "code_verifier": verifier})
}

// SignUp follows the desktop client's email registration flow. Supabase may
// return a session immediately, or require the user to verify a mailed code.
func SignUp(ctx context.Context, client *http.Client, email, password string) (*Session, bool, error) {
	b, err := authRequest(ctx, client, "/auth/v1/signup", map[string]string{"email": email, "password": password})
	if err != nil {
		return nil, false, err
	}
	var result struct {
		Session
		ID         string            `json:"id"`
		Email      string            `json:"email"`
		Identities []json.RawMessage `json:"identities"`
		User       struct {
			ID         string            `json:"id"`
			Email      string            `json:"email"`
			Identities []json.RawMessage `json:"identities"`
		} `json:"user"`
	}
	if err = json.Unmarshal(b, &result); err != nil {
		return nil, false, err
	}
	// GoTrue returns a bare user when email confirmation is enabled. Some
	// deployments wrap it under "user"; accept both response shapes.
	if result.User.ID == "" {
		result.User.ID = result.ID
		result.User.Email = result.Email
		result.User.Identities = result.Identities
	}
	if result.User.ID == "" {
		return nil, false, errors.New("airjelly signup: missing user")
	}
	if result.User.Identities != nil && len(result.User.Identities) == 0 {
		// Supabase returns a synthetic user for an existing email. The desktop
		// client probes the supplied password to distinguish confirmed and
		// unconfirmed accounts, then resends a code for the latter.
		existing, loginErr := LoginPassword(ctx, client, email, password)
		if loginErr == nil {
			return existing, false, nil
		}
		if strings.Contains(strings.ToLower(loginErr.Error()), "email not confirmed") {
			if err = ResendSignUpOTP(ctx, client, email); err != nil {
				return nil, false, err
			}
			return nil, true, nil
		}
		return nil, false, errors.New("airjelly signup: email already registered; sign in instead")
	}
	if result.AccessToken == "" {
		return nil, true, nil
	}
	result.Session.User.ID = result.User.ID
	result.Session.User.Email = result.User.Email
	if result.ExpiresAt == 0 && result.ExpiresIn > 0 {
		result.ExpiresAt = time.Now().Unix() + result.ExpiresIn
	}
	if result.RefreshToken == "" {
		return nil, false, errors.New("airjelly signup: incomplete session")
	}
	return &result.Session, false, nil
}

func VerifySignUpOTP(ctx context.Context, client *http.Client, email, code string) (*Session, error) {
	b, err := authRequest(ctx, client, "/auth/v1/verify", map[string]string{"email": email, "token": code, "type": "signup"})
	if err != nil {
		return nil, err
	}
	var session Session
	if err = json.Unmarshal(b, &session); err != nil {
		return nil, err
	}
	if session.AccessToken == "" || session.RefreshToken == "" || session.User.ID == "" {
		return nil, errors.New("airjelly verification: incomplete session")
	}
	if session.ExpiresAt == 0 && session.ExpiresIn > 0 {
		session.ExpiresAt = time.Now().Unix() + session.ExpiresIn
	}
	return &session, nil
}

func ResendSignUpOTP(ctx context.Context, client *http.Client, email string) error {
	_, err := authRequest(ctx, client, "/auth/v1/resend", map[string]string{"email": email, "type": "signup"})
	return err
}

// EnsureProfile mirrors the desktop client's post-login profile initialization.
// Supabase installations with a database trigger will already have the row.
func EnsureProfile(ctx context.Context, client *http.Client, session *Session) error {
	if session == nil || session.User.ID == "" || session.AccessToken == "" {
		return errors.New("airjelly profile: missing session")
	}
	key := strings.TrimSpace(os.Getenv("AIRJELLY_SUPABASE_ANON_KEY"))
	if key == "" {
		return errors.New("AIRJELLY_SUPABASE_ANON_KEY is required")
	}
	endpoint := Supabase + "/rest/v1/user_profiles?id=eq." + url.QueryEscape(session.User.ID) + "&select=id"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	b, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	_ = resp.Body.Close()
	if readErr != nil {
		return readErr
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("airjelly profile: HTTP %d: %s", resp.StatusCode, truncate(b))
	}
	var rows []json.RawMessage
	if err = json.Unmarshal(b, &rows); err != nil {
		return err
	}
	if len(rows) > 0 {
		return nil
	}
	name := strings.SplitN(session.User.Email, "@", 2)[0]
	if name == "" {
		name = "User"
	}
	payload, err := json.Marshal(map[string]string{"id": session.User.ID, "display_name": name})
	if err != nil {
		return err
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, Supabase+"/rest/v1/user_profiles", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		b, _ = io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("airjelly profile: HTTP %d: %s", resp.StatusCode, truncate(b))
	}
	return nil
}

func authRequest(ctx context.Context, client *http.Client, path string, payload any) ([]byte, error) {
	key := strings.TrimSpace(os.Getenv("AIRJELLY_SUPABASE_ANON_KEY"))
	if key == "" {
		return nil, errors.New("AIRJELLY_SUPABASE_ANON_KEY is required")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, Supabase+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", key)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("airjelly auth: HTTP %d: %s", resp.StatusCode, truncate(body))
	}
	return body, nil
}

func Refresh(ctx context.Context, client *http.Client, refreshToken string) (*Session, error) {
	return token(ctx, client, "refresh_token", map[string]string{"refresh_token": refreshToken})
}

// SignRequest signs the exact upstream method/path. The gateway key is supplied
// at runtime; it is never persisted in an auth file or committed to source.
func SignRequest(req *http.Request, accessToken string) error {
	if req == nil || req.URL == nil {
		return errors.New("airjelly: nil request")
	}
	if req.URL.Scheme != "https" || !strings.EqualFold(req.URL.Host, "airjelly-server-v2.up.railway.app") {
		return errors.New("airjelly: unexpected gateway origin")
	}
	keyString := strings.TrimSpace(os.Getenv("AIRJELLY_GATEWAY_HMAC_KEY"))
	if keyString == "" {
		return errors.New("AIRJELLY_GATEWAY_HMAC_KEY is required")
	}
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return errors.New("AIRJELLY_GATEWAY_HMAC_KEY must be hex")
	}
	if len(key) < 16 {
		return errors.New("AIRJELLY_GATEWAY_HMAC_KEY is too short")
	}
	if accessToken == "" {
		return errors.New("airjelly: missing access token")
	}
	nonceBytes := make([]byte, 16)
	_, err = rand.Read(nonceBytes)
	if err != nil {
		return err
	}
	nonceBytes[6] = (nonceBytes[6] & 0x0f) | 0x40
	nonceBytes[8] = (nonceBytes[8] & 0x3f) | 0x80
	nonce := fmt.Sprintf("%x-%x-%x-%x-%x", nonceBytes[:4], nonceBytes[4:6], nonceBytes[6:8], nonceBytes[8:10], nonceBytes[10:])
	ts := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(ts + nonce + strings.ToUpper(req.Method) + req.URL.Path))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-AirJelly-Timestamp", ts)
	req.Header.Set("X-AirJelly-Nonce", nonce)
	req.Header.Set("X-AirJelly-Signature", hex.EncodeToString(mac.Sum(nil)))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("x-ai-feature", "user")
	return nil
}

func ReadQuota(ctx context.Context, client *http.Client, accessToken string) (*Quota, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, Gateway+"/api/stripe/subscription", nil)
	if err != nil {
		return nil, err
	}
	if err = SignRequest(req, accessToken); err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("airjelly quota: HTTP %d: %s", resp.StatusCode, truncate(b))
	}
	var envelope struct {
		Plan   string `json:"plan"`
		Status string `json:"status"`
		Quota  Quota  `json:"quota"`
	}
	if err = json.Unmarshal(b, &envelope); err != nil {
		return nil, err
	}
	envelope.Quota.Plan = envelope.Plan
	envelope.Quota.Status = envelope.Status
	return &envelope.Quota, nil
}

func truncate(b []byte) string {
	if len(b) > 512 {
		b = b[:512]
	}
	return string(b)
}
