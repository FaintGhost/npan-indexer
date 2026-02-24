# Connect-RPC Timestamp Migration Design

## Context

当前 `proto/npan/v1/api.proto` 中与同步进度相关的时间字段仍是 `int64`（毫秒时间戳），例如：

- `CrawlStats.started_at` / `ended_at`
- `RootSyncProgress.updated_at`
- `SyncProgressState.started_at` / `updated_at`

之前已经完成 Connect-RPC 渐进接入与 protovalidate 规则落地。下一批目标是推进时间字段演进到 `google.protobuf.Timestamp`，但必须保持现有 REST/CLI/存储与老客户端兼容。

## Problem Statement

不能直接把已有 protobuf 字段类型从 `int64` 改成 `Timestamp`，这会破坏 protobuf 向后兼容与现有消费链路。需要一条低风险演进路径，让新老客户端在过渡期可并存。

## Goals

- 为 Connect 客户端提供 `google.protobuf.Timestamp` 字段。
- 保持现有 `int64` 字段在过渡期继续可用。
- 服务端保证新旧字段语义一致（同一时刻）。
- 前端消费侧优先读新字段，缺失时回退旧字段。

## Non-Goals

- 本批次不移除旧 `int64` 字段。
- 本批次不调整持久化 JSON 结构（仍以 `int64` 存储）。
- 本批次不做 `internal/httpx` 包结构重构。

## Options Considered

### Option A: 一次性替换旧字段为 Timestamp

优点：

- schema 更“纯净”

缺点：

- 破坏向后兼容
- 波及 Go/TS 生成、前端/CLI/存储与测试全链路

### Option B（推荐）: 双字段过渡（新增 Timestamp sidecar 字段）

优点：

- 保持兼容
- 可以渐进切换消费侧
- 风险可控

缺点：

- 过渡期存在新旧双字段维护成本

## Decision

采用 Option B：

- 在 proto 中新增 `*_ts` 的 `google.protobuf.Timestamp` 字段；
- 保留旧 `int64` 字段；
- 服务端输出同时填充新旧字段；
- 前端优先消费 `*_ts`，缺失时回退 `int64`。

## Detailed Design Summary

- Proto：
  - 为 `CrawlStats`、`RootSyncProgress`、`SyncProgressState` 增加 Timestamp sidecar 字段。
- 后端（Connect DTO 转换）：
  - 在 `internal/httpx/connect_admin.go` 的转换路径中同步填充 `*_ts`。
- 前端：
  - 增加 Timestamp 解析适配器，统一转换为 UI 所需数字时间戳或 `Date`。
  - 优先新字段，回退旧字段。
- 测试：
  - 覆盖 descriptor 层字段存在性、后端映射一致性、前端回退行为与全量回归。

## Success Criteria

- 生成链路通过（`buf lint` / `buf generate`）。
- Connect 返回中可观察到 `*_ts` 字段，并与旧 `int64` 对应同一时刻。
- 前端在新旧字段两种输入下均正确渲染。
- 全量 Go/关键前端测试通过。

## Open Questions

- `*_ts` 字段命名是否统一采用 `_ts` 后缀（建议）还是 `_timestamp` 后缀。
- 旧字段下线窗口（需要观察客户端升级覆盖率后再定）。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
