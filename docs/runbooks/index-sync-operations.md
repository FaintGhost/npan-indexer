# Npan 索引同步运行手册

## 1. 前置条件

- Meilisearch 已启动（推荐 `docker compose up -d`）。
- 已配置云盘认证参数（`NPA_TOKEN` 或 `NPA_CLIENT_ID/NPA_CLIENT_SECRET/NPA_SUB_ID`）。
- 已配置索引参数（`MEILI_HOST`、`MEILI_API_KEY`、`MEILI_INDEX`）。
- 已确认同步状态库路径（默认 `NPA_STATE_DB_FILE=./data/state/sync-state.sqlite`）。

## 1.1 状态持久化说明

当前运行时以 SQLite 作为同步状态主存储，统一保存：
- `progress`：全量/当前同步进度
- `sync_state`：增量游标（`lastSyncTime` 等）
- `checkpoint`：全量 crawl 断点

兼容策略：
- `NPA_PROGRESS_FILE` 与 `NPA_SYNC_STATE_FILE` 仍保留，用于首次读取时从 legacy JSON 惰性导入 SQLite。
- 导入后，运行时主读写路径仍是 `NPA_STATE_DB_FILE` 指向的 SQLite 文件。
- 本轮迁移不会自动删除旧 JSON；如需排障，可保留旧文件做人工对照。

## 2. 首次全量同步

1. 先执行测试与构建检查：
  - `go test ./...`
  - `go build ./...`
2. 执行全量任务（CLI 方式）：
  - `go run ./cmd/cli sync --mode full --token <token> --root-folder-ids 0`
  - 交互式排障建议使用默认人类可读进度：`--progress-output human`
  - 如需机器采集进度日志可切换：`--progress-output json`
  - 人类可读进度会附带估算进度字段：
    - `est=xx.x%(docs=a/b roots=c/d)`：按已知根目录 `estimatedTotalDocs` 估算，`a` 为已处理文档数（`files+folders`），`b` 为估算总文档数。
    - `est=n/a`：当前根目录无法可靠获取总量，不显示百分比（不影响同步本身）。
3. 或服务方式触发：
  - `POST /npan.v1.AdminService/StartSync` (body: `{"mode": "SYNC_MODE_FULL"}`)
  - 查看进度：`POST /npan.v1.AdminService/GetSyncProgress` (body: `{}`)
  - 取消同步：`POST /npan.v1.AdminService/CancelSync` (body: `{}`)
3. 同步完成后确认指标：
  - 遍历目录数量
  - 索引文档数量
  - 失败请求数量

## 3. 增量同步调度

- 建议每 5~15 分钟执行一次增量同步。
- 每次执行读取 `lastSyncTime`，仅处理变更文档。
- `lastSyncTime` 使用秒级时间戳；若历史状态为毫秒值，程序会自动兼容迁移。
- 同步成功后推进游标；失败时保留旧游标。
- CLI 示例：
  - `go run ./cmd/cli sync --mode incremental --incremental-query-words "* OR *" --window-overlap-ms 2000`

## 4. 检索与下载

- 搜索：`go run ./cmd/cli search-local --query "关键词"`
- 获取下载链接：`go run ./cmd/cli download-url --file-id 123 --token <token>`
- 下载链接为临时 URL，不要持久化到索引。

## 5. 告警建议

- 429 比例 > 5%（5 分钟窗口）告警。
- 同步任务连续失败 3 次告警。
- checkpoint 长时间不推进（例如 > 30 分钟）告警。

## 6. 故障恢复

1. 检查 Meili 健康：`curl $MEILI_HOST/health`
2. 检查云盘 token 是否过期。
3. 检查 SQLite 状态库是否存在且可更新：
  - 路径默认是 `./data/state/sync-state.sqlite`
  - 也可通过 `NPA_STATE_DB_FILE` 覆盖
4. 若需要人工确认当前进度，执行：
  - `go run ./cmd/cli sync-progress --state-db-file ./data/state/sync-state.sqlite`
5. 若增量游标异常：优先检查 SQLite 中的 `sync_state` 是否符合预期；必要时可用保留的 legacy JSON 做对照。
6. 若索引污染严重：清空目标索引后重跑全量。
7. 若怀疑是迁移问题：
  - 确认 `NPA_PROGRESS_FILE` / `NPA_SYNC_STATE_FILE` 仍指向原 JSON 文件。
  - 保留旧 JSON，不要先删除；程序会在 SQLite 缺失对应记录时惰性导入。
