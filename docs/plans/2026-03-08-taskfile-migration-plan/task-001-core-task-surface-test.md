# Task 001: [TEST] 核心 Task 入口与快速验证 (RED)

## Description

为新的 Task 主入口补充失败验证，锁定 namespaced 公开任务与快速验证聚合任务的最小行为：开发者能够看到 `guard:rest`、`test:go`、`test:web`、`verify:quick`，并且 `verify:quick` 作为 guard / Go / 前端测试的统一聚合入口存在。该任务只创建或更新验证资产，不实现生产配置。

## Execution Context

**Task Number**: 001 of 008
**Phase**: Foundation
**Prerequisites**: 仓库仍保留当前 `Makefile` 与现有测试入口；设计文档已确认 namespaced 公开任务范围

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

- Create: `Taskfile.yml`（仅建立最小占位或验证所需骨架时）
- Create or Modify: `tests/taskfile/taskfile-surface.test.sh` 或等价验证脚本
- Modify: `README.md`（仅在测试需要最小断言落点时）

## Steps

### Step 1: Verify Scenario

- 确认本任务只锁定公开任务表面与 `verify:quick` 聚合存在性。
- 不在本任务中覆盖 smoke / E2E 生命周期、CI workflow 或文档全面切换。

### Step 2: Implement Test (Red)

- 创建一个可在本地执行的验证入口，覆盖以下失败断言：
  - `task --list` 尚未展示 namespaced 公开任务
  - `verify:quick` 尚未作为聚合入口存在
  - `test:web` 未通过 Bun script 入口暴露
- 使用 shell 断言、文本匹配或等价轻量方式验证任务表面；不要依赖真实 smoke / E2E 环境。
- 如果需要引用 README 文本，只验证“主入口切换到 task”的最小信号，不在此任务中完成最终文档迁移。

### Step 3: Verify Red Failure

- 运行验证脚本或目标命令，确认新增断言因当前仓库尚未完成 Task 迁移而失败。
- 失败原因必须指向“Task 入口尚不存在或不完整”，而不是因为测试脚本路径、工具缺失或 shell 语法问题导致。

## Verification Commands

```bash
bash tests/taskfile/taskfile-surface.test.sh
```

## Success Criteria

- 新增验证资产稳定失败（Red）。
- 失败信息明确指向 Task 入口表面尚未建立。
- 验证不依赖真实 Docker、Playwright 或外部服务。
