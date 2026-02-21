# Task 020: Test rate limiting

**depends-on**: (none)

## Description

Write tests for per-IP rate limiting middleware. Tests should verify that requests exceeding the rate limit receive 429 responses, different IPs have independent limits, and limits recover after the window passes.

## Execution Context

**Task Number**: 020 of 032
**Phase**: Security Middleware
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- "单个 IP 超过请求速率限制返回 429"
- "不同 IP 各自独立计数"
- "速率限制窗口过后恢复正常"

## Files to Modify/Create

- Create: `internal/httpx/middleware_ratelimit_test.go`

## Steps

### Step 1: Verify Scenarios

- Confirm all 3 rate limiting scenarios exist in Feature 2

### Step 2: Implement Tests (Red)

- Create `internal/httpx/middleware_ratelimit_test.go` with:
  - `TestRateLimit_ExceedsLimit_Returns429` — send requests from same IP exceeding the limit; verify 429 status and response body contains "RATE_LIMITED" code
  - `TestRateLimit_DifferentIPs_Independent` — send requests from two different IPs; each below limit; all should return 200
  - `TestRateLimit_RecoveryAfterWindow` — exceed limit, wait for token replenishment, send another request; should return 200
  - `TestRateLimit_ResponseContainsRetryAfter` — 429 response should include `Retry-After` header
- Tests need to configure a low rate limit (e.g., 5 req/s) for fast testing
- Simulate different IPs using `X-Forwarded-For` header or by setting `RemoteAddr` on the request
- Use `golang.org/x/time/rate` (already in go.mod)
- **Verification**: Tests FAIL (middleware doesn't exist)

## Verification Commands

```bash
go test ./internal/httpx/ -run TestRateLimit -v
```

## Success Criteria

- All 3 BDD scenarios covered
- Tests use configurable rate limits for test speed
- Different IPs correctly isolated
- Recovery behavior tested
