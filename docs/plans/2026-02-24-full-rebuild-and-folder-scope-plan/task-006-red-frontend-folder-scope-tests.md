# Task 006: RED frontend folder-scope payload and UI tests

**depends-on**: task-001-openapi-contract-audit.md

## Scenario Reference

- Feature: Folder-Scoped Full Sync From Admin UI
- Scenario: 指定单个目录 ID 启动范围索引
- Scenario: 输入多个目录 ID（逗号分隔）
- Scenario: 空输入保持现有全库行为

## Objective

先补前端失败测试，锁定目录范围输入解析与请求体字段行为。

## Files

- `web/src/hooks/use-sync-progress.test.ts`
- `web/src/components/admin-page.test.tsx`

## Tasks

1. 为 `useSyncProgress.startSync` 增加测试：指定目录时请求体包含 `root_folder_ids` 与 `include_departments=false`。
2. 为 Admin 页面增加测试：目录 ID 输入解析（单个、多个、空值）。
3. 为非法输入（非数字、空 token）增加错误反馈或禁用行为测试（按现有测试风格）。

## Verification

- `cd web && bun vitest run src/hooks/use-sync-progress.test.ts`
- `cd web && bun vitest run src/components/admin-page.test.tsx`
- 预期：新增测试在实现前失败（RED）
