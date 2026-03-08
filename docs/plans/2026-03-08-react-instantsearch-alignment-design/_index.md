# React InstantSearch 纠偏设计

## Context

当前公开搜索页已经切到 React InstantSearch + Meilisearch 直连，但线上与手工验证暴露出两个高优先级问题：

- public 搜索当前是“提交后才搜”，不是官方示例的 search-as-you-type
- public 搜索结果质量明显差于 legacy `AppSearch`

根因已定位为两类偏差：

1. **交互行为偏差**
   - `web/src/routes/index.lazy.tsx` 中 public 分支的 `handleChange()` 只更新本地输入状态，不调用 `refine()`
   - `refine()` 只在 `handleSubmit()` 触发，因此用户连续输入时只会看到一次 `/multi-search`

2. **结果语义偏差**
   - legacy `AppSearch` 在 `internal/search/meili_index.go` 中包含 `preprocessQuery()`
   - legacy 还固定施加 `type=file`、`is_deleted=false`、`in_trash=false` 过滤
   - public 直连当前未补齐上述关键语义，因此召回与结果噪声都可能明显劣化

已确认本次目标档位为：

- **官方行为 + 关键对齐**
- 恢复官方 search-as-you-type 体验
- 补齐最关键的 legacy 搜索语义差异（query 预处理、默认过滤、结果对比与回滚门槛）
- 暂不扩大到排序规则或更深层的相关性调优

## Requirements

### Must

- public 搜索必须恢复为 **search-as-you-type**，输入变化应驱动 `/multi-search`
- 保留当前 280ms 左右的输入节奏，允许 debounce，但语义上不能退化成 submit-only
- Enter 与“搜索”按钮必须保留“立即触发当前查询”的能力
- public 搜索请求必须补齐 legacy 最关键的 query 预处理语义
- public 搜索必须始终施加与 legacy 公开搜索一致的默认过滤：
  - `type = file`
  - `is_deleted = false`
  - `in_trash = false`
- query、page、file_category 仍由 InstantSearch routing 统一拥有
- 必须建立 public vs legacy 的结果对比与发布/回滚门槛
- 必须保留 `AppSearch` fallback 与 `instantsearchEnabled` 运行时开关

### Should

- query 适配层只改写发往 Meilisearch 的 query，不改写输入框展示值与 URL 中原始 query
- 尽量继续沿用 `react-instantsearch` + `@meilisearch/instant-meilisearch` 官方栈，不回退为自定义搜索状态机
- 默认过滤应通过 InstantSearch 官方配置方式统一注入，而不是在结果层后过滤
- 结果对比样本应至少覆盖：普通关键词、扩展名关键词、版本号关键词、多词组合查询

### Won't

- 本轮不新增排序 UI 或重排排序规则
- 本轮不系统性调优 Meilisearch ranking rules / synonyms / typo / searchableAttributes
- 本轮不承诺完整复刻 legacy 的 `All -> Last` fallback 机制
- 本轮不删除 legacy `AppSearch` 双栈
- 本轮不改下载链路、管理后台、鉴权边界

## Option Analysis

### Option A（推荐）: 保持 InstantSearch 架构，只纠正输入语义与关键搜索语义

在现有 public bootstrap、routing、refinement、InfiniteHits 与下载受控链路基础上，最小补 4 项：

1. 输入即搜（search-as-you-type）
2. query 预处理 adapter
3. 默认过滤基线
4. 结果对比与回滚门槛

**优点**：
- 保持现有 React InstantSearch 架构与测试资产
- 改动聚焦，风险最小
- 能直接回应当前两个线上核心问题

**代价**：
- 需要在前端补一层与 Go 语义对齐的 query adapter
- 需要新增对比测试与发布门槛定义

### Option B: 回退到 legacy `AppSearch`，暂缓 public 直连

**优点**：
- 风险最低，用户体验立即回到旧行为

**缺点**：
- 放弃当前 public InstantSearch 路线
- 无法收敛既有投资
- 无法解决“官方行为与业务语义如何兼容”的设计问题

### Option C: 继续坚持纯官方默认模式，不补 legacy 语义

**不推荐**。

虽然更“纯”，但与当前线上反馈冲突。已有证据表明默认模式已造成明显业务回退，继续坚持只会放大问题。

## Rationale

选择 Option A 的原因：

- 当前 public 搜索的主要问题并非架构选型错误，而是**官方交互行为没有真正落地**，以及**关键业务语义未对齐**
- public bootstrap、routing、facet、下载边界已经有可复用资产，不应推倒重来
- 用户明确要求是“官方行为 + 关键对齐”，不是继续拿“首批不做 preprocess”当成范围借口
- 保留 `instantsearchEnabled` 作为灰度与回滚开关，可以控制实施风险

## Detailed Design

### 1. 恢复 search-as-you-type，消除 submit-only 偏差

public 搜索输入应恢复为官方语义：

- 输入变化后，经过约 280ms debounce，自动触发 `refine()`
- Enter / 搜索按钮可以立即 flush 当前查询
- 清空输入时，应清空当前 query 并恢复初始态

关键要求：

- 不能重新引入 public 模式下的“本地输入状态拥有权 > InstantSearch 状态拥有权”
- 输入框展示值、URL 状态、InstantSearch query 必须保持单一真相源

### 2. 引入最小 query adapter，对齐 legacy `preprocessQuery()` 关键语义

本轮只补最关键的 query 预处理能力，用于缩小 public 与 legacy 的召回差异。

最小对齐范围：

- `word.ext` 拆分与扩展名前置
- 版本号 `V/v` 前缀去除
- 扩展名词前置

约束：

- 只影响发往 Meilisearch 的 query
- 不改写输入框中展示的原始文本
- 不改写 URL 中持久化的 query
- adapter 逻辑需有独立测试，避免与组件状态耦合

### 3. 统一施加 public 默认过滤基线

public 搜索必须始终带上公开搜索默认过滤：

- `type = file`
- `is_deleted = false`
- `in_trash = false`

这组过滤是系统边界，不应依赖用户操作，也不应由结果列表渲染层去二次裁剪。

`file_category` refinement 仍保留，但它应叠加在上述默认过滤之上。

### 4. 保持现有 InstantSearch routing / refinement / hits 壳层

本轮不推翻以下已落地能力：

- `GetSearchConfig` public bootstrap
- `InstantSearch` 作为搜索状态拥有者
- `file_category` refinement 与 URL routing
- `useInfiniteHits` 结果渲染
- `AppDownloadURL` 下载边界

也就是说：

- **状态模型继续保持官方栈**
- **只纠正触发方式与关键搜索语义**

### 5. 将结果对比与回滚门槛升级为交付物

本轮必须产出一套可执行的对比与门槛定义，而不是只在 review 里口头记录。

最低要求：

- 对比 public 与 legacy 的命中总数
- 对比前 10 条结果
- 对比高亮输出
- 对比 preprocess 敏感查询（扩展名、版本号、多词组合）
- 对比默认过滤泄漏（folder / deleted / trash）

阻塞级差异至少包括：

- public 泄漏 folder / deleted / trash 文档
- legacy 有命中、public 因缺少 preprocess 而空结果
- public 仍然不是 search-as-you-type

若出现阻塞级差异，默认发布不得开启 public 搜索，必须保留或切回 legacy。

## Success Criteria

- 输入关键字时，public 模式下可在 network 中看到按输入变化触发的 `/multi-search`
- Enter / 搜索按钮仍可立即触发当前查询
- preprocess 敏感查询的 public 结果与 legacy 差距显著缩小
- public 默认不返回 folder / deleted / trash 结果
- `file_category` refinement 继续可用，且叠加在默认过滤之上
- 已建立 public vs legacy 的结果对比资产与发布/回滚门槛
- `instantsearchEnabled` 关闭后仍能稳定回退到 legacy `AppSearch`

## Risks and Mitigations

- 风险：输入即搜会增加请求量
  - 缓解：保留约 280ms debounce，并验证输入行为不退化为 submit-only
- 风险：前端 query adapter 与 Go `preprocessQuery()` 漂移
  - 缓解：把 adapter 独立成单文件并建立样例对齐测试
- 风险：默认过滤与 refinement 叠加后造成“看起来没结果”
  - 缓解：在测试和发布门槛里单独覆盖默认过滤 + file_category 组合场景
- 风险：仍存在未覆盖的 deeper relevance 差异
  - 缓解：本轮明确不做深度相关性调优，但要求结果对比能把剩余差异显式分类

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Best practices and considerations
