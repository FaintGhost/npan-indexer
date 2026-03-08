# Best Practices: Taskfile 迁移

## 1. 把 Task 当作编排层，不是脚本垃圾桶

Taskfile 最适合承担：

- 统一入口
- 任务分层
- 工作目录切换
- 前置条件检查
- 生命周期编排
- 文档化说明

不适合承担：

- 大量 Bash 函数
- 复杂数组与局部状态共享
- 需要长篇 `trap` / `set -euo pipefail` / 条件分支的脚本状态机

因此：

- 保留 `tests/smoke/smoke_test.sh`
- 保留 `docker-compose.ci.yml`
- 保留 `web/package.json` scripts

## 2. 公开任务要少而稳，帮助任务再细分

公开任务应围绕开发者真实心智组织：

- `guard:*`
- `test:*`
- `verify:*`

不要把所有 helper 都暴露在 README 中。

辅助任务（如 `stack:ci:*`、`smoke:run`、`e2e:run`）可以存在，但应作为支持层，而不是主要文档入口。

## 3. 对 `web` 任务优先使用 `dir` + `bun run <script>`

前端脚本真相源已经在 `web/package.json`：

- Task 不应重新发明 `vitest` 命令参数
- `test:web` 应通过 `dir: web` + `bun run test` 执行

这样能避免 README、CI 与 `package.json` scripts 继续漂移。

## 4. `deps` 只做并行前置，不做顺序生命周期

Task 的 `deps` 默认并行。

因此：

- `verify:quick` 可以用 `deps` 聚合 `guard:rest`、`test:go`、`test:web`
- `verify:smoke` / `verify:e2e` / `verify:all` 不应依赖 `deps` 表达顺序

凡是涉及：

- compose up
- smoke / e2e 执行
- compose down

都必须显式串行。

## 5. 本地清理可用 `defer`，CI 清理仍应留在 workflow 分支

Task 官方 `defer` 很适合本地聚合任务的“失败也清理”语义。

但在 GitHub Actions 中，仍建议保持：

- failure 时导日志
- always 时 down

原因：

- CI 需要保留失败后可观测性
- 如果把 down 完全包进单个 Task，workflow 可能拿不到容器日志

## 6. `preconditions` 用于工具检查，`requires` 留给参数化任务

本次迁移最直接的价值是：

- 缺少 `go`、`bun`、`rg`、`docker compose`、`curl`、`jq` 时尽早失败

因此优先使用：

- `preconditions`

而 `requires` 更适合未来存在显式变量输入的任务；本轮不应过度参数化。

## 7. 不要给验证任务加缓存跳过

Task 支持 `status`、`sources`、`generates`、`method` 等跳过机制，但本次迁移中：

- `guard:rest`
- `test:go`
- `test:web`
- `verify:smoke`
- `verify:e2e`

都属于验证任务，应默认每次执行。

原因：

- 测试与 smoke/E2E 的价值在于重新证明系统正确性
- 使用缓存跳过容易导致“命令跑了，但其实没验证”的错觉

## 8. 只改活跃入口，不清理历史档案

本次应更新：

- `README.md`
- `docs/STRUCTURE.md`
- `.github/workflows/ci.yml`

不应为了“全文一致”去追改：

- `docs/plans/**`
- `docs/archive/**`
- 历史 review / todo 记录

否则很容易把一次入口迁移膨胀成大规模文档清扫。

## 9. Playwright 容器继续尊重现有 Node/npm 现实

`docker-compose.ci.yml` 中的 Playwright 服务当前使用官方镜像，并在容器内执行：

- `npm install`
- `npx playwright test`

虽然仓库前端主包管理器是 Bun，但这里不应为了“表面统一”强行重写镜像行为。

规则是：

- 宿主机前端脚本优先 Bun
- Playwright 官方镜像内继续使用现有 Node/npm 路径

## 10. README 要把“何时跑什么”讲清楚

迁移完成后，README 不应只给出命令列表，还应明确：

- `verify:quick`：快速提交前检查
- `verify:smoke`：容器与服务链路验证
- `verify:e2e`：端到端交互验证
- `verify:all`：完整回归

否则用户虽然知道有 Taskfile，仍然不知道该在什么场景下跑什么任务。
