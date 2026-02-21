# Task 016: Test error response sanitization

**depends-on**: task-009

## Description

Write tests verifying that error responses from handlers do not leak internal implementation details. Tests should verify that Meilisearch errors, token failures, and internal panics all return generic user-facing messages while the actual error is logged server-side.

## Execution Context

**Task Number**: 016 of 032
**Phase**: Error Handling
**Prerequisites**: Route restructure (task-009) should be complete since tests target the restructured handlers

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- "500 错误不泄露堆栈"
- "Meilisearch 错误不直接透传"
- "Token 获取失败不泄露 Client Secret"

## Files to Modify/Create

- Create: `internal/httpx/error_sanitization_test.go`

## Steps

### Step 1: Verify Scenarios

- Confirm all 3 error sanitization scenarios exist in Feature 4

### Step 2: Implement Tests (Red)

- Create `internal/httpx/error_sanitization_test.go` with:
  - `TestErrorSanitization_InternalError_NoStack` — trigger an internal error; response body should NOT contain file paths, line numbers, or Go stack traces
  - `TestErrorSanitization_MeiliError_NoDetails` — when Meilisearch returns error, API response message should be "搜索服务暂不可用", NOT contain "meilisearch" or "meili"
  - `TestErrorSanitization_TokenError_NoSecret` — when token endpoint fails, response should NOT contain client_secret value; message should be "认证失败，请检查凭据"
  - `TestErrorSanitization_AllErrors_HaveUnifiedFormat` — all error responses have `code` and `message` fields; do NOT have `stack`, `trace`, `debug` fields
- Tests require mock/stub of query service and npan client to inject errors
- Use interface-based test doubles for external dependencies
- **Verification**: Tests should show current handlers leaking errors (Red state)

## Verification Commands

```bash
go test ./internal/httpx/ -run TestErrorSanitization -v
```

## Success Criteria

- Tests verify absence of internal info in responses
- Tests use test doubles for external dependencies
- Each test maps to a specific BDD scenario
