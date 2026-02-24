# Full Rebuild And Folder Scope Plan

## Goal

实现方案 B，并确保前后端通过 `api/openapi.yaml` 约束保持一致：

- 修复 `force_rebuild` 未清理 checkpoint 导致“伪全量”问题
- 增加 Admin 单目录/多目录范围索引能力
- 使用官方 `folder info` 填充 root 级统计预估并在完成后给出差异告警

## Constraints

- OpenAPI-first：任何新增/变更 API 字段必须先检查并更新 `api/openapi.yaml`
- 先写测试（RED），再实现（GREEN）
- 优先最小改动，复用现有 `SyncStartRequest` / `SyncProgressState` 契约
- 不引入破坏性 API 改动

## Execution Plan

- [Task 001: OpenAPI contract audit and generation guard](./task-001-openapi-contract-audit.md)
- [Task 002: RED backend checkpoint reset tests](./task-002-red-backend-checkpoint-reset-tests.md)
- [Task 003: GREEN backend checkpoint reset implementation](./task-003-green-backend-checkpoint-reset-impl.md)
- [Task 004: RED backend folder info estimate tests](./task-004-red-backend-folder-info-estimate-tests.md)
- [Task 005: GREEN backend folder info estimate and warnings](./task-005-green-backend-folder-info-estimate-and-warnings.md)
- [Task 006: RED frontend folder-scope payload and UI tests](./task-006-red-frontend-folder-scope-tests.md)
- [Task 007: GREEN frontend folder-scope payload and UI](./task-007-green-frontend-folder-scope-impl.md)
- [Task 008: RED frontend estimate warning display tests](./task-008-red-frontend-estimate-warning-tests.md)
- [Task 009: GREEN frontend estimate warning display](./task-009-green-frontend-estimate-warning-impl.md)
- [Task 010: Verification and regression checks](./task-010-verification.md)

## Batching Notes (for execution)

- Batch A: Task 001 + Task 002（先契约/后端红测）
- Batch B: Task 003 + Task 004（checkpoint 修复 + 估计红测）
- Batch C: Task 005 + Task 006（后端估计实现 + 前端红测）
- Batch D: Task 007 + Task 008（前端目录范围实现 + 告警红测）
- Batch E: Task 009 + Task 010（前端告警实现 + 全量验证）
