# Task 004: RED backend folder info estimate tests

**depends-on**: task-001-openapi-contract-audit.md

## Scenario Reference

- Feature: Official Folder Stats Estimate
- Scenario: 显式 root 使用 folder info 填充估计值
- Scenario: folder info 失败不阻断同步（降级）
- Feature: Completion Warning For Estimate Mismatch
- Scenario: 实际写入显著小于官方估计时生成警告

## Objective

先用失败测试固定“显式 root 估计来源”和“差异告警”行为。

## Files

- `internal/service/sync_manager_estimate_test.go`
- `internal/service/sync_manager_verification_test.go`

## Tasks

1. 增加 `discoverRootFolders` 场景测试：显式 root 调用 folder info 并回填 `RootNames` 与 `estimatedTotalDocs`。
2. 增加降级测试：folder info 失败时不阻断同步流程。
3. 增加 verification 告警测试：估计值与实际值差异超阈值时写入 warning。

## Verification

- `go test ./internal/service -run Estimate -v`
- `go test ./internal/service -run Verification -v`
- 预期：新增测试在实现前失败（RED）
