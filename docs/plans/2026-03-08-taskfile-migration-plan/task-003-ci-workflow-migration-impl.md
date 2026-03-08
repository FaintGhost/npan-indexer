# Task 003: [IMPL] GitHub Actions 迁移到 Task (GREEN)

**depends-on**: task-003-ci-workflow-migration-test.md

## Description

将 `.github/workflows/ci.yml` 迁移为通过 `task` 调度仓库验证任务，保持现有 job 结构、失败日志导出与最终清理路径，并确保 `Taskfile.yml` 变更可以触发 CI。

## Execution Context

**Task Number**: 006 of 008
**Phase**: CI Integration
**Prerequisites**: `task-003-ci-workflow-migration-test.md` 已完成并稳定失败

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

- Modify: `.github/workflows/ci.yml`
- Create or Modify: `tests/taskfile/taskfile-ci-workflow.test.sh` 或等价验证脚本

## Steps

### Step 1: Install and Call Task in Workflow

- 在 workflow 中为 runner 增加 Task 可用性准备步骤。
- 将各 job 的仓库内自动化调用统一切到 Task：
  - `guard:rest`
  - `test:go`
  - `test:web`
  - `stack:ci:up`
  - `smoke:run`
  - `e2e:run`
  - `stack:ci:logs`
  - `stack:ci:down`

### Step 2: Preserve Existing Job Structure

- 保持当前 `unit-test-go`、`unit-test-frontend`、`rest-guard`、`smoke-test`、`e2e-test` 的 job 粒度。
- smoke / e2e job 继续保留 failure logs 与 always cleanup 逻辑，不把这些分支隐藏进单个 task 调用中。

### Step 3: Update Trigger Paths

- 将 workflow 触发路径从 `Makefile` 切换到 `Taskfile.yml`。
- 确保任务入口切换后，修改 Task 配置仍会触发与原先修改 Makefile 等价的 CI。

### Step 4: Verify Green

- 运行 task-003 的 workflow 验证并确认通过。
- 如环境允许，使用本地静态检查或 GitHub Actions lint / dry-run 工具确认 YAML 结构未回退。

## Verification Commands

```bash
bash tests/taskfile/taskfile-ci-workflow.test.sh
```

## Success Criteria

- `.github/workflows/ci.yml` 不再调用 `make`。
- workflow 改为通过 `task` 调度仓库自动化任务。
- `Taskfile.yml` 已纳入 workflow 触发路径。
- smoke / e2e job 仍保留 failure logs 与 cleanup 分支。
- task-003 的失败验证转绿。
