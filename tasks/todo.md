# 任务计划（2026-02-24）

## 目标

- 设计并确认方案 B：修复 `force_rebuild` 下统计异常根因 + 增加单独索引目录能力 + 接入官方统计校验能力。

## 计划清单

- [x] 1. 复核当前全量同步、checkpoint、进度展示链路并形成根因结论
- [x] 2. 调研 FangCloud OpenAPI 可用于目录统计/校验的接口与参数约束
- [x] 3. 设计后端改造（checkpoint 清理、请求参数扩展、统计校验接口/服务）
- [x] 4. 设计前端改造（单目录输入、开关参数、结果展示与误用防护）
- [x] 5. 编写 BDD 场景与测试策略（后端单测 + 前端单测 + 集成验证）
- [x] 6. 产出设计文档包到 `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/`
- [x] 7. 回填评审结论与实现前置条件

## 评审记录（完成后填写）

- 结论：方案 B 可行，且能直接覆盖“全量+强制重建统计偏小”的核心根因。
- 根因优先级：`force_rebuild` 未清理 checkpoint > 目录范围误配置 > 上游分页异常。
- 实施建议：先做后端 checkpoint 修复与测试，再做前端目录范围输入，最后做统计告警展示。
- 设计产物：
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/_index.md`
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/architecture.md`
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/bdd-specs.md`
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/best-practices.md`

## 实施计划（OpenAPI 约束）

- [x] 1. 完成 `docs/plans/2026-02-24-full-rebuild-and-folder-scope-plan/` 任务分解
- [x] 2. 按计划执行后端 checkpoint 修复与测试
- [x] 3. 按计划执行后端 folder info 估计与告警
- [x] 4. 按计划执行前端目录范围输入与请求体对齐
- [x] 5. 按计划执行前端估计/告警展示
- [x] 6. 运行生成校验与回归测试并记录结果

## 新任务：Docker Registry 发布流水线

- [x] 1. 产出 `docker publish` 实施计划文档（任务拆分 + 验证步骤）
- [x] 2. 新增 GitHub Action：构建镜像并推送 Docker Hub + GHCR
- [x] 3. 补充仓库配置说明（Secrets、触发条件、镜像标签策略）
- [x] 4. 自检工作流语法与影响范围

## 新任务：ARM64 构建切换到 Self-hosted Runner

- [x] 1. 调整 `docker-publish` workflow 为分平台构建（amd64 / arm64）
- [x] 2. 将 `linux/arm64` 任务绑定到 `self-hosted Linux ARM64 debian13 trixie`
- [x] 3. 新增 manifest 合并步骤，确保两平台标签统一推送到 Docker Hub + GHCR
- [x] 4. 更新 README 对 runner 前置条件说明并完成工作流静态自检

## 新任务：修复 Docker Publish Action 的 secret output 警告与构建失败

- [x] 1. 分析 GitHub Actions 注解与失败日志，确认 `prepare` job output 被 secret 保护跳过的根因
- [x] 2. 改造 workflow，移除跨 job 传递镜像名（含 secret 风险），改为各 job 内本地计算
- [x] 3. 静态自检 workflow 改动并回填验证记录

## Review（Docker 发布流水线）

- 已新增 workflow：`.github/workflows/docker-publish.yml`
  - 触发：`push(main)`、`push(v*)`、`workflow_dispatch`
  - 行为：buildx 多平台构建并推送 Docker Hub + GHCR
  - 标签：`latest`、`ref`、`sha-*`
- 已更新文档：`README.md` 增加“镜像发布（GitHub Actions）”章节。
- 自检结果：
  - `git diff --check -- .github/workflows/docker-publish.yml README.md ...` 通过
  - 本地无法直接模拟 GitHub 托管运行环境；首次真实验证需在 GitHub 上触发 workflow。

## Review（ARM64 Self-hosted Runner 切换）

- Workflow 已拆分为 `prepare`、`build(matrix)`、`merge` 三段：
  - `amd64` 继续使用 GitHub 托管 `ubuntu-latest`
  - `arm64` 改为 `runs-on: [self-hosted, Linux, ARM64, debian13, trixie]`
- `build` 阶段按平台 push by digest，`merge` 阶段统一创建多架构 manifest 并打标签（Docker Hub + GHCR）。
- 文档已补充 ARM64 self-hosted runner 标签要求，避免 workflow 因 runner 不匹配而 pending。
- 验证结果：
  - `git diff --check -- .github/workflows/docker-publish.yml README.md tasks/todo.md` 通过
  - `GOCACHE=/tmp/go-build go test ./...` 通过（本机默认 `/root/.cache/go-build` 无写权限）
  - `cd web && bun vitest run` 通过（22 files / 184 tests）
  - 容器化链路（`docker compose ... up --wait` + `./tests/smoke/smoke_test.sh` + `docker compose --profile e2e run --rm playwright`）通过
    - 冒烟：34/34
    - E2E：32/32

## Review（Docker Publish secret output 修复）

- 根因：
  - `prepare` job 输出 `dockerhub_image` / `ghcr_image` 时，GitHub 判定输出值“可能包含 secret”（`DOCKERHUB_USERNAME` 参与拼接），因此跳过 job output。
  - 下游 `build` job 中 `docker/build-push-action` 的 `outputs` 里 `name=` 变为空字符串，触发 `ERROR: tag is needed when pushing to registry`。
- 修复：
  - 删除 `prepare` job 的跨 job 镜像名输出依赖。
  - 在 `build` / `merge` job 内分别通过 `Prepare image names` 步骤计算并使用 step outputs。
- 预期结果：
  - 消除 `prepare` 的 2 个 warning（skip output）
  - 消除 `Build linux/amd64` 与 `Build linux/arm64` 的 2 个 buildx error

## Review（本轮实施结果）

- 后端：
  - 已修复 `force_rebuild` 与 `resume=false` 场景下的 checkpoint 清理，避免残留断点污染全量统计。
  - 已接入 `GetFolderInfo`（`/api/v2/folder/{id}/info`）用于显式根目录估计值与名称回填。
  - 已在 verification 阶段追加 root 级“估计 vs 实际”差异告警。
- 前端：
  - Admin 页面新增目录 ID 输入（逗号分隔），支持单目录/多目录范围索引。
  - `startSync` 在传入 root IDs 时会自动发送 `include_departments=false`（契约字段）。
  - 根目录详情新增“估计/实际”展示，便于直接定位差异。
- OpenAPI 约束：
  - `api/openapi.yaml` 未发生契约漂移。
  - 生成校验已通过（`go generate ./api/...` + `bun run generate` + `git diff --exit-code -- api/types.gen.go web/src/api/generated/`）。
- 测试验证：
  - `go test ./...` 通过。
  - `cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/admin-page.test.tsx src/components/sync-progress-display.test.tsx` 通过。
  - `cd web && bun vitest run` 全量前端单测通过（22 files / 184 tests）。
  - `./tests/smoke/smoke_test.sh` 通过（34/34）。
  - `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright` 通过（32/32）。
