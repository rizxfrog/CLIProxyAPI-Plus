# CodeArts (华为云 CodeArts Work) — 逆向分析

来源：`CodeArts-0.1.26-f649f71-x64-prod.exe`（NSIS → `$PLUGINSDIR/app-64.7z`）
分析对象：Electron 主进程 `resources/app.asar` + Rust 内核 `resources/bin/agent-kernel.exe`

## 1. 包结构

```
app-64.7z
├── CodeArts.exe                     Electron 壳 (electron-builder)
└── resources/
    ├── product.json                 ★ 所有静态 endpoint / OAuth clientId 都在这
    ├── agent-config.yaml            内核附加配置 (--additional-config)
    ├── app.asar                     主进程 648KB (out/main/index.js) + 渲染层 (out/renderer)
    ├── app.asar.unpacked/out/mcp-servers/
    └── bin/
        ├── agent-kernel.exe         ★ Rust 内核 = opencode-server 0.1.48，真正发 LLM 请求
        ├── officecli.exe
        └── toolset-win32-x64.zip
```

关键结论：**LLM 调用不在 Electron 层，而在 `agent-kernel.exe`（Rust，opencode fork）里**。
主进程只负责 OAuth / 模型目录 / 福利接口，然后把 provider 配置 PATCH 给内核。

## 2. `product.json` 里的静态配置

```json
{
  "snapEngineUrl": "https://snap-access.cn-north-4.myhuaweicloud.com",
  "benefitApiUrl": "https://opengw.developer.huaweicloud.com",
  "apiHost": "https://ide.cn-north-4.myhuaweicloud.com",
  "auth": {
    "clientId": "codearts",
    "region": "cn-north-4",
    "iamStsHost": "https://sts.cn-north-4.myhuaweicloud.com",
    "portalHost": "https://codearts.huaweicloud.com/portal",
    "snapEngineApiHost": "https://snap-access.cn-north-4.myhuaweicloud.com/snap-manager",
    "authRedirectPath": "/oauth/callback",
    "loginTimeout": 300000,
    "pluginName": "snap_AIIDE",
    "pluginVersion": "5.1.0"
  },
  "kernel": { "envConfig": {
    "env": "hc",
    "inferhubBaseUrl": "https://snap-access.cn-north-4.myhuaweicloud.com/api/v2/,https://snap-access.cn-north-4.myhuaweicloud.cn/api/v2/",
    "knowledgeBaseUrl":  "https://snap-access.cn-north-4.myhuaweicloud.com,https://snap-access.cn-north-4.myhuaweicloud.cn",
    "agentCenterApiUrl": "https://snap-access.cn-north-4.myhuaweicloud.com:443,...",
    "policyCenterDomain":"https://snap-access.cn-north-4.myhuaweicloud.com,..."
  }}
}
```

## 3. OAuth 登录流程

代码位置 `app.asar → out/main/index.js`：`class AuthService`（~L1928+）、`class ApiService`（~L1768+）。

### 3.1 授权

```
GET {portalHost}/authorize
    ?theme=2&locale=zh-cn
    &uri_scheme=codearts
    &client_id=codearts
    &port={localhostPort}
    &code_challenge={PKCE_S256}
    &code_challenge_method=SHA-256
    &ticket_id={32字节随机hex}
    &plugin-name=snap_AIIDE
    &plugin-version=5.1.0
```
- `codeVerifier = hex(randomBytes(64))`（128 字符，非标准 43-128 base64url）
- `codeChallenge = base64url_nopad(sha256(codeVerifier))`
- 浏览器外开（`shell.openExternal`），本地回调用 `express` 监听
  `http://127.0.0.1:{port}/oauth/callback`（`portfinder` 从 10000 起）

### 3.2 拿 code 换凭证（主流程）

```
POST {iamStsHost}/v1/oauth2/tokens
Content-Type: application/x-www-form-urlencoded
DPoP: <ES256 JWT, typ=dpop+jwt, htm=POST, htu=同URL, iat, jti>

client_id=codearts
&code={code}
&code_verifier={codeVerifier}
&grant_type=authorization_code
&redirect_uri=http://127.0.0.1:{port}/oauth/callback
```
`DPoP` 头是 `jose.SignJWT(ES256/P-256)`，protected header 内嵌公钥 JWK。
真实 `sts` 服务必需该头，缺失报：
`{"error_code":"APIGW.0106","error_msg":"...Invalid header parameter: DPoP, required"}`

响应（**不是 OAuth 标准 token 响应**）：
```json
{
  "refresh_token": "...",
  "credentials": {
    "access_key_id":     "AK...",
    "secret_access_key": "SK...",
    "security_token":    "临时安全令牌",
    "expiration":        "2026-09-16T05:17:08Z"
  }
}
```
→ 即 **把 OAuth 换成了华为云临时 AK/SK/STS**，后续所有接口用华为云签名。

### 3.3 刷新

```
POST {iamStsHost}/v1/oauth2/tokens
client_id=codearts&grant_type=refresh_token&refresh_token={rt}&code_verifier={codeVerifier}
DPoP: <同上>
```
- 401/403 / `invalid_grant` / `refresh code has expired` / `kid not found` → 重登
- 刷新定时器 `tokenRefreshInterval = 3600000ms`，系统 resume 时也校验

### 3.4 取用户信息

```
GET {iamStsHost}/v5/caller-identity
X-Security-Token: {security_token}
（+ SDK-HMAC-SHA256 签名，见 §5.1）
```
响应：`principal_urn` / `principal_id` / `account_id` / `station_type`，用于取 `userName`。

### 3.5 降级：ticket 轮询（无 refresh_token）

回调带 `secret` 时（`?secret=..&redirect=..`，302 转发后轮询）：
```
GET {snapEngineApiHost}/v1/login/ticket?ticket_id=..&secret=..
plugin-name: snap_AIIDE
plugin-version: 5.1.0
```
每 2s 轮询，最多 10 次；`user_name` 非空即成功。响应：
```json
{ "user_id": "...", "user_name": "...", "domain_id": "...",
  "credential": { "access":"AK", "secret":"SK", "securitytoken":"...", "expires_at":"..." } }
```

### 3.6 回调备用通道

- 自定义协议 `codearts://...?code=...|error=...`（`handleProtocolCallback`）
- 浏览器若同时回调 `127.0.0.1:{port}/oauth/callback`，以先到为准

## 4. 大模型调用接口（内核实测）

> 方法：用 wine 跑 `agent-kernel.exe --host-init-stdin`，`inferhubBaseUrl` 指向本地 mock，
> PATCH provider 配置后发一条 message，抓取真实请求。

### 4.1 端点

| 用途 | 方法 | URL |
|---|---|---|
| chat | POST | `{inferhubBaseUrl}chat/completions` |
| 心跳 | PUT | `{snapEngineUrl}/snap-manager/v1/chat-session/heartbeat` |

`inferhubBaseUrl` = `https://snap-access.cn-north-4.myhuaweicloud.com/api/v2/`
⇒ **真实 chat 端点：`https://snap-access.cn-north-4.myhuaweicloud.com/api/v2/chat/completions`**（实测 401 存在）

### 4.2 请求头

```
content-type:      application/json
host:              snap-access.cn-north-4.myhuaweicloud.com
x-sdk-date:        20260916T041708Z
x-security-token:  {security_token}
authorization:     SDK-HMAC-SHA256 Access={AK}, SignedHeaders=host;x-sdk-date;x-security-token, Signature={hex}
x-auth-token:      {security_token}
user-session-id:   {内核 session id}
x-snap-traceid / traceid / x-ot-trace-id / x-ot-span-id: msg_xxx
x-ot-session-id / x-ot-parent-session-id: {session}
client-ip:         10.0.3.1
user-msg-id:       msg_xxx
created-time:      2026-09-16T04:17:08Z
model-id / model-name: {model}
accept:            */*
```
可白名单透传的头：`maas_type`、`x-snap-traceid`、`ideName`、`ideVersion`、`pluginName`、`pluginVersion`。

### 4.3 请求体（真实抓包）

标题生成调用（完整、小）：
```json
{
  "enable_thinking": true,
  "max_tokens": 32000,
  "messages": [{"role":"system","content":"..."},{"role":"user","content":"..."}],
  "model": "GLM-5.2",
  "stream": true,
  "stream_options": { "include_usage": true },
  "temperature": 0.5,
  "tool_stream": true,
  "user_prompt": "Generate a title for this conversation: ..."
}
```
主对话调用额外带（35 个工具的 agent 请求）：
```json
{ "tools": [ ... ], "tool_choice": "auto", "temperature": 0.0 }
```
- 流式：SSE，`data: {chunk}` … `data: [DONE]`；`usage` 在带 `stream_options.include_usage` 时下发
- **走完全标准 OpenAI Chat Completions 形态**（`messages`/`tools`/`stream`），额外字段是：
  `enable_thinking`、`tool_stream`、`user_prompt`

### 4.4 认证变体

内核里有两种：
1. **HWS SDK-HMAC-SHA256**（当前产品实际走这条）— `Using HWS SDK-HMAC-SHA256 signing`
2. Bearer / `x-auth-token` — `Using Bearer token auth` / `No auth headers set ... relying on chat.headers hook (e.g. x-auth-token)`
3. 请求体字段 `hws_access_key` / `hws_secret_key` / `hws_security_token`（仅在 kernel config 对应字段没填时才注入，产品默认走 header 签名）

### 4.5 心跳

```
PUT {snapEngineUrl}/snap-manager/v1/chat-session/heartbeat
user-session-id: {session}
x-snap-traceid: msg_xxx
（+ 同上 SDK-HMAC-SHA256 签名头）
body: {"status":"busy"}      # 空闲释放时 {"status":"idle"} 之类
```

## 5. 华为云签名 SDK-HMAC-SHA256（必须复刻）

```
ALGORITHM = "SDK-HMAC-SHA256"
X-Sdk-Date = UTC %Y%m%dT%H%M%SZ
SignedHeaders 默认 = host;x-sdk-date;x-security-token
canonicalUri   = 每段 encodeURIComponent，且末尾补 "/"
canonicalQuery = 按 key=value 排序拼接
bodyHash       = sha256(body)（或 X-Sdk-Content-Sha256 头优先）
canonicalRequest = METHOD\ncanonicalUri\ncanonicalQuery\ncanonicalHeaders\nsignedHeaders\nbodyHash
stringToSign     = "SDK-HMAC-SHA256\n{X-Sdk-Date}\n{sha256(canonicalRequest)}"
signature        = HMAC-SHA256(stringToSign, secretKey)
Authorization    = "SDK-HMAC-SHA256 Access={AK}, SignedHeaders={sh}, Signature={sig}"
```
注意内核实测的 `SignedHeaders` **只签名 host / x-sdk-date / x-security-token**（不含 content-type）。

## 6. 模型列表 & 模型参数

### 6.1 模型目录接口（主进程 → 渲染层）

| 用途 | 方法 | URL | 必需头 |
|---|---|---|---|
| 内置模型 | GET | `{snapEngineUrl}/v1/model/builtin` | `X-Security-Token`, `Agent-Type: PromptCenter`, `X-Language`, 签名 |
| Agent 模型 | GET | `{snapEngineUrl}/v1/agent-center/agents/detail?agent_id={id}` | `Agent-Type: AgentCenter` |
| 主 Agent 查找 | GET | `{snapEngineUrl}/v1/agent-center/agents/useragents?offset=0&limit=100&is_primary_agent=true` | `Agent-Type: AgentCenter` |
| LLM 厂商 | GET | `{snapEngineUrl}/v1/llm/vendors` | `Agent-Type: PromptCenter` |
| 免费福利模型 | GET | `{benefitApiUrl}/api/v1/gateway/config` | `X-Language: zh-cn`（**需先确认 `{snapEngineUrl}/v1/benefit-gateway-config` 的 `enabled`，见 §7.2.1**）|
| 额度账本 | GET | `{snapEngineUrl}/snap-manager/v1/statistics/plugin` | `X-Security-Token`, `Content-Type`（**不加 `Agent-Type`**，见 §7.2.2）|
| 权益开关 | GET | `{snapEngineUrl}/v1/benefit-gateway-config` | `Agent-Type: PromptCenter` |
| 思考强度档 | POST | `{policyCenterDomain}/v1/model/thinking-variants` body `{"model_names":[...]}` | `Agent-Type: PromptCenter` |
| 福利开关 | GET | `{policyCenterDomain}/v1/benefit-gateway-config` | |

`/v1/model/builtin` 响应：`{"builtinModels":[ {...}, ... ]}`
`/v1/agent-center/agents/detail` 响应：`{"gpts":{"models":[ {"model_parameters":{...}, "model_name", "model_type", "credit", "usage_token_num", "model_quota_num", "think_level"} ]}}`

### 6.2 模型参数字段（逆向自渲染层 `flattenBuiltinModels`）

| 字段 | 含义 | 映射到内核 |
|---|---|---|
| `model_id` / `id` | 模型 id | `modelId` |
| `model_name` / `name` / `display_name` | 展示名（也当 id 兜底） | `label` |
| `context_window` | 上下文窗口 | `limit.context` |
| `input_context_window` | 输入窗口 | `limit.input` / `limit.context` |
| `output_context_window` | 输出窗口 | `limit.output` |
| `input_length` | 输入长度 | `limit.input` |
| `max_tokens` | 最大输出 | `limit.output`（`output_context_window` 缺失时） |
| `truncate_length` | 截断长度 | （UI 用） |
| `supports_images` | 是否多模态 → `model_category=="图片理解"` | `multimodalModels` |
| `think_level` | `low`/`medium`/`high`/`xhigh`/`max` | 思考强度 |
| `credit` / `credit.ratio_display` | 消耗系数 | `creditRatio` |
| `usage_token_num` | 已用 token | `creditUsed` |
| `model_quota_num` | 配额 | `creditTotal` |
| `is_luxury` | 尊享模型 | 需 Ultimate Pro |
| `enable_queue` | 排队 | `isQueue` |
| `display_enabled` | 是否展示 | |
| `is_free_benefit` / `used_amount` / `total_quota` | 限时免费福利模型 | |
| `model_icon` / `model_desc` / `provider_type` | UI | |

默认值：`DEFAULT_INPUT_CONTEXT_TOKEN_LIMIT = 184000`，`DEFAULT_OUTPUT_CONTEXT_TOKEN_LIMIT = 16000`

### 6.3 真实模型目录（**带有效凭证实测**，权威）

用真实登录态请求 `GET /v1/agent-center/agents/useragents?is_primary_agent=true`
（`Agent-Type: AgentCenter`）拿到 6 个 primary agent，其中 `CodeAgent`
的 `detail` 携带 `model_parameters`。**这是内核对外的真实模型列表**，
渲染层硬编码的 `KERNEL_MODELS` 只是它的本地兜底：

| wire id（上行 `model`） | 显示名 | ctx | input_length | 说明 |
|---|---|---|---|---|
| `GLM-5.2` | GLM-5.2 | 202752 | 196608 | 深度推理，CodeAgent 默认 |
| `glm-5.2-sft-harmony` | **GLM-5.2-ArkTS-SPARK** | 202752 | 196608 | 增训鸿蒙/ArkTS |
| `openpangu-2.0-pro` | OpenPangu-2.0-Pro | 524288 | 512000 | 盘古旗舰 |
| `openpangu-2.0-flash` | OpenPangu-2.0-Flash | 524288 | 512000 | 盘古均衡 |

限时福利（`GET {benefitApiUrl}/api/v1/gateway/config`，`error_code=="0000"`）：

| wire id | ctx | max_tokens（上游宣称） |
|---|---|---|
| `deepseek-v4-flash-0731` | 1048576 | 393216 |
| `deepseek-v4-pro-0813` | 1048576 | 393216 |
| `glm-5.3-flash` | 1048576 | 131072 |

`GET /v1/model/builtin`（`Agent-Type: PromptCenter`，`X-Language: zh-cn|en-us`）
另外返回 `Qwen3-VL-235B`（`model_category == "图片理解"` → 多模态）。

> ⚠️ **id 与显示名不同名**：上行必须用 `glm-5.2-sft-harmony`，
> 只接受 `GLM-5.2-ArkTS-SPARK` 会 404。

**已不存在 / 不下发**（实测 404 `InferHub.002002009.404`）：
`GLM-5.2-ZX`、`Auto`（渲染层 `KERNEL_MODELS` 里仍有，但上游已无路由）；
`Qwen3-VL-235B` 是内嵌别名，上游实际落到 `kimi-k2.6` 路由。
另：`GLM-5.1` 不在 Agent 列表里，但网关仍可路由（实测 200）。

### 6.3.1 两个必须复刻的网关行为（实测）

**① `max_tokens` 硬上限 65536。**
网关做服务端范围校验，超过即 `HTTP 400 RANGE_VALIDATOR`：

```
{"error_code":"InferHub.001001005.400",
 "details":[{"error_code":"RANGE_VALIDATOR",
 "error_msg":"max_tokens -> max_tokens should be between 1 and 65536"}]}
```

这与目录里宣称的 131072 / 393216 **无关**：`GLM-5.2` 传 131072 一样 400。
所以实现里必须把 `max_tokens` **钳到 65536**（而不是照抄目录值），
否则客户端请求会直接失败。

**② 限时福利模型必须带 `maas_type: benefit`。**
`deepseek-v4-*` / `glm-5.3-flash` 不带该头时上游返回
`InferHub.002002009.404 The model is not registered`；
带上即正常。对普通模型带该头无害（路由相同）。
渲染层 `AiChannel` 里对应的逻辑就是
`...isFreeBenefit ? { "maas_type": "benefit" } : {}`。

### 6.4 怎么把模型交给内核（关键：max_tokens 的真正来源）

主进程算好目录后，`AgentKernelAdapter.syncProviderConfig()` 调：

```
PATCH /cag/global/config            (内核本地端口, 127.0.0.1:{port})
x-opencode-directory: <base64/编码后的工作目录>

{
  "provider": {
    "inferhub-provider": {
      "name": "inferhub-provider",
      "npm": "@ai-sdk/inferhub-provider",     // 内置 provider
      "options": { "baseURL": "{inferhubBaseUrl}" },
      "turnCompletionPolicy": "no_tool_calls",
      "models": {
        "GLM-5.2": { "id":"GLM-5.2", "reasoning":true,
                     "limit": {"context":202752,"input":196608,"output":131072},
                     "custom_config": {} }
      }
    }
  },
  "model": "inferhub-provider/GLM-5.2"
}
```

**实测验证**：`limit.output` → 请求体 `max_tokens` 一一对应（改 8192→8192，改 1024→1024）。
`limit.context/input` 用于内核侧上下文裁剪，不是一个单独的上行字段。

### 6.5 `contentWindow` / `maxToken` 结论

- 没有名为 `contentWindow` 的字段。语义对应 **`context_window`（= 内核 `limit.context`）**。
  后端把 context / input / output 作为**三个独立字段**下发（如 GLM-5.1：
  context=202752、input=196608、output=131072，三者不满足 context = input + output），
  客户端只做 `a || b` 兜底，不自行求和。
- `maxToken` 对应 **`max_tokens`（= 内核 `limit.output`）**，上行请求体字段就是 `max_tokens`。
- 若做自定义模型（`apiFormat: openai|anthropic`）：`inputContextWindow` / `outputContextWindow`
  映射到 `limit.input` / `limit.output`，默认 184000 / 16000。

## 7. 每日签到接口

前端（渲染层）流程：`getWelfareDelivery` → 找 `type == daily_claim && claimable && !claimed`
→ 弹窗 → `claimWelfare(campaignId, idempotentKey)` → `confirmWelfare(userBenefitId)`。

### 7.1 主签到（活动/福利体系，`policyCenterDomain` = snap-access）

```
GET  {policyCenterDomain}/v1/ops/delivery?channel=DESKTOP       # 活动列表/签到状态
POST {policyCenterDomain}/v1/ops/claim                          # 执行签到
     body: {"campaignId":"...","idempotentKey":"claim_{campaignId}_{Date.now()}","channel":"DESKTOP"}
POST {policyCenterDomain}/v1/ops/confirm                        # 确认领取
     body: {"campaignId":"...","userBenefitId":"..."}
POST {policyCenterDomain}/v1/ops/invite/generate                # 邀请码
     body: {"campaignId":"..."}
```
请求头（+ SDK-HMAC-SHA256 签名）：
`X-Security-Token`、`Agent-Type: PromptCenter`、`Content-Type: application/json`、`X-Language: zh-cn`

**实测订正（2026-09-17，真实账号完整跑通 claim→confirm 并从官方客户端源码核实）**：
- 每日签到活动的 `type` 实测是 **`USER_LOGIN`**（不是文档原本推测的 `daily_claim`），
  其 `extra.triggerEvent == "user.login"`、`campaignId: 1`（数字！）、claimable=true、status=ELIGIBLE。
  判定每日签到的可靠信号是 `triggerEvent == "user.login"`（或标题含「每日签到」）。
  `campaignId` 在 `/v1/ops/delivery` 响应里是 **JSON 数字**，按字符串解析时会导致整个
  响应反序列化失败（Go `UnmarshalTypeError`）。
- `confirm` 的 body **只有 `{"userBenefitId"}`**，与官方客户端一致
  （`out/main/index.js` `confirmWelfare`：`JSON.stringify({ userBenefitId })`）。
  之前「必须带 campaignId」的推断是错的——那是对伪 id 探测时得到的误导性报错。
- **`confirm` 是尽力而为（best-effort）**：积分在 `claim` 成功时即已入账。
  实测 2026-09-17：confirm 返回 HTTP 400 `HDN.1000 「userBenefitId : unknown exception」`，
  但官方客户端 UI 显示「今日已领」、积分已到账；且官方渲染层
  `await confirmWelfare(...)` **完全不检查其返回值**。
  （`HDN.1000` 的真实原因未知，可能与确认时机/幂等窗口有关；不影响签到结果。）
  因此：confirm 失败不得判定为签到失败，也不得触发简单签到兑底（避免重复入账）。
- 当日已领后再 `claim` 返回 `{"code":40001,"message":"尚未到达权益刷新时间，暂时无法领取"}`。
- `confirm` 的错误返回用 `error_code`/`error_msg`（如 `HDN.xxx`），
  与 `claim`/`delivery` 的 `code`/`message` 字段不同，判定要两者兼容。

响应判定：`code === 0 || code === 200 || data !== undefined`。
`/v1/ops/delivery` 的 `data.items[]` 字段：
`campaignId|campaign_id`、`type`（`DAILY_CLAIM`/`INVITE_USER`/`STUDENT_CERTIFIED`/`USER_LOGIN`）、
`title`、`description`、`benefitAmount`、`claimable`、`status`（`ELIGIBLE|CLAIMED|CONFIRMED|CONSUMED`）、
`displayConfig`(JSON 字符串，含 `jumpUrl`/`claimButtonText`)、`hooks[]`
`/v1/ops/claim` 返回 `data.id` 或 `data.userBenefitId`。
每日签到典型值：1000 积分；邀请/登录活动 100 积分。

### 7.2 简单签到（`benefitApiUrl` = opengw.developer.huaweicloud.com）

```
GET  {benefitApiUrl}/api/v1/benefit/claim      # 查询签到状态
POST {benefitApiUrl}/api/v1/benefit/claim      # 执行签到（body {}）
GET  {benefitApiUrl}/api/v1/user/tokens/balance# token 余额
```
判定：`error_code === "0000"`，数据在 `result`。

> 两个体系都存在，前端优先用 §7.1，失败再退回 §7.2（`window.api.benefit.claim`）。

### 7.2.1 ⚠️ opengw 整族已下线（2026-11-10 实测订正）

`opengw.developer.huaweicloud.com` 的三个端点（`benefit/claim`、`gateway/config`、
`user/tokens/balance`）现在对普通账号**一律返回 HTTP 200 + 业务码 `4000`**：

```json
{"error_code":"4000","error_msg":"invalid channel, group not found","result":null}
```

实测排除了下列解释（全部无效，返回不变）：

| 假设 | 实测 |
|---|---|
| 缺 `Agent-Type` | 带 `Agent-Type: PromptCenter` / `CodeArts` / `snap` 均 4000 |
| 缺 `X-Language` | 带/不带、`zh-cn`/`en-us` 均 4000 |
| 缺 channel 参数 | `?channel=codearts|DESKTOP|IDE|WEB|CLI` 均 4000 |
| 缺 channel 头 | `X-Channel`/`Channel`/`X-Group`/`X-App-Group`/`X-Benefit-Channel`/`X-Client-Channel`/`X-Channel-Code`/`X-Api-Channel`/`Channel-Id`/`X-App-Id`/`appId` 均 4000 |
| 签名错误 | 同一三元组对 snap 主机签名正常（`/v1/model/builtin` 200），且不带 `X-Security-Token` 时返回 `APIG.0301`（证明签名链路本身有效） |
| STS 轮换出的新 AK | 存留三元组与轮换后三元组结果相同 |
| 区域/域名 | `opengw.../huaweicloud.cn` 404；`ide.cn-north-4` 404；`huaweicloud.eu` 域名不存在 |

真正的开关在 **snap 主机**上：

```
GET {snapEngineUrl}/v1/benefit-gateway-config   (Agent-Type: PromptCenter)
→ {"enabled":false,"is_examination_domain":false,"examination_model_enable":false}
```

桌面版 0.2.8 的 `ModelService.getFreeBenefitModels()` 正是先读这个开关，
**两个通道（`benefit` / `education`）都关时就完全不去取 channel 模型**：

```js
const gatewayChecks = [
  { channel: "benefit",   result: await this.benefitService.getBenefitGatewayConfig() },
  { channel: "education", result: await this.benefitService.getEducationGatewayConfig() }
];
if (channels.length === 0) return [];   // 两个都关 → 直接跳过
```

因此：**`4000` 不是客户端缺陷，是该账号的权益计划被关闭**。同一凭证的推理、
模型目录、签到（`/v1/ops/delivery`）全部正常。

### 7.2.2 额度改读 `snap-manager/v1/statistics/plugin`（桌面版 0.2.8 实测）

0.2.8 的 `PackageInfoService` 已经**不再调用** `opengw` 的 balance 端点，
改为读 snap 主机的插件统计账本：

```
GET {snapEngineUrl}/snap-manager/v1/statistics/plugin
```

签名头**只需** `X-Security-Token` + `Content-Type`（**不加** `Agent-Type`，加了即 400）：

```js
const signedHeaders = sign({
  method: "GET",
  url: url2,
  headers: { "X-Security-Token": session.securityToken, "Content-Type": "application/json" }
}, session.accessKey, session.secretKey);
```

该端点**不接受任何 query 参数**（`?channel=` / `?start_date=` / `?plugin_name=` / `?version=` 全部 400）。

响应（真实账号，2026-11-10）：

```json
{
  "metrics": [
    {"name":"usageTokenChatMessages","usage_token_num":0,"package_token_amount":1},
    {"name":"usageTotalPackageCredit",   "package_credit_amount":0,"package_credit_used":0,
     "package_credit_remain":0,"package_credit_expiring_amount":0},
    {"name":"usageBasicPackageCredit",   "...":"同上"},
    {"name":"usageOnDemandPackageCredit","...":"同上"},
    {"name":"usageBonusPackageCredit",   "...":"同上"}
  ],
  "package": {"spec_code":"codearts.agent.individual.trial","status":"normal",
              "package_name_en":"Trial","package_name_cn":"体验版",
              "is_token_package":false,"is_credit_package":true},
  "start_date":"2026-10-16","end_date":"2026-11-17"
}
```

渲染层取数（`app.asar` 渲染层）：

| 展示项 | 指标名 | 字段 |
|---|---|---|
| 总积分 / 已用 / 剩余 | `usageTotalPackageCredit` | `package_credit_amount` / `package_credit_used` / `package_credit_remain` |
| 将过期 | 同上 | `package_credit_expiring_amount` |
| 基础 / 按需 / 赠送桶 | `usageBasicPackageCredit` / `usageOnDemandPackageCredit` / `usageBonusPackageCredit` | 同上 |
| 对话 token | `usageTokenChatMessages` | `usage_token_num` / `package_token_amount` |

> **零值是真实结果**：该账号是「体验版」(`codearts.agent.individual.trial`)，
> `is_credit_package: true` 但四个桶全为 0。这不是读取失败——因此实现里用
> `credits_known` 把「读到了、就是 0」和「读不到」区分开。

`/snap-manager/v1/package/overview` 在 0.2.8 里仍有引用，但对真实账号 **404**
（`APIG.0101 not published`），只在客户端读取 on-demand 指标时兜底尝试。


### 7.3 用户提示文案（确认语义）

- `每日签到领 {n} 积分`
- 已签：`今日已签到，明天再来吧`
- `welfare-benefit` 文案：`升级即送 2000 积分，每日签到领 1000 积分`

## 8. 实测验证过的端点（未鉴权）

| 端点 | 状态 |
|---|---|
| `POST sts.../v1/oauth2/tokens` | 400（需 DPoP）✅存在 |
| `GET sts.../v5/caller-identity` | 401 ✅ |
| `GET codearts.huaweicloud.com/portal/authorize` | 200 ✅ |
| `GET snap.../v1/model/builtin` | 401 ✅ |
| `GET snap.../v1/llm/vendors` | 401 ✅ |
| `POST snap.../api/v2/chat/completions` | 401 ✅ |
| `POST snap.../v1/model/thinking-variants` | 401 ✅ |
| `PUT snap.../snap-manager/v1/chat-session/heartbeat` | 401 ✅ |
| `GET snap.../snap-manager/v1/login/ticket` | 400 ✅ |
| `GET/POST snap.../v1/ops/delivery` `/claim` | 401 ✅ |
| `GET opengw.../api/v1/benefit/claim` | 401 ✅（**带凭证时 200 + `4000`，见 §7.2.1**）|
| `GET opengw.../api/v1/gateway/config` | 401 ✅（同上）|
| `GET opengw.../api/v1/user/tokens/balance` | 401 ✅（同上，**已被 §7.2.2 取代**）|
| `GET snap.../snap-manager/v1/statistics/plugin` | 200 ✅ 额度账本（§7.2.2）|
| `GET snap.../v1/benefit-gateway-config` | 200 ✅ 权益开关（§7.2.1）|

## 9. 实现状态（已在 CLIProxyAPI 落地）

已按本分析实现，实测验证结论：

| 环节 | 验证方式 | 结果 |
|---|---|---|
| OAuth `DPoP` 头 | 对真实 `sts.cn-north-4.myhuaweicloud.com` 发起 exchange | 错误从 `APIGW.0106 Invalid header parameter: DPoP, required` 前进为 `STS5.1805 invalid authorization code`，证明 DPoP/ES256/JWK 形态正确 |
| `SDK-HMAC-SHA256` 签名 | 用真实 `agent-kernel.exe` 抓到的请求体 + 头部做 golden vector | Go 实现复算出**完全一致**的签名 `41cc3658…` |
| 签名后真实调用 | 起真实服务，`/v1/chat/completions` → 真实华为云网关 | `APIG.0301 decrypt token fail`（伪造 STS 的预期结果），说明**签名已通过校验**、请求已进 IAM |
| `authorize` URL | 用生成的 URL 请求真实 portal | HTTP 200，返回正常授权页 |
| 模型列表 | 真实服务 `/v1/models` | 9 个模型（§6.3 表）及 `ca/` 前缀别名 |
| 全模型实调 | `/v1/chat/completions` 逐个真实调用 | 9/9 全部 200（含 `maas_type` 路由的福利模型）|
| `max_tokens` 钳制 | 传 `max_tokens=200000` | 实际按 65536 发上游，返回 200（未钳制时为 400）|
| management | `GET/PUT/PATCH/DELETE /v0/management/codearts-api-key`、`GET /v0/management/codearts-auth-url` | 全部可用 |
| 额度（`GET /v8/management/codearts-quota`） | 真实账号 2026-11-10 | 修复前 **502** `codearts: HTTP 200 code 4000: invalid channel, group not found`；修复后 **200**：`{"credits_known":true,"package":"Trial","spec_code":"codearts.agent.individual.trial","total_quota":0,...}`（该账号体验版积分确为 0，见 §7.2.2）|

代码位置：
- `internal/auth/codearts/sign.go` — SDK-HMAC-SHA256
- `internal/auth/codearts/dpop.go` — ES256 DPoP proof
- `internal/auth/codearts/codearts.go` — OAuth exchange / refresh / caller-identity
- `internal/auth/codearts/models.go` / `checkin.go` / `token.go` — 模型目录 / 签到 / 凭证落盘
- `internal/runtime/executor/codearts_executor.go` — 执行器
- `internal/runtime/executor/helps/codearts_http.go` — 签名 RoundTripper
- `internal/registry/model_definitions.go`（`codearts` 分节位于 `models/models.json`）— 内置模型
- `internal/api/handlers/management/auth_files_codearts.go` + `config_lists_codearts.go` — 登录与管理 API

## 10. 对 CLIProxyAPI 接入的建议

1. **凭证模型**：auth 文件保存 `accessKey / secretKey / securityToken / refreshToken /
   expiresAt / userName / domainId`（= 华为云临时 AK/SK/STS + OAuth refresh_token）。
2. **鉴权**：所有业务接口用 §5 SDK-HMAC-SHA256（`golang` 需自实现，逻辑很短；注意
   canonicalUri 末尾补 `/`、query 排序、`SignedHeaders` 与内核实测一致）。
3. **刷新**：`POST {sts}/v1/oauth2/tokens` `grant_type=refresh_token` + DPoP 头。
   DPoP 是 ES256 (`dpop+jwt`) 且 header 内嵌 JWK，需要用与首次授权**同一把** P-256 私钥，
   因此 auth 文件里要持久化 DPoP 私钥（JWK），否则 refresh 会失败。
4. **执行器**：OpenAI 兼容，POST `{inferhubBaseUrl}chat/completions`，
   额外字段 `enable_thinking`(=reasoning)、`tool_stream`(固定 true)、`user_prompt`；
   可忽略 `agent-kernel` 整套，直接发标准 OpenAI 请求。
   `max_tokens` **必须钳到 65536**（目录里的 131072/393216 会被网关 400 拒绝）。
   限时福利模型需额外加 `maas_type: benefit` 头（见 §6.3.1）。
5. **模型列表**：`/v1/agent-center/agents/detail`（`Agent-Type: AgentCenter`）为主，
   `/v1/model/builtin` 补多模态，`/api/v1/gateway/config` 补福利模型
   （**该端点需先确认权益开关，见 §7.2.1**）；
   静态兜底见 `internal/registry/models/models.json` 的 `codearts` 分节。
6. **额度**：读 `{snapEngineUrl}/snap-manager/v1/statistics/plugin`（§7.2.2），
   **不要**读 opengw 的 `user/tokens/balance`（整族已下线）。用 `credits_known`
   把「读到 0」与「读不到」区分开。
7. **签到**：`POST {policyCenterDomain}/v1/ops/claim`（+ `/confirm`），
   兜底 `POST {benefitApiUrl}/api/v1/benefit/claim`（**opengw 兜底在当前账号下不可用**，
   实测 `4000`；主路径 `/v1/ops/*` 正常）。
