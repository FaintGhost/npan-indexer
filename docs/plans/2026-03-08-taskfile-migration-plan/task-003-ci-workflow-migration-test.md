# Task 003: [TEST] GitHub Actions 迁移到 Task (RED)

## Description

为 `.github/workflows/ci.yml` 补充失败验证，锁定 CI workflow 对 `task` 的调用、`Taskfile.yml` 触发路径，以及 smoke / E2E job 继续保留日志导出与最终清理语义。该任务只创建或更新验证资产，不修改 workflow 生产配置。

## Execution Context

**Task Number**: 005 of 008
**Phase**: CI Integration
**Prerequisites**: `task-002-lifecycle-verification-impl.md` 已完成，Taskfile 已具备核心与生命周期任务

## BDD Scenario

```gherkin
Scenario: CI workflow 使用 task 并保持现有 job 结构
  Given GitHub Actions runner 已安装 Task
  When pull request 触发 CI workflow
  Then rest-guard、Go 单测、前端单测、smoke、e2e job 都应调用 task 而非 make
  And smoke 与 e2e job 仍保留各自的日志导出与最终清理步骤
  And 修改 Taskfile.yml 应触发该 workflow
```

**Spec Source**: `../2026-03-08-taskfile-migration-design/bdd-specs.md`

## Files to Modify/Create

- Create or Modify: `tests/taskfile/taskfile-ci-workflow.test.sh` 或等价验证脚本
- Modify: `.github/workflows/ci.yml`（仅在测试需要最小占位时）

## Steps

### Step 1: Verify Scenario

- 确认本任务只覆盖 workflow 入口切换与触发路径，不覆盖 README / 结构文档。
- workflow 级验证重点是命令路径与分支语义，而不是实际在本地执行 GitHub Actions。

### Step 2: Implement Test (Red)

- 创建验证脚本或等价断言，检查以下行为当前尚未成立：
  - workflow 仍在调用 `make` 或直接 shell，而不是统一调用 `task`
  - `Taskfile.yml` 尚未纳入触发路径
  - smoke / e2e job 的 failure logs 与 always cleanup 语义未与 Task 迁移对齐
- 通过 YAML 文本断言、结构匹配或等价方式验证，不依赖真实 GitHub Actions 环境。

### Step 3: Verify Red Failure

- 运行 workflow 验证并确认失败。
- 失败原因必须指向 workflow 尚未完成 Task 迁移，而不是验证脚本自身失效。

## Verification Commands

```bash
bash tests/taskfile/taskfile-ci-workflow.test.sh
```

## Success Criteria

- 新增 workflow 验证资产稳定失败（Red）。
- 失败明确指向 CI 仍未切到 Task。
- 测试不依赖 GitHub 远端执行。
