# Task 017: 测试 useSearch Hook（debounce、分页、竞态处理）

**depends-on**: task-010, task-004

## Description

为 useSearch 自定义 Hook 创建失败测试用例。覆盖 debounce、分页、请求竞态、去重等核心搜索逻辑。

## Execution Context

**Task Number**: 017 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 010 API 客户端已实现，Task 004 MSW 已配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 输入关键词后自动搜索并显示结果; 快速连续输入只触发一次请求; 新搜索取消进行中的旧请求; 滚动到底部自动加载下一页; 所有结果已加载时不再请求; 加载中不重复请求; 去重——同一文件不重复显示; 搜索框按 Enter 立即触发搜索

## Files to Modify/Create

- Create: `cli/src/hooks/use-search.test.ts`

## Steps

### Step 1: Test search triggers after debounce

- 调用 setQuery("MX40") → 280ms 后应发送搜索请求
- 返回 items 和 total

### Step 2: Test debounce coalesces rapid inputs

- 快速连续调用 setQuery("M"), setQuery("MX"), setQuery("MX40")
- 只应发送 1 次请求，query 为 "MX40"

### Step 3: Test immediate search (Enter key bypass)

- 调用 searchImmediate("固件") → 立即发送请求
- 取消待执行的 debounce 定时器

### Step 4: Test pagination (loadMore)

- 首次搜索返回 24 条且 total=50 → hasMore 为 true
- 调用 loadMore → 发送 page=2 请求
- 新结果追加到已有列表

### Step 5: Test no more pages

- items.length >= total → hasMore 为 false
- loadMore 不发送请求

### Step 6: Test deduplication

- 页 1 返回 id=[1,2,3]，页 2 返回 id=[3,4,5]
- 合并后去重为 5 条

### Step 7: Test request cancellation

- 搜索 A 进行中 → setQuery("B") → A 被 abort

### Step 8: Test reset

- 调用 reset → items 清空，total 归零

### Step 9: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-search.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖所有搜索核心行为
