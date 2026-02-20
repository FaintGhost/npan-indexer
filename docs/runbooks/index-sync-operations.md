# Npan 索引同步运行手册

## 1. 前置条件

- Meilisearch 已启动（推荐 `docker compose up -d`）。
- 已配置云盘认证参数（`NPA_TOKEN` 或 `NPA_CLIENT_ID/NPA_CLIENT_SECRET/NPA_SUB_ID`）。
- 已配置索引参数（`MEILI_HOST`、`MEILI_API_KEY`、`MEILI_INDEX`）。

## 2. 首次全量同步

1. 先执行测试与构建检查：
  - `go test ./...`
  - `go build ./...`
2. 执行全量任务（CLI 方式）：
  - `go run ./cmd/cli sync-full --token <token> --root-folder-ids 0`
  - 交互式排障建议使用默认人类可读进度：`--progress-output human`
  - 如需机器采集进度日志可切换：`--progress-output json`
3. 或服务方式触发：
  - `POST /api/v1/sync/full/start`
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
  - `go run ./cmd/cli sync-incremental --incremental-query-words "* OR *" --window-overlap-ms 2000`

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
3. 若增量游标损坏：回退到最近有效游标后重跑增量。
4. 若索引污染严重：清空目标索引后重跑全量。
