# Task 004: GREEN backend progress timestamp mapping

**depends-on**: task-003-red-backend-connect-progress-timestamp-tests

## Description

实现后端双字段映射：在 Connect progress 输出中同时填充 `int64` 与 `Timestamp` 字段。

## Execution Context

**Task Number**: 004 of 007  
**Phase**: Implementation (Green)  
**Prerequisites**: Task 003 已失败并定位缺口

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `Connect progress 返回新 Timestamp 字段`

## Files to Modify/Create

- Modify: `internal/httpx/connect_admin.go`
- Modify: `internal/httpx/connect_admin_test.go`

## Steps

### Step 1: Implement Mapping

- 在 DTO 转换路径补充 `int64 -> Timestamp` 的 sidecar 填充。
- 保证零值/空值场景有一致策略（避免无效时间）。

### Step 2: Verify Green

- 运行 Task 003 测试，确认新旧字段都正确且时刻一致。

### Step 3: Regression Check

- 回归已有 Admin Connect 测试，确保错误码与业务防线行为不变。

## Verification Commands

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*(Admin|Timestamp)' -count=1
```

## Success Criteria

- Task 003 转绿。
- 现有 Admin Connect 行为无回归。
