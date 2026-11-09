# LobsterAI (NetEase/Youdao) Provider — 协议契约与集成证据

Provider key: `lobsterai-server`
状态: 协议已恢复 (source-confirmed，静态分析)；实现见本仓库 `internal/auth/lobsteraiserver`、`sdk/auth/lobsterai_server.go`、`internal/runtime/executor/lobsterai_server_executor.go`。**尚未 live 验证**（无真实账号登录）。

## 0. 快照 (Snapshot)

| 项 | 值 |
| --- | --- |
| CLIProxyAPI commit | `f36dcddb82ab80f9593b3b63b2b5eb3fd027672c`（本实现基于此基线，branch `feature/dev`） |
| 客户端安装包 | `/run/media/van/Alaska/a/LobsterAI/LobsterAI.exe` (NSIS, 346,796,264 bytes) |
| 解包产物 | `/tmp/lobster/`：`app-64.7z` → Electron `app/` + `resources/app.asar`（自定义 16 字节前缀 asar） |
| 客户端版本 | `lobsterai` 2026.10.13, Electron, 基于 OpenClaw `v2026.8.1` |
| 检查日期 | 2026-10-15 |
| 归属 | 网易/有道（`lobsterai.project@rd.netease.com`），闭源商业软件，仅本地静态分析 |

来源标注约定: **source-confirmed** = 本地 JS 中明确代码；**runtime-confirmed** = 实机抓包；**inferred** = 由相邻证据推断；**unknown** = 未取得。

## 1. 来源与许可 (Provenance)

- 客户端为闭源商业软件；`package.json` 声明 MIT 但整体为专有发行，asar 带自定义头部混淆。分析仅用于协议互操作，**不复制**其代码。
- 本仓库内同类先例: `floatboat`（桌面客户端 OAuth + 托管网关，Anthropic Messages）、`codearts`（管理 OAuth + 只读额度）。这三者的登录/刷新/额度骨架被完整复用，只是把 wire format 从单一改成按模型选择。
- 复用本仓库自身抽象（translator registry、thinking pipeline、usage reporter、proxy-aware HTTP client），未引入新的并行实现。

## 2. Provider 身份

- 稳定 key: `lobsterai-server`（与 wire format 分离；不伪装成 `claude`/`codex`）。
- 后端: 生产 `https://lobsterai-server.youdao.com`；测试 `https://lobsterai-server.inner.youdao.com`（客户端 `endpoints.js`；可用 `LOBSTER_SERVER_BASE_URL` 覆盖）。**source-confirmed**。
- 同一 origin 既做账号后端（`/api/...`）又做推理网关（`/api/proxy/v1/...`）。

## 3. 授权 (Authorization) — 浏览器授权码 + 本地回环回调

客户端内 `authLocalCallbackServer.js` 起一个 `127.0.0.1:<随机端口>/auth/callback` 服务器，`state = crypto.randomBytes(24).toString('base64url')`，超时 5 分钟，校验 `code` 存在且 `state` 相等。登录页为 `{base}/login`，参数 `source`、`redirect_uri`、`state`。**source-confirmed**。

CLIProxyAPI 侧的对应实现：

- CLI 登录 (`--lobsterai-server-login`)：`internal/auth/lobsteraiserver/oauth_server.go` 起回环服务器（支持临时端口），`sdk/auth/lobsterai_server.go` 打开登录页，并在等待期间提供"粘贴回调 URL"兜底（15s 后提示），与 Codex 同形。
- 管理端登录：`GET /v0/management/lobsterai-server-auth-url`（或 `GET /v8/management/oauth/auth-url?provider=lobsterai-server`）→ `RequestLobsteraiServerToken` 返回 `{status,url,state,flow:"manual"}`，通过共享 `POST /oauth-callback` 写入 `.oauth-lobsterai-server-<state>.oauth` 回调文件，或直接 `POST /lobsterai-server-auth-callback {state,redirect_url}`。

### 登录页解析（重要修正，source-confirmed + runtime-confirmed）

登录页**不在后端 origin 上**：`{base}/login` 是一个需要鉴权的 API，浏览器直接访问会返回 HTTP 200 + JSON envelope `{"code":40100,"message":"登录已过期，请重新登录"}`（本机实测复现），**不能**当作登录页。

客户端真实来源（渲染层 `main-CT-gUJfU.js`）：

```js
// 1) 先向 overmind 取服务端下发的登录地址（可热更，无需发版）
const getLoginOvermindUrl = () => isTestModeEnabled()
  ? "https://api-overmind.youdao.com/openapi/get/luna/hardware/lobsterai/test/login-url"
  : "https://api-overmind.youdao.com/openapi/get/luna/hardware/lobsterai/prod/login-url";
// 2) 失败则回退门户 hash 路由
const getPortalLoginUrl = () => `${getPortalBase()}/login`;   // https://lobsterai.youdao.com/portal#/login
// fetchLoginUrl(): response.data.data.value，否则回退
```

运行时确认：`GET .../prod/login-url` → `{"data":{"value":"https://lobsterai.youdao.com/portal#/login"},"code":0}`；`GET https://lobsterai.youdao.com/portal` → 200，标题 `LobsterAI（有道龙虾）- 网易有道全场景办公助手 Agent`。

实现据此修正：`Client.LoginURL()` 先读该配置端点（5s 超时，失败静默回退），`BuildLoginURL` 把 `state`/`redirect_uri`/`source` 追加到 **hash query**（`#/login?...`，与客户端 `appendLoginParams` 一致；不能用 `url.URL.String()`，它会把 fragment 再转义一次导致双重编码）。

回调最终落在桌面客户端所属的 `127.0.0.1` 回环地址**或** `lobsterai://` 深链（客户端存在只带 `source`、不带 `redirect_uri` 的深链回退分支），因此管理端与 CLI 都按"粘贴回调 URL"处理。

**另两处必需参数（runtime-confirmed 修正）**：

1. `source` 必须是 `electron`。用其它值（曾用 `cli-proxy-api`）时门户**不会**执行授权码跳转，用户直接落到 `#/profile` 个人中心页、拿不到 `code`。
2. `redirect_uri` **不能省略**。省略时门户同样只显示个人中心页、不下发 `code`。省略即回退到 `http://127.0.0.1:18924/auth/callback`（`DefaultLoopbackCallbackPort`）。

管理端现在**优先起真实回环监听**（`RequestLobsteraiServerToken` 调 `NewOAuthServer(18924).Start()`）：本机浏览器点完登录即自动完成换取，无需复制粘帖；监听失败或远端浏览器够不到时，仍保留"粘贴回调 URL"路径（共享 `/oauth-callback` 或 `POST /lobsterai-server-auth-callback`）。

> 若回退到无人监听的 `127.0.0.1:18924`，浏览器会显示"无法连接"，此时 `code` 已出现在**地址栏**——这正是需要复制的内容（前端文案已说明）。

### 3a. 换取令牌 (source-confirmed)

`POST /api/auth/exchange`，body 为 `{authCode, state?, firstKeyfrom, latestKeyfrom, version, uuid?}`：

```js
// 客户端 main.js buildKeyfromPayload / withKeyfromBody
const buildKeyfromPayload = () => {
  const { firstKeyfrom, latestKeyfrom } = getKeyfromAttribution(getStore());
  const uuid = getOrCreateInstallationId();
  const userId = getAuthUserId();
  return { firstKeyfrom, latestKeyfrom, ...(uuid?{uuid}:{}), ...(userId?{userId}:{}), version: app.getVersion() };
};
const withKeyfromBody = (body) => ({ ...body, ...buildKeyfromPayload() });
```

响应统一为 envelope `{"code":0,"message":"ok","data":{...}}`，`data` 含 `accessToken`、`refreshToken`、`user`、`quota`。`keyfrom` 默认 `"official"`（`resources/keyfrom/keyfrom.json`）。**source-confirmed**。`userId` 在登录时通常为空，实现仅发送 `firstKeyfrom/latestKeyfrom/version[/uuid]`。

### 3b. 刷新 (source-confirmed)

`POST /api/auth/refresh` body `{refreshToken, firstKeyfrom, latestKeyfrom, version[, uuid]}`，响应 `{code:0,data:{accessToken,refreshToken?}}`；**未返回 `refreshToken` 时保留旧的**（`authSessionManager.js`）。终止失败：HTTP 401 或错误码 `40100`/`40101`/`NotMember`；其余为瞬态。实现按此语义处理（保留旧 refresh token）。

## 4. 推理 (Inference) — 已 live 实测

**唯一可用 surface 是 OpenAI Chat Completions**（实测，2026-10-15，免费账号）：

| 端面 | 实测结果 |
| --- | --- |
| `POST {base}/api/proxy/v1/chat/completions` | **200**，但**始终**返回 `text/event-stream`（即使 `"stream":false`，无 `stream` 字段也一样） |
| `POST {base}/api/proxy/v1/messages` | HTTP 200 + `event:error`：`不支持的模型: <m> (allowed providers: LobsterAI)` `code=40300` —— **所有模型都不可用** |
| `POST {base}/api/proxy/v1/responses` | **404 Not Found** |

实现据此设计：

- `RequestToFormat` 恒为 `openai`；Anthropic / Responses 前端由**翻译**到 OpenAI 完成，不做直通。
- `Execute`（非流式）分两条路：OpenAI 客户端 → 驱动流式并把 chunk 折叠成 `chat.completion`；其它前端 → 折叠成 OpenAI completion 后再翻译回客户端格式。
- `ExecuteStream` 强制 `stream:true`，并**逐行检测 in-band 错误**：`40300`（不支持模型）/`40301`（模型不可见或无访问权限）以 `event:error` 形式藏在 HTTP 200 里，直接透传会返回空回答。映射为 HTTP 状态（`40300→403`）后交给共享冷却策略。

鉴权与客户端标识头（`buildUpstreamRequestHeaders`，source-confirmed）：

- `Authorization: Bearer <accessToken>`
- `X-LobsterAI-Client-Capabilities: kimi-k3-agentic-v1,thinking-level-control-v1`
- `X-LobsterAI-Client-Version: <appVersion>`

### 4a. 模型目录（实测）

`GET {base}/api/models/available` → `{code:0,message:"success",data:[…]}`。条目字段为 **`modelId` / `modelName` / `apiFormat` / `accessible` / `contextWindow` / `supportsThinking` / `supportsImage` / `runtimeProfile`**（不是 `id`/`name`）。

- 本账号 30 条中 **7 条 `accessible=true`**，其余 `accessible=false` 且 `restrictionHint:"订阅套餐/购买加油包可用"`。
- 请求不可见模型 → `40301`；请求不存在模型 → `40300`。
- 因此**只注册 `accessible=true` 的模型**；`apiFormat` 目前全为 `openai`，仅作记录（`ModelInfo` 不再携带 per-model format 载体）。

实测：目录探测注册 7 个模型，`GET /v1/models` 返回这 7 个；请求未注册模型得到 `400 model_not_found`。

### 4b. 企业账号上下文头（source-confirmed，2026-11-09 补齐）

客户端把企业身份作为**两个请求头**加在**每一个已鉴权请求**上（`AuthSessionManager.fetch`
统一注入，见 `main.ts:6049-6056`），后端据此把账号限定到企业池：

| 头 | 取值 | 来源 |
| --- | --- | --- |
| `X-LobsterAI-Account-Mode` | `enterprise` | `EnterpriseAccountRequestHeader.AccountMode`（`shared/enterpriseAccount/constants.ts:67-70`） |
| `X-LobsterAI-Enterprise-Id` | 十进制字符串 | 同上（`EnterpriseAccountRequestHeader.EnterpriseId`） |

- **个人账号两个头都不发**（`buildEnterpriseAccountRequestHeaders(null) → {}`，`enterpriseAccount/context.ts:267-275`），
  这本身就是判据：本实现用「凭据里有没有 `enterprise_id`」决定发不发，而不是发一个空值。
- 上下文从 `GET {base}/api/enterprise/context` 读取（`Accept: application/json`）。归一化要求
  `enterpriseId` 为**正整数**、`enterpriseName` 非空、`role ∈ {super_admin, member}`
  （`normalizeEnterpriseAccountContext`）；`accountMode: "personal"` 或字段不全 → 视为无上下文。
  嵌套位置在客户端有 6 个候选键（`enterpriseContext`/`accountContext`/`organizationContext`/`context`/`user`/`data`），
  本实现按同一顺序查找。
- **个人账号的返回是「成功 + 明确说自己是个人」，不是错误码**（runtime-confirmed，2026-11-09）：

  ```json
  {"code":0,"message":"success","data":{"accountMode":"personal"}}
  ```

  即走 `accountMode == "personal"` 这条分支（与客户端 `findContextCandidate` 的
  `if (directMode === Personal) return null` 完全一致），而不是走终态码分支。
  本端点**不需要** `keyfrom` 查询串（实测直接用 Bearer 调用即返回上述结果）。
- 业务码 `41600` NotFound / `41602` NotMember / `41612` AccountModeMismatch 是**终态**（清掉已有上下文）；
  其余非 0（如 `41601` Unavailable）只代表本次读失败，**不得**抹掉已知上下文。
- 本实现：登录时（CLI 与管理端各一次）best-effort 读一次并写入凭据
  （`account_mode` / `enterprise_id` / `enterprise_name` / `enterprise_role` / `enterprise_member_id`）；
  之后由 synthesizer 暴露为属性，执行器、管理额度读、签到适配器统一带上这两个头。
- 实测：本账号为**个人账号**，`FetchEnterpriseContext` 返回 `(nil, nil)`，两个头均不发（见 §9）。

### 4c. 活动面的 `platform`（source-confirmed，2026-11-09 补齐）

活动接口的 `platform` **不是渠道名，而是运行环境判别字段**：客户端发的是自己的
`process.platform`（`main.ts:8236`/`8247` → `win32` / `darwin` / `linux`），后端对**其它任何值**
（如早先误用的 `cli`）静默返回空槽位。

- 本实现按 `runtime.GOOS` 映射到这同一套取值（`ActivityPlatformFor`）。
- `runtime.GOOS` 是**代理主机的系统**，不一定等于账号当初使用的环境，且没有任何 Electron 对应物的
  主机（freebsd 等）无法映射。因此：
  - 未知主机**不再**伪装成 `linux` 上报，而是由 `ActivityPlatformFor` 返回空、由适配器发出警告
    （`activity platform is unknown for <goos>; sent linux instead`），
  - 并提供凭据级覆盖 `activity_platform`，让运维显式声明该账号期望的环境。
- 实测：本机 `ActivityPlatformFor("windows") == "win32"`，活动槽位 `slotState=available`（见 §9）。

### 4d. 思考档位 (reasoning_effort) —— 2026-11-09 实测修正

上游契约（`GET /api/models/available`，live）：

```json
"thinkingConfig": { "options": [ {"level":"off","openclawLevel":"off"},
                                 {"level":"high","openclawLevel":"high"},
                                 {"level":"max","openclawLevel":"xhigh"} ],
                    "defaultLevel": "high" }
```

- `level` 是**产品侧展示名**，`openclawLevel` 才是 `reasoning_effort` 的 **wire 值**；两者只在最上档不同（UI 叫 `max`，wire 是 `xhigh`）。
- 19 个 `accessible` 模型中只有 **7 个**声明了 `thinkingConfig`（wire 集合恒为 `off/high/xhigh`，产品集合 `off/high/max`，默认 `high` 或 `max`）；其余 12 个 `supportsThinking:true` 但**无档位表**。
- 网关**不校验**该字段：`off/none/minimal/low/medium/high/xhigh/max` 实测**全部 200**。但**只有 `off` 真正关闭思考**（`off` 实测 reasoning_tokens 缺失，`none` 仍有 20 —— `none` 是空操作）。

**已修缺陷**：目录探测发布的 `ModelInfo` 只填 `Object/Created/OwnedBy/ID`，没有 `Thinking` 元数据。
思考管线按 **`toFormat`** 选 applier（本 provider 恒为 `openai` → OpenAI applier），而
`resolveOpenAIEffort` 在 `modelInfo.Thinking == nil` 时直接 `return "", false` —— **不写字段**，
于是 `reasoning_effort` 被静默丢弃，任何档位都跑在服务端默认。抓取上游真实请求体确认（修复前）：
body 只有 `model,messages,stream`。

**修复规则（目录条目是本频道的唯一权威）**：

| 上游声明 | 发布的 `Thinking` | 效果 |
| --- | --- | --- |
| 有档位表 | `{Levels: <openclawLevel 列表>, ZeroAllowed: 含 off}` | 集合内**原样透传**；集合外**显式 400**（`valid levels: off, high, xhigh`） |
| 仅 `supportsThinking` | `{}`（非 nil、空 Levels） | 客户端要什么发什么：不发明限制，也不发明许可 |
| 无思考 | `nil` | 不转发 |

- **交叉继承坑**：静态目录 `LookupStaticModelInfo(id)` 是**跨 provider 按 id 全局查**的，而
  `kimi-k3` 同时存在于 `kimi`/`codebuddy-ai`/`xiaohuanxiong` 段。修复前 `clone := *info` 会继承
  那个 provider 的档位表（例如 `xiaohuanxiong` 的 `[low, high]`），对本后端施加一份**无关的**校验。
  现在三种情况都显式覆盖。

验证（本地中转抓上游真实 body）：

| 请求 | 修复前 | 修复后 |
| --- | --- | --- |
| `reasoning_effort: high`（已声明模型） | body **无**该字段 | `"high"` |
| `off` | 无 | `"off"`（真正关闭） |
| `xhigh` | 无 | `"xhigh"` |
| `medium` / `max` | 无 | **400** `valid levels: off, high, xhigh` |
| `kimi-k3` + `medium`（未声明档位） | 无 | `"medium"`（原样透传，不设限） |
| `/v1/messages` + `thinking.budget_tokens` | 无 | `"high"`（canonical budget → 档位） |
| `/v1/responses` + `reasoning.effort` | 无 | `"high"` |
| 后缀 `deepseek-v4-flash(high)` | 无 | `"high"`（后缀语法是 `model(value)`，不是 `model-value`） |

**已知非最优（未改共享层）**：canonical「关闭」（`thinking.type=disabled` / 后缀 `(none)`）会被管线映射成
`"none"` —— 上游接受但**不生效**；要真正关闭需显式发 `off`。根因是 `standardLevelOrder` 里没有 `off`，
`ZeroAllowed` 只能产出 `"none"`，这是共享思考层的语义，不是本 provider 能决定的。

## 5. 配额 / 积分 (Quota) — 只读观测，三者分离

| 机制 | 含义 | 本实现 |
| --- | --- | --- |
| 用量记账 | 单次请求 token | 走共享 usage reporter（`helps.ParseClaudeUsage` / `ParseOpenAIUsage` + `StreamUsageBuffer`） |
| 额度/积分观测 | 账号余额与签到活动状态 | `GET /api/user/quota` + `GET /api/user/profile-summary`（两个读，见下）+ 只读 activity slot/context |
| 调度冷却 | 本地路由状态 | 保留上游 status/Retry-After，交给共享 manager（不混用） |

**两个额度读必须分开报，不能互相代替（2026-11-09 实测修正）**：

- `GET /api/user/quota`（**runtime-confirmed**）→ `data`：
  `{"subscriptionStatus":"free","planName":"免费","freeCreditsTotal":300,"freeCreditsRemaining":299.86,"freeCreditsUsed":0.14,"freeCreditsExpiresAt":"2026-10-29",…}`。
  这是**免费档额度桶**，也是客户端 `AuthIpcChannel.GetQuota` 用来做「低额度购买提示」闸门的那个读，**不是账号总额**。
- `GET /api/user/profile-summary`（**runtime-confirmed**，本仓库新增读）→ `data`：
  `{"id":123318,"nickname":"…","totalCreditsRemaining":200,"creditItems":[{"type":"campaign","label":"每日登录奖励","labelEn":"Daily Login Reward","creditsRemaining":100,"expiresAt":"2026-11-14T16:41:26"}, …]}`。
  这是**账号总额 + 分桶明细**，也是客户端登录按钮/账号菜单渲染的那个读
  （`LoginButton.tsx:575` `profileSummary.totalCreditsRemaining`、`:597` `profileSummary.creditItems`）。
- **两者可以差到「一个读 0、另一个读 200」**。2026-11-09 对同一账号实测：
  `quota.freeCreditsRemaining=0`（`freeCreditsExpired=true`）而 `profile-summary.totalCreditsRemaining=200`，
  200 全部来自两笔 `campaign`「每日登录奖励」。只读前者的实现会把一个还有 200 积分的账号显示成 0。
- 因此实现同时报两个读：`credits`（免费桶）与 `total_credits` + `credit_packages`（账号总额与明细），
  各自带 `*_known` 标志；缺失就是 unknown，绝不填 0。
- `GET /api/user/profile`（**runtime-confirmed**）→ `{"yid":"<登录邮箱>","nickname":"130****2189","id":123318}`；`id` 是**数字**、登录标识在 `yid`。
- **签到机制存在**：`GET /api/client-activities/slot?placement=desktop_sidebar&clientVersion=<v>&containerApiVersion=2&platform=<os>` → `{slotState, activity:{activityCode, configRevision}}`；`GET /api/client-activities/{code}/context?configRevision=<n>` → 上下文。类型 `daily_check_in`（动作 `check_in`）/`one_time_credit_reward`（动作 `claim`）。错误码 `51100..51106`，其中 `51104` = 已领取。**source-confirmed**（`activityClient.js`、`shared/activity/constants.js`；前端文案 `每日可领 {credits} 积分`、`登录领取每日积分礼`）。
- **签到默认不执行**：管理端点 `GET /lobsterai-server-quota` 只读复用 slot+context（`FetchActivitySnapshot`），没有任何控件能触发领取。
  需要自动签到时**必须由运维显式启用** `internal/checkin` 的 `lobsterai-server` 适配器
  （默认关闭，与 codearts / codebuddy-cn / dumate / trae 同一套调度框架），见 §10。
- 未知值诚实处理：`credits_known=false` / `total_credits_known=false` / `activity_available=false`，绝不伪造成 0 或无限。

## 6. 错误

- 上游错误体：按 surface 复用本仓库既有分类器 —— Anthropic surface 用 `classifyClaudeUpstreamErrorWithCooling`，OpenAI surface 用 `newOpenAICompatStatusError`，保留真实 status 与 `Retry-After`。**source-confirmed**（复用路径）。

## 7. 实现映射 (本仓库)

| 关注点 | 位置 |
| --- | --- |
| 协议客户端（登录/刷新/模型/额度/活动/企业上下文） | `internal/auth/lobsteraiserver/lobsteraiserver.go`（含 `FetchEnterpriseContext`、`SetEnterpriseContext`、`SetActivityPlatform`、`ActivityPlatformFor`） |
| 回环回调服务器 | `internal/auth/lobsteraiserver/oauth_server.go` |
| SDK authenticator | `sdk/auth/lobsterai_server.go` + `sdk/auth/refresh_registry.go` |
| CLI 登录 | `internal/cmd/lobsterai_server_login.go` + `cmd/server/main.go` (`--lobsterai-server-login`) |
| 嵌入式服务注册 | `sdk/cliproxy/service_auth.go` |
| 管理 OAuth | `internal/api/handlers/management/auth_files_lobsterai_server.go` + 路由 + `oauth_sessions.go` 归一化 + `auth_files_v8.go` |
| 管理配额 | `internal/api/handlers/management/auth_files_lobsterai_server_quota.go`（`GET /v0|/v8/management/lobsterai-server-quota`，含 `total_credits` / `credit_packages` / 签到计数） |
| 自动签到（可选） | `internal/checkin/lobsterai.go`（`NewLobsterAIAdapter`）+ `internal/checkin/capabilities.go` + `sdk/cliproxy/service_checkin.go`（默认关闭） |
| 执行器 | `internal/runtime/executor/lobsterai_server_executor.go` + `lobsterai_server_helpers.go`（`lobsteraiServerApplyEnterpriseHeaders` 在 `PrepareRequest` 与推理路径各调一次） |
| 执行器绑定 | `sdk/cliproxy/service_executors.go` (`case constant.LobsteraiServer`) |
| 模型目录探测 | `sdk/cliproxy/lobsterai_server_models.go` + `service_models.go` + `service_lifecycle.go`（`lobsteraiServerCatalogueModelInfo` 发布思考档位，见 §4d） |
| 静态回退目录 | `internal/registry/model_definitions.go` + `internal/registry/models/models.json`（`"lobsterai-server": []`）+ `catalog_sources.go` |
| 凭证再水合 | `internal/watcher/synthesizer/file.go` |
| 常量 | `internal/constant/constant.go` (`LobsteraiServer`) |
| TUI | `internal/tui/oauth_tab.go` |
| 管理面板前端 | `frontend/`（独立仓库，见 §7a） |

注册清单（完整，逐项已实现）：CLI flag/dispatch、两个默认 auth manager、refresh lead、executor 绑定、模型注册 + 异步目录探测 + 生命周期 goroutine、synthesizer 再水合、v0+v8 路由、v8 dispatch、provider 归一化、`nativeModelProviders`、TUI 条目、`internal/checkin` 适配器（`lobsterai-server`，默认关闭）。

## 7a. 管理面板前端 (frontend/)

`frontend/`（CLI-Proxy-API-Management-Center）是独立仓库，需单独提交改动；本 Provider 已按其 FloatBoat 先例完整镜像：

| 关注点 | 位置 |
| --- | --- |
| 登录入口 | `src/pages/OAuthPage.tsx`（provider card + `CALLBACK_SUPPORTED` + `DEDICATED_CALLBACK_COPY`）；图标 `src/assets/icons/lobsterai.svg` |
| OAuth provider 类型 | `src/services/api/oauth.ts` |
| 额度 API | `src/services/api/lobsteraiServerQuota.ts`（`GET /lobsterai-server-quota`；`credits_known` / `total_credits_known` 缺失即 unknown） |
| 额度适配器 | `src/features/quota/providers/lobsteraiserver/data.ts` |
| 额度渲染体 | `src/features/quota/providers/lobsteraiserver/LobsteraiServerQuotaBody.tsx`（积分总额 + 分桶明细 + 套餐 + 只读签到状态，无「领取」控件） |
| 注册 | `providers/index.ts`、`providers/types.ts`（`QuotaProviderType`/`QuotaStore`）、`constants.ts`（`QUOTA_TAB_ORDER`）、`logic.ts`（`QUOTA_FILTER_MAP`）、`QuotaPage.tsx`、`stores/useQuotaStore.ts`、`utils/quota/validators.ts`、`types/quota.ts`（`LobsteraiServerQuotaState` / `LobsteraiServerCreditPackage`）、`types/authFile.ts` |
| i18n | 7 个 locale 的 `auth_login.lobsterai_server_*` 与 `lobsterai_server_quota.*`（26 键，逐 locale 键集校验一致） |
| 测试 | `tests/lobsteraiServerQuota.test.ts`（payload 解析 / 两个读不可互换 / 分桶归一化 / unknown 语义 / view model） |

卡片只读、不提供领取控件；免费桶仅在**实际持有**时占一行 —— 它常见形态是 0 且已过期（§5 实测），
常驻一行 0 会让一个仍有钱的账号看起来是坏的。

前端验证：`bun run type-check` 通过；`bun run lint` 通过；`bun test` 中本 Provider 相关用例
（`lobsteraiServerQuota` 11 项、`quotaAdapterContract`、`quotaPageLogic`，共 38 项）全部通过。

## 7b. 与 `dsh-our-free-model` 的对照

同为 LobsterAI 渠道的另一实现（`dsh-our-free-model` 的 `channel-pack/src/lobsterai*`）可用作交叉校验；
本仓库与其协议结论一致（端点、`X-LobsterAI-Client-Capabilities`、`stream:true` 强制、`keyfrom` 归属字段）。
两处本仓库此前缺失、已按 §5 补齐：

1. **额度读**：对方一直只用 `profile-summary`（其注释明确写了「`/api/user/quota` 只显示 freeCreditsTotal=300，不含活动积分」），本仓库此前只读 `quota`。
2. **自动签到**：对方有 `auto-checkin.ts`（三步步进 + `idempotencyKey`）；本仓库此前只有只读观测与独立脚本。
   差异是设计选择而非能力缺失：本仓库的签到走 `internal/checkin` 统一调度（默认关闭、可配时区/窗口），
   而不是在每个渠道里自带一个签到循环。

## 8. 未确认 / 待实测

1. **不同账号的模型集合不同**（付费/订阅账号的 `accessible` 集合更大）：本实现按账号实时目录注册，未硬编码。免费账号 2026-10-15 为 7 个、2026-11-09 为 19 个，随账号状态变化。
2. `accessToken` 寿命：由 JWT `exp` 推导（本账号约 30 天，`exp-iat=2592000`），刷新链路已按 `refreshToken` 实现但**未在真实过期后实测**。
3. ~~企业账号上下文头（`getEnterpriseAccountHeaders`）：**unknown**~~ **已解决（2026-11-09）**：
   两个头的名称与取值、上下文端点、归一化规则、终态/瞬态业务码区分均 source-confirmed 并实现（§4b）。
   仍未实测的是**企业账号本身**：本账号是个人账号，只验证了「个人不发送这两个头」这一侧；
   企业侧（真实 `enterprise_id` 下发、企业池计费）无账号可测。
4. `/messages` 的 40300 是否对**所有**账号生效（而非仅免费账号）：仅本账号实测。
5. ~~签到活动接口返回 `slotState:empty`~~ **已解决并实测（2026-10-15）**：`slotState:empty` 是查询串里 `platform=cli` 造成的。后端把 `platform` 当作运行环境判别字段，只接受 `win32`/`darwin`/`linux`；非法值（如 `cli`）静默返回空槽位。修正为按 `runtime.GOOS` 发送真实平台后，管理端点实测 `activity_available:true`、`activity_code:daily-check-in-evergreen-prod-20260814`（见 §9 验证表）。
6. 前缀/别名/排除等共享管线未做账号级实测（走的是既有通用代码路径）。
7. **线上签到未真的执行过**：适配器只做了单测（含「已签不重发」与 `51104` 路径），
   真实账号的领取动作**本次未触发** —— 配额读取已 live 验证，但一次真实 `check_in` 会把该账号当天的
   100 积分领掉，属于不可回退的状态变更，需要明确授权后再做。

## 9. 验证状态 (Verification)

| 项 | 状态 |
| --- | --- |
| 协议契约 | source-confirmed（installer + asar 静态分析） |
| 单元测试 | `go test ./internal/auth/lobsteraiserver/` 通过；`./sdk/auth/`(LobsteraiServer)、`./internal/runtime/executor/`(LobsteraiServer)、`./sdk/cliproxy/`(LobsteraiServer)、`./internal/api/handlers/management/`(LobsteraiServer)、`./internal/registry/`、`./internal/watcher/...`、`./internal/cmd/...` 通过 |
| 编译 | `go build -o <tmp> ./cmd/server` 通过；`gofmt -l` 干净 |
| 目录→format→executor 链路 | **通过**（`TestLobsteraiServerCatalogueProbePublishesFetchedModels` mock 后端 → 探测 → 注册 → `LookupModelInfo` 回读 format → executor 端点/格式选择） |
| 管理登录记录持久化 | **通过**（mock 后端 → exchange → record → store 落盘，含 provider/type/refresh_token） |
| 真实登录/推理/额度 | **已 live 验证**（免费账号，2026-10-15） |
| ├ `/v1/models` | 200，7 个 accessible 模型（目录探测注册） |
| ├ OpenAI Chat 非流式 | 200 `PONG`，`finish_reason=stop`（上游始终 SSE，已折叠） |
| ├ OpenAI Chat 流式 | 200 `text/event-stream` |
| ├ Anthropic Messages | 200 `PONG`，`stop_reason=end_turn`（无 Anthropic 端面，翻译实现） |
| ├ OpenAI Responses | 200 `status=completed`，`output=[reasoning,message]` |
| ├ 工具调用 | 200 `finish_reason=tool_calls`，`{"city": "Paris"}` |
| ├ 计数 token | `POST /v1/messages/count_tokens` → `{"input_tokens":8}` |
| ├ in-band 错误 | 不可见模型/不存在模型 → 400 `model_not_found`；上游 `40300/40301` 已映射为真实状态 |
| └ 管理额度 | `/v0|/v8/management/lobsterai-server-quota` → `credits=299, credits_known=true`, `activity_available=true`（2026-10-15 修正 `platform` 后实测） |
| 额度两个读的差异 | **live 实测（2026-11-09）**：同一账号 `quota.freeCreditsRemaining=0`（`freeCreditsExpired=true`、`freeCreditsTotal=300`）而 `profile-summary.totalCreditsRemaining=200`，200 全部来自两笔 `campaign`「每日登录奖励」。只读 `quota` 会把该账号显示成 0 —— 这就是本次修正的动机 |
| `FetchCreditSummary` 解析 | **live 实测**：`total=200 known=true user="123318"`；`packages=2` 均为 `type=campaign / label=每日登录奖励 / label_en=Daily Login Reward`，`remaining` 各 100，`expires` 分别为 `2026-11-14T16:41:26`、`2026-11-15T00:44:15`；分桶合计 200 与总额一致 |
| 模型目录（2026-11-09 复测） | 19 个 `accessible=true`（含 `kimi-k3`），全部 `apiFormat=openai`，`contextWindow` 256k–1M |
| 签到适配器 | 单测 11 项：已签→`already_done` 且**不重发 POST**、可签→`success` 且断言 `configRevision`/`idempotencyKey`/空 `payload`、`51104`→`already_done`、`completed`→`already_done`、空槽位→`unavailable`、业务错误码→`failed`（消息带码）、缺凭据/异 provider→`blocked`、Identity、capability 登记 |
| 签到**启用路径** | **实测（2026-11-09）**：`PUT /v8/management/config.yaml` 带 `automation.checkin.providers.lobsterai-server.enabled: true` → **200**；修复前是 422 `automation.checkin: unsupported provider "lobsterai-server"`（详见 §10a）。未知 provider 仍 422。`GET .../automation/checkin/status` → `lobsterai-server=true` |
| 企业上下文（live） | **实测（2026-11-09）**：`GET /api/enterprise/context` → `{"code":0,"message":"success","data":{"accountMode":"personal"}}`；`FetchEnterpriseContext` → `(nil, nil)`（个人账号），故两个企业头均不发。走的是「明确个人模式」分支（与客户端一致），不是终态码分支 |
| 企业上下文（单测） | 归一化：个人模式 / 非正整数 id / 缺名 / 未知 role / 半成品载荷 → 无上下文；嵌套键可解析。终态码 `41600`/`41602`/`41612` → `(nil, nil)`；`41601` 与 5xx 与缺业务码 → 报错（不抹掉已知上下文）。`SetEnterpriseContext` 拒绝不完整身份并能清除既有身份 |
| 企业头下发 | 执行器 `PrepareRequest` 单测：个人/仅 mode/仅 id 三种不完整形态都**不发**头；完整形态发 `enterprise` + id。客户端侧单测：额度读、活动读带同一对头 |
| 活动 platform | 单测：`ActivityPlatformFor` 三个 Electron 主机正映射、未知主机（`freebsd`/`plan9`/空）→ 空串；槽位请求默认按主机映射、凭据 `activity_platform` 覆盖生效且清空后回落。**live**：本机映射 `windows→win32`，槽位 `slotState=available` |
| 签到脚本 | `docs/scripts/lobsterai_server_checkin.py` 实测签到成功：`creditsGranted=100`，`periodKey=2026-10-15`；重跑返回「今日已签到」不重复领取；`profile-summary` 新增 `每日登录奖励 +100.0` |
| 思考档位 | **实测（2026-11-09，本地中转抓上游真实 body）**：修复前 `reasoning_effort` 在所有档位下都**不进**上游 body；修复后 `high`/`off`/`xhigh` 原样透传、`medium`/`max` 显式 400、未声明档位的模型原样透传、Claude budget 与 Responses `reasoning.effort` 与 `(high)` 后缀均映射到 `"high"`。明细见 §4d |
| 已知与本改动无关的失败 | `sdk/cliproxy` 的 `TestRegisterModelsForAuthCodeBuddyCNOAuthIgnoresAPIKeyModels` 在干净 HEAD 上即失败（已用 stash 复现），非本改动引入 |

## 10. 自动签到 (两条路径)

签到动作（`POST /api/client-activities/{code}/actions/{actionId}`）默认**不由代理执行**：
管理端点与执行器都只读（`FetchActivitySnapshot` 永不执行动作）。需要自动签到时有两条路径：

### 10a. 内置调度器（推荐）

`internal/checkin` 的 `lobsterai-server` 适配器复用本仓库统一的签到框架（与 codearts /
codebuddy-cn / dumate / trae 同一套调度器），**默认关闭**，由运维在管理面板的
「自动签到」页显式启用并设置时区/窗口：

> **启用路径的一个坑（2026-11-09 修）**：provider id 存在**两份清单** ——
> `checkin.Capabilities()`（面板据此渲染卡片）与 `internal/config` 里的允许列表
> （`AutoCheckinConfig.Validate()` 的 switch）。后者当时只写了
> `codearts/codebuddy-cn/trae`，于是 `dumate` 与 `lobsterai-server` **能被列出、能被勾选，
> 但保存时必然 422 `unsupported provider`**。两处现已由测试钉死（见 §9）。
> 该 switch 是刻意的独立副本：`internal/config` 无法 import `internal/checkin`（会成环），
> 且配置文件里的拼写错误必须响亮失败。

- 先读 slot + context，`claimedToday` / `completed` / 业务码 `51104` 均映射为 `already_done`，
  **不会重复 POST**；
- 槽位未开放或活动未暴露动作 → `unavailable`；上游错误 → `failed`（消息带业务码）；
- 凭据缺失 → `blocked`（`credential_unavailable`），不伪装成成功；
- `platform` 按运行主机发送（`win32`/`darwin`/`linux`），**不可用 `cli`**（见 §8.5）；
- 幂等键 `cli-<毫秒>-<随机 8 字节 hex>`，满足 `^[A-Za-z0-9._:-]{1,64}$`。

### 10b. 独立脚本（无调度器时的替代）

```bash
python3 docs/scripts/lobsterai_server_checkin.py status        --auth-file data/auth_files/lobsterai-server-123318.json  # 只读
python3 docs/scripts/lobsterai_server_checkin.py               --auth-file data/auth_files/lobsterai-server-123318.json  # 签到
python3 docs/scripts/lobsterai_server_checkin.py --auth-dir data/auth_files --prefix lobsterai-server   # 批量
```

- 凭据来源：`--auth-file` / `--auth-dir --prefix` / `--token --backend`（或 `LOBSTERAI_AUTH_FILE` 等环境变量）。
- 幂等：`state.claimedToday=true` 时直接跳过；上游 `51104`（已领取）按成功处理；`idempotencyKey` 满足 `^[A-Za-z0-9._:-]{1,64}$`。
- `--refresh-token` 会先调 `POST /api/auth/refresh` 并把新 token 原样写回 auth 文件。
- `platform` 按 `sys.platform` 发送（`win32`/`darwin`/`linux`），**不可用 `cli`**。
- 离线测试：`python3 -m unittest test_lobsterai_server_checkin`（21 项，全 mock，不触网）。

## 11. 运行时语义说明 (Operational notes)

- **静态目录为空是刻意的**：后端 `/api/models/available` 才是权威来源，任意写死模型清单会为无权限账号暴露不可用模型。首个探测完成后模型才可见。
- **签到只由显式启用的调度器执行**：管理端点与执行器都只读；需要自动签到时启用
  `internal/checkin` 的 `lobsterai-server` 适配器（默认关闭），或使用 §10b 的独立脚本。
- **额度卡展示两个读**：免费桶（`credits`）与账号总额（`total_credits` + `credit_packages`）分列，
  建成一个数字会把「免费档已用完但活动积分仍在」的账号显示成 0（§5）。
- **本地额度重置只清本地状态**：`ResetQuota` 清路由/冷却，不重置上游订阅或积分。
- **provider key 与 wire format 解耦**：`lobsterai-server` 复用 OpenAI/Codex/Claude 三套 translator，不新增公共路由、不伪装成 codex。
