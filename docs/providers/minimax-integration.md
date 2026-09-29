# MiniMax Code (`minimax-code`) Provider Integration — Evidence Worksheet

Status: **implemented and mock-verified; live checks pending**. OAuth device-flow
constants and the Anthropic-Messages inference endpoint are fully source-confirmed.
The integration is complete end-to-end (auth → persistence → refresh → executor →
models → CLI + management login → tests); live verification against a real MiniMax
account has not been performed.

## Snapshot
- CLIProxyAPI commit: `06c4801edaac08938900fb0f1f84f5ab33544a56` (clean working tree).
- Client artifact: `/home/van/github/MiniMax-AI/minimax-code`, version `0.5.5`,
  commit `0f6ad52` (`fix(tui): distinguish Bash recaps ... (#369)`), MIT license.
  Already a public open-source projection; reuse is permitted but no private
  attribution embedded.
- Inspection date: 2026-06-18.

## Provider identity
- Recommended canonical key: `minimax` (already reserved in
  `internal/registry/model_definitions.go` for the Cline/OpenAI-compat
  MiniMax API surface; the managed account path reuses the same key with an
  OAuth `managed-login` auth kind).
- Account identity: JWT claim `account_id` and/or `sub` from the access token.
- Region variants: `cn` (`account.minimax.cn` / `agent.minimax.cn`) and
  `en` (`account.minimax.io` / `agent.minimax.io`). The client resolves the
  region from its build environment; CLIProxyAPI should expose it as a config
  field with `en` as default.

## Authorization (OAuth 2.0 Device Authorization Grant, RFC 8628) — SOURCE-CONFIRMED
Source: `packages/oauth-core/src/contracts.ts`, `oauth-client.ts`,
`endpoint-config.ts`.

- Client ID (public): `mcode-public`
- Scopes: `agent.default`
- Audience: `agent-backend`
- PKCE: S256 required (`code_challenge`, `code_challenge_method=S256`)
- Device authorization endpoint: `https://account.{region}/oauth2/device/code`
- Token endpoint: `https://account.{region}/oauth2/token`
- Revocation endpoint: `https://account.{region}/oauth2/revoke`
- Device authorization request body: `client_id`, `scope`, `audience`,
  `code_challenge`, `code_challenge_method=S256`.
- A `staging` build env sends `X-User-Pre: 1` on the device endpoint.
- User-facing verification URI is returned as `verification_uri`/`verification_url`
  plus `user_code`, `device_code`, `interval`, `expires_in`.

## Token exchange — SOURCE-CONFIRMED
- Grant `urn:ietf:params:oauth:grant-type:device_code` with `device_code` and
  `code_verifier` (or `user_code` when the account variant uses user-code polling).
  Polls `status` field: `pending`/`authorization_pending` -> continue,
  `slow_down` -> +5s interval, `denied`/`access_denied`/`expired` -> terminal error.
- Refresh grant: `grant_type=refresh_token`, `refresh_token`, `client_id`,
  `scope`, `audience`. Response JSON: `access_token`, `refresh_token`
  (rotated; previous retained when omitted), `token_type=Bearer`,
  `expires_in` (seconds), and `account_id`/`sub` from the JWT `access_token`.

## Credential storage (client) — SOURCE-CONFIRMED (shape reference only)
`StoredCredential`: `accessToken`, `refreshToken`, `tokenType=Bearer`,
`clientId=mcode-public`, `scopes`, `audience`, `expiresAtMs`, `generation`,
`subject?`, `accountId?`, `loginEpoch?`. CLIProxyAPI persists its own schema
(`Metadata.access_token`/`refresh_token`/`expired` + `account_id` attribute),
so we map to the shared auth store rather than copying this layout.

## Inference — CONFIRMED (endpoint + auth)
The managed MiniMax (MCode) provider is `authMode: "managed-login"` and uses the
**Anthropic Messages** wire format (`api = "anthropic-messages"`).

Decisive evidence: `packages/local-runtime-v2/src/service/model-system/resolution/
local-model-resolver.test.ts` asserts the managed preset resolves to
`model.baseUrl = https://agent.minimax.io/mavis/api/v1/llm` with headers
`Authorization: Bearer <oauth access token>` + `User-Agent: MiniMaxAgent`. The
configured preset baseURL is `https://agent.minimax.io/mavis/api/v1/llm/v1`;
`normalizeProviderBaseUrl` strips the trailing `/v1`, and the Anthropic SDK
(`@anthropic-ai/sdk` `buildURL`) then appends `/v1/messages`.

Managed inference surface (source-confirmed):
- Chat: `POST https://agent.minimax.io/mavis/api/v1/llm/v1/messages` (en) /
  `POST https://agent.minimax.cn/mavis/api/v1/llm/v1/messages` (cn).
- Count tokens: same base + `/v1/messages/count_tokens`
  (`packages/local-runtime/src/context/token-counter-adapters/messages.ts`).
- Headers: `Authorization: Bearer <access_token>`, `content-type: application/json`,
  `anthropic-version: 2023-06-01`, and `User-Agent: MiniMaxAgent`
  (`MANAGED_PROVIDER_USER_AGENT`).
- Managed routing headers (`X-Mavis-Session-Id`, `X-Mavis-Agent-Id`,
  `bedrock-lane`) are client-identity hints injected by
  `buildLocalProviderHeaders`; they are not required for a reverse proxy and are
  intentionally omitted.

## Models — SOURCE-CONFIRMED catalog
Source: `packages/config/src/config.ts` `MINIMAX_MODELS` / `MINIMAX_API_MODEL_CATALOG`.
- `MiniMax-M3` — context 512000 (option 1000000, hint higher_usage), output 128000,
  reasoning switchable default on, supports text/image/video input, file API.
- `MiniMax-M2.7-highspeed` — context 200000, output 128000, reasoning+tool_call.
- `MiniMax-M2.7` — context 200000, output 128000, reasoning+tool_call.
CLIProxyAPI already ships a `minimax/` prefixed model list (M1/M2/M2.1/M2.5/M2.7/
M2-her/M3/M2.1/01) of type `cline` (OpenAI-compat). The managed account uses the
`MiniMax-M3`/`MiniMax-M2.7*` IDs and the Anthropic-Messages format.

## Quota — UNKNOWN at this stage
- No quota endpoint is referenced in the managed-login path. The client tracks
  usage through response headers/websocket quota events in other providers; for
  MiniMax managed there is no public probe in the supplied source.
- Plan: expose unknown/limited and report observed response metadata only if the
  upstream emits usage in the Anthropic response `usage` block. Do NOT fabricate
  limits.

## Streaming / tools / thinking
- Anthropic Messages SSE streaming is reused from the Claude translator path.
- Thinking: `MiniMax-M3` exposes a switchable `thinking` block
  (`thinking.type=adaptive|disabled`). Reuse canonical thinking pipeline.
- Tools: Anthropic `tool_use`/`tool_result` reused from Claude translator.

## Verification plan
- Unit tests with `httptest` mock upstream emulating `/messages` Anthropic SSE.
- OAuth device-flow tests against a local mock token/device endpoint.
- `go build -o /tmp/cli-proxy-api ./cmd/server`.
- Live checks **pending** user MiniMax account + consent.

## Implementation coverage (as built)

Provider keys: `minimax` (international, `account.minimax.io`) and `minimax-cn`
(mainland, `account.minimax.cn`). Both share one implementation.

| Area | Files | Verification |
| --- | --- | --- |
| OAuth device flow, refresh, revoke | `internal/auth/minimax/minimax.go` | mock tests |
| SDK authenticators | `sdk/auth/minimax.go`, `sdk/auth/refresh_registry.go` | compile + registry tests |
| Executor (Anthropic Messages) | `internal/runtime/executor/minimax_executor.go` | mock tests (Claude/OpenAI/Responses, streaming, adaptive thinking, error status) |
| Executor binding | `sdk/cliproxy/service_executors.go` | binding test |
| Model catalog + registration | `internal/registry/models/models.json`, `model_definitions.go`, `model_updater.go`, `sdk/cliproxy/service_models.go` | registry + registration tests |
| CLI login | `internal/cmd/minimax_login.go`, `cmd/server/main.go` (`-minimax-login`, `-minimax-cn-login`) | build + `--help` |
| Management OAuth | `internal/api/handlers/management/auth_files_provider_oauth.go`, `oauth_sessions.go`, `server_management.go` (`/minimax-auth-url`, `/minimax-cn-auth-url`) | compile |
| TUI login choice | `internal/tui/oauth_tab.go` | compile |
| Restart/load synthesis | `internal/watcher/synthesizer/file.go` | synthesizer tests |
| Usage semantics | `sdk/cliproxy/usage/accounting.go` | usage tests |
| Evidence worksheet | `docs/providers/minimax-integration.md` | this file |

### Confirmed vs. unverified
- Source-confirmed: OAuth endpoints/client id/scope/audience/PKCE, device-code
  and refresh grants, JWT claims, the managed inference base URL, the Anthropic
  Messages wire format and Bearer auth, the model catalog, and the adaptive
  `thinking` dialect.
- Mock-verified: device-code login, refresh rotation retention, streaming and
  non-streaming Claude/OpenAI Chat/OpenAI Responses translation, adaptive
  thinking emission, upstream error status preservation, model registration and
  restart synthesis.
- Not verified live: real MiniMax account login, live inference/streaming, real
  refresh after expiry, quota/usage. The task's completion status is therefore
  **implemented and mock-verified; live checks pending**.
- Quota remains **unknown**; no fabricated limits. Management reset only clears
  local routing state.

