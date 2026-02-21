# Task 017: Implement error response sanitization

**depends-on**: task-016, task-001

## Description

Replace all `err.Error()` in handler error responses with generic user-facing messages. Log the full error server-side with slog. Use the `writeErrorResponse` helper from task-001 for all error responses.

## Execution Context

**Task Number**: 017 of 032
**Phase**: Error Handling
**Prerequisites**: Error response types (task-001) and sanitization tests (task-016) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: Feature 4, Scenarios 2-4

## Files to Modify/Create

- Modify: `internal/httpx/handlers.go` — replace all `writeError(c, status, err.Error())` calls

## Steps

### Step 1: Replace error responses in handlers

- Apply the mapping from best-practices.md #8:
  - Token handler: `err.Error()` → log error, return "认证失败，请检查凭据"
  - RemoteSearch handler: `err.Error()` → log error, return "搜索请求失败，请稍后重试"
  - DownloadURL handler: `err.Error()` → log error, return "获取下载链接失败"
  - LocalSearch handler: `err.Error()` → log error, return "搜索服务暂不可用"
  - StartFullSync handler: `err.Error()` → log error, return "启动同步失败"
  - GetFullSyncProgress handler: `err.Error()` → log error, return "无法读取同步进度"
- Each replacement should: `slog.Error("描述", "error", err, "handler", "HandlerName")` then `return writeErrorResponse(...)`
- Replace old `writeError` function calls with `writeErrorResponse` from errors.go

### Step 2: Verify (Green)

- Run tests from task-016
- **Verification**: `go test ./internal/httpx/ -run TestErrorSanitization -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestErrorSanitization -v
go test ./internal/httpx/ -v
go test ./... -count=1
```

## Success Criteria

- All error sanitization tests pass
- No handler returns raw `err.Error()` to clients
- All errors are logged server-side with handler context
- Uses `writeErrorResponse` consistently
