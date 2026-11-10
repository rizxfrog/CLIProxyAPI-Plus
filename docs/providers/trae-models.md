# Trae (SOLO CN) — 模型列表接口逆向分析

来源：
- 安装包 `reversewk/trae/TRAE_CN-Setup-x64.exe`（Inno Setup，v3.4.17，chess 2026-02-12 构建）
- 内含 `app.asar`（JS 主进程 + 渲染层）与 `icube_server_rs`（Rust 内核）
- 真实账号实测（2026-11-10，`uid=3815748660367363`，`AIRegion=CN`）

## 1. 结论

**Trae 没有可靠的模型列表接口，不要做自动拉取。**

渲染层确实有一个 `getModelListByFunction` 调用，但它走的是**本地内核的
WebSocket RPC**，不是可直接调用的 HTTPS 接口。REST 侧的三个候选接口实测均不可用
（见 §3）。唯一能确定账号可用模型的方式是**逐个探测聊天通道**（§5），
代价是每个候选模型一次真实请求。

因此 Trae 继续使用静态目录，不做后台探针。

## 2. 客户端侧的模型获取路径（为什么不能直接复用）

渲染层 `ModelService` → `_aiNativeModelService.getModelListByFunction()`：

```js
async getModelListByFunction(functions, forceRefresh) {
  const client = await this._clientManagerService.getClient();
  const r = await client.request(this._clientManagerService.wrapRequestWithCredential({
    service: "model",
    method: x.METHODS.ModelListByFunction,   // service="model"
    data: { functions, force_refresh: forceRefresh }
  }));
  // r.data.list -> [{ function, models: [...] }]
}
```

RPC 契约（来自 Rust 侧 `icube_server_rs/modules/ai-agent`，符号名可读）：

```
GetModelListByFunctionRequest  { functions: string[], force_refresh: bool }
GetModelListByFunctionResponse { list: [{ function, models: [...] }] }
```

传输层是 **WebSocket**（`new WebSocket(url)` 连接到本地内核），
`wrapRequestWithCredential` 只是注入凭据后原样透传：

```js
wrapRequestWithCredential(e) {
  if (!this._credentialStore.getState().loggedIn) throw Error("Please login first");
  return { ...e }
}
```

> 复刻它需要跑起本地 Rust 内核并维持 WebSocket 会话，不适用于服务端代理。

## 3. REST 接口实测（全部不可用）

生产 host 只有一个可用：`https://trae-api-cn.mchost.guru`
（`api.trae.com.cn` / `api.trae.cn` 对 `/api/ide/*` 一律 404）。

逆向出的配置表（`app.asar` 内 `api` 对象）：

```json
"modelListAPI":           "api/ide/v1/model_list",
"fusionModelListAPI":     "api/ide/v1/get_model_list",
"fusionModelDetailAPI":   "api/ide/v1/get_detail_param",
"batch_get_detail_param": "api/ide/v1/batch_get_detail_param",
"chatAPI":                "api/ide/v1/chat",
"llmRawChatAPI":          "api/ide/v1/llm_raw_chat"
```

| 接口 | 实测结果 |
|---|---|
| `GET/POST /api/ide/v1/model_list` | ❌ **恒定 4001** `param is invalid`，附加 `model_configs: []` |
| `GET /api/ide/v1/get_model_list` | ❌ 404 |
| `GET/POST /api/ide/v1/get_detail_param` | ⚠️ 200，但 `config_info_list` **恒为空** |
| `GET /api/ide/v1/providers` | ✅ 200（但用途见 §4） |

### 3.1 `model_list` 的 4001 不是版本门控

尝试过的参数（GET 与 POST 各一组）：`{}`、`{model_names:[]}`、`{config_names:[]}`、
`{config_name:''}`、`{model_name:''}`、`{names:[]}`、`{model_list:[]}`、
`{functions:['solo_work_lite']}`、`{functions:[...], force_refresh:true}` —— 全部 4001。

尝试过的客户端版本（`X-Ide-Version` / `X-Ide-Version-Code`）：

| 版本头 | 结果 |
|---|---|
| `0.1.43` / `20260716` | 4001 |
| `0.1.52` / `20260901` | 4001 |
| `9.9.9` / `20990101` | 4001 |

版本不影响结果，所以这不是 `0db67577` 提交里说的“版本门控”效应。
该提交的观察（旧版本导致模型变少）是真的，但机制不在这个接口上。

### 3.2 `get_detail_param` 不是模型列表，且是死常量

`get_detail_param` 是 **BYOK 自定义模型的详情**接口：返回
`{"allow_tenant_user_add_model":true,"config_info_list":[],"metadata":{}}`。
账号没有自定义模型时 `config_info_list` 恒为空，与参数无关
（试过 `config_name` / `config_names` / `model_name` / `model_names` /
`keys` / `config_name_list`，GET 与 POST 都一样）。

> ⚠️ **仓库里的注释是错的**：
> `internal/runtime/executor/trae_executor.go` 写着
> `// TraeModelsPath is the get_detail_param endpoint path.`
> 该常量 `TraeModelsPath` 全仓库**从未被调用**，且它既不是列表接口。

## 4. `/api/ide/v1/providers` 的真实用途

返回 26 个**第三方 BYOK 供应商**与 134 个模型，例如：

```json
{"code":0,"providers":[
  {"id":"deepseek","base_url":"https://api.deepseek.com","billing_mode":"paygo",
   "model_detail":[{"display_name":"DeepSeek-V4-Pro","model_name":"deepseek-v4-pro"}]},
  {"id":"volcengine-plan","billing_mode":"quota",
   "base_url":"https://ark.cn-beijing.volces.com/api/coding/v3",
   "model_detail":[{"model_name":"glm-5.3"}, ...]}
]}
```

这是**自定义模型配置界面**的数据源（openrouter / zhipu / moonshot / siliconflow /
gitee / aws bedrock 等 BYOK 入口），**不是** Trae 官方给该账号提供的模型列表。
把它当模型目录会导致大量该账号实际不可用的条目。

## 5. 唯一可行方式：聊天通道探测法

`POST {host}/api/agent/v3/llm_utils_chat` 对可用模型返回 200 + SSE 流，
对不可用模型返回 `event:error` + `{"code":4001,"message":"... param is invalid"}`。

请求体（`content` 必须是**数组**，字符串会 4001 并暴露结构要求）：

```json
{
  "function": "solo_work_lite",
  "messages": [{"role": "user", "content": [{"type": "text", "text": "hi"}]}],
  "stream": true,
  "config_name": "<模型名>"
}
```

### 5.1 实测结果（该账号 2026-11-10）

**可用 14 个**：

```
glm-5.2              glm-5.3            glm-5.3-flash       glm-5-turbo
DeepSeek-V4-Pro      DeepSeek-V4-Flash  deepseek-v4.1-flash
kimi-k3              kimi-k2.7-code     kimi-k2.6
qwen3.8-max          qwen-3.7-plus      minimax-m3          Doubao-Seed-2.1-Pro
```

**明确不可用**（含大小写差异）：

```
GLM-5.2  GLM-5.1  GLM-5  glm-5.1
deepseek-v4-pro  deepseek-v4-flash
Kimi-K2.7-Code  qwen3.7-max  qwen3.7-plus  Qwen3.8-Max  MiniMax-M3
doubao-seed-2-1-pro-260628  solo_work_lite  auto
```

> ⚠️ **模型名大小写敏感**。`glm-5.2` ✅ 而 `GLM-5.2` ❌；
> `DeepSeek-V4-Pro` ✅ 而 `deepseek-v4-pro` ❌；`qwen3.8-max` ✅ 而 `Qwen3.8-Max` ❌。
> 这是实测差异，不是猜测。

### 5.2 与静态目录的差异

| | 数量 |
|---|---|
| `models.json` 的 `trae` 分节 | 17 |
| 该账号实测可用 | 14 |
| **静态目录缺失** | **`glm-5-turbo`** |
| 静态目录有、本次未测到 | `Doubao-Seed-Evolving`、`Doubao-Seed-2.1-Turbo`、`DeepSeek-V4-Flash-Official`、`DeepSeek-V4-Pro-Official` |

“未测到”不等于不可用：探测只覆盖了候选列表，后四个带 `-Official` / 变体后缀的
条目没有出现在探测集内。要确认需再补一轮探测。

## 6. 鉴权与请求头

```
Authorization: Cloud-IDE-JWT <access_token>
X-Cloudide-Token: <access_token>
X-Ide-Token:      <access_token>
X-Uid:            <uid>
X-App-Id:         6eefa01c-1036-4c7e-9ca5-d891f63bfcd8
X-App-Version:    default
X-Ide-Version:    0.1.52
X-Ide-Version-Code:   20260901
X-App-Version-Code:   20260901
User-Agent:       Trae/0.1.52
```

凭据有效性可用 `POST https://api.trae.com.cn/cloudide/api/v3/trae/GetUserInfo`
（body `{"uid":"<uid>"}`）验证，200 且 `Result.UserID` 匹配即有效。

## 7. 为什么不接入自动更新

| 方案 | 可行性 |
|---|---|
| 复刻渲染层 RPC | ❌ 需要本地 Rust 内核 + WebSocket 会话 |
| 调用 `model_list` | ❌ 恒定 4001，参数结构未知 |
| 调用 `get_detail_param` | ❌ 只返回 BYOK 自定义模型 |
| 调用 `providers` | ❌ 返回 BYOK 供应商目录，非账号可用模型 |
| 聊天通道探测 | ⚠️ 可行但每个候选模型一次真实请求（计费/限流风险） |
| 保持静态目录 | ✅ 现状 |

## 8. 建议的后续改动（未实施）

1. 把 `glm-5-turbo` 补进 `models.json` 的 `trae` 分节。
2. 修正 `internal/runtime/executor/trae_executor.go:58` 的错误注释，
   或直接删除从未使用的 `TraeModelsPath` 常量。
3. 若要补全静态目录，需对 §5.2 的“未测到”条目再跑一轮探测。

## 9. 复现步骤

```bash
# 1) 解出安装包数据块（Inno Setup 的 zlb 容器）
7z x TRAE_CN-Setup-x64.exe -oext
# ext/[0] 以 "zlb\x1a" 开头，其后自偏移 4 起是 LZMA1 流（props=0x5d, dict=8MiB）

# 2) 包装成标准 .lzma 后解压（得到 ~1.88GB）
#    头 = props(1) + dictSize(4, LE) + uncompressedSize(8 = 0xFFFFFFFFFFFFFFFF)
7z x stream.lzma -oout

# 3) 在 out/stream 中定位接口
grep -a -o 'api/ide/v1/[a-z_]*' out/stream | sort -u
grep -a -o 'GetModelListByFunction[A-Za-z]*' out/stream | sort -u
```
