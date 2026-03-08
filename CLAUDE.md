# CLAUDE.md

> 面向代码代理（coding agent）的项目接手说明。请与平台级 `AGENTS.md` 一起阅读：
> - `AGENTS.md` 负责通用协作规则/工作流
> - 本文件负责仓库事实、入口位置、验证命令、常见坑

## 0. 一句话说明

`npan` 是一个将 Npan 云盘文件元数据同步到 Meilisearch 的服务，提供：
- Web 搜索页面（React + Vite）
- 管理后台（同步启动/取消/进度）
- Connect-RPC API（主路径）

当前迁移状态：
- 已接入 `buf` 生成链路
- 后端使用 `connect-go`
- 前端使用 `connect-es` + `connect-query`
- 运行时已全面切换到 Connect（不再提供 `/api/v1/*`）

## 1. 技术栈与运行边界

- 后端：Go 1.25+, Echo v5, Meilisearch, Prometheus
- 前端：React 19, Vite, TanStack Router, Bun, Vitest, Playwright
- RPC：Buf + Protobuf + Connect-RPC（Connect/gRPC/gRPC-Web handler 由 connect-go 生成）
- 状态持久化：SQLite（`modernc.org/sqlite`，默认文件 `./data/state/sync-state.sqlite`）
- 契约：
  - Protobuf（Connect 路径）：`proto/npan/v1/api.proto`

## 2. 目录地图（优先阅读）

### 服务端核心

- `cmd/server`：HTTP 服务启动入口（加载配置、启动 Echo、嵌入前端）
- `cmd/cli`：CLI 入口（同步、进度查询等）
- `internal/httpx`：HTTP 路由、鉴权、中间件、Connect server adapter
- `internal/service`：业务服务层（同步编排、搜索等）
- `internal/npan`：Npan API/OAuth 客户端封装
- `internal/search`：Meilisearch 查询与索引交互
- `internal/indexer`：同步/抓取/索引写入逻辑
- `internal/config`：环境变量配置与校验

### 前端核心

- `web/src/routes`：页面路由（搜索页 `/`、管理页 `/admin`）
- `web/src/components`：页面与 UI 组件
- `web/src/hooks`：前端 hooks（下载、鉴权、热键等）
- `web/src/lib/connect-transport.ts`：Connect transport / QueryClient 配置
- `web/src/lib/*adapter.ts`：Proto <-> UI domain 映射
- `web/e2e`：Playwright E2E（admin/search/download/边界场景）

### 契约与生成代码

- `proto/npan/v1/api.proto`：Connect/Buf 主契约（RPC + message）
- `buf.yaml` / `buf.gen.yaml`：Buf lint/codegen 配置
- `gen/go/npan/v1`：Buf 生成的 Go protobuf / connect-go 代码
- `web/src/gen`：Buf 生成的前端 protobuf / connect-es / connect-query 代码

## 3. 路由现状（迁移后很关键）

在 `internal/httpx/server.go` 中，当前是 Connect-only：

- Connect-RPC（主路径）
  - `/npan.v1.HealthService/*`
  - `/npan.v1.AppService/*`
  - `/npan.v1.AuthService/*`
  - `/npan.v1.SearchService/*`
  - `/npan.v1.AdminService/*`

实践建议：
- 前端新逻辑优先接 Connect
- 改 E2E 时优先校验 Connect `POST /npan.v1.*` 请求体

## 4. 生成链路（改契约时必须看）

### 4.1 Connect / Protobuf（Buf）

修改 `proto/npan/v1/api.proto` 后：

```bash
buf lint
buf generate
```

会更新：
- `gen/go/npan/v1/*.pb.go`
- `gen/go/npan/v1/npanv1connect/*.connect.go`
- `web/src/gen/**/*`

## 5. 开发与验证命令（默认使用 Bun）

### 本地开发

```bash
# 启动依赖（Meilisearch）
docker compose up -d meilisearch

# 后端
go run ./cmd/server

# 前端（可选独立 dev）
cd web && bun install && bun run dev
```

### 单元测试

```bash
# Go
GOCACHE=/tmp/go-build go test ./...

# Frontend
cd web && bun vitest run
```

### Docker 冒烟 / E2E（推荐回归链）

```bash
# 冒烟（34 项）
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
./tests/smoke/smoke_test.sh

# Playwright E2E（32 项）
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright

# 清理
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
```

注意：
- `tests/smoke/smoke_test.sh` 默认端口已对齐 `docker-compose.ci.yml`（`11323` / `19091`）
- `docker-compose.yml`（开发/部署）端口仍是 `1323` / `9091`

## 6. 常见坑（迁移后高频）

### 6.1 E2E 等待条件失效

现象：`waitForRequest` / `waitForResponse` 大量超时。

高概率原因：
- 页面已经改为 Connect `POST /npan.v1.*`，但测试仍在等旧协议路径
- Connect 请求参数在 JSON body 中，不在 URL query 上（例如 `page=2`）

处理方式：
- 先核对真实请求路径与 method
- 优先校验 request body，而不是 URL query
- 超时按场景收紧（3s/5s/10s），不要默认 30s

### 6.2 生成代码不一致

现象：本地编译通过但 CI 失败。

排查顺序：
1. 是否改了 `proto/...` 但漏跑 `buf generate`
2. 是否漏提 `gen/go` 与 `web/src/gen` 产物
3. 是否修改了连接层 adapter 但未同步测试

### 6.3 `go:embed` / 前端产物

后端会嵌入前端构建产物（`web/dist`）。
- 本地 `go run ./cmd/server` 前若无 `web/dist`，需要先构建前端或使用仓库已有产物
- Dockerfile 会自动构建前端并复制到镜像

### 6.4 SQLite 状态库与 legacy JSON

当前同步状态默认已切到 SQLite：
- 默认路径：`NPA_STATE_DB_FILE=./data/state/sync-state.sqlite`
- `cmd/server/main.go` 与 `internal/cli/root.go` 都通过 `storage.NewSQLiteStateStores(...)` 初始化状态存储
- `progress` / `sync_state` / `checkpoint` 统一写入 SQLite

兼容边界：
- `NPA_PROGRESS_FILE` 与 `NPA_SYNC_STATE_FILE` 仍保留，但角色仅是 legacy 导入来源
- 不要把 legacy JSON 当作运行时主状态源
- 排障时优先检查 SQLite 文件，再决定是否对照旧 JSON
- CLI `sync-progress` 默认读 SQLite，可通过 `--state-db-file` 指定路径

## 7. 改动建议（给接手 agent）

### 如果你在改前端功能

优先阅读：
- `web/src/routes/index.lazy.tsx`
- `web/src/components/admin-sync-page.tsx`
- `web/src/lib/connect-transport.ts`
- `web/src/lib/connect-*-adapter.ts`

并执行至少：
```bash
cd web && bun vitest run
```

### 如果你在改后端 API / 同步逻辑

优先阅读：
- `internal/httpx/server.go`
- `internal/httpx/handlers*.go`
- `internal/service/*`
- `internal/indexer/*`

并执行至少：
```bash
GOCACHE=/tmp/go-build go test ./...
```

### 如果你在改契约（字段/RPC）

请同时考虑：
- Connect protobuf 契约（Buf）
- E2E 与 smoke 的断言路径/请求格式是否需要更新

## 8. 提交前最小检查单

```bash
# 1) 契约生成（如改了契约）
buf lint && buf generate

# 2) 单测
GOCACHE=/tmp/go-build go test ./...
cd web && bun vitest run

# 3) 长链路（改了接口/页面/鉴权/同步流程时）
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
./tests/smoke/smoke_test.sh
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
```

---

如果你刚接手这个仓库，建议第一轮只做两件事：
1. 跑通 `go test ./...` 与 `cd web && bun vitest run`
2. 阅读 `internal/httpx/server.go` 和 `web/src/routes/index.lazy.tsx`，确认 Connect-only 结构
