# TRAE SOLO CN — live model list

Provider key: `trae`
状态: **已在真实账号上 live 验证**（凭据 `trae-3273661518457721.json`，
`type=trae`/`auth_kind=oauth`，2026-11-27 实测）。模型列表 37 个可路由模型，
`v1/chat/completions` 走 `config_name` 正常返回。

## 1. 证据来源

安装包：`/mnt/usb/a/trae/TRAE_CN-Setup-x64.exe`
Inno Setup 6.4.0.1，`ProductName: TRAE CN`，`FileVersion: 3.4.17`（2025-02-12）。
解包命令：`innoextract -e TRAE_CN-Setup-x64.exe`。

| 证据 | 位置 |
| --- | --- |
| 网关 host | `resources/app/product.json` → `bootConfig.agent.trae.normal = https://trae-api-cn.mchost.guru` |
| App ID | `product.json` → `bootConfig.ug.appId = 6eefa01c-1036-4c7e-9ca5-d891f63bfcd8` |
| 路由表 | `resources/app/modules/ai-agent/ai_agent.dll` 内联字符串：`…api/ide/v1/model_listapi/ide/v1/get_model_listapi/ide/v1/get_detail_paramapi/ide/v1/batch_get_detail_param…` |
| 桌面请求头 | 同 dll：`X-APP-ID/X-App-Version/X-Ide-Version-Code/X-App-Version-Code/X-Plugin-Channel/X-Device-Brand/X-Device-Cpu/X-Device-Id/X-Machine-Id/X-Web-Id/X-Os-Version/X-Device-Type/x-ide-version/x-ide-version-type/request-traffic-type` + `x-ide-token` |
| 桌面 UI 调用 | `resources/app/node_modules/@byted-icube/solo-lite/dist/3160.4b0fc74a.mjs`：`request({service:"model",method:"model_list_by_function",data:{functions,force_refresh,show_custom_model:true}})` |
| 响应字段 | `ai_agent.dll` 类型表：`ModelConfigInfo{usage,display_config,display_contact_config,model_detail_list,custom_models,config_switch,is_invisible_to_user}`（live 已验证） |
| 硬编码回退 | `node_modules/@byted-icube/ai-modules-chat/dist/index.mjs` 内 `CN_DEFAULT_MODELS*` / `US_DEFAULT_MODELS*`，条目形如 `{name,display_name,multimodal,is_default,config_source,model_type,custom_config}` |
| 真实版本 | `product.json` → `version=1.107.1`, `traeVersionCode=20250325` |

## 2. 端点契约（全部 runtime-confirmed）

真实凭据 `data/auth_files/trae-3273661518457721.json`（`type=trae`, `auth_kind=oauth`，
未过期）在真实网关 `https://trae-api-cn.mchost.guru` 上实测。

### 2.1 模型来源：get_detail_param（唯一正确来源）

```
POST /api/ide/v1/get_detail_param
{"function":"solo_work_lite","need_prompt":false}
→ {"code":0,"allow_tenant_user_add_model":true,"metadata":{},
   "config_info_list":[{"config_name":"glm-5.2","config_switch":true,
     "is_invisible_to_user":false,"usage":"chat_completion",
     "config_source":1,
     "context_window_tokens":{"dev":200000},
     "display_config":{"display_name":"GLM-5.2","multimodal":false,
       "model_capability":"reasoning_model","max_mode":false,…},
     "reasoning_effort_config":{…},"model_detail_list":[…],…}]}
```

该账号返回 **54 行**，`is_invisible_to_user`/`config_switch` 过滤后 **37 个可路由模型**。
路由键是 `config_name`；`display_name` 与 `multimodal` 在 `display_config` 里；
窗口在 `context_window_tokens.{dev,max}`（该账号只发 `dev`）。

### 2.2 model_list：存在但**不适合**本 provider

```
GET /api/ide/v1/model_list?type=chat     → 200，5 行
GET /api/ide/v1/model_list?type=builder  → 200，0 行
不加 type / 其它任何 type                 → 400 {"code":4001,"message":"…param is invalid…"}
```

`type` 只有 `chat` / `builder` 两个合法值（其余 20+ 候选全部 400）。

致命点：`type=chat` 返回的是 **IDE 编辑器模型**（`name=seed_m8`、`deepseek-V3`…），
把它们当 `config_name` 发到 solo 通道会被拒：

```
config_name=glm-5.2   → 200 + 正常 SSE
config_name=seed_m8   → event:error {"code":4001,"message":"…param is invalid…"}
config_name=deepseek-V3 → event:error {"code":4001,…}
```

**结论**：`model_list` 不是本 provider 的模型来源；原代码里零引用的
`TraeModelsPath = "/api/ide/v1/get_detail_param"` 才是对的，一度被我误改成
`model_list`，现已改回并接线。

### 2.3 端点对照

| 路径 | 结果 |
| --- | --- |
| `POST /api/ide/v1/get_detail_param` | 200 + `config_info_list`（正确来源） |
| `GET /api/ide/v1/model_list?type=chat` | 200，但为 IDE 编辑器模型，solo 通道不可用 |
| `GET /api/ide/v1/model_list`（无 type） | 400 `4001 param is invalid` |
| `POST /api/ide/v1/batch_get_detail_param` | 200，但主体是 `client_config`/skills，模型在 `function_configs` |
| `GET /api/ide/v1/get_model_list` | 404（旧版路径，勿用） |
| `POST /api/agent/v3/llm_utils_chat` | 200 SSE（对话通道，`config_name` 生效） |

### 2.4 版本头不是 4001 的原因

安装包真实版本是 `1.107.1` / `traeVersionCode=20250325`，而代码里发的是
`0.1.52` / `20260901`。实测把 `X-Ide-Version(-Code)` 换成真实值、换 `User-Agent`、
补 `X-Device-Id` 等组合，`model_list` **仍然全部 4001**；而对话通道用原头部
就是 200。所以 4001 是缺 `type` 参数，不是版本门控。（`TraeIdeVersion` 注释里
"旧版本导致 4001" 的说法对 `model_list` 不成立。）

## 3. 实现

| 位置 | 内容 |
| --- | --- |
| `internal/auth/trae/agent.go` | `ModelClient.FetchCatalogue`、`ParseDetailParam`、`ApplyAgentHeaders` / `ApplyDesktopHeaders`、`AppID` / `IdeVersionCode` / `DetailParamPath` / `FunctionSoloWorkLite` 常量 |
| `internal/auth/trae/schema.go` | `DescribeModelList` / `SchemaReport`：值脱敏的 schema 自检 |
| `sdk/cliproxy/trae_models.go` | `asyncProbeTraeCatalogue`、`executeTraeCatalogueProbe`、`traeModelsForCatalogue`、`refreshTraeCatalogues`、`runTraeModelRefresh` |
| `sdk/cliproxy/service_models.go` | 注册分支（`len(models)>0` 与空表两条路径）触发 probe |
| `sdk/cliproxy/service_lifecycle.go` | `go s.runTraeModelRefresh(ctx)` |
| `internal/runtime/executor/trae_executor.go` | 头部改由 `internal/auth/trae` 统一生成，与 probe 共用一份身份定义 |

行为：

- 注册凭据后立即后台拉取一次；成功则用网关列表**替换**该凭据的注册模型。
- 失败或空列表：保留当前模型（静态回退），10 分钟后重试。
- 成功后 3 小时（`registry.ModelsRefreshInterval`）过期，每分钟扫描一次到期账号。
- 拉取结果仍走完整管线：`excluded_models` → alias → plugin → settings → prefix，
  与静态注册一致，因此 `prefix: "trae"` 仍然生效。
- 模型 `ID` 取 `config_name`（对话端点所需的路由键），
  `display_config.display_name` / `context_window_tokens.{max,dev}` /
  `display_config.multimodal` 透传；`config_switch=false` 与
  `is_invisible_to_user=true` 的行被丢弃。
- 探测上下文与 `antigravityContext` 生命周期绑定，不走请求上下文，
  不会因为注册调用返回而取消。并发受 `modelRegistrationMaxWorkersPerCategory` 限制。

## 4. 自检与剩余限制

### 4.1 已由真实响应确认的字段

`get_detail_param` 的 `config_info_list` 行实测字段：

```
config_name            string   ← 路由键（chat 的 config_name）
config_switch          bool     ← false 即禁用
is_invisible_to_user   bool     ← true 为内部模型，桌面端不展示
usage                  string   "chat_completion"
config_source          number
context_window_tokens  object{dev,max}   ← 实测多为只有 dev
display_config         object{display_name,multimodal,model_capability,max_mode,…}
reasoning_effort_config object
model_detail_list      array
display_contact_config string JSON
extra_config           string JSON
```

`context_window_tokens` 是对象而非数字，客户端读法：

```js
e.context_window_tokens?.max ?? e.context_window_tokens?.dev
```

解析器取 `max` → `dev` → `default` → 平铺数字，多形状兼容。

### 4.2 自检：首次响应打印 schema 报告

第一次成功拉取会打一条 `info`（每凭据一次，之后不重复），只含**键名与类型**：

```
trae: catalogue schema for trae-3273661518457721.json (one-shot, values redacted):
  envelope keys: allow_tenant_user_add_model, config_info_list, metadata;
  config_info_list: 54 entries;
  row keys: config_name(string), config_switch(bool),
            context_window_tokens(object{dev,max}|object{dev}),
            display_config(object{can_use_super_model,…,multimodal,…}), …;
  parsed: 37, skipped: 17
```

- `ak`、`sk`、`token`、`access_token`、`encrypted_model_param`、`encrypted_prompt*`、
  `custom_config`、`extra_config`、`uid`、`machine_id`、`device_id`、`email` 等
  只显示 `<redacted>`，**永不输出值**。
- `parsed: 37, skipped: 17` 即过滤后的可路由数与被过滤数。
- 缺 `config_name` / 缺上下文窗口 / 空数组 / 非 JSON 各会追加一条 anomaly。

### 4.3 剩余限制

- 只服务 `solo_work_lite` 一个 function。账号里其它 function
  （`solo_agent`、`solo_coder`、`builder`…）的模型集不同，需要通过
  `get_detail_param` 换 `function` 才能列出；当前 provider 的对话通道只有
  `solo_work_lite`，所以没有暴露。
- 未实现 `batch_get_detail_param` 与 `model_list_by_function`：前者主体是
  `client_config`/skills（模型在 `function_configs`），后者是 `jsonrpsee` RPC。
- 上游会淘汰模型（实测该账号已无 `glm-5.2`，`TraeDefaultModel` 因此改为
  `glm-5.3`）。这正是 live fetch 的价值：静态表会一直列出已下线模型。
- 该账号对部分模型返回 `code=1005`（配额/权益），属于账号额度，不是模型列表问题。
