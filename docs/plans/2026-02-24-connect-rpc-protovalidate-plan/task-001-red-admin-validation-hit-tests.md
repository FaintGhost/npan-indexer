# Task 001: RED admin validation hit tests

## Description

为 Admin 相关输入约束建立失败测试，先证明“当前尚未通过 proto 规则拦截”的缺口存在，再进入规则实现阶段。

## Execution Context

**Task Number**: 001 of 006  
**Phase**: Testing (Red)  
**Prerequisites**: 无

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`  
**Scenario**: `命中 proto 规则时由 validation interceptor 返回 invalid_argument`

## Files to Modify/Create

- Create: `internal/httpx/connect_validation_interceptor_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD 场景要求：命中 proto 规则时必须在 interceptor 层返回 `CodeInvalidArgument`，且不进入业务 handler。

### Step 2: Implement Test (Red)

- 新增 interceptor 级测试用例（Admin 维度）：
  - 选择 `StartSyncRequest` 中当前未由 proto 规则覆盖、且应受约束的字段（例如 `root_workers`、`progress_every`）。
  - 构造非法请求，断言返回 `CodeInvalidArgument`。
  - 通过 fake `next` handler 记录调用次数，断言命中规则时 `next` 不应被调用。
- 使用 test doubles 隔离依赖：
  - 只测试 interceptor，不访问真实网络/存储/外部服务。

### Step 3: Verify Failure

- 运行新增测试，确认在尚未添加 proto 规则前失败，且失败原因是行为断言不满足（非编译错误）。

## Verification Commands

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*Validation.*Admin' -count=1
```

## Success Criteria

- 新增 Admin 校验命中测试稳定失败（Red）。
- 失败指向“规则尚未在 proto 中定义”。
