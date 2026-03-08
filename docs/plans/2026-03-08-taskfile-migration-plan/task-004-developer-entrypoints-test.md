# Task 004: [TEST] 开发者入口与文档切换 (RED)

## Description

为 README、`docs/STRUCTURE.md` 与根入口切换补充失败验证，锁定“task 成为活跃主入口、Makefile 被移除、长驻开发命令与验证命令分层说明清晰”这三类最终交付行为。该任务只创建或更新验证资产，不执行最终文档迁移。

## Execution Context

**Task Number**: 007 of 008
**Phase**: Developer Experience
**Prerequisites**: `task-003-ci-workflow-migration-impl.md` 已完成，Taskfile 与 CI 都已可用

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

- Create or Modify: `tests/taskfile/taskfile-docs-entrypoints.test.sh` 或等价验证脚本
- Modify: `README.md`（仅在测试需要最小占位时）
- Modify: `docs/STRUCTURE.md`（仅在测试需要最小占位时）

## Steps

### Step 1: Verify Scenario

- 确认本任务只覆盖开发者活跃入口切换和根 Makefile 移除。
- 不在本任务中重新验证 CI workflow 或生命周期任务细节。

### Step 2: Implement Test (Red)

- 创建验证资产，检查以下行为当前尚未全部成立：
  - README 仍存在 `make rest-guard`、`make smoke-test`、`make e2e-test`
  - `docs/STRUCTURE.md` 仍将 `make` 作为发布前建议主入口
  - 根 `Makefile` 仍存在
  - 文档尚未清楚区分长驻开发命令与验证命令
- 使用轻量 shell / 文本断言，不依赖真实 task 执行结果。

### Step 3: Verify Red Failure

- 运行文档与入口验证并确认失败。
- 失败原因必须指向活跃文档和根入口尚未切换，而不是脚本路径或匹配规则错误。

## Verification Commands

```bash
bash tests/taskfile/taskfile-docs-entrypoints.test.sh
```

## Success Criteria

- 新增文档与入口验证资产稳定失败（Red）。
- 失败清楚指向 README / `docs/STRUCTURE.md` / `Makefile` 仍未完成切换。
- 测试不依赖外部服务。
