# npan

Npan 外部索引服务：把 Npan 云盘文件元数据同步到 Meilisearch，提供可直接使用的搜索与下载能力。

- 后端：Go + Echo
- 前端：React + Vite
- 搜索：Meilisearch
- RPC：Buf + Connect-RPC（connect-go / connect-es）

## 1. 你会得到什么

- Web 搜索页：`/`（关键词检索、分页、下载）
- 管理后台：`/admin/`（启动同步、取消同步、查看进度）
- Connect-RPC API（主路径）：`/npan.v1.*`

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
NPA_CLIENT_ID=...
NPA_CLIENT_SECRET=...
NPA_SUB_ID=...
# 可选：覆盖默认 SQLite 状态库位置
# NPA_STATE_DB_FILE=./data/state/sync-state.sqlite
```

说明：
- `NPA_ADMIN_API_KEY` 必填，且长度必须 >= 16。
- 若只做最小功能联调，也可先提供 `NPA_TOKEN`，跳过 OAuth 三元组。
- `NPA_STATE_DB_FILE` 是同步状态的主存储，默认路径为 `./data/state/sync-state.sqlite`。
- `NPA_PROGRESS_FILE` 与 `NPA_SYNC_STATE_FILE` 仍保留为 legacy JSON 导入来源，用于首次惰性迁移与人工对照，不再是运行时主状态源。

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

### 2.7 本地 Docker + 真实凭据跑 admin live E2E

### 2.6 本地 Docker + 真实凭据跑 admin live E2E

当你需要对 `/admin` 关键环节做真实数据实测（不使用 route mock）时，使用 live 覆盖配置。
支持两种认证输入：

- 直接提供 `NPA_TOKEN`
- 或提供 `NPA_CLIENT_ID/NPA_CLIENT_SECRET/NPA_SUB_ID`，由服务端自动换取 token

示例（OAuth 三元组方式）：

```bash
# 1) 启动服务（提供真实 OAuth 凭据；也可改为直接传 NPA_TOKEN）
NPA_CLIENT_ID='<your-client-id>' \
NPA_CLIENT_SECRET='<your-client-secret>' \
NPA_SUB_ID='<your-sub-id>' \
NPA_SUB_TYPE='enterprise' \
docker compose \
  -f docker-compose.ci.yml \
  -f docker-compose.e2e-live.yml \
  up --build -d --wait --wait-timeout 120

# 2) 运行 admin live E2E（InspectRoots + 全量启动）
docker compose \
  -f docker-compose.ci.yml \
  -f docker-compose.e2e-live.yml \
  --profile e2e run --rm playwright \
  npx playwright test e2e/tests/admin.live.spec.ts

# 3) 清理
docker compose \
  -f docker-compose.ci.yml \
  -f docker-compose.e2e-live.yml \
  --profile e2e down --volumes
```

说明：
- live 测试通过 `E2E_LIVE=1` 启用（由 `docker-compose.e2e-live.yml` 注入）。
- `docker-compose.e2e-live.yml` 会把服务端认证来源切到你传入的真实 token 或 OAuth 三元组。
- `InspectRoots` live 用例会自动预热：若尚无根目录 catalog，会先触发一次全量启动并等待目录列表出现。

可选性能调优（用于 `/npan.v1.AdminService/InspectRoots`）：

- `NPA_INSPECT_ROOTS_MAX_CONCURRENCY`：目录详情并发度，默认 `6`
- `NPA_INSPECT_ROOTS_PER_FOLDER_TIMEOUT`：单目录请求超时（Go duration），默认 `10s`

## 3. 首次同步（两种方式）

### 3.1 方式 A：管理后台（推荐）

1. 打开 `http://localhost:1323/admin/`
2. 输入 `NPA_ADMIN_API_KEY`
3. 选择同步模式（全量 / 增量）
4. 点击启动，同步状态会自动轮询刷新

### 3.2 方式 B：API 调用

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

> 项目默认前端包管理器是 `bun`。

## 5. 常用命令

```bash
# 后端测试
GOCACHE=/tmp/go-build go test ./...

# 前端测试
cd web && bun vitest run

# 防回退检查（禁止在运行时代码引入 /api/v1 路径）
make rest-guard

# CI 冒烟测试（34 项）
make smoke-test

# 冒烟 + E2E（32 项）
make e2e-test
```

## 6. API 入口总览

### 6.1 Connect-RPC（运行时主路径）

- 公开：
  - `/npan.v1.HealthService/*`
  - `/npan.v1.AppService/*`
- 管理（需 API Key）：
- `/npan.v1.AppService/*`
- `/npan.v1.AuthService/*`
- `/npan.v1.SearchService/*`
- `/npan.v1.AdminService/*`

### 6.2 健康检查（HTTP）

- `GET /healthz`
- `GET /readyz`

鉴权头支持：
- `X-API-Key: <key>`
- `Authorization: Bearer <key>`

## 7. 契约与代码生成说明（开发者关心）

当前仓库使用 Connect 契约（Buf/Proto）：`proto/npan/v1/api.proto`。

对应生成链路：

```bash
# Connect / protobuf / connect-go / connect-es
buf lint
buf generate
```

## 8. CI / 测试环境端口说明

`docker-compose.ci.yml` 使用的是 CI 端口映射：

- 应用：`11323 -> 1323`
- 指标：`19091 -> 9091`

`tests/smoke/smoke_test.sh` 默认值已经对齐该映射，可直接运行。

## 9. 常见问题

### Q1: `readyz` 失败

优先检查：
- `MEILI_HOST` 是否能连通
- Meilisearch 是否已 healthy
- `MEILI_API_KEY` 是否与 Meilisearch 一致

### Q2: 管理页一直提示 API Key 无效

检查：
- `.env` 中 `NPA_ADMIN_API_KEY` 是否 >= 16
- 输入值是否与 `.env` 完全一致
- 是否修改后未重启容器

### Q3: E2E 大量超时

迁移后页面走 Connect 路径，若自定义测试仍在等待旧 HTTP 路径，会超时。请改为等待 `/npan.v1.*` 请求。

## 10. 许可证

MIT
