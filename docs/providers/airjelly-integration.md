# AirJelly Lite managed account

The `airjelly` provider exposes the single verified Free-tier model
`airjelly-lite`. It uses the AirJelly desktop account's Supabase session and
forwards OpenAI Chat Completions to the signed AirJelly gateway. Both JSON and
SSE responses, including tool calls, are supported.

## Runtime setup

Set these environment variables on the proxy process:

- `AIRJELLY_SUPABASE_ANON_KEY`: the public Supabase publishable key used by the
  installed AirJelly client. It is needed for password login, OAuth code exchange,
  and session refresh.
- `AIRJELLY_GATEWAY_HMAC_KEY`: the client's current gateway signing key in hex.
  Keep it in a private environment file; never put it in config.yaml, an auth
  file, logs, or source control. Client updates may rotate it.

The management UI offers Google OAuth with PKCE, email/password login, and
email registration. As in the desktop client, registration may return a session
immediately or require an emailed signup code. The UI can verify or resend that
code, then saves the resulting session. A first Google sign-in also creates an
AirJelly account through Supabase when the email is new. After sign-in, the
backend initializes the AirJelly user profile if it is absent.
Google returns to the desktop client's `http://localhost:19877` callback. Copy
the complete callback URL from the browser address bar and paste it into the
AirJelly card. The backend keeps the PKCE verifier for ten minutes and consumes
it once. The callback URL, password, and tokens are never returned to the UI
after exchange. A successful login stores an `airjelly-*.json` credential under
the configured auth directory. The existing credentials page can list, disable,
refresh, and delete it.

The account's `prefix` is `airjelly`. Servers with `routing.force-model-prefix:
true` call the model as `airjelly/airjelly-lite`. Servers without that setting
can also call `airjelly-lite`. The upstream request always uses `airjelly-lite`.

The quota card reads `GET /api/stripe/subscription` through the same signed
gateway and shows the plan, daily, monthly (when present), permanent, and total
remaining credits. A zero or absent monthly cap is not shown as a paid pool.
The session refreshes shortly before expiry using the Supabase refresh grant.

This integration targets the current desktop client protocol, which is not a
documented public developer API. If the desktop client changes its signing or
authentication contract, update the runtime key and verify the endpoints again.
