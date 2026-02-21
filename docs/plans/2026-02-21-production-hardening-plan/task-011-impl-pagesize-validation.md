# Task 011: Implement pageSize validation

**depends-on**: task-010

## Description

Implement pageSize validation in both the handler layer (reject > 100 with 400) and query service layer (cap at 100 as defense-in-depth). Add a `validatePageSize` function and a `maxPageSize` constant.

## Execution Context

**Task Number**: 011 of 032
**Phase**: Input Validation
**Prerequisites**: pageSize tests (task-010) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: Feature 3, pageSize scenarios

## Files to Modify/Create

- Create: `internal/httpx/validation.go`
- Modify: `internal/httpx/handlers.go` — add pageSize validation call in search handlers
- Modify: `internal/search/query_service.go` — add defense-in-depth cap on pageSize

## Steps

### Step 1: Create validation.go

- Create `internal/httpx/validation.go` with:
  - `const maxPageSize int64 = 100`
  - `func validatePageSize(pageSize int64) error` — returns error if pageSize <= 0 or > maxPageSize

### Step 2: Integrate into handlers

- In `LocalSearch` and `RemoteSearch` handlers in `handlers.go`, after parsing `page_size` query parameter, call `validatePageSize` and return 400 with `writeErrorResponse` on error

### Step 3: Add query service cap

- In `internal/search/query_service.go`, in the normalization logic, cap `PageSize` at 100 if it exceeds

### Step 4: Verify (Green)

- Run tests from task-010
- **Verification**: `go test ./internal/httpx/ -run TestValidatePageSize -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestValidatePageSize -v
go test ./internal/search/ -v
go test ./... -count=1
```

## Success Criteria

- All pageSize tests pass
- Handler rejects pageSize > 100 with 400
- Query service silently caps at 100 as defense-in-depth
