# CLI Proxy API

[English](README.md) | **简体中文** | [日本語](README_JA.md)

一个本地代理服务，把你的 CLI 订阅账号（Codex / Claude / Gemini / Grok / Kimi 等）统一转换成 OpenAI / Gemini / Claude 兼容的 API 接口，支持多账号轮询负载均衡。

管理前端仓库：[Cli-Proxy-API-Management-Center](https://github.com/rizxfrog/Cli-Proxy-API-Management-Center)（Web 管理面板，对应本服务的 Management API）。

## 支持的 Provider

| Provider | 标识 | 认证方式 | 说明 |
|---|---|---|---|
| Gemini | `gemini` | API Key | Google Gemini，配置 `gemini-api-key` |
| Gemini Interactions | `gemini-interactions` | API Key | 原生 Google Interactions API，配置 `interactions-api-key` |
| Vertex AI | `vertex` | API Key | 配置 `vertex-api-key` |
| AI Studio | `aistudio` | WebSocket 中继 | 通过内置 WS 中继接入（运行时注册，无需配置） |
| Antigravity | `antigravity` | OAuth | `--antigravity-login` |
| Claude | `claude` | OAuth / API Key | `--claude-login` 或 `claude-api-key` |
| Codex (OpenAI) | `codex` | OAuth / API Key | `--codex-login`、`--codex-device-login` 或 `codex-api-key` |
| xAI (Grok) | `xai` | OAuth / API Key | `--xai-login` 或 `xai-api-key` |
| Kimi | `kimi` | OAuth | `--kimi-login` |
| CodeBuddy 中国版 | `codebuddy-cn` | OAuth / API Key | `--codebuddy-cn-login` 或 `codebuddy-cn-api-key`（网关 `copilot.tencent.com`） |
| CodeBuddy 国际版 | `codebuddy-ai` | OAuth / API Key | `--codebuddy-ai-login` 或 `codebuddy-ai-api-key`（网关 `www.codebuddy.ai`） |
| Devin | `devin` | OAuth | `--devin-login` |
| TRAE SOLO CN | `trae` | 桌面端凭据 | 配置 `trae-api-key`；登录通过管理面板发起 |
| 百度搭子 | `dumate` | 桌面端凭据 | 在Windows上安装Dumate并登录后，用脚本dumate-creds.py导出登录凭证json。dumate目前没有合适的OAuth |
| DeepSeek Web | `deepseek-web` | 浏览器会话 | 配置 `deepseek-web-api-key`（从 chat.deepseek.com 复制 userToken） |
| Qwen Web | `qwen-web` | 网页登录 | 通过管理面板 Web 登录接入 |
| OpenAI 兼容 | `openai-compatibility` | API Key | 任意 OpenAI 兼容端点，配置 `openai-compatibility` |

> 具体字段与示例见 `config.example.yaml` 注释。

## 快速开始

### 1. 准备配置文件

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，至少需要：

- `api-keys`：客户端访问本代理时使用的密钥（自定义）。
- `remote-management.secret-key`：管理后台的密钥（明文填写，启动时会自动哈希）。
- `remote-management.allow-remote: true`：需要远程访问管理面板时开启。
- `auth-dir`：OAuth 凭证存放目录（默认 `~/.cli-proxy-api`，本项目使用 `data/auth_files`）。

### 2. 本地运行

需要 Go 1.26+：

```bash
go build -o cli-proxy-api ./cmd/server
./cli-proxy-api --config config.yaml
```

常用参数：

| 参数 | 说明 |
|---|---|
| `--config <path>` | 指定配置文件路径 |
| `--tui` | 启动终端管理 UI |
| `--standalone` | TUI 模式下内嵌启动本地服务 |
| `--local-model` | 只使用内置模型表，不从远端更新 |
| `--no-browser` | OAuth 登录时不自动打开浏览器 |
| `--claude-login` / `--codex-login` / `--codex-device-login` / `--antigravity-login` / `--kimi-login` / `--codebuddy-cn-login` / `--codebuddy-ai-login` / `--xai-login` / `--devin-login` | 对应厂商的 OAuth 登录流程 |

### 3. 添加账号

OAuth 登录（示例以 Codex 为例，会打开浏览器完成授权）：

```bash
./cli-proxy-api --config config.yaml --codex-login
```

登录成功后凭证会写入 `auth-dir`，多放几个账号即可自动轮询。

也可以直接在 `config.yaml` 里配置各家 API Key（`gemini-api-key`、`codex-api-key`、`claude-api-key`、`openai-compatibility` 等），详见 `config.example.yaml` 注释。

### 4. 调用 API

服务默认监听 `8317` 端口，使用 `api-keys` 中配置的密钥，以 OpenAI 兼容方式调用：

```bash
curl http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer <你的 api-key>" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-5", "messages": [{"role": "user", "content": "hi"}]}'
```

同时兼容 Gemini、Claude（`/v1/messages`）等协议，客户端只需把 base URL 指向 `http://localhost:8317` 即可。

## Docker 部署

### 使用预构建镜像（推荐）

仓库自带 `docker-compose.yml`：

```bash
# 准备配置和数据目录
cp config.example.yaml config.yaml
mkdir -p data

# 启动
docker compose up -d
```

说明：

- 镜像：`rizxfrog/cli-proxy-api:latest`
- 端口：`8317:8317`
- 挂载：`./config.yaml` → 容器内配置文件，`./data` → 容器内数据目录（凭证、模型表等持久化）
- `docker-compose.yml` 默认使用名为 `1panel-network` 的外部网络，不需要时可删除 `networks` 部分

也可以使用交互脚本（可选预构建镜像或源码构建）：

```bash
./docker-build.sh
```

### 从源码构建镜像

```bash
docker build -t cli-proxy-api:local .
docker run -d --name cli-proxy-api \
  -p 8317:8317 \
  -v $(pwd)/config.yaml:/CLIProxyAPI/config.yaml \
  -v $(pwd)/data:/CLIProxyAPI/data \
  cli-proxy-api:local
```

Docker 容器内没有浏览器，OAuth 登录请使用 `--no-browser` 参数手动复制链接完成授权，或先在宿主机登录后将 `auth-dir` 中的凭证文件挂载进容器。

## 管理面板

管理前端代码仓库：**[Cli-Proxy-API-Management-Center](https://github.com/rizxfrog/Cli-Proxy-API-Management-Center)**

- 面板静态资源由后端自动从该仓库 release 下载（`remote-management.disable-control-panel` 可关闭）。
- 配置项 `remote-management.panel-github-repository` 指定面板来源仓库。
- 所有管理接口位于 `/v0/management/*`，请求需携带 `remote-management.secret-key`。

## License

MIT License - 详见 [LICENSE](LICENSE)。
