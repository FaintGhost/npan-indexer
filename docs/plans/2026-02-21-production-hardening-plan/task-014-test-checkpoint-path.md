# Task 014: Test checkpoint path validation

**depends-on**: (none)

## Description

Write tests for checkpoint path validation. The checkpoint_template parameter in sync start requests must be restricted to the `data/checkpoints` directory, reject absolute paths, and prevent path traversal attacks.

## Execution Context

**Task Number**: 014 of 032
**Phase**: Input Validation
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 3 — Scenario Outline "checkpoint_template 路径遍历攻击被拒绝"

## Files to Modify/Create

- Modify: `internal/httpx/validation_test.go` (append)

## Steps

### Step 1: Verify Scenario

- Confirm 3 path traversal examples exist in the scenario outline

### Step 2: Implement Tests (Red)

- Add to `internal/httpx/validation_test.go`:
  - `TestValidateCheckpointTemplate_PathTraversal_RejectsRelative` — "../../../etc/passwd" → error
  - `TestValidateCheckpointTemplate_AbsolutePath_Rejects` — "/etc/shadow" → error
  - `TestValidateCheckpointTemplate_TraversalInCheckpoints_Rejects` — "data/checkpoints/../../secrets" → error
  - `TestValidateCheckpointTemplate_ValidPath_Accepts` — "data/checkpoints/my-checkpoint" → ok
  - `TestValidateCheckpointTemplate_EmptyString_Accepts` — "" → ok (means default)
- Tests call `validateCheckpointTemplate(template string) error`
- **Verification**: Tests FAIL

## Verification Commands

```bash
go test ./internal/httpx/ -run TestValidateCheckpointTemplate -v
```

## Success Criteria

- All 3 BDD scenario outline examples covered
- Tests verify specific error messages
- Empty template is accepted (uses default)
