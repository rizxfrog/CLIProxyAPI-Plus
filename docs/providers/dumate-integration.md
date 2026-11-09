# Dumate (百度搭子) Provider — protocol contract and integration evidence

Provider key: `dumate`
Status: **implemented and live-verified** against a running Dumate desktop client on this workstation.
Implementation: `internal/auth/dumate/`, `internal/watcher/synthesizer/file.go`, `sdk/cliproxy/service_executors.go`,
`sdk/cliproxy/service_models.go`, `internal/registry/`.

This provider deliberately implements **no OAuth login**. Dumate has no authorization server, no client id,
no device/user token endpoint and no refresh grant. The operator exports a credential file from the running
desktop client and uploads it, exactly like an API-key credential.

## 0. Snapshot

| Item | Value |
| --- | --- |
| CLIProxyAPI revision | `91473903` on `feature/dev` |
| Client artifact | `Dumate-win-setup-x64-latest.exe` (NSIS), product version `1.0.89`, signed by Baidu |
| Client data dir | `%APPDATA%\qianfan-desktop-app` |
| Backend binary | `dumate-main-server.exe` (`qianfan-desktop-main`), spawned by the Electron shell |
| Inspection date | 2026-11-07 |
| Evidence kinds | `runtime-confirmed` (observed against the live client), `source-confirmed` (client string/bundle), `inferred`, `unknown` |

## 1. Provenance

- All protocol facts below were recovered from the locally installed client and from live loopback traffic
  (`%(APPDATA)s\qianfan-desktop-app\log\electron.log`, `log\desktop-main\*.log`). No third-party reference
  implementation was copied, and no code was ported; the Go provider is an original reimplementation of the
  observed wire contract.
- No client bundle, binary, or log is committed to this repository.

## 2. Provider identity

- Stable key: `dumate`, used consistently across the constant, synthesizer, executor, model catalog, tests and
  documentation.
- Identity is separate from the wire format: the executor reuses the shared OpenAI Chat Completions translator
  without claiming to be `openai`, `codex`, or an `openai-compatibility` config entry.
- The upstream is **local**: an OpenAI-compatible surface served by `dumate-main-server` on
  `http://127.0.0.1:<ephemeral port>`. That process forwards to `https://dumate-svc.baidu.com/gateway/apis/v1/chat/completions`
  (`runtime-confirmed`, `log\desktop-main\Gateway.log`).

## 3. Credential model (no OAuth)

The client authenticates its local backend with three runtime values plus a placeholder bearer token:

| Field | Meaning | Lifetime |
| --- | --- | --- |
| `account_id` | signed-in Baidu account id | stable per account |
| `device_id` | per-installation fingerprint (`1\|windows.<sha256>`) | stable per installation |
| `inapp_key` | per-process in-app key, `crypto.randomBytes(32).toString('hex')` in the client | **regenerated on every client start** |
| `api_key` | always sent as `nokey`; real authorization is the headers above | constant |

Because the in-app key and the backend port both change on restart, the credential is static and
**non-refreshable**: `Refresh` cannot rotate anything, so no refresh lead is registered for `dumate` and the
auto-refresh loop never schedules it (`runtime-confirmed`: `ProviderRefreshLead("dumate", nil) == nil`).

Note: the shared file synthesizer records `auth_kind=oauth` for every file credential. Dumate keeps that
convention so global `oauth-excluded-models` rules keep working; it has no effect on refresh scheduling.

### 3.1 Credential file schema

```json
{
  "type": "dumate",
  "base_url": "http://127.0.0.1:52194/api/qianfanproxy/v1",
  "api_key": "nokey",
  "account_id": "<account id>",
  "device_id": "1|windows.<sha256>",
  "inapp_key": "<64 hex chars>"
}
```

- `base_url` must include the `/api/qianfanproxy/v1` prefix because the shared executor appends
  `/chat/completions`. A value without the prefix gets it appended automatically.
- `port` may be supplied instead of `base_url`; the synthesizer then builds
  `http://127.0.0.1:<port>/api/qianfanproxy/v1`.
- `api_key` defaults to `nokey` when omitted.
- `headers` may be supplied and overrides the generated `X-Dumate-*` headers.

Generate the file with `dumate-creds.py --export auths/dumate.json` (see §9).

## 4. Inference contract

| Aspect | Finding | Kind |
| --- | --- | --- |
| Endpoint | `POST {base_url}/chat/completions` | runtime-confirmed |
| Auth | `Authorization: Bearer nokey` + `X-Dumate-Account-Id` + `X-Dumate-Device-Id` + `X-Dumate-Inapp-Key` | runtime-confirmed |
| Body | OpenAI Chat Completions (`model`, `messages`, `stream`, `max_tokens`, `reasoning_effort`) | runtime-confirmed |
| Streaming | standard SSE: `data: {...}` chunks with `choices[].delta`, final chunk carries `usage`, terminated by `data: [DONE]` | runtime-confirmed |
| Reasoning | `reasoning_effort` accepted for `none`/`minimal`/`low`/`medium`/`high`/`xhigh`; reasoning text returned in `reasoning_content` | runtime-confirmed |
| Model listing | **none**: `GET {base_url}/models` returns 404 `dumate api not registered` | runtime-confirmed |
| Responses API | **none**: `POST {base_url}/responses` returns 404 | runtime-confirmed |
| Anthropic Messages | `POST {base_url}/messages` exists but rejects models bound to an OpenAI engine (`anthropic request must hit an Anthropic engine gapi, got engine="OpenAI"`) | runtime-confirmed |

Because the upstream is OpenAI Chat Completions, the shared executor provides the OpenAI Chat Completions,
OpenAI Responses and Anthropic Messages frontends by translating to that format. Anthropic clients therefore
never touch the backend's own `/messages` route.

## 5. Models

`internal/registry/models/models.json` (`dumate` section) is authoritative; there is no discovery endpoint.

| Model ID | Resolved upstream (`X-Dumate-Gw-Realmodel`) | Display name | Kind |
| --- | --- | --- | --- |
| `glm-5` | `glm-5.2` | GLM-5.2 | runtime-confirmed |
| `deepseek-v4-pro` | `deepseek-v4-pro-0813` | DeepSeek-V4 Pro | runtime-confirmed |
| `deepseek-v4-flash-0731` | `deepseek-v4.1-flash-low-latency` | DeepSeek-V4 Flash (0731) | runtime-confirmed |
| `deepseek-v41-flash` | `deepseek-v4.1-flash-low-latency` | DeepSeek-V4.1 Flash | runtime-confirmed |

Availability is account/plan dependent. IDs observed in client logs but **rejected** by the backend on this
account (`invalid_model`) are deliberately excluded: `glm-5.2`, `glm-4.6`, `deepseek-v3.1`, `deepseek-v3.2`,
`kimi-k2`.

Only text input is advertised. Image input was not exercised and is **unknown**.

The fork's remote catalog (`rizxfrog/CLIProxyAPI-Plus`, `internal/registry/models/models.json`) must be updated
with the same `dumate` section, otherwise a remote catalog refresh replaces the embedded catalog and the
`dumate` models disappear until the section is published there.

## 6. Quota

**unknown / unsupported.** No quota, balance or rate-limit endpoint was found on the local backend, and no
quota response headers were observed. Dumate does not participate in quota observation or scheduler cooldown
from this provider; usage accounting works and comes from the upstream `usage` object.

## 7. Errors

| Condition | Observation |
| --- | --- |
| Unknown model | HTTP 404, `{"error":{"code":"invalid_model","message":"model \`X\` does not exist. ..."}}` |
| Empty body | HTTP 400, `{"message":"invalid llm request body: request body is empty","success":false}` |
| Missing base URL | executor returns 401 `missing provider baseURL` |
| Wrong engine for Anthropic route | HTTP 400, `anthropic request must hit an Anthropic engine gapi, got engine="OpenAI"` (not reached via this provider) |

The client is a local process: when it is not running, `base_url` fails to connect and the request errors out
normally. No account switching happens mid-stream; retries use the shared manager policy.

## 8. Registration checklist

| Item | State |
| --- | --- |
| Provider constant | `internal/constant/constant.go` → `Dumate` |
| Credential shape helpers | `internal/auth/dumate/dumate.go` |
| File-synthesizer mapping | `internal/watcher/synthesizer/file.go` (base URL, bearer token, `X-Dumate-*` headers) |
| Executor binding | `sdk/cliproxy/service_executors.go` → `executor.NewOpenAICompatExecutor(constant.Dumate, cfg)` |
| Model registration | `sdk/cliproxy/service_models.go` → `registry.GetDumateModels()` |
| Model catalog | `internal/registry/models/models.json`, `GetDumateModels`, channel dispatch, static lookup |
| Management provider list | `internal/api/handlers/management/provider_models.go` → `dumate` |
| Credential upload | `POST /v8/management/credentials` (multipart) or legacy `POST /v0/management/auth-files`; panel **Upload File** |
| Credential exporter | `dumate-creds.py` (repository root; not part of the Go module) |
| Panel icon + label | `frontend/src/assets/icons/dumate.png`, `AUTH_FILE_ICONS.dumate`, `TYPE_COLORS.dumate`, `auth_files.filter_dumate` |
| Quota probe | `internal/auth/dumate/console.go` + `auth_files_dumate_quota.go` → `GET /v8/management/dumate-quota` |
| Console cookie source | decrypted from the client's own store (`console_store.go`); an explicit `console_cookie` in the credential overrides it |
| Daily check-in | `internal/checkin/dumate.go`, registered in `checkin.Capabilities()` and `sdk/cliproxy/service_checkin.go` |
| Panel quota tab | `src/features/quota/providers/dumate/*`, `dumateQuota.ts`, plus the four quota registries |
| Panel check-in card | none — the panel renders one card per backend capability |
| CLI login / management OAuth start | **not applicable** — no OAuth flow exists |
| Refresh lead registration | **not applicable** — credential is static |
| Manual-refresh action | **not applicable** — nothing to rotate |

## 9. Usage

### 9.1 Export the credential

```bash
py dumate-creds.py                     # print a summary (secrets masked)
py dumate-creds.py --export auths/dumate.json
```

The exporter reads the client log (`account_id`, `inapp_key`, backend port), the device fingerprint from
`qianfan_desk_xdg/data/sts/device_id.json`, and confirms the port by fingerprinting the backend's inference
path. Start the Dumate client and sign in first.

### 9.2 Upload

```bash
curl -X POST "http://127.0.0.1:8317/v8/management/credentials" \
  -H "Authorization: Bearer <management key>" \
  -F "file=@auths/dumate.json"
```

`/v8/management/credentials` is the multipart upload route the management panel uses. The legacy
`/v0/management/auth-files` route (`?name=...` raw JSON or multipart) also accepts the file. The auth-file
watcher picks it up immediately; no restart is required.

In the panel, use **Auth Files → Upload File** and select the exported JSON.

The WebUI panel's file-upload control accepts the same file. Uploading a file whose `base_url`/`port` is
stale makes requests fail until the client is restarted and the file re-exported.

### 9.3 Call the proxy

```bash
# list the registered models
curl http://127.0.0.1:8317/v1/models -H "Authorization: Bearer <proxy api key>"

# OpenAI Chat Completions (non-streaming)
curl http://127.0.0.1:8317/v1/chat/completions \
  -H "Authorization: Bearer <proxy api key>" -H "Content-Type: application/json" \
  -d '{"model":"glm-5","messages":[{"role":"user","content":"hello"}]}'

# OpenAI Chat Completions (streaming)
curl -N http://127.0.0.1:8317/v1/chat/completions \
  -H "Authorization: Bearer <proxy api key>" -H "Content-Type: application/json" \
  -d '{"model":"deepseek-v4-pro","messages":[{"role":"user","content":"hello"}],"stream":true}'

# Anthropic Messages
curl http://127.0.0.1:8317/v1/messages \
  -H "Authorization: Bearer <proxy api key>" -H "Content-Type: application/json" \
  -d '{"model":"glm-5","max_tokens":64,"messages":[{"role":"user","content":"hello"}]}'
```

## 10. Verification

Commands run on 2026-11-07:

```bash
go build ./...
go test ./internal/auth/dumate/... ./internal/registry/... ./internal/watcher/synthesizer/...
go test ./internal/api/... ./sdk/cliproxy/... ./internal/runtime/executor/... ./internal/constant/...
```

Live evidence collected by running the freshly built server against a real Dumate client and credential:

| Check | Result |
| --- | --- |
| `GET /v1/models` | 200; `glm-5`, `deepseek-v4-pro`, `deepseek-v4-flash-0731`, `deepseek-v41-flash` present |
| `POST /v1/chat/completions` (non-streaming) | 200; `content="proxy ok"`, `finish_reason=stop`, `usage.total_tokens=55` |
| `POST /v1/chat/completions` (streaming) | 200; 19 SSE chunks, re-assembled `"streaming works"`, final chunk carries `usage` |
| `POST /v1/messages` (Anthropic) | 200; `stop_reason=end_turn`, thinking block + `text="anthropic ok"`, `usage` reported |
| `POST /v1/responses` | 200; `status=completed`, reasoning + text output |
| `POST /v8/management/credentials` (multipart) | 200; credential listed with `provider=dumate`, 4 models appear without a restart |
| `GET /v8/management/provider-models` | 200; `dumate` section with 4 models |
| `GET /v8/management/dumate-quota` | 200 with the live wallet: `available_points=5520.00`, `subscription_points=4993.91`, `bonus_points=520.00`, `expires_at` set |
| `GET /v8/management/automation/checkin/providers` | 200; `dumate` listed alongside `codearts`, `codebuddy-cn`, `trae` |
| Enabling the check-in provider | Was **422** `automation.checkin: unsupported provider "dumate"` until 2026-11-09 (the config allowlist had drifted); now `PUT /v8/management/config.yaml` with `dumate: {enabled: true}` returns **200** and the status endpoint reports it enabled |
| Panel (`frontend/`) in a real browser | Auth Files banner, Quota Management tab (live wallet rendered), and the Automatic check-in Dumate card in both light and dark themes |

Frontend checks: `bun test tests/dumateProvider.test.ts tests/dumateQuota.test.ts`, `bun run type-check`,
`bun run lint`, `bun run build`, plus the whole `bun test` suite — the 27 failures are byte-identical to the
pre-change baseline (verified by diffing the failure names), and the new suites add passing coverage.

Backend checks: `gofmt`, `go build ./...`, and the focused packages
(`internal/auth/dumate`, `internal/checkin`, `internal/api/handlers/management`, `internal/watcher/...`,
`sdk/cliproxy`, `internal/registry`). `sdk/cliproxy` reported three timing-sensitive failures under a
parallel multi-package run that pass in isolation and on a single-package run.

## 11. Unsupported / blocked / unverified

- **No OAuth login**: no authorization server exists; credentials are user-exported. This is a capability gap
  in the upstream product, not an unimplemented feature.
- **No refresh**: the in-app key is regenerated per client process, and the backend port changes on restart.
- **No quota**: no observable quota source; not reported rather than reported as zero or unlimited.
- **Image/audio input**: untested — the catalog advertises text only.
- **Tool calling**: not exercised in this session; the shared translator provides it, but no live proof.
- **Multi-account**: one credential file describes one live client session; multiple concurrent clients on the
  same machine are not addressed.
- **Panel quota/refresh affordances**: the panel shows the points wallet as a quota tab; there is no
  manual-refresh action, matching the backend, which registers no refresh lead for this provider.
- **Live check-in claim not triggered**: the adapter is unit-tested for the issued / already-done / refused /
  upstream-failure paths and its read call is live-verified, but no claim was executed against the real account
  — today's bonus was already claimed, and enabling the provider would have meant rewriting the operator's
  config to prove a code path the shared scheduler already covers for the other three providers.
- **Quota bucket families**: only the plan and bonus families are summed. A third family (purchased top-ups)
  has no observed sample, so it is not mapped; `available_points` would show the difference if one appears.
- **Cookie store portability**: reading the client's own store assumes the proxy runs on the same host as the
  client, which inference already requires. A deployment elsewhere must supply `console_cookie` explicitly.
- **Release packaging**: the panel release is published from `rizxfrog/Cli-Proxy-API-Management-Center`; the
  quota tab and icon reach users only after that repository cuts a release.

## 12. Points wallet (Quota Management)

The wallet lives on `console.bce.baidu.com`, **not** on the client's local backend, so it needs the Baidu
console session rather than the inference credential.

### 12.1 Where the cookie comes from

The client stores its cookie jar in `auth.json` (field `cookies`) encrypted with the raw 32-byte key in
`.cookie-key`, as `base64(iv[12] ‖ tag[16] ‖ ciphertext)` under AES-256-GCM — the same layout
`lib/crypto/cookie.js` implements. The proxy decrypts that store directly
(`internal/auth/dumate/console_store.go`) instead of asking the operator to copy a live session cookie into a
second file: the local client is already required for inference, and a re-login then takes effect on the next
read without re-exporting. A credential may still carry `console_cookie` to override the store, which is the
path for a deployment that does not run the client on the same host.

The Go side must append the tag *after* the ciphertext (`crypto/cipher` convention), while the stored layout
puts it before: the two orders are not interchangeable and mixing them fails authentication.

Only unexpired `*.baidu.com` cookies are sent; a jar with no usable cookie is reported as an error rather than
sent as an empty header.

### 12.2 Endpoint

```http
GET /v8/management/dumate-quota?auth_index=<index>
```

```json
{
  "provider": "dumate",
  "observed_at": "2026-11-07T15:03:59Z",
  "available_points": "5520.00",
  "used_points": "6.09",
  "is_subscribed": true,
  "throttled": false,
  "throttle_reason": "",
  "throttle_type": "",
  "subscription_points": "4993.91",
  "bonus_points": "520.00",
  "expires_at": 1794653785
}
```

Upstream contract (`GET /api/dumate/points/quota_overview?timezone=…&clientType=desktop&ignoreLoginBonus=true`)
and its envelope are documented in `docs/reverse-note/dumate-oauth_CN.md` §21.

Design notes:

- Decimal strings are **passed through** in the console's own units; the handler never re-rounds the totals.
- `subscription_points` sums the plan buckets and `bonus_points` sums every granted or event bucket. The
  console's other package types (purchased top-ups) have no observed sample, so they are deliberately not
  folded into a bucket: `available_points` remains the console's own total, which makes a new family visible
  as a difference instead of silently landing in the wrong row.
- `expires_at` is the soonest future expiry across active buckets. An inactive bucket never contributes.
- The upstream origin is fixed by the client; credential metadata cannot redirect a privileged management read.
- A credential whose cookie is unreadable returns a clear 400 rather than a fabricated zero balance.

The panel renders these as Available / Plan / Bonus / Points used plus the expiry line, matching the app's own
Settings → Points screen.

## 13. Daily check-in

```http
GET  /api/dumate/points/loginBonusInfo   # read state
POST /api/dumate/points/loginBonus       # claim (no body, no query)
```

The adapter (`internal/checkin/dumate.go`) reads the state first and only claims when the console does not
already count today:

- `hasIssued` is the console's own flag; `signInDays` covers a timezone disagreement between the scheduler's
  business date and the console's day.
- A claim that answers success-without-grant (`result:false`) is reported as `already_done`, not as a failure,
  and is not retried blindly.
- Upstream errors are reported as `failed` with the console's message; the credential never appears in the
  message.
- An unreadable cookie store is `blocked` with the read error as a warning, so the operator sees why.

The provider is listed by `checkin.Capabilities()` and is **disabled by default**; the operator enables it and
picks the timezone/window from the panel's Automatic check-in page. The panel needs no per-provider code: it
renders one card per capability returned by
`GET /v8/management/automation/checkin/providers`.

**Correction (2026-11-09)**: being listed was not enough. The provider id lives in two lists —
`checkin.Capabilities()` (which the panel renders) and the allowlist in
`internal/config/auto_checkin.go` (`AutoCheckinConfig.Validate()`) — and the latter still named only
`codearts/codebuddy-cn/trae`. Enabling Dumate therefore saved with
`422 automation.checkin: unsupported provider "dumate"`. The allowlist now covers every built-in
provider, `checkin.Capabilities()` and the adapter set come from one list
(`checkin.BuiltinAdapters()`), and a test pins the two together.
