# npan

Npan 外部索引服务：把 Npan 云盘文件元数据同步到本地搜索索引，提供 Web 搜索页、管理后台、Connect-RPC API 和运维 CLI。

- 后端：Go + Echo
- 前端：React + Vite + Bun
- 搜索：Meilisearch / Typesense
- RPC：Buf + Connect-RPC（connect-go / connect-es）

## 1. 你会得到什么

- Web 搜索页：`/`
- 管理后台：`/admin/`
- Connect-RPC API：`/npan.v1.*`
- CLI：`go run ./cmd/cli ...`

## 2. 快速部署（推荐：Docker Compose）

### 2.1 前置条件

- Docker 24+
- Docker Compose v2

### 2.2 准备配置

```bash
cp .env.example .env
cp .env.meilisearch.example .env.meilisearch
cp .env.typesense.example .env.typesense
```

至少修改 `.env` 中这些字段：

```bash
NPA_ADMIN_API_KEY=your-admin-key-minimum-16-chars

# 二选一：
NPA_TOKEN=...
# 或
NPA_CLIENT_ID=...
NPA_CLIENT_SECRET=...
NPA_SUB_ID=...
# 可选：覆盖默认 SQLite 状态库位置
# NPA_STATE_DB_FILE=./data/state/sync-state.sqlite
```

说明：

- `NPA_ADMIN_API_KEY` 必填，且长度必须 `>= 16`。
- `NPA_SEARCH_BACKEND` 默认是 `meilisearch`；切换到 Typesense 时设为 `typesense`。
- 上游认证支持两种来源：
  - 直接提供 `NPA_TOKEN`
  - 提供 `NPA_CLIENT_ID` / `NPA_CLIENT_SECRET` / `NPA_SUB_ID`，由服务端换取 token
- 若只做最小功能联调，也可先提供 `NPA_TOKEN`，跳过 OAuth 三元组。
- `NPA_STATE_DB_FILE` 是同步状态的主存储，默认路径为 `./data/state/sync-state.sqlite`。
- `NPA_PROGRESS_FILE` 与 `NPA_SYNC_STATE_FILE` 仍保留为 legacy JSON 导入来源，用于首次惰性迁移与人工对照，不再是运行时主状态源。
- 若要启用浏览器直连公开搜索，可配置 `NPA_PUBLIC_INSTANTSEARCH_ENABLED=true`，并按后端分别提供对应的 public host / index / search-only key。
- Typesense 索引实现位于 `internal/search/typesense_index.go`，对应单测位于 `internal/search/typesense_index_test.go`。

### 2.3 启动

```bash
docker compose up -d --build
```

默认端口：

- 应用：`http://localhost:1323`
- 指标：`http://localhost:9091/metrics`
- Meilisearch：`http://localhost:7700`
- Typesense：`http://localhost:8108`

### 2.4 验证服务是否可用

```bash
curl -sS http://localhost:1323/healthz
curl -sS http://localhost:1323/readyz
```

期望响应：

```json
{"status":"ok"}
{"status":"ready"}
```

### 2.5 访问页面

- 搜索页：`http://localhost:1323/`
- 管理页：`http://localhost:1323/admin/`

### 2.6 SQLite 状态库说明

当前同步状态默认写入单一 SQLite 文件：

- 默认路径：`./data/state/sync-state.sqlite`
- 配置项：`NPA_STATE_DB_FILE`

状态库中统一保存：
- 全量同步进度（progress）
- 增量游标（sync state）
- crawl checkpoint

排障建议：
- 先看 `NPA_STATE_DB_FILE` 指向的 SQLite 文件是否存在、是否持续更新。
- 若需要对照迁移前状态，可同时保留 `NPA_PROGRESS_FILE` 与 `NPA_SYNC_STATE_FILE` 指向的 JSON 文件；它们仅作为 legacy 导入来源，不再是默认读路径。
- CLI `sync-progress` 默认也会优先读取 SQLite，可用 `--state-db-file` 显式指定状态库。

### 2.7 可选：启用浏览器公开搜索

当你希望搜索页优先走浏览器直连 public InstantSearch 链路时，按当前后端配置：

```bash
NPA_PUBLIC_INSTANTSEARCH_ENABLED=true

# Meilisearch backend
MEILI_PUBLIC_SEARCH_HOST=http://127.0.0.1:7700
MEILI_PUBLIC_SEARCH_INDEX=npan_items
MEILI_PUBLIC_SEARCH_API_KEY=<search-only-key>

# Typesense backend
TYPESENSE_PUBLIC_SEARCH_HOST=http://127.0.0.1:8108
TYPESENSE_PUBLIC_SEARCH_INDEX=npan_items
TYPESENSE_PUBLIC_SEARCH_API_KEY=<search-only-key>
```

注意：

- `MEILI_PUBLIC_SEARCH_API_KEY` 必须是 dedicated search-only key，不能复用私有 `MEILI_API_KEY`。
- `TYPESENSE_PUBLIC_SEARCH_API_KEY` 也必须是 dedicated search-only key，不能复用私有 `TYPESENSE_API_KEY`。
- 前端会先通过 `AppService.GetSearchConfig` 拉取公开搜索配置。
- 返回里会包含 `provider`，由前端在 Meilisearch / Typesense InstantSearch client 之间切换。
- 若公开搜索配置不完整，或 `NPA_PUBLIC_INSTANTSEARCH_ENABLED=false`，搜索页会自动回退到 legacy Connect `AppSearch` 链路。
- 以上 4 个变量由 `npan` 应用读取，并通过 `AppService.GetSearchConfig` 下发给浏览器；不要写到 `.env.meilisearch` / `meilisearch` 服务。
- `MEILI_PUBLIC_SEARCH_HOST` 必须是浏览器可访问的地址；生产环境不要填 Docker 内网地址，例如 `http://meilisearch:7700`。
- 服务端不会在 `MEILI_PUBLIC_SEARCH_API_KEY` 为空时回落复用私有 `MEILI_API_KEY`。
- `TYPESENSE_PUBLIC_SEARCH_HOST` 也必须是浏览器可访问的地址；生产环境不要填 Docker 内网地址，例如 `http://typesense:8108`。
- 服务端不会在 `TYPESENSE_PUBLIC_SEARCH_API_KEY` 为空时回落复用私有 `TYPESENSE_API_KEY`。

#### 获取 public search key（最小步骤）

1. 在 `.env.meilisearch` 中配置 `MEILI_MASTER_KEY` 并启动 Meilisearch。
2. 先用 `GET /keys` 检查现有 key；若拿不到真实 `key` 字段，就新建一个 search-only key。
3. 把返回 JSON 里的 `key` 写入应用 `.env` 的 `MEILI_PUBLIC_SEARCH_API_KEY`；不要填 `uid`。

```bash
curl -sS \
  -H 'Authorization: Bearer <MEILI_MASTER_KEY>' \
  http://localhost:7700/keys
```

```bash
curl -sS -X POST \
  -H 'Authorization: Bearer <MEILI_MASTER_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{"actions":["search"],"indexes":["npan_items"]}' \
  http://localhost:7700/keys
```

#### 最小验证

```bash
curl -sS -X POST \
  -H 'Content-Type: application/json' \
  -d '{}' \
  http://localhost:1323/npan.v1.AppService/GetSearchConfig
```

返回里应能看到 `provider`、`instantsearchEnabled`、`host`、`indexName`、`searchApiKey`。

```bash
curl -sS -X POST \
  -H 'Authorization: Bearer <MEILI_PUBLIC_SEARCH_API_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{"queries":[{"indexUid":"npan_items","q":"test","limit":1}]}' \
  http://localhost:7700/multi-search
```

如果这里返回 `invalid_api_key`，优先检查填的是不是 `key` 而不是 `uid`，以及该 key 是否具有目标索引的 `search` 权限。

### 2.8 本地 Docker + 真实凭据跑 admin live E2E

当你需要对 `/admin` 关键流程做真实数据验证时，可叠加 `docker-compose.e2e-live.yml`。

当前约束：

- `docker-compose.e2e-live.yml` 会注入 `E2E_LIVE=1`。
- 该 compose 文件当前要求 `NPA_CLIENT_ID`、`NPA_CLIENT_SECRET`、`NPA_SUB_ID` 必填。
- 你也可以额外提供 `NPA_TOKEN`；运行时会优先使用该 token，但 OAuth 三元组仍需要满足 compose 变量展开约束。

示例：

```bash
NPA_CLIENT_ID='<your-client-id>' \
NPA_CLIENT_SECRET='<your-client-secret>' \
NPA_SUB_ID='<your-sub-id>' \
NPA_SUB_TYPE='enterprise' \
NPA_TOKEN='' \
docker compose \
  -f docker-compose.ci.yml \
  -f docker-compose.e2e-live.yml \
  up --build -d --wait --wait-timeout 120

docker compose \
  -f docker-compose.ci.yml \
  -f docker-compose.e2e-live.yml \
  --profile e2e run --rm playwright \
  npx playwright test e2e/tests/admin.live.spec.ts

docker compose \
  -f docker-compose.ci.yml \
  -f docker-compose.e2e-live.yml \
  --profile e2e down --volumes
```

可选性能调优（用于 `/npan.v1.AdminService/InspectRoots`）：

- `NPA_INSPECT_ROOTS_MAX_CONCURRENCY`：目录详情并发度，默认 `6`
- `NPA_INSPECT_ROOTS_PER_FOLDER_TIMEOUT`：单目录请求超时，默认 `10s`

## 3. 同步与常用操作

### 3.1 方式 A：管理后台（推荐）

1. 打开 `http://localhost:1323/admin/`
2. 输入 `NPA_ADMIN_API_KEY`
3. 选择同步模式（全量 / 增量）
4. 启动同步并观察进度、统计与根目录详情

### 3.2 方式 B：Connect API

启动全量同步：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{"mode":"SYNC_MODE_FULL"}' \
  http://localhost:1323/npan.v1.AdminService/StartSync
```

查询同步进度：

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

其他常用管理 RPC：

- `POST /npan.v1.AdminService/InspectRoots`
- `POST /npan.v1.AdminService/GetIndexStats`
- `POST /npan.v1.AdminService/WatchSyncProgress`（stream）

### 3.3 方式 C：CLI

首次全量同步：

```bash
go run ./cmd/cli sync --mode full --root-folder-ids 0
```

增量同步：

```bash
go run ./cmd/cli sync --mode incremental --incremental-query-words "* OR *" --window-overlap-ms 2000
```

查看全量同步进度：

```bash
go run ./cmd/cli sync-progress
```

其他常用 CLI：

```bash
go run ./cmd/cli token
go run ./cmd/cli search-remote --query "关键词"
go run ./cmd/cli search-local --query "关键词"
go run ./cmd/cli download-url --file-id 123
```

说明：

- CLI 默认从环境变量读取凭据和当前搜索后端配置（Meilisearch / Typesense）。
- `sync` 支持 `--progress-output human|json`。
- `sync-progress` 默认优先读取 `NPA_STATE_DB_FILE` 指向的 SQLite 状态库，也可通过 `--state-db-file` 显式指定。
- `NPA_PROGRESS_FILE`、`NPA_SYNC_STATE_FILE`、`NPA_CHECKPOINT_FILE` 仍可用于 legacy 导入与排障对照，但不再是运行时主状态源。

## 4. 本地开发（不走 Docker 全栈）

### 4.1 启动依赖

```bash
docker compose up -d meilisearch
```

### 4.2 启动后端

```bash
go run ./cmd/server
```

### 4.3 启动前端（可选，独立 dev）

```bash
cd web
bun install
bun run dev
```

### 4.4 前端产物与嵌入说明

- 前端包管理器和脚本入口在 `web/package.json`，默认使用 `bun`。
- 生产模式下，后端通过 `web/embed.go` 中的 `//go:embed all:dist` 嵌入 `web/dist`。
- 如果本地直接运行 `go run ./cmd/server` 时缺少 `web/dist`，请先执行：

```bash
cd web && bun install && bun run build
```

## 5. 常用命令

长驻开发命令请使用上面的“本地开发”章节；下面统一列出仓库验证入口。

```bash
# 查看公开任务
task --list

# 快速验证（guard:rest + Go + 前端单测）
task verify:quick

# 单独执行某一类验证
task guard:rest
task test:go
task test:web

# 前端类型检查
cd web && bun run typecheck

# CI 冒烟测试
task verify:smoke

# 完整链路回归（smoke + Playwright E2E）
task verify:e2e

# 全量回归（verify:quick -> verify:e2e）
task verify:all

# 契约变更后的生成链路
buf lint
buf generate
```

## 6. API 与鉴权概览

### 6.1 Connect-RPC（运行时主路径）

当前运行时已是 Connect-only，主路径都在 `/npan.v1.*` 下。

- `/npan.v1.HealthService/*`
  - Connect 版健康检查 / 就绪检查
  - 不要求 API Key
- `/npan.v1.AppService/*`
  - 挂载 `EmbeddedAuth()`
  - 面向内嵌前端请求与公开搜索 bootstrap
  - 包含 `GetSearchConfig`、`AppSearch`、`AppDownloadURL`
- `/npan.v1.AuthService/*`
  - 挂载 `APIKeyAuth()`
  - 用于 `CreateToken`
- `/npan.v1.SearchService/*`
  - 挂载 `APIKeyAuth()`
  - 包含 `RemoteSearch`、`LocalSearch`、`DownloadURL`
- `/npan.v1.AdminService/*`
  - 挂载 `APIKeyAuth()` + `ConfigFallbackAuth()` + `RateLimitMiddleware()`
  - 包含 `StartSync`、`InspectRoots`、`GetIndexStats`、`GetSyncProgress`、`WatchSyncProgress`、`CancelSync`

### 6.2 健康检查（HTTP）

- `GET /healthz`
- `GET /readyz`

### 6.3 鉴权头支持

要求 API Key 的路由支持：

- `X-API-Key: <key>`
- `Authorization: Bearer <key>`

## 7. 契约与代码生成说明

当前主契约：`proto/npan/v1/api.proto`

生成链路：

```bash
buf lint
buf generate
```

生成产物：

- Go：`gen/go/npan/v1/*.pb.go`
- Go Connect：`gen/go/npan/v1/npanv1connect/*.connect.go`
- Frontend：`web/src/gen/**/*`

## 8. CI / 测试环境端口说明

`docker-compose.ci.yml` 使用以下端口映射：

- 应用：`11323 -> 1323`
- 指标：`19091 -> 9091`
- Meilisearch：`17700 -> 7700`

补充说明：

- `tests/smoke/smoke_test.sh` 默认 `BASE_URL=http://localhost:11323`、`METRICS_URL=http://localhost:19091`。
- `task verify:smoke` 会自动拉起 `docker-compose.ci.yml` 并运行 smoke script。
- `task verify:e2e` 会在 smoke 后继续运行 Playwright 容器。
- `docker-compose.ci.yml` 中的 Playwright 服务默认命令为 `npm install 2>/dev/null; npx playwright test`。

## 9. 常见问题

### Q1: `readyz` 失败

优先检查：

- `MEILI_HOST` 是否可达
- Meilisearch 是否 healthy
- `MEILI_API_KEY` 是否与当前实例一致

### Q2: 管理页一直提示 API Key 无效

检查：

- `.env` 中 `NPA_ADMIN_API_KEY` 是否 `>= 16`
- 输入值是否与服务端配置完全一致
- 是否修改后未重启进程或容器

### Q3: 本地 `go run ./cmd/server` 失败并提示找不到 `dist`

后端会在编译时嵌入 `web/dist`。如果当前工作区没有前端构建产物，请先执行：

```bash
cd web && bun install && bun run build
```

### Q4: E2E 大量超时

迁移后页面已经改为 Connect `POST /npan.v1.*`。如果测试仍在等待旧 REST `/api/v1/*` 路径，或仍按 URL query 断言分页参数，就会超时。

### Q5: 公开搜索没有切到 InstantSearch

检查：

- `MEILI_PUBLIC_INSTANTSEARCH_ENABLED=true`
- `MEILI_PUBLIC_SEARCH_HOST`、`MEILI_PUBLIC_SEARCH_INDEX`、`MEILI_PUBLIC_SEARCH_API_KEY` 是否完整
- `MEILI_PUBLIC_SEARCH_API_KEY` 是否为 dedicated search-only key

## 10. 许可证

MIT
