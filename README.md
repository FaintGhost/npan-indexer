# npan

一个面向生产的 Npan 外部索引服务（Go）。

- Web: `Echo v5`
- Index: `Meilisearch`
- Runtime: Go 1.25+
- 能力: 全量同步、断点续跑、本地检索、下载链接代理、401 自动刷新 token

## 快速开始

1. 准备配置。

```bash
cp .env.example .env
```

2. 启动 Meilisearch（可选，若你已有实例可跳过）。

```bash
docker compose up -d meilisearch
```

3. 启动 API 服务。

```bash
go run ./cmd/server
```

4. 查看 CLI 命令。

```bash
go run ./cmd/cli --help
```

## 常用命令

```bash
# 测试
go test ./...

# 竞争检测
go test -race ./...

# 构建
go build ./...

# 启动 HTTP 服务
go run ./cmd/server

# 查询同步进度
go run ./cmd/cli sync-progress
```

## API 入口

- `GET /healthz`
- `POST /api/v1/token`
- `GET /api/v1/npan/search`
- `GET /api/v1/search/local`
- `GET /api/v1/download-url`
- `POST /api/v1/sync/full/start`
- `GET /api/v1/sync/full/progress`
- `POST /api/v1/sync/full/cancel`

## 安全基线

- 推荐设置 `NPA_ADMIN_API_KEY`，保护 `/api/v1/*`（请求头：`X-API-Key`）。
- 默认 `NPA_ALLOW_CONFIG_AUTH_FALLBACK=false`，HTTP 接口不会自动回退服务端凭据。
- 生产环境不要提交 `.env`，只提交 `.env.example`。

## 项目结构

```text
.
├─ cmd/
│  ├─ server/            # HTTP API 入口
│  └─ cli/               # CLI 入口
├─ internal/             # 业务核心实现
├─ data/                 # 运行时状态文件（默认不提交）
├─ docs/
│  ├─ runbooks/          # 运维手册
│  ├─ reference/         # 外部参考资料
│  └─ archive/           # 历史记录归档（含 legacy-ts）
├─ docker-compose.yml
├─ go.mod
└─ .env.example
```

## 文档索引

- 结构说明：`docs/STRUCTURE.md`
- 运维手册：`docs/runbooks/index-sync-operations.md`
- 历史归档：`docs/archive/README.md`
