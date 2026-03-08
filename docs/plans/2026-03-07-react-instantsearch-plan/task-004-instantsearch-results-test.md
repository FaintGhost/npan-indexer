# Task 004: [TEST] InstantSearch 结果列表与高亮 (RED)

**depends-on**: task-002-search-bootstrap-fallback-impl.md

## Description

为 InstantSearch 结果列表、高亮与无限滚动补充失败测试，锁定前端已从旧 `AppSearch` 分页状态机切换到官方 InstantSearch hooks 驱动的结果渲染模型。该任务不实现生产代码。

## Execution Context

**Task Number**: 007 of 013
**Phase**: Search Migration
**Prerequisites**: `task-002-search-bootstrap-fallback-impl.md` 已完成，且搜索页已具备运行时配置引导与新旧链路分支

## BDD Scenario

```gherkin
Scenario: 输入关键字后触发直连 Meilisearch 搜索
  Given 搜索页已成功初始化 InstantSearch
  When 用户输入 "report" 并提交搜索
  Then 浏览器应向 Meilisearch 发起搜索请求
  And 结果列表应展示 hits
  And 状态文案应基于 InstantSearch 返回的结果数量更新

Scenario: InfiniteHits 驱动无限滚动
  Given 当前查询已有第一页结果且存在下一页
  When 用户滚动到结果列表底部
  Then 前端应通过 InfiniteHits 继续加载下一批 hits
  And 已加载结果应继续可见

Scenario: 命中结果名称显示高亮
  Given Meilisearch 返回带有 _formatted.name 的 hits
  When 结果卡片渲染文件名称
  Then 页面应展示高亮后的名称
  And 未命中高亮的结果应展示原始名称
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/components/file-card.test.tsx`
- Modify: `web/src/tests/mocks/handlers.ts`
- Create: `web/src/components/search-results.test.tsx`
- Create: `web/src/lib/meili-hit-adapter.test.ts`

## Steps

### Step 1: Verify Scenario

- 确认本任务聚焦“直连结果渲染 + InfiniteHits + 高亮”三类行为，不覆盖 routing/refinement 与下载链路。
- 明确失败应指向“结果仍由旧 `AppSearch` 状态机拥有”，而不是配置引导缺失。

### Step 2: Implement Test (Red)

- 使用 MSW 或模块替身模拟 Meilisearch hits、下一页 hits 与 `_formatted.name` 响应。
- 在搜索页测试中增加“启用 InstantSearch 后应消费 hits 而非 `AppSearch` 结果”的失败断言。
- 为结果列表与卡片增加 InfiniteHits 累积展示、高亮存在/缺失两条分支断言。

### Step 3: Verify Red Failure

- 运行目标前端测试并确认新增用例失败。
- 失败原因应明确指向“结果列表、高亮或 InfiniteHits 尚未接入 InstantSearch”。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/file-card.test.tsx src/components/search-results.test.tsx src/lib/meili-hit-adapter.test.ts
```

## Success Criteria

- 新增用例稳定失败（Red）。
- 使用 MSW / module mock 隔离网络与 search client。
- 失败信息明确指向结果渲染层尚未迁移到 InstantSearch。
