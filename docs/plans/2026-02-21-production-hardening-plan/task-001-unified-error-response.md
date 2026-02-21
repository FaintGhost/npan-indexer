# Task 001: Create unified error response types

## Description

Create the foundation error response infrastructure that will be used by all subsequent tasks. This includes the `ErrorResponse` struct, predefined error code constants, and a `writeErrorResponse` helper function. Also create a global HTTP error handler that catches panics and unhandled errors.

## Execution Context

**Task Number**: 001 of 032
**Phase**: Foundation
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 4 "错误响应格式" — Scenario "错误响应使用统一 JSON 结构"

## Files to Modify/Create

- Create: `internal/httpx/errors.go`

## Steps

### Step 1: Verify Scenario

- Ensure Feature 4 Scenario "错误响应使用统一 JSON 结构" exists in BDD specs
- Review the `ErrorResponse` struct design in `../2026-02-21-production-hardening-design/architecture.md` Section 4

### Step 2: Create ErrorResponse struct and helpers

- Create `internal/httpx/errors.go` with:
  - `ErrorResponse` struct with `Code`, `Message`, `RequestID` fields (JSON tags)
  - Error code constants: `ErrCodeUnauthorized`, `ErrCodeBadRequest`, `ErrCodeNotFound`, `ErrCodeConflict`, `ErrCodeRateLimited`, `ErrCodeInternalError`
  - `writeErrorResponse(c, status, code, message)` helper that extracts request ID from context and returns JSON response
  - `customHTTPErrorHandler` that catches echo HTTP errors and unhandled errors, logs full error internally via slog, returns sanitized `ErrorResponse` to client

### Step 3: Write unit tests

- Create tests in `internal/httpx/errors_test.go`:
  - Test that `writeErrorResponse` returns correct JSON structure with code, message, request_id
  - Test that `customHTTPErrorHandler` returns generic message for unhandled errors
  - Test that error responses never contain stack traces or internal paths

### Step 4: Verify

- Run tests and ensure they pass
- **Verification**: `go test ./internal/httpx/ -run TestErrorResponse -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestErrorResponse -v
go test ./internal/httpx/ -run TestCustomHTTPErrorHandler -v
```

## Success Criteria

- `ErrorResponse` struct exists with proper JSON tags
- Error code constants are defined
- `writeErrorResponse` returns well-formed JSON with request_id
- `customHTTPErrorHandler` logs full error but returns generic message
- All tests pass
