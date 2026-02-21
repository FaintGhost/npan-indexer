# Task 015: Implement checkpoint path validation

**depends-on**: task-014

## Description

Implement the `validateCheckpointTemplate` function that sanitizes checkpoint template paths using `filepath.Clean`, rejects absolute paths, path traversal, and paths outside `data/checkpoints`. Integrate into the sync start handler.

## Execution Context

**Task Number**: 015 of 032
**Phase**: Input Validation
**Prerequisites**: Checkpoint path tests (task-014) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 3 — "checkpoint_template 路径遍历攻击被拒绝"

## Files to Modify/Create

- Modify: `internal/httpx/validation.go` — add `validateCheckpointTemplate`
- Modify: `internal/httpx/handlers.go` — add validation call in `StartFullSync` handler

## Steps

### Step 1: Implement validateCheckpointTemplate

- In `validation.go`, implement `func validateCheckpointTemplate(template string) error`
- Use `filepath.Clean` to normalize the path
- Reject: absolute paths (`filepath.IsAbs`), paths containing `..`, paths not prefixed with `data/checkpoints`
- Empty string is valid (uses default)

### Step 2: Integrate into handler

- In `StartFullSync` handler, validate `checkpoint_template` before passing to sync manager
- Return 400 with `writeErrorResponse` on invalid path

### Step 3: Verify (Green)

- Run tests from task-014
- **Verification**: `go test ./internal/httpx/ -run TestValidateCheckpointTemplate -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestValidateCheckpointTemplate -v
go test ./... -count=1
```

## Success Criteria

- All checkpoint path tests pass
- Path traversal attacks blocked
- Uses `filepath.Clean` for normalization
