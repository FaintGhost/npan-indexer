# Task 013: Implement type parameter whitelist

**depends-on**: task-012

## Description

Implement type parameter whitelist validation in the handler and add quote-wrapping defense in meili_index.go as defense-in-depth against filter injection.

## Execution Context

**Task Number**: 013 of 032
**Phase**: Input Validation
**Prerequisites**: Type whitelist tests (task-012) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 3 — "type 参数只接受白名单值"

## Files to Modify/Create

- Modify: `internal/httpx/validation.go` — add `validateType` function
- Modify: `internal/httpx/handlers.go` — add type validation call in search handlers
- Modify: `internal/search/meili_index.go` — wrap type value in quotes as defense-in-depth

## Steps

### Step 1: Add validateType to validation.go

- Define `var allowedTypes = map[string]bool{"all": true, "file": true, "folder": true}`
- Implement `func validateType(typeParam string) error` — empty string is ok (default), otherwise must be in allowedTypes map

### Step 2: Integrate into handlers

- In search handlers (`LocalSearch`, `DemoSearch`), validate the `type` query parameter before passing to search service
- Return 400 with `writeErrorResponse` on invalid type

### Step 3: Defense-in-depth in meili_index.go

- In `internal/search/meili_index.go`, change the filter construction from `fmt.Sprintf("type = %s", params.Type)` to `fmt.Sprintf("type = '%s'", params.Type)` — add single quotes around the value

### Step 4: Verify (Green)

- Run tests from task-012
- **Verification**: `go test ./internal/httpx/ -run TestValidateType -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestValidateType -v
go test ./internal/search/ -v
go test ./... -count=1
```

## Success Criteria

- All type whitelist tests pass
- Injection strings are rejected at handler level
- Filter values are quote-wrapped at Meilisearch level
- Dual-layer defense implemented
