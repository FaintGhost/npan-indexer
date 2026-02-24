# Task 003: GREEN backend checkpoint reset implementation

**depends-on**: task-002-red-backend-checkpoint-reset-tests.md

## Scenario Reference

- Feature: Force Rebuild Must Reset Crawl Checkpoints
- Scenario: 强制重建忽略残留 checkpoint
- Scenario: 非强制重建且 resume=false 也应重置 checkpoint
- Scenario: resume=true 保留当前断点行为

## Objective

在全量路径中按 root 清理 checkpoint，使强制重建和明确禁用续跑的行为与语义一致。

## Files

- `internal/service/sync_manager.go`

## Tasks

1. 在全量路径计算出 root checkpoint 映射后，加入 checkpoint 重置步骤。
2. 将重置条件与 `force_rebuild` / `resume_progress=false` 对齐。
3. 保持 `resume_progress=true` 的现有行为不变。
4. 确保错误路径返回清晰错误，不进行半重置运行。

## Verification

- `go test ./internal/service -run Checkpoint -v`
- `go test ./internal/service -run CursorUpdate -v`
