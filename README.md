# npan

Npan 外部索引服务 — 将 Npan 云盘文件元数据同步到 Meilisearch，提供高性能本地搜索和下载链接代理。

- **Web 框架:** Echo v5
- **搜索引擎:** Meilisearch
- **运行时:** Go 1.25+
- **核心能力:** 自适应同步（自动/全量/增量）、断点续跑、本地全文检索、远程搜索代理、下载链接代理、401 自动刷新 token

## 快速开始

### 1. 准备配置

```bash
cp .env.example .env
cp .env.meilisearch.example .env.meilisearch
```

编辑 `.env`，至少填写：

```bash
NPA_ADMIN_API_KEY=your-admin-key-minimum-16-chars  # 必填，>= 16 字符
NPA_CLIENT_ID=xxx                                   # OAuth 凭据
NPA_CLIENT_SECRET=xxx
NPA_SUB_ID=123
```

### 2. 启动服务

**方式 A — Docker Compose 一键启动（生产推荐）：**

```bash
docker compose up -d --build
```

**方式 B — 本地开发：**

```bash
# 先启动 Meilisearch
docker compose up -d meilisearch

# 再启动 API 服务
go run ./cmd/server
```

服务默认监听 `:1323`，Prometheus 指标在 `:9091`，支持优雅关闭（SIGINT/SIGTERM）。

### 3. CLI 工具

```bash
go run ./cmd/cli --help
```

### 4. Web 前端

服务启动后访问 `http://127.0.0.1:1323/`，React 19 + Vite 单页应用，支持即时搜索和无限滚动。`/admin` 路径提供同步管理面板（模式选择、实时进度、取消）。页面通过 `/api/v1/app/*` 端点访问，凭据由服务端配置处理。

## API 端点

### 公开端点（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 存活检查 |
| GET | `/readyz` | 就绪检查（检测 Meilisearch 连通性） |
| GET | `/app` | Web 搜索页面 |

### App API（内嵌认证，凭据由服务端配置提供）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/app/search` | 应用搜索（仅文件类型） |
| GET | `/api/v1/app/download-url` | 获取下载链接 |

### API（需要 API Key）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/token` | 获取 OAuth access token |
| GET | `/api/v1/search/remote` | 远程搜索（代理 Npan API） |
| GET | `/api/v1/search/local` | 本地搜索（Meilisearch） |
| GET | `/api/v1/download-url` | 获取下载链接 |

### Admin API（需要 API Key）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/sync` | 启动同步（支持 mode: auto/full/incremental） |
| GET | `/api/v1/admin/sync` | 查询同步进度 |
| DELETE | `/api/v1/admin/sync` | 取消同步任务 |

API Key 通过请求头传递：`X-API-Key: <key>` 或 `Authorization: Bearer <key>`。

## 常用命令

<!-- AUTO-GENERATED:makefile-targets -->
### Makefile 快捷命令

| 命令 | 说明 |
|------|------|
| `make test` | Go 单元测试（`-short -count=1 -race`） |
| `make test-frontend` | 前端测试（`cd web && bun run test`） |
| `make generate` | 生成 Go + TypeScript 类型定义 |
| `make generate-check` | 生成并检查是否有未提交的差异 |
| `make smoke-test` | 启动 Docker CI 环境并运行 34 项冒烟测试 |
| `make e2e-test` | 冒烟测试 + Playwright E2E 测试（32 项） |
<!-- /AUTO-GENERATED:makefile-targets -->

### Go 命令

```bash
# 测试
go test ./...

# 竞争检测
go test -race ./...

# 构建
go build ./cmd/server
go build ./cmd/cli

# 启动 HTTP 服务
go run ./cmd/server

# 查询同步进度
go run ./cmd/cli sync-progress

# 自适应同步（有游标走增量，否则全量）
go run ./cmd/cli sync

# 强制全量同步
go run ./cmd/cli sync --mode full

# 强制增量同步
go run ./cmd/cli sync --mode incremental

# JSON 格式进度输出
go run ./cmd/cli sync --progress-output json

# 指定增量查询词与窗口回看
go run ./cmd/cli sync --mode incremental --incremental-query-words "* OR *" --window-overlap-ms 2000
```

## 镜像发布（GitHub Actions）

仓库已提供 `.github/workflows/docker-publish.yml`，用于构建并推送镜像到：

- Docker Hub: `docker.io/<DOCKERHUB_USERNAME>/<repo>`
- GHCR: `ghcr.io/<github_owner>/<repo>`

### 触发条件

- push 到 `main`
- push `v*` tag（例如 `v1.0.0`）
- 手动触发 `workflow_dispatch`

### 需要的仓库 Secrets

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

> GHCR 使用 `GITHUB_TOKEN` 推送，workflow 已配置 `packages: write` 权限。

### Runner 要求（ARM64）

- workflow 会拆分 `amd64` / `arm64` 构建：
  - `amd64` 使用 GitHub 托管 `ubuntu-latest`
  - `arm64` 使用 self-hosted runner，标签要求：
    `self-hosted`, `Linux`, `ARM64`, `debian13`, `trixie`
- 如果 ARM64 runner 离线或标签不匹配，`arm64` build job 会处于 `pending`，最终导致 manifest 合并阶段无法完成。

### 标签策略

- `latest`（仅默认分支）
- `ref` 标签（分支名 / tag 名）
- `sha-<commit>` 短标签

## 环境变量

### 必填

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NPA_ADMIN_API_KEY` | 管理 API 密钥（>= 16 字符） | — |

### 服务

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_ADDR` | 监听地址 | `:1323` |
| `METRICS_ADDR` | Prometheus 指标端口（留空禁用） | `:9091` |
| `NPA_BASE_URL` | Npan OpenAPI 地址 | `https://npan.novastar.tech:6001/openapi` |
| `NPA_ALLOW_CONFIG_AUTH_FALLBACK` | 允许 API 接口回退服务端凭据 | `false` |

### Npan OAuth

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NPA_TOKEN` | 直接使用的 access token | — |
| `NPA_CLIENT_ID` | OAuth client ID | — |
| `NPA_CLIENT_SECRET` | OAuth client secret | — |
| `NPA_SUB_ID` | 授权主体 ID | — |
| `NPA_SUB_TYPE` | 授权主体类型 | `user` |
| `NPA_OAUTH_HOST` | OAuth 服务地址 | Npan 默认值 |

### Meilisearch

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MEILI_HOST` | Meilisearch 地址 | `http://127.0.0.1:7700` |
| `MEILI_API_KEY` | Meilisearch API Key | — |
| `MEILI_INDEX` | 索引名称 | `npan_items` |

### 同步

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NPA_SYNC_MAX_CONCURRENT` | 最大并发请求数（1-20） | `2` |
| `NPA_SYNC_MIN_TIME_MS` | 请求最小间隔（ms） | `200` |
| `NPA_SYNC_ROOT_WORKERS` | 根目录并发 worker 数 | `2` |
| `NPA_SYNC_PROGRESS_EVERY` | 进度报告频率 | `1` |
| `NPA_ROOT_FOLDER_IDS` | 同步根目录 ID 列表（逗号分隔） | `0` |
| `NPA_INCLUDE_DEPARTMENTS` | 是否包含部门文件 | `true` |
| `NPA_DEPARTMENT_IDS` | 部门 ID 列表（逗号分隔） | — |
| `NPA_SYNC_STATE_FILE` | 增量游标状态文件路径 | `./data/progress/incremental-sync-state.json` |
| `NPA_INCREMENTAL_QUERY_WORDS` | 增量查询词 | `* OR *` |
| `NPA_SYNC_WINDOW_OVERLAP_MS` | 增量窗口回看毫秒数 | `2000` |

### 重试策略

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NPA_MAX_RETRIES` | 最大重试次数（0-10） | `3` |
| `NPA_BASE_DELAY_MS` | 基础重试延迟（ms） | `500` |
| `NPA_MAX_DELAY_MS` | 最大重试延迟（ms） | `5000` |
| `NPA_JITTER_MS` | 重试抖动（ms） | `200` |

## Docker 部署

### 一键启动（推荐）

使用 docker compose 同时启动 Meilisearch + npan 服务：

```bash
# 准备配置
cp .env.example .env
cp .env.meilisearch.example .env.meilisearch

# 编辑 .env 填入必要配置（NPA_ADMIN_API_KEY, OAuth 凭据等）
# 编辑 .env.meilisearch 设置 MEILI_MASTER_KEY

# 构建并启动
docker compose up -d --build

# 查看日志
docker compose logs -f npan

# 停止
docker compose down
```

服务启动后：
- Web UI: `http://<host>:1323/`
- Admin: `http://<host>:1323/admin`
- Prometheus 指标: `http://<host>:9091/metrics`
- Meilisearch: `http://<host>:7700`

> docker compose 中 `MEILI_HOST` 会自动设为 `http://meilisearch:7700`（容器内网络），无需手动配置。

### 单独构建镜像

```bash
# 构建镜像
docker build -t npan .

# 运行（需自行指定 Meilisearch 地址）
docker run -d \
  --env-file .env \
  -p 1323:1323 \
  -p 9091:9091 \
  -v npan-data:/app/data \
  npan
```

镜像特性：多阶段构建、非 root 用户运行、内置健康检查。

## 安全

### 认证与授权

- `NPA_ADMIN_API_KEY` **必填**，启动时校验（空 key 直接 panic），保护 `/api/v1/*` 和 `/api/v1/admin/*` 端点
- API Key 使用 constant-time 比较，防止计时攻击
- `NPA_ALLOW_CONFIG_AUTH_FALLBACK` 默认关闭，防止 API 接口意外使用服务端凭据

### HTTP 安全中间件栈

<!-- AUTO-GENERATED:middleware-stack -->
| 中间件 | 说明 |
|--------|------|
| RequestID | 为每个请求生成唯一 ID |
| Recover | 捕获 panic 防止进程崩溃 |
| SecureHeaders | 设置 `X-Content-Type-Options: nosniff`、`X-Frame-Options: DENY`、`Referrer-Policy`、`Permissions-Policy` |
| RequestLogger | 结构化请求日志 |
| BodyLimit(1MB) | 限制请求体大小，防止大 payload DoS |
| RateLimitMiddleware | 全局 20 rps / burst 40；Admin 路由 5 rps / burst 10 |
| IPExtractor(Direct) | 直连 IP 提取（部署在反代后需改为 XFF） |
| HTTPErrorHandler | 统一 JSON 错误响应，5xx 消息自动脱敏为 `"服务器内部错误"` |
<!-- /AUTO-GENERATED:middleware-stack -->

### 输入验证

- `type` 参数白名单校验（`all`、`file`、`folder`），防止 Meilisearch 过滤注入
- `page_size` 上限 100，防止资源耗尽
- `checkpoint_template` 路径遍历防护（禁止绝对路径、`..`、必须在 `data/checkpoints/` 下）

### 其他

- 错误响应不泄露内部信息（堆栈、地址、凭据等），仅返回结构化错误码和消息
- 敏感配置在日志中自动脱敏（`[REDACTED]`）
- 生产环境不要提交 `.env`，仅提交 `.env.example`

## 项目结构

```text
.
├── cmd/
│   ├── server/          # HTTP API 服务入口
│   └── cli/             # CLI 工具入口
├── internal/
│   ├── cli/             # CLI 命令定义
│   ├── config/          # 配置加载与验证
│   ├── httpx/           # HTTP 路由、中间件、Handler
│   ├── indexer/         # 索引写入逻辑
│   ├── logx/            # 日志初始化
│   ├── models/          # 数据模型
│   ├── npan/            # Npan API 客户端与认证
│   ├── search/          # Meilisearch 查询服务
│   ├── service/         # 同步管理器等业务服务
│   └── storage/         # 进度持久化
├── web/                # React 19 + Vite 前端（搜索页 + Admin 同步管理）
├── data/                # 运行时状态文件（不提交）
├── docs/
│   ├── plans/           # 设计与实施计划
│   ├── runbooks/        # 运维手册
│   ├── reference/       # 外部参考资料
│   └── archive/         # 历史记录归档
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── .env.example
└── .env.meilisearch.example
```

## 文档

- 结构说明: `docs/STRUCTURE.md`
- 运维手册: `docs/runbooks/index-sync-operations.md`
- 历史归档: `docs/archive/README.md`
