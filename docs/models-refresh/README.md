# 模型目录自动更新（models.json 远程刷新）

本目录记录 CLIProxyAPI（本 fork）**自动从远程仓库更新模型目录**的机制。

相关代码：

| 关注点 | 位置 |
|---|---|
| 周期与常量 | `internal/registry/model_updater.go` |
| 更新器 / 拉取 / 发布 | `internal/registry/catalog_sources.go` |
| 源配置校验 | `internal/registry/catalog_config.go` |
| codex 目录源 | `internal/registry/codex_client_models_updater.go` |
| devin 目录源 | `internal/registry/devin_models_updater.go` |
| 启动接线 | `sdk/cliproxy/service_lifecycle.go:455` |
| 热重载接线 | `sdk/cliproxy/service_config.go:171` |
| 刷新回调消费 | `sdk/cliproxy/service_plugins.go:354` |
| 手动刷新 API | `internal/api/handlers/management/model_definitions.go` |

## 1. 结论

**会。** 启动时立即拉取一次，之后每 **3 小时** 拉取一次；配置热重载会重建更新器；另有手动端点可立即触发。

三个目录各自独立刷新：general（`models.json`）、codex-client、devin。

失败或校验不过时**保留上一份有效目录**，绝不擦除。

## 2. 三个目录与源

```go
// internal/registry/model_updater.go:17
const ModelsRefreshInterval = 3 * time.Hour
```

| 目录 | 更新器 | URL（按顺序尝试） | 周期 |
|---|---|---|---|
| **general** | `generalCatalogUpdater` | `https://raw.githubusercontent.com/rizxfrog/CLIProxyAPI-Plus/refs/heads/feature/dev/internal/registry/models/models.json` | 3h |
| codex-client | `codexCatalogUpdater` | `https://raw.githubusercontent.com/router-for-me/models/refs/heads/main/codex_client_models.json`<br>`https://models.router-for.me/codex_client_models.json` | 3h |
| devin | `devinCatalogUpdater` | `https://raw.githubusercontent.com/router-for-me/models/refs/heads/main/devin_models.json`<br>`https://models.router-for.me/devin_models.json` | 3h |

注意：

- **general 只有一个 URL**，没有镜像；codex/devin 各有 raw + CDN 两个。
- general 指向**本 fork 自己的仓库**（`rizxfrog/CLIProxyAPI-Plus`），因为该文件含 fork 专有 provider 分节。合并上游时**不得**回退成 `router-for-me/models`——上游那份只有官方 provider，会丢分节（`preserveLocalCatalogSections` 只能补救一部分，见 §6）。
- 必须是 **raw** URL。`github.com/.../blob/...` 返回 HTML，会在校验阶段被拒（§4）。

`modelsURLs` 的注释里已写明这一点（`internal/registry/model_updater.go:19-25`）。

## 3. 触发时机

| 时机 | 路径 |
|---|---|
| **启动立即一次** | `StartModelCatalogUpdaters(ctx, cfg.Models, cfg.Home.Enabled)` → `configureCatalogs` → 每个 `configure()` 起一个 goroutine，先 `refresh()` 一次 |
| **此后每 3 小时** | 同一 goroutine 的 `time.NewTicker(ModelsRefreshInterval)` |
| **配置热重载** | `UpdateModelCatalogSources(newCfg.Models, newCfg.Home.Enabled)` → 重建三个更新器 |
| **手动** | `POST /v8/management/models/refresh`（见 §8），三个目录**并行**刷新 |

`configure()`（`catalog_sources.go:46`）的几个关键分支：

```go
if u.cancel != nil && u.source == source && u.parentDone == ctx.Done() && u.ctx.Err() == nil {
    return                      // 源与生命周期都没变 → 幂等，不重建
}
u.generation++                  // 换源即换代
u.publishedSeq = 0
if source == "disabled" { return }          // 完全不拉
go func() {
    u.refresh(...)                          // 立即一次
    if source == embeddedCatalogSource { return }   // 内嵌 → 不再周期拉
    ticker := time.NewTicker(ModelsRefreshInterval)
    ...
}()
```

所以：

- `--local-model` → `source == "embed"` → **只加载一次内嵌目录，永不联网**。
- `source == "disabled"`（home 模式，见 §7）→ **完全不启动拉取**。

## 4. 拉取与发布

### 4.1 取值

`readCatalogSource`（`catalog_sources.go:242`）：

- 路径为**绝对路径** → `os.Open`，直接从磁盘读。
- 否则 → HTTP GET，**必须 200**，否则 `catalog HTTP status %d`。
- 读取上限 `maxCodexClientModelsSize`（**8 MiB**）；超限报 `catalog exceeds size limit`。

`catalogFetcherWithSource`（`:282`）：按 URL 顺序逐个试，**第一个"能读且校验通过"的胜出**；全都失败则报 `no valid model catalog source`。

### 4.2 发布

`publishCatalogBytes`（`:324`）：

```
unmarshal → validateModelsCatalog → preserveLocalCatalogSections → detectChangedProviders → 写入 store
```

校验失败 → 拒绝整份（`model catalog rejected; keeping last valid catalog`）。这是防"拉到 HTML 页"的关口。

### 4.3 并发纪律

`catalogUpdater` 用 `generation` + `requestSeq` + `publishedSeq` 保证：

- 换代（换源）后，**旧代的在途请求作废**；
- 更旧的 seq 不得覆盖更新的发布；
- 取消后返回的读操作也不会写入。

注释原话：*"serializes source changes with publication, not network I/O"* —— 即**不持锁做网络 I/O**。

## 5. 失败语义：绝不擦除

| 情况 | 行为 |
|---|---|
| 网络/HTTP 失败 | `log.Warn("model catalog refresh failed; keeping last valid catalog")`，保留上一版 |
| 校验失败 | `log.Warn("model catalog rejected; keeping last valid catalog")`，保留上一版 |
| 全部 URL 失败 | 保留上一版 |
| 超 8 MiB | 拒绝 |

`internal/registry/codex_client_models_test.go:179` 有对应断言：*"failed remote refresh replaced last valid catalog"*。

## 6. Fork 专有分节保护

`preserveLocalCatalogSections`（`catalog_sources.go:351`）：远程若**缺失**某分节，用本地（当前已发布，初始为内嵌）的补上；**远程若有则以远程为准**（上游日后可接管）。

```go
if len(parsed.Meta) == 0 { parsed.Meta = old.Meta }
// 逐分节：len(*incoming) == 0 && len(local) > 0 → *incoming = local
```

保护清单（15 项）：

```
meta, codebuddy-cn, codebuddy-ai, deepseek-web, trae, cline,
qoder-cn, qoder-ai, minimax, floatboat, lobsterai-server,
xiaohuanxiong, codearts, loomy, dumate
```

⚠️ 该机制**只能补救"远程缺失"**。如果**本地内嵌**那份就缺某个分节（见 §10），而远程又恰好没有该分节，结果仍然是空——保护逻辑没有可补的源。

## 7. 配置与覆盖

配置结构（`internal/registry/catalog_config.go:11`，`config.ModelCatalogs` 是它的别名）：

```yaml
models:
  catalog: ""          # general
  codex-catalog: ""    # codex-client
  devin-catalog: ""    # devin
```

`Validate`（`:18`）拒绝**相对路径**与**非 http(s) 协议**：

```
models.<name> must be an http(s) URL or an absolute local path
```

### 取值优先级

general 目录的源按以下顺序确定（`effectiveCatalogSources` + `defaultGeneralCatalogSource`）：

| 顺序 | 来源 | 说明 |
|---|---|---|
| 1 | `models.catalog` | 显式配置；可为绝对路径或 http(s) URL |
| 2 | `--local-model` | → `"embed"`：用内嵌，**不联网、无周期** |
| 3 | `MODELS_FILE` | 环境变量，绝对路径 |
| 4 | `MODELS_URL` | 环境变量，**优先级低于 `MODELS_FILE`** |
| 5 | 空（`""`） | 回落到内置 URL 列表（即 §2 的 general URL） |

注意第 5 条：**源为空字符串 = 用内置 URL 列表**，不是"不更新"。只有 `"disabled"` 与 `"embed"` 才不联网。

### home 模式

`home.enabled = true` 时（`effectiveCatalogSources`）：

```go
if home { general = "disabled"; devin = "disabled" }
// 注意：codex 未被置为 disabled
```

同时 provider 级探针也会因 `antigravityHomeEnabled()` 而跳过。

## 8. 手动刷新 API

| 方法 | 路径 |
|---|---|
| POST | `/v8/management/models/refresh` |
| POST | `/v8/management/routing/model-definitions/refresh` |

两者都指向 `RefreshModelDefinitions`，内部调 `RefreshModelCatalogsWithPolicy`，**三个目录并行**（`sync.WaitGroup`），返回：

```json
{
  "ok": true, "status": "ok",
  "catalog": "<实际源>", "codex_catalog": "...", "devin_catalog": "...",
  "sources": {...},
  "catalogs": { "catalog": {...}, "codex-catalog": {...}, "devin-catalog": {...} },
  "changed_providers": ["..."]
}
```

失败返回 **502** 并带同样字段（`internal/api/handlers/management/model_definitions.go`）。某个目录 `Skipped: true` 表示源是 `"disabled"` 或 `"embed"`。

> v0 只有 `GET /v0/management/model-definitions/:channel`（读静态定义），**没有**刷新端点。v0 已废弃，新功能只落在 v8。

## 9. 与 provider 级"实时目录探针"的区别

这是**两层**机制，不要混淆：

| | 目录级（本文档） | provider 级 |
|---|---|---|
| 对象 | `models.json` 整份静态目录 | 单个账号的真实可用模型 |
| 源 | fork 仓库的 raw JSON | 各厂商接口（如 Loomy `GET /models`） |
| 驱动 | `internal/registry` | `sdk/cliproxy/*_models.go` |
| 周期常量 | `ModelsRefreshInterval` | 同样引用 `registry.ModelsRefreshInterval` 作为信任期 |

provider 级探针**复用**同一个常量作为目录信任期，但其扫描间隔、退避与门控是各自实现的（如 `loomyModelRefreshScanInterval = time.Minute`、`*ModelRetryInterval = 10 * time.Minute`）。

引用 `registry.ModelsRefreshInterval` 的 provider 探针：

```
sdk/cliproxy/antigravity_models.go:26        antigravityCapabilityCacheTTL
sdk/cliproxy/codearts_models.go:25
sdk/cliproxy/codebuddy_models.go:24
sdk/cliproxy/floatboat_models.go:25
sdk/cliproxy/lobsterai_server_models.go:25
sdk/cliproxy/loomy_models.go:24
sdk/cliproxy/xiaohuanxiong_models.go:25
```

## 10. 实测验证（2026-11-10）

```
GET https://raw.githubusercontent.com/rizxfrog/CLIProxyAPI-Plus/refs/heads/feature/dev/internal/registry/models/models.json
→ HTTP 200, text/plain, 393781 bytes, 27 个分节
→ 与 HEAD 逐分节条目数完全一致 ✅
```

fork 专有分节均在远程存在：

| 分节 | 条目 |
|---|---|
| cline | 458 |
| codebuddy-ai | 30 |
| floatboat | 26 |
| codebuddy-cn | 20 |
| trae | 17 |
| qoder-cn / qoder-ai | 14 / 14 |
| codearts | 9 |
| deepseek-web | 8 |
| xiaohuanxiong | 7 |
| loomy | 6 |
| dumate | 4 |
| minimax | 3 |
| **lobsterai-server** | **0** |

`lobsterai-server` 静态分节为空是**预期**：该 provider 的真实目录由后端按账号返回，静态列表只是兜底，实际目录完全来自 `asyncProbeLobsteraiServerCatalogue` 的探针。

### ⚠️ 关于工作区状态的提醒

`internal/registry/models/models.json` 曾被发现处于**被修改**状态（` M`，非本文档作者改动）：

| | 分节数 |
|---|---|
| HEAD（会 `go:embed` 进二进制的版本） | **27** |
| 被修改的工作区版本 | **22** |

被删除的 5 个分节：`codebuddy-cn`、`codebuddy-ai`、`xiaohuanxiong`、`codearts`、`loomy`。

影响面：

- **远程拉取不受影响**：远程含这 5 节，会照常发布。
- **`--local-model` 构建会缺这 5 个 provider 的模型**，因为 `go:embed` 内嵌的是工作区那份。
- **若该文件被提交**，内嵌兜底将永久缺这 5 节，且 `preserveLocalCatalogSections` 也补不回来（本地没有可补的源）。

排查时若发现 5 个 provider 的静态目录为空，先检查这个文件是否被改成 22 分节版本。

## 11. 排查清单

| 症状 | 检查点 |
|---|---|
| 目录没更新 | 日志里找 `model catalog refresh failed` / `model catalog rejected`；确认 `home.enabled` 是否把 general 置为 `disabled` |
| 拉到但被拒 | 源是不是 raw URL；文件是否合法 JSON 且能通过 `validateModelsCatalog`；是否超 8 MiB |
| 某个 provider 分节消失 | 是否命中 `preserveLocalCatalogSections` 清单；内嵌那份是否本身就缺（§10） |
| 完全没联网 | 是否用了 `--local-model`（→ `embed`）；`configure` 对 `embed` 直接 return，无 ticker |
| 手动刷新无效果 | `POST /v8/management/models/refresh` 返回里看 `catalogs.*.skipped` 是否为 true |
