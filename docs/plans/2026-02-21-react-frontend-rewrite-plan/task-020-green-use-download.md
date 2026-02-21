# Task 020: 实现 useDownload Hook

**depends-on**: task-019

## Description

实现 useDownload 自定义 Hook，使 Task 019 测试通过。

## Execution Context

**Task Number**: 020 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 019 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 2 - 下载功能所有场景

## Files to Modify/Create

- Create: `cli/src/hooks/use-download.ts`

## Steps

### Step 1: Implement useDownload Hook

- 内部维护 Map<number, string> 缓存
- download(fileId): 检查缓存 → 未命中则 fetchAPI → 缓存结果
- 状态管理：Map<fileId, "idle" | "loading" | "success" | "error">
- 成功后调用 window.open(url, '_blank', 'noopener,noreferrer')
- 成功后 1.5 秒 setTimeout 恢复 idle
- 空 download_url 视为错误

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-download.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 019 所有测试通过
