# Best Practices: React InstantSearch 直连 Meilisearch

## 1. 依赖与初始化

- 严格使用官方推荐组合：
  - `react-instantsearch`
  - `@meilisearch/instant-meilisearch`
  - `instantsearch.css`
- `searchClient` 在模块级初始化或通过稳定工厂创建，避免每次渲染重复实例化
- `InstantSearch` 只挂载一次，避免上层状态变化导致搜索状态重置

## 2. 安全边界

- 浏览器环境只能持有 **search-only key**
- 严禁将 Meilisearch admin/master key 打包到前端
- 公共搜索可使用 search-only key；如果未来变成按用户或目录隔离，必须升级为 tenant token
- 优先通过运行时配置接口返回公开搜索配置，而不是把敏感配置常量直接固化到 Vite 构建产物

## 3. 路由与状态所有权

- query、page、refinements 的唯一真相源应为 InstantSearch routing
- 禁止同时保留一套手工 `URLSearchParams + history.replaceState` 逻辑
- 旧页面中的 `query/activeQuery/activeFilter` 本地双状态应被清理，避免双向同步竞态

## 4. Facet / Filter 设计

- 公开搜索的分类筛选必须基于索引字段，而不是基于前端已加载 hits 二次过滤
- `file_category` 适合作为固定枚举 refinement 字段
- 只把真正需要公开筛选的字段加入 `FilterableAttributes`
- refinement UI 的计数、空态和 URL 状态都以 InstantSearch 返回结果为准

## 5. Displayed / Searchable / Sortable Attributes

- `DisplayedAttributes` 只保留公开搜索页渲染必需字段，避免把无关元数据暴露到浏览器
- `SearchableAttributes` 保持聚焦，优先 `name_base`、`name_ext`、`name`、`path_text`
- `SortableAttributes` 只保留真实需要的字段；如果首批不暴露排序 UI，不必额外新增复杂排序产品逻辑
- 如需后续接入排序控件，先复用现有 `modified_at`、`size`、`created_at` 能力

## 6. 高亮与结果渲染

- 名称高亮优先使用 InstantSearch 官方高亮能力，而不是继续扩散手工 HTML 处理逻辑
- 只对真正需要展示的文本字段开启高亮
- 如果继续使用现有 `FileCard`，应将高亮映射限定在名称字段，避免把 URL 参数或不可信文本直接注入 HTML

## 7. 保留视觉资产，替换数据层

- 现有页面的视觉结构、按钮、卡片、空态可以保留
- 推荐使用 InstantSearch hooks 接入现有组件，而不是整页换成默认 widgets DOM
- 这样既符合官方栈，也能最大限度减少视觉返工

## 8. 性能与体验

- 首批继续保持小而稳定的每页加载量，避免过大 `hitsPerPage`
- `InfiniteHits` 是首选无限滚动方式，但必须配合克制的分页上限
- 不要为追求“滚得更深”而盲目放大 `maxTotalHits`
- 搜索体验优先使用官方默认状态模型，只有在验证确认明显回退时，才为 query preprocessing 增加适配层

## 9. 测试策略

- 单测覆盖：
  - 公开搜索配置加载
  - search-only key 安全边界
  - routing 状态恢复
  - `file_category` refinement
  - 结果高亮
  - 下载链路未回归
- E2E 需要从“等待 `/npan.v1.AppService/AppSearch`”切换到“等待浏览器直连 Meilisearch 请求”
- 必做新旧链路结果对比，至少覆盖：
  - 命中总数
  - 前 10 条结果
  - 高亮输出
  - 空态与错误态

## 10. 可观测性

- 搜索流量绕过 Go 后端后，现有后端搜索缓存/命中指标不再代表真实首页搜索流量
- 必须补至少一层新的观测：
  - 前端搜索埋点
  - Meilisearch 侧查询延迟/错误率
  - 公开搜索配置接口的调用成功率
- 下载链路仍保留后端监控，不受本次迁移影响

## 11. 灰度与回滚

- 首批必须保留 `AppSearch` 旧链路
- 通过运行时开关决定是走 InstantSearch 还是 AppSearch
- 不在首批删除旧测试与旧 handler，直到预发和线上灰度稳定

## 12. ADR-style 决策记录

- ADR-01：公共搜索采用 React InstantSearch + `instant-meilisearch` 官方栈
- ADR-02：浏览器只持有 search-only key，下载仍保留 Connect 后端受控链路
- ADR-03：分类筛选升级为索引字段 `file_category`，不再保留前端本地扩展名过滤
- ADR-04：首批保留 `AppSearch` 作为灰度与回滚兜底
- ADR-05：首批以官方默认直连模式为准，不提前复刻旧后端 query preprocess 行为
