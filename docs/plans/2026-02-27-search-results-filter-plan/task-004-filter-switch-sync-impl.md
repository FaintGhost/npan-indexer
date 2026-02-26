# Task 004: [IMPL] 筛选切换 URL 同步与请求契约不变 (GREEN)

**depends-on**: task-004-filter-switch-sync-test.md

## Description

实现筛选切换的 URL 回写逻辑，并保持 `AppSearch` 请求体仅包含原有查询字段。

## Execution Context

**Task Number**: 008 of 012  
**Phase**: Integration  
**Prerequisites**: `task-004-filter-switch-sync-test.md` 已 Red

## BDD Scenario

```gherkin
Scenario: 用户切换筛选会同步更新 URL
  Given 当前 URL 为 /?q=test
  And 当前筛选为 all
  When 用户点击图片筛选
  Then URL 应更新为包含 ext=image
  And 列表应仅展示图片类结果

Scenario: 选择 all 时可移除 ext 参数
  Given 当前 URL 为 /?q=test&ext=video
  When 用户切换到全部筛选
  Then URL 中 ext 参数应被移除或归一化为默认
  And 列表恢复展示全部结果

Scenario: 切换筛选不改变后端请求契约
  Given 用户已发起搜索并拿到结果
  When 用户在文档和压缩包筛选间切换
  Then 不应引入新的后端查询字段
  And Connect 请求路径仍为 /npan.v1.AppService/AppSearch
```

**Spec Source**: `../2026-02-27-search-results-filter-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/routes/index.lazy.tsx`

## Steps

### Step 1: Implement Logic (Green)
- 在筛选切换事件中更新 URL 参数。
- 选择 `all` 时执行 URL 归一化处理。
- 明确保持 `useInfiniteQuery(appSearch)` 请求参数结构不变。

### Step 2: Verify Green
- 运行 task-004 测试并确保通过。

### Step 3: Regression Check
- 复跑搜索页测试，确认无回归。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- URL 与筛选状态保持双向一致。
- Connect 请求契约完全不变。
