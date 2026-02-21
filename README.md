# npan

Npan 外部索引服务 — 将 Npan 云盘文件元数据同步到 Meilisearch，提供高性能本地搜索和下载链接代理。

- **Web 框架:** Echo v5
- **搜索引擎:** Meilisearch
- **运行时:** Go 1.25+
- **核心能力:** 全量同步、断点续跑、增量同步、本地全文检索、远程搜索代理、下载链接代理、401 自动刷新 token

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

### 2. 启动 Meilisearch

```bash
docker compose up -d meilisearch
```

若已有 Meilisearch 实例，设置 `MEILI_HOST` 和 `MEILI_API_KEY` 即可跳过。

### 3. 启动 API 服务

```bash
go run ./cmd/server
```

服务默认监听 `:1323`，支持优雅关闭（SIGINT/SIGTERM）。

### 4. CLI 工具

```bash
go run ./cmd/cli --help
```

### 5. Web 搜索页面

服务启动后访问 `http://127.0.0.1:1323/app`，纯 HTML 搜索页面，支持即时搜索和无限滚动。页面通过 `/api/v1/app/*` 端点访问，凭据由服务端配置处理，无需用户输入 token。

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
| POST | `/api/v1/admin/sync/full` | 启动全量同步 |
| GET | `/api/v1/admin/sync/full/progress` | 查询同步进度 |
| POST | `/api/v1/admin/sync/full/cancel` | 取消同步任务 |

API Key 通过请求头传递：`X-API-Key: <key>` 或 `Authorization: Bearer <key>`。

## 常用命令

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

# 全量同步（默认人类可读进度）
go run ./cmd/cli sync-full

# 全量同步（结构化 JSON 进度）
go run ./cmd/cli sync-full --progress-output json

# 增量同步
go run ./cmd/cli sync-incremental

# 增量同步（显式指定查询词与窗口回看）
go run ./cmd/cli sync-incremental --incremental-query-words "* OR *" --window-overlap-ms 2000
```

## 环境变量

### 必填

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NPA_ADMIN_API_KEY` | 管理 API 密钥（>= 16 字符） | — |

### 服务

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_ADDR` | 监听地址 | `:1323` |
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

### 重试策略

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NPA_MAX_RETRIES` | 最大重试次数（0-10） | `3` |
| `NPA_BASE_DELAY_MS` | 基础重试延迟（ms） | `500` |
| `NPA_MAX_DELAY_MS` | 最大重试延迟（ms） | `5000` |
| `NPA_JITTER_MS` | 重试抖动（ms） | `200` |

## Docker

```bash
# 构建镜像
docker build -t npan .

# 运行
docker run -d \
  --env-file .env \
  -p 1323:1323 \
  -v npan-data:/app/data \
  npan
```

镜像特性：多阶段构建、非 root 用户运行、内置健康检查。

## 安全

- `NPA_ADMIN_API_KEY` **必填**，启动时校验，保护 `/api/v1/*` 和 `/api/v1/admin/*` 端点
- API Key 使用 constant-time 比较，防止计时攻击
- `NPA_ALLOW_CONFIG_AUTH_FALLBACK` 默认关闭，防止 API 接口意外使用服务端凭据
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
├── web/
│   └── app/             # Web 搜索页面（纯 HTML）
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
