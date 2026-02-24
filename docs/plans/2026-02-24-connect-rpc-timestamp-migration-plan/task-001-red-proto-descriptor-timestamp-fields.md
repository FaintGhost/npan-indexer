# Task 001: RED proto descriptor timestamp fields

## Description

先建立失败测试，验证当前消息 descriptor 尚未包含 `*_ts` 字段，作为后续 schema 变更的 Red 基线。

## Execution Context

**Task Number**: 001 of 007  
**Phase**: Testing (Red)  
**Prerequisites**: 无

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `Connect progress 返回新 Timestamp 字段`

## Files to Modify/Create

- Create: `internal/httpx/connect_timestamp_descriptor_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD 场景要求：迁移后 descriptor 必须可见 `*_ts` 字段。

### Step 2: Implement Test (Red)

- 编写 descriptor 级测试（基于 protoreflect）检查目标消息字段存在性与类型期望。
- 当前阶段应失败，证明字段尚未引入。

### Step 3: Verify Failure

- 运行指定测试，确认失败原因是字段缺失（非编译错误）。

## Verification Commands

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Timestamp.*Descriptor' -count=1
```

## Success Criteria

- 测试稳定失败（Red）。
- 失败信息指向 `*_ts` 字段缺失。
