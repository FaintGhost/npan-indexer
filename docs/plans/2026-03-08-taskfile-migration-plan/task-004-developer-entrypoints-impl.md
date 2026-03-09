# Task 004: [IMPL] 开发者入口与文档切换 (GREEN)

**depends-on**: task-004-developer-entrypoints-test.md

## Description

完成开发者活跃入口的最终切换：更新 README 与 `docs/STRUCTURE.md`，清晰区分长驻开发命令与 Task 验证命令，并删除根 `Makefile`，让 `Taskfile.yml` 成为仓库唯一自动化主入口。

## Execution Context

**Task Number**: 008 of 008
**Phase**: Developer Experience
**Prerequisites**: `task-004-developer-entrypoints-test.md` 已完成并稳定失败

## BDD Scenario

```gherkin
Scenario: 开发者通过 README 与 task --list 找到命名空间任务
  Given 仓库根目录已存在 Taskfile.yml 且已删除根 Makefile
  And README 已更新为以 task 作为自动化主入口
  When 开发者执行 "task --list"
  Then 输出中应出现 guard:rest、test:go、test:web、verify:quick、verify:smoke、verify:e2e
  And README 中不再把 make 作为活跃主入口
  And 开发者可以区分快速验证、smoke 与 E2E 三种使用场景

Scenario: 本地开发说明区分长驻命令与验证命令
  Given README 已更新本地开发章节
  When 开发者阅读“本地开发”与“常用命令”部分
  Then 他应看到 go run 与 bun run dev 仍作为长驻开发命令
  And 他应看到 task 负责 guard、测试、smoke 与 E2E 验证入口
  And 文档不会同时给出彼此冲突的主入口说明
```

**Spec Source**: `../2026-03-08-taskfile-migration-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `README.md`
- Modify: `docs/STRUCTURE.md`
- Delete: `Makefile`
- Create or Modify: `tests/taskfile/taskfile-docs-entrypoints.test.sh` 或等价验证脚本

## Steps

### Step 1: Update Active Documentation

- 将 README 中的自动化命令入口切换到 namespaced Task 任务。
- 在 README 中明确区分：
  - 长驻开发命令（`go run ./cmd/server`、`bun run dev`）
  - 快速验证（`verify:quick`）
  - 容器链路验证（`verify:smoke`）
  - 全链路回归（`verify:e2e` / `verify:all`）
- 将 `docs/STRUCTURE.md` 的发布前建议切换到 Task 入口。

### Step 2: Remove Legacy Entry Point

- 删除根 `Makefile`，确保仓库不再保留双自动化入口。
- 检查 README、`docs/STRUCTURE.md` 与 CI workflow 中不存在活跃 `make` 主入口引用。

### Step 3: Verify Green

- 运行 task-004 的文档与入口验证并确认通过。
- 运行 `task --list`，确认开发者从命令行可发现 namespaced 任务。
- 视环境允许，补跑一次 `task verify:quick` 证明文档与真实入口一致。

## Verification Commands

```bash
bash tests/taskfile/taskfile-docs-entrypoints.test.sh
task --list
task verify:quick
```

## Success Criteria

- README 与 `docs/STRUCTURE.md` 已切换为 Task 主入口。
- 根 `Makefile` 已删除。
- 开发者文档能区分长驻开发命令与验证命令。
- 仓库活跃入口不再依赖 `make`。
- task-004 的失败验证转绿。
