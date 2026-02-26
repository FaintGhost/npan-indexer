# Architecture

## Overview

本设计在不改后端契约的前提下，为搜索页增加“前端扩展名单选筛选 + URL 参数持久化”。

核心路径保持不变：

- 数据仍来自 `AppSearch`
- 分页与去重仍由 `mergePages` 负责
- 筛选仅发生在前端渲染层

```
SearchInput(query)
  -> activeQuery (debounce)
  -> AppSearch(query, page, pageSize)
  -> mergePages(dedupe by source_id)
  -> items
  -> filterByCategory(items, ext)
  -> filteredItems
  -> render list/status/count/empty
```

## Existing Baseline

- 搜索页入口：`web/src/routes/index.lazy.tsx`
- 合并去重：`mergePages`（按 `source_id`）
- 卡片展示：`web/src/components/file-card.tsx`
- 当前扩展名识别逻辑参考：`web/src/lib/file-icon.ts`

## Proposed File-Level Changes

| File | Change Type | Description |
|------|------------|-------------|
| `web/src/routes/index.lazy.tsx` | Modify | 新增筛选状态读取/写入（URL），新增 `filteredItems` 派生计算，替换列表和状态文案的数据源 |
| `web/src/lib/file-category.ts` | Add | 新增扩展名分类与匹配函数（含多段扩展名） |
| `web/src/components/search-page.test.tsx` | Modify | 新增 URL 初始化、切换筛选、过滤结果、非法参数回退等测试 |
| `web/src/tests/accessibility.test.tsx` | Modify | 新增筛选控件可访问性断言 |
| `web/src/lib/file-category.test.ts` | Add | 覆盖扩展名映射与边界条件 |

## URL Contract (Frontend)

- 参数：`ext`
- 枚举：`all | doc | image | video | archive | other`
- 默认：`all`
- 异常值：回退到 `all`

## Component/Data Responsibilities

### Search Page

- 负责 URL 参数解析和回写
- 负责从 `items` 计算 `filteredItems`
- 负责将筛选结果映射到：
  - 列表渲染
  - 计数条
  - 空态文案
  - 状态文案

### Category Utility

- 负责根据 `doc.name` 提取扩展名
- 负责分类判定
- 不依赖 `highlighted_name`，避免 HTML 高亮带来噪声

## State Model

- 保留现有 state：`query`、`activeQuery`
- 新增：`activeFilter`（来源于 URL）
- 不新增“过滤结果 state”，只保留派生值：`filteredItems`

## Interaction Sequence

### Initial Load

1. 读取 URL 的 `ext`
2. 校验并归一化到合法枚举
3. 根据 `activeFilter` 过滤 `items`

### User Changes Filter

1. 用户选择新分类
2. 更新 URL `ext`
3. 重新计算 `filteredItems`
4. 更新计数/列表/空态

### User Clears Search

1. 清空 `query` 与 `activeQuery`
2. 过滤回到默认逻辑（通常 `all`）
3. 可同步清理 URL 中与搜索相关参数

## Backward Compatibility

- 不影响后端 RPC 路径与请求结构
- URL 无 `ext` 时行为与当前一致
- `ext` 非法值自动兜底，不导致页面错误

## Risks

- 筛选后可见项下降导致触底更快，分页请求可能更频繁
- 分类规则边界（如 `tar.gz`、无扩展名）易出现误判

## Verification Matrix

| Concern | Verification |
|---------|--------------|
| URL 同步 | 刷新/后退前进状态一致 |
| 过滤准确 | 分类规则单测 + 页面行为测试 |
| 回归安全 | 保持 Connect 请求参数不变 |
| 可访问性 | radiogroup/radio 语义与键盘交互测试 |
