# CLI Proxy API

[English](README.md) | [简体中文](README_CN.md) | **日本語**

CLI のサブスクリプションアカウント（Codex / Claude / Gemini / Grok / Kimi など）を、OpenAI / Gemini / Claude 互換の API エンドポイントに統合するローカルプロキシサービスです。複数アカウントのラウンドロビンによる負荷分散に対応しています。

管理フロントエンドリポジトリ：[Cli-Proxy-API-Management-Center](https://github.com/rizxfrog/Cli-Proxy-API-Management-Center)（本サービスの Management API に対応する Web 管理パネル）。

## 対応 Provider 一覧

| Provider | ID | 認証方式 | 説明 |
|---|---|---|---|
| Gemini | `gemini` | API Key | Google Gemini。`gemini-api-key` を設定 |
| Gemini Interactions | `gemini-interactions` | API Key | ネイティブ Google Interactions API。`interactions-api-key` を設定 |
| Vertex AI | `vertex` | API Key | `vertex-api-key` を設定 |
| AI Studio | `aistudio` | WebSocket リレー | 内蔵 WS リレー経由で接続（実行時に登録。設定不要） |
| Antigravity | `antigravity` | OAuth | `--antigravity-login` |
| Claude | `claude` | OAuth / API Key | `--claude-login` または `claude-api-key` |
| Codex (OpenAI) | `codex` | OAuth / API Key | `--codex-login`、`--codex-device-login` または `codex-api-key` |
| xAI (Grok) | `xai` | OAuth / API Key | `--xai-login` または `xai-api-key` |
| Kimi | `kimi` | OAuth | `--kimi-login` |
| CodeBuddy 中国版 | `codebuddy-cn` | OAuth / API Key | `--codebuddy-cn-login` または `codebuddy-cn-api-key`（ゲートウェイ `copilot.tencent.com`） |
| CodeBuddy 国際版 | `codebuddy-ai` | OAuth / API Key | `--codebuddy-ai-login` または `codebuddy-ai-api-key`（ゲートウェイ `www.codebuddy.ai`） |
| Devin | `devin` | OAuth | `--devin-login` |
| TRAE SOLO CN | `trae` | デスクトップクライアント資格情報 | `trae-api-key` を設定。ログインは管理パネルから開始 |
| 百度搭子 | dumate | デスクトップ用の認証情報 | WindowsにDumateをインストールしてログインした後、スクリプトdumate-creds.pyでログイン認証情報のJSONをエクスポートします。dumateには現在適切なOAuthがありません |
| DeepSeek Web | `deepseek-web` | ブラウザセッション | `deepseek-web-api-key` を設定（chat.deepseek.com から userToken をコピー） |
| Qwen Web | `qwen-web` | Web ログイン | 管理パネルの Web ログイン経由で接続 |
| OpenAI 互換 | `openai-compatibility` | API Key | 任意の OpenAI 互換エンドポイント。`openai-compatibility` を設定 |

> 各項目の正確なフィールドと例は `config.example.yaml` のコメントを参照してください。

## クイックスタート

### 1. 設定ファイルを用意する

```bash
cp config.example.yaml config.yaml
```

`config.yaml` を編集します。最低限、次の項目が必要です：

- `api-keys`：クライアントが本プロキシへアクセスする際に使うキー（任意に設定）。
- `remote-management.secret-key`：管理パネルのキー（平文で記入。起動時に自動でハッシュ化されます）。
- `remote-management.allow-remote: true`：管理パネルへリモートアクセスする場合に有効化。
- `auth-dir`：OAuth 資格情報の保存ディレクトリ（既定は `~/.cli-proxy-api`。本プロジェクトでは `data/auth_files` を使用）。

### 2. ローカルで実行する

Go 1.26+ が必要です：

```bash
go build -o cli-proxy-api ./cmd/server
./cli-proxy-api --config config.yaml
```

よく使う引数：

| 引数 | 説明 |
|---|---|
| `--config <path>` | 設定ファイルのパスを指定 |
| `--tui` | ターミナル管理 UI を起動 |
| `--standalone` | TUI モードで組み込みローカルサーバーを起動 |
| `--local-model` | 内蔵モデルカタログのみを使用し、リモート更新を行わない |
| `--no-browser` | OAuth ログイン時にブラウザを自動で開かない |
| `--claude-login` / `--codex-login` / `--codex-device-login` / `--antigravity-login` / `--kimi-login` / `--codebuddy-cn-login` / `--codebuddy-ai-login` / `--xai-login` / `--devin-login` | 各プロバイダーの OAuth ログインフロー |

### 3. アカウントを追加する

OAuth ログイン（以下は Codex の例。ブラウザが開き、認可を完了します）：

```bash
./cli-proxy-api --config config.yaml --codex-login
```

ログインに成功すると資格情報が `auth-dir` に書き込まれます。アカウントを複数追加すると自動でラウンドロビンされます。

`config.yaml` に各プロバイダーの API Key を直接設定することもできます（`gemini-api-key`、`codex-api-key`、`claude-api-key`、`openai-compatibility` など）。詳細は `config.example.yaml` のコメントを参照してください。

### 4. API を呼び出す

サービスは既定でポート `8317` をリッスンします。`api-keys` のキーを使用し、OpenAI 互換の形式で呼び出します：

```bash
curl http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-5", "messages": [{"role": "user", "content": "hi"}]}'
```

Gemini や Claude（`/v1/messages`）などのプロトコルにも対応しています。クライアントは base URL を `http://localhost:8317` に向けるだけです。

## Docker デプロイ

### プリビルドイメージを使う（推奨）

リポジトリに `docker-compose.yml` が同梱されています：

```bash
# 設定ファイルとデータディレクトリを用意
cp config.example.yaml config.yaml
mkdir -p data

# 起動
docker compose up -d
```

補足：

- イメージ：`rizxfrog/cli-proxy-api:latest`
- ポート：`8317:8317`
- マウント：`./config.yaml` → コンテナ内の設定ファイル、`./data` → コンテナ内のデータディレクトリ（資格情報やモデルカタログなどを永続化）
- `docker-compose.yml` は既定で `1panel-network` という外部ネットワークを使用します。不要な場合は `networks` セクションを削除してください

対話式スクリプトを使うこともできます（プリビルドイメージかソースビルドを選択）：

```bash
./docker-build.sh
```

### ソースからイメージをビルドする

```bash
docker build -t cli-proxy-api:local .
docker run -d --name cli-proxy-api \
  -p 8317:8317 \
  -v $(pwd)/config.yaml:/CLIProxyAPI/config.yaml \
  -v $(pwd)/data:/CLIProxyAPI/data \
  cli-proxy-api:local
```

コンテナ内にはブラウザがないため、OAuth ログインでは `--no-browser` 引数を使ってリンクを手動でコピーして認可を完了するか、先にホスト側でログインしてから `auth-dir` の資格情報ファイルをコンテナへマウントしてください。

## 管理パネル

管理フロントエンドリポジトリ：**[Cli-Proxy-API-Management-Center](https://github.com/rizxfrog/Cli-Proxy-API-Management-Center)**

- パネルの静的アセットは、そのリポジトリの release からバックエンドが自動でダウンロードします（`remote-management.disable-control-panel` で無効化できます）。
- 設定項目 `remote-management.panel-github-repository` でパネルの取得元リポジトリを指定します。
- すべての管理エンドポイントは `/v0/management/*` にあり、`remote-management.secret-key` の付与が必要です。

## License

MIT License - 詳細は [LICENSE](LICENSE) を参照してください。
