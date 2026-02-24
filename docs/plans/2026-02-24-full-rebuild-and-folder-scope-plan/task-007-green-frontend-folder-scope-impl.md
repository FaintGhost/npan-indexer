# Task 007: GREEN frontend folder-scope payload and UI

**depends-on**: task-006-red-frontend-folder-scope-tests.md

## Scenario Reference

- Feature: Folder-Scoped Full Sync From Admin UI
- Scenario: 指定单个目录 ID 启动范围索引
- Scenario: 输入多个目录 ID（逗号分隔）
- Scenario: 空输入保持现有全库行为

## Objective

实现 Admin 页面目录范围输入，并确保请求体按 OpenAPI 契约发送。

## Files

- `web/src/components/admin-sync-page.tsx`
- `web/src/hooks/use-sync-progress.ts`

## Tasks

1. 在 Admin 页新增目录 ID 输入控件（支持逗号分隔）。
2. 将解析后的目录 ID 数组传入 `startSync`。
3. 当目录范围非空时，发送 `include_departments=false`；空范围时不覆盖现有默认行为。
4. 保持现有 mode / forceRebuild 交互逻辑与兼容性。

## Verification

- `cd web && bun vitest run src/hooks/use-sync-progress.test.ts`
- `cd web && bun vitest run src/components/admin-page.test.tsx`
