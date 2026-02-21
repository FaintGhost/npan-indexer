# Task 001: Test config timeout fields

**depends-on**: (none)

## Description

为 Config 结构体新增 HTTP server 超时配置字段创建测试。测试应验证：默认值正确、环境变量可覆盖默认值。

## Execution Context

**Task Number**: 001 of 012
**Phase**: Foundation
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 1.1 (服务启动时配置超时参数), Scenario 1.2 (超时参数可通过配置覆盖)

## Files to Modify/Create

- Create: `internal/config/config_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD specs 中 Scenario 1.1 和 1.2 存在

### Step 2: Implement Test (Red)

- 创建 `internal/config/config_test.go`
- 测试 `TestLoad_DefaultTimeouts`: 不设置任何超时环境变量时调用 `Load()`，验证 4 个超时字段等于默认值（ReadHeaderTimeout=5s, ReadTimeout=10s, WriteTimeout=30s, IdleTimeout=120s）
- 测试 `TestLoad_CustomTimeouts`: 设置 `SERVER_READ_HEADER_TIMEOUT=8s`、`SERVER_READ_TIMEOUT=15s`、`SERVER_WRITE_TIMEOUT=60s`、`SERVER_IDLE_TIMEOUT=180s`，调用 `Load()`，验证字段等于自定义值
- 注意：测试需要在 setUp 中 `os.Setenv` 并在 tearDown 中 `os.Unsetenv`
- **Verification**: `go test ./internal/config/... -run TestLoad_DefaultTimeouts` 和 `TestLoad_CustomTimeouts` 应编译失败（Config 结构体还没有这些字段）

### Step 3: Verify Red

- 运行测试确认因缺少字段而编译失败

## Verification Commands

```bash
go test ./internal/config/... -run "TestLoad_DefaultTimeouts|TestLoad_CustomTimeouts" -v
```

## Success Criteria

- 测试代码编译失败（Red），因为 Config 没有 timeout 字段
- 测试逻辑正确映射 BDD Scenario 1.1 和 1.2
