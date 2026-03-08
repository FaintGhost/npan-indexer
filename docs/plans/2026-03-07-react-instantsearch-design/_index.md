# React InstantSearch 直连 Meilisearch 设计

## Context

当前公开搜索页 `/` 仍通过 Connect `AppService.AppSearch` 发起搜索，请求链路和状态管理集中在：

- `web/src/routes/index.lazy.tsx`
- `internal/httpx/connect_app_auth_search.go`
- `internal/search/meili_index.go`

当前实现已经具备：

- 基于 Meilisearch 的全文检索与 `name` 高亮
- 前端无限滚动、手工 debounce、前台恢复 refetch
- 下载链路通过 `AppService.AppDownloadURL` 生成受控下载地址
- 本地扩展名筛选（`web/src/lib/file-category.ts`），但它只作用于“前端已加载结果”，并不是真正的服务端 facet/filter

已确认本次边界：

- 采用 **方案 A：浏览器直连 Meilisearch**
- 搜索页访问边界是 **公共搜索**，不同访问者共享同一份可见结果集
- 必须 **严格遵循 Meilisearch 官方 React InstantSearch 最佳实践**
- 下载、同步、管理后台仍保留现有 Connect/Go 后端边界

## Requirements

### Must

- 使用官方推荐栈：`react-instantsearch` + `@meilisearch/instant-meilisearch`
- 搜索状态由 `InstantSearch` 统一管理，包括 query、page、refinements 和 routing
- 浏览器只使用 **search-only key**，严禁暴露 Meilisearch admin/master key
- 搜索页切换为直连 Meilisearch 后，下载仍继续走 `AppService.AppDownloadURL`
- 扩展名分类从“前端二次过滤”升级为“索引字段 + facet/filter”
- 搜索 URL 必须可刷新恢复、可分享、支持前进后退
- 首批落地必须保留 `AppSearch` 作为灰度和回滚兜底链路

### Should

- 通过运行时搜索配置接口下发 `host/indexName/searchKey`，避免把搜索配置硬编码进前端构建产物
- 优先使用 InstantSearch hooks 复用现有 `SearchInput` / `FileCard` / 筛选 chips，而不是整体替换为默认 widgets DOM
- 将公开搜索返回字段收敛到最小必要集合，避免把无关元数据暴露给浏览器
- 保留现有视觉壳层（hero/docked 布局、按钮、空态、错误态）

### Won't

- 本批次不设计多租户隔离或 tenant token
- 本批次不改动管理后台与同步主流程
- 本批次不强行保留后端 `preprocessQuery()` 与 `All -> Last` fallback 行为；先按官方直连模式落地，再以验证结果决定是否补充适配层
- 本批次不删除 `AppSearch`

## Option Analysis

### Option A（已选）: React InstantSearch + 浏览器直连 Meilisearch

- 优点：完全符合官方文档路径；InstantSearch 可直接提供 routing、InfiniteHits、Highlight、facet/filter 等能力；前端不再维护自定义搜索状态机
- 代价：需要补公开搜索配置、安全边界、索引 facet 字段与测试体系改造

### 被拒绝的替代方案

- 保留 `AppSearch`，只局部替换 UI：无法吃到官方 InstantSearch 的完整状态模型与 facet/filter 能力
- 使用自定义 searchClient 代理后端：更适合方案 B，但与当前已选方案不一致

## Rationale

- 当前索引层已经具备较好的 Meilisearch settings 基础（`internal/search/meili_index.go`），适合向官方 InstantSearch 方案收口
- 当前前端扩展名筛选存在计数与分页语义不准确的问题，改为真正的 facet/filter 后可一次性解决
- 公共搜索边界允许使用 search-only key，因此方案 A 在安全模型上成立
- 下载链路仍需服务端持有上游凭据，因此搜索与下载分离是最自然的边界

## Detailed Design

### 1. 公开搜索配置

新增一个轻量的公开搜索配置接口，由应用后端在页面初始化时返回：

- `host`
- `indexName`
- `searchApiKey`
- `instantsearchEnabled`

推荐放在现有 `AppService` 下，例如 `GetSearchConfig`。其职责仅限于：

- 向前端公开“允许被浏览器使用的只读搜索配置”
- 为后续 key 轮换、灰度、回滚保留控制点

这不是搜索代理，不承载搜索流量本身。

### 2. 搜索数据层切换

首页从：

- `useInfiniteQuery(appSearch)`
- 手工去重
- 手工 URL 参数同步
- 前端本地扩展名过滤

切换为：

- `InstantSearch` 作为唯一搜索状态源
- `instantMeiliSearch(...)` 创建 `searchClient`
- `routing` 负责 query / page / file_category 等 URL 同步
- `useInfiniteHits` 驱动无限滚动结果流
- `useSearchBox` 驱动输入框与立即搜索行为
- `useRefinementList` 或等价 hooks 驱动分类筛选 chips
- `Highlight` 或 `_formatted` 驱动名称高亮渲染

### 3. 保留现有 UI 壳层，替换内部状态来源

为了兼顾官方最佳实践与当前页面视觉系统：

- 保留 `SearchInput`、`FileCard`、空态、错误态和 hero/docked 结构
- 不再让这些组件读取本地搜索 state
- 改为由 InstantSearch hooks 提供 query、hits、stats、refinements 和 status

换句话说：

- **状态模型换成官方**
- **视觉组件尽量保留现有项目资产**

### 4. 索引字段与 settings 调整

当前索引已有 `name_ext`，但未用于公开 facet/filter。首批需要新增并暴露：

- `file_category`：`doc | image | video | archive | other`
- 将 `file_category` 加入 `FilterableAttributes`
- 视需要将 `name_ext` 也加入 `FilterableAttributes`
- 保持 `type=file`、`in_trash=false`、`is_deleted=false` 作为公开搜索默认过滤条件

同时建议收敛公开展示字段：

- 保留：`doc_id`、`source_id`、`type`、`name`、`name_ext`、`file_category`、`path_text`、`modified_at`、`created_at`、`size`
- 不在公开搜索结果中暴露与页面无关的内部字段

### 5. 搜索交互语义

首批交互目标：

- 保持 280ms 左右的输入 debounce 手感
- 仍支持 Enter 和按钮立即提交
- 结果列表使用 `InfiniteHits`
- URL 中持久化：query、page、file category
- 分类筛选采用真正的 facet/filter，而不是本地 `items.filter(...)`
- 结果总数、空态和状态文案全部以 InstantSearch 返回状态为准

### 6. 下载链路保持后端受控

搜索结果命中后：

- 点击下载按钮仍调用 `AppService.AppDownloadURL`
- 浏览器不直接从 Meilisearch 获得下载地址
- 搜索和下载形成双通道：
  - 搜索：浏览器 -> Meilisearch
  - 下载：浏览器 -> Connect -> 上游 Npan API

### 7. 灰度与回滚

首批上线必须保留双栈：

- `InstantSearch` 为默认候选新链路
- `AppSearch` 保留为旧链路和回滚开关

建议由 `GetSearchConfig.instantsearchEnabled` 或同等运行时开关控制前端选择哪条搜索链路，确保：

- 出现 CORS / 网络 / 结果语义问题时可立即切回旧实现
- E2E 可以同时覆盖新旧两条链路

## Success Criteria

- 打开搜索页后，直连链路开启时浏览器不再调用 `AppService.AppSearch`
- 浏览器只持有 search-only key，且页面和构建产物中不存在 admin/master key
- 分类筛选成为真正的 facet/filter，数量、列表和空态一致
- 搜索 URL 可刷新恢复，并支持回退/前进
- 下载按钮行为与当前页面一致
- 现有页面视觉壳层保留，前端测试与 E2E 完成迁移

## Risks and Mitigations

- 风险：浏览器直连后丢失当前后端 query 预处理与 fallback 体验
  - 缓解：把“结果差异对比”纳入验收；首批不做额外自定义，若确有明显回退，再单独设计 query adapter
- 风险：搜索流量绕过 Go 后端后，现有搜索指标失真
  - 缓解：补充前端埋点与 Meilisearch 侧监控，并保留下载/配置接口监控
- 风险：CORS、网络拓扑或 HTTPS 配置不完整导致浏览器不可达
  - 缓解：优先使用同源反向代理路径或完整预发验证，并保留 `AppSearch` 回滚
- 风险：如果仍沿用本地扩展名过滤，会与 InstantSearch refinement 双重冲突
  - 缓解：首批删除前端本地分类过滤逻辑，统一以 `file_category` refinement 为准

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
