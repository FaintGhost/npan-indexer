# Task 002: GREEN proto add timestamp sidecar fields

**depends-on**: task-001-red-proto-descriptor-timestamp-fields

## Description

在 proto 中新增 Timestamp sidecar 字段并更新生成产物，使 Task 001 的字段存在性测试转绿。

## Execution Context

**Task Number**: 002 of 007  
**Phase**: Implementation (Green)  
**Prerequisites**: Task 001 已失败并定位到字段缺失

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `Connect progress 返回新 Timestamp 字段`

## Files to Modify/Create

- Modify: `proto/npan/v1/api.proto`
- Modify: `gen/go/npan/v1/api.pb.go`（generated）
- Modify: `gen/ts/npan/v1/api_pb.ts`（generated）

## Steps

### Step 1: Add Timestamp Sidecar Fields

- 为进度相关消息新增 `*_ts` 字段（`google.protobuf.Timestamp`）。
- 保留旧 `int64` 字段，避免破坏兼容。

### Step 2: Regenerate Artifacts

- 执行 `buf lint` 与 `buf generate` 更新生成代码。

### Step 3: Verify Green

- 重新执行 Task 001 测试，确认 descriptor 检查通过。

## Verification Commands

```bash
XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf lint
XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf generate
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Timestamp.*Descriptor' -count=1
```

## Success Criteria

- proto 与生成产物更新完成。
- Task 001 转绿。
