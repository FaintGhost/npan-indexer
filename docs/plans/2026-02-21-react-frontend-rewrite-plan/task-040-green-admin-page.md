# Task 040: 实现管理页面

**depends-on**: task-039

## Description

实现管理页面组件（`/app/admin` 路由），组合 Admin 相关 hooks 和组件，使 Task 039 测试通过。

## Execution Context

**Task Number**: 040 of 046
**Phase**: Integration (Green)
**Prerequisites**: Task 039 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 5 + Feature 6 - 管理页认证与同步管理所有场景

## Files to Modify/Create

- Modify: `cli/src/routes/admin.tsx` — 实现 AdminPage 组件

## Steps

### Step 1: Compose AdminPage

- 使用 useAdminAuth 管理认证状态
- 使用 useSyncProgress 管理同步状态
- 条件渲染: needsAuth → ApiKeyDialog，否则 → 管理面板

### Step 2: Admin Panel layout

- "启动全量同步"按钮（同步运行中时 disabled）
- "取消同步"按钮（仅同步运行中时显示，点击弹确认对话框）
- SyncProgress 组件展示进度
- 返回搜索页链接

### Step 3: Wire up 401 interception

- API 请求返回 401 → 调用 on401 → 显示 ApiKeyDialog

### Step 4: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/routes/admin.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 039 所有集成测试通过
