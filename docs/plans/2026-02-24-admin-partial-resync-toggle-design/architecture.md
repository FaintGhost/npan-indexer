# Architecture

## Current Behavior (Root Cause)

- 前端 `web/src/components/admin-sync-page.tsx` 将目录 ID 输入直接用于 `startSync()`
- 前端 `web/src/hooks/use-sync-progress.ts` 会把 `root_folder_ids` 发到 `/api/v1/admin/sync`
- 后端 `internal/service/sync_manager.go` 在 full path 中依据本次 `roots` 创建/恢复 progress
- 当本次只传一个 root 时，`progress.Roots` 与 `progress.RootProgress` 被重建为该集合
- `web/src/components/sync-progress-display.tsx` 的根目录详情直接渲染 `progress.rootProgress`

结果：

- 用户执行一次单目录 scoped full 后，UI 根目录详情只剩该目录

## Target UX Flow

```text
[输入目录 ID] --(点击“拉取目录详情”)--> [目录册列表新增条目 + 默认勾选]
                                      |
                                      v
                         [根目录详情(带 toggle)]
                                      |
                          (勾选异常目录若干个)
                                      |
                                      v
                          [启动同步 -> 全量(仅勾选目录)]
```

## API Design

### 1) New Admin API: Inspect Roots (recommended)

`POST /api/v1/admin/roots/inspect`

请求（示例）：

```json
{
  "folder_ids": [123456, 789012]
}
```

响应（示例）：

```json
{
  "items": [
    {
      "folder_id": 123456,
      "name": "PIXELHUE",
      "item_count": 4151,
      "estimated_total_docs": 4152
    }
  ],
  "errors": [
    {
      "folder_id": 789012,
      "message": "folder not found"
    }
  ]
}
```

说明：

- 用于“拉取目录详情”按钮，不启动同步
- 支持部分成功，便于批量输入时给出可用结果
- 底层复用已有 `npan.GetFolderInfo()`

### 2) Start Sync Request Extension (scoped full + preserve catalog)

扩展 `/api/v1/admin/sync` 请求字段（建议）：

- `selected_root_ids?: number[]`
  - 表示本次执行范围（来自 UI toggle）
- `preserve_root_catalog?: boolean`
  - 为 `true` 时，局部全量运行后保留历史 root 详情用于展示

兼容策略：

- 若未传 `selected_root_ids`，沿用现有 `root_folder_ids`
- CLI 与现有客户端保持兼容，不需要立即跟进

备注：

- 也可直接复用 `root_folder_ids` 作为“执行范围”，再增加 `preserve_root_catalog`
- 本设计推荐显式区分“手动拉取目录输入”和“本次执行勾选目录”，减少前端状态混淆

## Progress Model Evolution

### Problem

当前 `SyncProgressState` 的 `Roots/RootProgress` 同时承担两种职责：

- 本次执行范围（进度条）
- 根目录详情列表（历史可视化）

这两个职责在局部补同步场景下冲突。

### Recommended Shape (Backward-Compatible Extension)

在保持现有字段可用的前提下，新增目录册字段：

- `catalogRoots?: number[]`
- `catalogRootNames?: map`
- `catalogRootProgress?: map`

语义：

- `roots` / `completedRoots` / `activeRoot`: 本次运行范围与进度
- `catalog*`: 用于 UI 根目录详情与 toggle 列表（历史 + 新拉取）

兼容策略：

- 旧前端仍读取 `rootProgress` 可正常工作
- 新前端优先读取 `catalogRootProgress`，缺失时回退 `rootProgress`

### Minimal Alternative (no new response fields)

若要最小化 OpenAPI 变更：

- 仍使用 `rootProgress` 承载目录册（包含历史项）
- `roots` 仅表示本次执行范围

可行但缺点：

- `rootProgress` 语义变宽，后续维护容易误解

本设计建议优先采用 `catalog*` 显式字段。

## Backend Changes

### `internal/httpx/handlers.go`

- 新增 `InspectRoots` handler
- 解析批量 `folder_ids`
- 调用 `npan.API.GetFolderInfo()`
- 返回部分成功结果

### `internal/httpx/server.go`

- 注册新路由：`POST /api/v1/admin/roots/inspect`

### `api/openapi.yaml`

- 新增 `InspectRootsRequest/Response` schema 与 endpoint
- 扩展 `SyncStartRequest`（若采用 `selected_root_ids` / `preserve_root_catalog`）
- 扩展 `SyncProgressState`（若采用 `catalog*` 字段）

### `internal/service/sync_manager.go`

- 全量 scoped run 时：
  - 执行范围使用本次勾选 roots
  - 若 `preserve_root_catalog=true`，从 existing progress 合并历史目录册字段
  - 本次运行结束后更新目录册中被执行 roots 的最新统计
- 增量路径：
  - 继续保留目录册字段（与当前保留 roots/rootProgress 的思路一致）

## Frontend Changes

### `web/src/components/admin-sync-page.tsx`

- 将目录输入区改为“拉取目录详情”用途
- 增加：
  - `拉取目录详情` 按钮
  - `按勾选目录同步` 按钮（全量模式下可见）
- 状态：
  - `fetchFolderIdsInput`
  - `selectedRootIds`
  - `inspectLoading/inspectError`

### `web/src/components/sync-progress-display.tsx`

- 根目录详情行新增 toggle（受控）
- 可选展示来源：
  - `catalogRootProgress`（优先）
  - `rootProgress`（回退）
- 对“历史项但不在本次 roots 中”的条目增加弱提示（如“历史统计”）

### `web/src/hooks/use-sync-progress.ts`

- 新增 `inspectRoots(folderIds)` 调用
- `startSync()` 支持传入：
  - `selectedRootIds`
  - `preserveRootCatalog`

## State Rules

- 页面初始化：
  - 拉取 progress
  - 用目录册（或 `rootProgress` 回退）初始化 toggle 列表
- 拉取目录详情成功：
  - 合并到目录册显示
  - 新条目默认勾选（待用户确认）
- 同步运行中：
  - 禁用 toggle、拉取按钮、模式切换、force_rebuild

## Error Handling

- `InspectRoots` 部分失败不阻断成功项合并
- 对无效 ID 显示逐项错误，不清空已有目录册
- 若同步启动失败，保持当前勾选状态不丢失

## Testing Hooks

- 后端：
  - handler：批量 inspect 的部分成功响应
  - service：scoped full + preserve catalog 后 progress 不覆盖历史目录册
- 前端：
  - 拉取目录详情不触发同步请求
  - toggle 勾选决定 `selected_root_ids`
  - 局部同步后根目录详情仍显示历史项
