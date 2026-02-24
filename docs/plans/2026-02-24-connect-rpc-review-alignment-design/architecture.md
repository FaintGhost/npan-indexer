# Architecture

## Current State Snapshot

基于当前分支（`feat/connect-rpc-migration-s1-s2`）：

- `.proto` 契约统一在 `proto/npan/v1/api.proto`
- `buf` 生成链路已包含：
  - `connect-go`
  - `connect-es`
  - `query-es`
- Connect 服务已接入并挂载：
  - `HealthService`
  - `AppService`
  - `AuthService`
  - `SearchService`
  - `AdminService`
- Connect 拦截器已接入：
  - 统一错误拦截器（未知错误 -> `internal`）
  - validation interceptor（已接入 `protovalidate` runtime，但 schema 尚未写规则）

## Review Suggestions Status Matrix

### 已完成（无需重复返工）

- 枚举零值使用 `*_UNSPECIFIED`
- `connect-es` 与 `query-es` 生成链路
- Connect 服务在 Echo 中渐进挂载（REST 并存）
- Connect 统一错误拦截器
- Connect validation interceptor 基础设施

### 待采纳（建议下一批执行）

- 在 `.proto` 中引入 `buf/validate/validate.proto` 并补充规则注解

### 暂缓（单独立项）

- `google.protobuf.Timestamp` 替换现有 `int64` 时间戳字段
- `internal/rpc` 包抽离与依赖注入重组
- Connect 路由路径策略调整（例如 `/rpc/*`）

## Why Timestamp Is Deferred

`Timestamp` 迁移不是单点改动，会同时影响：

- Go 端模型与 DTO 转换（当前大量 `int64` 毫秒语义）
- 存储层（进度 JSON 持久化结构）
- 前端类型与展示逻辑（当前默认 `number` 时间戳）
- 与既有 REST 契约的兼容性验证

因此它应作为独立批次处理，避免与 schema 校验规则落地相互干扰。

## Next Batch Scope (Recommended)

### 1. Proto Validation Rules (Incremental)

第一批只覆盖“高收益 + 低争议”的请求字段：

- `StartSyncRequest`
  - `root_folder_ids`（若提供则需为正整数）
  - `department_ids`（若提供则需为正整数）
  - `root_workers`、`progress_every`（如提供需为正数）
- `InspectRootsRequest`
  - `folder_ids` 非空且元素为正整数
- 分页类请求（如 `LocalSearchRequest`、`AppSearchRequest`）
  - `page >= 1`
  - `page_size` 在合理区间内（例如 1-100）

说明：

- 不强制第一批覆盖全部请求，先验证流程与错误码行为。

### 2. Connect Validation Behavior

数据流：

```text
Connect Request
  -> validation interceptor (protovalidate)
    -> 命中规则错误: 返回 invalid_argument
    -> 无规则/通过: 进入业务 handler
      -> 业务错误: 由 error interceptor 统一映射/透传
```

关键点：

- schema 校验与业务校验可短期并存，但要避免重复/冲突错误文案。
- 优先把“结构性约束”迁到 proto（空、范围、格式、数组元素约束）。
- 业务语义约束（如 `force_rebuild + scoped`）仍保留在 handler/service 防线。

### 3. Test Strategy

建议新增或补充：

- 拦截器集成测试
  - 命中 protovalidate 规则时返回 `connect.CodeInvalidArgument`
  - 对未配置规则的消息不拦截（no-op）
- 现有 Admin/Search Connect 测试回归
  - 确认错误拦截器与 validation interceptor 顺序不破坏既有行为
- 生成链路验证
  - `buf lint`
  - `buf generate`

## Future Batch Gate: Timestamp Migration

只有在满足以下条件后才进入 `Timestamp` 迁移批次：

- Connect 客户端消费侧（前端/脚本）已确认可接受 `Timestamp` 表示形式
- 现有 REST 与 Connect 的时间字段对齐策略已确定（并完成测试方案）
- 已完成影响面清单（模型、存储、前端、测试、文档）

建议该批次单独产出 design + writing-plans，再实施。

## Future Batch Gate: `internal/rpc` Extraction

只有在满足以下条件后才考虑包抽离：

- Connect 服务数量继续增加，`internal/httpx` 的职责开始明显混杂
- 需要统一复用 RPC server 构造逻辑（非 Echo 场景、独立集成测试等）
- 结构收益明显大于短期迁移成本

在当前阶段，保持 Connect 实现在 `internal/httpx` 内部更符合“最小影响面”原则。
