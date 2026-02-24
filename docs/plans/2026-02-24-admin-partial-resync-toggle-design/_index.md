# Admin Partial Resync Toggle Design

## Context

当前 Admin 同步页把“目录 ID 输入”和“启动同步”绑定在一起：

- 用户输入 `root_folder_ids` 后点击“启动同步”，会直接发起目录范围全量同步。
- 这会让后端本次进度状态以该范围重建 `Roots/RootProgress`，从而导致下方“根目录详情”只剩本次目录。

用户希望的交互是：

- 先做一次完整全量同步。
- 在下方“根目录详情”里通过 toggle 勾选异常目录。
- 在已有同步数据基础上，只对勾选目录做局部补同步（局部全量）。
- “拉取目录详情”应该是单独按钮，用来把新拿到的目录 ID 加入可选列表，而不是立即触发同步。

## Problem Statement

现有设计把“目录发现/录入”和“同步执行”耦合在一起，导致两个问题：

- 易误操作：输入目录 ID 就改变下一次同步范围。
- 进度展示被覆盖：局部同步后 `progress.rootProgress` 只保留本次 root，历史根目录详情消失。

## Goals

- 将“拉取目录详情”和“启动同步”解耦。
- 在“根目录详情”区域支持每个根目录的 toggle 选择。
- 支持“首次全量后，局部补同步异常目录”的高频运维场景。
- 局部同步后保留历史根目录详情列表，不再只剩单个目录。
- 保持现有 `/api/v1/admin/sync` 基本语义兼容（已有客户端不受破坏）。

## Non-Goals

- 本次不改变索引文档结构（不新增 `root_id` 到 Meili 文档）。
- 本次不实现“按目录删除陈旧文档”的额外清理策略。
- 本次不重做整个 Admin 页面布局，仅做交互重构与必要 API 扩展。

## Requirements

- Admin 页面新增“拉取目录详情”按钮（基于输入目录 ID 列表执行，不启动同步）。
- 拉取成功的目录应合并进根目录详情列表，并可被 toggle 选择。
- 全量模式下支持“仅同步已勾选根目录”（局部补同步）。
- 局部补同步完成后，根目录详情列表仍保留历史项（至少包含之前完整同步过的根目录）。
- 运行中禁止修改选择状态、拉取目录详情、重复发起同步。
- `force_rebuild=true` 与“局部补同步”组合需要明确限制（默认禁止）。

## Options Considered

### Option A（推荐）: 新增目录详情拉取接口 + 后端保留根目录目录册（catalog）

- 新增 Admin API（示例）`POST /api/v1/admin/roots/inspect`
- 返回目录基础信息（`id/name/item_count`）
- 局部全量同步时，后端执行范围仍使用本次选中 roots，但在 progress 中保留历史 `rootProgress/rootNames` 作为目录册显示

优点：

- 目录列表以服务端 progress 为准，刷新页面后仍可见
- 解决“列表被覆盖”根因，不依赖浏览器本地缓存
- 对现有 `/api/v1/admin/sync` 行为改动较小（扩展字段）

代价：

- 需要新增一个 Admin API 与 OpenAPI 契约
- `SyncProgressState` 的“执行范围”和“展示目录册”语义需要在设计中明确

### Option B: 仅前端本地缓存目录册（localStorage）并与 progress 合并显示

- 不新增后端 API（可选仍需目录详情 API）
- 前端单独保存历史根目录详情，显示时合并 `progress.rootProgress + local cache`

优点：

- 后端改动小

缺点：

- 状态源不一致（服务端一份、浏览器一份）
- 换浏览器/清缓存/多管理员并行时体验不一致
- 不能从根本上解决 progress 被覆盖

## Decision

选择 Option A。

核心理由：

- 用户反馈的痛点是“局部同步后列表被覆盖”，根因在服务端 progress 语义。
- 该问题应在服务端状态层修复，而不是靠前端缓存掩盖。

## Detailed Design Summary

- 前端：
  - 保留目录 ID 输入框，但从“同步范围输入”改为“拉取目录详情输入”。
  - 增加“拉取目录详情”按钮。
  - 在根目录详情里增加 toggle，并提供“按勾选目录启动全量补同步”入口。
- 后端：
  - 新增 Admin 目录详情查询接口（批量）。
  - 在全量 scoped run 中引入“保留根目录目录册”能力，避免覆盖历史 root 详情。
  - 新增/扩展 progress 字段区分“本次执行 roots”和“目录册 roots”（设计见 `architecture.md`）。
- 测试：
  - 覆盖 UI 行为、请求体契约、progress 保留语义、运行中禁用与互斥规则。

## Open Questions

- 拉取目录详情成功后，新加入的目录默认是否自动勾选用于下一次同步（当前设计默认：自动勾选）。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
