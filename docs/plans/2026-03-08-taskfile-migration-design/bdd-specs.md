# BDD Specifications: Taskfile 迁移

## Feature: Taskfile 成为仓库自动化主入口

### Scenario: 开发者通过 README 与 task --list 找到命名空间任务

```gherkin
Given 仓库根目录已存在 Taskfile.yml 且已删除根 Makefile
And README 已更新为以 task 作为自动化主入口
When 开发者执行 "task --list"
Then 输出中应出现 guard:rest、test:go、test:web、verify:quick、verify:smoke、verify:e2e
And README 中不再把 make 作为活跃主入口
And 开发者可以区分快速验证、smoke 与 E2E 三种使用场景
```

### Scenario: 本地开发说明区分长驻命令与验证命令

```gherkin
Given README 已更新本地开发章节
When 开发者阅读“本地开发”与“常用命令”部分
Then 他应看到 go run 与 bun run dev 仍作为长驻开发命令
And 他应看到 task 负责 guard、测试、smoke 与 E2E 验证入口
And 文档不会同时给出彼此冲突的主入口说明
```

## Feature: REST 守卫通过 Task 保持现有防回退语义

### Scenario: 运行时代码引入 /api/v1 时 guard:rest 失败

```gherkin
Given 某个运行时代码文件新增了字符串 "/api/v1/"
When 开发者执行 "task guard:rest"
Then 任务应以非零状态失败
And 输出应提示运行时代码不得包含 REST /api/v1 路径
```

### Scenario: guard:rest 不误报历史文档与非运行时代码

```gherkin
Given docs/plans、docs/archive、tasks、Markdown、YAML 与 web/dist 中存在 "/api/v1/" 文本
When 开发者执行 "task guard:rest"
Then 这些位置不应导致任务失败
And 守卫的排除规则应与现有 Makefile 行为保持等价
```

## Feature: 快速验证入口聚合独立校验任务

### Scenario: verify:quick 运行 guard、Go 单测与前端单测

```gherkin
Given 开发者已安装 go、bun 与 rg
When 开发者执行 "task verify:quick"
Then 任务应触发 guard:rest、test:go 与 test:web
And test:web 应在 web 目录通过 bun script 执行前端测试
And 任一子任务失败都应使 verify:quick 失败
```

## Feature: smoke 验证保留完整生命周期

### Scenario: verify:smoke 从冷启动运行 smoke 并自动清理

```gherkin
Given 本地尚未启动 docker-compose.ci.yml 对应的服务
When 开发者执行 "task verify:smoke"
Then 任务应先启动 CI compose 栈并等待服务就绪
And 任务应以默认 BASE_URL=http://localhost:11323 与 METRICS_URL=http://localhost:19091 执行 smoke_test.sh
And 无论 smoke 成功还是失败都应执行 compose down --volumes
```

## Feature: E2E 验证保留完整容器链路

### Scenario: verify:e2e 先运行 smoke 再执行 Playwright，并在结束后清理环境

```gherkin
Given 本地尚未启动 docker-compose.ci.yml 对应的服务
When 开发者执行 "task verify:e2e"
Then 任务应启动 CI compose 栈
And 任务应先以默认 BASE_URL=http://localhost:11323 与 METRICS_URL=http://localhost:19091 执行 smoke_test.sh
And 再执行 Playwright 容器验证
And 无论验证成功还是失败都应执行 compose down --volumes
```

### Scenario: verify:all 作为全量回归总入口

```gherkin
Given 开发者需要一次性完成提交前全量回归
When 开发者执行 "task verify:all"
Then 任务应先完成快速验证
And 再执行完整的 E2E 链路
And 最终结果应对任一失败步骤返回非零状态
```

## Feature: GitHub Actions 改为通过 Task 调度

### Scenario: CI workflow 使用 task 并保持现有 job 结构

```gherkin
Given GitHub Actions runner 已安装 Task
When pull request 触发 CI workflow
Then rest-guard、Go 单测、前端单测、smoke、e2e job 都应调用 task 而非 make
And smoke 与 e2e job 仍保留各自的日志导出与最终清理步骤
And 修改 Taskfile.yml 应触发该 workflow
```

## Suggested Automated Verification

## Documentation Verification

- `README.md`
  - 命令示例改为 `task ...`
  - 本地开发与验证入口分层说明清晰
- `docs/STRUCTURE.md`
  - 发布前建议中的验证命令改为 Task 入口

## Workflow Verification

- `.github/workflows/ci.yml`
  - 触发路径从 `Makefile` 改为 `Taskfile.yml`
  - 安装 Task 后调用 `task guard:rest`、`task test:go`、`task test:web`
  - smoke / e2e job 保持日志导出与 cleanup 分支

## Runtime Verification

- 本地执行：
  - `task guard:rest`
  - `task test:go`
  - `task test:web`
  - `task verify:smoke`
  - `task verify:e2e`
- 若迁移涉及 CI：
  - 以 GitHub Actions 或等价本地 dry-run 验证 workflow 命令路径
