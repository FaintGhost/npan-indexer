# Npan 索引同步运行手册

## 1. 前置条件

- Meilisearch 已启动：
  - 本地开发/部署通常使用 `docker compose up -d`
  - CI 回归使用 `docker compose -f docker-compose.ci.yml ...`
- 已配置云盘认证参数：
  - `NPA_TOKEN`
  - 或 `NPA_CLIENT_ID` / `NPA_CLIENT_SECRET` / `NPA_SUB_ID`
- 已配置索引参数：`MEILI_HOST`、`MEILI_API_KEY`、`MEILI_INDEX`
- 已确认同步状态库路径（默认 `NPA_STATE_DB_FILE=./data/state/sync-state.sqlite`）。
- 若计划从管理接口直接回退使用服务端凭据，确认：
  - `NPA_ALLOW_CONFIG_AUTH_FALLBACK=true`
  - 且服务端环境中存在有效 `NPA_TOKEN` 或完整 OAuth 三元组

## 1.1 状态持久化说明

当前运行时以 SQLite 作为同步状态主存储，统一保存：
- `progress`：全量/当前同步进度
- `sync_state`：增量游标（`lastSyncTime` 等）
- `checkpoint`：全量 crawl 断点

兼容策略：
- `NPA_PROGRESS_FILE` 与 `NPA_SYNC_STATE_FILE` 仍保留，用于首次读取时从 legacy JSON 惰性导入 SQLite。
- 导入后，运行时主读写路径仍是 `NPA_STATE_DB_FILE` 指向的 SQLite 文件。
- 本轮迁移不会自动删除旧 JSON；如需排障，可保留旧文件做人工对照。

## 2. 端口与入口

### 2.1 开发 / 部署 compose

- 应用：`http://localhost:1323`
- 指标：`http://localhost:9091/metrics`
- Meilisearch：`http://localhost:7700`

### 2.2 CI compose

- 应用：`http://localhost:11323`
- 指标：`http://localhost:19091/metrics`
- Meilisearch：`http://localhost:17700`

### 2.3 当前主入口

- 管理 RPC：`/npan.v1.AdminService/*`
- 健康检查：`GET /healthz`、`GET /readyz`
- 运维 CLI：`go run ./cmd/cli ...`

## 3. 首次全量同步

### 3.1 预检

建议先执行：

```bash
task verify:quick
cd web && bun run typecheck
```

如果要回归容器链路，再执行：

```bash
task verify:smoke
```

### 3.2 方式 A：CLI 触发（适合本地排障）

```bash
go run ./cmd/cli sync --mode full --root-folder-ids 0
```

常用补充参数：

- `--progress-output human`：默认，适合人工观察
- `--progress-output json`：适合日志采集
- `--root-workers <n>`：覆盖根目录并发度
- `--checkpoint-template <path>`：覆盖 checkpoint 文件位置
- `--sync-state-file <path>`：覆盖增量游标文件位置

说明：

- `sync` 默认从环境变量读取认证参数与 Meilisearch 配置。
- 首次全量一般不需要额外传 `--incremental-query-words` 或 `--window-overlap-ms`。
- 默认 `root_folder_ids` 来自 `NPA_ROOT_FOLDER_IDS`，若未配置则回落到 `0`。

### 3.3 方式 B：管理 Connect API 触发（推荐现网入口）

启动全量同步：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{"mode":"SYNC_MODE_FULL"}' \
  http://localhost:1323/npan.v1.AdminService/StartSync
```

查询进度：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{}' \
  http://localhost:1323/npan.v1.AdminService/GetSyncProgress
```

取消同步：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{}' \
  http://localhost:1323/npan.v1.AdminService/CancelSync
```

补充说明：

- `AdminService` 路由要求 API Key。
- 当前服务端还会为 `AdminService` 挂载 `ConfigFallbackAuth()` 与限流中间件。
- 同步、InspectRoots 等管理操作统一走 `POST /npan.v1.AdminService/*`，不要再使用历史 `/api/v1/*` 路径。

### 3.4 完成后建议核对

- `POST /npan.v1.AdminService/GetIndexStats`
- `POST /npan.v1.AdminService/GetSyncProgress`
- `http://localhost:9091/metrics` 或 CI 对应 `19091`

重点关注：

- 遍历目录数量
- 索引文档数量
- 失败请求数量
- 是否仍有运行中的同步任务

## 4. 增量同步调度

- 建议每 5~15 分钟执行一次增量同步。
- 增量游标运行时主存储位于 `NPA_STATE_DB_FILE` 指向的 SQLite `sync_state` 命名空间。
- `NPA_SYNC_STATE_FILE` 仍保留为 legacy JSON 导入来源，不再是默认主读写路径。
- 增量查询词默认来自 `NPA_INCREMENTAL_QUERY_WORDS`，默认值是 `* OR *`。
- 回看窗口默认来自 `NPA_SYNC_WINDOW_OVERLAP_MS`，默认值是 `2000` 毫秒。
- 同步成功后会写入新的 `lastSyncTime`；失败时保留旧游标。

CLI 示例：

```bash
go run ./cmd/cli sync --mode incremental --incremental-query-words "* OR *" --window-overlap-ms 2000
```

## 5. 根目录巡检与索引统计

拉取目录详情：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{"folderIds":[0]}' \
  http://localhost:1323/npan.v1.AdminService/InspectRoots
```

查看索引统计：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{}' \
  http://localhost:1323/npan.v1.AdminService/GetIndexStats
```

可调参数：

- `NPA_INSPECT_ROOTS_MAX_CONCURRENCY`
- `NPA_INSPECT_ROOTS_PER_FOLDER_TIMEOUT`

## 6. 检索与下载

本地索引搜索：

```bash
go run ./cmd/cli search-local --query "关键词"
```

远程平台搜索：

```bash
go run ./cmd/cli search-remote --query "关键词"
```

获取下载链接：

```bash
go run ./cmd/cli download-url --file-id 123
```

说明：

- 下载链接是临时 URL，不应持久化到索引。
- 浏览器公开搜索下的下载链路仍经 `AppService.AppDownloadURL` 受控下发。

## 7. 告警建议

- 429 比例 > 5%（5 分钟窗口）告警。
- 同步任务连续失败 3 次告警。
- checkpoint 或增量游标长时间不推进告警。
- `InspectRoots` 长时间超时或持续部分失败告警。

## 8. 故障恢复

1. 检查 Meilisearch 健康：`curl "$MEILI_HOST/health"`
2. 检查 `GET /healthz` 与 `GET /readyz`
3. 检查云盘 token 是否过期，或 OAuth 三元组是否仍可换取 token。
4. 检查 SQLite 状态库是否存在且可更新：
  - 路径默认是 `./data/state/sync-state.sqlite`
  - 也可通过 `NPA_STATE_DB_FILE` 覆盖
5. 若需要人工确认当前进度，执行：
  - `go run ./cmd/cli sync-progress --state-db-file ./data/state/sync-state.sqlite`
6. 若增量游标异常：优先检查 SQLite 中的 `sync_state` 是否符合预期；必要时可用保留的 legacy JSON 做对照。
7. 若全量 checkpoint 异常：先核对 SQLite 中的 checkpoint 是否更新；必要时再对照 `NPA_CHECKPOINT_FILE` 对应的 legacy 文件。
8. 若索引污染严重：清空目标索引后重跑全量。
9. 若怀疑是迁移问题：
  - 确认 `NPA_PROGRESS_FILE` / `NPA_SYNC_STATE_FILE` 仍指向原 JSON 文件。
  - 保留旧 JSON，不要先删除；程序会在 SQLite 缺失对应记录时惰性导入。
10. 若 smoke / E2E 超时：确认测试是否仍等待旧 REST `/api/v1/*` 路径，当前应校验 Connect `POST /npan.v1.*`
