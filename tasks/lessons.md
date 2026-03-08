# Lessons

## 2026-03-08

- 用户纠正：检索第三方搜索文档时，必须先查项目实际使用产品的官方文档；当前场景是 Meilisearch，不应直接拿 Algolia 文档当一手依据。
- 规则：
  - 涉及 Meilisearch + InstantSearch 集成时，先查 `meilisearch.com` 官方文档。
  - 只有在 Meilisearch 官方文档缺少底层实现细节时，才下钻到上游 InstantSearch/Algolia 文档，并明确标注它只是兼容层参考，不是当前项目的一手来源。
  - 若用户已指定官方文档入口，后续检索必须以该入口为主，不要自行切换到其他厂商文档。
- 用户纠正：当前 InstantSearch 的目标不是“严格保持首批官方默认直连模式”，而是“官方行为 + 关键对齐”——既要恢复 search-as-you-type，也要补齐最关键的 legacy 搜索语义差异。
- 规则：
  - 当线上反馈已证明“官方默认模式”产生明显业务回退时，必须及时收敛目标为“官方交互行为 + 关键业务语义对齐”，不能继续拿旧设计中的范围约束当成不修的理由。
  - 搜索体验纠偏时，要先区分“交互行为偏差”（如 submit-only vs search-as-you-type）与“结果语义偏差”（如 query preprocess、默认过滤），分别定位并建模到设计/计划中。
  - 若现有 design 文档已与用户新确认的目标冲突，必须先补新的 design，再进入 writing-plans；不要直接复用过时 design 继续拆计划。

## 2026-03-04

- 用户偏好：前端视觉更偏浅蓝色，不要过深或发黑。
- 规则：
  - 若用户明确给出色彩偏好，先统一全局主题变量，再收口组件级 accent，避免页面局部残留旧色。
  - 优先保留错误红、警告黄等语义色，只替换主品牌色与成功提示色，保证状态可辨识度。
  - 深色按钮需要控制饱和度和明度，避免接近黑色（如 `slate-900`）导致整体观感过重。
- 用户纠正：主页初始态应模仿 Google 搜索，视口内仅居中搜索框，不应出现滚动条；搜索后才进入 docked 列表模式。
- 规则：
  - Hero 初始态下，结果区必须从文档流中塌陷（例如 `max-height: 0` + `overflow: hidden`），避免“透明但仍占位”导致滚动条。
  - 若主内容在 Hero 态保留容器节点，必须同时清除会制造额外高度的留白（如 `main` 的 `padding-bottom`、结果区 `margin-top`）。
  - 搜索后进入 docked 模式时再恢复结果区占位与滚动，不影响结果列表交互。
- 用户纠正：docked 模式下“键盘删空输入框”应与“点击输入框右侧叉号”行为一致，都要回到居中的 hero 模式。
- 规则：
  - 输入变空时（`trim()==''`）必须执行与 `handleClear` 对齐的状态收敛：清 `activeQuery`、`setDocked(false)`、重置筛选与 URL 参数。
  - 禁止出现“叉号清空会回中，但键盘删空不回中”的双轨交互。

## 2026-03-03

- 用户纠正：Web 侧命令优先使用 Bun（如 `bun run test`），不要默认走 npm。
- 规则：
  - 进入 `web/` 后先读取 `package.json` scripts，再用 `bun run <script>` 执行。
  - 当 Bun 测试失败时，先定位失败用例并修复，不直接切回 npm 作为默认路径。
  - 命令输出中若出现与功能无关的环境告警（例如 shell prompt 写缓存失败），应在结果里说明“可忽略”，并聚焦测试结论。
- 用户纠正：涉及契约与类型链路时，先看 `CLAUDE.md/README` 确认“生成代码边界”再改。
- 规则：
  - 先区分 `web/src/gen`（Buf 生成）与 `web/src/lib/*`（手写映射/校验层），禁止直接改生成产物兜底。
  - 处理前端路由类型问题时，优先补齐 route tree 生成链路，再跑 `typecheck`，不要只靠临时文件绕过。

## 2026-02-24

- 用户纠正：项目已全面迁移到 Connect-RPC，新增能力应以 `proto/npan/v1/api.proto` 为唯一契约源。
- 规则：
  - 在实现前先检查 `proto/npan/v1/api.proto` 是否已覆盖所需字段。
  - 若有字段变更，先改 proto，再执行 `buf lint && buf generate`。
  - 禁止在运行时代码新增 `/api/v1/*` 路径；历史文档（`docs/archive/**`、`docs/plans/**`、`tasks/**`）可保留。
- 用户纠正：新增功能后必须跑完整验证链（前后端单测、冒烟测试、E2E 测试）。
- 规则：
  - 默认收口命令至少包含：`go test ./...`、`cd web && bun vitest run`。
  - 若仓库提供 Docker 冒烟与 E2E（本项目有），必须执行并记录结果：
    - `docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120`
    - `./tests/smoke/smoke_test.sh`
    - `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright`
    - `docker compose -f docker-compose.ci.yml --profile e2e down --volumes`
  - 若环境缺失依赖（如 `make`/Docker 权限），需明确说明并给出等价命令或申请提权后执行。
- 用户纠正：Docker Publish workflow 出现 `Skip output ... may contain secret` 与 `tag is needed when pushing to registry`。
- 规则：
  - 禁止将由 `secrets.*` 参与拼接得到的值作为 **job outputs** 跨 job 传递。
  - 镜像名（尤其含 DockerHub 用户名）统一在各 job 内本地计算并使用 **step outputs**。
  - 出现 `buildx tag is needed` 时，优先排查 `name=` 是否为空（变量/输出被跳过）。
- 用户纠正：merge 阶段 `imagetools create` 出现 `failed to parse source "...@sha256:"`。
- 规则：
  - 当 `*` 可能匹配多个文件时，禁止用 `printf` 一次性混合拼接镜像引用（易触发格式串重复）。
  - manifest source 统一使用显式循环：逐个 digest 组装 `${TARGET_IMAGE}@sha256:<digest>`。
  - 对 shell 拼接逻辑，优先先在本地用多输入样例验证展开行为再提交。
- 用户纠正：merge 阶段 `docker.io/...@sha256:... not found`。
- 规则：
  - 多 registry 发布时，若 build 阶段只保存单一 digest artifact，必须明确该 digest 的“来源 registry”。
  - `imagetools create` 的 source 优先统一使用单一 canonical registry（本项目选 GHCR），避免跨 registry digest 不一致/不可见。
  - 若需要双 registry 发布，推荐策略是“单 registry push-by-digest + merge 阶段跨 registry 打标签复制”。
- 用户纠正：测试应仅在源码变更时触发。
- 规则：
  - CI workflow 触发优先使用 `paths` 白名单，而不是仅靠 `paths-ignore`。
  - 将 `tasks/**`、文档、临时记录等高频非源码变更排除在 CI 测试触发条件之外。
  - 涉及 workflow/Docker 构建链路调整时，默认先做静态检查；是否跑全量测试根据是否触及源码决定。
- 用户纠正：`go:embed all:dist` 在 CI 仍失败（checkout 缺少 `web/dist` 目录）。
- 规则：
  - `go:embed` 依赖目录若不是稳定跟踪产物，必须在 CI 对应 job 前置创建占位目录/文件。
  - 不能仅依赖本地存在但未跟踪文件（如 `.gitkeep`）；要以 GitHub checkout 的最小状态为准验证。
- 用户纠正：先修问题再测试；并且禁止用类型断言/`any` 兜底类型错误。
- 规则：
  - 当 `lint`/`typecheck` 失败时，先修复本轮引入的问题，再进入 smoke/E2E 长链路测试。
  - 测试代码中读取 `request.json()` 一律先用 `unknown`，再通过类型守卫收窄（如 `assertRecord`/`getRecord`）。
  - 禁止通过 `as` 或 `any` 绕过类型系统；若需从可空值读取，使用返回值型守卫（如 `requireValue`）而非断言。
- 用户纠正：E2E 不应默认拉满超时，很多场景应快速失败。
- 规则：
  - `waitForRequest`/`waitForResponse` 默认使用短超时（优先 3s-5s），仅对已知慢路径使用 10s+。
  - UI 断言超时按场景分级：即时交互 3s、常规异步 5s、确实重负载流程才给 10s。
  - 出现批量超时时先检查“等待条件是否匹配当前协议/路径”（如 REST -> Connect 迁移），再考虑放宽超时。
- 用户纠正：线上从 DockerHub 拉取 `latest` 后仍表现为旧版本，需要排查发布产物而非仅看代码分支。
- 规则：
  - 发布后优先核对镜像标签对应的 `org.opencontainers.image.revision`（或 `sha-<commit>` tag）是否匹配目标提交。
  - 多平台 manifest 若行为异常，需检查各平台 digest 是否被错误混入历史条目（尤其 self-hosted runner 的临时目录残留）。
  - CI 上传构建 digest artifact 时必须使用“每次运行唯一且清理过”的临时目录，不能复用固定 `/tmp` 路径。
- 用户纠正：冒烟与 E2E 覆盖不足，遗漏了 Connect streaming 在中间件包装下的运行时错误（`http.Flusher` 缺失）。
- 规则：
  - 所有会包装 `http.ResponseWriter` 的中间件，必须透传 streaming 相关接口（至少 `http.Flusher`，并保留 `Unwrap` 能力）。
  - 新增/修改 Connect server streaming 能力时，必须加“启用全部中间件（含 Prometheus）”的后端回归测试。
  - 每次修复线上 streaming 问题后，E2E 至少补一个可观测守卫（console/page error 或网络级断言），避免仅靠功能按钮通过。
  - 服务端改动验证 E2E 时，默认使用 `docker compose ... up --build`，避免旧镜像掩盖修复结果。
