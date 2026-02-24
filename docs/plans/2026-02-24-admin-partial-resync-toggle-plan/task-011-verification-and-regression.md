# Task 011: End-to-end verification and regression suite

**depends-on**: task-002-green-backend-inspect-roots-api-and-contract.md, task-004-green-backend-catalog-preserve-impl.md, task-006-green-frontend-inspect-and-autoselect-impl.md, task-008-green-frontend-running-lock-and-guard-impl.md, task-010-green-frontend-catalog-fallback-impl.md

## Description

执行完整验证，确认新交互与后端语义在单测、契约与回归链路全部通过。

## Execution Context

**Task Number**: 011 of 011  
**Phase**: Verification  
**Prerequisites**: 所有 Green 任务完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: 覆盖所有场景（局部补同步、拉取解耦、运行锁、force_rebuild 互斥、catalog 回退）

## Files to Modify/Create

- Modify: `tasks/todo.md`（回填验证结果）
- Modify: `tasks/lessons.md`（若出现用户纠正或回归问题）

## Steps

### Step 1: Contract and Generation Check

- 执行 OpenAPI 相关生成命令并确认无未预期差异。

### Step 2: Targeted Tests

- 运行本次新增/修改的后端与前端定向测试。

### Step 3: Full Regression

- 运行 `go test ./...` 与 `cd web && bun vitest run`。
- 若环境允许，执行 smoke + e2e 流程。

### Step 4: Result Documentation

- 在 `tasks/todo.md` 记录验证命令、通过/失败结果、残余风险。

## Verification Commands

```bash
go generate ./api/...
cd web && bun run generate
go test ./internal/httpx ./internal/service -count=1
cd web && bun vitest run src/components/admin-page.test.tsx src/components/sync-progress-display.test.tsx src/hooks/use-sync-progress.test.ts
go test ./...
cd web && bun vitest run
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
```

## Success Criteria

- 关键行为测试全部通过，且红绿链路可回溯。
- OpenAPI 与生成产物一致。
- 回归测试无新增失败。
- `tasks/todo.md` 有完整验证记录。
