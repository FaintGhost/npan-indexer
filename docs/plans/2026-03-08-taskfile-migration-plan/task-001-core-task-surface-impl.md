# Task 001: [IMPL] 核心 Task 入口与快速验证 (GREEN)

**depends-on**: task-001-core-task-surface-test.md

## Description

建立根级 `Taskfile.yml` 的核心公开任务表面，落地 `guard:rest`、`test:go`、`test:web`、`verify:quick` 等 namespaced 入口，并通过前置检查、工作目录切换与聚合关系，使 Task 成为新的最小可用自动化主入口。

## Execution Context

**Task Number**: 002 of 008
**Phase**: Foundation
**Prerequisites**: `task-001-core-task-surface-test.md` 已完成并稳定失败

## BDD Scenario

```gherkin
Scenario: 开发者通过 README 与 task --list 找到命名空间任务
  Given 仓库根目录已存在 Taskfile.yml 且已删除根 Makefile
  And README 已更新为以 task 作为自动化主入口
  When 开发者执行 "task --list"
  Then 输出中应出现 guard:rest、test:go、test:web、verify:quick、verify:smoke、verify:e2e
  And README 中不再把 make 作为活跃主入口
  And 开发者可以区分快速验证、smoke 与 E2E 三种使用场景

Scenario: verify:quick 运行 guard、Go 单测与前端单测
  Given 开发者已安装 go、bun 与 rg
  When 开发者执行 "task verify:quick"
  Then 任务应触发 guard:rest、test:go 与 test:web
  And test:web 应在 web 目录通过 bun script 执行前端测试
  And 任一子任务失败都应使 verify:quick 失败
```

**Spec Source**: `../2026-03-08-taskfile-migration-design/bdd-specs.md`

## Files to Modify/Create

- Create: `Taskfile.yml`
- Create or Modify: `tests/taskfile/taskfile-surface.test.sh` 或等价验证脚本
- Modify: `web/package.json`（仅当需要补齐稳定测试 script 名称时）

## Steps

### Step 1: Establish Public Task Surface

- 在根级 `Taskfile.yml` 中建立 namespaced 公开任务：
  - `guard:rest`
  - `test:go`
  - `test:web`
  - `verify:quick`
  - 以及为后续任务预留的 `verify:smoke`、`verify:e2e`、`verify:all`
- 为公开任务补齐 `desc`；对聚合任务补齐 `summary`，让 `task --list` 能直接反映用途。

### Step 2: Preserve Existing Executors

- 让 `guard:rest` 继续复用当前 Makefile 的排除规则和运行时 guard 语义。
- 让 `test:go` 保持当前 Go 测试参数。
- 让 `test:web` 通过 `dir` 切换到 `web/` 并调用 Bun script，而不是重新发明前端测试命令。

### Step 3: Add Preconditions and Aggregate Quick Verification

- 为 `guard:rest`、`test:go`、`test:web` 添加最小必要的工具前置检查。
- 让 `verify:quick` 作为快速回归聚合入口，能够统一触发 guard / Go / 前端测试，并在任一失败时返回非零状态。

### Step 4: Verify Green

- 运行 task-001 的失败验证并确认通过。
- 分别运行 `task guard:rest`、`task test:go`、`task test:web`、`task verify:quick`，确认新入口与既有执行器语义一致。

## Verification Commands

```bash
bash tests/taskfile/taskfile-surface.test.sh
task guard:rest
task test:go
task test:web
task verify:quick
```

## Success Criteria

- 根 `Taskfile.yml` 已建立核心公开任务表面。
- `task --list` 能展示 namespaced 任务。
- `test:web` 继续通过 Bun script 作为真实入口。
- `verify:quick` 成为新的快速回归总入口。
- task-001 的失败验证转绿。
