# Task 004: [TEST] 筛选切换 URL 同步与请求契约不变 (RED)

**depends-on**: task-001-url-filter-state-impl.md

## Description

补充筛选切换行为测试，验证 URL 同步更新，以及切换筛选时不会改变 Connect 请求契约。

## Execution Context

**Task Number**: 007 of 012  
**Phase**: Integration  
**Prerequisites**: URL 初始化能力已具备

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

- Modify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Verify Scenario
- 确认 URL 同步和请求契约场景在 BDD 中定义完整。

### Step 2: Implement Test (Red)
- 新增筛选切换后的 URL 断言。
- 新增请求体断言，确保切换筛选不引入新字段。
- 使用 MSW 记录请求，作为网络替身。

### Step 3: Verify Red Failure
- 执行目标测试并确认失败。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- URL 同步与契约不变测试处于 Red。
- 失败信息清晰指向行为差异。
