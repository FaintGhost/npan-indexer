# Task 003: Test rate limit middleware mounting

**depends-on**: (none)

## Description

为速率限制中间件挂载创建测试。验证 NewServer 返回的 Echo 实例在高频请求时返回 429，以及不同路由组使用不同限流参数。

## Execution Context

**Task Number**: 003 of 012
**Phase**: Rate Limit
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 2.1 (搜索端点速率限制), Scenario 2.3 (管理端点独立限流)

## Files to Modify/Create

- Create: `internal/httpx/server_ratelimit_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD specs 中 Scenario 2.1 和 2.3 存在

### Step 2: Implement Test (Red)

- 创建 `internal/httpx/server_ratelimit_test.go`
- 测试 `TestNewServer_RateLimitOnSearchEndpoint`: 创建 NewServer，使用 `httptest` 向 `/api/v1/app/search?query=test` 发送大量请求（超过 burst），验证后续请求返回 429 且包含 `Retry-After` header
- 测试 `TestNewServer_RateLimitOnAdminEndpoint`: 向 `/api/v1/admin/sync/full/progress` 发送请求（需带 API Key header），验证管理端点在更低阈值时返回 429
- 需要构造最小的 Handlers 实例（可以传入 mock/nil 的 queryService 和 syncManager，因为请求会被 rate limiter 拦截在到达 handler 之前）
- **注意**: NewServer 当前签名 `NewServer(handlers *Handlers, adminAPIKey string)` 不接受限流参数，测试应假设使用默认值
- **Verification**: 测试应失败（Red），因为 NewServer 当前不挂载 RateLimitMiddleware

### Step 3: Verify Red

- 运行测试确认请求不返回 429（当前无限流）

## Verification Commands

```bash
go test ./internal/httpx/... -run "TestNewServer_RateLimit" -v
```

## Success Criteria

- 测试因缺少限流而失败（Red）
- 测试逻辑正确映射 BDD Scenario 2.1 和 2.3
