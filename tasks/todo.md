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

## 新任务：修复 merge 阶段 manifest source 拼接错误

- [x] 1. 基于报错 `invalid reference format` 回溯 `sources` 拼接逻辑并复现变量展开问题
- [x] 2. 将 `printf` 拼接改为显式循环构造每个平台 digest 的完整引用
- [x] 3. 完成静态自检并记录根因与修复

## 新任务：优化 Docker Build 缓存与修复 merge digest not found

- [x] 1. 优化 Docker 构建缓存（BuildKit cache mount + gha/registry 双缓存）
- [x] 2. 收敛构建上下文（排除与镜像产物无关的高频变更目录）
- [x] 3. 修复 `merge` 阶段按 Docker Hub digest 查找失败（统一以 GHCR digest 为 source）
- [x] 4. 完成静态检查与完整回归测试

## 新任务：CI 测试仅在源码变更时触发

- [x] 1. 收敛 `.github/workflows/ci.yml` 触发条件为源码白名单路径
- [x] 2. 覆盖后端/前端/测试与构建关键文件，排除文档与任务记录类变更
- [x] 3. 完成 workflow 静态检查并记录变更影响

## 新任务：修复 CI 中 go:embed 在 dist 仅含隐藏文件时失败

- [x] 1. 基于 GitHub Actions 日志定位 `web/embed.go` 匹配模式失败根因
- [x] 2. 调整 embed 指令以支持 `dist` 目录仅包含 `.gitkeep` 的场景
- [x] 3. 运行 Go 测试验证 `cmd/server` 与 `npan/web` 编译路径恢复正常

## 新任务：修复 CI checkout 缺少 web/dist 导致 go:embed 失败

- [x] 1. 从最新 run 日志确认 `pattern all:dist: no matching files found` 的真实原因是目录缺失
- [x] 2. 在 `unit-test-go` 中补充 `web/dist` 占位创建步骤，确保 Go 测试前满足 embed 前提
- [x] 3. 触发并观察新一轮 CI 验证结果

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

## Review（merge source 拼接错误修复）

- 根因：
  - `sources="$(printf '%s@sha256:%s ' "${TARGET_IMAGE}" *)"` 在 `*` 匹配多个 digest 文件时，`printf` 会重复使用格式串。
  - 第二轮输出会把某个 digest 当成“镜像名”，并因缺少配对参数生成 `digest@sha256:`，触发 `invalid reference format`。
- 修复：
  - 改为 `for digest in *; do ...; done` 显式构造：
    `${TARGET_IMAGE}@sha256:<digest>`。
  - 同步修复 Docker Hub 与 GHCR 两个 manifest 合并步骤。
- 预期结果：
  - `merge` 阶段不再出现 `failed to parse source "...@sha256:"`。

## Review（Build 缓存优化 + merge digest not found 修复）

- 缓存优化：
  - `Dockerfile` 启用 BuildKit cache mount：
    - bun 依赖缓存：`/root/.bun/install/cache`
    - go mod 缓存：`/go/pkg/mod`
    - go build 缓存：`/root/.cache/go-build`
  - workflow `build` job 增加双缓存后端：
    - `type=gha`（快速近端缓存）
    - `type=registry`（GHCR 持久缓存，跨 runner 复用）
  - `.dockerignore` 新增排除：`tasks/`、`.claude/`、`web/e2e/`、`web/playwright-report/`、`web/test-results/`、`web/tsconfig.tsbuildinfo`。
- `merge` 报错根因：
  - 之前 `build` 同时向 Docker Hub + GHCR push by digest，但 artifact 仅保存了单一 digest。
  - `merge` 在 Docker Hub 用该 digest 查源时可能不存在，触发 `not found`。
- 修复方案：
  - `build` 改为仅向 GHCR push by digest（digest 作为统一 source）。
  - `merge` 在 Docker Hub / GHCR 两个 manifest 创建步骤都使用 GHCR digest source，再分别打目标仓库标签。
- 预期结果：
  - 同源码重复构建命中率显著提升。
  - `merge` 阶段不再出现 `docker.io/...@sha256:... not found`。

## Review（CI 源码路径触发约束）

- 背景：
  - 你要求“测试只在源码发生变更时进行”。
- 调整：
  - 将 `.github/workflows/ci.yml` 的 `push` / `pull_request` 触发从 `paths-ignore` 改为 `paths` 白名单。
  - 白名单包含：`api/**`、`cmd/**`、`internal/**`、`tests/**`、`web/**`、`go.mod`、`go.sum`、`Dockerfile`、`docker-compose*.yml`、`Makefile`。
- 影响：
  - 仅修改 `docs/**`、`tasks/**`、`.claude/**`、普通 markdown 等非源码文件时，不再触发整套 CI 测试。

## Review（go:embed dist 匹配失败修复）

- 根因：
  - CI checkout 后 `web/dist` 只有 `.gitkeep`，原模式 `//go:embed dist/*` 不匹配隐藏文件。
  - 导致 `go test ./...` 报错：`pattern dist/*: no matching files found`。
- 修复：
  - `web/embed.go` 改为 `//go:embed all:dist`，允许包含隐藏文件并递归嵌入目录。
- 验证：
  - `GOCACHE=/tmp/go-build go test ./... -short` 通过（`cmd/server` 与 `npan/web` 已不再 setup failed）。

## Review（CI checkout 缺少 web/dist 修复）

- 根因：
  - GitHub checkout 环境中 `web/dist` 不存在（`.gitkeep` 未被跟踪），导致 `//go:embed all:dist` 仍然匹配失败。
- 修复：
  - 在 `.github/workflows/ci.yml` 的 `unit-test-go` job 增加步骤：
    - `mkdir -p web/dist`
    - `touch web/dist/.gitkeep`
  - 保证 `go test ./...` 执行前 embed 路径必定存在。
- 结果预期：
  - `unit-test-go` 不再因 `web/embed.go` setup failed 而中断。

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

## 新任务：Admin 局部补同步交互重构（Brainstorming 设计）

- [x] 1. 复盘当前 Admin 页面“目录 ID 输入即同步”的交互与根因链路
- [x] 2. 确认“局部同步后根目录详情被覆盖”的根因位于服务端 progress 语义
- [x] 3. 输出可选方案（前端本地缓存 vs 后端目录册保留）并给出推荐
- [x] 4. 产出设计文档包到 `docs/plans/2026-02-24-admin-partial-resync-toggle-design/`
- [x] 5. 补充 BDD 场景（toggle 局部补同步 / 拉取目录详情 / 运行中禁用 / 互斥规则）
- [x] 6. 回填评审记录与未决问题

## Review（Admin 局部补同步交互重构 / 设计）

- 结论：
  - 采用“新增目录详情拉取接口 + 后端保留根目录目录册（catalog）”方案。
  - 目标是把“目录发现”和“同步执行”解耦，并修复 scoped full 后 UI 列表被覆盖的问题。
- 根因：
  - 当前 `root_folder_ids` 输入直接驱动 `/api/v1/admin/sync`。
  - full path 会按本次 roots 重建 progress，`rootProgress` 被覆盖。
  - 前端根目录详情直接渲染 `progress.rootProgress`，因此只剩本次目录。
- 设计产物：
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-design/_index.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-design/architecture.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-design/best-practices.md`
- 未决问题（实现前确认）：
  - 拉取目录详情成功后，新加入目录是否默认自动勾选（设计稿当前默认“自动勾选”）。

## 新任务：Admin 局部补同步交互重构（Writing Plans）

- [x] 1. 基于设计文档产出实施计划目录 `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/`
- [x] 2. 将任务拆分为 BDD 驱动的 Red/Green 粒度（每任务单文件）
- [x] 3. 补全任务依赖、验证命令与回归收口任务
- [x] 4. 回填执行移交说明（下一步进入 `executing-plans`）

## Review（Admin 局部补同步交互重构 / 实施计划）

- 计划目录：
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/_index.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-001-red-backend-inspect-roots-tests.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-002-green-backend-inspect-roots-api-and-contract.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-003-red-backend-catalog-preserve-tests.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-004-green-backend-catalog-preserve-impl.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-005-red-frontend-inspect-and-autoselect-tests.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-006-green-frontend-inspect-and-autoselect-impl.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-007-red-frontend-running-lock-and-guard-tests.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-008-green-frontend-running-lock-and-guard-impl.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-009-red-frontend-catalog-fallback-tests.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-010-green-frontend-catalog-fallback-impl.md`
  - `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/task-011-verification-and-regression.md`
- 关键决策已固化：
  - 拉取目录详情成功后，新目录默认自动勾选（按用户确认）。

## 新任务：Admin 局部补同步交互重构（Executing Plans）

- [x] 1. 实现后端 `inspect roots` 接口（批量、部分成功）并接入 Admin 路由鉴权
- [x] 2. 扩展同步启动请求字段 `preserve_root_catalog` 并加 `force_rebuild + scoped` 防线
- [x] 3. 扩展同步进度模型与 DTO：支持 `catalogRoots/catalogRootNames/catalogRootProgress`
- [x] 4. 在 full scoped 运行中保留历史根目录目录册，修复“列表被覆盖”
- [x] 5. 前端实现“拉取目录详情”独立按钮与 `inspectRoots` 流程
- [x] 6. 前端在根目录详情中增加 toggle，并按勾选目录发起全量补同步
- [x] 7. 新拉取目录默认自动勾选（按用户确认）
- [x] 8. 更新 OpenAPI + 生成 Go/TS 客户端类型
- [x] 9. 补齐后端/前端测试并完成回归验证

## Review（Admin 局部补同步交互重构 / 实施结果）

- 后端能力：
  - 新增 `POST /api/v1/admin/roots/inspect`（目录详情拉取，支持部分成功）。
  - `POST /api/v1/admin/sync` 支持 `preserve_root_catalog`。
  - 当 `force_rebuild=true` 且提供 `root_folder_ids` 时返回 `400`，避免危险组合误用。
  - `SyncProgressState` 新增目录册字段：`catalogRoots` / `catalogRootNames` / `catalogRootProgress`。
  - full scoped run 在 `preserve_root_catalog=true` 时合并历史 `rootProgress`，避免根目录详情被覆盖。
- 前端交互：
  - 目录输入框改为“拉取目录详情”用途，不再直接触发同步。
  - 新增“拉取目录详情”按钮，独立 loading/error 状态。
  - `SyncProgressDisplay` 在根目录详情行新增 toggle；Admin 页面按勾选根目录发起全量补同步。
  - 新拉取目录默认自动勾选（按用户选择 `1` 落地）。
  - 前端渲染优先使用 `catalog*`，缺失时回退 `rootProgress/rootNames`。
- OpenAPI 与生成：
  - 已更新 `api/openapi.yaml`。
  - 已执行 `GOCACHE=/tmp/go-build go generate ./api/...`。
  - 已执行 `cd web && bun run generate`。
- 测试验证：
  - `GOCACHE=/tmp/go-build go test ./internal/httpx ./internal/service -count=1` 通过。
  - `GOCACHE=/tmp/go-build go test ./...` 通过。
  - `cd web && bun vitest run src/components/admin-page.test.tsx src/hooks/use-sync-progress.test.ts src/components/sync-progress-display.test.tsx` 通过。
  - `cd web && bun vitest run` 通过（22 files / 187 tests）。
  - 容器链路：
    - `docker compose -f docker-compose.ci.yml --profile e2e up --build -d --wait --wait-timeout 180` 通过。
    - `BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh` 通过（34/34）。
    - `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright` 通过（31 passed + 1 flaky 重试后通过）。
- 额外检查：
  - `git diff --check` 通过。
  - `cd web && bun run typecheck` 仍有历史性问题（`routeTree.gen` 缺失导致路由类型报错），与本次改动无直接耦合；本次新增测试中的类型错误已修复。

## 新任务：Admin 局部补同步交互优化（移除手动目录 ID 输入）

- [x] 1. 移除 Admin 页面手动输入目录 ID 的 UI 与解析逻辑
- [x] 2. 保留“刷新目录详情”按钮并改为基于当前根目录列表执行
- [x] 3. 更新前端测试用例，验证无输入框时的局部补同步流程
- [x] 4. 运行相关前端测试回归

## Review（Admin 局部补同步交互优化 / 移除手动目录 ID 输入）

- 交互调整：
  - Admin 页面不再提供手动输入目录 ID 的入口，避免与下方 toggle 重复。
  - 保留“刷新目录详情”按钮，仅刷新当前已知根目录（`catalogRoots`/`rootProgress` 推导）。
  - 当尚无根目录列表时，按钮禁用并提示需先完成一次全量同步。
- 实现范围：
  - 仅前端改动（`web/src/components/admin-sync-page.tsx`、`web/src/components/admin-page.test.tsx`），后端接口保持兼容。
- 验证：
  - `cd web && bun vitest run src/components/admin-page.test.tsx src/components/sync-progress-display.test.tsx src/hooks/use-sync-progress.test.ts` 通过（33 tests）。

## 新任务：Connect-RPC 新 review 收敛（Brainstorming 设计）

- [x] 1. 读取最新 `review.md`，对照当前 Connect-RPC 迁移状态做建议分流（已完成/待采纳/暂缓）
- [x] 2. 并行调研现有代码架构、schema 最佳实践与 BDD 场景（子代理）
- [x] 3. 产出设计文档包到 `docs/plans/2026-02-24-connect-rpc-review-alignment-design/`
- [x] 4. 回填评审结论与后续执行边界（避免重复返工）

## Review（Connect-RPC 新 review 收敛 / Brainstorming 设计）

- 结论：
  - 新 `review.md` 中关于 `enum 0 值使用 *_UNSPECIFIED`、`connect-es/query-es`、Connect 后端渐进接入的建议，当前分支已基本落地。
  - 本轮真正需要规划的后续项是：
    - 在 `.proto` 中补 `protovalidate` 规则注解（利用已接入的 validation interceptor）
    - 明确 `google.protobuf.Timestamp` 迁移策略与触发条件（继续暂缓到后续批次）
- 明确不做（本设计批次）：
  - 不立即把时间戳字段从 `int64` 全量切到 `Timestamp`
  - 不新建 `internal/rpc` 包并搬迁现有 Connect handler（避免与功能推进耦合）
  - 不改 Connect 路由路径前缀（保持现有生成路径与 Echo 挂载方式）
- 设计产物：
  - `docs/plans/2026-02-24-connect-rpc-review-alignment-design/_index.md`
  - `docs/plans/2026-02-24-connect-rpc-review-alignment-design/architecture.md`
  - `docs/plans/2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md`
  - `docs/plans/2026-02-24-connect-rpc-review-alignment-design/best-practices.md`
- 后续建议（进入 writing-plans 前的推荐范围）：
  - 优先做 `protovalidate` 注解的增量落地（先覆盖 `StartSyncRequest`、`InspectRootsRequest`、分页类请求）
  - 为 validation interceptor 增加正反向测试（命中规则 / 无规则 no-op）
  - 将 Timestamp 迁移单独立项，先完成影响面清点与兼容方案选择

## 新任务：Connect-RPC protovalidate 增量落地（Writing Plans）

- [x] 1. 基于 `docs/plans/2026-02-24-connect-rpc-review-alignment-design/` 输出实施计划目录
- [x] 2. 按 BDD 场景拆分 Red/Green 任务（Admin 规则、分页规则、兼容性守门）
- [x] 3. 为每个任务补齐依赖、影响文件与验证命令
- [x] 4. 回填执行移交说明（下一步进入 `executing-plans`）

## Review（Connect-RPC protovalidate 增量落地 / 实施计划）

- 计划目录：
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/_index.md`
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/task-001-red-admin-validation-hit-tests.md`
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/task-002-green-admin-proto-validation-rules.md`
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/task-003-red-search-pagination-validation-hit-tests.md`
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/task-004-green-search-pagination-proto-validation-rules.md`
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/task-005-green-noop-and-business-guard-regression.md`
  - `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/task-006-verification-and-timestamp-compat-gate.md`
- 范围收敛：
  - 本批次只做 `protovalidate` 注解与校验测试补齐。
  - 明确不做 `Timestamp` 字段迁移与 `internal/rpc` 包抽离。
