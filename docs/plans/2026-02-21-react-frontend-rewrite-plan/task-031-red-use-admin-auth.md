# Task 031: 测试 useAdminAuth Hook（localStorage + 验证 + 401 拦截）

**depends-on**: task-004

## Description

为 useAdminAuth 自定义 Hook 创建失败测试用例。

## Execution Context

**Task Number**: 031 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 5 - 首次访问管理页且无存储的 Key 时弹出输入对话框; 输入有效 API Key 后存储并关闭对话框; 已存储的 API Key 自动使用; 输入无效 API Key 显示错误提示; 已存储的 Key 失效时重新提示输入; 空输入提交不发送验证请求

## Files to Modify/Create

- Create: `cli/src/hooks/use-admin-auth.test.ts`

## Steps

### Step 1: Test no stored key → needsAuth is true

### Step 2: Test stored key → needsAuth is false, apiKey available

### Step 3: Test validate valid key → stores to localStorage, needsAuth becomes false

### Step 4: Test validate invalid key (401) → does not store, returns error

### Step 5: Test empty key → validation error, no request sent

### Step 6: Test on401 → clears localStorage, needsAuth becomes true

### Step 7: Test getHeaders returns X-API-Key header

### Step 8: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-admin-auth.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖 API Key 完整生命周期
