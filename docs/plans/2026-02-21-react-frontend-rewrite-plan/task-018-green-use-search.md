# Task 018: 实现 useSearch Hook

**depends-on**: task-017

## Description

实现 useSearch 自定义 Hook，使 Task 017 测试通过。封装搜索状态管理、debounce、分页、请求竞态处理。

## Execution Context

**Task Number**: 018 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 017 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 搜索相关所有场景

## Files to Modify/Create

- Create: `cli/src/hooks/use-search.ts`

## Steps

### Step 1: Implement useSearch Hook

- 状态：query, items, total, page, loading, hasMore, error
- setQuery(q): 设置 query + 启动 debounce 定时器（280ms）
- searchImmediate(q): 取消 debounce + 立即搜索
- loadMore(): 加载下一页（追加模式）
- reset(): 清空所有状态
- 内部使用 AbortController 管理请求生命周期
- 内部使用 requestSeq 序列号防止过时响应
- 去重：基于 source_id 的 Set
- 使用 fetchAPI 发送请求

### Step 2: Cleanup

- useEffect cleanup 中 abort 进行中的请求
- clearTimeout 清理 debounce 定时器

### Step 3: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-search.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 017 所有测试通过
- Hook 在组件卸载时正确清理资源
