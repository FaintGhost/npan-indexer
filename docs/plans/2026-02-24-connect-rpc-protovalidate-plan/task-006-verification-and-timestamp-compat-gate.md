# Task 006: Verification and timestamp compatibility gate

**depends-on**: task-004-green-search-pagination-proto-validation-rules, task-005-green-noop-and-business-guard-regression

## Description

执行本批次回归收口，并显式验证“未引入 Timestamp 契约变化”的兼容性守门条件。

## Execution Context

**Task Number**: 006 of 006  
**Phase**: Verification  
**Prerequisites**: Task 004、Task 005 已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`  
**Scenario**: `本批次不引入 Timestamp 契约变化`

## Files to Modify/Create

- Modify: `tasks/todo.md`（回填执行结果）

## Steps

### Step 1: Contract & Generation Verification

- 运行 `buf lint` / `buf generate`。
- 执行 `git diff --check`，确认无格式问题。

### Step 2: Runtime Regression Verification

- 执行 Connect 相关与全量 Go 测试，确认无回归。

### Step 3: Timestamp Compatibility Gate

- 检查 `proto/npan/v1/api.proto`，确认本批次未将现有 `int64` 时间字段替换为 `google.protobuf.Timestamp`。
- 将该守门结果写入 `tasks/todo.md` 的 Review 区域。

## Verification Commands

```bash
buf lint
buf generate
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect|Routes|Health|Admin' -count=1
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./... -count=1
git diff --check
```

## Success Criteria

- 回归命令全部通过。
- `Timestamp` 暂缓边界被显式验证并记录。
- 执行结果已回填 `tasks/todo.md`。
