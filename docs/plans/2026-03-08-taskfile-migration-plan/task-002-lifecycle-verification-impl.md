# Task 002: [IMPL] smoke / E2E 生命周期任务 (GREEN)

**depends-on**: task-002-lifecycle-verification-test.md

## Description

补齐 `stack:ci:*`、`smoke:run`、`e2e:run`、`verify:smoke`、`verify:e2e`、`verify:all` 等生命周期任务，使 Task 能完整承接当前 Makefile 的 smoke / E2E 编排能力，并保留默认端口、自动清理与全量回归入口语义。

## Execution Context

**Task Number**: 004 of 008
**Phase**: Lifecycle Verification
**Prerequisites**: `task-002-lifecycle-verification-test.md` 已完成并稳定失败

## BDD Scenario

```gherkin
Scenario: verify:smoke 从冷启动运行 smoke 并自动清理
  Given 本地尚未启动 docker-compose.ci.yml 对应的服务
  When 开发者执行 "task verify:smoke"
  Then 任务应先启动 CI compose 栈并等待服务就绪
  And 任务应以默认 BASE_URL=http://localhost:11323 与 METRICS_URL=http://localhost:19091 执行 smoke_test.sh
  And 无论 smoke 成功还是失败都应执行 compose down --volumes

Scenario: verify:e2e 先运行 smoke 再执行 Playwright，并在结束后清理环境
  Given 本地尚未启动 docker-compose.ci.yml 对应的服务
  When 开发者执行 "task verify:e2e"
  Then 任务应启动 CI compose 栈
  And 任务应先以默认 BASE_URL=http://localhost:11323 与 METRICS_URL=http://localhost:19091 执行 smoke_test.sh
  And 再执行 Playwright 容器验证
  And 无论验证成功还是失败都应执行 compose down --volumes

Scenario: verify:all 作为全量回归总入口
  Given 开发者需要一次性完成提交前全量回归
  When 开发者执行 "task verify:all"
  Then 任务应先完成快速验证
  And 再执行完整的 E2E 链路
  And 最终结果应对任一失败步骤返回非零状态
```

**Spec Source**: `../2026-03-08-taskfile-migration-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `Taskfile.yml`
- Create or Modify: `tests/taskfile/taskfile-lifecycle.test.sh` 或等价验证脚本

## Steps

### Step 1: Add Support Tasks for Compose Lifecycle

- 在 `Taskfile.yml` 中补齐辅助任务：
  - `stack:ci:up`
  - `stack:ci:logs`
  - `stack:ci:down`
  - `smoke:run`
  - `e2e:run`
- 这些任务只负责复用现有执行器，不重写 smoke 脚本或 Playwright 命令本身。

### Step 2: Implement Public Lifecycle Tasks

- 实现 `verify:smoke`，使其从冷启动执行 compose up、smoke 脚本与自动清理。
- 实现 `verify:e2e`，使其从冷启动执行 compose up、Playwright 容器验证与自动清理。
- 实现 `verify:all`，使其按顺序执行快速验证与完整 E2E 链路，而不是依赖并行 `deps` 表达顺序。

### Step 3: Preserve Cleanup and Port Semantics

- 为本地聚合任务保留“失败也清理”的语义。
- 明确 `verify:smoke` 使用 `BASE_URL=http://localhost:11323` 与 `METRICS_URL=http://localhost:19091`。
- 确保 `verify:e2e` 继续复用 `docker-compose.ci.yml` 的 Playwright profile。

### Step 4: Verify Green

- 运行 task-002 的生命周期验证并确认通过。
- 视环境情况运行真实 smoke / E2E 命令，证明 Task 编排与现有执行器保持一致。

## Verification Commands

```bash
bash tests/taskfile/taskfile-lifecycle.test.sh
task verify:smoke
task verify:e2e
task verify:all
```

## Success Criteria

- `verify:smoke`、`verify:e2e`、`verify:all` 已具备完整生命周期语义。
- 默认端口与自动清理语义保持不变。
- `verify:all` 先跑快速验证，再跑完整链路。
- task-002 的失败验证转绿。
