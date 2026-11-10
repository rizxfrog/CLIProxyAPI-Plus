# Xiaohuanxiong Provider

This document records how the `xiaohuanxiong` provider works in CLIProxyAPI, the
protocol it implements, and how the implementation was derived from the official
desktop client.

## What it is

`xiaohuanxiong` (商汤小浣熊 / Raccoon) is a **hosted gateway** provider. SenseTime's
office assistant exposes an OpenAI-compatible chat-completions endpoint that
fronts several upstream model families:

```
https://xiaohuanxiong.com/api/web/llm/v2/chat/completions
```

The provider reuses the standard OpenAI request/response translation; what is
provider-specific is credential acquisition and the per-model-family
deep-thinking dialect.

## Credentials

Xiaohuanxiong has **no standard OAuth2 authorization server**. The desktop client
performs a browser-based authorization-code exchange:

1. It opens the web login page in the system browser, adding the parameters the
   upstream expects:
   ```
   https://xiaohuanxiong.com/code/authorize?login_source=desktop&appname=办公小浣熊客户端
   ```
2. The site redirects to a desktop deep link carrying a one-time code:
   ```
   office-raccoon://auth/callback?code=<AUTHORIZATION_CODE>
   ```
3. The code is exchanged for tokens:
   ```
   POST /api/web/auth/v1/login_with_authorization_code
   {"authorization_code": "<code>"}
   ```
   Response `data` carries `access_token`, `refresh_token`, `office_identity`,
   `office_org_name`, and `office_org_role`.

The upstream error code `200035` means the authorization code is missing,
expired, or already consumed.

Access tokens are short-lived JWTs. When `exp` is within 300 seconds, the token
is refreshed:

```
POST /api/web/auth/v1/refresh
{"refresh_token": "<refresh_token>"}
```

An opaque token (manual paste) has no refresh path; it is used until it expires.

### Login flow

```
GET  /v0/management/xiaohuanxiong-auth-url          -> { url, state, callback }
POST /v0/management/xiaohuanxiong-auth-callback     -> { state, code | redirect_url }
```

The GET endpoint returns the login URL and waits in the background. The flow
completes through any of:

- `POST /v0/management/oauth-callback` with `provider: "xiaohuanxiong"`
- `GET  /v0/management/oauth-callback?provider=xiaohuanxiong&code=...`
- `POST /v0/management/xiaohuanxiong-auth-callback` with the callback URL (or a
  bare code) — the escape hatch when the `office-raccoon://` deep link cannot
  reach the proxy.

The `code` / `redirect_url` field of the dedicated endpoint accepts both formats:

```
office-raccoon://auth/callback?code=<AUTHORIZATION_CODE>
<AUTHORIZATION_CODE>
```

Because the browser triggers the deep link client-side, it never reaches a server
and the operator has to paste it. `ParseCallbackCode` therefore also tolerates the
other places the code surfaces: an `https` callback, a fragment form, and a query
pair pasted without its leading `?` (`code=<code>`, `authorization_code=<code>`).
Anything that is not URL- or query-shaped is taken verbatim as the code, so a
typed or DevTools-copied code needs no massaging.

Provider aliases accepted by the callback router: `xiaohuanxiong`,
`xiaohuanxiong.com`, `raccoon`, `sensetime-raccoon`.

## Configuration

```yaml
xiaohuanxiong-api-key:
  - api-key: "YOUR_ACCESS_TOKEN"
    refresh-token: "YOUR_REFRESH_TOKEN"     # optional: enables automatic rotation
    base-url: "https://xiaohuanxiong.com/api/web/llm/v2"  # optional override
    weight: 1
    prefix: "xhx"                           # optional: require xhx/glm-5-3
    proxy-url: "socks5://proxy.example.com:1080"          # optional
    excluded-models:
      - "kimi-k3"
    models:                                 # optional: override the built-in catalog
      - name: "glm-5-3"
        alias: "glm-5-3"
```

The key entry is a dedicated struct (not an alias of the CodeBuddy shape)
because it needs `refresh-token`. A key entry without `api-key` is dropped, and
duplicate `api-key` + `base-url` pairs are collapsed.

## Models

### Live per-account catalogue

Each account's list is read from the gateway itself:

```
GET {origin}/api/web/llm/v2/model_catalog
Authorization: Bearer <access_token>
```

The desktop client calls the same endpoint (`fetchCatalog` in `app.asar`,
normalized by `normalizeLocalAgentRoutingCatalog`). `code` must be `0`, and the
models live in the category whose `type` is `chat`. A non-zero code is an error,
never an empty catalogue.

`sdk/cliproxy/xiaohuanxiong_models.go` drives the refresh with the same shape as
the other live catalogues: the static list is published first as a fallback, a
detached per-account fetch runs at registration, a success republishes the
account's models, and the scan repeats every minute with a 3-hour trust window
and a 10-minute backoff after a failure. A failed or empty fetch keeps the
fallback list published.

Measured on 2026-11-10 for one personal account — 9 chat models. `visible` is the
gateway's own presentation flag, not a routing restriction: the hidden rows are
still routable and one of them is the default.

| Model | Context | `ability_level` | Billing | `visible` |
| ----- | ------- | --------------- | ------- | --------- |
| `raccoon-090dd3` (default) | 1000000 | 3 | normal ×1 | no |
| `raccoon-19b265` | 1000000 | 3 | normal ×1 | no |
| `raccoon-405a1c` | 1000000 | 2 | normal ×1 | no |
| `sn-sensenova-6-8-flash` | 256000 | 1 | lite ×0.5 | yes |
| `sn-sensenova-6-8-flash-lite` | 256000 | 1 | lite ×0.5 | yes |
| `sn-glm-5-3` | 1000000 | 2 | normal ×0.75 | yes |
| `sn-kimi-k3` | 1000000 | 2 | normal ×1 | yes |
| `sn-glm-5-3-flash` | 1000000 | 2 | lite ×0.2 | yes |
| `sn-deepseek-v4-1-flash-1017` | 1000000 | 3 | normal ×0.25 | yes |

> **Never send a fallback ID.** For an unknown model name the gateway answers
> HTTP 200 and silently serves the account default instead of returning an error,
> so a stale ID such as `glm-5-3` yields a correct-looking response produced by
> `raccoon-090dd3`. Third-party models carry an `sn-` prefix on the wire.

### Fallback catalogue

The catalog ships as the `xiaohuanxiong` section of the shared
`internal/registry/models/models.json`, refreshed together with the other
providers by the model catalog updater.

Remote refresh replaces the shared catalog wholesale, so the fork-only
sections (this one included) are carried over by
`preserveLocalCatalogSections` whenever an upstream payload omits them; a
section that the remote payload does start to publish takes over from the
embedded copy.

| Model ID | Context | Max output | Thinking dialect |
| -------- | ------- | ---------- | ---------------- |
| `raccoon-chat-ml-5-5` | 180000 | 80000 | default |
| `glm-5-3` | 200000 | 65536 | `reasoning_effort` high/low |
| `glm-5-3-flash` | 200000 | 65536 | `reasoning_effort` high/low |
| `kimi-k3` | 256000 | 65536 | default |
| `deepseek-v4-pro-0813` | 163840 | 65536 | `extra_body.thinking.type` |
| `sn-sensenova-6-8-flash-lite` | 131072 | 65536 | none (no thinking channel) |
| `sensenova-6-8-flash` | 131072 | 65536 | none |

> These IDs are the embedded fallback only. All but
> `sn-sensenova-6-8-flash-lite` were measured to be absent from the live
> catalogue, so they are replaced as soon as the first probe succeeds. The
> thinking dialect stays keyed on model-name markers and is applied per family.

## Deep thinking

The gateway speaks the OpenAI wire protocol, so the shared thinking pipeline
dispatches on the `openai` format and emits a generic `reasoning_effort`. Each
upstream model family expects a different dialect, so the executor rewrites the
**final upstream body** (the `outgoingTransforms` hook, the same mechanism
CodeBuddy CN uses):

| Model family | Thinking enabled | Thinking disabled |
| ------------ | ---------------- | ----------------- |
| `glm-5-3` / `glm-5.3` | `reasoning_effort: "high"` | `reasoning_effort: "low"` |
| contains `deepseek` / `doubao` | `extra_body.thinking.type: "enabled"` | `extra_body.thinking.type: "disabled"` |
| contains `qwen` | `extra_body.enable_thinking: true` | `extra_body.enable_thinking: false` |
| contains `gemini` | `reasoning_effort: "high"` | `reasoning_effort: "none"` — except `gemini-2.5-pro` / `gemini-3.1-pro`, which never receive a disable value |
| `sensenova-` / `sn-sensenova-` | `reasoning_effort: "high"` | `reasoning_effort: "none"` |
| everything else | `reasoning_effort: "high"` | field removed |

Rules that the implementation guarantees, with tests:

- A body **without** `reasoning_effort` carries no thinking intent and is
  returned byte-for-byte unchanged, so a request that never asked for thinking
  is never given a disable value.
- For families that use `extra_body`, the generic top-level `reasoning_effort` is
  stripped so an unsupported field never leaks upstream.
- Disabled thinking is never turned into an enabled state for any family.

Reasoning text is read from the `reasoning` / `reasoning_content` delta aliases,
and inline `<think>...</think>` markup is relocated to the thinking channel by
the shared pipeline.

## Request correlation

The upstream accepts the desktop client's correlation headers. They are optional
and are not injected by default, but can be supplied per credential through the
`headers` map:

```
X-RACCOON-Session-ID, X-RACCOON-Turn-ID, X-RACCOON-Title, X-RACCOON-Call-Kind
x-client-name, x-client-platform, x-client-version, x-client-os-version,
x-client-channel, x-client-device-id
```

## Verification

- `internal/auth/xiaohuanxiong` — code parsing, authorization URL parameters,
  code exchange, refresh, catalog fetch, JWT expiry, and token persistence.
- `internal/thinking/provider/xiaohuanxiong` — every dialect branch, the
  no-intent passthrough, the `extra_body` strip, the Gemini pro carve-out, and
  the disabled-never-enables invariant.
- `internal/runtime/executor` — provider key, credential bridging, and the
  outgoing transform.
- `internal/registry` — embedded catalog contents, thinking metadata, and
  channel dispatch.
- `internal/config` — the `refresh-token` field is accepted (regression guard:
  an earlier alias-based shape silently dropped it, which would have disabled
  rotation without any error).
- `internal/watcher/synthesizer` — the auth record carries the refresh token and
  the expected base URL.
- `internal/api/handlers/management` — provider normalization, callback file
  naming agreement with the shared writer, and the waiter's code/error/cancel
  paths.
- `sdk/cliproxy/xiaohuanxiong_models_test.go` — catalogue parsing (chat category
  only, blank names dropped, non-zero code rejected), per-model mapping
  (identity, context window, image capability from the `vision` tag), the origin
  derivation that strips the gateway path, session-scoped catalogue keys, the
  end-to-end probe against a stub gateway, and the failed-fetch retry path.

Live verification performed during development:

```
GET https://xiaohuanxiong.com/api/web/llm/v2/model_catalog
  (no header)                        -> 401 {"code":200001,"message":"authorization_empty_error"}
  (raw token, no Bearer prefix)      -> 401 {"code":200002,"message":"authorization_invalid_error"}
  (Authorization: Bearer <token>)    -> 200 {"code":0,"message":"success","data":{"categories":[...]}}
```

An authenticated fetch returned 9 chat models for the account under test, and the
same names were confirmed routable on `POST /api/web/llm/v2/chat/completions`.
The 401 shapes also confirm the `/models`-style `token:` header is NOT accepted
here: this gateway wants `Authorization: Bearer`.

## Limitations

- **The fallback IDs are stale.** They are replaced by the live catalogue after
  the first successful probe, but until then a fallback ID must not be used: the
  gateway answers HTTP 200 for an unknown name and silently serves the account
  default. Any id can be overridden per credential via `models[].name`.
- **`max_completion_tokens` is not carried per model.** The live catalogue
  publishes `params.max_tokens`, but the registry entry keeps the fallback's
  value; only `context_window` is taken from the live row.
- **Hidden models are published too.** The three `raccoon-*` rows are
  `visible:false` yet routable, and one is the account default, so filtering them
  would drop a usable model. Accounts that want them gone can list them in
  `excluded_models`.
- **The thinking dialect is keyed on model-name markers** copied from the
  official client. A newly added upstream family would fall through to the
  default family.
- **DeepSeek/Doubao and Qwen families emit a top-level `reasoning_effort` from
  the shared pipeline that is then moved into `extra_body`.** If the upstream
  also honours the top-level field for those models, the strip is a no-op
  change in behavior.
