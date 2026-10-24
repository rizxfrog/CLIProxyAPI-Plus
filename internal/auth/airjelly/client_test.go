package airjelly

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestSignRequest(t *testing.T) {
	key := strings.Repeat("ab", 32)
	t.Setenv("AIRJELLY_GATEWAY_HMAC_KEY", key)
	req, err := http.NewRequest("POST", Gateway+"/api/v1/llm/chat/completions", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = SignRequest(req, "test-token"); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer test-token" {
		t.Fatal("missing bearer header")
	}
	mac := hmac.New(sha256.New, mustHex(t, key))
	_, _ = mac.Write([]byte(req.Header.Get("X-AirJelly-Timestamp") + req.Header.Get("X-AirJelly-Nonce") + "POST" + req.URL.Path))
	if !hmac.Equal(mustHex(t, req.Header.Get("X-AirJelly-Signature")), mac.Sum(nil)) {
		t.Fatal("signature mismatch")
	}
}

func mustHex(t *testing.T, value string) []byte {
	t.Helper()
	b, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestRefreshUsesSupabaseGrant(t *testing.T) {
	t.Setenv("AIRJELLY_SUPABASE_ANON_KEY", "public-test-key")
	client := &http.Client{Transport: testTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/auth/v1/token" || req.URL.Query().Get("grant_type") != "refresh_token" || req.Header.Get("apikey") != "public-test-key" {
			t.Fatalf("unexpected refresh request: %s", req.URL)
		}
		body, _ := io.ReadAll(req.Body)
		if !strings.Contains(string(body), `"refresh_token":"old"`) {
			t.Fatal("refresh token omitted")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"new","refresh_token":"next","expires_in":3600,"user":{"id":"account"}}`)), Header: make(http.Header)}, nil
	})}
	s, err := Refresh(context.Background(), client, "old")
	if err != nil || s.AccessToken != "new" || s.RefreshToken != "next" || s.ExpiresAt <= 0 {
		t.Fatalf("refresh failed: %#v %v", s, err)
	}
}

func TestEmailRegistrationAndVerification(t *testing.T) {
	t.Setenv("AIRJELLY_SUPABASE_ANON_KEY", "public-test-key")
	var paths []string
	client := &http.Client{Transport: testTransport(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("apikey") != "public-test-key" {
			t.Fatal("missing Supabase key")
		}
		paths = append(paths, req.URL.Path)
		body, _ := io.ReadAll(req.Body)
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		var response string
		switch req.URL.Path {
		case "/auth/v1/signup":
			if payload["email"] != "new@example.com" || payload["password"] != "password" {
				t.Fatal("wrong signup payload")
			}
			response = `{"id":"user-1","email":"new@example.com","identities":[{"id":"identity-1"}]}`
		case "/auth/v1/verify":
			if payload["type"] != "signup" || payload["token"] != "123456" {
				t.Fatal("wrong verification payload")
			}
			response = `{"access_token":"access","refresh_token":"refresh","expires_in":3600,"user":{"id":"user-1","email":"new@example.com"}}`
		case "/auth/v1/resend":
			if payload["type"] != "signup" || payload["email"] != "new@example.com" {
				t.Fatal("wrong resend payload")
			}
			response = `{}`
		default:
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(response)), Header: make(http.Header)}, nil
	})}
	got, pending, err := SignUp(context.Background(), client, "new@example.com", "password")
	if err != nil || got != nil || !pending {
		t.Fatalf("expected pending signup, got %#v %t %v", got, pending, err)
	}
	if err = ResendSignUpOTP(context.Background(), client, "new@example.com"); err != nil {
		t.Fatal(err)
	}
	got, err = VerifySignUpOTP(context.Background(), client, "new@example.com", "123456")
	if err != nil || got.AccessToken != "access" || got.User.ID != "user-1" || got.ExpiresAt <= 0 {
		t.Fatalf("verification failed: %#v %v", got, err)
	}
	if strings.Join(paths, ",") != "/auth/v1/signup,/auth/v1/resend,/auth/v1/verify" {
		t.Fatalf("wrong route sequence: %v", paths)
	}
}

func TestEnsureProfileCreatesMissingRow(t *testing.T) {
	t.Setenv("AIRJELLY_SUPABASE_ANON_KEY", "public-test-key")
	var methods []string
	client := &http.Client{Transport: testTransport(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("apikey") != "public-test-key" || req.Header.Get("Authorization") != "Bearer access" {
			t.Fatal("profile request missing credentials")
		}
		methods = append(methods, req.Method)
		if req.Method == http.MethodGet {
			if req.URL.Query().Get("id") != "eq.user-1" {
				t.Fatal("wrong profile lookup")
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[]`)), Header: make(http.Header)}, nil
		}
		body, _ := io.ReadAll(req.Body)
		if !strings.Contains(string(body), `"display_name":"new"`) {
			t.Fatal("wrong profile payload")
		}
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(``)), Header: make(http.Header)}, nil
	})}
	session := &Session{AccessToken: "access"}
	session.User.ID, session.User.Email = "user-1", "new@example.com"
	if err := EnsureProfile(context.Background(), client, session); err != nil {
		t.Fatal(err)
	}
	if strings.Join(methods, ",") != "GET,POST" {
		t.Fatalf("wrong profile sequence: %v", methods)
	}
}

func TestSignUpExistingEmailUsesPasswordSession(t *testing.T) {
	t.Setenv("AIRJELLY_SUPABASE_ANON_KEY", "public-test-key")
	client := &http.Client{Transport: testTransport(func(req *http.Request) (*http.Response, error) {
		response := `{"user":{"id":"user-1","email":"existing@example.com","identities":[]}}`
		if req.URL.Path == "/auth/v1/token" {
			response = `{"access_token":"access","refresh_token":"refresh","expires_in":3600,"user":{"id":"user-1","email":"existing@example.com"}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(response)), Header: make(http.Header)}, nil
	})}
	session, pending, err := SignUp(context.Background(), client, "existing@example.com", "password")
	if err != nil || pending || session == nil || session.AccessToken != "access" {
		t.Fatalf("existing account should sign in: %#v %t %v", session, pending, err)
	}
}

// Runs only when explicitly given the local desktop session and a runtime key.
func TestLiveQuota(t *testing.T) {
	path := os.Getenv("AIRJELLY_TEST_SESSION_FILE")
	if path == "" || os.Getenv("AIRJELLY_GATEWAY_HMAC_KEY") == "" {
		t.Skip("live credentials not configured")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var container map[string]string
	if err = json.Unmarshal(b, &container); err != nil {
		t.Fatal(err)
	}
	var session Session
	for _, v := range container {
		if err = json.Unmarshal([]byte(v), &session); err != nil {
			t.Fatal(err)
		}
		break
	}
	if session.AccessToken == "" {
		t.Fatal("no access token")
	}
	q, err := ReadQuota(context.Background(), http.DefaultClient, session.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if q.DailyLimit <= 0 || q.TotalRemaining < 0 {
		t.Fatalf("unexpected quota values: %+v", q)
	}
	t.Logf("plan=%s daily=%.0f/%.0f total=%.0f", q.Plan, q.DailyRemaining, q.DailyLimit, q.TotalRemaining)
}
