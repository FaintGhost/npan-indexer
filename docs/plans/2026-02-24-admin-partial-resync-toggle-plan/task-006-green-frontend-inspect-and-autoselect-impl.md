# Task 006: GREEN frontend inspect decoupling and auto-select implementation

**depends-on**: task-005-red-frontend-inspect-and-autoselect-tests.md

## Description

实现目录详情拉取交互与默认自动勾选逻辑，将目录输入从“同步范围”改为“目录详情拉取”用途。

## Execution Context

**Task Number**: 006 of 011  
**Phase**: Frontend (Green)  
**Prerequisites**: Task 005 红测已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `拉取目录详情只更新列表，不启动同步`、`批量拉取目录详情部分成功`

## Files to Modify/Create

- Modify: `web/src/components/admin-sync-page.tsx`
- Modify: `web/src/hooks/use-sync-progress.ts`
- Modify: `web/src/lib/sync-schemas.ts`（若新增 progress 字段或 inspect 响应 schema）
- Modify: `web/src/components/sync-progress-display.tsx`（新增 toggle 支撑时）

## Steps

### Step 1: Implement Inspect Action

- 在 hook 层新增 inspect roots 请求能力。
- 在页面新增“拉取目录详情”按钮与独立 loading/error 状态。

### Step 2: Implement Auto-Select

- inspect 成功项自动加入已勾选集合。
- 保留用户已存在勾选，不做覆盖式重置。

### Step 3: Keep Sync Trigger Separate

- “启动同步”只基于当前勾选集合提交，不读取输入框原始字符串。

### Step 4: Verify (Green)

- 运行 Task 005 测试转绿。

## Verification Commands

```bash
cd web && bun vitest run src/components/admin-page.test.tsx src/hooks/use-sync-progress.test.ts
```

## Success Criteria

- Task 005 红测转绿。
- 目录拉取与同步动作完全解耦。
- 新拉取目录默认自动勾选（已按用户确认生效）。
