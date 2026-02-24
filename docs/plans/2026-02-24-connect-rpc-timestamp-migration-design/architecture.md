# Architecture

## Current State

- 进度模型核心时间字段使用 `int64`（毫秒）：
  - `proto/npan/v1/api.proto`
  - `internal/models/models.go`
  - `internal/service/sync_manager.go`
  - `internal/httpx/connect_admin.go`
- 前端当前以数字时间戳消费进度，格式化逻辑依赖 `number`。

## Compatibility Constraint

protobuf 字段类型不可就地替换。把旧字段从 `int64` 改为 `Timestamp` 属于破坏性变更，不可作为渐进迁移方案。

## Target Architecture (Phase 1)

### Proto Layer

为以下消息新增 Timestamp sidecar 字段（示意）：

- `CrawlStats.started_at_ts` / `ended_at_ts`
- `RootSyncProgress.updated_at_ts`
- `SyncProgressState.started_at_ts` / `updated_at_ts`

旧 `int64` 字段保留不动。

### Backend Mapping Layer

在 Connect 响应转换层填充双字段：

- 继续输出旧 `int64` 字段；
- 同时基于旧值生成 `timestamppb.Timestamp` 输出到新字段。

不改 `internal/models` 与 `storage` 持久化结构，避免扩大变更半径。

### Frontend Consumption Layer

新增统一时间解析适配器：

- 优先读取 `*_ts`
- 若缺失则回退旧 `int64`
- 对上层 UI 暴露统一的时间数值/日期对象

## Migration Phases

### Phase 1（本计划覆盖）

- 新增 sidecar 字段
- 后端双写输出
- 前端优先新字段并支持回退
- 完整回归验证

### Phase 2（后续独立批次）

- 统计客户端覆盖率
- 确认是否可停用旧字段
- 若可行，执行旧字段清理（需单独 design + plan）

## Risks & Mitigations

- 风险：前端/测试 fixture 仅使用旧字段导致新字段未覆盖
  - 缓解：新增双输入测试（仅新字段、仅旧字段、同时存在）
- 风险：新旧字段时区/精度不一致
  - 缓解：统一按 UTC 毫秒转换，断言同一时刻
- 风险：误把持久化也改为 `Timestamp` 扩大影响
  - 缓解：本批次明确禁止修改存储模型结构
