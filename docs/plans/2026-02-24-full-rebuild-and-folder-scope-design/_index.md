# Full Rebuild And Folder Scope Design

## Context

用户在手动选择“全量 + 强制重建”后，观察到某些目录统计显著偏小（示例：官方 GUI 显示 4152，而本次仅 393 文件 / 119 文件夹 / 119 页）。

本设计聚焦 3 个目标：

1. 修复 `force_rebuild` 仍可能复用旧 checkpoint 的根因。
2. 增加“单独索引文件夹”能力（Admin UI 可输入目录 ID）。
3. 引入官方统计能力作为可视化预估与完成后告警依据。

## Requirements

- `force_rebuild=true` 必须保证从根目录第 0 页重新开始，不复用旧 checkpoint。
- Admin 页面支持输入一个或多个目录 ID，执行目录范围索引。
- 当用户执行目录范围索引时，默认不混入部门目录（显式 `include_departments=false`）。
- 对显式目录 ID，显示官方统计预估（来自 Npan OpenAPI）。
- 同步完成后对“预估总项 vs 实际写入项”做差异提示。
- 保持兼容现有 API/CLI 行为，不破坏历史调用。

## Rationale

当前“统计偏小但状态 done”的高概率根因是：

- `RunFullCrawl()` 会优先 `Load()` checkpoint；
- `force_rebuild` 仅清空索引，不会清理 checkpoint 文件；
- 若存在残留 checkpoint，会从旧断点继续，导致“全量名义、增量式实际”结果。

因此，修复要点不是改分页逻辑，而是确保“强制重建 = 全新遍历状态”。

## Detailed Design

### 1) Force Rebuild 清理 Checkpoint（根因修复）

- 在 `SyncManager.run()` 全量路径中，确定 `force_rebuild=true` 或 `resume=false` 后：
  - 计算本次生效根目录对应的 checkpoint 文件路径；
  - 在每个 root 开跑前先执行 `CheckpointStore.Clear()`（或等价 `os.Remove` 封装）；
  - 再进入 `runSingleRoot()`。
- 目的：保证 `RunFullCrawl()` 的 `Load()` 返回空状态，从 `root/page=0` 开始。

### 2) 单独索引文件夹（目录范围）

- 后端已支持 `root_folder_ids` 与 `include_departments` 请求字段，复用现有能力。
- 前端 Admin 新增“目录范围”输入区：
  - 输入单个/多个目录 ID（逗号分隔）；
  - 启动时请求体携带：
    - `root_folder_ids: number[]`
    - `include_departments: false`（当用户指定目录范围时默认关闭）
- 保留“全量全库”行为：
  - 若目录 ID 为空，沿用现有默认配置。

### 3) 官方统计能力接入（预估 + 告警）

- 新增 `npan.API` 方法：
  - `GetFolderInfo(ctx, folderID)` -> 读取 `/api/v2/folder/{id}/info`。
- 在 `discoverRootFolders()` 中，对于显式 `root_folder_ids`：
  - 调 `GetFolderInfo` 填充：
    - `rootNameMap[rootID]`
    - `rootEstimateMap[rootID] = item_count + 1`
- `RootProgress.estimatedTotalDocs` 已有字段，直接复用。
- 在全量完成后，追加差异告警：
  - 对每个 root 计算 `actual = filesIndexed + foldersVisited`
  - 若 `estimatedTotalDocs > 0` 且差异超过阈值（默认 5% 或绝对值 > 20），写入 `verification.warnings`。

### 4) UI 展示增强

- Admin 进度卡新增：
  - 当前运行模式（全量/增量）
  - 目录范围（若有）
  - 每个 root 的 `estimatedTotalDocs` 与 `actual` 差值状态（仅有 estimate 时显示）
- 避免误解：
  - 在“强制重建”提示文案中明确说明“会重置断点并全量重跑”。

## Scope Boundaries

- 本次不引入新的独立统计端点（优先复用现有同步链路与 progress 结构）。
- 本次不调整全局分页策略（`page_count` 仍由上游 API 返回驱动）。
- 本次不改动索引文档结构（不新增 root_id 字段到 Meili 文档）。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
