# Task 004: Mount rate limit middleware

**depends-on**: task-003

## Description

在 `NewServer()` 中挂载 `RateLimitMiddleware`。全局层使用 20 rps / burst 40，管理端点叠加 5 rps / burst 10 的更严格限流。

## Execution Context

**Task Number**: 004 of 012
**Phase**: Rate Limit
**Prerequisites**: Task 003 测试已创建

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 2.1 (搜索端点速率限制), Scenario 2.3 (管理端点独立限流)

## Files to Modify/Create

- Modify: `internal/httpx/server.go` — 在 `NewServer()` 中添加 `e.Use(RateLimitMiddleware(20, 40))`，在 admin group 中添加 `RateLimitMiddleware(5, 10)`

## Steps

### Step 1: Add global rate limit

- 在 `NewServer()` 的 middleware 链中（`middleware.RequestLogger()` 之后）添加全局限流：`e.Use(RateLimitMiddleware(20, 40))`

### Step 2: Add admin-specific rate limit

- 在 admin group 创建时添加更严格的限流：
  ```
  admin := e.Group("/api/v1/admin", APIKeyAuth(adminAPIKey), RateLimitMiddleware(5, 10))
  ```

### Step 3: Verify Green

- 运行 Task 003 创建的测试，验证全部通过
- **Verification**: `go test ./internal/httpx/... -run "TestNewServer_RateLimit" -v`

## Verification Commands

```bash
go test ./internal/httpx/... -run "TestNewServer_RateLimit" -v
go build ./cmd/server/...
```

## Success Criteria

- Task 003 的测试全部通过（Green）
- `cmd/server` 编译成功
