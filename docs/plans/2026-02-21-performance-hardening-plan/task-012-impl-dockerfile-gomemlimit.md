# Task 012: Dockerfile GOMEMLIMIT

**depends-on**: (none)

## Description

在 Dockerfile 中添加 `GOMEMLIMIT` 和 `GOGC` 环境变量配置，为 2C2G 同机部署优化 Go 运行时内存管理。

## Execution Context

**Task Number**: 012 of 012
**Phase**: Runtime Tuning
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 5.1 (Dockerfile 包含 GOMEMLIMIT), Scenario 5.2 (运行时可覆盖)

## Files to Modify/Create

- Modify: `Dockerfile` — 添加 ENV GOMEMLIMIT 和 GOGC

## Steps

### Step 1: Add environment variables

- 在 Dockerfile 的生产镜像阶段（`FROM alpine:3.21` 之后，`ENTRYPOINT` 之前），添加：
  ```dockerfile
  ENV GOMEMLIMIT=512MiB
  ENV GOGC=100
  ```
- 使用 `ENV` 而非 `ARG`，允许运行时通过 `docker run -e GOMEMLIMIT=768MiB` 覆盖

### Step 2: Verify Dockerfile builds

- 运行 `docker build` 验证 Dockerfile 语法正确
- **Verification**: `docker build -t npan-test .`（如环境支持）
- 或至少验证 Dockerfile 语法无误

## Verification Commands

```bash
# 验证 Dockerfile 语法（grep 检查）
grep -q "GOMEMLIMIT" Dockerfile && echo "OK" || echo "MISSING"
grep -q "GOGC" Dockerfile && echo "OK" || echo "MISSING"
```

## Success Criteria

- Dockerfile 包含 `ENV GOMEMLIMIT=512MiB`
- Dockerfile 包含 `ENV GOGC=100`
- `ENV` 放在 `ENTRYPOINT` 之前
