# Task 005: GREEN no-op and business guard regression

**depends-on**: task-002-green-admin-proto-validation-rules

## Description

补齐“无规则 no-op”与“业务语义防线保留”的回归测试，确保 schema 增强不会破坏既有行为边界。

## Execution Context

**Task Number**: 005 of 006  
**Phase**: Regression  
**Prerequisites**: Task 002 已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`  
**Scenario**: `未配置规则的消息保持 no-op 行为`  
**Scenario**: `业务语义校验继续保留在后端防线`

## Files to Modify/Create

- Modify: `internal/httpx/connect_validation_interceptor_test.go`
- Modify: `internal/httpx/connect_admin_test.go`

## Steps

### Step 1: Add No-op Regression Test

- 为未配置规则的消息（例如 `GetSyncProgressRequest`）新增回归断言：
  - 请求不应被 validation interceptor 错误拦截为 `invalid_argument`。
  - 请求应继续进入既有 handler 路径并返回原有语义错误码。

### Step 2: Add Business Guard Regression Test

- 强化 `force_rebuild + scoped roots` 互斥场景断言：
  - 错误码与错误文案保持既有业务语义。
  - 规则增强后该防线仍由业务层负责。

### Step 3: Verify Regression

- 运行 Admin / Validation 相关测试，确认行为稳定。

## Verification Commands

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*(Validation|Admin).*' -count=1
```

## Success Criteria

- no-op 行为有测试保护。
- 业务语义防线有测试保护，且未被 schema 规则替代。
