# CLI Proxy API

**English** | [简体中文](README_CN.md) | [日本語](README_JA.md)

A local proxy service that unifies your CLI subscription accounts (Codex / Claude / Gemini / Grok / Kimi, etc.) into OpenAI / Gemini / Claude compatible API endpoints, with multi-account round-robin load balancing.

Management frontend repository: [Cli-Proxy-API-Management-Center](https://github.com/rizxfrog/Cli-Proxy-API-Management-Center) (a web management panel that talks to this service's Management API).

## Supported Providers

| Provider | ID | Auth | Notes |
|---|---|---|---|
| Gemini | `gemini` | API Key | Google Gemini; configure `gemini-api-key` |
| Gemini Interactions | `gemini-interactions` | API Key | Native Google Interactions API; configure `interactions-api-key` |
| Vertex AI | `vertex` | API Key | Configure `vertex-api-key` |
| AI Studio | `aistudio` | WebSocket relay | Connected through the built-in WS relay (registered at runtime; no config needed) |
| Antigravity | `antigravity` | OAuth | `--antigravity-login` |
| Claude | `claude` | OAuth / API Key | `--claude-login` or `claude-api-key` |
| Codex (OpenAI) | `codex` | OAuth / API Key | `--codex-login`, `--codex-device-login`, or `codex-api-key` |
| xAI (Grok) | `xai` | OAuth / API Key | `--xai-login` or `xai-api-key` |
| Kimi | `kimi` | OAuth | `--kimi-login` |
| CodeBuddy CN | `codebuddy-cn` | OAuth / API Key | `--codebuddy-cn-login` or `codebuddy-cn-api-key` (gateway `copilot.tencent.com`) |
| CodeBuddy International | `codebuddy-ai` | OAuth / API Key | `--codebuddy-ai-login` or `codebuddy-ai-api-key` (gateway `www.codebuddy.ai`) |
| Devin | `devin` | OAuth | `--devin-login` |
| TRAE SOLO CN | `trae` | Desktop client credential | Configure `trae-api-key`; sign-in is initiated from the management panel |
| DuMate | `dumate` | Desktop client credential | install Dumate on windows and login it, use dumate-creds.py to export the json。i haven't find a fittable OAuth for dumate |
| DeepSeek Web | `deepseek-web` | Browser session | Configure `deepseek-web-api-key` (copy the userToken from chat.deepseek.com) |
| Qwen Web | `qwen-web` | Web login | Connected through the management panel Web login |
| OpenAI Compatible | `openai-compatibility` | API Key | Any OpenAI-compatible endpoint; configure `openai-compatibility` |

> See the comments in `config.example.yaml` for the exact fields and examples.

## Quick Start

### 1. Prepare the Config File

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml`. At minimum you need:

- `api-keys`: the key clients use to access this proxy (set your own).
- `remote-management.secret-key`: the management panel key (enter it in plaintext; it is hashed automatically at startup).
- `remote-management.allow-remote: true`: enable this when the management panel is accessed remotely.
- `auth-dir`: the directory for OAuth credentials (defaults to `~/.cli-proxy-api`; this project uses `data/auth_files`).

### 2. Run Locally

Requires Go 1.26+:

```bash
go build -o cli-proxy-api ./cmd/server
./cli-proxy-api --config config.yaml
```

Common flags:

| Flag | Description |
|---|---|
| `--config <path>` | Path to the config file |
| `--tui` | Start the terminal management UI |
| `--standalone` | Start an embedded local server in TUI mode |
| `--local-model` | Use only the embedded model catalogs; skip remote updates |
| `--no-browser` | Do not open the browser automatically during OAuth login |
| `--claude-login` / `--codex-login` / `--codex-device-login` / `--antigravity-login` / `--kimi-login` / `--codebuddy-cn-login` / `--codebuddy-ai-login` / `--xai-login` / `--devin-login` | OAuth login flow for the corresponding provider |

### 3. Add Accounts

OAuth login (the example below uses Codex; a browser opens to complete authorization):

```bash
./cli-proxy-api --config config.yaml --codex-login
```

After a successful login the credential is written to `auth-dir`. Add several accounts to enable automatic round-robin.

You can also configure each provider's API key directly in `config.yaml` (e.g. `gemini-api-key`, `codex-api-key`, `claude-api-key`, `openai-compatibility`). See the comments in `config.example.yaml` for details.

### 4. Call the API

The service listens on port `8317` by default. Use a key from `api-keys` and call it in an OpenAI-compatible way:

```bash
curl http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-5", "messages": [{"role": "user", "content": "hi"}]}'
```

It is also compatible with the Gemini and Claude (`/v1/messages`) protocols. Clients only need to point their base URL at `http://localhost:8317`.

## Docker Deployment

### Use the Prebuilt Image (Recommended)

The repository ships with a `docker-compose.yml`:

```bash
# Prepare the config and data directories
cp config.example.yaml config.yaml
mkdir -p data

# Start
docker compose up -d
```

Notes:

- Image: `rizxfrog/cli-proxy-api:latest`
- Port: `8317:8317`
- Mounts: `./config.yaml` → the config file inside the container; `./data` → the data directory inside the container (persists credentials, model catalogs, etc.)
- `docker-compose.yml` uses an external network named `1panel-network` by default; remove the `networks` section if you do not need it

You can also use the interactive script (choose a prebuilt image or build from source):

```bash
./docker-build.sh
```

### Build the Image from Source

```bash
docker build -t cli-proxy-api:local .
docker run -d --name cli-proxy-api \
  -p 8317:8317 \
  -v $(pwd)/config.yaml:/CLIProxyAPI/config.yaml \
  -v $(pwd)/data:/CLIProxyAPI/data \
  cli-proxy-api:local
```

There is no browser inside the container, so for OAuth login use the `--no-browser` flag and complete authorization by copying the link manually, or log in on the host first and mount the credential files from `auth-dir` into the container.

## Management Panel

Management frontend repository: **[Cli-Proxy-API-Management-Center](https://github.com/rizxfrog/Cli-Proxy-API-Management-Center)**

- The panel's static assets are downloaded automatically from that repository's releases (`remote-management.disable-control-panel` disables this).
- The `remote-management.panel-github-repository` option specifies the source repository for the panel.
- All management endpoints live under `/v0/management/*` and require the `remote-management.secret-key`.

## License

MIT License - see [LICENSE](LICENSE) for details.
