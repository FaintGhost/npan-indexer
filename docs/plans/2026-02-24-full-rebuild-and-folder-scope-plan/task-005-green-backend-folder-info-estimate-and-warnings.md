# Task 005: GREEN backend folder info estimate and warnings

**depends-on**: task-004-red-backend-folder-info-estimate-tests.md

## Scenario Reference

- Feature: Official Folder Stats Estimate
- Scenario: 显式 root 使用 folder info 填充估计值
- Scenario: folder info 失败不阻断同步（降级）
- Feature: Completion Warning For Estimate Mismatch
- Scenario: 实际写入显著小于官方估计时生成警告
- Scenario: 差异在阈值内不告警

## Objective

接入 `folder info` 并把估计值与完成后差异告警串到现有同步链路。

## Files

- `internal/npan/types.go`
- `internal/npan/client.go`
- `internal/service/sync_manager.go`

## Tasks

1. 在 `npan.API` 增加 `GetFolderInfo` 能力并完成 HTTP 客户端实现。
2. 在 root 发现逻辑中，对显式 `root_folder_ids` 拉取 folder info 回填名称和估计值。
3. 在同步完成 verification 阶段增加 root 级估计差异告警。
4. 对 folder info 失败场景做降级（记录告警/日志，不中断同步）。

## Verification

- `go test ./internal/npan -run FolderInfo -v`
- `go test ./internal/service -run Estimate -v`
- `go test ./internal/service -run Verification -v`
