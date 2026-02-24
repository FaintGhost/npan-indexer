# Task 002: RED backend checkpoint reset tests

**depends-on**: task-001-openapi-contract-audit.md

## Scenario Reference

- Feature: Force Rebuild Must Reset Crawl Checkpoints
- Scenario: 强制重建忽略残留 checkpoint
- Scenario: 非强制重建且 resume=false 也应重置 checkpoint
- Scenario: resume=true 保留当前断点行为

## Objective

先补失败测试，锁定 checkpoint 重置语义，避免修复回归。

## Files

- `internal/service/sync_manager_routing_test.go`

## Tasks

1. 增加测试：预置 checkpoint 文件 + `force_rebuild=true`，断言运行时不沿用旧断点。
2. 增加测试：`resume_progress=false` 时也应清 checkpoint。
3. 增加测试：`resume_progress=true` 保持现有续跑行为（不清理）。

## Verification

- `go test ./internal/service -run Checkpoint -v`
- 预期：新增测试在实现前失败（RED）
