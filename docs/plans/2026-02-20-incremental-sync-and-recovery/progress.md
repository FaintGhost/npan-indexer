# 进度记录

## 2026-02-20

- 已完成：计划与 BDD 规格落盘。
- 已完成：新增 `sync-incremental` 命令，接入增量同步链路。
- 已完成：新增增量变更抓取实现（按更新时间窗口分页拉取、去重、upsert/delete 拆分）。
- 已完成：Meilisearch 写入改为等待任务结束（settings/upsert/delete），失败不推进后续状态。
- 已完成：`sync-full` 在中断信号下等待同步协程收尾后退出，并输出最终状态摘要。
- 已完成：新增测试
  - `internal/indexer/incremental_fetch_test.go`
  - `internal/indexer/incremental_sync_test.go`

## 2026-02-21

- 已定位：`sync-incremental` 出现“执行成功但 0 变更”问题的两个根因
  - 增量窗口时间戳单位错配（毫秒传给秒级 `updated_time_range`）。
  - 增量查询词固定 `*` 在当前平台语义下返回空集合。
- 已完成：增量游标统一为秒级语义，并兼容历史毫秒游标自动迁移。
- 已完成：增量查询词改为可配置（`NPA_INCREMENTAL_QUERY_WORDS` / `--incremental-query-words`），默认值调整为 `* OR *`。
- 已完成：新增 `internal/npan/client_search_updated_test.go`，覆盖增量窗口请求参数与默认查询词。
- 已完成：补充 `internal/indexer/incremental_sync_test.go`，覆盖毫秒游标兼容与秒级写回行为。
- 已完成：命令级验证，历史毫秒状态下 `since_used` 已按秒输出。

## 验证结果

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./...` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test -race ./...` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go run ./cmd/cli sync-incremental --sync-state-file /tmp/npan-incremental-state-test.json --window-overlap-ms 2000 --incremental-query-words '* OR *'` 通过。
