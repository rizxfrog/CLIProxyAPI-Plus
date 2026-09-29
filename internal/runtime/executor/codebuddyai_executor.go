package executor

import (
	"context"
	"net/http"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

// CodeBuddyAIBaseURL is the international CodeBuddy AI OpenAI-compatible gateway
// chat-completions endpoint. The executor derives its base URL from the
// synthesized auth attributes; this constant documents the canonical endpoint.
const CodeBuddyAIBaseURL = "https://www.codebuddy.ai/v2/chat/completions"

// CodeBuddyAIExecutor talks to the international CodeBuddy AI OpenAI-compatible
// gateway (https://www.codebuddy.ai).
//
// The international build shares the entire /v2 REST surface, request
// construction, authentication interceptor, and SSE parsing with the CodeBuddy
// CN gateway. Only the host and the account's authentication domain differ, so
// the executor reuses the same CodeBuddy behavior: streaming is forced at the
// executor boundary (the gateway rejects non-stream chat requests), reasoning
// effort is mirrored into reasoning_summary, and CLI agent system prompts are
// neutralized to avoid the upstream content filter.
type CodeBuddyAIExecutor struct {
	*OpenAICompatExecutor
}

// NewCodeBuddyAIExecutor constructs a CodeBuddy AI executor.
func NewCodeBuddyAIExecutor(cfg *config.Config) *CodeBuddyAIExecutor {
	e := NewOpenAICompatExecutor(constant.CodeBuddyAI, cfg)
	e.httpClientFactory = helps.NewCodeBuddyAIHTTPClient
	e.outgoingTransforms = applyCodeBuddyAIOutgoingTransforms
	return &CodeBuddyAIExecutor{
		OpenAICompatExecutor: e,
	}
}

// Identifier returns the executor identifier.
func (e *CodeBuddyAIExecutor) Identifier() string { return constant.CodeBuddyAI }

// Execute forces streaming upstream and aggregates the SSE chunks back into a
// single OpenAI JSON response for non-streaming clients.
func (e *CodeBuddyAIExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	return executeCodeBuddyAggregated(ctx, e.OpenAICompatExecutor, codeBuddyAIAuthDefaults, auth, req, opts)
}

// ExecuteStream forces streaming and delegates to the OpenAI-compatible executor.
func (e *CodeBuddyAIExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	return executeCodeBuddyStream(ctx, e.OpenAICompatExecutor, codeBuddyAIAuthDefaults, auth, req, opts)
}

// PrepareRequest injects CodeBuddy AI OAuth or API-key credentials into ad-hoc requests.
func (e *CodeBuddyAIExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	return e.OpenAICompatExecutor.PrepareRequest(req, prepareCodeBuddyAuth(auth, codeBuddyAIAuthDefaults))
}

// HttpRequest executes an ad-hoc CodeBuddy AI request with normalized credentials.
func (e *CodeBuddyAIExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	return e.OpenAICompatExecutor.HttpRequest(ctx, prepareCodeBuddyAuth(auth, codeBuddyAIAuthDefaults), req)
}

// Refresh rotates CodeBuddy AI OAuth credentials using the stored refresh token.
func (e *CodeBuddyAIExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	return refreshCodeBuddyAuth(ctx, e.OpenAICompatExecutor, codeBuddyAIAuthDefaults, auth)
}
