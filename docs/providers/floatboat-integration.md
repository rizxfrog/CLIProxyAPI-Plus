# FloatBoat (aoe.chat Agent OS) Provider — 协议契约与集成证据

Provider key: `floatboat`
状态: 协议已恢复 (source-confirmed)；实现见本仓库 `internal/auth/floatboat`、`internal/runtime/executor/floatboat_executor.go`。

## 0. 快照 (Snapshot)

| 项 | 值 |
| --- | --- |
| CLIProxyAPI commit | `4bf02fb2`（本实现基于此基线） |
| 客户端安装包 | `/tmp/floatboat/Floatboat-Setup-0.5.9-x64.exe` (NSIS, 308,304,824 bytes) |
| 解包产物 | `/tmp/floatboat/app/` (Electron) + `/tmp/floatboat/asar/` (`app.asar` 解包) |
| 客户端版本 | Floatboat 0.5.9, Electron 39.4.0, build `d974bee` (dirty), builtAt 2026-09-28 |
| 检查日期 | 2026-10-07 |
| 归属 | aoe.chat「Floatboat - Agent OS」，闭源商业软件，仅本地静态分析 |

来源标注约定: **source-confirmed** = 本地 JS 中明确代码；**runtime-confirmed** = 实机抓包；**inferred** = 由相邻证据推断；**unknown** = 未取得。

## 1. 来源与许可 (Provenance)

- 客户端为闭源商业软件，无可用开源许可。分析仅用于协议互操作，**不复制**其代码。
- 本仓库内同类先例: `minimax`（managed account + Anthropic Messages 上游）、`xiaohuanxiong`、`codearts`。复用本仓库自身抽象。

## 2. Provider 身份

- 稳定 key: `floatboat`（与 wire format 分离；不伪装成 claude/codex）。
- 品牌: aoe.chat / FloatBoat；DeepSeek 变体为独立发行 `deepseek-agent`（deep link scheme `dsagent`，本 Provider 只实现 `floatboat`）。
- deep link scheme: `aoe`（`index.js` 中 `variants.floatboat.deepLinkScheme="aoe"`）。**source-confirmed**。

## 3. 授权 (Authorization) — Deep-link + 浏览器中转 authorization-code

FloatBoat 使用自有账号体系（`floatboat.ai` 后端），**不是** Anthropic 的 `cai` OAuth（后者只是捆绑的 Claude Code CLI 自带，被自有账号体系覆盖）。推理凭据由后端签发的 `newapi` key 提供。

### 3a. 登录 URL 构造 (source-confirmed)

Renderer `AuthStore`（`index-B7VyprCl.js`）：

```js
const Y3e="aoe.pendingOAuthState", zQt=600*1e3;
// UQt: 取 sign-in 表面 URL，追加 state
function UQt(n,e=fs){ const t=m0({id:"sign-in",releaseVariant:e});
  const i=new URL(t.url); return i.searchParams.set("state",n), i.toString(); }
// oauthLogin
const i=crypto.randomUUID(); VQt({state:i,createdAt:Date.now()});
const s=UQt(i,fs); await Zc.login(s);
```

- sign-in 表面: `{"sign-in":{owner:"product-web",path:"/desktop-sign-in"}}`，origin `https://floatboat.ai`。
- 即登录页 `https://floatboat.ai/desktop-sign-in?state=<uuid>`。
- state 存 localStorage `aoe.pendingOAuthState` `{state,createdAt}`，TTL 600s，回调校验 `x.state===a`。**source-confirmed**。

### 3b. 浏览器打开 + callback_scheme (source-confirmed)

主进程 `AuthServerChannel.login`：

```js
async login(o){ const n=new URL(o);
  const [u]=getDeepLinkProtocolsForRuntime(app,...);
  n.protocol[set](u);                       // "aoe:"
  n.searchParams.set("callback_scheme", appVersion);
  await getAuthManager().login(n.toString()); }
```

### 3c. 回调 deep link (source-confirmed / host+path 为 inferred)

- 解析: `createDeepLinkLogContext` → `hasCode:searchParams.has("code")`、`hasState`、`hasApiKey:searchParams.has("apikey")`、`hasError`。
- Renderer 回调读取: `code`, `state`, `apikey|api_key`（`replace(/ /g,"+")`）, `error`。
- scheme `aoe`，host/path 未在明文常量中出现（由 `isAuthCallbackDeepLink` 判断）— 记为 **inferred**。

### 3d. 令牌交换 / 用户信息 / 新 API key (source-confirmed + runtime-confirmed)

后端 origin = `https://floatboat.ai`（配置键 `api-backend-url`；`billing-url` 同域）。

响应统一为 envelope：`{"code":0,"message":"ok","data":{…}}`（**runtime-confirmed**）。

| 用途 | 请求 | 响应字段 |
| --- | --- | --- |
| 换令牌 | `POST /api/desktop/auth/exchange` body `{code,state,deviceId}` | `data.access_token, data.refresh_token`（**无 `expires_in`**；寿命由 JWT `exp-iat` 推导 = 900s） |
| 用户信息 | `GET /api/desktop/user/me` `Authorization: Bearer <accessToken>` | `user{id,email,name,image,isAdmin}, credits, hasActiveSubscription, membershipLevel, canUseUnlimitedMode, userContext`（envelope 内；实现同时支持从 JWT claims 回退取身份） |
| 取推理 key | `POST /api/desktop/newapi/key` `Authorization: Bearer <accessToken>` | `data.api_key`（**runtime-confirmed**，返回稳定 key） |
| 刷新 | `POST /api/desktop/auth/refresh` body `{refresh_token}` | `data.access_token, data.refresh_token, data.expires_in(=900), data.refresh_expires_in(=2592000)`；`refresh_token` **不轮换**（**runtime-confirmed**）。注意：**路径在 `/api/desktop/auth/` 下，不是 `/api/v1/`**（后者实测 404） |
| 登出 | `POST /api/desktop/auth/sign-out` | `{"code":0,"message":"ok"}`（**runtime-confirmed**） |

落地为 AuthManager 记录: `{userId, accessToken, refreshToken, expiresIn, refreshExpiresIn, apiKey, userEmail, userName, userAvatar, quotaRemaining, billingMembership{hasActiveSubscription,membershipLevel,refreshedAt}}`（本地加密存储）。**source-confirmed**。

### 3e. 设备码 / 账号移交 (source-confirmed，参数形状完整)

| 用途 | 请求 |
| --- | --- |
| 创建配对 | `POST /api/desktop/auth/device-pair` body `{desktop_device_id}` → `{pairing_token, pairing_code, qr_payload, status, expires_at_ms, expires_in}` |
| 查询配对 | `GET /api/desktop/auth/device-pair/status?pairing_token=` → `{status, expires_at_ms, scanned_at_ms, approved_at_ms, rejected_at_ms, consumed_at_ms, request_device{name,platform,model,device_id,ip}}` |
| 批准配对 | `POST /api/desktop/auth/device-pair/approve` body `{pairing_token}` → `{status}` |
| 账号移交认领 | `POST /api/desktop/auth/account-handoff/claim` body `{request_id,nonce}` → `{requestId,status,targetPath,browserLabel,expiresAtMs}` |
| 移交决策 | `POST /api/desktop/auth/account-handoff/{approve\|reject}` body `{request_id,nonce}` → `{status}` |

`createDevicePairSession` 的 path 常量被混淆，但其 URL 拼接与返回值同族；记为 `/api/desktop/auth/device-pair` **inferred**（`/status`、`/approve` 为 source-confirmed）。

## 4. 推理 (Inference)

- 上游基址: **`https://newapi.aoe.chat`**，同时用于 `api-anthropic-base-url`（Anthropic Messages）、`api-google-base-url`（Gemini）、`api-llm-utility-base-url`。**source-confirmed**。
- 端点: `${base}/v1/messages`（Anthropic Messages 语义，`anthropic-version` 由客户端注入）。
- 鉴权: **runtime-confirmed 为 `Authorization: Bearer <newapi api_key>` 单一路径**（`x-api-key` 同样被接受）。JWE 分支（`fbk1.<...>` 5 段紧凑 JWE）在本账号上**未被强制**：真实 `/v1/messages` 与 `/v1/chat/completions` 请求均以 Bearer 成功返回 200，无需签名头。
- `max_tokens` 非必填：`/v1/messages` 在不带 `max_tokens` 时仍返回 200（网关补默认值）。
- 协议族：同一 `api_key` 可同时驱动 `/v1/messages`（Anthropic）、`/v1/chat/completions`（OpenAI）与 Gemini 端点；**即使模型未在目录中声明 `anthropic` 端点类型，`/v1/messages` 仍可用**（runtime-confirmed，例如 `glm-5.2`）。
- 上游会**改写响应中的 `model` 字段**（例如请求 `deepseek-flash` 返回 `gemini-3.8-flash`，请求 `deepseek-v4-pro` 返回 `gemini-3.1-pro-preview`）。这是网关侧的后端映射，不是错误。
- **模型目录为在线获取**（`GET ${base}/api/pricing`，`Bearer api_key`），而非写死。响应含 `data[]`（90 项全量）与 **`auto_groups`（本 key 所属分组）**；每个条目带 `enable_groups`、`supported_endpoint_types`、`model_ratio`。
  - 可用性判定 = `auto_groups` ∩ 条目 `enable_groups`（`enable_groups` 是跨账号并集，单独使用会纳入本 key 无权分组）∩ `supported_endpoint_types` 含对话族（openai/anthropic/gemini）。
  - 该规则在本账号得 46 项；**实测仅 26 项真正可调用**，其余返回 `402 credits_exhausted` / `400 invalid_request` / `404 model_not_found` / `502 upstream_unavailable`，另有 1 项 VIP 专属（`gpt-6-astra`）。故目录是**候选集**，最终可用性由运行期请求决定。
  - 客户端 bundle 中的 7 项默认列表（`claude-sonnet-4-6`、`claude-opus-4-6` 等）已**过时**：前两者在上游目录中不存在。静态表已按实测可调用集合重写为回退值，运行期由在线探测覆盖。
- 本地模型旁路: `runtime.floatboatCore.llm.endpoint` + `.apiKey`（用户自填 OpenAI 兼容 endpoint）— 与本 Provider 无关。

## 5. 配额 (Quota)

**runtime-confirmed**，网关提供 new-api 标准计费读：

| 端点 | 观测响应 |
| --- | --- |
| `GET ${base}/v1/dashboard/billing/subscription` | `{"has_payment_method":true,"soft_limit_usd":0.8,"hard_limit_usd":0.8,"system_hard_limit_usd":0.8,"access_until":0}` |
| `GET ${base}/v1/dashboard/billing/usage` | `{"object":"list","total_usage":0.4}` |

- 两者均以 `api_key` 鉴权。剩余额度 = `hard_limit_usd - total_usage`（本次观测 ≈ **$0.4**）。
- 语义澄清，三者相互独立，本实现不混用：
  1. **实测计费额度**：上述两读，即 `FloatboatQuota`（端点 `GET /v0/management/floatboat-quota?auth_index=…`）。
  2. **登录时快照**：凭据上的 `has_active_subscription` / `quota_remaining`（登录响应带入，**不随用量刷新**），仅原样透传展示。
  3. **调度冷却**：本地路由状态，不是账号余额，不由此推断。
- `quota_remaining`（快照字段）的窗口/单位/重置语义仍为 **unknown**；因此两读任一失败时**不报告余额**（`remaining_known=false`），而非填 0 或视为无限。

## 6. 错误

- 上游为标准 Anthropic Messages 错误体（`{"type":"error","error":{"type":...,"message":...}}`）。本实现复用 `classifyClaudeUpstreamErrorWithCooling` 的状态码/重试语义（401/403/429/5xx）。**inferred**（由复用路径得出，需实测确认 429 的重试提示格式）。

## 7. 实现映射 (本仓库)

| 关注点 | 位置 |
| --- | --- |
| 协议客户端 | `internal/auth/floatboat/floatboat.go` |
| SDK authenticator | `sdk/auth/floatboat.go` + `sdk/auth/refresh_registry.go` |
| CLI 登录 | `internal/cmd/floatboat_login.go` + `cmd/server/main.go` |
| 嵌入式服务注册 | `sdk/cliproxy/service_auth.go` |
| 管理 OAuth | `internal/api/handlers/management/auth_files_floatboat.go` + 路由 + `oauth_sessions.go` 归一化 |
| 执行器 | `internal/runtime/executor/floatboat_executor.go` |
| 执行器绑定 | `sdk/cliproxy/service_executors.go` |
| 模型注册 | `sdk/cliproxy/service_models.go` + `sdk/cliproxy/floatboat_models.go`（在线探测） + `internal/registry/model_definitions.go` + `models.json`（回退） |
| 配额 | `internal/api/handlers/management/auth_files_floatboat_quota.go`（`GET /v0/management/floatboat-quota`） + 客户端 `FetchBilling` |
| 常量 | `internal/constant/constant.go` |

## 8. 未确认 / 待实测 (Blocker 清单)

1. ~~`/api/desktop/auth/exchange` 响应键名大小写~~ —— **已实测**：真实响应为 envelope `{"code":0,"message":"ok","data":{"access_token","refresh_token"}}`（snake_case，无 `expires_in`，寿命由 JWT `exp-iat` 推导）。
2. 回调 deep link 的 host/path —— 已实测为 `aoe://auth/callback?code=…&state=…`；实现同时接受 `code`/`state`/`error`。
3. ~~refresh 端点~~ —— **已实测修正**：正确端点为 `POST {backend_url}/api/desktop/auth/refresh` + `{"refresh_token":…}`，返回 envelope，含 `expires_in:900`/`refresh_expires_in:2592000`，且**不轮换** `refresh_token`。此前依据 bundle 推断的 `/api/v1/auth/refresh` 是**错误路径（404）**，已修复。
4. ~~是否强制 JWE 鉴权~~ —— **已实测排除**：Bearer 足够（见 §4）。
5. `quota_remaining`（快照字段）的窗口/单位/重置语义 —— 仍为 **unknown**；已改用网关计费读作为实测额度（见 §5）。
6. `scope=home` / CLIProxyAPIHome 模式下的模型探测与配额端点 —— 未验证（Home 模式下 `asyncProbeFloatboatCatalogue` 按既有约定直接返回）。
7. 管理面板 floatboat 额度卡片 —— **已实现**（见 §9）。同时修复了一个接入缺陷：面板固定调用 `/v8/management/*`（`MANAGEMENT_API_PREFIX = '/v8/management'`，无 v0 回退），而后端的 provider 额度路由此前**只注册在 v0**，因此所有 provider 额度卡片在该面板上都会 404。已将 `/v8/management/floatboat-quota` 注册到 v8 路由组（其余 provider 仍缺 v8 注册，属既有问题，未扩大改动范围）。

## 9. 验证状态 (Verification)

| 项 | 状态 |
| --- | --- |
| 协议契约 | source-confirmed（installer + asar 静态分析）+ **runtime-confirmed**（真实登录与真实推理） |
| 单元测试 | `go test ./internal/auth/floatboat/...` **14/14 通过**（含目录过滤、envelope、JWT 寿命、billing 计算、refresh 路径回归）；`./internal/registry/...`、`./internal/runtime/executor/`、`./internal/api/handlers/management/`、`./internal/watcher/...` 通过 |
| 注册链路 | **端到端通过**：独立实例（8317 之外）加载真实凭据后 `provider-models` 返回 floatboat **41 项，source 全部为 `runtime`**（在线目录已完全覆盖静态回退） |
| 管理登录入口 | `GET /v0/management/floatboat-auth-url` 返回 `https://floatboat.ai/desktop-sign-in?state=…`；用户已完成**真实登录**并产出 auth 文件 |
| 路由 | `/v1/chat/completions` → `https://newapi.aoe.chat/v1/messages`，`Authorization: Bearer <api_key>`；真实网关 `/v1/messages` 与 `/v1/chat/completions` 均返回 **200** |
| 真实推理（直连网关） | **通过**：`deepseek-flash` 正确回答 `17*23=391`；`/v1/messages` 与 `/v1/chat/completions` 均成功；`deepseek-flash` 为推理模型（单次约 125 reasoning tokens） |
| 真实推理（经代理） | **通过**：经本代理 `POST /v1/chat/completions` → 200（`deepseek-flash`，内容 `391`）；`POST /v1/messages` → 200（`OK`） |
| 流式（经代理） | **通过**：SSE 正常，含 `data: [DONE]` |
| 工具调用（经代理） | **通过**：OpenAI 工具定义 → Anthropic 翻译 → 返回 `tool_calls`（`get_weather{city:"Paris"}`，`finish_reason: tool_calls`） |
| Gemini 前端 | **通过**：`GET /v1beta/models` → 200 |
| 额度端点 | **通过**：`GET /v0/management/floatboat-quota?auth_index=…` → 200，实测数值（`hard_limit 0.8`、`remaining_known true`、`groups ["default"]`） |
| 刷新 + 轮换持久化 | **通过**：注入无效 `api_key` + 过期 `access_token`（保留有效 refresh_token）后，经代理推理仍 **200**；凭据被重铸（`api_key` 恢复有效、`expired` 推进到新 token）并落盘；`refresh_token` 按上游行为**不轮换** |
| 管理面板额度卡片 | **通过（浏览器实测）**：额度页出现 `Floatboat` tab（计数 1），卡片显示 Plan=Free、Remaining balance `USD -13.60 / USD 0.80`（0%）、Spent `USD 14.40`、Groups `default`。前端 `tsc --noEmit` 通过、`bun run build` 通过、`bun test tests/floatboatQuota.test.ts` 8/8，`tests/quotaPageLogic.test.ts` 与 `tests/quotaAdapterContract.test.ts` 全绿 |
| 前端单测 | **通过**：新增 `tests/floatboatQuota.test.ts`（8 个用例：解析、未知额度不伪造、空载荷拒绝、分组清洗、视图模型） |
| 已知与本改动无关的失败 | Go：`TestServiceCatalogStartupAndConfigReload`、`TestRegisterModelsForAuthCodeBuddyCNOAuthIgnoresAPIKeyModels` —— 已在**干净 HEAD worktree** 复现为既有失败。前端：25 个失败中 24 个为既有（feature/dev 既有改动，已在 stash 干净树上复现：CodeArts/Cline/Qoder/CodeBuddy/Xiaohuanxiong 契约、visual-config 解析、cooldown 渲染、v8 scalar 一致性），仅 `buildTabCounts` 由新增 tab 引起并已修正预期 |

## 10. 运行时语义说明 (Operational notes)

- **静态表是回退，不是真相**：`models.json` 的 `floatboat` 段（26 项已按实测可调用集合重写）仅在运行期探测失败或尚未完成时生效；注册成功后由在线目录（本账号 41 项）覆盖。`provider-models` 视图按设计合并静态与运行期条目，其中 `source=runtime` 的才是实际注册结果。
- **共享目录上游仍带旧段**：`router-for-me/models` 的 `models.json` 目前仍含过时的 floatboat 段（`claude-sonnet-4-6` 等）。因为“incoming 段总是获胜”，远端刷新会把该旧段写回本地回退。这不影响实际路由（运行期探测覆盖），如需彻底对齐需更新上游 `models` 仓库。`floatboat` 已加入 `preserveLocalCatalogSections`，在**上游缺失该段**时保留本地回退。
- **`used` 增长**：验证期间的大量模型探测计入网关计费，观测值由 0.4 上升至 14.4（订阅上限 0.8）。这是真实用量，不是实现缺陷。
