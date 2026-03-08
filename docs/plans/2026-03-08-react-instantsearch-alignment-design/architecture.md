# Architecture: React InstantSearch 纠偏

## Scope

本轮不是重做 public 搜索架构，而是在现有 React InstantSearch 基础上做“官方行为 + 关键对齐”纠偏。

保留不变的部分：

- `GetSearchConfig` 运行时配置引导
- `InstantSearch` 作为 public 搜索状态拥有者
- `file_category` refinement 与 URL routing
- `useInfiniteHits` 结果列表与高亮壳层
- 下载继续通过 `AppService.AppDownloadURL`
- `instantsearchEnabled` 作为新旧链路切换开关

本轮新增/修正的架构能力：

1. 输入即搜（search-as-you-type）
2. query preprocess adapter
3. public 默认过滤基线
4. 结果对比与发布/回滚门槛

## Current Root Causes

### 1. 输入框不是 search-as-you-type

当前 public 分支中：

- `handleChange()` 只更新 `inputValue`
- `refine()` 仅在 `handleSubmit()` 触发

这使 public 搜索退化为 submit-only，与官方示例和用户预期都不一致。

### 2. public 请求缺少 legacy 关键语义

legacy `AppSearch` 通过 `internal/search/meili_index.go` 提供：

- `preprocessQuery()`
- 公开搜索默认过滤（`type=file`、`is_deleted=false`、`in_trash=false`）

public 直连当前未补齐这两类关键语义，因此召回与结果噪声都可能明显偏离 legacy。

## Target Architecture

```text
┌────────────────────────────────────────────────────────────┐
│ Browser / React                                            │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ InstantSearch Provider                               │  │
│  │  - routing owns query/page/file_category             │  │
│  │  - search-as-you-type input                          │  │
│  │  - Configure injects default filters                 │  │
│  │  - custom search client adapter preprocesses query   │  │
│  └───────────────────────────┬──────────────────────────┘  │
│                              │                             │
│                     POST /multi-search                     │
└──────────────────────────────┼─────────────────────────────┘
                               │
                 ┌─────────────▼─────────────┐
                 │       Meilisearch         │
                 │ public index + search key │
                 └─────────────┬─────────────┘
                               │
      ┌────────────────────────▼────────────────────────┐
      │ AppService.GetSearchConfig / AppDownloadURL    │
      │ legacy AppSearch fallback remains available    │
      └─────────────────────────────────────────────────┘
```

## Key Decisions

### 1. 输入变化驱动 refine，提交只做“立即 flush”

public 搜索必须恢复为：

- `onChange` 经过 debounce 后触发搜索
- `onSubmit` 立即触发当前查询
- `onClear` 清空当前 query 并恢复初始态

这样既能回到官方交互，又保留现有页面的 Enter / 按钮操作习惯。

### 2. query adapter 只改 outbound query，不改输入显示或 URL

为了尽量小地补齐 legacy 搜索语义，adapter 只作用于“发送到 Meilisearch 的 query”，而不改变：

- 输入框中的文本
- URL 中持久化的 query
- 用户对自己输入内容的感知

这能避免把“业务兼容逻辑”扩散进 routing 与 UI。

### 3. 默认过滤使用官方配置入口统一注入

public 默认过滤属于系统硬边界，不应由：

- 用户手动选择
- 结果渲染层二次裁剪
- 各个组件分别拼接

而应由单一配置入口统一注入到搜索请求中，确保：

- 总数正确
- refinement 计数正确
- 空态正确
- 不泄漏 folder / deleted / trash 文档

### 4. `file_category` refinement 保持为用户态过滤

`file_category` 是用户可见的筛选，不是系统默认过滤。它应继续：

- 由 refinement hooks 驱动
- 与 URL 同步
- 叠加在默认过滤之上

### 5. 结果对比与回滚门槛纳入架构交付

这轮不是“修完看感觉”，而是把以下产物纳入交付：

- 代表性查询集
- public vs legacy 的对比输出
- 阻塞级差异定义
- 默认启用 / 回滚条件

## Expected File Touch Points

### Primary Frontend Files

- `web/src/routes/index.lazy.tsx`
  - public 输入模型纠偏
  - debounce / submit / clear 行为调整
  - public 搜索配置组合
- `web/src/lib/meili-search-client.ts`
  - search client 扩展为带 query adapter 的稳定工厂
- `web/src/components/search-input.tsx`
  - 仅在需要时补轻微交互适配，不改变组件角色
- `web/src/components/search-results.tsx`
  - 主要用于验证默认过滤与结果计数未回退
- `web/src/components/search-filters.tsx`
  - 用于验证 refinement 继续叠加默认过滤

### Suggested New Frontend Utility

- `web/src/lib/search-query-normalizer.ts`
  - 承载最小 legacy preprocess 对齐逻辑
  - 便于单测与长期维护

### Existing Backend Reference Files

- `internal/search/meili_index.go`
  - 作为 preprocess 与默认过滤的语义来源
- `internal/httpx/connect_app_auth_search.go`
  - 作为 legacy AppSearch 默认约束来源

### Verification Assets

- `web/src/components/search-page.test.tsx`
- `web/src/lib/meili-search-client.test.ts` 或等价测试文件
- `web/e2e/tests/search.spec.ts`
- `tasks/todo.md`

## Non-Goals

- 不重建 public bootstrap 协议
- 不删除 public / legacy 双栈
- 不在本轮引入更深的排序与相关性调优
- 不在本轮复刻完整 `All -> Last` 搜索 fallback
