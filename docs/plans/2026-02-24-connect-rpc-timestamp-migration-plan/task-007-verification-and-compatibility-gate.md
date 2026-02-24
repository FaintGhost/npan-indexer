# Task 007: verification and compatibility gate

**depends-on**: task-004-green-backend-progress-timestamp-mapping, task-006-green-frontend-timestamp-consumer-adapter

## Description

执行全链路收口验证，并确认“旧字段仍在、存储结构未改、兼容边界达成”。

## Execution Context

**Task Number**: 007 of 007  
**Phase**: Verification  
**Prerequisites**: Task 004、Task 006 已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `进度持久化结构保持兼容`  
**Scenario**: `生成链路与回归验证通过`

## Files to Modify/Create

- Modify: `tasks/todo.md`（回填执行结果）

## Steps

### Step 1: Contract & Generation

- 执行 `buf lint`、`buf generate`，确认生成产物完整。

### Step 2: Runtime Regression

- 执行 Connect/Go 全量测试。
- 若涉及前端改动，执行关键 Vitest 套件。

### Step 3: Compatibility Gate

- 检查 proto：旧 `int64` 字段仍保留。
- 检查模型与存储：未强制迁移为 `Timestamp` 持久化。
- 在 `tasks/todo.md` 记录验证结果与后续清理建议。

## Verification Commands

```bash
XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf lint
XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf generate
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -count=1
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./... -count=1
cd web && bun vitest run
git diff --check
```

## Success Criteria

- 合同生成、后端、前端回归均通过。
- 兼容门槛检查有记录，且旧字段仍可用。
