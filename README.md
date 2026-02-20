# Npan Index Service (Go)

基于 Go 的 Npan 外部索引服务，使用 `Echo v5` + `Meilisearch`，支持：
- 全量同步（含断点恢复）
- 本地索引检索
- 实时下载链接代理
- token 自动刷新（401 后自动刷新并重试 1 次）

## 目录结构

```text
.
├─ cmd/
│  ├─ server/                # API 服务入口
│  └─ cli/                   # CLI 入口
├─ internal/
│  ├─ cli/                   # CLI 子命令实现
│  ├─ config/                # 配置与 .env 加载
│  ├─ httpx/                 # Echo 路由与 handler
│  ├─ indexer/               # 全量/增量/限流/重试
│  ├─ models/                # 领域模型
│  ├─ npan/                  # Npan 认证与 API 客户端
│  ├─ search/                # Meili 封装与查询
│  ├─ service/               # 业务编排（同步调度、下载代理）
│  └─ storage/               # JSON 持久化（checkpoint/progress）
├─ data/
│  ├─ checkpoints/           # 运行时断点文件（默认忽略提交）
│  ├─ progress/              # 运行时进度文件（默认忽略提交）
│  └─ dumps/                 # 导出文件（默认忽略提交）
├─ docs/
│  ├─ archive/               # 历史计划/工作记录归档
│  ├─ reference/             # 参考资料
│  └─ runbooks/              # 运维手册
├─ .env.example
├─ docker-compose.yml
├─ go.mod
└─ go.sum
```

## 快速开始

1. 准备环境变量：

```bash
cp .env.example .env
```

2. 启动 Meilisearch（可选）：

```bash
docker compose up -d meilisearch
```

3. 启动 API 服务：

```bash
go run ./cmd/server
```

4. 使用 CLI：

```bash
go run ./cmd/cli --help
```

## 核心命令

```bash
# 构建
go build ./...

# 启动服务
go run ./cmd/server

# CLI 帮助
go run ./cmd/cli --help

# 查看同步进度
go run ./cmd/cli sync-progress
```

## 主要 API

- `GET /healthz`
- `POST /api/v1/token`
- `GET /api/v1/npan/search`
- `GET /api/v1/search/local`
- `GET /api/v1/download-url`
- `POST /api/v1/sync/full/start`
- `GET /api/v1/sync/full/progress`
- `POST /api/v1/sync/full/cancel`

## token 自动刷新说明

若请求 Npan 返回 `401`，客户端会自动尝试刷新 token 并重试一次。
启用条件：
- `NPA_CLIENT_ID`
- `NPA_CLIENT_SECRET`
- `NPA_SUB_ID`

仅配置 `NPA_TOKEN` 时，使用静态 token，不触发自动刷新。
