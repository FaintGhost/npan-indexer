# Task 039: 测试管理页面完整流程

**depends-on**: task-032, task-034, task-036, task-038

## Description

为管理页面的完整用户流程创建集成测试。

## Execution Context

**Task Number**: 039 of 046
**Phase**: Integration (Red)
**Prerequisites**: 所有 Admin 相关 hooks 和组件已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 5 - 管理页认证所有场景; Feature 6 - 同步管理所有场景

## Files to Modify/Create

- Create: `cli/src/routes/admin.test.tsx`

## Steps

### Step 1: Test no API key → shows dialog

### Step 2: Test valid API key → dialog closes, admin panel shows

### Step 3: Test start sync → success message + polling starts

### Step 4: Test progress display during sync

### Step 5: Test cancel sync → confirmation → cancel request

### Step 6: Test 401 during operation → dialog reappears

### Step 7: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/routes/admin.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 集成测试覆盖管理页面完整流程
