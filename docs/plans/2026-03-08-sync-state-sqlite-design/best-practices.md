# Best Practices: 同步状态 SQLite 迁移

## 1. 驱动选型

优先使用 `modernc.org/sqlite`：

- 纯 Go 驱动，更适合当前 `CGO_ENABLED=0` 的 Docker 构建链路。
- 可通过 `database/sql` 统一管理连接与事务。
- 避免引入 `go-sqlite3` 的 CGO 依赖，减少构建和部署变量。

## 2. 保持“描述对象不变，存储介质替换”

本次迁移的重点是可靠性，而不是重构同步领域模型。

因此应优先：

- 保持 `models.SyncProgressState`、`models.SyncState`、`models.CrawlCheckpoint` 不变。
- 保持 Connect 响应与 CLI JSON 输出不变。
- 仅替换底层状态持久化实现与 wiring。

## 3. 单一状态源优先

一旦 SQLite 接入完成：

- 运行时读写应优先走 SQLite。
- JSON 文件只作为 legacy import source。
- 不要长期维持“JSON + SQLite 双写常态化”，否则可靠性与排障复杂度都会回升。

## 4. checkpoint key 逻辑保持兼容

当前代码和进度模型里已经把 checkpoint 表示为路径字符串：

- `SyncProgressState.CheckpointTemplate`
- `RootSyncProgress.CheckpointFile`

本轮不要试图删除这些字段。

推荐做法：

- 把它们视为 **checkpoint logical key**。
- SQLite 内部用该 key 存取 checkpoint payload。
- 对外继续显示原字符串，保证 CLI、前端、测试不受影响。

## 5. 迁移必须非破坏、可回看

首次切换到 SQLite 时：

- 不要自动删除旧 JSON。
- 不要自动 hard rename 覆盖旧文件。
- 迁移失败时应保留旧 JSON，以便人工排查或手工回滚。

如果后续要做清理，应作为单独任务和单独发布窗口处理。

## 6. 事务边界要围绕“单次写入完整对象”

虽然当前采用 JSON payload，而不是拆分关系表，但仍要确保：

- 单次 `Save()` 是原子的
- `Clear()` 不会留下半状态
- 任何读到的 payload 都应是可反序列化的完整 JSON

建议：

- 对 `UPSERT` / `DELETE` 使用显式事务或保证单语句原子性。
- 对高频写路径统一封装，不要在服务层拼 SQL。

## 7. 测试优先，且必须隔离外部依赖

### 存储层

- 使用临时目录中的 SQLite 文件。
- 不依赖真实网络、Meilisearch、Docker。

### 服务层

- 使用 stub API、stub index、fake limiter。
- 只验证同步状态迁移后的行为，不把失败归咎于外部 API。

### 接口层

- 使用已有 Connect test 模式。
- 验证协议语义不变，而不是只验证内部类型替换。

## 8. 验证顺序

建议严格按以下顺序收口：

1. `go test ./internal/storage ./internal/service ./internal/httpx ./internal/cli ./internal/config`
2. `GOCACHE=/tmp/go-build go test ./...`
3. `cd web && bun vitest run`
4. `docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120`
5. `./tests/smoke/smoke_test.sh`
6. `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright`
7. `docker compose -f docker-compose.ci.yml --profile e2e down --volumes`

## 9. 常见误区

### 误区 A：把 SQLite 当成“只是把 JSON 文件换个后缀”

如果仍然保留多处直接 `NewJSON...` 的 wiring，实际上并没有完成迁移。

### 误区 B：顺手重做进度模型

这会放大改动面，增加前端和接口回归风险。

### 误区 C：先实现再补测试

本项目对同步语义已有较多回归测试；本轮必须继续沿用 BDD / Red-Green，先锁行为，再切存储。

### 误区 D：为了省事引入 CGO 驱动

这会直接破坏当前 Docker 构建链路，与仓库现状冲突。
