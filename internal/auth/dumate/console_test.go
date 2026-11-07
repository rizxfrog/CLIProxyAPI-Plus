package dumate

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testCookie = "BDUSS=abc; bce-sessionid=def"

func consoleServer(t *testing.T, handler http.HandlerFunc) (*ConsoleClient, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewConsoleClientWithBaseURL(server.Client(), server.URL), server
}

func successEnvelope(t *testing.T, result any) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"logId": "log-1", "code": 0, "message": "", "success": true, "result": result,
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return string(raw)
}

func TestFetchPointsOverview(t *testing.T) {
	var seenPath, seenCookie, seenUA string
	client, _ := consoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		seenCookie = r.Header.Get("Cookie")
		seenUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(successEnvelope(t, map[string]any{
			"isSubscribed": true,
			"usedPoints":   "6.09",
			"totalPoints":  "5520.00",
			"modelThrottleInfo": map[string]any{
				"throttled": true, "reason": "daily_limit", "throttleType": "soft",
			},
			"subscription": []map[string]any{{
				"packageId": "pkg-1", "usedPoints": "6.09", "totalPoints": "5000.00",
				"startDate": 1794048985, "expireDate": 1794653785, "pointsExpireDate": 1794653785,
				"packageType": "plan_experience", "source": "event_bonus", "status": "active",
			}},
			"incremental": []map[string]any{{
				"packageId": "pkg-2", "usedPoints": "0.00", "totalPoints": "500.00",
				"startDate": 1793980800, "expireDate": 1796572799, "packageType": "grant_point",
				"source": "login_bonus", "status": "active",
			}},
		})))
	})

	overview, err := client.FetchPointsOverview(context.Background(), testCookie, "")
	if err != nil {
		t.Fatalf("FetchPointsOverview() error = %v", err)
	}
	if overview.TotalPoints != "5520.00" || overview.UsedPoints != "6.09" || !overview.IsSubscribed {
		t.Fatalf("overview = %+v", overview)
	}
	if len(overview.Subscription) != 1 || overview.Subscription[0].PointsExpireDate != 1794653785 {
		t.Fatalf("subscription = %+v", overview.Subscription)
	}
	if len(overview.Incremental) != 1 || overview.Incremental[0].Source != "login_bonus" {
		t.Fatalf("incremental = %+v", overview.Incremental)
	}
	if !overview.ModelThrottleInfo.Throttled || overview.ModelThrottleInfo.Reason != "daily_limit" {
		t.Fatalf("throttle = %+v", overview.ModelThrottleInfo)
	}
	// The client's own query contract: the default timezone is applied, and the
	// unclaimed daily bonus stays out of the displayed total.
	for _, want := range []string{
		"timezone=" + strings.ReplaceAll(ConsoleTimezoneDefault, "/", "%2F"),
		"clientType=desktop", "ignoreLoginBonus=true",
	} {
		if !strings.Contains(seenPath, want) {
			t.Fatalf("request uri %q missing %q", seenPath, want)
		}
	}
	if seenCookie != testCookie {
		t.Fatalf("cookie = %q", seenCookie)
	}
	if !strings.Contains(seenUA, "Electron") {
		t.Fatalf("user agent = %q", seenUA)
	}
}

func TestFetchPointsOverviewHonoursTimezone(t *testing.T) {
	var seenPath string
	client, _ := consoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		_, _ = w.Write([]byte(successEnvelope(t, map[string]any{"totalPoints": "1.00"})))
	})
	if _, err := client.FetchPointsOverview(context.Background(), testCookie, "Asia/Tokyo"); err != nil {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(seenPath, "timezone=Asia%2FTokyo") {
		t.Fatalf("request uri = %q", seenPath)
	}
}

func TestConsoleBusinessError(t *testing.T) {
	client, _ := consoleServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"logId":"log-9","code":1001,"message":"登录凭证已过期","success":false}`))
	})
	_, err := client.FetchPointsOverview(context.Background(), testCookie, "")
	var upstreamErr *ConsoleError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("error = %v, want *ConsoleError", err)
	}
	if upstreamErr.Code != 1001 || upstreamErr.LogID != "log-9" {
		t.Fatalf("console error = %+v", upstreamErr)
	}
	if !strings.Contains(upstreamErr.Error(), "登录凭证已过期") {
		t.Fatalf("message = %q", upstreamErr.Error())
	}
}

func TestConsoleUnauthorized(t *testing.T) {
	client, _ := consoleServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"code":0,"message":"unauthorized"}`))
	})
	_, err := client.FetchPointsOverview(context.Background(), testCookie, "")
	var upstreamErr *ConsoleError
	if !errors.As(err, &upstreamErr) || upstreamErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("error = %v", err)
	}
}

func TestConsoleRequiresCookie(t *testing.T) {
	client := NewConsoleClient(http.DefaultClient)
	if _, err := client.FetchLoginBonusInfo(context.Background(), "  "); err == nil {
		t.Fatal("expected an error for a missing cookie")
	}
}

func TestLoginBonusInfoAndClaim(t *testing.T) {
	var claimMethod, claimPath string
	client, _ := consoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case consoleLoginBonusInfo:
			_, _ = w.Write([]byte(successEnvelope(t, map[string]any{
				"totalTimes": 3, "totalPoints": 1500, "hasIssued": true,
				"signInDays": []string{"2026-11-05", "2026-11-06", "2026-11-07"},
			})))
		case consoleLoginBonusPath:
			claimMethod, claimPath = r.Method, r.URL.Path
			_, _ = w.Write([]byte(successEnvelope(t, true)))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	info, err := client.FetchLoginBonusInfo(context.Background(), testCookie)
	if err != nil {
		t.Fatalf("FetchLoginBonusInfo() error = %v", err)
	}
	if info.TotalTimes != 3 || info.TotalPoints != 1500 || !info.HasIssued {
		t.Fatalf("info = %+v", info)
	}
	if !info.SignedInOn("2026-11-07") || info.SignedInOn("2026-11-08") || info.SignedInOn("") {
		t.Fatalf("SignedInOn mismatch for %+v", info.SignInDays)
	}

	claimed, err := client.ClaimLoginBonus(context.Background(), testCookie)
	if err != nil || !claimed {
		t.Fatalf("ClaimLoginBonus() = %v, %v", claimed, err)
	}
	// The claim is a POST: a GET would be a different upstream contract.
	if claimMethod != http.MethodPost || claimPath != consoleLoginBonusPath {
		t.Fatalf("claim request = %s %s", claimMethod, claimPath)
	}
}

func TestClaimLoginBonusReportsRefusal(t *testing.T) {
	client, _ := consoleServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(successEnvelope(t, false)))
	})
	claimed, err := client.ClaimLoginBonus(context.Background(), testCookie)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if claimed {
		t.Fatal("a refused claim must not report true")
	}
}

// encryptJar writes the client's storage layout so the request-side decryption
// path is exercised against a value it can verify.
func encryptJar(t *testing.T, key, plaintext []byte) string {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}
	iv := make([]byte, consoleCookieIVLength)
	if _, errRead := rand.Read(iv); errRead != nil {
		t.Fatalf("iv: %v", errRead)
	}
	sealed := gcm.Seal(nil, iv, plaintext, nil)
	ciphertext, tag := sealed[:len(sealed)-consoleCookieTagLength], sealed[len(sealed)-consoleCookieTagLength:]
	blob := append(append(append([]byte{}, iv...), tag...), ciphertext...)
	return base64.StdEncoding.EncodeToString(blob)
}

func TestDecryptCookieJarRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte(`[{"name":"BDUSS","value":"abc","domain":".baidu.com"}]`)
	encoded := encryptJar(t, key, plaintext)

	got, err := decryptCookieJar(encoded, key)
	if err != nil {
		t.Fatalf("decryptCookieJar() error = %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("plaintext = %s", got)
	}
	if _, errWrong := decryptCookieJar(encoded, []byte("ffffffffffffffffffffffffffffffff")); errWrong == nil {
		t.Fatal("a wrong key must fail authentication")
	}
	if _, errShort := decryptCookieJar("AAAA", key); errShort == nil {
		t.Fatal("a truncated jar must be rejected")
	}
}

func TestConsoleCookieHeaderFiltersJar(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	expired := float64(now.Unix() - 60)
	future := float64(now.Unix() + 3600)
	jar := []storedCookie{
		{Name: "BDUSS", Value: "keep", Domain: ".baidu.com", Expires: &future},
		{Name: "bce-sessionid", Value: "keep2", Domain: ".bce.baidu.com"},
		{Name: "BAIDUID", Value: "stale", Domain: ".baidu.com", Expires: &expired},
		{Name: "pt_login_sig", Value: "other", Domain: ".ptlogin2.qq.com"},
		{Name: "ui", Value: "other", Domain: ".graph.qq.com"},
		{Name: "  ", Value: "nameless", Domain: ".baidu.com"},
		{Name: "EMPTY", Value: "", Domain: ".baidu.com"},
	}
	raw, err := json.Marshal(jar)
	if err != nil {
		t.Fatalf("marshal jar: %v", err)
	}
	header, err := consoleCookieHeader(raw, now)
	if err != nil {
		t.Fatalf("consoleCookieHeader() error = %v", err)
	}
	if header != "BDUSS=keep; bce-sessionid=keep2" {
		t.Fatalf("header = %q", header)
	}

	if _, errEmpty := consoleCookieHeader([]byte(`[]`), now); errEmpty == nil {
		t.Fatal("an empty jar must be reported, not sent as an empty header")
	}
}

func TestResolveConsoleCookiePrefersCredential(t *testing.T) {
	got, err := ResolveConsoleCookie(map[string]any{FieldConsoleCookie: "  explicit  "})
	if err != nil || got != "explicit" {
		t.Fatalf("ResolveConsoleCookie() = %q, %v", got, err)
	}
}

func TestConsoleCookieDomainMatching(t *testing.T) {
	for domain, want := range map[string]bool{
		".baidu.com":           true,
		"baidu.com":            true,
		".bce.baidu.com":       true,
		"login.bce.baidu.com":  true,
		".passport.baidu.com":  true,
		".hm.baidu.com":        true,
		".ptlogin2.qq.com":     false,
		".baidu.com.evil.test": false,
		"notbaidu.com":         false,
		"":                     false,
	} {
		if got := isBaiduCookieDomain(domain); got != want {
			t.Fatalf("isBaiduCookieDomain(%q) = %v, want %v", domain, got, want)
		}
	}
}
