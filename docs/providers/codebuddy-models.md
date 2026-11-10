# CodeBuddy (国内 / 国际) — 模型目录动态拉取

来源：`reversewk/codebuddy/cli/package`（`tencent-ai-codebuddy-code-2.182.0.tgz`）
的对象 `dist/codebuddy.js` 中的 `CloudProductProvider` / `CloudProductManagerImpl`，
以及仓库提供的只读探查脚本 `docs/scripts/workbuddy_models_cn.py` /
`workbuddy_models_intl.py`。

## 1. 结论

CodeBuddy 两版**都能**动态拉取模型目录，机制与 CodeArts 一致（后台探针 + 定期刷新），
但接口形态完全不同：CodeArts 有专门的模型目录接口，CodeBuddy 则是
**从一份「云产品配置」里取 `data.models` 字段**。

| 版本 | Provider 常量 | Origin | 生产 host |
|---|---|---|---|
| 国内 | `codebuddy-cn` | `https://copilot.tencent.com` | 同左（备用 `www.workbuddy.cn`、`wb.tencentbuddy.com`）|
| 国际 | `codebuddy-ai` | `https://www.codebuddy.ai` | 同左（预发 `staging-codebuddy.tencent.com`）|

## 2. 接口

两个版本**完全同构**，只有 host 不同。

| 用途 | 方法 | 路径 | 说明 |
|---|---|---|---|
| 模型目录 | GET | `{origin}/v3/config` | `data` 是整份云产品配置；模型只是其中一个字段 |
| 凭证有效性 | GET | `{origin}/v2/plugin/account` | **必须**先调用 |
| 企业级目录 | GET | `{origin}/console/enterprises/{id}/config/models` | 个人账号 403，不实现 |

### 2.1 鉴权与请求头

```
Authorization: Bearer <access_token>     # Keycloak JWT
Accept: application/json
X-Domain: <origin host>                  # copilot.tencent.com / www.codebuddy.ai
User-Agent: CLI/<ver> CodeBuddy/<ver>    # 决定返回哪套目录，见 §4
```

`repos` 查询参数可省略（实测带不带、带真实仓库地址，返回内容一致）。CLI 侧仍会带：
`params: { repos: <buildRepoParams(gitRepos)> }`。

### 2.2 响应形状

```json
{
  "code": 0,
  "data": {
    "models": [
      {"id":"glm-5.2","name":"GLM-5.2","maxInputTokens":1000000,"maxOutputTokens":50000,
       "supportsImages":true,"supportsToolCall":true,
       "reasoning":{"supportedEfforts":["low","medium","high"],"canDisableThinking":true},
       "tags":["chat"]}
    ],
    "agents": [{"name":"cli","models":["glm-5.2"]}],
    "productFeatures": {"CommandsSmartMerge": true}
  }
}
```

字段映射（`internal/auth/codebuddycn/catalog.go`）：

| 上游 | 注册表 |
|---|---|
| `id` | `ID` |
| `name` | `DisplayName`（缺省用 `id`）|
| `maxInputTokens` | `ContextLength` |
| `maxOutputTokens` | `MaxCompletionTokens` |
| `supportsImages` | `SupportedInputModalities` = `TEXT,IMAGE` |
| `reasoning.supportedEfforts` | `Thinking.Levels` |
| `reasoning.canDisableThinking` | `Thinking.ZeroAllowed` |

## 3. 两个必须复刻的陷阱

### 3.1 token 失效返回 **HTTP 200 + `models: null`**

这是最危险的一条：凭证失效时 `/v3/config` **不回 401**，而是 200 加 `data.models == null`，
与**匿名请求**的响应完全同构。若把它当作"目录为空"处理，会**静默撤销账号上所有模型**。

所以实现里先查 `/v2/plugin/account` 判定凭证，再取目录；`models == null` 一律报错：

```go
if !present || rawModels == nil {
    return nil, fmt.Errorf("codebuddy: catalogue is null for a valid account")
}
```

对应脚本里同样的判断（`fetch_config`）：
> `Catalog is null: the credential is missing, expired or revoked`

### 3.2 User-Agent 决定返回哪一套目录

缺失或格式不对直接 **HTTP 400 `{"code":12403,"msg":"check ua, get coding copilot version error"}`**。

| `--product` | User-Agent | 国内 | 国际 |
|---|---|---|---|
| `cli` | `CLI/<v> CodeBuddy/<v>` | 34 条 | 28 条 |
| `ide` | `WorkBuddy/<v>` | 39 条（含 IDE 专属）| **23 条**（反而更少）|

> ⚠️ 国际站 `ide` UA 返回的不是 IDE 目录而是更小的 CLI 目录，与国内站行为相反。
> 因此国际版固定用 `cli`。

版本号对网关不透明（任意 `x.y.z` 都能解析），当前用 `2.63.2` / `5.3.14`。

## 4. `data.models` ≠ CLI 实际提供的模型

`data.models` 比 `agents[name=="cli"].models` **更宽**。默认 `scope=cli` 的理由：

- 多出的行里有 **CLI 不对外展示**的（`glm-5.0-turbo`、`kimi-k2.5`，网关仍受理 code 11133）
- 有**网关目录根本不认**的（`minimax-m2.5`、`glm-5.0`、`glm-4.7/4.6/4.6v`、`kimi-k2-thinking`、
  `hy4-preview-x` 回 code 11102 `service info not found`）
- 有**图像/视频生成行**（code 11103）

`agents[cli].models` 是唯一实测全部可调的集合。

**当前实现取的是 `data.models` 全量**（与静态目录口径一致），并额外暴露
`Catalog.CLIModelIDs` 供后续按 CLI 口径收窄。若要严格对齐客户端，应改为按
`agents[cli].models` 过滤。

生成模型的判据（客户端与脚本一致）：`tags` 含
`text-to-image` / `image-to-image` / `text-to-video` / `image-to-video`。

## 5. 运行时接线

```
internal/auth/codebuddycn/catalog.go   FetchCatalog / UserAgentFor / 类型
sdk/cliproxy/codebuddy_models.go       后台探针：拉取 → 映射 → 发布 → 退避刷新
```

与 CodeArts 相同的时间参数：

| 参数 | 值 |
|---|---|
| 扫描间隔 | 1 分钟 |
| 目录信任时长 | `registry.ModelsRefreshInterval`（3 小时）|
| 失败退避 | 10 分钟 |

行为约定：

- **静态目录先发布**（`registry.GetCodeBuddyCNModels()` / `GetCodeBuddyAIModels()`），
  探针随后覆盖；失败或为空则保留兜底。
- **显式配置 `models:` 时不参与探针**（`codeBuddyConfigPinsModels`）。
- 两版共用一个探针，按 `auth.Provider` 选择 `NewClientWithProxyURL`（CN）或
  `NewAIClientWithProxyURL`（国际），区别只在 host。

## 6. 实测验证

假上游 + 真实 server 二进制：

```
GET /v1/models → 200, count 2
live-cb-model     ← /v3/config 的 data.models
live-cb-second
```

只返回这两个，**静态目录 30 条一条不剩**，证明覆盖发布生效。

单测：`internal/auth/codebuddycn/catalog_test.go`（6 条，含 UA 产品线、
null 目录拒绝、凭证先探测、12403 拒绝）、
`sdk/cliproxy/codebuddy_models_test.go`（4 条，含端到端发布）。

## 7. 已知未做

1. **企业级目录** `/console/enterprises/{id}/config/models` — 个人账号 403，企业账号才有意义。
2. **`agents[cli].models` 收窄** — 见 §4。
3. **`codebuddy-cn` 静态目录为空** — `models.json` 没有该键，
   `GetCodeBuddyCNModels()` 返回空表。当前依赖动态探针提供模型；
   若探针失败（凭据失效），该 provider 将没有任何模型可用。
   可用脚本补齐：
   ```bash
   uv run docs/scripts/workbuddy_models_cn.py models \
       --auth-dir <auth-dir> --prefix codebuddy-cn \
       --out-file internal/registry/models/models.json
   ```
