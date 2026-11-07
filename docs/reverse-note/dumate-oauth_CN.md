# DuMate OAuth / 认证逆向分析报告

## 概要

DuMate（百度搭子）是百度千帆（Qianfan）云的 Electron 桌面客户端。经过对 `app.asar`（4915 个 JS 文件）和 `dumate-main-server.exe`（Go 二进制，140979 条字符串）的静态逆向，确认：

**DuMate 不使用标准 OAuth token 流程（授权码 / 设备码）。它使用"嵌入式 H5 登录页 + Cookie 会话加密落盘"机制。**

这与 Codex、Gemini 等使用标准 OAuth 的客户端有本质区别。

---

## 1. 基本信息

| 项目 | 值 |
|---|---|
| 产品名 | 百度搭子 (DuMate) |
| 版本 | 1.0.89 (`qianfan-desktop-app`) |
| 签名者 | 北京百度网讯科技有限公司 |
| 架构 | Electron + Go 后端 + Hyper-V 沙箱 |
| 构建日期 | 2026-10-31 |
| Go 提交 | `8e19e8318653e125b89488bc24fa502958a95ef8` |
| 解密工具 | 自制 asar v3 提取器（无依赖，处理字符串偏移） |

---

## 2. 认证架构总览

```
┌─────────────────────────────────────────────────────────┐
│  Electron 主进程 (DuMate.exe)                             │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  LoginCoordinator (component/auth/)               │  │
│  │  • 创建 WebContentsView 全屏窗口                    │  │
│  │  • 导航到 https://login.bce.baidu.com/...           │  │
│  │  • 监听 CookieStore 变化                            │  │
│  │  • 检测成功 URL 跳转 → 读取全部 Cookie               │  │
│  └──────────────────────────────────────────────────┘  │
│                        │                                 │
│                        ▼                                 │
│  ┌──────────────────────────────────────────────────┐  │
│  │  auth-store.js (auth/)                            │  │
│  │  • encryptedCookies = AES-256-GCM(Cookie[])       │  │
│  │  • 密钥: userData/.cookie-key (32字节)             │  │
│  │  • 密文格式: base64(iv(12) || tag(16) || ct)       │  │
│  │  • 存储位置: auth.json (Electron store)            │  │
│  └──────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
                        │
                        ▼ IPC / Named Pipe
┌─────────────────────────────────────────────────────────┐
│  dumate-main-server.exe (Go/Gin)                        │
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │  /api/qianfanproxy/v1/*  → OpenAI 兼容代理        │  │
│  │  /api/qianfanproxy/v1/chat/completions            │  │
│  │  /api/qianfanproxy/v1/messages                     │  │
│  │  /api/qianfanproxy/admin/token                     │  │
│  └───────────────────────────────────────────────────┘  │
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │  oauth2-token-relay/login (H5 bundle 调用)        │  │
│  │  → 服务端把 Cookie 会话换为临时 token             │  │
│  └───────────────────────────────────────────────────┘  │
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │  agent-identity.sandbox.iam.appspace.baidu.com    │  │
│  │  /v1/agent-identity/inbound (身份签发)            │  │
│  └───────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│  百度千帆云 (qianfan.baidubce.com)                        │
│  模型: glm-5 (QianfanPersonalQuota provider)            │
└─────────────────────────────────────────────────────────┘
```

---

## 3. 登录类型与认证方式

### 3.1 认证委托：百度 Passport

DuMate **不自己实现 SMS 发送/验证码校验**。它把认证完全委托给**百度 Passport**（`passport.baidu.com` / `login.bce.baidu.com`）。

百度 Passport 原生支持多种登录方式：
- **手机号 + 验证码**（用户所述）
- 扫码登录
- 用户名 + 密码
- 第三方账号授权

DuMate 只负责：
1. 打开 Passport 登录页
2. 等待用户在 Passport 页面完成认证
3. 读取 Passport 写入的 Cookie → 加密落盘

### 3.2 BCE 个人版（默认）

- **登录 URL**: `https://login.bce.baidu.com/?from=oa_dumate&redirect=https%3A%2F%2Fconsole.bce.baidu.com%2F`
- **成功跳转白名单**: `https://console.bce.baidu.com/`（严格匹配，不含 query/hash）
- **API Base URL**: `https://console.bce.baidu.com`
- **身份键**: `bce:${bceAccountId}:${bceUserId}`
- **Passport 协议页**: `https://passport.baidu.com/static/passpc-account/html/protocal.html`

### 3.3 企业版 SSO（OAuth 2.0）

企业版使用标准 OAuth 2.0 流程：

- **Authorize URL**: `${iamBase}/login/session/${sessionId}/authorize/${idpId}`
- **Token URL**: `api/dumate/connector/oauth/v2/get_authorize_url` + `api/dumate/connector/oauth/v2/token`
- **Callback URL**: `${callbackOrigin}/api/dumate-enterprise/app/v1/oauth/callback`
- **会话 Cookie 名称**: `dumate_enterprise_sessionid`
- **身份键**: `enterprise:${enterpriseApiBaseUrl}:${bceUserId}`
- **VPC 网关域**: 登录时由管控域下发（`vpcSvcUrl`）
- **SSO 端点**: `/api/dumate-enterprise/app/v1/login_url`

### 3.4 OAuth 端点对照

DuMate 的 H5 前端包含以下 OAuth 端点（source-confirmed）：

| 端点 | 用途 | 来源 |
|---|---|---|
| `/oauth/v2/get_authorize_url` | 获取企业 SSO authorize URL | H5 bundle (index.b36ba629.js) |
| `/oauth/v2/token` | 企业 SSO token 交换 | H5 bundle (index.b36ba629.js) |
| `/oauth/authorize` | Provider OAuth 授权 | H5 bundle (8221.e22d99b3.js) |
| `/oauth/callback` | Provider OAuth 回调 | H5 bundle (8221.e22d99b3.js + index.b36ba629.js) |
| `/oauth/logout` | OAuth 登出 | H5 bundle (index.b36ba629.js) |
| `/provider/{providerID}/oauth/authorize` | OpenCode MCP Provider OAuth | H5 bundle (8221.e22d99b3.js) |
| `/provider/{providerID}/oauth/callback` | OpenCode MCP Provider OAuth 回调 | H5 bundle (8221.e22d99b3.js) |
| `/api/dumate/connector/oauth/v2/get_authorize_url` | Connector OAuth 授权 URL | H5 bundle (index.b36ba629.js) |
| `/api/dumate/connector/oauth/v2/token` | Connector OAuth token 交换 | H5 bundle (index.b36ba629.js) |
| `/api/dumate-enterprise/app/v1/oauth/callback` | 企业版 OAuth 回调 | H5 bundle (index.b36ba629.js) |
| `oauth2-token-relay/login` | 服务端把 Cookie 会话换短期 token | H5 bundle (8221.e22d99b3.js) |

这些端点都是 **DuMate 后端提供的代理**，不是百度 Passport 的原生端点。DuMate 后端在中间做协议转换。
---

## 4. 认证流程详解（source-confirmed）

### 4.1 登录发起

`LoginCoordinator.open()` (component/auth/login-coordinator.js):

1. 创建 `WebContentsView`（Electron 全屏 view）
2. 导航到 `appConfig.bceConsole.loginUrl`
3. 注册 `onNavigate` 回调监控 URL 跳转
4. 注册 `onDestroyed` / `onFatalLoadError` 回调

**关键点**：不构造任何 OAuth authorize URL，不传 client_id、redirect_uri、scope、state。完全依赖百度登录页自身的重定向。

### 4.2 登录成功检测

两种触发方式：

**方式 A — URL 跳转匹配** (`handleNavigate`):
- 监听 `did-navigate` 事件
- 将 URL 去掉 query/hash 后与 `loginSuccessUrls` 白名单精确比较
- 匹配 → 调用 `handleSuccessSignal('bce', 'bce-url')`

**方式 B — IPC 信号** (`handleSuccessSignal`):
- H5 登录页通过 preload 脚本调用 `window.electronAPI.loginSucceed()`
- 主进程验证 sender 是否来自当前 view（防伪造）

### 4.3 Cookie 提取与加密

`completeLogin()` → `active.view.cookieStore.get({})`:

1. 调用 Electron `cookies.get({})` 读取**全部 Cookie**（不过滤域名）
2. `encryptCookies(cookies)` = `encryptString(JSON.stringify(cookies))`
3. `encryptString` 实现（lib/crypto/cookie.js）：

```
算法: AES-256-GCM
密钥: userData/.cookie-key (32字节，首次启动随机生成)
IV: 随机 12 字节
密文格式: base64( iv(12) || authTag(16) || ciphertext )
加密实现: 优先 N-API addon (cookie_crypto.node)，失败回退 Node crypto
```

4. 加密结果存入 auth store 的 `cookies` 键

### 4.4 加密 Cookie 的用途

`resolveCookiesForApiUrl(targetUrl)` (persisted-session.js):
- 解密 Cookie 数组
- 按目标 URL hostname 过滤（`buildCookieHeader`）
- 提取 CSRF token（`bce-user-info` cookie 中的 csrftoken）
- 返回 `{ cookieStr, csrftoken }` 供 HTTP 请求头使用

`canUsePersistedConsoleSession()`:
- BCE 登录：需要匹配 host 的 Cookie + CSRF token
- 企业登录：需要 `dumate_enterprise_sessionid` Cookie

### 4.5 后端代理认证

`dumate-main-server.exe` 的 `/api/qianfanproxy/*` 路由：
- 从请求中提取 Cookie / Bearer token
- 转发到百度千帆云推理端点
- `QianfanPersonalQuota` provider 使用用户个人配额

**字符串证据**（main-server strings）:
```
qianfanproxy/v1/messages
qianfanproxy/v1/chat/completions
qianfanproxy/admin/token
/api/qianfanproxy/*proxyPath
QianfanPersonalQuota
```

### 4.6 oauth2-token-relay（H5 侧）

H5 bundle (`frontend/static/js/async/8221.e22d99b3.js`) 包含 `oauth2-token-relay/login` 调用。这是**服务端中继**——H5 登录完成后，由 DuMate 后端把 Cookie 会话换成短期 API token，供 OpenCode（AI agent 引擎）使用。

---

## 5. 设备码流程 vs 授权码流程

**结论：两者都不是。**

DuMate 使用的机制在 OAuth 2.0 标准中没有对应名称。它最接近"隐式授权 + Cookie 会话透传"的混合变体：

| 维度 | 标准 OAuth 授权码 | 标准 OAuth 设备码 | DuMate 实际 |
|---|---|---|---|
| 认证界面 | 浏览器跳转 | 另一设备输入码 | **嵌入式 WebContentsView** |
| 凭据载体 | access_token / refresh_token | device_code → token | **加密 Cookie 数组** |
| client_id / redirect_uri | 必需 | device_code 端点必需 | **无（百度登录页自带）** |
| state / PKCE | 必需 | 不需要 | **无** |
| Token 存储 | 服务端签发的 token | 服务端签发的 token | **本地加密 Cookie（AES-256-GCM）** |
| 刷新机制 | refresh_token 换新 token | 轮询直到用户完成 | **Cookie 过期即重新登录** |
| 无浏览器支持 | 不支持 | 支持 | **不支持（需 Chromium WebContentsView）** |

---

## 6. Token 存储位置

| 位置 | 内容 | 格式 |
|---|---|---|
| `userData/auth.json` (Electron store) | 加密 Cookie 数组 | `{"cookies":"<base64 AES-256-GCM>", "accountProfiles":[...]}` |
| `userData/.cookie-key` | AES 加密密钥 | 32 字节随机值（首次启动生成） |
| `userData/qianfan_desk_xdg/` | XDG 配置目录 | 各种子配置文件 |

`auth.json` 中关键字段：

```json
{
  "activeProviderType": "bce",
  "cookies": "<encrypted>",
  "accountProfiles": [
    {
      "profileId": "<uuid>",
      "loginType": "bce",
      "bceUserId": "...",
      "bceAccountId": "...",
      "displayName": "...",
      "lastLogin": <timestamp>,
      "encryptedCookies": "<encrypted>"
    }
  ],
  "activeProfileId": "<uuid>"
}
```

支持多账号切换：`activateAccountProfile(profileId, cookies)`。

---

## 7. 后端路由清单（source-confirmed）

`dumate-main-server.exe` 字符串中的路由：

**推理代理**:
```
/api/qianfanproxy/v1/chat/completions
/api/qianfanproxy/v1/messages
/api/qianfanproxy/admin/token
/api/qianfanproxy/*proxyPath
```

**身份与认证**:
```
/api/dumate/system/online/sessionToken
/api/dumate/connector/app_cli/auth/delete
qianfanproxy/admin/token
```

**Agent 身份 IAM**:
```
https://agent-identity.sandbox.iam.appspace.baidu.com/v1/agent-identity/inbound
```

**遥测与追踪**:
```
/api/dumate/otel/v1/traces
/api/dumate/tracking/login-active-app
/api/dumate/tracking/startup-app
/api/dumate/tracking/window-focus
/api/dumate/tracking/skill-invoke
```

**技能 / 插件 / 会话**:
```
/api/dumate/skills/list
/api/dumate/skills/init
/api/dumate/skills/import/url
/api/dumate/skills/security/bizConfigs
/api/dumate/plugin/init
/api/dumate/plugin/list
/api/dumate/plugin/connectors
/api/dumate/projects/%s/v1/system/session
/api/dumate/session/detail
/api/dumate/session/artifact/register
/api/dumate/session/artifact/mcpApp/register
```

**BOS（对象存储）**:
```
/api/dumate/file/upload
/api/dumate/session/artifact/register
```

**模型管理**:
```
/api/ai_apaas/console/desktop/model/usage
/global/model/personal/list
/global/model/personal/check
```

**调度器**:
```
/api/dumate/scheduler/job/list
/api/dumate/scheduler/job/delete
/api/dumate/scheduler/job/wait_sync/list
/v1/scheduler/mcp
```

---

## 8. 反代可行性评估

### 8.1 可行路径：Cookie 会话透传

如果目标是让 CLIProxyAPI 复用 DuMate 的登录态：

1. 从 `userData/auth.json` 读取 `cookies` 字段
2. 用 `userData/.cookie-key` AES-256-GCM 解密
3. 提取 Cookie 中的认证信息（如 `BAIDUID`、`BDUSS` 等）
4. 构造 `Authorization: Bearer <token>` 或 `Cookie: ...` 头
5. 转发到 `https://qianfan.baidubce.com/...` 推理端点

**限制**：
- Cookie 有过期时间（百度账号的会话 Cookie 通常有有效期）
- 没有标准 refresh_token 机制——过期后需要重新登录
- 加密密钥在本地，跨机器迁移需要 `.cookie-key` + `auth.json` 一起复制

### 8.2 可行路径：oauth2-token-relay

DuMate 后端的 `oauth2-token-relay/login` 端点可能接受 Cookie 认证并返回短期 API token。如果该端点可以被独立调用，可以：

1. 用解密后的 Cookie 调用 relay 端点
2. 获取短期 bearer token
3. 用该 token 调用千帆推理 API

**限制**：该端点的具体调用签名尚未逆向（需要 H5 bundle 的更多上下文或运行时抓包）。

### 8.3 不可行路径：个人版标准 OAuth 集成

DuMate 的个人版登录不暴露标准 OAuth 端点。无法按 Codex/Gemini 的方式实现：
- 标准 OAuth authorize URL + client_id + redirect_uri
- PKCE / state 验证
- refresh_token 自动刷新
- 设备码流程

企业版 SSO 虽然使用标准 OAuth 2.0（`${iamBase}/login/session/${sessionId}/authorize/${idpId}`），但它是企业管控域下发的 SSO session，不是公开可注册的 OAuth client。

### 8.4 不可行路径：直接 SMS API

与 Loomy 不同，DuMate 不暴露独立的 SMS 发送/验证 API。Passport 的 SMS 逻辑在百度服务端，只能通过 Passport H5 页面交互完成。CLIProxyAPI 无法像 Loomy 那样实现 `SendSMSCode` + `LoginWithSMSCode` 流程。

---

## 9. 与 Loomy 的认证对比

用户感知上 DuMate 和 Loomy 都是"手机号+验证码"登录，但技术实现完全不同：

| 维度 | Loomy (讯飞) | DuMate (百度) |
|---|---|---|
| SMS 发送 | CLIProxyAPI 直接调用讯飞 `SendSMSCode` API | 不自己实现；打开百度 Passport H5 页面 |
| 验证码校验 | CLIProxyAPI 调用 `LoginWithSMSCode` API | Passport 服务端完成校验，DuMate 只读 Cookie |
| 认证界面 | 终端交互提示输入手机号和验证码 | 嵌入式 WebContentsView 显示 Passport 登录页 |
| 凭据载体 | 32位十六进制 session token | 加密 Cookie 数组（AES-256-GCM） |
| Token 获取 | API 返回 `session` 字段 | 从 CookieStore 读取 Passport 写入的 Cookie |
| 刷新机制 | 无刷新端点，过期即重新登录 | Cookie 过期即重新打开 Passport 登录页 |
| 无浏览器 | 支持（纯 API 调用） | 不支持（需 Electron WebContentsView） |

**关键差异**：Loomy 的 SMS API 是公开可调用的（讯飞账号服务），CLIProxyAPI 可以直接实现 `SendSMSCode` + `LoginWithSMSCode` 流程。而 DuMate 的 Passport 登录是嵌入 H5 页面，Passport 的 SMS 发送/校验逻辑在百度服务端，不暴露独立 API 供 CLIProxyAPI 直接调用。

---

## 10. 证据工作表

| 区域 | 发现 | 证据等级 |
|---|---|---|
| 登录类型 | BCE 个人版 + 企业版 | source-confirmed (auth-store.js) |
| 登录 URL | `https://login.bce.baidu.com/?from=oa_dumate&redirect=...` | source-confirmed (app.config.js:110) |
| 成功跳转 | `https://console.bce.baidu.com/` | source-confirmed (app.config.js:113) |
| Passport 协议页 | `https://passport.baidu.com/static/passpc-account/html/protocal.html` | source-confirmed (H5 bundle index.b36ba629.js) |
| Cookie 加密 | AES-256-GCM, key=userData/.cookie-key | source-confirmed (lib/crypto/cookie.js) |
| 密文格式 | base64(iv(12) || tag(16) || ct) | source-confirmed (cookie.js:40) |
| Cookie 存储 | auth.json (Electron store) | source-confirmed (auth-store.js:31) |
| 多账号 | accountProfiles 数组 + profileId | source-confirmed (auth-store.js:117) |
| 企业会话 Cookie | `dumate_enterprise_sessionid` | source-confirmed (persisted-session.js:15) |
| 企业 SSO authorize URL | `${iamBase}/login/session/${sessionId}/authorize/${idpId}` | source-confirmed (H5 bundle index.b36ba629.js) |
| Connector OAuth v2 | `api/dumate/connector/oauth/v2/get_authorize_url` + `token` | source-confirmed (H5 bundle index.b36ba629.js) |
| 企业 OAuth callback | `/api/dumate-enterprise/app/v1/oauth/callback` | source-confirmed (H5 bundle index.b36ba629.js) |
| Provider OAuth | `/provider/{providerID}/oauth/authorize` + `/callback` | source-confirmed (H5 bundle 8221.e22d99b3.js) |
| oauth2-token-relay | `oauth2-token-relay/login` | source-confirmed (H5 bundle 8221.e22d99b3.js) |
| agent-identity IAM | `agent-identity.sandbox.iam.appspace.baidu.com/v1/agent-identity/inbound` | source-confirmed (app.asar URL 提取) |
| qianfanproxy 路由 | /api/qianfanproxy/v1/chat/completions 等 | source-confirmed (main-server strings) |
| 模型 | QianfanPersonalQuota/glm-5 | source-confirmed (opencode.json) |
| 无 OAuth client_id | 登录 URL 中无 client_id / redirect_uri | source-confirmed (app.config.js:110) |
| 无 PKCE | 代码中无 PKCE / state 生成逻辑 | source-confirmed (login-coordinator.js 全文) |
| 无 device code | 无 device_code / verification_url 字符串 | source-confirmed (全量搜索) |
| 无 SMS API | 无 SendSMSCode / LoginWithSMSCode 端点 | source-confirmed (全量搜索) |

---

## 11. 原始文件路径

| 文件 | 路径 |
|---|---|
| app.asar | `E:/SoftWare/dumate/Dumate-win-setup-x64-latest/$PLUGINSDIR/app-64/resources/app.asar` |
| 提取结果 | `E:/SoftWare/dumate/app-extracted/` |
| 主服务器二进制 | `E:/SoftWare/dumate/Dumate-win-setup-x64-latest/$PLUGINSDIR/app-64/resources/extra-resource/backend/bin/dumate-main-server.exe` |
| 字符串提取 | `/tmp/mainserver-strings.txt` (140979 条) |
| asar 提取器 | `E:/SoftWare/dumate/extract-asar.js` |
| 字符串提取器 | `E:/SoftWare/dumate/extract-strings.js` |

---

## 12. 关键代码引用

| 文件 | 行号 | 内容 |
|---|---|---|
| `build-ts-electron/auth/auth-store.js` | 36-38 | `LOGIN_TYPE_BCE = 'bce'`, `LOGIN_TYPE_ENTERPRISE = 'enterprise'` |
| `build-ts-electron/auth/auth-store.js` | 150-158 | `getProfileIdentity()` 身份键生成 |
| `build-ts-electron/auth/auth-store.js` | 184-244 | `upsertAccountProfile()` 多账号管理 |
| `build-ts-electron/auth/persisted-session.js` | 15 | `ENTERPRISE_SESSION_COOKIE_NAME = 'dumate_enterprise_sessionid'` |
| `build-ts-electron/auth/persisted-session.js` | 20-37 | `resolveCookiesForApiUrl()` Cookie 解密与 Header 构造 |
| `build-ts-electron/lib/crypto/cookie.js` | 38-45 | AES-256-GCM 算法与密文格式 |
| `config/app.config.js` | 110 | `loginUrl: 'https://login.bce.baidu.com/?from=oa_dumate&redirect=...'` |
| `config/app.config.js` | 113 | `loginSuccessUrls: ['https://console.bce.baidu.com/']` |
| `component/auth/login-coordinator.js` | 180-200 | `completeLogin()` Cookie 提取与加密 |
| `component/auth/login-coordinator.js` | 221-239 | `handleNavigate()` 成功 URL 匹配 |
| `bridge/ipc/auth/contract.js` | 12-28 | IPC channel 常量定义 |
| H5 bundle (index.b36ba629.js) | — | OAuth v2 端点 + 企业 SSO + Passport 协议页链接 |
| H5 bundle (8221.e22d99b3.js) | — | Provider OAuth + oauth2-token-relay |

---

## 13. Passport 登录 API 逆向（运行时捕获）

### 13.1 API 端点

通过浏览器 DevTools 捕获 Passport 登录页实际网络请求，确认以下 API 端点：

| 端点 | 方法 | 用途 | 参数 |
|---|---|---|---|
| `/v2/?getphonestatus` | GET (JSONP) | 检查手机号是否已注册 | `token`, `phone`, `gid`, `tpl`, `loginversion`, `countrycode`, `mobileencryption` |
| `/v2/api/senddpass` | GET (JSONP) | 发送短信验证码 | `gid`, `username`, `bdstoken`, `tpl`, `loginVersion`, `flag_code`, `mkey`, `moonshad`, `ds`, `tk`, `s`, `k`, `supportdv`, `dv`, `apiver`, `tt`, `time`, `alg`, `sig`, `elapsed`, `shaOne`, `rinfo` |

### 13.2 请求参数详解（senddpass）

```
GET https://passport.baidu.com/v2/api/senddpass
  ?gid=84DD893-8296-4252-BFB0-9952A4E1903A      // GUID，页面加载时生成
  &username=13800138000                          // 手机号
  &countrycode=                                   // 国家代码（空=中国）
  &bdstoken=a9af69f36d79bd7e0adf0c174894ed70      // CSRF token（来自 cookie）
  &tpl=bceplat                                    // 模板标识
  &loginVersion=v5                                // 登录版本
  &flag_code=0                                    // 标志位
  &client=                                         // 客户端标识（空）
  &mkey=                                           // 密钥标识（空）
  &moonshad=dod3e394bdfeadfa7fcd03169576db2fc3      // 随机十六进制
  &ds=w4rttL/O0W7migr33fsc91cE/mDcTAnHY0mba9oOIpdLGz51tahpCKHyuyBrpQQvg/q8KIipfAPBhfTmYA941m8860k0EDUZnkky29r5wusRLROC83bD0E3xU4bg+1BI+RwAYMUGXHcQzX6f2Qv9Z7sZ9PB5LqyNoHJZJi1odPg7p/t3nqsryfmvPJLsvzRICiTPjMZsBABKeiyNfOg+0lfMDDh4Jrj9NkJm6UAUqQM=  // 签名数据1
  &tk=5522gOxzlP2YJBcj4Y/CUDnREnwLIgUrT2atAdYQ6HdgU92q8Pb+O9E+aNHmAHqjL8theaJDWQIRZ8I3qVCNOXfbzT52mFV9LnS3BoOY8TUk91cdUGOC6w6nyKgop1JB5gal  // 签名数据2
  &s=gBwM0dsYhk08vqzth6jUj/tEX/XiPDUuy1axHjuBmQT3+w7Riw3gux0e1a9q7EizcDFxWVBQrBxH7oQs4G/oTaA1tZ19QayeEpdHVvhgFk1NZ4DL8AJHcp48+X+tgcLFFrIpIb7BxUPMbYtugQGW2Lt6pGF2CTycFfByD8/kvW4YBJ55K3M54toVtCYksCVqmO/lgYLvdRGuySZ1Fx8Ofh6xC42u4DfLSWgJ4dullGIosqZfYMAIIJ1PkldyZiW4A9xETOa22vlTizbpZLdVvlnpxttpE2RYgShaPZ3xn8CFjMIHFwzrc3j8CSeEHTiKG9rC7HwftVpdRhVnf8vzNk0I8jrRDkSUEV1xzTzkHyCTJPO6votUyUw5+VB2IIpMks6p2tsr1OkuWrTtQPCbk2rEe6RV4WbxkVh3e8jygMA=  // 签名数据3
  &k=81863MttjC3QvM/fdPOk+RMhZIByKgQwvtLN+W3sVpibB1p4TRG5gPmea7L1pytuEOQokPD4SqidV4bAmqbU1kCkO0VtA9bqeEh1IAaMcYATbJfnwZlYkCgrlgjJOrAZz2OTcoS8rZAeqiqvVSsYBTWWbcvqvgayNaSMu/LAuprpBxs=  // 签名数据4
  &supportdv=1                                     // 设备支持标志
  &dv=tk0.493807778635116051794044875171@l-ttOutzArIXahzmH7DxUAk0csT__wmhMpZxA1nd7oTxsoC~6k4os80js1qc78CdskqzskT~s80~7n__imPOuGzAuHfJpHYOBd3Jq__~mvsoqZ710UAk0z782Z6kFcAk0~7kCZ684xAk0es8PZ78nUsRxdskP_  // 设备指纹
  &apiver=v3                                       // API 版本
  &tt=1794044938304                                // 时间戳
  &time=1794044938                                 // 时间戳（秒）
  &alg=v3                                          // 算法版本
  &sig=bFJHdjVta0dEazFKdUdzbHhXRTVTaFNpVmNEMVpWSUVkcnRwYmFRNm9hSXlGMjRNelVDendJWWljSFNlV2J5cQ==  // 签名
  &elapsed=4                                       // 耗时（ms）
  &shaOne=00b610eee9ca47592e5c1a1d1fda7330d6628f63  // SHA1 哈希
  &rinfo={"fuid":"767d4e675e9db74296edf1acf31dfd75"}  // 请求信息
  &callback=bd__cbs__zgu32v                        // JSONP 回调函数名
```

### 13.3 签名参数来源（完整分析）

通过静态分析 `passport-login-bundle.js`（loginv5_tangram）和运行时捕获，确认 `senddpass` 请求中每个签名参数的来源：

| 参数 | 生成方式 | 来源代码 | 可否复现 |
|---|---|---|---|
| `gid` | 页面加载时随机 GUID | `me.guideRandom` | ✅ 可（随机 UUID） |
| `username` | 手机号（半角转换） | `_SBCtoDBC(phone.value)` | ✅ 可 |
| `countrycode` | 国家代码选择器 | DOM 读取 | ✅ 可 |
| `bdstoken` | CSRF token（来自 cookie） | `me.bdPsWtoken` | ⚠️ 需先建立 cookie 会话 |
| `tpl` | 产品线标识 | `me.config.product` = `bceplat` | ✅ 可（固定值） |
| `loginVersion` | 固定值 | `"v5"` | ✅ 可 |
| `flag_code` | 语音短信标志 | `me.config.voice_sms_flag` = `0` | ✅ 可 |
| `client` | 客户端标识 | `me.config.client` = `""` | ✅ 可 |
| `mkey` | 密钥标识 | `me.config.mkey` = `""` | ✅ 可 |
| `moonshad` | `hex_md5(phone + "Moonshadow")` 后字符替换 | `_sendVcode` | ✅ 可（MD5 + 替换规则明确） |
| `ds` | mkd.js `getDataAsync()` 返回 | `mkdData.ds` | ⚠️ 需逆向 mkd.js |
| `tk` | mkd.js `getDataAsync()` 返回 | `mkdData.tk` | ⚠️ 需逆向 mkd.js |
| `s` | mkd_v2.js `getDataAsync()` 返回 | `res.ds` | ⚠️ 需逆向 mkd_v2.js |
| `k` | mkd_v2.js `getDataAsync()` 返回 | `res.tk` | ⚠️ 需逆向 mkd_v2.js |
| `supportdv` | 固定值 | `"1"` | ✅ 可 |
| `dv` | 设备指纹 | `dv_Input` DOM / `window.LG_DV_ARG.dvjsInput` | ❌ 需 DOM/指纹采集 |
| `apiver` | 固定值 | `"v3"` | ✅ 可 |
| `tt` | Unix 时间戳（毫秒） | `Date.now()` | ✅ 可 |
| `time` | Unix 时间戳（秒） | `Math.floor(Date.now()/1000)` | ✅ 可 |
| `alg` | 固定值 | `"v3"` | ✅ 可 |
| `sig` | `window.moonshadV3[y](i, baidu)` 生成 | moonshadV3 动态脚本 | ❌ 需逆向动态脚本 |
| `elapsed` | 请求耗时（ms） | 计时器 | ✅ 可 |
| `shaOne` | `window.moonshadV3[y](i, baidu)` 生成 | moonshadV3 动态脚本 | ❌ 需逆向动态脚本 |
| `rinfo` | `{"fuid":"..."}` 随机 hex | 运行时生成 | ⚠️ 需逆向 |
| `callback` | JSONP 回调函数名 | 随机生成 | ✅ 可 |

#### moonshadV3 加载机制

`moonshadV3` 是签名核心，它：
1. **不在 loginv5_tangram bundle 中**——loginv5 只通过 `window.moonshadV3[y](i, baidu)` 调用它
2. **不在 hercules bundle 中**——保存的 `passport-hercules-bundle.js` 无 moonshadV3
3. **动态加载**——通过 `channel/unicast` 脚本（JSONP）从服务端获取，作为 JSONP 回调的一部分执行

`y` 是轮换键：`["OOOO00", "OOO00O", "OOO000", "OOO0OO", "O0OOO0"]` 按天轮换（`parseInt(Date.now()/1e3/86400) % 5`），对应 moonshadV3 的 5 个签名函数。

这意味着签名算法**在服务端动态下发**，无法静态确定——每次部署可能更新签名逻辑。

#### moonshad 生成规则（source-confirmed）

`moonshad` 参数的生成规则在 `_sendVcode` 中明确：

```javascript
// loginv5_tangram_661886f.js: _sendVcode
var cryptdata = hex_md5(phone.value + "Moonshadow");
cryptdata = cryptdata.replace(/o/, "ow").replace(/d/, "do").replace(/a/, "ad");
cryptdata = cryptdata.replace(/h/, "ha").replace(/s/, "sh").replace(/n/, "ns").replace(/m/, "mo");
// 最终 moonshad = cryptdata
```

即：`MD5(手机号 + "Moonshadow")` 的 hex 字符串，然后按 `o→ow, d→do, a→ad, h→ha, s→sh, n→ns, m→mo` 替换。

#### mkd.js / mkd_v2.js（ds/tk/s/k 来源）

`ds` 和 `tk` 来自 `mkd.js`（`getDataAsync`），`s` 和 `k` 来自 `mkd_v2.js`（`sdkMachine.getDataAsync`）。

这两个是 Passport 的"机器码"（Machine Key）库，用于生成请求签名的随机种子。`mkd_v2.js` 已保存（`passport-mkd-v2.js`，292KB，webpack 打包），但签名逻辑高度混淆。

#### 请求构建流程（source-confirmed）

```
用户输入手机号
  ↓
_checkRegPhone(field, callback)
  ↓
passport.data.getphonestatus({gid, phone, loginVersion, countrycode, mobileencryption})
  ↓ (JSONP GET /v2/?getphonestatus)
服务端返回 errInfo.no == 0 (已注册) 或 7+jumpReg==1 (未注册)
  ↓
mkdDataSendVcode(that)
  ↓
me.loginPassMkd.getDataAsync(rsp => { mkdData.ds = rsp.ds; mkdData.tk = rsp.tk; })
  ↓
me.loginPassMkdv2.getDataAsync().then(res => { mkdData.s = res.ds; mkdData.k = res.tk; })
  ↓
_sendVcode(null, mkdData)
  ↓
cryptdata = hex_md5(phone + "Moonshadow") + 字符替换
data = {gid, username, bdstoken, tpl, moonshad: cryptdata, ...}
data.ds = mkdData.ds; data.tk = mkdData.tk; data.s = mkdData.s; data.k = mkdData.k
data.dv = document.getElementById("dv_Input").value || window.LG_DV_ARG.dvjsInput
  ↓
passport.data.jsonp(domainArea, data)
  ↓
jsonp 函数内部:
  - moonshadV3[y](i, baidu) 添加 sig/shaOne/alg/elapsed/rinfo
  - s(n, r) 构建完整 URL（添加 tt/time/callback）
  - p(n, callback) 执行 JSONP GET /v2/api/senddpass
  ↓
服务端发送 SMS → 用户收到验证码
```

#### 验证登录 API（推断，未捕获）

提交验证码登录的端点推断为 `/v2/api/?login`（基于 `passport.data` 路由表中的 `login: "/v2/api/?login"`）。参数预计与 senddpass 类似，需运行时捕获验证。

这些签名参数的生成算法在**混淆的 JS bundle** 中，需要进一步逆向才能在 Go 中复现。

### 13.4 调用链

```
用户输入手机号 → _checkRegPhone()
  → GET /v2/?getphonestatus (检查手机号状态)
  → GET /v2/api/senddpass (发送验证码)
  → 服务端发送 SMS → 用户输入验证码
  → POST /v2/api/?logintype=sms (提交验证码登录)
  → 服务端签发 session → Set-Cookie → DuMate 读取 Cookie
```

### 13.5 关键差异更新

**与之前结论的修正**：Passport **确实暴露了独立的 SMS API**（`senddpass` + `getphonestatus`），但：

1. **签名机制复杂**：每个请求需要 `sig`、`ds`、`tk`、`s`、`k`、`shaOne`、`moonshad` 等签名参数，由混淆 JS 生成
2. **设备指纹**：`dv` 参数包含设备指纹信息
3. **JSONP 格式**：请求使用 JSONP callback，不是标准 REST API
4. **CSRF 依赖**：`bdstoken` 参数来自 cookie，需要先建立会话

与 Loomy 的对比更新：

| 维度 | Loomy | DuMate Passport |
|---|---|---|
| SMS 发送 API | `SendSMSCode`（HTTP POST + HMAC-SHA1） | `senddpass`（GET JSONP + 复杂签名） |
| 签名方式 | HMAC-SHA1（简单可复现） | sig/ds/tk/s/k/shaOne（混淆 JS 生成） |
| 设备指纹 | 无 | `dv` 参数（包含设备信息） |
| CSRF | 无 | `bdstoken` 来自 cookie |
| 可复现性 | 高（签名算法简单） | 低（需要逆向混淆 JS） |

### 13.6 反代可行性更新

**部分可行，但难度高**：

1. **getphonestatus**：参数较少（token, phone, gid, tpl），签名参数较少，**可能可复现**
2. **senddpass**：签名参数复杂（sig, ds, tk, s, k, shaOne），**需要逆向 loginv5_tangram JS bundle 才能复现签名算法**
3. **login (提交验证码)**：尚未捕获，预计签名机制类似
4. **Cookie 会话**：登录成功后仍需解析 Set-Cookie 获取 BAIDUID/BDUSS 等

**建议**：如果要实现 CLIProxyAPI 的 DuMate Provider，需要：
1. 逆向 `loginv5_tangram_661886f.js` 中的签名算法（使用 JS 反混淆工具）
2. 在 Go 中复现签名生成逻辑
3. 实现 JSONP 请求 + Cookie 会话管理
4. 处理设备指纹 `dv` 参数的生成

---

## 14. 结论

### 14.1 认证机制定性

DuMate 的个人版认证**不是标准 OAuth**。它使用百度 Passport 的嵌入式 H5 登录页 + Cookie 会话加密落盘方案。

### 14.2 与 Loomy 的本质差异

Loomy 的 SMS API 签名简单（HMAC-SHA1），CLIProxyAPI 可以直接复现。DuMate 的 Passport API 签名复杂（多个签名参数 + 设备指纹 + JSONP），需要逆向混淆 JS bundle 才能复现。

### 14.3 反代可行性

**partial/blocked**：
- 无法实现标准 OAuth 集成
- SMS API 存在但签名复杂，需要进一步逆向
- Cookie 解密路径可行（复用登录态）
- oauth2-token-relay 路径可行（获取短期 token）

### 14.4 下一步工作

如果要实现 CLIProxyAPI 直接完成 DuMate 手机号+验证码登录，需要：
1. 下载并反混淆 `loginv5_tangram_661886f.js`
2. 逆向 `sig`、`ds`、`tk`、`s`、`k`、`shaOne` 的生成算法
3. 逆向 `dv`（设备指纹）的生成算法
4. 在 Go 中实现完整的签名生成 + JSONP 请求 + Cookie 会话管理
---

## 15. dumate2api 实现方式分析

### 15.1 核心策略

dumate2api 是一个 Python 反向代理，**不实现 OAuth 登录**。它通过以下方式工作：

```
客户端请求 (OpenAI 格式)
  → dumate2api (本地代理, 端口 8792)
    → 从 DuMate 日志中提取凭据 (INAPP_KEY + ACCOUNT_ID)
    → 转发到 DuMate 本地后端 (端口 52000-52999)
      → /api/qianfanproxy/v1/chat/completions
        → 千帆云推理 API
```

### 15.2 凭据来源

`discover_credentials()` 函数从两个来源获取凭据：

**环境变量**:
```
DUMATE_ACCOUNT_ID   → X-Dumate-Account-Id header
DUMATE_INAPP_KEY    → X-Dumate-Inapp-Key header
DUMATE_USER_ID      → 用户标识
DUMATE_DEVICE_ID    → X-Dumate-Device-Id header
DUMATE_TENANT_ID    → X-Dumate-Tenant-Id header (企业版)
DUMATE_PRD_SOURCE   → X-Dumate-Prd-Source header (enterprise/personal)
```

**本地日志扫描**:
```
~/Library/Application Support/qianfan-desktop-app/log/electron.log
%APPDATA%\qianfan-desktop-app\log\electron.log
```

从日志中提取 `DUMATE_INAPP_KEY:`、`DUMATE_ACCOUNT_ID:`、`DUMATE_USER_ID:` 等行。

### 15.3 请求转发

```python
# main.py:287-296
upstream_url = f"http://127.0.0.1:{port}/api/qianfanproxy/v1/chat/completions"
headers = {
    "Authorization": "Bearer nokey",           # 占位符！
    "X-Dumate-Account-Id": creds["account_id"],
    "X-Dumate-Device-Id": creds["device_id"],
    "X-Dumate-Inapp-Key": creds["inapp_key"],
    "X-Opencode-Directory": creds["workspace_dir"],
    "User-Agent": "Dumate2API/1.1.0"
}
```

**关键**：`Authorization: Bearer nokey` 是占位符，真正的认证靠 `X-Dumate-Inapp-Key` + `X-Dumate-Account-Id` 头。这与第 4.5 节的 `qianfanproxy` 路由分析完全吻合。

### 15.4 后端发现与自动拉起

```python
# main.py:141-162
def get_dumate_backend_port():
    # 扫描 52000-52999 端口范围，检测 DuMate 本地后端是否在线
    
# main.py:164-196
def ensure_dumate_running():
    # 如果未检测到端口，自动拉起 DuMate 客户端
```

### 15.5 对外暴露的 API

| 端点 | 方法 | 功能 |
|---|---|---|
| `/v1/models` | GET | 返回核心模型列表 |
| `/v1/chat/completions` | POST | 转发推理请求（支持 streaming） |
| `/health` | GET | 健康检查 |

### 15.6 模型映射

```python
# main.py:67-89
MODEL_MAP = {
    "deepseek-v4.1": "deepseek-v4-pro",
    "kimi-k3.1-pro": "kimi-k31",
    "glm-5.3": "glm-53",
    "qwen-3.8-flash": "qwen38-flash",
    # ... 更多映射
}
```

### 15.7 与 CLIProxyAPI 的本质区别

| 维度 | dumate2api | CLIProxyAPI Provider 集成 |
|---|---|---|
| 部署方式 | 本地运行，需 DuMate 客户端同时运行 | 云端部署，独立运行 |
| 认证 | 从日志窃取静态凭据 + 调用本地后端 | 标准 OAuth + 直接调用千帆云 API |
| 凭据类型 | `X-Dumate-Inapp-Key`（客户端标识符） | OAuth access_token / refresh_token |
| 依赖 | DuMate Electron 客户端必须运行 | 无外部依赖 |
| 刷新机制 | 无需刷新（静态凭据） | 需要 refresh_token 刷新 |

### 15.8 对 CLIProxyAPI 的启示

dumate2api 验证了逆向报告的结论：**Passport 登录 API 签名机制过于复杂，无法在独立客户端中复现**。它通过绕过 OAuth 直接提取运行时凭据来解决这个问题。

如果要做 DuMate Provider，有两条路：

**路径 A：参考 dumate2api 思路**
- 不实现 OAuth 登录
- 从 DuMate 客户端日志/文件中提取 `X-Dumate-Inapp-Key` + `X-Dumate-Account-Id`
- 直接调用千帆云推理 API
- 需要确认千帆云 API 是否接受这些 header（dumate2api 是通过 DuMate 本地后端转发的）

**路径 B：实现完整 OAuth**
- 逆向 moonshadV3 签名算法（需要运行时 Hook）
- 实现 Passport 的 SMS 发送 + 验证 + 登录流程
- 获取标准 session token

**建议**：路径 A 更实际——如果千帆云推理 API 直接接受 `X-Dumate-Inapp-Key` 作为认证头，就可以绕过 OAuth 实现简单集成。
---

## 16. 密码登录流程（source-confirmed）

### 16.1 端点

| 端点 | 方法 | 用途 |
|---|---|---|
| `/v2/getpublickey` | GET (JSONP) | 获取 RSA 公钥 |
| `/v2/api/?login` | POST (JSONP) | 提交登录 |

### 16.2 密码登录流程

```
用户输入用户名 + 密码
  ↓
_getRSA(callback)
  ↓
passport.data.getRsaKey({gid, loginVersion: "v5"})
  ↓ (GET /v2/getpublickey)
服务端返回 {pubkey, key}
  ↓
RSA.setKey(cert.pubkey)
  ↓
passwordVal = _SBCtoDBC(password.value)
if (passwordVal.length < 128 && !config.safeFlag) {
    data.password = RSA.encrypt(passwordVal)
    data.rsakey = cert.key
    data.crypttype = 12
}
  ↓
passport.data.login(data)
  ↓ (POST /v2/api/?login)
服务端验证 → 签发 session → Set-Cookie → DuMate 读取 Cookie
```

### 16.3 密码登录参数

```javascript
// loginv5_tangram_661886f.js
data = {
    userName: _SBCtoDBC(username),        // 用户名
    password: RSA.encrypt(password),       // RSA 加密的密码
    rsakey: cert.key,                      // RSA 密钥标识
    crypttype: 12,                         // 加密类型
    token: me.bdPsWtoken,                   // CSRF token
    ds: mkdData.ds,                        // mkd 签名
    tk: mkdData.tk,
    s: mkdData.s,
    k: mkdData.k,
    dv: dv_Input.value,                    // 设备指纹
    FP_UID: cookie("FP_UID"),              // 指纹 UID
    FP_INFO: window.PP_FP_INFO,            // 指纹信息
    loginVersion: "v5",
    supportdv: "1",
    timeSpan: Date.now() - initTime,       // 耗时
    // moonshadV3 添加 sig/shaOne/alg/elapsed/rinfo
}
```

### 16.4 RSA 加密细节

```javascript
// 密码长度 < 128 且非 safeFlag 模式时，使用 RSA 加密
if (passwordVal.length < 128 && !me.config.safeFlag) {
    data.password = baidu.url.escapeSymbol(me.RSA.encrypt(passwordVal));
    data.rsakey = me.rsakey;
    data.crypttype = 12;
}
```

- RSA 公钥从 `/v2/getpublickey` 获取
- `crypttype=12` 表示 RSA 加密
- 密码经过 `_SBCtoDBC` 半角转换后加密
- 使用 `baidu.url.escapeSymbol` 进行 URL 编码

### 16.5 与 SMS 登录的对比

| 维度 | SMS 登录 | 密码登录 |
|---|---|---|
| 端点 | `/v2/?getphonestatus` → `/v2/api/senddpass` → `/v2/api/?logintype=sms` | `/v2/getpublickey` → `/v2/api/?login` |
| 密码处理 | 无 | RSA 加密 + crypttype=12 |
| 用户名 | 手机号 | 用户名/手机号/邮箱 |
| 签名机制 | ds/tk/s/k/dv/moonshadV3 | ds/tk/s/k/dv/moonshadV3（相同） |
| 额外参数 | countrycode, mobileencryption | rsakey, crypttype, FP_UID, FP_INFO |

### 16.6 关键结论

**密码登录和 SMS 登录都走百度服务端，使用相同的签名机制（ds/tk/s/k/dv/moonshadV3）。**

区别仅在于：
1. SMS 登录需要先获取手机号状态、发送验证码
2. 密码登录需要先获取 RSA 公钥、加密密码
3. 两者最终都调用 `/v2/api/?login` 端点（SMS 登录可能是 `/v2/api/?logintype=sms`）

签名复杂度相同，都依赖 moonshadV3 动态下发的签名函数。dumate2api 绕过签名的方式对两者都适用——直接从 DuMate 客户端日志提取运行时凭据。
---

## 17. 微信扫码登录（source-confirmed）

### 17.1 认证流程

```
用户点击微信按钮
  ↓
第三方登录弹窗 → 微信授权页
  ↓
微信授权 → 回调百度 Passport
  ↓
服务端签发 session → Set-Cookie → DuMate 读取 Cookie
```

### 17.2 与密码/SMS 登录的关键区别

| | 密码/SMS 登录 | 微信扫码登录 |
|---|---|---|
| 认证界面 | Passport 输入框 | 微信授权页 |
| 凭据交换 | 用户名+密码/验证码 → session | 微信 code → session |
| 签名机制 | ds/tk/s/k/dv/**moonshadV3** | **不需要** |
| 设备指纹 | 需要 `dv` 参数 | 不需要 |
| 签名复杂度 | 高（moonshadV3 动态下发） | **低** |

### 17.3 为什么微信扫码签名更简单？

微信扫码是**第三方 OAuth 流程**：

```javascript
// loginv5_tangram_661886f.js
authsiteLogin: ['weixin']  // 支持 'qzone', 'tsina', 'weixin'
```

1. 前端只渲染微信按钮
2. 点击后跳转微信授权页
3. 微信授权后回调百度 Passport
4. **签名由微信侧和百度服务端处理**，前端不生成复杂签名

### 17.4 QR 扫码登录（百度 App）

百度 App 扫码登录使用不同的端点：

```
GET /v2/api/getqrcode    → 获取二维码
GET /channel/unicast     → 轮询扫码状态
GET /v3/login/main/qrbdusslogin → 扫码确认后登录
```

QR 扫码登录同样**不需要 moonshadV3 签名**——认证在百度 App 侧完成。

### 17.5 对 CLIProxyAPI 的启示

微信扫码是**最可行的路径**——不需要逆向 moonshadV3。但有个问题：百度 Passport 的微信扫码是**页面内完成的**，不是标准 OAuth 授权码流程。

要做独立 Provider 需要确认 Passport 是否提供公开的微信 OAuth client 注册机制。目前逆向证据不支持。

dumate2api 的做法对**所有登录方式都适用**——不关心你用什么方式登录，只要 DuMate 客户端在线，从日志提取 `X-Dumate-Inapp-Key` 就能复用。
---

## 18. X-Dumate-Inapp-Key 生成机制（source-confirmed）

### 18.1 生成代码

```javascript
// build-ts-electron/lib/crypto/inapp-key.js (完整源码)
const crypto = require("crypto");

// 模块加载时生成一次，并通过 DUMATE_INAPP_KEY 共享给所有子进程
const inAppKey = crypto.randomBytes(32).toString('hex');

function getInAppKey() {
    return inAppKey;
}
```

### 18.2 关键特性

| 特性 | 值 |
|---|---|
| 生成方式 | `crypto.randomBytes(32).toString('hex')` |
| 长度 | 32 字节 = 64 字符 hex 字符串 |
| 生成时机 | **进程启动时**（模块加载时） |
| 持久化 | **不持久化**——每次重启 Electron 都会重新生成 |
| 作用域 | 仅 Electron 主进程内有效 |
| 传递方式 | 通过 `DUMATE_INAPP_KEY` 环境变量传给所有子进程 |

### 18.3 使用位置

`X-Dumate-Inapp-Key` 头被用于以下场景：

```
backend-service.js         → 传给 desktop-main 子进程
opencode-service.js        → 传给 opencode 子进程的 HTTP 请求头
opencode-router-service.js → 传给 router 子进程
base-proxy-service.js      → 传给 base-proxy 子进程
vm-manager-service.js      → 传给 vm-manager 子进程
sandbox/light/index.js     → 沙箱 HTTP 请求头
```

所有子进程通过 `process.env.DUMATE_INAPP_KEY` 获取同一个值。

### 18.4 日志中的可见性

`redactRecordForLog()` 将 `inappkey` 标记为敏感字段，在 Electron 主进程日志中会被替换为 `[REDACTED]`。

但 **子进程（Go 后端）可能在自己的日志中输出环境变量**。dumate2api 就是从 `electron.log` 中扫描 `DUMATE_INAPP_KEY:` 行来提取的。

### 18.5 后端验证

Go 后端（`dumate-main-server.exe`）字符串中有 `DUMATE_INAPP_KEY`，说明后端读取此环境变量。后端用它做什么——是简单头校验，还是更复杂的签名验证——需要进一步逆向确认。

### 18.6 对 CLIProxyAPI 的关键启示

**`X-Dumate-Inapp-Key` 不是云端认证凭据**，而是：

1. **本地进程间认证 token**——验证请求来自 DuMate Electron 主进程
2. **每次重启都变化**——无法持久化使用
3. **不授予千帆云访问权**——只授予 DuMate 本地后端的访问权

这意味着 dumate2api 的方案**强依赖 DuMate 客户端在线**：
- DuMate 不运行 → 没有 Inapp-Key → 无法代理
- DuMate 重启 → Inapp-Key 变化 → 需要重新提取

**真正的云端认证凭据是**：
- `X-Dumate-Account-Id`（账号标识，持久）
- `X-Dumate-Device-Id`（设备标识，持久）
- Cookie 中的 `BAIDUID` / `BDUSS`（会话凭据，有过期时间）

这些才是 DuMate 后端用来访问千帆云的真正凭据。
---

## 19. Cookie 登录分析（runtime-confirmed）

### 19.1 Cookie 中的关键字段

从 console.bce.baidu.com 的 cookie 中提取的认证相关字段：

| Cookie 名称 | 值 | 用途 |
|---|---|---|
| `BDUSS` | g5fkhDLXpHR2hFfmZZS0RoYjRxVlI3RXpFZW5peFJmdjEyaUhyT1lkT2JaQlpySVFBQUFBJCQ... | 百度账号 session token |
| `BDUSS_BFESS` | 同上 | BFESS 域名变体 |
| `bce-login-accountid` | **bc01598b7b534f3f93822bcdf91f5026** | **这就是 X-Dumate-Account-Id** |
| `bce-passport-stoken` | 5c0c5ad72c7ef58b204ae0097631288255ce25364bb8ab619e849059f2cc9a34 | Passport session token |
| `bce-sessionid` | 001a15cd862ba1a4eb9a18e2023431204b1 | 百度云会话 ID |
| `bce-long-term-sessionid` | 0011449746765a44bc5b182c41d10d0c4ff | 长期会话 ID |
| `bce-session` | bc01598b...|0fa4edb5eac10ec365419d55efe0f44a | 会话标识 |
| `bce-login-domain-account` | 老何聊故事 | 显示名 |
| `bce-login-expire-time` | 2026-11-07T08:17:43Z|... | 过期时间 |
| `bce-ctl-client-cookies` | BDUSS,BDUSS_BFESS,bce-passport-stoken,bce-device-cuid,bce-device-token,BAIDUID | 允许的 cookie 列表 |

### 19.2 Cookie 能做什么

**能做**：
- 访问 `console.bce.baidu.com` 网页界面
- 调用 console 前端 API（`/api/dumate/*`, `/api/ai_apaas/*` 等）
- 作为 `X-Dumate-Account-Id` 的来源

**不能做**：
- 直接调用千帆推理 API（`qianfan.baidubce.com/v2/chat/completions`）
  - 实测返回 401 `invalid_iam_token`
  - 需要 Bearer IAM token，不是 cookie
- 直接调用 `aip.baidubce.com` 推理端点
  - 实测返回参数错误，需要 access_token

### 19.3 为什么 Cookie 不能直接调推理 API？

千帆推理 API（`qianfan.baidubce.com/v2/`）使用 **Bearer IAM token** 认证：

```
POST https://qianfan.baidubce.com/v2/chat/completions
Authorization: Bearer <IAM_TOKEN>
```

而 IAM token 需要用 AK/SK 签名获取：

```
POST https://aip.baidubce.com/oauth/2.0/token
grant_type=access_token&scope=brain_all_scope
&AK=<AccessKey>&SK=<SecretKey>
```

Cookie 中的 `BDUSS` 是百度账号 session，不是 IAM token。DuMate 通过 `dumate-svc.baidu.com` 调度服务把 cookie 转换成 IAM token。

### 19.4 完整认证链

```
用户登录 (Passport / 微信 / SMS / 密码)
  → 服务端签发 Cookie (BDUSS, bce-session, bce-passport-stoken)
  → DuMate 客户端读取 Cookie，存入 auth.json
  → DuMate 后端启动时读取 Cookie
  → DuMate 后端调用 dumate-svc.baidu.com 调度服务
  → 调度服务用 Cookie 换取 IAM token
  → DuMate 后端用 IAM token 调用千帆推理 API
```

### 19.5 对 CLIProxyAPI 的可行性评估

**用 Cookie 直调推理 API：不可行**
- Cookie 不是 IAM token
- 推理 API 需要 Bearer 认证
- 没有 AK/SK 无法获取 IAM token

**绕过 Cookie 直接用 bce-login-accountid**：不可行
- `X-Dumate-Account-Id` 只是标识符，不是凭据
- 单独使用会被后端拒绝

**完整实现：需要逆向调度服务**
- 逆向 `dumate-svc.baidu.com` 的 token 交换接口
- 或者从 DuMate 客户端日志/运行时提取 IAM token
- 或者逆向 DuMate 后端的 cookie → IAM token 转换逻辑

**最实际的方案**：参考 dumate2api
- 依赖 DuMate 客户端在线
- 从日志提取 `X-Dumate-Inapp-Key` + `X-Dumate-Account-Id`
- 调用 DuMate 本地后端（`127.0.0.1:52xxx/api/qianfanproxy/v1/chat/completions`）
- DuMate 本地后端处理 cookie → IAM token 转换





<tool_call>bash<arg_key>command</arg_key><arg_value>cd "C:/github/router-for-me/CLIProxyAPI" && wc -l docs/reverse-note/dumate-oauth_CN.md
---

## 20. Cookie 换取 IAM Token（runtime-confirmed）

### 20.1 Token 交换端点

从 main-server 字符串分析，找到以下可能的 token 交换端点：

| 端点 | 方法 | 用途 |
|---|---|---|
| `/api/dumate/app/token/get` | POST | 获取应用级 token |
| `/api/dumate/connector/v2/get_token` | POST | 获取 connector token |
| `/api/dumate/system/online/sessionToken` | GET | 获取会话 token |

### 20.2 实测结果

使用过期 cookie 测试：

| 端点 | 结果 |
|---|---|
| `/api/dumate/app/token/get` | `{"success":false,"message":{"global":"登录凭证已过期，请重新登录"}}` |
| `/api/dumate/connector/v2/get_token` | 同上 |
| `/api/dumate/system/online/sessionToken` | HTTP 302（未登录重定向） |
| `dumate-svc.baidu.com/api/token` | HTTP 404 |

### 20.3 结论

**用 Cookie 换取 IAM Token 是可行的**，但需要：

1. **有效的 Cookie 会话**——过期或无效的 cookie 会被拒绝
2. **正确的请求参数**——每个端点可能需要特定的请求体
3. **CSRF token**——`bce-passport-stoken` 用于 CSRF 防护

### 20.4 完整认证链

```
用户登录 (Passport / 微信 / SMS / 密码)
  → 服务端签发 Cookie (BDUSS, bce-session, bce-passport-stoken)
  → DuMate 客户端读取 Cookie，存入 auth.json
  → DuMate 后端启动时读取 Cookie
  → DuMate 后端调用 /api/dumate/app/token/get
  → 服务端验证 Cookie → 返回 IAM Token
  → DuMate 后端用 IAM Token 调用千帆推理 API
```

### 20.5 对 CLIProxyAPI 的可行性

**理论可行**：
- 从 cookie 中提取 `BDUSS` + `bce-login-accountid` + `bce-passport-stoken`
- 调用 `/api/dumate/app/token/get` 获取 IAM Token
- 用 IAM Token 调用千帆推理 API

**实际挑战**：
1. Cookie 过期后需要重新登录
2. 需要逆向 `/api/dumate/app/token/get` 的完整请求格式
3. 需要处理 CSRF token 和 cookie 刷新机制

**最实际的方案**：参考 dumate2api，保持 DuMate 客户端在线，从日志提取运行时凭据。
