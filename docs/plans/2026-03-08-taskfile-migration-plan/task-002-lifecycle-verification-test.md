# Task 002: [TEST] smoke / E2E 生命周期任务 (RED)

## Description

为 `verify:smoke`、`verify:e2e`、`verify:all` 及其底层辅助任务补充失败验证，锁定容器生命周期、默认端口与自动清理语义。该任务只创建或更新验证资产，不实现 Task 生命周期逻辑。

## Execution Context

**Task Number**: 003 of 008
**Phase**: Lifecycle Verification
**Prerequisites**: `task-001-core-task-surface-impl.md` 已完成，根 `Taskfile.yml` 已存在核心任务表面

## BDD Scenario

```gherkin
Scenario: verify:smoke 从冷启动运行 smoke 并自动清理
  Given 本地尚未启动 docker-compose.ci.yml 对应的服务
  When 开发者执行 "task verify:smoke"
  Then 任务应先启动 CI compose 栈并等待服务就绪
  And 任务应以默认 BASE_URL=http://localhost:11323 与 METRICS_URL=http://localhost:19091 执行 smoke_test.sh
  And 无论 smoke 成功还是失败都应执行 compose down --volumes

Scenario: verify:e2e 运行 Playwright 并在结束后清理环境
  Given 本地尚未启动 docker-compose.ci.yml 对应的服务
  When 开发者执行 "task verify:e2e"
  Then 任务应启动 CI compose 栈
  And 任务应执行 Playwright 容器验证
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

- Create or Modify: `tests/taskfile/taskfile-lifecycle.test.sh` 或等价验证脚本
- Modify: `Taskfile.yml`（仅在测试需要声明最小占位任务时）

## Steps

### Step 1: Verify Scenario

- 确认本任务只覆盖生命周期任务语义：
  - `stack:ci:up`
  - `stack:ci:down`
  - `stack:ci:logs`
  - `smoke:run`
  - `e2e:run`
  - `verify:smoke`
  - `verify:e2e`
  - `verify:all`
- 不在本任务中迁移 GitHub Actions 或文档入口。

### Step 2: Implement Test (Red)

- 创建独立验证资产，检查以下行为当前尚未成立：
  - lifecycle 辅助任务不存在或未暴露
  - `verify:smoke` 未声明默认端口语义或自动清理语义
  - `verify:e2e` 未声明 Playwright 容器链路与清理语义
  - `verify:all` 尚未串联快速验证与完整链路
- 采用静态 Taskfile 结构断言、任务摘要断言或等价轻量方式；不要在 Red 阶段要求真实启动 Docker Compose。

### Step 3: Verify Red Failure

- 运行生命周期验证并确认失败。
- 失败原因必须指向生命周期任务缺失或语义不完整，而不是因为本地 Docker 环境不可用。

## Verification Commands

```bash
bash tests/taskfile/taskfile-lifecycle.test.sh
```

## Success Criteria

- 新增生命周期验证资产稳定失败（Red）。
- 失败清楚指出 smoke / E2E / all 任务语义仍未完成。
- 测试不依赖真实容器或外部网络。
