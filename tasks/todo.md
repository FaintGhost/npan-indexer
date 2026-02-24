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

## 新任务：Connect-RPC protovalidate 增量落地（Executing Plans）

- [x] 1. 新增 validation interceptor 级测试（Admin 规则命中、分页规则命中、无规则 no-op）
- [x] 2. 在 `buf.yaml` 增加 protovalidate 依赖并更新 `buf.lock`
- [x] 3. 在 `proto/npan/v1/api.proto` 补充 Admin 请求规则（`StartSyncRequest`、`InspectRootsRequest`）
- [x] 4. 在 `proto/npan/v1/api.proto` 补充分页请求规则（`AppSearchRequest`、`LocalSearchRequest`）
- [x] 5. 回归业务语义防线测试（`force_rebuild + scoped roots`）
- [x] 6. 执行 lint/generate/测试回归并确认 Timestamp 守门条件

## Review（Connect-RPC protovalidate 增量落地 / 实施结果）

- 代码改动：
  - 新增 `internal/httpx/connect_validation_interceptor_test.go`：
    - `TestConnectValidationInterceptor_AdminStartSyncHitRule`
    - `TestConnectValidationInterceptor_PaginationHitRule`
    - `TestConnectValidationInterceptor_NoRuleMessageNoop`
  - 更新 `internal/httpx/connect_admin_test.go`：
    - 强化 `force_rebuild + scoped roots` 业务防线断言（错误信息包含 `force_rebuild`）
  - 更新 `buf.yaml` 并新增 `buf.lock`，引入 `buf.build/bufbuild/protovalidate` 依赖。
  - 更新 `proto/npan/v1/api.proto`：
    - 引入 `buf/validate/validate.proto`
    - `StartSyncRequest` 增加 root/dept IDs 与 worker 参数范围约束
    - `InspectRootsRequest` 增加 `folder_ids` 非空+正整数约束
    - `AppSearchRequest`、`LocalSearchRequest` 增加分页范围约束
  - 生成产物更新：
    - `gen/go/npan/v1/api.pb.go`
    - `gen/ts/npan/v1/api_pb.ts`
- BDD Red/Green 证据：
  - Red（Task 001）：`TestConnectValidationInterceptor_AdminStartSyncHitRule` 初次运行失败（预期 `invalid_argument` 未命中）
  - Green（Task 002）：补 Admin 规则后同用例转绿
  - Red（Task 003）：`TestConnectValidationInterceptor_PaginationHitRule` 初次运行失败（预期 `invalid_argument` 未命中）
  - Green（Task 004）：补分页规则后同用例转绿
- 验证结果：
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf dep update` 通过
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf lint` 通过
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf generate` 通过
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect.*(Validation|Admin).*' -count=1` 通过
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -run 'Connect|Routes|Health|Admin' -count=1` 通过
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./... -count=1` 通过（先创建 `web/dist/.gitkeep` 占位）
  - `git diff --check` 通过
- Timestamp 守门结论：
  - 本批次未引入 `google.protobuf.Timestamp`，`proto/npan/v1/api.proto` 中相关时间字段仍为 `int64`（`started_at` / `updated_at` 等），符合“兼容性优先、暂缓迁移”的约束。

## 新任务：Connect-RPC Timestamp 迁移（Brainstorming 设计）

- [x] 1. 盘点时间字段影响面（proto、Go DTO、service、存储、前端消费与测试）
- [x] 2. 对比迁移策略（一次性替换 vs 双字段过渡）并收敛推荐方案
- [x] 3. 产出设计文档包到 `docs/plans/2026-02-24-connect-rpc-timestamp-migration-design/`
- [x] 4. 回填评审结论与迁移边界（本批次不直接删旧 `int64` 字段）

## Review（Connect-RPC Timestamp 迁移 / Brainstorming 设计）

- 核心结论：
  - 不能“就地把已有字段类型从 `int64` 改成 `Timestamp`”，这会破坏 protobuf 向后兼容。
  - 推荐采用“双字段过渡”：
    - 保留现有 `int64` 字段（兼容 REST/CLI/存储与老客户端）
    - 新增 `*_ts`（`google.protobuf.Timestamp`）字段给 Connect 新客户端优先消费
  - 过渡期由服务端同时填充新旧字段，前端优先读新字段，缺失时回退旧字段。
- 范围边界：
  - 本批次目标是“增量引入 Timestamp 字段并打通兼容消费链路”。
  - 本批次不移除旧 `int64` 字段，不修改进度持久化结构，不做破坏式清理。
- 设计产物：
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-design/_index.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-design/architecture.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-design/best-practices.md`

## 新任务：Connect-RPC Timestamp 迁移（Writing Plans）

- [x] 1. 基于 Timestamp 设计文档产出实施计划目录
- [x] 2. 按 BDD 场景拆分 Red/Green 任务（契约、后端映射、前端消费、兼容守门）
- [x] 3. 为任务补齐依赖、影响文件与验证命令
- [x] 4. 回填执行移交说明（下一步进入 `executing-plans`）

## Review（Connect-RPC Timestamp 迁移 / 实施计划）

- 计划目录：
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/_index.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-001-red-proto-descriptor-timestamp-fields.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-002-green-proto-add-timestamp-sidecar-fields.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-003-red-backend-connect-progress-timestamp-tests.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-004-green-backend-progress-timestamp-mapping.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-005-red-frontend-timestamp-fallback-tests.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-006-green-frontend-timestamp-consumer-adapter.md`
  - `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/task-007-verification-and-compatibility-gate.md`
- 迁移策略已固化：
  - 先“新增 + 双写 + 回退读取”，再评估后续清理旧字段；
  - 兼容优先，不做同批次破坏式替换。

## 新任务：Connect-RPC Timestamp 迁移（Executing Plans）

- [x] 1. 新增 proto descriptor Red 测试，验证 `*_ts` 字段缺失（Red）
- [x] 2. 在 `proto/npan/v1/api.proto` 为进度消息新增 Timestamp sidecar 字段并生成代码（Green）
- [x] 3. 新增后端 progress DTO 映射 Red 测试，验证 `*_ts` 未填充（Red）
- [x] 4. 在 `internal/httpx/connect_admin.go` 实现 `int64 -> Timestamp` 双字段映射（Green）
- [x] 5. 新增前端 hook / 组件 Timestamp 优先与回退测试（Red）
- [x] 6. 实现前端时间兼容适配层，支持 `Timestamp | int64` 消费（Green）
- [x] 7. 执行 lint/generate/后端/前端回归并完成兼容门槛检查

## Review（Connect-RPC Timestamp 迁移 / 实施结果）

- 代码改动：
  - `proto/npan/v1/api.proto`
    - 引入 `google/protobuf/timestamp.proto`
    - 为以下消息新增 Timestamp sidecar 字段（保留旧 `int64` 不动）：
      - `CrawlStats.started_at_ts` / `ended_at_ts`
      - `RootSyncProgress.updated_at_ts`
      - `SyncProgressState.started_at_ts` / `updated_at_ts`
  - 生成产物更新：
    - `gen/go/npan/v1/api.pb.go`
    - `gen/ts/npan/v1/api_pb.ts`
  - 新增后端测试：
    - `internal/httpx/connect_timestamp_descriptor_test.go`（descriptor 字段存在性）
  - 更新后端实现与测试：
    - `internal/httpx/connect_admin.go` 新增 `millisToProtoTimestamp(...)` 并在 progress DTO 转换中双写 `*_ts`
    - `internal/httpx/connect_admin_test.go` 增加 `toProtoSyncProgressState` sidecar 映射断言
  - 前端兼容适配：
    - `web/src/lib/sync-schemas.ts`
      - 扩展 `SyncProgress`/`CrawlStats`/`RootProgress` schema，允许可选 Timestamp sidecar 字段
      - 新增 `timestampLikeToMillis` / `preferTimestampMillis`
    - `web/src/hooks/use-sync-progress.ts`
      - 新增 `normalizeSyncProgressTimestamps(...)`，在拉取 progress 后统一归一化时间字段
    - `web/src/components/sync-progress-display.tsx`
      - `ElapsedTime` 改为优先使用 `*_ts` sidecar，回退旧 `int64`
    - 测试更新：
      - `web/src/hooks/use-sync-progress.test.ts`
      - `web/src/components/sync-progress-display.test.tsx`
- BDD Red/Green 证据：
  - Red（Task 001）：`TestConnectTimestampDescriptor_ProgressMessagesHaveSidecarFields` 失败，提示 `started_at_ts/updated_at_ts` 字段缺失
  - Green（Task 002）：新增 proto sidecar 字段并生成后，descriptor 测试转绿
  - Red（Task 003）：`TestToProtoSyncProgressState_PopulatesTimestampSidecar` 失败，提示 `started_at_ts` 未填充
  - Green（Task 004）：后端 DTO 映射补齐后测试转绿
  - Red（Task 005）：前端新增两条用例失败（hook 未优先取 sidecar；组件 elapsed 仍走旧值）
  - Green（Task 006）：前端适配层上线后两条用例转绿
- 验证结果：
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf lint` 通过
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf generate` 通过
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -count=1` 通过
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./... -count=1` 通过（先创建 `web/dist/.gitkeep` 占位）
  - `cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/sync-progress-display.test.tsx` 通过（30 tests）
  - `cd web && bun vitest run` 通过（22 files / 189 tests）
  - `git diff --check` 通过
- 兼容门槛检查（Task 007）：
  - 旧 `int64` 字段仍保留在 `proto/npan/v1/api.proto`（未做破坏式替换）。
  - `internal/models` 与 `internal/storage` 持久化结构未改为 `Timestamp`，仍沿用 `int64` 存储。
  - 服务端仅在 Connect DTO 输出层新增 sidecar 双写，符合“最小影响面”策略。

## 新任务：Connect-ES 兼容性补强（Review 跟进）

- [x] 1. 复核 review 与现有后端实现，确认阻塞项范围
- [x] 2. 补齐 CORS `AllowHeaders` 与 `ExposeHeaders`，覆盖 Connect-ES 浏览器调用场景
- [x] 3. 新增/更新单测，锁定 CORS 配置回归
- [x] 4. 运行 `go test ./internal/httpx -count=1` 并回填结果

## Review（Connect-ES 兼容性补强 / 实施结果）

- 目标：
  - 处理 review 中标记的阻塞项：Connect 协议自定义 Header 未放行、错误头未暴露给浏览器。
- 范围：
  - `internal/httpx/middleware_security.go`
  - `internal/httpx/middleware_security_test.go`
- 变更：
  - `CORSConfig(...).AllowHeaders` 新增：
    - `Connect-Protocol-Version`
    - `Connect-Timeout-Ms`
    - `Grpc-Timeout`
  - `CORSConfig(...).ExposeHeaders` 新增：
    - `Connect-Error-Reason`
    - `Connect-Error-Details`
  - 新增单测：
    - `TestCORSConfig_ConnectHeadersIncluded`
    - `TestParseCORSOrigins_TrimEmpty`
- 验证：
  - `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -count=1` 通过
- 备注：
  - 当前改动覆盖的是 CORS 配置 helper；若未来需要跨域直连（非同源/非 dev proxy），需确认服务启动路径已实际挂载 `middleware.CORSWithConfig(CORSConfig(...))`。
- 非本轮（记录）：
  - Validation 错误结构化为 `errdetails.BadRequest` 属体验优化项，后续单独设计实施。

## 新任务：Connect-Query 前端迁移（Stage 4 / Admin 第一批）

- [x] 1. 复核新 review 与当前代码状态，确认 CORS 项已完成、迁移入口为 Admin
- [x] 2. 在 `web/src/main.tsx` 接入 `TransportProvider + QueryClientProvider`
- [x] 3. 将 `useSyncProgress` 的网络层迁移为 `@connectrpc/connect-query`（保持外部 API 不变）
- [x] 4. 增加 proto-to-UI 适配层，处理 enum/int64/timestamp 到现有 UI 数据结构
- [x] 5. 更新 Admin 相关单测（provider wrapper + Connect 路由 mock）并通过回归
- [x] 6. 回填本轮 Review 结果与验证命令

## Review（Connect-Query 前端迁移 / 实施结果）

- 目标：
  - 开始 Stage 4 前端迁移，优先替换 Admin 同步链路，减少手写 REST fetch 逻辑。
- 范围（第一批）：
  - `web/src/main.tsx`
  - `web/src/hooks/use-sync-progress.ts`
  - `web/src/components/admin-page.test.tsx`
  - `web/src/hooks/use-sync-progress.test.ts`
  - 新增 Connect 适配层/测试辅助文件（按实现落地）
- 关键实现：
  - 全局接入 Connect Query Provider：
    - `web/src/main.tsx` 包裹 `TransportProvider + QueryClientProvider`
    - 新增 `web/src/lib/connect-transport.ts`（全局 `QueryClient` / Connect transport / API Key 拦截器）
  - Admin Hook 网络层迁移：
    - `web/src/hooks/use-sync-progress.ts` 改为使用 `@connectrpc/connect-query` 的 `useQuery/useMutation`
    - 保留外部返回 API 与轮询/乐观更新行为，`AdminSyncPage` 无需重写
  - Proto 适配层：
    - 新增 `web/src/lib/connect-admin-adapter.ts`
    - 处理 `enum`（`SyncStatus`/`SyncMode`）、`int64(bigint)`、`Timestamp` 到现有 UI schema 结构的映射
  - 生成产物与路径修复：
    - `buf.gen.yaml` 中 TS 插件输出路径切换到 `web/src/gen`
    - `buf.build/bufbuild/es` 插件开启 `include_imports: true`，补齐 `buf/validate/validate_pb.ts`
    - 解决前端导入 `api_pb.ts` 时的 `buf/validate` 悬空依赖问题
  - 测试辅助：
    - 新增 `web/src/tests/test-providers.tsx`，为组件/Hook 测试提供 QueryClient + Transport Provider
  - 本地开发代理：
    - `web/vite.config.ts` 增加 `^/npan\\.v1\\.` 代理到后端，支持 Vite dev 下 Connect 路由
- 测试改造：
  - `web/src/hooks/use-sync-progress.test.ts`
    - 增加 Provider wrapper
    - 将 Admin 同步链路 mock 从 REST 路由切换到 Connect 路由
    - 适配 Connect 的 proto JSON（enum 名称、int64 字符串、Timestamp RFC3339）
  - `web/src/components/admin-page.test.tsx`
    - 增加 Provider wrapper
    - Admin 进度/启动/目录详情改为 Connect mock（`useAdminAuth` 在下一批次已迁移到 Connect）
- 验证：
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf generate` 通过
  - `XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf lint` 通过
  - `cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/admin-page.test.tsx` 通过（19 tests）
  - `cd web && bun vitest run` 通过（22 files / 189 tests）
  - `git diff --check` 通过
- 备注：
  - `cd web && bun run typecheck` 当前仍失败（缺少 `routeTree.gen` / 路由类型未生成），与本轮 Connect 迁移改动无直接关系，属于现有前端生成链路前置条件问题。
- 兼容原则：
  - `AdminSyncPage` 组件层尽量不改，先保持 `useSyncProgress` 返回结构稳定；
  - 先完成 Admin，再评估 Search 迁移与旧 `api-client` 清理窗口。

## 新任务：Connect-Query 前端迁移（Stage 4 / Search + Admin Auth）

- [x] 1. 盘点 `api-client` 在前端的剩余使用点，确认目标为 `use-search` + `use-admin-auth`
- [x] 2. 新增 App proto-to-UI 适配层（搜索结果映射）
- [x] 3. 将 `use-search` 迁移到 Connect（保持 debounce/分页/去重行为）
- [x] 4. 将 `use-admin-auth` 的 API Key 校验迁移到 Connect 调用
- [x] 5. 更新 Search/AdminAuth 相关测试（Connect 路由 mock / provider wrapper）并回归
- [x] 6. 回填本轮 Review 结果与验证记录

## Review（Connect-Query 前端迁移 / Search + Admin Auth）

- 目标：
  - 将 Search 与 Admin 鉴权链路从 `api-client` 切到 Connect，收敛手写 REST 调用。
- 范围：
  - `web/src/hooks/use-search.ts`
  - `web/src/hooks/use-admin-auth.ts`
  - `web/src/lib/connect-app-adapter.ts`
  - `web/src/hooks/use-search.test.ts`
  - `web/src/components/search-page.test.tsx`
  - `web/src/hooks/use-admin-auth.test.ts`
  - `web/src/components/admin-page.test.tsx`
  - `web/src/tests/accessibility.test.tsx`
- 关键改动：
  - `use-search` 改为 `@connectrpc/connect-query` 的 `useMutation(appSearch)`，保留：
    - debounce（280ms）
    - loadMore 分页
    - 按 `source_id` 去重
    - `searchImmediate`/`reset` 行为
  - 新增 `connect-app-adapter`，把 `AppSearchResponse`（enum/int64/camelCase）映射到现有 UI `SearchResponse`（number/snake_case）。
  - `use-admin-auth` 改为 `callUnaryMethod(..., getSyncProgress)` 进行 API Key 校验：
    - `Code.NotFound` 仍视为“鉴权通过但暂无进度”
    - `Code.Unauthenticated` 映射“API Key 无效”
  - Search/Admin/Auth 相关测试统一改为 Connect 路由 mock，且对依赖 Connect Query 的测试挂载 provider wrapper。
- 清理结果：
  - `api-client` 在业务代码中已不再用于 Admin/Search（仅 `use-download` 仍在使用）。
- 验证：
  - `cd web && bun vitest run src/hooks/use-search.test.ts src/components/search-page.test.tsx src/hooks/use-admin-auth.test.ts src/components/admin-page.test.tsx` 通过（26 tests）
  - `cd web && bun vitest run` 通过（22 files / 189 tests）
  - `git diff --check` 通过

## 新任务：Stage 4 AdminSyncPage 组件内直连 Connect（已完成）

- [x] 1. 移除 `AdminSyncPage` 对 `use-sync-progress` 的依赖，改为组件内直接使用 `@connectrpc/connect-query` 的 `useQuery/useMutation`
- [x] 2. 保持现有交互行为不变（刷新目录详情、全量勾选目录、强制重建约束、取消同步、消息提示）
- [x] 3. 回归验证 `admin-page` 相关测试与前端全量测试
- [x] 4. 回填本轮 Review 记录（改动点 + 验证结果）

## Review（Stage 4 / AdminSyncPage 组件内直连 Connect）

- 关键改动：
  - `web/src/components/admin-sync-page.tsx` 已移除 `useSyncProgress` 依赖。
  - 组件内直接接入 `@connectrpc/connect-query`：
    - `useQuery(getSyncProgress)` + 手动 `refetch`
    - `useMutation(startSync / inspectRoots / cancelSync)`
  - 保留并内联了原有同步状态机能力：
    - 首次加载、运行中轮询、`NotFound` 视为“暂无进度”
    - 目录详情拉取后写回 `catalogRoots/catalogRootProgress`
    - 启动同步后的运行态更新与取消后的刷新
  - UI 交互与文案保持不变（刷新目录详情、全量勾选、强制重建确认、取消同步确认）。
- 额外修正：
  - 启动/取消同步结果改为基于 mutation 返回值判断成功，避免依赖异步 error 状态导致误提示成功。
- 验证：
  - `cd web && bun vitest run src/components/admin-page.test.tsx` 通过（5 tests）
  - `cd web && bun vitest run` 通过（22 files / 189 tests）
  - `git diff --check` 通过

## 新任务：Stage 4 清理遗留 use-sync-progress（继续2）

- [x] 1. 删除未再被生产代码使用的 `web/src/hooks/use-sync-progress.ts`
- [x] 2. 删除对应测试 `web/src/hooks/use-sync-progress.test.ts` 并确认测试套件收敛
- [x] 3. 运行前端回归测试（重点 `admin-page` + 全量 vitest）
- [x] 4. 回填本轮 Review（清理范围与验证结果）

## Review（Stage 4 / 清理遗留 use-sync-progress）

- 清理范围：
  - 删除 `web/src/hooks/use-sync-progress.ts`
  - 删除 `web/src/hooks/use-sync-progress.test.ts`
- 背景：
  - `AdminSyncPage` 已在上一轮改为组件内直接使用 Connect Query/Mutation，`use-sync-progress` 不再被生产代码引用。
- 结果：
  - `rg -n "use-sync-progress" web/src -g'*.ts*'` 无匹配，遗留引用已清零。
  - 前端测试套件从 `22 files / 189 tests` 收敛为 `21 files / 175 tests`（仅因移除该 hook 测试）。
- 验证：
  - `cd web && bun vitest run src/components/admin-page.test.tsx` 通过（5 tests）
  - `cd web && bun vitest run` 通过（21 files / 175 tests）
  - `git diff --check` 通过
