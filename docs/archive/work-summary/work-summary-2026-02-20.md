# Npan 外部索引项目工作总结（2026-02-20）

## 本次目标

- 基于 Npan OpenAPI 构建独立搜索索引链路，绕过平台弱搜索。
- 支持全量遍历、增量同步、Meilisearch 检索、按需下载 URL 获取。
- 支持长任务可恢复执行，并可观察同步进度。

## 已完成内容

- 认证与基础能力
- `get-token.ts` 与 `search-file.ts` 支持 `.env` 回退。
- 新增 `src/npan/auth.ts`，支持 `token` 或 `client_id/client_secret/sub_id` 换取 token。

- 索引核心
- 新增全量遍历与限流：`src/indexer/full-crawl.ts`、`src/indexer/rate-limiter.ts`。
- 新增重试策略：`src/indexer/retry-policy.ts`（429/5xx 指数退避）。
- 新增 checkpoint：`src/indexer/checkpoint-store.ts`（中断后续跑）。
- 新增增量同步：`src/indexer/incremental-sync.ts`、`src/indexer/sync-state-store.ts`。

- Meilisearch 集成
- 新增映射层：`src/search/meili-mapper.ts`。
- 新增索引操作：`src/search/meili-index.ts`。
- 新增查询服务：`src/search/query-service.ts`。

- 下载代理
- 新增 `src/download/download-url-service.ts`。
- 支持常见错误映射（404/403/429）。

- CLI 工具
- `src/cli/sync-full-index.ts`：全量同步入口。
- `src/cli/search-local-index.ts`：索引检索。
- `src/cli/get-download-url.ts`：按 `file_id` 拉实时下载链接。
- `src/cli/sync-progress.ts`：查看同步进度。

- 发现逻辑增强
- 修复“只扫个人空间”的问题。
- 新增部门根目录发现：
- `GET /api/v2/user/departments`
- `GET /api/v2/folder/department_folders?department_id=...`
- 默认同时扫描个人根和部门根。

- 进度与恢复
- 新增进度文件：`./data/progress/full-sync-progress.json`。
- 记录根目录状态、累计统计、当前目录/页、最近错误。
- 默认 `--resume-progress=true`，已完成根目录自动跳过。
- 修复续跑统计回退问题（`files/pages/folders` 只增不减）。

- 运维与文档
- 新增压测/自检脚本：`scripts/run-load-check.sh`。
- 新增运行手册：`docs/runbooks/index-sync-operations.md`。
- 更新 `.env.example`：补充 Meili、进度、续跑相关变量。

## 测试与验证结果

- 单元与集成测试
- 命令：`bun test`
- 结果：`23 pass, 0 fail`

- 负载检查脚本
- 命令：`bash scripts/run-load-check.sh`
- 结果：通过。
- 说明：若宿主机未映射 7700 端口，脚本会 fallback 容器 IP 健康检查。

## 关键问题与处理记录

- 问题 1：全量结果只有约 189 文件。
- 原因：仅扫描了个人空间 root=0。
- 处理：增加部门根目录发现并纳入扫描。

- 问题 2：长跑后 token 失效。
- 现状：手动重启可从 checkpoint 续跑。
- 备注：后续可加“401 自动刷新 token”能力。

- 问题 3：progress 统计回退（例如 2036 -> 194）。
- 原因：续跑时覆盖了累计统计。
- 处理：改为 `base + delta` 累加并加单调保护。

## 当前使用方式

- 启动全量同步（建议持续日志）
- `bun run src/cli/sync-full-index.ts --progress-every 1`

- 查看当前进度
- `bun run src/cli/sync-progress.ts`

- 检索索引
- `bun run src/cli/search-local-index.ts --query "关键词"`

- 获取下载链接
- `bun run src/cli/get-download-url.ts --file-id 123 --token <token>`

## 建议的下一步

- 增加 401 自动刷新 token（不中断长任务）。
- 在进度中增加“预估总量”字段（按根目录 `item_count`）与更直观百分比。
- 增加增量调度 CLI（定时任务入口）。
- 若需宿主机直连 Meili，补充 compose 端口映射 `7700:7700`。

## 当日追加记录（Go 重写）

- 已新增 Go 重写版本，Web 框架使用 `github.com/labstack/echo/v5@v5.0.4`。
- 已完成模块分层：
  - `internal/npan`：认证与 OpenAPI 客户端。
  - `internal/search`：Meilisearch 设置、写入、查询。
  - `internal/indexer`：全量遍历、重试、限流、增量同步。
  - `internal/storage`：checkpoint/progress/sync-state JSON 持久化。
  - `internal/service`：下载代理与同步任务调度。
  - `internal/httpx`：Echo 路由与 HTTP handler。
- 已实现 goroutine 并发：
  - 全量同步后台启动采用 goroutine。
  - 多根目录并行同步采用 goroutine + semaphore 控制并发。
- 已提供 API：
  - `GET /healthz`
  - `POST /api/v1/token`
  - `GET /api/v1/npan/search`
  - `GET /api/v1/search/local`
  - `GET /api/v1/download-url`
  - `POST /api/v1/sync/full/start`
  - `GET /api/v1/sync/full/progress`
  - `POST /api/v1/sync/full/cancel`
- 验证结果：
  - `go test ./...` 通过
  - `go build ./...` 通过

## 当日追加记录（生产级审查修复）

- 安全修复
  - 新增 `NPA_ADMIN_API_KEY`，支持对 `/api/v1/*` 开启 API Key 访问保护（`X-API-Key`）。
  - 新增 `NPA_ALLOW_CONFIG_AUTH_FALLBACK`（默认 `false`），默认禁止 HTTP 请求回退服务端配置凭据。
- 稳定性修复
  - 多 root 同步改为共享全局 limiter，避免总并发/QPS 随 root 数放大。
  - 任一 root 失败后立即取消其余 goroutine，减少无效上游流量。
  - checkpoint/progress/sync-state 改为原子写（临时文件 + fsync + rename），降低状态损坏风险。
- 回归测试补充
  - 新增 `internal/httpx/handlers_test.go`（鉴权与凭据回退行为）。
  - 新增 `internal/storage/json_store_test.go`（原子写与持久化加载）。
- 验证结果
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./...` 通过。
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test -race ./...` 通过。
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...` 通过。
