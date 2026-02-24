# Task 003: RED search pagination validation hit tests

**depends-on**: task-002-green-admin-proto-validation-rules

## Description

为分页类请求建立第二组失败测试，覆盖 Search/App 请求在 proto 规则缺失时无法由 interceptor 前置拦截的问题。

## Execution Context

**Task Number**: 003 of 006  
**Phase**: Testing (Red)  
**Prerequisites**: Task 002 已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`  
**Scenario**: `命中 proto 规则时由 validation interceptor 返回 invalid_argument`

## Files to Modify/Create

- Modify: `internal/httpx/connect_validation_interceptor_test.go`

## Steps

### Step 1: Verify Scenario

- 确认分页类场景也应由 schema 校验前置拦截，而不是仅依赖 handler 内部判断。

### Step 2: Implement Test (Red)

- 新增/扩展 interceptor 测试用例（Search/App 维度）：
  - 构造 `LocalSearchRequest` 或 `AppSearchRequest` 的非法分页参数。
  - 断言返回 `CodeInvalidArgument`。
  - 断言 fake `next` handler 未执行。
- 继续使用 test doubles，禁止真实外部依赖。

### Step 3: Verify Failure

- 在尚未补分页 proto 规则前运行测试，确认失败且失败原因聚焦为“规则未命中”。

## Verification Commands

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Validation.*Pagination' -count=1
```

## Success Criteria

- 新增分页校验命中测试稳定失败（Red）。
- 失败原因可直接映射到规则缺失。
