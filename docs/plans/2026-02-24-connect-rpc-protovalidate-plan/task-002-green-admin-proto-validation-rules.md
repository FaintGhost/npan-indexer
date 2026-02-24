# Task 002: GREEN admin proto validation rules

**depends-on**: task-001-red-admin-validation-hit-tests

## Description

为 Admin 请求补齐 `protovalidate` 规则并通过生成校验，让 Task 001 的失败用例转绿。

## Execution Context

**Task Number**: 002 of 006  
**Phase**: Implementation (Green)  
**Prerequisites**: Task 001 已完成且失败原因明确

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`  
**Scenario**: `命中 proto 规则时由 validation interceptor 返回 invalid_argument`

## Files to Modify/Create

- Modify: `buf.yaml`
- Modify: `proto/npan/v1/api.proto`
- Modify: `gen/go/npan/v1/api.pb.go`（generated）
- Modify: `gen/ts/npan/v1/api_pb.ts`（generated）
- Modify: `gen/ts/npan/v1/api_connect.ts`（如生成产物变更）

## Steps

### Step 1: Add Contract Dependency

- 在 `buf.yaml` 增加 protovalidate 依赖，确保可导入 `buf/validate/validate.proto`。

### Step 2: Add Proto Rules

- 在 `StartSyncRequest`、`InspectRootsRequest` 等 Admin 高价值请求字段添加增量规则：
  - 非空/范围/正整数约束（按设计文档范围）。
- 保持业务语义规则（如 `force_rebuild + scoped`）仍在 handler/service 防线，不迁移到 schema。

### Step 3: Regenerate Artifacts

- 执行 `buf lint` 与 `buf generate`，更新生成文件。

### Step 4: Verify Green

- 重新执行 Task 001 用例，确认由 interceptor 返回 `CodeInvalidArgument` 且 `next` 不被调用。

## Verification Commands

```bash
buf lint
buf generate
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Validation.*Admin' -count=1
```

## Success Criteria

- Admin 维度 proto 规则落地且生成通过。
- Task 001 用例从 Red 转 Green。
