# Architecture: Taskfile 迁移

## Scope

本轮只解决“自动化入口统一”问题，不重做测试体系本身。

保留不变的部分：

- `tests/smoke/smoke_test.sh` 仍作为 smoke 真实执行器
- `docker-compose.ci.yml` 仍作为 smoke / E2E 使用的容器栈
- `web/package.json` 仍作为前端测试的真实脚本入口
- GitHub Actions 仍保留现有 job DAG 与 artifact 上传职责

本轮新增/调整的架构能力：

1. 以 `Taskfile.yml` 取代根 `Makefile`
2. 用命名空间任务统一公开验证入口
3. 将“本地一键验证”与“CI 细粒度辅助任务”分层
4. 把活跃文档和 CI 触发器同步到 Task 入口

## Current Root Causes

### 1. 自动化入口分裂

当前入口同时存在三套表达：

- `Makefile`：只覆盖部分本地验证
- README / `docs/STRUCTURE.md`：仍显示 `make ...`
- `.github/workflows/ci.yml`：多数 job 直接写 shell，只有 `rest-guard` 走 `make`

结果是：

- 本地与 CI 命令不一致
- 文档与实际执行路径容易漂移
- 未来修改自动化时容易漏改某一处入口

### 2. smoke / E2E 的关键语义不只是命令本身

当前 Makefile 的 `smoke-test` / `e2e-test` 不只是命令缩写，还承载了：

- 使用 `docker-compose.ci.yml` 启动容器栈
- 固定 `BASE_URL` / `METRICS_URL` 默认值
- 任务结束后自动 `down --volumes`

若迁移时丢失这些语义，Taskfile 虽然“可用”，但行为会回退。

### 3. Task 的并发语义与当前生命周期需求不完全同构

Task 的 `deps` 适合并行独立任务，但不适合表达：

- `stack:ci:up`
- `smoke:run`
- `e2e:run`
- `stack:ci:down`

这类严格顺序且带清理语义的流程。

因此迁移设计必须显式区分：

- 能并行的校验任务
- 必须串行的生命周期任务

## Target Topology

```text
┌──────────────────────────────────────────────────────────┐
│ Public Task Layer                                        │
│                                                          │
│  guard:rest   test:go   test:web                         │
│        \         |         /                             │
│           └── verify:quick ──┐                           │
│                               ├── verify:all             │
│  verify:smoke ────────────────┘        │                 │
│  verify:e2e  ───────────────────────────┘                │
└──────────────────────────────┬───────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────┐
│ Support Task Layer                                       │
│                                                          │
│  stack:ci:up   stack:ci:logs   stack:ci:down            │
│  smoke:run     e2e:run                                  │
└──────────────────────────────┬───────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────┐
│ Existing Executors                                        │
│                                                          │
│  go test ./...                                           │
│  cd web && bun run test                                  │
│  ./tests/smoke/smoke_test.sh                             │
│  docker compose ... run --rm playwright                  │
└──────────────────────────────────────────────────────────┘
```

## Key Decisions

### 1. 单根 Taskfile，而不是多文件 include

目标文件：

- `Taskfile.yml`

原因：

- 当前任务规模仍小
- 主要目标是替换 Make 入口，不是重构整个自动化结构
- 单文件更利于本次审阅、迁移和 CI 触发管理

未来只有在任务明显扩张到多个领域时，再考虑 include。

### 2. 公开任务全部改为命名空间风格

推荐公开任务：

- `guard:rest`
- `test:go`
- `test:web`
- `verify:quick`
- `verify:smoke`
- `verify:e2e`
- `verify:all`

设计原则：

- `guard:*` 只放保护性检查
- `test:*` 只放纯测试入口
- `verify:*` 承担提交前验证策略与聚合职责

这让 `task --list` 本身就具有可导航性。

### 3. 本地聚合任务与 CI 辅助任务分层

#### 本地聚合任务

- `verify:quick`
  - 适合日常快速回归
- `verify:smoke`
  - 适合 Docker 链路验证
- `verify:e2e`
  - 适合 smoke + Playwright 的完整回归
- `verify:all`
  - 作为最高层总入口

#### CI 辅助任务

- `stack:ci:up`
- `stack:ci:logs`
- `stack:ci:down`
- `smoke:run`
- `e2e:run`

原因：

- CI 需要在 failure / always 分支中保留日志导出与清理能力
- 如果把全部逻辑塞进 `verify:smoke` / `verify:e2e`，workflow 的失败可观测性会变差

### 4. Task 负责编排，不重写脚本状态机

保留现有执行器的理由：

- `tests/smoke/smoke_test.sh` 已包含断言函数、输出统计与错误行为
- Playwright 仍依赖 `docker-compose.ci.yml` 中的 `playwright` service
- `web/package.json` 已是前端脚本真相源

因此 Taskfile 只做：

- 入口命名
- 依赖检查
- 目录切换
- 生命周期编排
- 文档化说明

而不是把复杂 Bash 逻辑重写进 YAML。

### 5. 生命周期语义按“本地”与“CI”分开表达

#### 本地路径

`verify:smoke` / `verify:e2e` 应直接负责编排完整生命周期：

- up
- run
- down

这里可以利用 Task 的 `defer` 能力保留“失败也清理”的语义。

#### CI 路径

workflow 继续显式表达：

- up
- run
- failure logs
- always down

这样做的原因不是重复，而是让 GitHub Actions 的失败分支仍可观测。

### 6. `deps` 只用于独立任务，不用于生命周期顺序

适合 `deps` 的位置：

- `verify:quick` 可并行依赖：`guard:rest`、`test:go`、`test:web`

不适合 `deps` 的位置：

- `verify:smoke`
- `verify:e2e`
- `verify:all`

这些任务必须显式串行，以防 docker compose 生命周期乱序。

## File Touch Points

### Primary Files

- `Taskfile.yml`
  - 新增，作为唯一自动化入口
- `Makefile`
  - 删除
- `.github/workflows/ci.yml`
  - 改为安装并调用 Task
  - 更新触发路径
- `README.md`
  - 更新常用命令与验证分层说明
- `docs/STRUCTURE.md`
  - 更新发布前建议中的验证命令

### Existing Execution Files (no semantic rewrite)

- `tests/smoke/smoke_test.sh`
  - 保持 smoke 断言脚本角色
- `docker-compose.ci.yml`
  - 保持 CI 栈与 Playwright profile
- `web/package.json`
  - 保持前端测试脚本真相源

## Workflow Mapping

### Current -> Target

- `make rest-guard` -> `task guard:rest`
- `go test ./... -short -count=1 -race` -> `task test:go`
- `cd web && bun run test` -> `task test:web`
- 本地 cold-start smoke -> `task verify:smoke`
- 本地 full-chain E2E -> `task verify:e2e`

### GitHub Actions Mapping

- 触发路径：`Makefile` -> `Taskfile.yml`
- `rest-guard` job：shell -> `task guard:rest`
- `unit-test-go` job：shell -> `task test:go`
- `unit-test-frontend` job：shell -> `task test:web`
- `smoke-test` job：`task stack:ci:up` + `task smoke:run` + failure logs + always down
- `e2e-test` job：`task stack:ci:up` + `task e2e:run` + failure logs + always down

## Verification Strategy

### Developer Strategy

- 改 Go / 前端代码但未触及容器链路：`task verify:quick`
- 改服务启动、配置、容器、HTTP / RPC 行为：`task verify:smoke`
- 改页面交互、下载、完整端到端流程：`task verify:e2e`
- 需要完整提交前收口：`task verify:all`

### CI Strategy

- 保留当前 job 粒度，便于单项失败快速定位
- 由 workflow 负责平台初始化与 artifact 上传
- 由 Task 负责仓库内自动化语义统一

## Non-Goals

- 不把 `go run ./cmd/server`、`bun run dev` 强行收编为 Task 必需入口
- 不在本轮引入多环境参数矩阵或复杂变量模板
- 不把 smoke / E2E 做成可跳过式缓存任务
