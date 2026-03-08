# Architecture: React InstantSearch 直连 Meilisearch

## Component Diagram

```text
┌──────────────────────────────────────────────────────────┐
│                      Browser / React                     │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │ Search Page                                        │  │
│  │                                                    │  │
│  │  ┌──────────────────────────────────────────────┐  │  │
│  │  │ InstantSearch Provider                       │  │  │
│  │  │ - searchClient from instant-meilisearch      │  │  │
│  │  │ - routing                                     │  │  │
│  │  │ - useSearchBox / useInfiniteHits             │  │  │
│  │  │ - useRefinementList / Highlight              │  │  │
│  │  └───────────────────────┬──────────────────────┘  │  │
│  │                          │                         │  │
│  │         search           │          download       │  │
│  └──────────────────────────┼─────────────────────────┘  │
└─────────────────────────────┼────────────────────────────┘
                              │
             ┌────────────────▼────────────────┐
             │      AppService.GetSearchConfig │
             │      AppService.AppDownloadURL  │
             └────────────────┬────────────────┘
                              │
          ┌───────────────────▼───────────────────┐
          │              Meilisearch              │
          │  public index + search-only key       │
          │  filterable/sortable/displayed attrs  │
          └───────────────────┬───────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │ Go Sync / Indexer │
                    │ builds documents  │
                    │ maintains settings│
                    └───────────────────┘
```

## Key Architectural Decisions

### 1. 由 InstantSearch 统一拥有搜索状态

当前 `web/src/routes/index.lazy.tsx` 里同时维护：

- `query`
- `activeQuery`
- 手工分页合并
- 手工 URL 参数同步
- 本地扩展名筛选

迁移后，统一改由 `InstantSearch` 持有 query、page、refinements 与 routing。这样可以：

- 消除重复状态源
- 直接获得官方 routing 和 InfiniteHits 能力
- 避免“本地筛选与服务端总数不一致”的语义错位

### 2. 通过运行时配置接口下发公开搜索配置

虽然公共搜索可以合法使用 search-only key，但本设计仍引入轻量配置接口，而不是把配置硬编码进前端：

- 更利于 key 轮换
- 更利于灰度开关
- 更利于回滚到 `AppSearch`
- 更利于未来演进为 tenant token

该接口不是搜索代理，不承担搜索流量。

### 3. 使用官方 hooks 复用现有页面壳层

`react-instantsearch` 支持 widgets 与 hooks。对于当前项目，推荐：

- 用 hooks 获取 search state 和 hits
- 继续复用现有 `SearchInput`、`FileCard`、空态和错误态
- 不将页面整体替换成默认 Algolia 风格 widgets DOM

这既符合官方栈，也能最大限度保留当前搜索页设计资产。

### 4. 新增 file_category，收敛公开搜索 facet 语义

当前 `web/src/lib/file-category.ts` 的分类逻辑只存在于前端。为了让 InstantSearch refinement 成为真正的服务端筛选，需要：

- 在索引文档中新增 `file_category`
- 在 `internal/search/mapper.go` 中映射文件分类
- 在 `internal/search/meili_index.go` 中将 `file_category` 加入 `FilterableAttributes`
- 使用 `file_category` 替代本地 `items.filter(...)`

### 5. 公开搜索结果字段最小化

公开搜索页只需要：

- 基本标识与下载所需主键
- 名称与高亮渲染字段
- 修改时间、创建时间、大小
- 分类与路径文本

因此首批不应让浏览器拿到不必要的内部字段，公开搜索 settings 需要同步收敛 `DisplayedAttributes`。

### 6. 首批不复刻旧后端 query preprocess

旧链路在 `internal/search/meili_index.go` 中还包含：

- `preprocessQuery()`
- `MatchingStrategy` 的 `All -> Last` fallback

这些行为会影响结果排序与召回，但并不是官方 InstantSearch 默认模式。为了严格遵循官方最佳实践，首批设计选择：

- 先按官方直连模式落地
- 通过结果对比验证其是否足够
- 仅在验证发现明显业务回退时，再单独设计 adapter 层

### 7. 保留 `AppSearch` 作为灰度与回滚兜底

本设计不删除 `AppSearch`，理由：

- 浏览器直连的 CORS / 网络 / 代理配置风险较高
- 现有 E2E 和监控需要过渡
- 结果语义需要有对照基线

因此首批应该保留双栈，并通过运行时开关决定实际使用哪条链路。

## File Changes Summary

### Modified Files

| File | Change |
|------|--------|
| `proto/npan/v1/api.proto` | 新增公开搜索配置 RPC 与响应消息 |
| `internal/httpx/connect_app_auth_search.go` | 实现公开搜索配置 handler，保留 `AppDownloadURL` 与 `AppSearch` |
| `internal/models/models.go` | 为索引文档增加 `file_category` |
| `internal/search/mapper.go` | 在索引映射阶段写入 `file_category` |
| `internal/search/meili_index.go` | 更新公开搜索相关 settings（filterable/displayed attributes） |
| `web/package.json` | 引入 `react-instantsearch`、`@meilisearch/instant-meilisearch`、`instantsearch.css` |
| `web/src/routes/index.lazy.tsx` | 切换为 InstantSearch provider + hooks 驱动搜索页 |
| `web/src/components/search-input.tsx` | 适配 `useSearchBox` 驱动的输入与提交行为 |
| `web/src/components/file-card.tsx` | 适配 InstantSearch hit / 高亮渲染 |
| `web/e2e/tests/search.spec.ts` | 从等待 Connect `AppSearch` 改为等待 Meilisearch 直连请求 |
| `web/e2e/fixtures/seed.ts` | 为测试索引补充 `file_category` 与 refinement 所需 settings |

### New Files

| File | Purpose |
|------|---------|
| `web/src/lib/search-config.ts` | 拉取公开搜索配置并做运行时校验 |
| `web/src/lib/meili-search-client.ts` | 初始化 `instantMeiliSearch` 并导出单例 search client |
| `web/src/lib/meili-hit-adapter.ts` | 将 Meilisearch hit 映射为现有 UI 组件所需结构 |
| `web/src/components/search-filters.tsx` | 基于 InstantSearch hooks 的分类筛选 UI |
| `web/src/components/search-results.tsx` | 基于 `useInfiniteHits` 的结果列表封装 |

### Unchanged Files

| File | Reason |
|------|--------|
| `web/src/hooks/use-download.ts` | 下载仍通过 Connect `AppDownloadURL`，职责不变 |
| `web/src/components/admin-sync-page.tsx` | 管理后台与本次公开搜索迁移无关 |
| `internal/service/*` | 同步编排逻辑无需感知前端搜索方式变化 |

## Migration Phases

### Phase 1: 安全与配置准备

- 新增公开搜索配置接口
- 准备 search-only key
- 验证浏览器到 Meilisearch 的网络可达性和 CORS
- 引入运行时开关，允许前端在 `instantsearch` / `appsearch` 间切换

### Phase 2: 索引字段与 settings 收敛

- 新增 `file_category`
- 更新 `FilterableAttributes`
- 收敛 `DisplayedAttributes`
- 为测试种子数据补 facet 字段

### Phase 3: 前端搜索页切换

- 引入 `react-instantsearch` / `instant-meilisearch`
- 使用 hooks 替换现有搜索状态机
- 保留下载按钮和视觉壳层
- 删除本地扩展名过滤逻辑

### Phase 4: 验证与灰度

- 前端单测迁移
- E2E 改为等待 Meilisearch 请求
- 结果对比：新旧链路命中数、前 10 结果、高亮、空态
- 预发灰度后再决定是否默认开启
