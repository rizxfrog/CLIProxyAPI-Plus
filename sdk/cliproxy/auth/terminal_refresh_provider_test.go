package auth

import (
	"errors"
	"fmt"
	"testing"
)

// vendorTerminalError stands in for a provider error type that declares its own
// terminal refresh failures (e.g. codearts.Error with STS5.1806).
type vendorTerminalError struct {
	terminal bool
}

func (e *vendorTerminalError) Error() string { return "vendor: refresh rejected" }

func (e *vendorTerminalError) IsTerminalRefreshError() bool { return e.terminal }

// TestTerminalRefreshErrorHonoursProviderDeclaration pins that the shared loop
// consults a provider's own terminality declaration.
//
// Without it, a vendor whose error wording matches none of the shared OAuth
// tokens (Huawei Cloud STS says "the refresh token has been used", not
// "refresh_token_reused") is retried forever after a spent token, so the account
// stays wedged on a retryable error instead of prompting for sign-in.
func TestTerminalRefreshErrorHonoursProviderDeclaration(t *testing.T) {
	auth := &Auth{Provider: "codearts"}

	if !isTerminalRefreshError(auth, &vendorTerminalError{terminal: true}) {
		t.Fatal("a provider-declared terminal error must be terminal")
	}
	if isTerminalRefreshError(auth, &vendorTerminalError{terminal: false}) {
		t.Fatal("a provider-declared retryable error must stay retryable")
	}
	// The declaration must survive wrapping, because executors wrap refresh
	// errors with context before handing them back.
	wrapped := fmt.Errorf("codearts: rotate credential: %w", &vendorTerminalError{terminal: true})
	if !isTerminalRefreshError(auth, wrapped) {
		t.Fatal("a wrapped provider-declared terminal error must be terminal")
	}
	if !errors.Is(wrapped, wrapped) {
		t.Fatal("sanity: errors.Is must see through the wrap")
	}
}
