# npan

Npan 外部索引服务：把 Npan 云盘文件元数据同步到 Meilisearch，提供 Web 搜索页、管理后台、Connect-RPC API 和运维 CLI。

- 后端：Go + Echo
- 前端：React + Vite + Bun
- 搜索：Meilisearch
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
```

说明：

- `NPA_ADMIN_API_KEY` 必填，且长度必须 `>= 16`。
- 上游认证支持两种来源：
  - 直接提供 `NPA_TOKEN`
  - 提供 `NPA_CLIENT_ID` / `NPA_CLIENT_SECRET` / `NPA_SUB_ID`，由服务端换取 token
- 若要启用浏览器直连 Meilisearch 的公开搜索，还需要配置 `MEILI_PUBLIC_*`，并使用 dedicated search-only key。

### 2.3 启动

```bash
docker compose up -d --build
```

默认端口：

- 应用：`http://localhost:1323`
- 指标：`http://localhost:9091/metrics`
- Meilisearch：`http://localhost:7700`

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

### 2.6 可选：启用浏览器公开搜索

当你希望搜索页优先走浏览器直连 Meilisearch 的 public InstantSearch 链路时，配置：

```bash
MEILI_PUBLIC_SEARCH_HOST=http://127.0.0.1:7700
MEILI_PUBLIC_SEARCH_INDEX=npan_items
MEILI_PUBLIC_SEARCH_API_KEY=<search-only-key>
MEILI_PUBLIC_INSTANTSEARCH_ENABLED=true
```

注意：

- `MEILI_PUBLIC_SEARCH_API_KEY` 必须是 dedicated search-only key，不能复用私有 `MEILI_API_KEY`。
- 前端会先通过 `AppService.GetSearchConfig` 拉取公开搜索配置。
- 若公开搜索配置不完整，或 `MEILI_PUBLIC_INSTANTSEARCH_ENABLED=false`，搜索页会自动回退到 legacy Connect `AppSearch` 链路。

### 2.7 本地 Docker + 真实凭据跑 admin live E2E

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

- CLI 默认从环境变量读取凭据和 Meilisearch 配置。
- `sync` 支持 `--progress-output human|json`。
- 默认进度文件与状态文件来自 `.env` / `.env.example` 中的 `NPA_PROGRESS_FILE`、`NPA_SYNC_STATE_FILE`、`NPA_CHECKPOINT_FILE`。

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

```bash
# 后端测试（与 Makefile 一致，含 -short/-count=1/-race）
make test

# 前端测试
make test-frontend

# 前端类型检查
cd web && bun run typecheck

# 防回退检查（禁止在运行时代码中重新引入 /api/v1 路径）
make rest-guard

# CI compose 冒烟回归
make smoke-test

# CI compose 冒烟 + Playwright E2E
make e2e-test

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
- `make smoke-test` 会自动拉起 `docker-compose.ci.yml` 并运行 smoke script。
- `make e2e-test` 会在 smoke 后继续运行 Playwright 容器。
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
