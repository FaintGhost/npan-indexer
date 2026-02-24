# Best Practices

## 1. 先做 Review 分流，再做代码改动

外部 review 往往混合了：

- 已完成建议
- 可采纳建议
- 不适合当前阶段的建议

在开始改代码前，先做一轮状态分流，可以避免：

- 重复返工
- 因示例代码路径不匹配导致误接入
- 把结构重构和功能推进绑定在同一批

## 2. 让 `.proto` 承担“结构性约束”，业务语义留在服务层

适合迁移到 `protovalidate` 的内容：

- 必填/非空
- 数值范围
- 数组元素范围
- 字符串长度与基础格式

不适合只靠 `protovalidate` 的内容：

- `force_rebuild + scoped roots` 互斥
- 与外部系统状态相关的前置条件
- 依赖运行时上下文的鉴权/授权决策

实践原则：

- schema 规则做“快速失败”
- handler/service 做“业务防线”

## 3. Validation Interceptor 要增量启用，不追求一次覆盖所有 RPC

当前项目已经接入了 validation interceptor，最佳落地方式是：

- 先给高频/高风险入口加规则（Admin、Search、分页类请求）
- 通过测试确认错误码与文案行为稳定
- 再逐步扩展到其余请求

这样可以更快验证收益，同时降低批量改 proto 带来的回归风险。

## 4. `Timestamp` 迁移应单独立项，不与 schema 校验同批

`google.protobuf.Timestamp` 是有价值的，但在当前阶段直接切换会放大风险：

- 影响 DTO 转换与前端类型
- 影响测试快照与现有序列化假设
- 需要明确 REST 与 Connect 的兼容路径

建议流程：

- 先完成 schema validation 收敛
- 再单独产出 Timestamp 迁移设计与计划
- 最后分批实施并验证

## 5. 对 review 示例代码保持“检索优先”

实践规则：

- 先查当前代码是否已实现类似能力
- 先确认真实模块路径与 API 签名
- 再决定采纳/改写 review 示例

本项目已出现过典型案例：

- `protovalidate` 的真实 Go import 路径需以本地模块解析为准，而不能直接照抄外部示例。

## 6. 保持渐进迁移边界稳定

当前 Connect 迁移策略是：

- REST 与 Connect 并存
- 优先复用现有 service/业务逻辑
- 在 `internal/httpx` 内增量扩展，而非同时做大规模包重构

在没有明确结构收益前，不要引入 `internal/rpc` 抽离作为“顺手优化”。

## 7. 验证闭环必须覆盖生成链路与运行时行为

仅跑单测不够，至少要覆盖：

- 契约检查：`buf lint`
- 代码生成：`buf generate`
- Connect 路由/错误码行为：`go test ./internal/httpx ...`
- 全量回归：`go test ./...`

如果本批次涉及前端生成或消费端类型，再追加：

- `cd web && bun run generate`
- `cd web && bun vitest run`
