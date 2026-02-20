# 增量索引与恢复能力改进计划（2026-02-20）

## 背景结论

- 当前增量核心是“`LastSyncTime` 游标 + `doc_id` 幂等覆盖”，但尚未接入 CLI/API 执行入口。
- 当前全量恢复有 checkpoint 与 resume 能力，但中断后退出流程仍可优化（先等待取消落盘再退出）。
- 当前写入 Meilisearch 仅提交异步任务，不等待任务最终状态，存在“状态前进但异步任务失败”的一致性窗口。
- 新发现：增量查询存在两个生产缺陷
  - 增量窗口时间戳使用毫秒，但 OpenAPI `updated_time_range` 实际接受秒级时间戳，导致窗口过滤异常。
  - 增量查询词固定为 `*`，在当前平台检索语义下会稳定返回 0 结果，造成“成功但无变更”的假象。

## 本次目标

1. 增加 `sync-incremental` CLI 命令，打通增量索引链路。
2. 增加 Meilisearch 写入任务等待，确保提交成功后再继续流程。
3. 优化全量命令中断收尾，提升恢复状态可观测性。
4. 增加最小自动化测试，覆盖增量核心行为。
5. 修复增量窗口时间戳单位错误，并兼容历史毫秒游标状态。
6. 修复增量默认查询词，避免 `*` 导致的空结果误判。

## 范围

- 代码：`internal/indexer`、`internal/search`、`internal/cli`、`cmd/server`。
- 文档：`docs/plans/...` 进度记录。
- 不包含：调度系统（cron/worker）与 HTTP 增量接口。

## 验收标准

- `go run ./cmd/cli sync-incremental` 可执行并更新 `NPA_SYNC_STATE_FILE`。
- Meili 写入失败时明确报错，不推进增量游标。
- `sync-full` 收到中断信号后会等待任务取消并输出最终状态摘要。
- `go test ./...`、`go test -race ./...`、`go build ./...` 全通过。

## Design Documents

- [BDD Specifications](./bdd-specs.md)
- [Progress](./progress.md)
