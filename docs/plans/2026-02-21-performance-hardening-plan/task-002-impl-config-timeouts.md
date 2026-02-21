# Task 002: Implement config timeouts and http.Server wiring

**depends-on**: task-001

## Description

在 Config 结构体中添加 4 个 `time.Duration` 超时字段，在 `Load()` 中从环境变量读取并设置默认值。同时在 `cmd/server/main.go` 中将这些字段设置到 `http.Server` 上。

## Execution Context

**Task Number**: 002 of 012
**Phase**: Foundation
**Prerequisites**: Task 001 测试已创建

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 1.1 (服务启动时配置超时参数), Scenario 1.2 (超时参数可通过配置覆盖)

## Files to Modify/Create

- Modify: `internal/config/config.go` — 添加 4 个 timeout 字段到 Config 结构体，添加 `readDuration` 辅助函数，在 `Load()` 中读取环境变量
- Modify: `cmd/server/main.go` — 在 `http.Server{}` 初始化中设置 ReadHeaderTimeout、ReadTimeout、WriteTimeout、IdleTimeout

## Steps

### Step 1: Add duration reader

- 在 `config.go` 中添加 `readDuration(key string, fallback time.Duration) time.Duration` 辅助函数
- 使用 `time.ParseDuration` 解析环境变量值（支持 "5s", "10s" 格式）

### Step 2: Add Config fields

- 在 Config 结构体中添加：`ServerReadHeaderTimeout`, `ServerReadTimeout`, `ServerWriteTimeout`, `ServerIdleTimeout`（类型 `time.Duration`）

### Step 3: Read from env in Load()

- 在 `Load()` 中添加 4 行读取：
  - `SERVER_READ_HEADER_TIMEOUT` 默认 5s
  - `SERVER_READ_TIMEOUT` 默认 10s
  - `SERVER_WRITE_TIMEOUT` 默认 30s
  - `SERVER_IDLE_TIMEOUT` 默认 120s

### Step 4: Wire into http.Server

- 在 `cmd/server/main.go` 的 `http.Server{}` 初始化中添加：
  - `ReadHeaderTimeout: cfg.ServerReadHeaderTimeout`
  - `ReadTimeout: cfg.ServerReadTimeout`
  - `WriteTimeout: cfg.ServerWriteTimeout`
  - `IdleTimeout: cfg.ServerIdleTimeout`

### Step 5: Verify Green

- 运行 Task 001 创建的测试，验证全部通过
- **Verification**: `go test ./internal/config/... -v`

## Verification Commands

```bash
go test ./internal/config/... -run "TestLoad_DefaultTimeouts|TestLoad_CustomTimeouts" -v
go build ./cmd/server/...
```

## Success Criteria

- Task 001 的测试全部通过（Green）
- `cmd/server` 编译成功
- `http.Server` 包含 4 个超时字段
