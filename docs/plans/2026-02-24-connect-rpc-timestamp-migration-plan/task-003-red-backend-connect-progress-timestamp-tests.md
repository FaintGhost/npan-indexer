# Task 003: RED backend connect progress timestamp tests

**depends-on**: task-002-green-proto-add-timestamp-sidecar-fields

## Description

先写后端映射失败测试，证明当前 Connect progress 响应尚未正确填充 `*_ts` 字段。

## Execution Context

**Task Number**: 003 of 007  
**Phase**: Testing (Red)  
**Prerequisites**: Task 002 已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `Connect progress 返回新 Timestamp 字段`

## Files to Modify/Create

- Modify: `internal/httpx/connect_admin_test.go`

## Steps

### Step 1: Verify Scenario

- 明确断言目标：GetSyncProgress 响应需同时具备新旧时间字段且时刻一致。

### Step 2: Implement Test (Red)

- 增加 Connect 集成测试：
  - 构造有进度状态的场景；
  - 断言 `*_ts` 非空且与旧字段可对齐到同一 UTC 时刻。
- 当前实现下测试应失败（`*_ts` 尚未填充）。

### Step 3: Verify Failure

- 运行测试，确认失败原因为映射缺失或不一致。

## Verification Commands

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Admin.*Timestamp' -count=1
```

## Success Criteria

- 测试稳定失败（Red）。
- 失败信息指向后端映射缺口。
