# Task 032: 实现 useAdminAuth Hook

**depends-on**: task-031

## Description

实现 useAdminAuth Hook，使 Task 031 测试通过。

## Execution Context

**Task Number**: 032 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 031 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 5 - 管理页认证所有场景

## Files to Modify/Create

- Create: `cli/src/hooks/use-admin-auth.ts`

## Steps

### Step 1: Implement useAdminAuth

- 从 localStorage ("npan_admin_api_key") 读取已存储的 key
- needsAuth: boolean — 是否需要用户输入 key
- validate(key): 发送测试请求验证 key → 成功则存储到 localStorage
- on401(): 清除 localStorage，设置 needsAuth=true
- getHeaders(): 返回 { "X-API-Key": storedKey }
- localStorage key 名为 "npan_admin_api_key"

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-admin-auth.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 031 所有测试通过
