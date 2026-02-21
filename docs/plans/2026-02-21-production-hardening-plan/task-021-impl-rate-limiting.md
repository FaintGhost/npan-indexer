# Task 021: Implement rate limiting

**depends-on**: task-020, task-001

## Description

Implement per-IP token bucket rate limiting middleware using `golang.org/x/time/rate`. Each IP gets an independent rate limiter. Stale limiters are periodically cleaned up to prevent memory leaks.

## Execution Context

**Task Number**: 021 of 032
**Phase**: Security Middleware
**Prerequisites**: Rate limit tests (task-020) and error response types (task-001) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: Feature 2, all 3 scenarios

## Files to Modify/Create

- Create: `internal/httpx/middleware_ratelimit.go`

## Steps

### Step 1: Implement rate limit middleware

- Create `internal/httpx/middleware_ratelimit.go` with:
  - A `rateLimiterStore` struct that maps IP → `*rate.Limiter` with a sync.Mutex
  - A `getLimiter(ip string)` method that returns (or creates) a limiter for the IP
  - A cleanup goroutine or last-seen tracking to evict stale entries
  - `func RateLimitMiddleware(rps float64, burst int) echo.MiddlewareFunc` — extracts client IP, checks rate limit, returns 429 with `Retry-After` header on limit exceeded
- Extract client IP from `c.RealIP()` (Echo's built-in method that checks X-Forwarded-For)
- On 429, return `writeErrorResponse` with `ErrCodeRateLimited` and set `Retry-After` header

### Step 2: Verify (Green)

- Run tests from task-020
- **Verification**: `go test ./internal/httpx/ -run TestRateLimit -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestRateLimit -v
go test -race ./internal/httpx/ -run TestRateLimit -v
```

## Success Criteria

- All rate limit tests pass
- Race detector passes (concurrent access safe)
- Memory doesn't grow unboundedly (stale entries cleaned)
- Uses `golang.org/x/time/rate` (already a dependency)
