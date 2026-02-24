# BDD Specifications

## Feature: Root Details Are Selectable For Partial Re-sync

### Scenario: 首次全量后在根目录详情中勾选异常目录进行局部补同步

```gherkin
Given 用户已经完成一次全量同步
And 管理页显示多个根目录详情条目
When 用户在根目录详情中仅勾选目录 1001 和 1003
And 用户选择全量模式
And 用户点击启动同步
Then 前端应向 /api/v1/admin/sync 发送本次勾选目录范围
And 后端应仅执行目录 1001 和 1003 的全量爬取
And 未勾选目录不应在本次运行中被执行
```

### Scenario: 局部补同步后根目录详情列表仍保留历史条目

```gherkin
Given 历史 progress 中存在根目录 1001,1002,1003 的详情
When 用户执行仅包含目录 1002 的局部全量同步
Then 同步完成后的根目录详情列表仍应包含 1001,1002,1003
And 目录 1002 的统计应更新为本次结果
And 目录 1001 与 1003 的历史统计不应被清空
```

## Feature: Fetch Root Details Is Decoupled From Sync

### Scenario: 拉取目录详情只更新列表，不启动同步

```gherkin
Given 用户在目录输入框输入 "123456"
When 用户点击“拉取目录详情”
Then 前端应调用目录详情查询接口
And 前端不应调用 /api/v1/admin/sync
And 根目录详情列表应新增目录 123456 条目
```

### Scenario: 批量拉取目录详情部分成功

```gherkin
Given 用户输入 "1001,999999,1002"
And 其中 999999 为无效目录
When 用户点击“拉取目录详情”
Then 有效目录 1001 和 1002 应加入根目录详情列表
And 页面应显示 999999 的错误信息
And 已有目录详情条目不应被清空
```

## Feature: Interaction Safety During Running Sync

### Scenario: 运行中禁止修改目录选择和拉取目录详情

```gherkin
Given 当前同步任务状态为 running
When 用户尝试点击根目录 toggle 或“拉取目录详情”
Then 控件应为禁用状态
And 不应发出新的目录详情查询或同步请求
```

### Scenario: 运行中禁止重复发起局部补同步

```gherkin
Given 当前同步任务状态为 running
When 用户点击“启动同步”
Then 按钮应被禁用
And 页面仅显示“同步进行中”状态
```

## Feature: Force Rebuild Guardrails

### Scenario: force_rebuild 与局部补同步互斥

```gherkin
Given 用户勾选了部分根目录作为本次同步范围
And 用户打开 force_rebuild 开关
When 用户尝试启动同步
Then 页面应阻止提交并提示 force_rebuild 仅允许全量全库执行
And 不应调用 /api/v1/admin/sync
```

### Scenario: 全量全库时允许 force_rebuild

```gherkin
Given 用户未指定局部根目录范围
And 用户打开 force_rebuild 开关
When 用户点击启动同步
Then 前端应发送 force_rebuild=true
And 同步请求应按全量全库语义启动
```

## Feature: Catalog Fallback Compatibility

### Scenario: 后端未返回 catalog 字段时前端回退到 rootProgress

```gherkin
Given 管理页收到旧版 SyncProgressState（仅包含 rootProgress）
When 页面渲染根目录详情
Then 页面仍应正确显示根目录详情列表
And toggle 功能仍可基于 rootProgress 工作
```

## Suggested Automated Tests

## Frontend (Vitest + RTL)

- `web/src/components/admin-page.test.tsx`
  - 拉取目录详情按钮不触发同步
  - toggle 勾选后启动同步发送勾选范围
  - running 状态禁用 toggle 与拉取按钮
  - `force_rebuild + scoped selection` 被阻止
- `web/src/hooks/use-sync-progress.test.ts`
  - 新增 `inspectRoots()` 请求与错误处理
  - `startSync()` 请求体包含 `selected_root_ids/preserve_root_catalog`
- `web/src/components/sync-progress-display.test.tsx`
  - 优先渲染 `catalogRootProgress`
  - 回退渲染 `rootProgress`

## Backend (Go tests)

- `internal/httpx/handlers_test.go`
  - `InspectRoots` 批量部分成功响应
- `internal/service/sync_manager_progress_test.go` 或新增测试文件
  - scoped full + preserve catalog 不覆盖历史目录册
- `internal/service/sync_manager_routing_test.go`
  - `force_rebuild` 与 scoped selection 互斥（若后端也做防线）
