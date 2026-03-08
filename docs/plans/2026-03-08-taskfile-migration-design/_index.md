# Taskfile 迁移设计

## Context

当前仓库的自动化入口存在三类割裂：

- 根 `Makefile` 只覆盖 5 个任务：`rest-guard`、`test`、`test-frontend`、`smoke-test`、`e2e-test`
- `README.md` 与 `docs/STRUCTURE.md` 仍把 `make` 作为活跃入口之一
- `.github/workflows/ci.yml` 只有 `rest-guard` job 走 `make`，其余 job 直接执行 shell 命令，导致本地与 CI 缺少统一编排层

同时，当前 `smoke-test` / `e2e-test` 的关键价值不只是“能跑命令”，而是：

- 使用 `docker-compose.ci.yml` 启动完整 CI 栈
- 依赖固定默认端口：`11323` / `19091`
- 通过退出清理语义保证容器与 volume 不残留

用户已明确确认两项方向：

- **直接移除 Makefile**，不做双轨兼容
- **公开任务采用命名空间风格**，而不是继续保留旧 target 名称

## Problem Statement

如果只把 `Makefile` 机械翻译成 `Taskfile.yml`，仍会遗留三个问题：

1. 开发者入口不统一：README、CI、文档与本地习惯仍会继续漂移
2. 生命周期语义容易回退：smoke / E2E 的自动清理与失败排查路径可能丢失
3. 任务边界不清晰：哪些适合本地一键跑，哪些适合 CI 分 job 调度，当前没有清晰分层

## Goals

- 用根级 `Taskfile.yml` 取代根 `Makefile`，作为仓库自动化主入口
- 覆盖当前 guard / Go 单测 / 前端单测 / smoke / E2E 能力
- 用命名空间任务统一开发者验证入口
- 保留当前 smoke / E2E 的默认端口、Compose 拓扑与清理语义
- 让本地与 CI 共享同一套任务语义，但允许 CI 保留更细粒度的 job 编排
- 更新活跃文档与 CI 触发器，避免入口继续漂移

## Non-Goals

- 本次不重写 `tests/smoke/smoke_test.sh` 的断言逻辑
- 本次不调整 `docker-compose.ci.yml` 的服务拓扑或 Playwright 容器行为
- 本次不清理历史归档设计文档中所有 `make` 文本引用
- 本次不把所有长驻开发命令都包装成 Task（如 `go run ./cmd/server`、`bun run dev`）
- 本次不扩展到发布、部署、镜像发布等与 Makefile 无关的自动化领域

## Requirements

### Must

- 新增根 `Taskfile.yml`，并删除根 `Makefile`
- 公开任务使用命名空间风格，至少覆盖：
  - `guard:rest`
  - `test:go`
  - `test:web`
  - `verify:quick`
  - `verify:smoke`
  - `verify:e2e`
- `test:web` 必须继续以 `web/package.json` 中的 Bun script 为真实入口
- `verify:smoke` 必须保留现有 cold-start 本地体验：启动 CI compose 栈、运行 smoke、自动清理
- `verify:e2e` 必须保留现有 full-chain 本地体验：启动 CI compose 栈、先运行 smoke、再运行 Playwright，并自动清理
- `.github/workflows/ci.yml` 必须改为调用 `task`，且修改 `Taskfile.yml` 时能触发 CI
- `README.md` 与 `docs/STRUCTURE.md` 必须改为以 `task` 为主的验证入口

### Should

- 采用单一根 `Taskfile.yml`，而不是在当前规模下拆多个 include 文件
- 用 Task 负责“编排”，继续保留 shell 脚本与 Docker Compose 作为真实执行器
- 区分“本地一键验证任务”和“CI 细粒度辅助任务”
- 对工具依赖使用 `preconditions` 做前置检查（如 `go`、`bun`、`docker compose`、`rg`、`curl`、`jq`）
- 对 `web` 任务使用 `dir` 而不是 `cd web && ...`
- 对外公开任务全部补齐 `desc`，复杂任务补 `summary`

### Won't

- 不保留旧 Make target 作为 alias
- 不引入多层 Taskfile 命名空间拆分
- 不把 GitHub Actions 的 artifact 上传等平台能力搬进 Task
- 不对 smoke / E2E 使用 `sources` / `generates` 做缓存跳过

## Option Analysis

### Option A（推荐）: 单根 Taskfile + 命名空间公开任务 + CI 辅助任务分层

在根目录新增单一 `Taskfile.yml`，将面向开发者的常用入口收敛为命名空间任务；同时保留一组供 CI 复用的细粒度辅助任务。

优点：

- 最符合“直接移除 Makefile”的目标
- 公开入口稳定、可发现，`task --list` 可直接作为命令目录
- 本地与 CI 共用同一语义层，但不强迫 CI 退化成单 job 大命令
- 能清晰保留 smoke / E2E 生命周期与失败排查策略

代价：

- 需要同时改 `Taskfile.yml`、README、`docs/STRUCTURE.md` 与 CI workflow
- 需要重新定义公开任务命名与分层

### Option B: 先上 Taskfile，保留 Makefile 转发层

不采用。

原因：与用户已确认的“直接移除 Makefile”冲突，而且会继续保留双入口漂移问题。

### Option C: 一开始就拆成多个 Taskfile include

不推荐。

当前任务数量有限，过早拆分会增加认知成本，也会把本次迁移从“入口统一”变成“结构重构”。

## Rationale

选择 Option A 的原因：

- 当前问题的核心不是 Make 语法本身，而是**入口分散 + 生命周期重复 + 文档漂移**
- 单根 Taskfile 足以承载现有规模，不需要引入更多结构复杂度
- 命名空间任务可以清晰表达职责边界，例如 `guard:*`、`test:*`、`verify:*`
- Task 官方能力（`dir`、`preconditions`、`defer`、`summary`）足以覆盖当前 Makefile 的主要语义，同时比 Make 更适合作为团队文档化入口

## Detailed Design

### 1. 公开任务分层

推荐对开发者公开以下任务：

- `guard:rest`
  - 保留现有 `/api/v1/` 运行时代码防回退检查
- `test:go`
  - 保留当前 Go 单测参数：`go test ./... -short -count=1 -race`
- `test:web`
  - 在 `web/` 目录执行 `bun run test`
- `verify:quick`
  - 聚合 `guard:rest`、`test:go`、`test:web`
  - 作为最常用的快速本地回归入口
- `verify:smoke`
  - 面向本地冷启动验证：启动 CI compose 栈，执行 smoke，自动清理
- `verify:e2e`
  - 面向本地全链路验证：启动 CI compose 栈，执行 Playwright，自动清理
- `verify:all`
  - 串行执行 `verify:quick` 与 `verify:e2e`
  - 作为全量回归总入口

### 2. 辅助任务分层

为避免把 CI 逻辑硬塞进单个大任务，建议补充一组细粒度辅助任务：

- `stack:ci:up`
- `stack:ci:logs`
- `stack:ci:down`
- `smoke:run`
- `e2e:run`

它们的定位是：

- 让本地 `verify:*` 组合任务可以复用底层执行器
- 让 GitHub Actions 继续保留“失败时导日志、最后清理”的 job 级控制权
- 避免把 `smoke` / `e2e` 的复杂 shell 状态机重新塞回 Task YAML

### 3. 本地与 CI 使用策略

#### 本地开发者入口

- 日常提交前：`task verify:quick`
- 需要容器链路验证：`task verify:smoke`
- 需要全链路回归：`task verify:e2e` 或 `task verify:all`

本地长驻开发命令保持不变：

- 后端：`go run ./cmd/server`
- 前端：`cd web && bun run dev`

也就是说，本次迁移统一的是**自动化验证入口**，不是把所有开发命令都任务化。

#### CI 入口

CI 继续保留现有 job DAG，但 job 内统一改为调用 `task`：

- `rest-guard` job -> `task guard:rest`
- `unit-test-go` job -> `task test:go`
- `unit-test-frontend` job -> `task test:web`
- `smoke-test` job -> `task stack:ci:up` -> `task smoke:run` -> failure 时 `task stack:ci:logs` -> always `task stack:ci:down`
- `e2e-test` job -> `task stack:ci:up` -> `task e2e:run` -> failure 时 `task stack:ci:logs` -> always `task stack:ci:down`

这样可以同时满足：

- 本地有一键验证入口
- CI 仍保留可观测性与失败日志导出

### 4. 生命周期与清理语义

当前 Makefile 的关键语义是“失败也清理”。迁移后需要保留两种路径：

- 本地 `verify:smoke` / `verify:e2e`
  - 使用 Task 的 `defer` 或等价编排，在任务结束时执行 `stack:ci:down`
- GitHub Actions
  - 保持 workflow 中的 failure/always 分支，先导日志，再执行 `stack:ci:down`

关键设计点：

- 不把 CI failure logs 完全埋进一个本地聚合任务里
- 不让 `deps` 承担顺序语义；`stack:ci:up -> smoke:run -> e2e:run -> stack:ci:down` 必须显式串行表达

### 5. 文档与触发器更新范围

必须更新的活跃入口：

- `README.md`
- `docs/STRUCTURE.md`
- `.github/workflows/ci.yml`
- 删除根 `Makefile`

明确不追溯修改：

- `docs/plans/**`
- `docs/archive/**`
- `tasks/**` 中的历史记录

### 6. 执行器保持不变

Taskfile 负责编排，真实执行器保持现状：

- Go 测试：`go test`
- 前端测试：`bun run test`
- Smoke：`./tests/smoke/smoke_test.sh`
- E2E：`docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright`

这样可以保证本次迁移聚焦在“入口统一”，而不是引入新的测试行为差异。

## Success Criteria

- 根目录存在 `Taskfile.yml`，根 `Makefile` 被删除
- `task --list` 可以展示命名空间公开任务
- 开发者能通过 `task verify:quick`、`task verify:smoke`、`task verify:e2e` 完成分层验证
- `README.md` 与 `docs/STRUCTURE.md` 中不再把 `make` 作为活跃主入口
- `.github/workflows/ci.yml` 中不再调用 `make`
- 修改 `Taskfile.yml` 时会触发 CI
- `guard:rest` 继续保持当前排除规则，不误报历史文档与生成产物
- `verify:smoke` / `verify:e2e` 继续保持自动清理语义

## Risks and Mitigations

- 风险：把顺序流程错误地写成 `deps`，导致 compose 生命周期乱序
  - 缓解：仅将 `deps` 用于独立前置任务；生命周期步骤显式串行
- 风险：本地聚合任务清理过早，导致 CI 无法导出失败日志
  - 缓解：本地聚合任务与 CI 辅助任务分层，CI 保留 job 级 failure/always 控制
- 风险：Taskfile 过度承载 shell 逻辑，后续难维护
  - 缓解：继续保留 `smoke_test.sh` 与 Compose 配置作为真实执行器
- 风险：README 只改命令、不解释验证层级，开发者仍不清楚何时跑什么
  - 缓解：在 README 中增加“快速验证 / smoke / E2E / 全量回归”的分层说明

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为场景与测试策略
- [Architecture](./architecture.md) - 任务拓扑、CI 集成与文件触点
- [Best Practices](./best-practices.md) - Taskfile 迁移约束与实施注意事项
