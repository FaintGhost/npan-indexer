# npan

`npan` 是一个把 Npan / Fangcloud 云盘文件元数据同步到本地搜索引擎的服务，提供：

- Web 搜索页
- 管理后台（启动/取消同步、查看进度、检查根目录）
- Connect-RPC API
- 运维 CLI

当前项目状态：

- 运行时已全面切换到 Connect-RPC，不再提供 `/api/v1/*`
- 搜索后端支持 `Meilisearch` 和 `Typesense`
- 同步状态默认持久化到 SQLite
- 前端默认使用 `Bun`

## 1. 功能概览

启动后你会得到这些入口：

- 搜索页：`/`
- 管理页：`/admin/`
- Connect-RPC：`/npan.v1.*`
- CLI：`go run ./cmd/cli ...`

项目主要能力：

- 全量同步根目录和部门目录到本地索引
- 基于时间窗口的增量同步
- 管理后台查看同步状态、索引统计和根目录检查结果
- 公开搜索可按配置切换为浏览器直连 public search
- 支持本地搜索、远程搜索、下载链接生成

## 2. 技术栈

- 后端：Go 1.25+、Echo v5
- 前端：React 19、Vite、TanStack Router、Bun
- RPC：Buf + Protobuf + Connect-RPC
- 搜索：Meilisearch / Typesense
- 状态存储：SQLite（`modernc.org/sqlite`）
- 测试：Vitest、Playwright、Go test

## 3. 快速开始

### 3.1 前置条件

- Docker 24+
- Docker Compose v2

### 3.2 准备环境变量

```bash
cp .env.example .env
cp .env.meilisearch.example .env.meilisearch
cp .env.typesense.example .env.typesense
```

至少需要配置这些值：

```bash
NPA_ADMIN_API_KEY=your-admin-key-minimum-16-chars

# 二选一
NPA_TOKEN=...

# 或使用 OAuth 三元组
NPA_CLIENT_ID=...
NPA_CLIENT_SECRET=...
NPA_SUB_ID=...
```

常用配置说明：

- `NPA_ADMIN_API_KEY` 必填，长度必须 `>= 16`
- `NPA_SEARCH_BACKEND` 默认是 `meilisearch`，可切换为 `typesense`
- `NPA_TOKEN` 适合最小部署；如不提供，则由服务端使用 OAuth 三元组换取 token
- `NPA_STATE_DB_FILE` 默认是 `./data/state/sync-state.sqlite`
- `NPA_ROOT_FOLDER_IDS` 默认是 `0`
- `NPA_INCREMENTAL_QUERY_WORDS` 默认是 `* OR *`

### 3.3 启动

```bash
docker compose up -d --build
```

默认端口：

- 应用：`http://localhost:1323`
- 指标：`http://localhost:9091/metrics`
- Meilisearch：`http://localhost:7700`
- Typesense：`http://localhost:8108`

### 3.4 健康检查

```bash
curl -sS http://localhost:1323/healthz
curl -sS http://localhost:1323/readyz
```

期望响应：

```json
{"status":"ok"}
{"status":"ready"}
```

### 3.5 访问页面

- 搜索页：`http://localhost:1323/`
- 管理页：`http://localhost:1323/admin/`

## 4. 同步与管理

### 4.1 管理后台（推荐）

1. 打开 `http://localhost:1323/admin/`
2. 输入 `NPA_ADMIN_API_KEY`
3. 选择同步模式（全量 / 增量）
4. 启动同步并观察进度、统计和根目录检查结果

说明：

- 增量同步使用时间窗口 + 重叠区间
- 当 live 目录计数与本地索引子树计数出现漂移时，系统会优先做目录级定向补偿，而不是总是整根目录重扫

### 4.2 Connect-RPC

启动同步：

```bash
curl -sS -X POST \
  -H 'X-API-Key: <your-admin-key>' \
  -H 'Content-Type: application/json' \
  -d '{"mode":"SYNC_MODE_FULL"}' \
  http://localhost:1323/npan.v1.AdminService/StartSync
```

查看进度：

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
- `POST /npan.v1.AdminService/WatchSyncProgress`

### 4.3 CLI

首次全量同步：

```bash
go run ./cmd/cli sync --mode full --root-folder-ids 0
```

增量同步：

```bash
go run ./cmd/cli sync --mode incremental --incremental-query-words "* OR *"
```

查看同步进度：

```bash
go run ./cmd/cli sync-progress
```

其他常用命令：

```bash
go run ./cmd/cli token
go run ./cmd/cli search-remote --query "关键词"
go run ./cmd/cli search-local --query "关键词"
go run ./cmd/cli download-url --file-id 123
```

## 5. 搜索后端与公开搜索

### 5.1 切换搜索后端

通过 `NPA_SEARCH_BACKEND` 选择：

```bash
NPA_SEARCH_BACKEND=meilisearch
# 或
NPA_SEARCH_BACKEND=typesense
```

### 5.2 启用公开搜索（浏览器直连）

```bash
NPA_PUBLIC_INSTANTSEARCH_ENABLED=true

# Meilisearch
MEILI_PUBLIC_SEARCH_HOST=http://127.0.0.1:7700
MEILI_PUBLIC_SEARCH_INDEX=npan_items
MEILI_PUBLIC_SEARCH_API_KEY=<search-only-key>

# Typesense
TYPESENSE_PUBLIC_SEARCH_HOST=http://127.0.0.1:8108
TYPESENSE_PUBLIC_SEARCH_INDEX=npan_items
TYPESENSE_PUBLIC_SEARCH_API_KEY=<search-only-key>
```

注意：

- public search key 必须是 search-only key
- public host 必须是浏览器可访问的地址
- 如果公开搜索配置不完整，前端会回退到服务端 `AppSearch` 链路

最小验证：

```bash
curl -sS -X POST \
  -H 'Content-Type: application/json' \
  -d '{}' \
  http://localhost:1323/npan.v1.AppService/GetSearchConfig
```

## 6. 状态存储

同步状态默认写入 SQLite：

- 默认路径：`./data/state/sync-state.sqlite`
- 配置项：`NPA_STATE_DB_FILE`

状态库统一保存：

- 同步进度
- 增量游标
- crawl checkpoint

兼容说明：

- `NPA_PROGRESS_FILE`
- `NPA_SYNC_STATE_FILE`
- `NPA_CHECKPOINT_FILE`

这些 legacy JSON 路径仍可用于导入和人工对照，但不再是运行时主状态源。

## 7. 本地开发

### 7.1 启动依赖

```bash
docker compose up -d meilisearch
```

如需 Typesense，也可以一并启动：

```bash
docker compose up -d meilisearch typesense
```

### 7.2 启动后端

```bash
go run ./cmd/server
```

### 7.3 启动前端

```bash
cd web
bun install
bun run dev
```

### 7.4 前端构建产物

后端会嵌入 `web/dist`。如果你直接运行 `go run ./cmd/server` 且本地没有构建产物，请先执行：

```bash
cd web && bun install && bun run build
```

## 8. 常用验证命令

```bash
# 查看任务
task --list

# Go
task test:go

# Frontend
task test:web
cd web && bun run typecheck

# Connect-only 运行时守卫
task guard:rest

# Docker 冒烟
task verify:smoke

# Docker 冒烟 + Playwright E2E
task verify:e2e

# 完整回归
task verify:all

# 契约变更后
buf lint
buf generate
```

CI 端口映射：

- 应用：`11323 -> 1323`
- 指标：`19091 -> 9091`
- Meilisearch：`17700 -> 7700`

## 9. API 与鉴权

当前运行时是 Connect-only，主路径都在 `/npan.v1.*` 下。

- `/npan.v1.HealthService/*`
- `/npan.v1.AppService/*`
- `/npan.v1.AuthService/*`
- `/npan.v1.SearchService/*`
- `/npan.v1.AdminService/*`

HTTP 健康检查：

- `GET /healthz`
- `GET /readyz`

要求 API Key 的接口支持两种头：

- `X-API-Key: <key>`
- `Authorization: Bearer <key>`

## 10. 代码生成

主契约文件：

- `proto/npan/v1/api.proto`

生成命令：

```bash
buf lint
buf generate
```

生成产物：

- Go protobuf：`gen/go/npan/v1/*.pb.go`
- Go Connect：`gen/go/npan/v1/npanv1connect/*.connect.go`
- Frontend：`web/src/gen/**/*`

## 11. 常见问题

### Q1: `readyz` 失败

优先检查：

- `MEILI_HOST` / `TYPESENSE_HOST` 是否可达
- 当前搜索后端是否已启动
- API key 是否与实例匹配

### Q2: 管理页提示 API Key 无效

检查：

- `NPA_ADMIN_API_KEY` 是否长度 `>= 16`
- 输入值是否与服务端配置完全一致
- 修改 `.env` 后是否已重启服务

### Q3: 本地运行时报错找不到 `dist`

请先构建前端：

```bash
cd web && bun install && bun run build
```

### Q4: 增量同步很快结束，但数量不一致

先在管理页执行根目录检查，确认 live 计数和本地索引计数是否存在漂移。当前版本会优先做目录级定向补偿；如果仍无法恢复，再考虑做一次明确的全量同步。

### Q5: E2E 大量超时

优先确认测试是否仍在等待旧 REST `/api/v1/*` 路径。当前页面请求已经是 Connect `POST /npan.v1.*`，很多参数也在请求体里而不是 URL query。

## 12. 许可证

MIT
