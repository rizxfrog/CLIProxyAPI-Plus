package executor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	lobsteraiserverauth "github.com/router-for-me/CLIProxyAPI/v8/internal/auth/lobsteraiserver"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// lobsteraiServerProxyBaseURL resolves the gateway base URL for the credential
// (base+"/api/proxy/v1"), falling back to the production default.
func lobsteraiServerProxyBaseURL(auth *cliproxyauth.Auth) string {
	if raw := lobsteraiServerAttributeString(auth, "base_url"); raw != "" {
		return strings.TrimRight(raw, "/")
	}
	return lobsteraiserverauth.DefaultProxyBaseURL
}

// lobsteraiServerBackendURL resolves the account backend origin (without the
// proxy path) for login/refresh calls.
func lobsteraiServerBackendURL(auth *cliproxyauth.Auth) string {
	if raw := lobsteraiServerAttributeString(auth, "backend_url"); raw != "" {
		return strings.TrimRight(raw, "/")
	}
	// Derive the backend origin from the proxy base URL when it carries the
	// documented proxy path.
	if base := lobsteraiServerProxyBaseURL(auth); strings.HasSuffix(base, lobsteraiserverauth.DefaultProxyBasePath) {
		return strings.TrimRight(strings.TrimSuffix(base, lobsteraiserverauth.DefaultProxyBasePath), "/")
	}
	return lobsteraiserverauth.DefaultBackendURL
}

// lobsteraiServerChatURL is the only inference surface the gateway implements.
// /messages answers every model with code 40300 ("不支持的模型") and /responses is
// 404, so Anthropic- and Responses-shaped upstream calls cannot work here.
func lobsteraiServerChatURL(auth *cliproxyauth.Auth) string {
	return lobsteraiServerProxyBaseURL(auth) + "/chat/completions"
}

// lobsteraiServerCredential returns the access token that authorizes the gateway.
func lobsteraiServerCredential(auth *cliproxyauth.Auth) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if token := strings.TrimSpace(auth.Attributes["access_token"]); token != "" {
			return token
		}
		if token := strings.TrimSpace(auth.Attributes["api_key"]); token != "" && !strings.HasPrefix(token, "sk-xxx") {
			return token
		}
	}
	return lobsteraiServerAttributeString(auth, "access_token")
}

// lobsteraiServerClientVersion resolves the X-LobsterAI-Client-Version value.
// The upstream treats an empty version as a legacy client, so the pinned client
// version is always sent.
func lobsteraiServerClientVersion(auth *cliproxyauth.Auth) string {
	if version := lobsteraiServerAttributeString(auth, "client_version"); version != "" {
		return version
	}
	return lobsteraiserverauth.ClientVersion
}

// lobsteraiServerAttributeString reads a trimmed string value from attributes,
// then metadata.
func lobsteraiServerAttributeString(auth *cliproxyauth.Auth, key string) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if value := strings.TrimSpace(auth.Attributes[key]); value != "" {
			return value
		}
	}
	if auth.Metadata != nil {
		if value, ok := auth.Metadata[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// lobsteraiServerApplyEnterpriseHeaders adds the account's enterprise identity.
//
// The backend scopes an enterprise account to its pool using this pair, and the
// desktop client sends it on every authenticated request. A personal account has
// no context, so neither header is sent: a fabricated id would scope the request
// to a pool the account is not a member of.
func lobsteraiServerApplyEnterpriseHeaders(req *http.Request, auth *cliproxyauth.Auth) {
	if req == nil {
		return
	}
	if !strings.EqualFold(lobsteraiServerAttributeString(auth, "account_mode"), lobsteraiserverauth.AccountModeEnterprise) {
		return
	}
	enterpriseID := lobsteraiServerAttributeString(auth, "enterprise_id")
	if enterpriseID == "" {
		return
	}
	req.Header.Set(lobsteraiserverauth.HeaderAccountMode, lobsteraiserverauth.AccountModeEnterprise)
	req.Header.Set(lobsteraiserverauth.HeaderEnterpriseID, enterpriseID)
}

// lobsteraiServerUpstreamError classifies a non-2xx upstream response with the
// OpenAI-compatible status/retry semantics so the manager's cooldown policy sees
// the real status and Retry-After.
func lobsteraiServerUpstreamError(statusCode int, headers http.Header, body []byte) error {
	return newOpenAICompatStatusError(statusCode, headers, body)
}

// lobsteraiServerStreamErrorPayload reports the in-band error the gateway sends
// as an SSE "event:error" frame inside an HTTP 200 response. The gateway signals
// an unsupported model (40300) and an inaccessible model (40301) this way, so
// treating the 200 as success would silently return an empty completion.
//
// It accepts either the bare JSON payload or a full SSE frame ("event:error"
// followed by a "data:" line), because callers observe both shapes.
func lobsteraiServerStreamErrorPayload(body []byte) (int, string, bool) {
	payload := lobsteraiServerErrorJSON(body)
	if payload == nil {
		return 0, "", false
	}
	message := strings.TrimSpace(gjson.GetBytes(payload, "error.message").String())
	if message == "" {
		return 0, "", false
	}
	// The gateway reports application-level rejections with its own five-digit
	// code (40300 unsupported model, 40301 inaccessible model), not an HTTP
	// status. Derive the HTTP status from it so the manager's cooldown/retry
	// policy sees a class it understands; anything unparsable is a bad gateway.
	status := http.StatusBadGateway
	switch code := gjson.GetBytes(payload, "error.code").Int(); {
	case code >= 400 && code < 600:
		status = int(code)
	case code >= 40000 && code < 60000:
		status = int(code / 100)
	}
	return status, message, true
}

// lobsteraiServerErrorJSON extracts the JSON payload from a bare object or an
// SSE frame, returning nil when no error payload is present.
func lobsteraiServerErrorJSON(body []byte) []byte {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || !bytes.Contains(trimmed, []byte(`"error"`)) {
		return nil
	}
	// A bare JSON object (the non-stream shape).
	if bytes.HasPrefix(trimmed, []byte("{")) && json.Valid(trimmed) {
		return trimmed
	}
	// An SSE frame: pick the JSON out of the "data:" line.
	for _, line := range bytes.Split(trimmed, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		candidate := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		if json.Valid(candidate) && bytes.Contains(candidate, []byte(`"error"`)) {
			return candidate
		}
	}
	return nil
}

// lobsteraiServerSSEToCompletion extracts the JSON objects from an SSE body and
// folds them into a single OpenAI chat.completion. It lets a non-stream client be
// served from the gateway's stream-only response.
func lobsteraiServerSSEToCompletion(sse []byte) []byte {
	var lines bytes.Buffer
	for _, line := range bytes.Split(sse, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmed, []byte("data:")) {
			continue
		}
		candidate := bytes.TrimSpace(bytes.TrimPrefix(trimmed, []byte("data:")))
		if len(candidate) == 0 || bytes.Equal(candidate, []byte("[DONE]")) {
			continue
		}
		if !json.Valid(candidate) {
			continue
		}
		lines.Write(candidate)
		lines.WriteByte('\n')
	}
	if lines.Len() == 0 {
		return nil
	}
	return lobsteraiServerAggregateStream(lines.Bytes())
}

// lobsteraiServerAggregateStream folds OpenAI chat.completion.chunk objects (one
// per line) into a single chat.completion. The gateway streams every response
// regardless of the requested stream flag, so a stream:false client needs the
// chunks reassembled.
func lobsteraiServerAggregateStream(raw []byte) []byte {
	if len(bytes.TrimSpace(raw)) == 0 {
		return raw
	}
	// A raw non-stream body (a gateway change, or an ad-hoc call) is already a
	// chat.completion and passes through untouched.
	if trimmed := bytes.TrimSpace(raw); bytes.HasPrefix(trimmed, []byte("{")) && json.Valid(trimmed) {
		if gjson.GetBytes(trimmed, "object").String() == "chat.completion" {
			return trimmed
		}
	}
	// aggregateCodeBuddyCNChunks already implements the chunk-folding contract
	// for OpenAI chat.completion.chunk objects, including streamed tool-call
	// argument fragments and reasoning_content.
	aggregated := aggregateCodeBuddyCNChunks(raw)
	if !gjson.ValidBytes(aggregated) || len(gjson.GetBytes(aggregated, "choices").Array()) == 0 {
		return raw
	}
	return aggregated
}

// lobsteraiServerForceStream pins the upstream stream flag. The gateway ignores
// a false value (it always answers with SSE), but sending true keeps the wire
// shape aligned with the response for logging and diagnostics.
func lobsteraiServerForceStream(body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	updated, errSet := sjson.SetBytes(body, "stream", true)
	if errSet != nil {
		return body
	}
	return updated
}
