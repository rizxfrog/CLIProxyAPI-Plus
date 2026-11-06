# Loomy (讯飞 Loomy) Provider — 协议契约与集成证据

Provider key: `loomy`
状态: 协议已恢复（source-confirmed，来源为本地提供的参考实现与客户端常量）；实现见
`internal/auth/loomy`、`sdk/auth/loomy.go`、`internal/runtime/executor/loomy_executor.go`。
**尚未 live 验证**（本次未使用真实账号登录，也未抓取上游流量）。

## 0. 快照 (Snapshot)

| 项 | 值 |
| --- | --- |
| CLIProxyAPI commit | `db5587a6`（基线；工作区另有未跟踪的 `frontend/`，与本改动无关） |
| 参考实现 | `/home/van/github/deepseek-ai/dsh-our-free-model/vendor/channel-pack/src/loomy*.ts` |
| 上游出处 | `dsh-codearts-auth`（gitee `iJetLi/deepseek-harness-codearts`），提交 `345f0a07`，MIT |
| 检查日期 | 2026-10-20 |
| 参考文件摘要 (sha256) | `loomy.ts` `4b27d3f9…3269`；`loomy-auth.ts` `4397dc7e…fd44`；`loomy-oauth.ts` `f8acff0f…36d2`；`loomy-sign.ts` `b72480b0…389e`；`loomy-adapter.ts` `92a5c025…fdb`；`loomy-product.ts` `6e89ad62…bddf`；`loomy-credits.ts` `fd820350…9756`；`loomy-wechat.ts` `4087e60c…1f1f`；`loomy-wechat-login.ts` `f05f9324…6f4f` |
| 客户端逆向样本 | `Loomy-0.9.39-win`（Electron，`resources/app.asar` 解包），检查日期 2026-11-06 |

来源标注约定：**source-confirmed** = 参考实现或客户端常量中明确的代码；**runtime-confirmed** =
参考实现记录的实机抓包结论；**inferred** = 由相邻证据推断；**unknown** = 未取得。

## 1. 来源与许可 (Provenance)

- 参考实现为 MIT（`vendor/channel-pack/LICENSE`，与上游一致）。本次**未复制**其代码：
  仅按协议契约在 Go 中重新实现，保留错误语义与常量来源注释。
- 本仓库同类先例：`xiaohuanxiong`（HTTPS 常量 + 管理端登录 + 只读额度）、
  `lobsterai-server`（托管网关 + 单 wire format + 额度/冷却分离）。

## 2. Provider 身份

- 稳定 key: `loomy`（与 wire format 分离；不伪装成 `openai`/`codex`）。
- 业务/推理：`https://loomyad.xunfei.cn/api/v1`（**source-confirmed**，`loomy.ts` `LOOMY_API_BASE`）。
- 账号：`https://account.xfinfr.com`（**source-confirmed**，`LOOMY_ACCOUNT_BASE`）。
- 推理 wire format：OpenAI Chat Completions + 标准 SSE（**source-confirmed**，
  `loomy.ts` 模块头明确「标准 OpenAI 兼容 + 标准 SSE，无加密、无信封」）。

## 3. 授权 (Authorization) — 手机短信验证码（主）/ 微信扫码

### 3.1 短信验证码（本实现的主路径）

```
POST {accountBase}/login/phone/sendMsgCode  {base, param:{ccode:"86", phone, expire:300}}   → data.msgid
POST {accountBase}/login/phone/checkCode    {base, param:{ccode:"86", phone, mcode, msgid, expire:1209600}}
                                                                                            → data.{session, userid}
```

- `expire: 1209600` = 14 天，是向服务端**声明**的有效期；响应**不含**到期时间戳，
  故 `expired` 由本地按登录时刻 + 14 天推算（**source-confirmed**，`loomy-oauth.ts`）。
- `base` 信封固定字段：`appid: GM3LOOMY`、`modelid: Web`、`version: 1.0.0`、`devid: web`、
  `ua: Loomy|Desktop|Electron|macOS`（Windows 上也是这个值，照抄）、`traceid`（32 位 hex）
  （**source-confirmed**，`buildLoomyAccountBody`）。
- 认证是 **HMAC-SHA1 签名**，头为 `Authorization: account {ak}:{sig}`，不是 Bearer。
  AccessKey ID/Secret 随客户端分发（`loomy-product.ts`），泄露不涉及用户数据
  （**source-confirmed**，`loomy-product.ts` 模块头）。

### 3.2 微信扫码（协议已恢复，本仓库未接线为入口）

```
GET open.weixin.qq.com/connect/qrconnect?appid=…&redirect_uri=<官方地址>&…   → HTML 内嵌二维码 uuid
GET open.weixin.qq.com/connect/qrcode/<uuid>                                → 二维码图片（实测 JPEG）
GET long.open.weixin.qq.com/connect/l/qrconnect?uuid=<uuid>&last=<prev>     → 长轮询状态机
POST {accountBase}/login/thirdAccount/bind/auth      {tcode:{code}, type:"wx"} → {bind, rcode, nickname, headpic}
POST {accountBase}/login/thirdAccount/bind/skip      {rcode, expire}          → {session, userid}   (bind=1)
POST {accountBase}/login/thirdAccount/bind/sendMsg   {rcode, phone, ccode, expire} → {msgid}        (bind=0)
POST {accountBase}/login/thirdAccount/bind/checkCode {rcode, mcode, msgid, expire} → {session,userid,phone}
```

状态机（依据授权页内嵌 JS 的 `switch(window.wx_errcode)`，**source-confirmed**）：

| errcode | 含义 | 动作 |
| --- | --- | --- |
| 408 | 待扫码 | 继续轮询 |
| 404 | 已扫码待确认 | 继续轮询 |
| 405 | **已确认**，`wx_code` 在该帧 | 用 code 换 rcode |
| 403 | 用户取消 | 终止 |
| 402 | 二维码失效 | 重新取 uuid |

`redirect_uri` 必须是官方地址 `https://loomy.xunfei.cn/oauth/wechat/callback`（微信校验域名白名单）；
该页实测 404 且**不需要可达**，code 来自长轮询（**runtime-confirmed**，参考实现）。

本仓库实现了 `FetchWeChatUUID` / `FetchWeChatQRImage` / `PollWeChat` / `BindWeChat*`
（库层完整、有单测），但**未**接线为管理端或 CLI 入口：微信扫码需要一个承载二维码页面的
本地 HTTP 服务，属于面板职责。见 §9。

## 4. 刷新 (Refresh) — 不存在

- **没有 refresh 端点**（**source-confirmed**，`loomy.ts` 模块头明确「没有 refresh 端点」）。
- 实现口径：`LoomyExecutor.Refresh` 只用一次最廉价的只读请求探测存活
  （`GET /points/records?pageNo=1&pageSize=1&recordType=all`，无副作用、不消耗积分）。
- 收到业务码 `100002` → 返回 `401 unauthorized`，让管理器停止选号（而非每次冷却后重试）；
  传输失败**不**判死，原样上抛（网络抖动不该让用户重新登录）。
- `RefreshLead()` 返回 `nil`：探活无法推进固定窗口，注册 lead 会让自动刷新循环每轮立即到期
  反复探活。真实过期由本地 `expired` 字段驱动选号跳过。

## 5. 推理 (Inference)

### 5.1 端点与凭据头（**source-confirmed**，两套头必须记住）

| 端点 | 必需头 |
| --- | --- |
| `POST /chat/completions` | `Authorization: Bearer <session>`；官方客户端**同时**发 `token: <session>` |
| `GET /models`、`/points/*`、`/onboarding/*` | **仅**小写 `token: <session>`（带 Bearer 会回 `100002 缺少 token`） |

本实现：chat 同时发两个头（与客户端一致，且对上游将来改判据更稳）；业务读只发 `token`。

### 5.2 流式

- 标准 SSE：`data:` 帧 + `[DONE]`；`stream: true` 由执行器固定（网关总是流式）。
- 错误可能出现在 HTTP 200 内：裸信封 `{code, desc}` 或 `data:` 帧含 `error`。
  两者都按失败处理，避免「静默空回答」。

### 5.3 思考档位

- 远端 `reasoning_efforts`（实测 8 个 chat 模型一致：`none/low/medium/high/xhigh`），
  请求字段 `reasoning_effort`（**source-confirmed**，`loomy-adapter.ts`）。
- 上游对未知字段**静默忽略**（三种字段名都返回 200），故字段名依据是远端自己的命名。
- 默认档：参考实现有意选 `high`（远端声明 `low`）；本仓库**不**在协议层覆盖默认档，
  由 canonical thinking 管线决定，模型目录声明完整档位集合，`none` 标记为可关闭。

## 6. 模型 (Models)

- 目录来源：`GET {apiBase}/models`（**source-confirmed**），按 `type === "chat"` 过滤；
  **不能**用输入模态过滤（5 个 chat 模型支持图片输入，与生图模型无关）。
- 字段映射：`name` → 展示名；`context_length` → 上下文；`reasoning_efforts` → 档位；
  `capabilities.reasoning` → 是否声明推理；`capabilities.input_modalities` 含 `image` → 视觉。
- 兜底目录：`internal/registry/models/models.json` 的 `loomy` 分节（2026-09-26 实测 chat 切片的重新生成版，当前 6 条）。
  远端拉取失败或为空时保留兜底表并延期重试，**不**撤销全部模型（那会让凭据不可用）。
- 已知分歧：`spark-x` 远端声明 1048576，Loomy 客户端本地表强制 262144；兜底表**采信远端**，
  并在描述中记录该分歧。

### 6.1 客户端逆向证据（Loomy-0.9.39-win，2026-11-06）

解包 `resources/app.asar` 还原了客户端取模型列表的调用链（**source-confirmed**，客户端代码）：

- 端点与鉴权印证 §5.1：`GET {baseURL}/models`，头**仅**小写 `token: <session>`
  （`authMode:'token'`，不发 Bearer），另带 `Accept/Content-Type: application/json`，
  10s 超时（`electron/model-service.js` `fetchModelsFromProvider`）。
- `baseURL` 不是模板常量：运行时由 `pointsConfig.baseUrl` 强制覆写 provider `imodel`
  （`electron/bundled-resources.js`），非 `/api/v1` 结尾则追加——即模型列表与积分接口
  共用同一 baseUrl 解析链；生产值 `https://loomyad.xunfei.cn/api/v1`（见 §7.1），
  `opencode-home-template/opencode.json` 里的 `ossptest…/api/v1` 只是被覆盖的测试默认值。
- 刷新链：renderer `window.electronAPI.refreshSessionModels()` → IPC
  `opencode:config:refreshSessionModels` → `syncSessionAuthProviderModels`，只对
  `useSessionAuth:true` 的主 provider 拉取，anthropic companion 由主 provider 派生。
- 响应为 OpenAI 兼容 `{data:[…]}`；客户端消费字段：`id`/`name`、`protocol`
  （按 openai/anthropic 分桶路由）、`is_default_model`/`is_default_small_model`
  （服务端指定默认主/小模型，本仓库未消费）、`context_length`、`input_token_limit`、
  `capabilities.input_modalities`/`output_modalities`/`function_calling`、`reasoning`、
  `tool_call`、`attachment`。
- **与 §6 的差异**：客户端**不按 `type === "chat"` 过滤**，而是按 `protocol` 分桶；
  `type` 过滤是参考实现（dsh channel pack）的口径。两者对 5 个带图片输入的 chat 模型
  结论一致（都不按输入模态过滤）。
- 失败兜底：HTTP 404/500 时按 host 匹配内置 fallback 表（星火等不支持 `/models` 的
  供应商），不视为致命——与本仓库「远端拉取失败保留兜底目录」的策略同构。
- 可执行验证脚本：`docs/scripts/loomy_models.py`（模型列表 + §7.1 额度读取）。

## 7. 额度 (Quota) — 三个独立职责

| 职责 | 本实现的来源 | 说明 |
| --- | --- | --- |
| 用量统计 | 共享 `ExecutorUsageReporter` + SSE usage | 与其它 provider 一致 |
| 额度观测 | `GET /points/records?…`（只读） | 两个池分别返回：`permanent`、`daily`；读接口不返回的每日窗口字段（`daily_quota`/`daily_consumed`/`daily_cycle_date`）**缺省**，未知与 0 严格区分 |
| 调度/冷却 | 上游 402/`积分不足` → `402 + credentialScoped` | 冷却凭据而非单个模型；本地 `POST /quota/reset` 只清路由状态，不清上游积分 |

- `POST /points/first-login` 是**写**端点（每日额度初始化，幂等，判据 `alreadyProcessed`），
  只在登录后尽力调用一次，**不**在面板打开时调用（**source-confirmed**，`loomy-credits.ts`）。
- `dailyQuota` 只在写端点响应里出现；读接口缺省时**不**硬编码 5000。

### 7.1 客户端逆向证据（Loomy-0.9.39-win，2026-11-06）

解包 `resources/app.asar`（Electron 壳，业务逻辑全在 JS）逐层还原了积分详情页取额度的调用链，
与 §4/§7 的口径相互印证（**source-confirmed**，客户端代码）：

- 取额度接口即本实现探活/额度观测所用的同一端点：
  `GET {pointsBase}/api/v1/points/records?pageNo=1&pageSize=1&recordType=all`，
  头**仅**小写 `token: <session>`（`electron/points-service.js` `queryRecords` →
  `request()`；渲染层 `o9()` 用 `pageSize=1` 的流水查询同时取余额）。
- 原始响应字段（非本仓库归一化后的名字）：`data.balance` = 永久/长期积分（客户端强制
  `Number.isFinite`，缺失即报错），`data.dailyBalance` = 每日赠送积分；另有
  `dailyLimitPoints`/`dailyConsumedPoints`/`dailyRemainingPoints`（个人）与 `teamDaily*`
  （团队）被客户端以「可空」方式解析（非有限值即丢弃）。故静态证据只能证明读接口
  **可能**返回每日窗口字段，不能证明必然出现——§7「缺省即未知」的口径保持不变。
  客户端把首条响应缓存到 localStorage `loomy-points-summary`。
- 业务码语义印证 §4：`code !== '000000'` 为失败，`100002` = 未登录/登录失效
  （客户端触发 re-login 后重试）。
- 生产 base URL 证据：`resources/.env.prod` 为 `LOOMYENC1:` 前缀的 AES-256-GCM 密文
  （口令硬编码于 `electron/utils/env-file-crypto.js`，客户端自述仅为混淆），解密得
  `LOOMY_POINTS_BASE_URL=https://loomyad.xunfei.cn`，与 §2 `LOOMY_API_BASE` 一致；
  安装包内默认值 `points-config.js` 指向测试环境 `ossptest.voicecloud.cn/loomy/integration`。
- 其余积分端点（同 base、同 `token` 头）：`/api/v2/points/records`（`direction=debit|credit`、
  `granularity=msg`，消耗/获得标签页）、`/api/v1/team-points/balance`（`currentBalance`）、
  `/api/v2/team-points/records`、`/api/v1/points/by-chat?chatId=`、`/api/v1/points/by-turn`、
  `/api/v1/points/first-login`、`/api/v1/points/redemption-codes/redeem`、
  `/api/v1/points/activation`、`/api/v1/pet-work(/rewards)`、`/api/v1/invitation-codes`、
  `/api/v1/user/identity`。客户端调用链：renderer `window.electronAPI.points.*` →
  IPC `points:query-records` 等（`electron/ipc/account-business-ipc.js`）→ `PointsService`。

## 8. 注册清单 (Registration checklist)

| 项 | 位置 | 状态 |
| --- | --- | --- |
| Provider key | `internal/constant/constant.go`（`Loomy`） | 已实现 |
| 协议客户端 | `internal/auth/loomy/{loomy,sign,login,token}.go` | 已实现 |
| SDK 认证器 | `sdk/auth/loomy.go` + `service_auth.go` + `internal/cmd/auth_manager.go` | 已实现 |
| CLI 入口 | `--loomy-login`（`cmd/server/main.go`、`internal/cmd/loomy_login.go`） | 已实现 |
| 管理端登录 | `POST /v8/management/loomy-sms-code`、`POST /v8/management/loomy-login`（v0 同路径） | 已实现 |
| 管理端额度 | `GET /v8/management/loomy-quota`（v0 同路径） | 已实现 |
| 管理端 auth-url 提示 | `StartOAuthV8` 对 `provider=loomy` 返回 SMS 指引 | 已实现 |
| 凭据合成 | `internal/watcher/synthesizer/file.go`（`base_url`、`access_token`、OAuth kind） | 已实现 |
| 执行器 | `internal/runtime/executor/loomy_executor.go`（完整 `ProviderExecutor`） | 已实现 |
| 执行器绑定 | `sdk/cliproxy/service_executors.go` | 已实现 |
| 模型注册 | `sdk/cliproxy/service_models.go` + `internal/registry/model_definitions.go`（共享目录 `loomy` 分节） | 已实现 |
| 目录探活/刷新 | `sdk/cliproxy/loomy_models.go` + `service_lifecycle.go` | 已实现 |
| 管理端模型目录 | `internal/api/handlers/management/provider_models.go` | 已实现 |
| 翻译器 | 复用既有 OpenAI 翻译器（无新增） | 不适用 |
| 思考管线 | 复用 canonical 管线 + 模型 `thinking.levels` | 不适用（无新 applier） |
| 外部前端 UI | `frontend/`（独立仓库 Cli-Proxy-API-Management-Center 的检出） | 已实现（登录卡片 + 额度卡），见 §9 |
| CLIProxyAPIHome | 未提供源码 | **未验证** |

## 9. 未实现 / 未验证

1. **管理面板 UI**：已实现于 `frontend/`（独立仓库 `Cli-Proxy-API-Management-Center` 的检出，本仓库不跟踪其内容）。
   - 登录页 `src/pages/OAuthPage.tsx` 新增 Loomy 卡片：手机号 → `POST /loomy-sms-code`，验证码 →
     `POST /loomy-login`；不走共享的 `oauth/auth-url` 流程（`startAuth` 对 `loomy` 直接转短信步骤）。
     客户端模块 `src/services/api/loomy.ts` 只持有 `msg_id`，登录响应只有 auth 文件名，凭据不进页面。
   - 额度页 `src/features/quota/providers/loomy/` 读取 `loomy-quota`，两个池分列；
     `daily_quota`/`daily_consumed` 缺省即「该读接口不返回」，渲染时省略该行而非显示 0。
   - 该检出与后端分开提交/发布；后端不依赖其产物（`dist/` 由发布流程生成 `management.html`）。
2. **微信扫码入口**：库层完整（含 404/405 状态机单测），但没有承载二维码页面的本地服务，
   因此没有对外入口。若需要，应在管理面板侧实现（参考实现的弹窗页是它自己的做法）。
3. **TUI**：TUI 的登录页只支持 URL / 设备码流程（`qwen-web` 的表单登录同样不在其中），
   Loomy 沿用同一先例，未加入 TUI。
4. **live 验证**：无真实账号，未做 OAuth 登录、真实推理、真实额度读取与冷却恢复验证。
   `internal/auth/loomy`、执行器、管理端点均为 mock 验证。
5. **Home 协调**：未提供 CLIProxyAPIHome 源码，凭据刷新/元数据/模型/冷却的协调改动未评估。

## 10. 验证命令与结果

已运行（2026-10-20）：

```bash
gofmt -w ./internal ./sdk ./cmd
go build -o /tmp/cpa-build-check ./cmd/server   # 通过
go test ./internal/auth/loomy/ ./sdk/auth/ ./internal/registry/ \
        ./internal/runtime/executor/ ./internal/api/handlers/management/ \
        ./internal/watcher/... ./sdk/cliproxy/...   # 全部通过
```

本地端到端（mock 上游 `127.0.0.1:18318`，真实 `cli-proxy-api` 进程）：

| 检查 | 结果 |
| --- | --- |
| `/v1/models` 经目录探活后只剩上游声明的模型 | 通过（`spark-x`） |
| OpenAI Chat 非流式 | 通过（`hello loomy`，usage 正确） |
| OpenAI Chat 流式 | 通过（增量帧 + `[DONE]`） |
| Anthropic `/v1/messages` 非流式 + 流式 | 通过（`content_block_*` / `message_delta` / `message_stop`，usage 正确） |
| chat 头上游实收 | `Authorization: Bearer <session>` + `token: <session>` |
| 业务读头上游实收 | 仅 `token: <session>` |
| HTTP 200 内嵌 `error` 帧 | 502 上报，非空回答 |
| HTTP 200 内嵌 `100002` | 401 `invalid_api_key`，凭据进入 `unauthorized` 冷却 |
| HTTP 200 内嵌 `积分不足` | 402，凭据级冷却，另一账号继续服务（failover 200） |
| 管理端 `loomy-quota` | 200，两池分列；读接口不返回的每日窗口字段**缺省**（`{permanent,daily,total}`） |
| 管理端 `loomy-sms-code` 无密钥 | 401（受管理鉴权保护） |

管理面板（`frontend/`，同一后端进程 + mock 上游）：

| 检查 | 结果 |
| --- | --- |
| `bun run build` | 通过（单文件 `dist/index.html`） |
| `bun test tests/loomyOAuthUi.test.ts tests/loomyQuota.test.ts` | 通过 |
| 登录页 Loomy 卡片 | 渲染标题/提示/手机号输入，无裸 i18n key |
| 发送验证码 | `POST /v8/management/loomy-sms-code` 发出，卡片转为验证码输入 + Resend |
| 提交验证码 | `POST /v8/management/loomy-login` 发出，状态变 `Authentication succeeded!` |
| 额度页 Loomy tab | 计数为 2；查询一张卡后渲染 `Total 19,992 / Permanent 15,000 / Daily 4,992`，不显示未返回的每日上限 |
| 读接口缺省每日窗口 | `daily_quota`/`daily_consumed` 不出现（mock 返回 `{permanent,daily,total}`） |

未运行全量 `go test ./...` 的结论：仓库存在**与本改动无关**的既有失败（基线 worktree 复现）：
`internal/client/codex/models` 的 `TestCodexClientModelsResponseAppliesMaxContextLengthOverride`、
`internal/util` 的 `TestInPlaceByteWritesAreReviewed`。
