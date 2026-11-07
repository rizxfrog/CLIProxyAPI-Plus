package dumate

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// The desktop client keeps the Baidu console session in its own store:
//
//	<data dir>/auth.json   {"cookies":"<base64>"} plus the account profiles
//	<data dir>/.cookie-key a raw 32-byte AES-256-GCM key
//
// and encrypts the cookie jar as base64(iv[12] || tag[16] || ciphertext), which
// is exactly the layout internal/auth/dumate decrypts here.
//
// The proxy reads that store directly instead of asking the operator to copy a
// live session cookie into a second file. The local client is already required
// for inference (the base URL is a loopback port), so this adds no new host
// coupling, and a re-login takes effect on the next read without re-exporting.
//
// A credential may still carry a console_cookie field to override the store,
// which is the path for a deployment that does not run the client locally.
const (
	clientDataDirName = "qianfan-desktop-app"
	authStoreFileName = "auth.json"
	cookieKeyFileName = ".cookie-key"

	consoleCookieKeyLength = 32
	consoleCookieIVLength  = 12
	consoleCookieTagLength = 16
)

// storedCookie is one entry of the client's decrypted cookie jar. The client
// stores Chromium-style cookie records rather than a raw header string.
type storedCookie struct {
	Name           string   `json:"name"`
	Value          string   `json:"value"`
	Domain         string   `json:"domain"`
	Path           string   `json:"path"`
	Expires        *float64 `json:"expires"`
	ExpirationDate *float64 `json:"expirationDate"`
}

type storedAuthStore struct {
	Cookies string `json:"cookies"`
}

// clientDataDirs lists the per-platform locations of the client data directory,
// in preference order. The home-relative Windows path is included because the
// preferred form depends on APPDATA, which a service or container process may
// not inherit.
func clientDataDirs() []string {
	dirs := make([]string, 0, 2)
	if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
		dirs = append(dirs, filepath.Join(appData, clientDataDirName))
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return dirs
	}
	switch runtime.GOOS {
	case "darwin":
		dirs = append(dirs, filepath.Join(home, "Library", "Application Support", clientDataDirName))
	case "windows":
		dirs = append(dirs, filepath.Join(home, "AppData", "Roaming", clientDataDirName))
	default:
		dirs = append(dirs, filepath.Join(home, ".config", clientDataDirName))
	}
	return dirs
}

// ResolveConsoleCookie returns the console cookie for a credential.
//
// An explicit console_cookie in the credential wins; otherwise the value is read
// from the running client's store. A missing store is reported as an error so
// the caller can explain why the wallet is unreadable instead of showing zero.
func ResolveConsoleCookie(metadata map[string]any) (string, error) {
	if value := ConsoleCookie(metadata); value != "" {
		return value, nil
	}
	return ClientConsoleCookie()
}

// ClientConsoleCookie decrypts the console cookie from the client's own store.
func ClientConsoleCookie() (string, error) {
	var firstErr error
	for _, dir := range clientDataDirs() {
		cookie, err := readClientConsoleCookie(dir)
		if err == nil {
			return cookie, nil
		}
		if firstErr == nil {
			// Report the preferred location's failure, not the last one tried.
			firstErr = err
		}
	}
	if firstErr != nil {
		return "", firstErr
	}
	return "", fmt.Errorf("dumate: no %s data directory found", clientDataDirName)
}

func readClientConsoleCookie(dir string) (string, error) {
	key, err := os.ReadFile(filepath.Join(dir, cookieKeyFileName))
	if err != nil {
		return "", fmt.Errorf("dumate: read %s: %w", cookieKeyFileName, err)
	}
	if len(key) != consoleCookieKeyLength {
		return "", fmt.Errorf("dumate: %s is %d bytes, want %d", cookieKeyFileName, len(key), consoleCookieKeyLength)
	}
	raw, err := os.ReadFile(filepath.Join(dir, authStoreFileName))
	if err != nil {
		return "", fmt.Errorf("dumate: read %s: %w", authStoreFileName, err)
	}
	var store storedAuthStore
	if errUnmarshal := json.Unmarshal(raw, &store); errUnmarshal != nil {
		return "", fmt.Errorf("dumate: decode %s: %w", authStoreFileName, errUnmarshal)
	}
	if strings.TrimSpace(store.Cookies) == "" {
		return "", fmt.Errorf("dumate: %s has no cookie jar; sign in to the client", authStoreFileName)
	}
	plaintext, err := decryptCookieJar(store.Cookies, key)
	if err != nil {
		return "", err
	}
	header, err := consoleCookieHeader(plaintext, time.Now())
	if err != nil {
		return "", err
	}
	return header, nil
}

// decryptCookieJar opens base64(iv || tag || ciphertext) with the client key.
func decryptCookieJar(encoded string, key []byte) ([]byte, error) {
	blob, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("dumate: decode cookie jar: %w", err)
	}
	if len(blob) <= consoleCookieIVLength+consoleCookieTagLength {
		return nil, fmt.Errorf("dumate: cookie jar is truncated")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("dumate: cookie key: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("dumate: cookie cipher: %w", err)
	}
	iv := blob[:consoleCookieIVLength]
	tag := blob[consoleCookieIVLength : consoleCookieIVLength+consoleCookieTagLength]
	ciphertext := blob[consoleCookieIVLength+consoleCookieTagLength:]
	// The stored layout keeps the tag before the ciphertext, while crypto/cipher
	// expects it appended. Copy so the two slices never alias one buffer.
	sealed := make([]byte, 0, len(ciphertext)+len(tag))
	sealed = append(sealed, ciphertext...)
	sealed = append(sealed, tag...)
	plaintext, errOpen := gcm.Open(nil, iv, sealed, nil)
	if errOpen != nil {
		return nil, fmt.Errorf("dumate: decrypt cookie jar: %w", errOpen)
	}
	return plaintext, nil
}

// consoleCookieHeader turns the decrypted jar into a Cookie header value.
//
// Only unexpired Baidu cookies are sent: the console rejects a request that
// carries nothing useful, and the analytics hosts in the same jar (hm.baidu.com)
// are irrelevant to it. Ties on cookie name are preserved in order, matching
// what the desktop client's own HTTP stack sends.
func consoleCookieHeader(plaintext []byte, now time.Time) (string, error) {
	var jar []storedCookie
	if err := json.Unmarshal(plaintext, &jar); err != nil {
		return "", fmt.Errorf("dumate: decode cookie jar: %w", err)
	}
	nowSeconds := float64(now.Unix())
	parts := make([]string, 0, len(jar))
	for _, entry := range jar {
		name := strings.TrimSpace(entry.Name)
		if name == "" || entry.Value == "" {
			continue
		}
		if !isBaiduCookieDomain(entry.Domain) {
			continue
		}
		expiry := entry.ExpirationDate
		if expiry == nil {
			expiry = entry.Expires
		}
		if expiry != nil && *expiry > 0 && *expiry <= nowSeconds {
			continue
		}
		parts = append(parts, name+"="+entry.Value)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("dumate: no usable Baidu console cookie in the client store")
	}
	return strings.Join(parts, "; "), nil
}

func isBaiduCookieDomain(domain string) bool {
	normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(domain), "."))
	return normalized == "baidu.com" || strings.HasSuffix(normalized, ".baidu.com")
}
