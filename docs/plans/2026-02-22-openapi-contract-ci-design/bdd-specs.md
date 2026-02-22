# BDD Specifications

## Feature: OpenAPI 契约一致性

### Scenario: 从 spec 生成的 Go types 与现有 handler 响应结构匹配
```gherkin
Given OpenAPI spec 中定义了 SyncProgressState schema
And schema 的 status 字段是 enum [idle, running, done, error, cancelled, interrupted]
When 执行 `make generate-go`
Then api/types.gen.go 包含 SyncProgressState struct
And struct 的 Status 字段类型为生成的 enum 类型
And 所有 JSON tag 与 spec 中的 property name 一致
```

### Scenario: 从 spec 生成的 Zod schemas 能解析后端 JSON 响应
```gherkin
Given OpenAPI spec 中定义了 SearchResponse schema
When 执行 `make generate-ts`
Then web/src/api/generated/ 目录下生成 Zod schema
And 使用生成的 schema 解析后端 /api/v1/app/search 的 JSON 响应能通过验证
```

### Scenario: Spec 变更后未重新生成代码时 CI 失败
```gherkin
Given 开发者修改了 openapi.yaml 中的某个 schema
But 没有执行 `make generate`
When CI 运行 `make generate-check`
Then CI 检测到 api/types.gen.go 与 spec 不同步
And CI 失败并提示 "Generated code is out of date. Run 'make generate' and commit."
```

## Feature: CI 冒烟测试

### Scenario: 服务栈正常启动
```gherkin
Given docker-compose.ci.yml 定义了 meilisearch 和 npan 服务
When CI 执行 `docker compose up -d --wait`
Then meilisearch 容器 healthcheck 通过
And npan 容器 healthcheck 通过
And GET http://localhost:1323/healthz 返回 200
And 响应 JSON 包含 {"status": "ok"}
```

### Scenario: 未认证请求被拒绝
```gherkin
Given 服务栈已启动
When 不带 X-API-Key 请求 GET /api/v1/admin/sync
Then 返回 401 Unauthorized
And 响应 JSON 包含 {"code": "UNAUTHORIZED"}
```

### Scenario: 认证后管理端点可用
```gherkin
Given 服务栈已启动
And 环境变量 NPA_ADMIN_API_KEY 设置为 test-admin-key
When 带 X-API-Key: test-admin-key 请求 GET /api/v1/admin/sync
Then 返回 200 或 404（取决于是否有同步进度）
And 响应 JSON 符合 ErrorResponse 或 SyncProgressState schema
```

### Scenario: 搜索端点返回正确结构
```gherkin
Given 服务栈已启动且 Meilisearch 可达
When 请求 GET /api/v1/app/search?q=test
Then 返回 200
And 响应 JSON 包含 items 数组和 total 数字字段
```

## Feature: 开发工作流

### Scenario: 开发者添加新 API 字段的完整流程
```gherkin
Given 开发者需要给 SyncProgressState 添加新字段 estimatedEndTime
When 开发者编辑 api/openapi.yaml 添加 estimatedEndTime property
And 执行 `make generate`
Then Go types 中自动出现 EstimatedEndTime 字段
And TS Zod schema 中自动出现 estimatedEndTime 验证
And 开发者在后端 handler 中赋值该字段
And 前端代码使用生成的类型引用该字段，获得编译期检查
```
