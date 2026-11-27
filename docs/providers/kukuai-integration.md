# 库库AI (KuKu AI / GenFlow) Provider — 协议契约与集成证据

Provider key: `kukuai`
状态: 协议已恢复 (source-confirmed)；实现未开始。

## 0. 快照 (Snapshot)

| 项 | 值 |
| --- | --- |
| CLIProxyAPI commit | `7e6579fb2a23975842b8f2c8a193fa7b169d0baf` |
| 工作区状态 | dirty: `M test/chat-completions-test.http` (与本任务无关, 保留) |
| 客户端安装包 | `/run/media/van/Alaska/a/kukkuAI/KukuAI_1.6.6.132.exe` (NSIS, 160,734,104 bytes) |
| 解包产物 | `/tmp/kuku_analysis/app/` (`resources/app.asar` → out/main, out/preload, out/renderer) |
| 客户端版本 | 1.6.6 (win_version 1.6.6.132), 构建 2025-08/2026-09 |
| 检查日期 | 2026-09-29 |
| 归属 | Baidu「库库AI / GenFlowPro」; 非开源, 仅本地静态分析 |

来源标注约定: **source-confirmed** = 本地 JS/二进制中的明确代码; **inferred** = 由相邻证据推断; **unknown** = 未取得。

## 1. 来源与许可 (Provenance)

- 客户端为闭源商业软件 (Baidu)，无可用开源许可。分析仅用于协议互操作，**不复制**其代码。
- 本仓库内同类先例: `deepseek-web` (`docs/deepseek-web.md`)、`xiaohuanxiong` (`internal/auth/xiaohuanxiong/`)、`qwen-web`。复用本仓库自身的抽象，不搬运第三方实现。

## 2. Provider 身份

- 稳定 key: `kukuai` (与 wire format 分离；不要伪装成 codex/claude)。
- 账户身份: 百度账号 `uk`/`uid` + `BDUSS`；企业空间 `cid`。
- 区域: 目前仅中国区 `kuku.baidu.com`。**source-confirmed**。

## 3. 授权 (Authorization)

库库AI 的推理后端**不使用标准 OAuth**，而是百度 passport **Cookie 会话**。存在两条独立的授权路径，必须区分:

### 3a. LLM 推理凭据 = passport Cookie (source-confirmed)
- 登录入口页: `https://kuku.baidu.com/genflowpro/login`
- 关键 Cookie: `BDUSS`, `BDUSS_BFESS`, `PTOKEN`, `PTOKEN_BFESS`, `PASS_STOKEN`, `STOKEN`
  - 来源: `out/main/main.js` L3714-3723 (`mo`/`_o` 列表), L3799-3805 (`Eo()` 抽取)。
- 业务 STOKEN 交换 (source-confirmed + native-confirmed):
  - `genflow_engine_get_login_auth_param(bduss, ptoken)` (genflowengine.dll) → `appid=%d&bduss=%s&ptoken=%s&return_type=1&tpl=%s&tpl_list=%s&sig=%s`
    - 常量: `pass_appid` 默认 1(占位 `%d`), `pass_tpl`=`genflowpro`, `pass_key`=`34868b316de55815273b2616954b0286` (硬编码)。
    - `tpl_list` 由 `pass_tpl`+`pass_key` 经 `tolower` 变换得出；`sig` = 对拼接串做哈希后小写 (疑为 MD5/BCrypt 系)。
    - 来源: IDA `sub_1801C6DF0` @ `genflow_engine_get_login_auth_param_0` (0x180089910)。
  - `POST https://passport.baidu.com/v3/login/api/auth` (form-urlencoded) → 取 `data.stoken_list.genflowpro`。
    - 来源: main.js L4326-4360 (`exchangeStoken`)。
- 通用参数 `createAuthCommonParams`: `app_id=123971023`, `channel=chunlei`, `version=<app>`, `clienttype=<device>` (main.js L4411-4424)。**source-confirmed**。
- Cookie 持久化: electron session `persist:genflow-auth`。

### 3a'. 手机号 + 短信登录 — 实现于百度 passport 官方面页, 不在此客户端内

**重要结论: 桌面端的手机号/短信登录不是库库AI 自己的代码，而是嵌入的百度 passport 官方登录页与 SDK。**

证据链 (source-confirmed):
1. 主进程把 `https://kuku.baidu.com/genflowpro/login` 作为 Electron `BrowserWindow` 加载 (main.js `loadInitialLoginPage`/`loadLoginUrl` @ L4200-4310；`login-webview.js` preload)。
2. 该页面由远程脚本渲染: 页面只含一个脚本 `https://nd-static.bdstatic.com/m-static/genflow-login/js/index.o61gd118.js`
   (sha256 `10bc7f04...43e2a4`)。本地 asar 中**无**任何短信端点/中文登录文案。
3. 该脚本加载百度 passport 官方 SDK: `https://passport.baidu.com/passApi/js/wrapper.js` (sha256 `6d0322cc...0bebd9`)，
   并以 `product=genflowpro, loginVersion=v5, tpl=netdisk, sms:5, qrcode:0, ...` 初始化 `passport.use('login', ...)`。
4. SDK 再加载 v5 登录模块 `.../passApi/js/loginv5_bf939ca.js` (sha256 `b94828f5...84da`)，短信 UI 与接口都在其中。
5. 全量本地搜索 (`genflowengine.dll`/`kernel.dll`/`netbase.dll`/`app.asar`/renderer) **仅**出现主机 `passport.baidu.com`/`wappass.baidu.com`，
   **零**个 `/wp/api/login/sms`、`senddpass`、`logincheck` 等路径 → 确认短信流程不在此客户端实现。

passport v5 的短信登录接口 (source-confirmed 于 `loginv5_bf939ca.js`):

| 步骤 | 方法 | 端点 | 关键参数 |
| --- | --- | --- | --- |
| 获取会话/token | GET/JSONP | `https://passport.baidu.com/v2/api/?getapi` | `tpl=netdisk` → 返回 `bdstoken` 等 |
| 发送短信验证码 | JSONP | `https://passport.baidu.com/v2/api/senddpass` | `gid, username=手机号, countrycode, bdstoken, tpl, loginVersion=v5, flag_code, client, mkey, moonshad, ds, tk, s, k, supportdv=1, dv` |
| 语音验证码(可选) | JSONP | `https://wappass.baidu.com/wp/api/login/sms?is_voice_sms=1` | 同上 |
| 提交登录 | JSONP/GET | `https://passport.baidu.com/v2/api/?login` | `username, password(短信=验证码), tpl, staticpage, codestring, verifycode, loginVersion=v5, apiver=v3, safeflg, logintype, ...` |
| 图形/短信验证码串 | GET | `https://passport.baidu.com/v2/?reggetcodestr` / `?regsmscodestr` | `codestring` |
| 校验验证码 | JSONP | `https://passport.baidu.com/v2/?checkvcode` | `verifycode, codestring, loginVersion=v5` |
| 图形验证码图片 | GET | `https://passport.baidu.com/cgi-bin/genimage?<params>` | — |
| 设备指纹 | 本地 | `https://wappass.baidu.com/static/touch/js/lib/fingerprint.js` + 隐藏域 `dv_Input` | `dv`, `supportdv=1` |
| 短信反爬参数 (本地可复现) | 本地 | `cryptdata = md5(phone + 'Moonshadow')` → 依次对**首个匹配**做替换: `o→ow`, `d→do`, `a→ad`, `h→ha`, `s→sh`, `n→ns`, `m→mo`；作为 `moonshad` 参数 | 来源 `loginv5_bf939ca.js`: `hex_md5(phone.value+'Moonshadow')` 后链式 `.replace(/x/,'..')` |
| 失败跳转/换绑 | — | `https://passport.baidu.com/v2/?reg&regphone=1&tpl=netdisk&u=<callback>`；`/m-static/genflow-login/thirdparty/pass_v3_jump.html` | — |

登录成功后 passport SDK 在 `passport.baidu.com`/`.baidu.com` 域写入 `BDUSS`(+`STOKEN` 经 `v3/login/api/auth`)；客户端通过
`authSession.cookies.get({url:'https://kuku.baidu.com'})` 抽取 `Eo()` 字段，再 `setLoginState` 给原生引擎。

**可行性结论 (关键):** 手机号+短信登录可以实现，但它是**百度 passport 通用流程**(不是库库AI 专有协议)，需要:
1. 处理 `dv` 设备指纹 + `moonshad` 反爬变换 + `codestring/verifycode` 风控；
2. 处理 JSONP 回调、`bdstoken` 会话与 `countrycode`；
3. 不实现「图形验证码」交互则无法稳定通过风控——**headless 自动化风险高，可能触发百度账号风控/封禁**。

> 按 SKILL 规则: 这仍是**Cookie 会话凭据**，不是 OAuth。本 Provider 应支持「导入 BDUSS(+STOKEN)」为一等入口；
> 手机号登录只能作为**可选辅助**(依赖第三方 passport 页变动，且属百度非公开接口)，需用户明确同意并自行承担账号风险。

### 3b. Connector 子系统的百度 OAuth (与 LLM 推理无关, source-confirmed)
- `https://openapi.baidu.com/oauth/2.0/authorize`，`response_type=token`(**implicit**)，`scope=basic,netdisk`，
  `client_id=f7be8LH2EtOLDttPV53JKj3sVEOcaBAA`，loopback 端口 `[47801,48501,49001,63399]`，回调 `/callback`、`/callback/deliver`，
  `qrcode=1`；产物是 `KUKU_BDPAN_CLI_TOKEN` (Bearer)，仅用于连接百度网盘的**外部 MCP/知识库** connector。
- 来源: main.js L27538 (`HE`/`UE`)、L27560-27710 (flow)、L27704-27710 (`QE` clientId/ports)。
- **结论**: 该 OAuth 不能用于 `/wenchain/genflow/chat` 推理。仅当用户要求"库库AI 内置网盘 MCP 工具"时才需要实现。

## 4. Token 交换 / 刷新

- 推理凭据为 Cookie，无独立 access/refresh token。
- 刷新 = 重新获取/校验 Cookie。`BDUSS` 有效期长但可失效；STOKEN 可过期。
- `Refresh` 语义: 校验 Cookie 是否仍有效（对 `GET /api/genflowpro/ping` 或 `settings/profile` 探测），失效则标记 need-relogin，**不伪造刷新**。**inferred**。
- 终端失败: 401/未登录 → 需要用户重新导入 Cookie。

## 5. 推理 (Inference) — source-confirmed

传输说明 (重要, 影响实现):
- 业务 API (`/api/*`, `/bizapi/*`, `/wenchain/genflowpro/*`) 由渲染进程经**原生 C++ HTTP** (`callNative('httpDownloadMemory')`) 发出:
  - 公共参数作为 **query 参数** 追加: `Qu(...commonParams, e.params)`；`commonParams = {app_id:123971023} + scopeParams(device/cid)`。
  - POST 的 `params` 默认以 `application/x-www-form-urlencoded` 编码为 **body** (`$u`)，`json:true` 时用 `application/json`。
  - Cookie 通过 native 层 (`headerCookies`) 附加。
  - 来源: `out/renderer/assets/index-CO77MoJW.js` (`Gu` http client, `commonParams`, `$u`, `ed`)。
- 推理 chat (`/wenchain/genflow/chat`) 由渲染进程用**浏览器 fetch** 直接发出 (`credentials:'include'`), **不走** native HTTP。
  - 来源: `out/renderer/assets/dist-hoFjsj6Q.js` (`ee()` transport)。

- 端点: **`POST https://kuku.baidu.com/wenchain/genflow/chat`**
- 头部: `Content-Type: application/json`, `Accept: text/event-stream`；凭据为 Cookie (`credentials: include`)。
- 来源: `out/renderer/assets/dist-hoFjsj6Q.js` (transport `ee()`, path 常量 `/wenchain/genflow/chat`)。

### 请求体 (来源: `useInputToEditorOps-CD8wHctf.js` `getSendMsgParams`)
```json
{
  text, rich_input_params, fsid:[], quotes:[], skills:[], experts:[],
  model_name, model_display_name, think_mode,
  project_id, session_id, reply_id,
  show_text, device_id, premake_data,
  sheet_selection?, project_category?, template_name?, security_id?, market_code?
}
```
- `model_name` = 模型 id（来自 model/list 的 `model_name` 字段，见 §7）。
- `think_mode` = 思考档位 id（`think_list` 的 `id`）。
- `session_id` 复用同一次对话；`reply_id` 为轮次 id。
- `premake_data` = `{premake_query_id, scene?}`（防滥用/相关推荐，见 §9）。

### 流式响应 (SSE, 自定义分帧)
- 帧以 **两个空格** 分隔；每帧行: `event:<TYPE>` / `data:<JSON>` / `id:<ID>`。
- `data` 为 `[DONE]` 或 JSON。JSON 解析 (`re`): `event_type ?? event`, `data ?? details.run_object`, `status`, `message`, `content`。
- `content` 归一化: 字符串直接输出；数组取 `{type:"text", text:{value}}` 的 `text.value` 拼接。
- 事件类型枚举 (source-confirmed):
  `USER_QUERY, USER_STEER, USER_ANSWER, SESSION_INFO, SESSION_TITLE, REPLY_START, REPLY_END,
   MODEL_CALL_START, MODEL_CALL_END, TEXT_BLOCK_START/DELTA/END, THINKING_BLOCK_START/DELTA/END,
   TOOL_CALL_START/DELTA/END, TOOL_RESULT_START/TEXT_DELTA/END, PLAN_UPDATE, ARTIFACT,
   REQUIRE_USER_CONFIRM, USER_CONFIRM_RESULT, REQUIRE_EXTERNAL_EXECUTION, EXTERNAL_EXECUTION_RESULT,
   AWAITING_INPUT, HINT_BLOCK, ERROR, STREAM_ERROR, SUMMARY_REFERENCES, LIST_REFERENCES, RISK,
   COMMERCIAL, COMMERCIAL_OFFICEAGENT, COMMERCIAL_NOTICE, COMMON_NOTICE, FINANCE_SIGNAL,
   CONTEXT_COMPACTED, ESTIMATE_POINT, ACTUAL_POINT, TURN_DONE, CUSTOM_CARD`
- 另有较早/外层信封 `message | thoughts_execute | tool_call_info | done | run_end | error` (dist 顶部 `v`)。
- 翻译映射: `TEXT_BLOCK_DELTA → content`, `THINKING_BLOCK_DELTA → reasoning_content`, `TURN_DONE/REPLY_END → finish`.

## 6. 会话 / 工具

- 有状态: 服务端 `session_id`；会话 CRUD 在 `/api/genflowpro/workspace/*` 与 `/wenchain/genflowpro/*`。
- 工具: `TOOL_CALL_*` / `TOOL_RESULT_*` (服务端工具)；`REQUIRE_EXTERNAL_EXECUTION` = 客户端沙箱/浏览器 Agent 执行。
- 思维链: `THINKING_BLOCK_*`。
- **字段 (source-confirmed)**: 增量文本在 `data.delta` — `TEXT_BLOCK_DELTA`→正文，`THINKING_BLOCK_DELTA`→思维链。
  来源: `out/renderer/assets/runtime-BF9xEphj.js` (`typeof e.data.delta=='string' ? e.data.delta : void 0`)。
- 当前实现每轮新建 `reply_id` (UUID)、`session_id=""`；后续可缓存 session 实现原生多轮会话。

## 7. 模型 (Models) — source-confirmed

- 列表: `GET /wenchain/genflowpro/model/list?channel=genfp` (搜索场景 `channel=kuku_pc_genflowpro_search` / `kukutob_pc_genflowpro_search`)。
- 响应: `data.model_list[]`, `data.think_list[]`, `data.default_model_id`, `data.default_think_id`。
  - 模型项: `{id, model_name, display_name, description, cost_ratio, vip_type===1→vipOnly}`。
  - 思考项: `{id, think_name, description}`。
- UI 实际展示 (截图): `DeepSeek-V4.1-Flash, DeepSeek-V4-Pro, DeepSeek-V4-Flash, GLM-5.3-Flash, GLM-5.3(会员优先), GLM-5.2`，含 `cost_ratio` 倍率与会员锁。
- **必须运行时拉取并注册**；禁止硬编码静态表 (skill gate)。

## 8. 额度 / 积分 (Quota) — source-confirmed

真实接口 (bizapi 前缀, 仍走 `main` host):
- `GET /bizapi/gfpro/getgfvipremain` → 主额度快照。
  - 响应 `data`: `isVip, isExpire, isTrial, vipType, vipStartTime, vipEndTime, isAuthPay, authPayType, authPayTime, authGoodsId, list[], businessTotalPoint, businessLeftPoint, busAllLeftPoint`。
  - `list[]`: `{assetType, assetName, totalPoint, bonusPoint, vipPoint, chargePoint, freezePoint}`；
    `assetType` 枚举 `[1,2,3,4]` (3=?, 4=?)。来源: `use-cashier-*.js` zod schema `x`,`b`。
- `GET /bizapi/gfpro/gfgetpointexpendconf?source_app=...` → 每分钟消耗配置 (point→second)。
- `GET /bizapi/gfpro/gfpoint2second?source_app=...` → `data.point2second.pricePerMinute` (语音/笔记计费)。
- 企业: `/bizapi/gfpro/getbusinessquota|getbusinessinfo|getbusinessdetails|getbusinessgroupconsume|getbusinessdeplist|savebusinessuserquota`。
- 免费积分: `/api/genflowpro/freepoint/{homenew,inviteCode,inviteBind,rewardClaim,taskComplete}`。
- 流式内额度事件: `ESTIMATE_POINT` (预估), `ACTUAL_POINT` (实际扣减/剩余)。
- 权益校验: `GET/POST /api/genflowpro/settings/authorize` (`{feature}`), `GET /api/genflowpro/settings/featureauth`, `GET /api/genflowpro/settings/profile` (含 `vipInfo`, `tokenRemain`)。

> 额度方向: `*LeftPoint`/`busAllLeftPoint` = 剩余；`cost_ratio`/`pricePerMinute` = 单位换算。reset 语义 = 会员周期 (`vipEndTime`)，**inferred**。
> 按 SKILL: 使用 `getgfvipremain` 作为**主动探测** (authenticated probe)，保留 observed time/units；`ACTUAL_POINT` 为**被动信号**；两者与调度 cooldown 分离。

## 9. 反滥用 / 可行性风险 — 需重点标注

- `premake_data` / `premake_query_id` / `scene`，`device_id`，`show_text` 可能参与上游风控 (inferred，未取得服务端校验证据)。
- `GET /wenchain/genflow/idallochstr` (会话预分配) 出现在发送前流程。
- 客户端使用 electron 指纹 + CDN 反爬 (amisConf/`/act/api/conf`)。**unknown**: 是否存在 header 级签名；静态未见 `X-Sign` 类，但需运行时抓包确认。
- 结论标记: 该 Provider 属于 **private/undocumented web API**，与 `deepseek-web` 同类风险 (ToS + 账号风控)。文档必须明示。

## 10. 错误 (Errors)

- 信封: `{errno, errmsg, show_msg}` (main.js 多处) 或 `{status:{code,msg,sLogid}, data}` (wenchain)。
- 未登录 → 401/需重登；额度不足 → `insufficient_type: automation|points` (`point_sufficient`)。
- **inferred**: 429 语义未在静态中确认。

## 11. 验证计划

- 单元: httptest mock `/wenchain/genflow/chat` SSE 帧 (含 split/空帧/`[DONE]`/in-band ERROR/工具轮次)；
  mock `/wenchain/genflowpro/model/list`、`/bizapi/gfpro/getgfvipremain`。
- 实时 (需用户授权账号): 导入 Cookie → model/list → 非流式 → 流式 → 工具轮次 → 额度探测。逐项记录。

## 12. 实现状态 (截至本次改动)

已实现并 mock 验证 (未 live 验证):

- `internal/constant`: `KukuAI = "kukuai"`。
- 配置: `KukuAIKey = CodeBuddyCNKey` (`kukuai-api-key`)，含 `SanitizeKukuAIKeys`、weight 校验、`config.example.yaml` 示例。
- 凭据合成: `internal/watcher/synthesizer/config.go` `synthesizeKukuAIKeys` (api_key / base_url / prefix / weight / models_hash / proxy)。
- 执行器: `internal/runtime/executor/kukuai_executor.go` (`ProviderExecutor`: Execute / ExecuteStream / Refresh / CountTokens / HttpRequest)。
  - 请求体: `{text, model_name, model_display_name, think_mode, session_id, reply_id, device_id, ...}`。
  - 认证: `Cookie` 头 (BDUSS/STOKEN)，支持原始 Cookie 串或 JSON。
  - 翻译: GenFlow 事件 → OpenAI chunk (`data.delta` → `content` / `reasoning_content`；`TURN_DONE` → `stop`)。
- SSE helpers: `internal/runtime/executor/helps/kukuai_sse.go` (双空格 + 换行双分帧 + 平衡 JSON 提取)、`kukuai_stream.go` (聚合器 + 非流式组装)。
- 模型: `internal/registry/kukuai_models.go` + `models/kukuai_models.json` (嵌入，避免远端目录刷新覆盖)；`GetKukuAIModels` / `LookupKukuAIModel`。
- 注册: `service_executors.go` (case + `baselineExecutorAuths`)、`service_models.go` (`buildKukuAIConfigModels` + `resolveConfigKukuAIKey`)、`service_provider_models.go`、`management/provider_models*.go`、`sdk/config` 别名。
- 测试: executor (Execute/ExecuteStream/错误/Refresh/CountTokens)、helps (分帧/拆包/聚合/in-band error)、registry、synthesizer。

未实现 / 待办:

- **live 验证** (需用户授权 BDUSS): model/list 真实 id、非流式/流式真实链路、工具轮次、额度。
- 额度 (`/bizapi/gfpro/getgfvipremain`) 主动探测 + `ACTUAL_POINT`/`ESTIMATE_POINT` 被动信号接入 `quota_signals.go`。
- 管理面凭据导入入口 (目前仅 config，与 `deepseek-web` 一致)；CLI/TUI 登录项。
  注: `config_auth_index.go` 未对 `deepseek-web` 做特殊处理，故 `kukuai` 保持同等水平即可。
- 前端管理面 (Cli-Proxy-API-Management-Center) 表单/logo。

职责区分 (未合并): 用量计账 / 额度观测 / 调度 cooldown 保持独立。

### 冒烟示例

```yaml
kukuai-api-key:
  - api-key: "BDUSS=<value>; STOKEN=<value>"
    weight: 1
```

```bash
curl -s http://127.0.0.1:8317/v1/models -H 'Authorization: Bearer <proxy-key>' | jq '.data[].id' | grep -E 'glm-5.3|deepseek-v4-pro'

curl -s http://127.0.0.1:8317/v1/chat/completions \
  -H 'Authorization: Bearer <proxy-key>' -H 'Content-Type: application/json' \
  -d '{"model":"glm-5.3","messages":[{"role":"user","content":"hi"}],"stream":true}'
```

## 13. 集成清单 (registration checklist)

- [x] `internal/constant`: `KukuAI = "kukuai"`
- [x] 配置类型 + sanitize + weight + example
- [x] 凭据合成 (`synthesizeKukuAIKeys`)
- [x] 执行器 (`kukuai_executor.go`)
- [x] SSE helpers (`helps/kukuai_sse.go`, `helps/kukuai_stream.go`)
- [x] 模型目录 (`registry/kukuai_models.go` + json)
- [x] 执行器绑定 (`registerExecutorForAuth` + `baselineExecutorAuths`)
- [x] per-auth 模型 (`service_models.go` + `service_provider_models.go`)
- [x] 管理模型目录 (`provider_models*.go`)
- [x] SDK 别名 (`sdk/config/config.go`)
- [x] v8 布局迁移 (`internal/config/config_v8.go` `v8KeyFamilies`) — 否则凭证会被迁移安静丢弃
- [x] 主动刷新 lead (`sdk/auth/refresh_registry.go`, 6h 被动校验)
- [x] 上游格式列表 (`conductor_execution.go` 显式列出 `kukuai`)
- [x] 单元测试 + `gofmt` + `go build ./cmd/server`
- [ ] 额度探测 (helps + management 路由)
- [ ] 管理 API/UI 导入入口
- [ ] live 验证
- 前端管理面 (Cli-Proxy-API-Management-Center, `/home/van/github/router-for-me/Cli-Proxy-API-Management-Center`) 需与后端对齐。

## 14. 运行时确认 (live-verified, 2026-09-30)

用导入的 BDUSS cookie 对生产环境实测。**仅 BDUSS 即可**（无需 STOKEN）。以下均为 runtime-confirmed。

### 14.1 主链路是「SSE 控制通道 + envelope」，**不是**单个 REST 调用

经修正：`/wenchain/genflow/chat` 在生产为 **404**（客户端 SDK 内部路径，未在线上暴露）。真实链路：

1. `GET https://pan.baidu.com/rest/2.0/xpan/nas?method=uinfo` → `uk`（Baidu uid, 如 `1234567890`，必需）。
2. `POST /wenchain/genflowpro/sse/msgexchange`，body `{"client_session_id":"<uuid>","device_type":401}`，
   `Accept: text/event-stream` → 长连接，首帧 `type: connected`。**必须先建连**。
3. `POST /wenchain/genflowpro/sendmsg`，body 为 envelope（同连接用途）：
```json
{"type":"message","sub_type":"chat_create","client_session_id":"<同 2 的 uuid>",
 "session_id":"","uk":"<uk>","cid":0,"channel":"kuku_pc_genflowpro_v1","device_type":401,
 "created_at":<ms>,"msg_type":1,"sync_type":0,"sync_id":"","v":"cd1755270722192e4536f605",
 "data":{"text":"...","rich_input_params":[{"type":"text","data":{"content":"..."}}],
   "fsid":[],"quotes":[],"skills":[],"experts":[],
   "model_name":"<model_list[].model_name>","model_display_name":"<display_name>",
   "think_mode":3,"project_id":""}}
```
   → 返回 `{status:{code:0},data:{session_id,reply_id,traceparent}}`。**首轮不要传 `reply_id`**（传了报 `invalid reply_id`）。
   `data.rich_input_params` 是**数组**（不是对象），否则 `bind params error`。
4. SSE 推送 `type: panel` / `sub_type: message_modify`，其 `data` 含**预签名** `data_url` / `data_url_with_thought`
   （`https://wenchain-genflow.bj.bcebos.com/...`，签名在 query，**GET 无需 cookie**）。
5. `GET data_url` → **GenFlow 事件数组**（本 Provider 的响应体）。实测事件序列：
   `USER_QUERY → REPLY_START → HINT_BLOCK×2 → MODEL_CALL_START → SESSION_TITLE → TEXT_BLOCK_START →
   TEXT_BLOCK_DELTA → TEXT_BLOCK_END → MODEL_CALL_END → REPLY_END → ACTUAL_POINT → TURN_DONE → SUG → FINISH`
6. 文本在 `data.delta`（实测 `TEXT_BLOCK_DELTA` 的 `data.delta == "OK"`）。`REPLY_END.data.finished_reason == "completed"`。

### 14.2 模型 (runtime-confirmed)

`GET /wenchain/genflowpro/model/list?channel=genfp` → `status.code=0`，`data.model_list` (13 项)。
**`model_name` 不是显示名**：

| id | model_name (用于请求) | display_name | vip_type | cost_ratio |
|---|---|---|---|---|
| 1 | `auto` | Auto | 0 | |
| 2 | `gateway-deepseek-v4.1-flash-volcengine` | DeepSeek-V4.1-Flash | 0 | 0.16x |
| 3 | `gateway-deepseek-v4-pro-tencent` | DeepSeek-V4-Pro | 0 | 0.52x |
| 4 | `gateway-deepseek-v4-flash-tencent` | DeepSeek-V4-Flash | 0 | 0.15x |
| 5 | `gateway-glm-5.3-flash` | GLM-5.3-Flash | 0 | 0.07x |
| 6 | `glm-5.3` | GLM-5.3 | 1 | 0.77x |
| 7 | `gateway-glm-5.2` | GLM-5.2 | 0 | 0.67x |
| 8 | `gateway-glm-5.1-kuaishou` | GLM-5.1 | 0 | 0.56x |
| 9 | `ernie-5.1` | 文心 5.1 | 0 | 0.53x |
| 10 | `ms-kimi-k3` | Kimi-K3 | 1 | 3.60x |
| 11 | `gateway-kimi-k2.7-code-tencent` | Kimi-K2.7-Code | 0 | 0.58x |
| 12 | `gateway-kimi-k2.6` | Kimi-K2.6 | 0 | 0.56x |
| 13 | `ali-minimax/minimax-m3` | MiniMax-M3 | 0 | 0.35x |

`think_list`: id 1..4 = 低/中/高/极高；`default_think_id=3`，`default_model_id=1`。

### 14.3 额度 (runtime-confirmed)

`GET /bizapi/gfpro/getgfvipremain` → `status.code=0`，`data.list` 为资产数组：
`[{assetType:1,assetName:"token",totalPoint:"499.44"},{assetType:2,assetName:"duration",totalPoint:"7200"},
  {assetType:3,assetName:"scheduled_task",totalPoint:"2"},{assetType:4,assetName:"realtime_t",totalPoint:"0"}]`，
另有 `isVip/isExpire/isTrial/vipType`。
流式事件里另有 `ACTUAL_POINT`（本次扣减）于事件数组中（`data_from: commercial`）。

### 14.4 未确认 / 阻塞

- `GET /api/genflowpro/workspace/getsessiondetail` 返回 `errno:-6`（业务未鉴权态）；`settings/profile` 同样 `errno:-6` → 说明**存在额外的业务登录态**（cookie 只有 BDUSS 时 `uk=0`）。推理本身可用，但会话/档案接口需要补该登录态（可能需要 `STOKEN`/`PTOKEN` 或 `stoken` 交换）。
- 上述 `errno:-6` 也解释了首次（无 `uk`）发送无回复。

## 15. 会话/上下文模型实测 (runtime-confirmed, 2026-09-30)

问题：上游大模型接口是否无状态？上下文是否在服务端？

结论：**上游是「服务端会话」模型，没有无状态大模型接口。** 因此反代必须自己承担会话映射（同 deepseek-web）。

### 实测证据

| # | 实验 | 结果 | 含义 |
|---|---|---|---|
| 1 | `chat_create` 请求体 | 只有 `{text, rich_input_params, model_name, think_mode, project_id, ...}`，**无 `messages[]`** | 客户端不传历史 |
| 2 | `he.RETRY_LAST` 请求体 | 只有 `{session_id, reply_id}`，**连 text 都不重发** | 上下文按 `(session_id, reply_id)` 存服务端 |
| 3 | 首轮 `session_id:""` | 200，服务端签发新的 `session_id`+`reply_id` | 会话由服务端创建 |
| 4 | 传 `reply_id=<客户端 UUID>` | `200002 invalid reply_id` | `reply_id` 必须是**服务端签发**的值 |
| 5 | 传 `reply_id=<随机 32hex>` | `200002 invalid reply_id` | 非服务端签发即拒 |
| 6 | 只传 `session_id`（省略 `reply_id`） | 200，返回**同一 `session_id`** + 新 `reply_id`；SSE 推新 `message_modify` | 省略 `reply_id` 即「在本会话新起一轮」，是正确续聊姿势 |
| 7 | 新会话问「刚才的暗号」 | 回答「我这边没有收到过要记住的暗号」 | 上下文**确实按会话隔离**（不是全局共享） |
| 8 | 同会话 T1 自述 | 「在本轮对话里我会记得它」 | 模型自证存在会话级上下文 |

> 未定项：第 6 项（仅 `session_id` 续聊）中 T2 的**回复正文在探测窗口内未抓到**（SSE 已推 `message_modify`，但未再取到 `REPLY_END` 事件包）。故「同会话多轮是否 100% 保留上文」标为 **strongly-inferred（强证据但未最终确认）**，实现时需再验一次。

### 对反代的结论

- **不能**把下游 `messages[]` 当成上游的输入契约（上游不吃数组）。
- 两条路线：
  - **A. 会话映射（推荐）**：下游会话 → 上游 `(session_id, reply_id, uk, client_session_id)` 映射（带 TTL）。
    首轮省略 `reply_id`；续聊**只带最新一条** + 复用 `session_id`，省略 `reply_id` 起新轮。完整保留服务端记忆/工具。
  - **B. 无状态模拟**：每轮新建会话，把 `messages[]` 拍平成一个 `text`（`kukuAIMessagesToText` 现有思路）。
    简单但丢服务端记忆/工具/文件上下文，且每轮建会话有成本。
- 工具调用（`REQUIRE_EXTERNAL_EXECUTION`）依赖会话状态 → 若要做工具，必须选 A。
