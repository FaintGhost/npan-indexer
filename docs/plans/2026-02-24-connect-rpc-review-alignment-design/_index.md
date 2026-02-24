# Connect-RPC Review Alignment Design

## Context

最新 `review.md` 对 Connect-RPC 迁移提出了几类建议：

- Protobuf schema 最佳实践（枚举零值、时间戳类型、protovalidate）
- `buf` 生成链路（`connect-go` / `connect-es` / `query-es`）
- Stage 3 后端接入与绞杀式迁移方式

当前分支已完成较大一部分建议（包括 `query-es`、Connect handler 接入、错误拦截器、validation interceptor 基础设施、Admin Connect 路由与测试）。如果不先做一次审查收敛，后续容易重复返工或在错误层面引入高风险改造（例如立即切换全部时间戳字段到 `google.protobuf.Timestamp`）。

## Problem Statement

需要把新 review 建议转化为“可执行且与当前代码状态一致”的下一步方案，而不是直接照搬：

- 哪些建议已经完成，应记录为已采纳；
- 哪些建议值得下一批执行（例如 proto 中补 `protovalidate` 规则）；
- 哪些建议暂不执行（例如 `Timestamp` 全量迁移、`internal/rpc` 包抽离），并给出明确原因与触发条件。

## Goals

- 形成一份面向下一批执行的 Connect-RPC review 收敛设计。
- 明确 review 建议的状态分流：已完成 / 待采纳 / 暂缓。
- 为 `protovalidate` 注解落地定义最小增量范围与验证策略。
- 为 `Timestamp` 迁移定义延后策略与进入条件，避免当前批次引入兼容性风险。

## Non-Goals

- 本次不直接实现 proto 字段的 `Timestamp` 迁移。
- 本次不进行 `internal/httpx` -> `internal/rpc` 的包层重构。
- 本次不调整 Connect 路由路径（例如统一加 `/rpc/` 前缀）。
- 本次不执行功能代码变更（仅产出设计与 BDD 规格）。

## Requirements

- 设计结论必须基于当前代码库状态，而不是仅基于 review 示例代码。
- 保持渐进迁移原则：REST 与 Connect 共存，不引入不必要重构。
- 优先利用已接入的 Connect validation interceptor，避免重复造校验逻辑。
- 明确验证闭环（`buf lint`、`buf generate`、Go 测试、前端生成/测试）。

## Options Considered

### Option A: 全量采纳 review（立即做 Timestamp + protovalidate + internal/rpc 抽离）

优点：

- 最大程度贴近 review 建议
- 一次性完成 schema 与目录重构

缺点：

- 变更面过大，混合 schema 演进与包重构，回归成本高
- `Timestamp` 迁移会波及 Go 模型、存储、前端类型与序列化假设
- 不符合当前分支“最小改动、渐进迁移”的既有决策

### Option B（推荐）: 增量收敛（优先 protovalidate 注解，Timestamp/包重构暂缓）

优点：

- 复用已接入的 validation interceptor，收益快、风险低
- 与当前分支状态和测试结构一致，易验证
- 能把 `Timestamp` 与包抽离拆成独立议题，降低耦合

缺点：

- review 的部分建议需要延后，短期内仍保留 `int64` 时间戳

### Option C: 仅记录 review，不继续推进 schema 硬化

优点：

- 无开发成本

缺点：

- 已接入的 validation interceptor 价值未释放
- schema 规则仍分散在 handler 代码中，后续迁移成本继续累积

## Decision

选择 Option B。

理由：

- 当前分支已经完成 Connect 路由/服务接入与基础拦截器，最自然的下一步是补齐 `.proto` 规则注解，让 schema 真正承载输入约束。
- `Timestamp` 全量迁移影响面大，且之前已明确为了兼容性暂缓；应单独立项处理。
- `internal/rpc` 抽离属于结构优化，不应与功能推进（schema hardening）绑定在一批。

## Detailed Design Summary

- Review 建议分流：
  - 已完成：枚举 `UNSPECIFIED` 零值、`query-es`、Connect 后端渐进接入、错误拦截器、validation interceptor 基础设施。
  - 下一批采纳：在 `.proto` 增量加入 `protovalidate` 规则注解（先覆盖高价值请求）。
  - 暂缓：`Timestamp` 迁移、`internal/rpc` 包抽离、路由前缀改造。
- 下一批实现建议（供 `writing-plans` 使用）：
  - 先做 `.proto` 注解与 `buf` 依赖校验；
  - 再补 Connect 校验命中/未命中（no-op）测试；
  - 最后做回归验证与文档记录。

## Success Criteria

- 设计文档明确列出 review 建议的状态分流和理由。
- BDD 规格覆盖 `protovalidate` 规则命中、无规则 no-op、兼容性边界。
- 架构文档明确 `Timestamp` 暂缓与未来触发条件，避免下一轮再次争议。

## Open Questions

- `Timestamp` 迁移未来采用哪种策略更合适：
  - 一次性替换现有字段；
  - 新增并行字段（新旧并存一段时间）后再移除旧字段。
- `protovalidate` 第一批规则是否覆盖所有请求，还是仅覆盖 Admin + Search 的高频入口。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
