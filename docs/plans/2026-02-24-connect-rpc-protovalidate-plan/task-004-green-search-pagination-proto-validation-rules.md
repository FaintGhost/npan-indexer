# Task 004: GREEN search pagination proto validation rules

**depends-on**: task-003-red-search-pagination-validation-hit-tests

## Description

在 `.proto` 中为 Search/App 分页请求补齐规则，完成第二组 Red -> Green 转换。

## Execution Context

**Task Number**: 004 of 006  
**Phase**: Implementation (Green)  
**Prerequisites**: Task 003 已完成并稳定失败

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`  
**Scenario**: `命中 proto 规则时由 validation interceptor 返回 invalid_argument`

## Files to Modify/Create

- Modify: `proto/npan/v1/api.proto`
- Modify: `gen/go/npan/v1/api.pb.go`（generated）
- Modify: `gen/ts/npan/v1/api_pb.ts`（generated）
- Modify: `gen/ts/npan/v1/api_connect.ts`（如生成产物变更）

## Steps

### Step 1: Add Pagination Rules

- 在 `LocalSearchRequest`、`AppSearchRequest` 等分页字段上增加范围规则（如 `page >= 1`、`page_size` 上下界）。
- 仅增加与现有行为一致的约束，避免改变既有业务语义。

### Step 2: Regenerate Artifacts

- 执行 `buf lint` 与 `buf generate`，确认契约和生成产物一致。

### Step 3: Verify Green

- 重新执行 Task 003 用例，确认非法分页参数可在 interceptor 层被拦截。

## Verification Commands

```bash
buf lint
buf generate
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Validation.*Pagination' -count=1
```

## Success Criteria

- 分页相关 proto 规则已落地并通过生成校验。
- Task 003 用例由 Red 转 Green。
