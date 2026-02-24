# BDD Specifications

## Feature: Review 建议状态分流

### Scenario: 已完成建议被识别为已采纳而非重复实施

```gherkin
Given 当前分支已经包含 Connect-RPC 路由挂载、query-es 生成链路和枚举 UNSPECIFIED 零值
When 维护者参考新的 review.md 进行下一步规划
Then 这些建议应被标记为“已完成”
And 不应再安排重复改造任务
And 设计文档应记录对应理由与后续边界
```

### Scenario: 暂缓项在设计中被明确记录触发条件

```gherkin
Given review 建议将时间戳字段迁移到 google.protobuf.Timestamp
When 当前阶段评估发现会影响模型、存储、前端类型和兼容性验证
Then 设计文档应将 Timestamp 迁移标记为“暂缓”
And 文档应记录进入该批次的前置条件
```

## Feature: Protovalidate 注解增量落地

### Scenario: 命中 proto 规则时由 validation interceptor 返回 invalid_argument

```gherkin
Given StartSyncRequest 或 InspectRootsRequest 在 proto 中配置了 protovalidate 规则
When 客户端发送不满足规则的请求（例如 folder_ids 为空或 page_size 为 0）
Then Connect validation interceptor 应拦截请求
And 返回 connect CodeInvalidArgument
And 业务 handler 不应继续执行
```

### Scenario: 未配置规则的消息保持 no-op 行为

```gherkin
Given 某个 RPC 请求消息暂未配置 protovalidate 规则
When 客户端发送该请求且结构合法
Then validation interceptor 不应阻断请求
And 请求应继续进入现有业务 handler
And 现有测试行为应保持不变
```

## Feature: 兼容性优先的渐进迁移

### Scenario: 本批次不引入 Timestamp 契约变化

```gherkin
Given 当前 Connect 与 REST 仍以 int64 时间戳字段保持兼容
When 执行本批次“review 收敛”工作
Then 不应修改 proto 中现有时间字段类型为 Timestamp
And 不应引入前端时间字段类型的大面积变化
And 变更范围应集中在 validation 规则与测试补充
```

### Scenario: 业务语义校验继续保留在后端防线

```gherkin
Given force_rebuild 与 scoped roots 的互斥属于业务语义约束
When proto 增量引入 protovalidate 规则
Then 该互斥校验仍应保留在 handler 或 service 中
And API 行为应继续返回既有错误码语义
```

## Suggested Automated Tests

## Go

- `internal/httpx/connect_*_test.go`
  - 新增 protovalidate 命中案例（`invalid_argument`）
  - 校验无规则消息仍可通过（no-op）
- `internal/httpx/server_routes_test.go`
  - 回归路由与鉴权行为未变化

## Generation / Contract

- `buf lint`
- `buf generate`
- `git diff --check`

## Regression

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect|Routes|Health|Admin' -count=1`
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./... -count=1`

## Optional Full Validation (when implementation lands)

- `cd web && bun vitest run`
- `./tests/smoke/smoke_test.sh`
- `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright`
