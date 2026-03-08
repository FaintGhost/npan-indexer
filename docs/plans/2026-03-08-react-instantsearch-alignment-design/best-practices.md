# Best Practices: React InstantSearch 纠偏

## 1. 先恢复官方输入语义，再补业务语义

- search-as-you-type 是当前 public 搜索的第一优先级行为修正
- 不要在 submit-only 语义上继续叠加更多兼容逻辑
- 先让输入变化稳定驱动请求，再评估结果质量对齐

## 2. 保持 InstantSearch 为唯一搜索状态源

- query、page、refinements 的真相源继续是 InstantSearch
- 不要重新引入 public 模式下的双状态模型（例如“本地输入值拥有权 > InstantSearch query”）
- 输入框如果需要本地节奏控制，只能作为 debounce 辅助层，不能重新拥有搜索结果语义

## 3. query adapter 要独立、可测、最小化

- 将 legacy preprocess 对齐逻辑放到独立 utility 文件
- adapter 只负责 outbound query 改写，不负责 UI、URL 或组件状态同步
- 使用独立单测覆盖代表性样例：
  - `规格书.pdf`
  - `firmware v3.2.1`
  - `mx40 spec pdf`
- 禁止把 adapter 逻辑散落在组件事件处理函数中

## 4. 默认过滤必须在搜索请求层生效

- `type=file`、`is_deleted=false`、`in_trash=false` 属于系统默认过滤
- 不能依赖结果展示层再二次过滤，否则会导致：
  - 结果总数错误
  - facet 计数错误
  - 空态错误
  - 分页错误
- 用户态 refinement（`file_category`）只能叠加在默认过滤之上

## 5. 保持官方栈，不回退到自定义搜索状态机

- 继续使用 `react-instantsearch` + `@meilisearch/instant-meilisearch`
- 尽量通过官方配置入口和 hooks 完成纠偏
- 不要因为要补 preprocess / 默认过滤，就把搜索重新退回手写 `useEffect + fetch` 模式

## 6. Enter / 按钮是“立即触发”，不是“唯一触发”

- 回车与按钮保留当前产品习惯
- 但它们在语义上是 debounce 的 flush / accelerate，不应替代输入即搜主路径
- 测试里要分别覆盖：
  - 输入停顿触发
  - 输入后立即 Enter 触发
  - 输入后立即点按钮触发

## 7. 结果对比必须有明确样本与阻塞级定义

- 不要用“感觉差不多”作为发布依据
- 至少覆盖这些对比维度：
  - 命中总数
  - 前 10 条结果
  - 高亮输出
  - 空态 / 错误态
  - 默认过滤泄漏
  - preprocess 敏感查询
- 阻塞级差异至少包括：
  - folder / deleted / trash 泄漏
  - legacy 有结果、public 无结果的关键查询
  - public 仍不是 search-as-you-type

## 8. 回滚条件要前置定义

- 继续保留 `instantsearchEnabled` 作为发布开关
- 若出现阻塞级差异，默认启用必须停止
- 验证不仅要证明“开关开能用”，还要证明“关掉后能稳定回退 legacy”

## 9. 测试策略

- 单测优先覆盖：
  - 输入即搜触发频次
  - Enter / 按钮立即触发
  - query adapter 输入输出对齐
  - 默认过滤注入
  - refinement 与默认过滤叠加
- E2E 重点覆盖：
  - 浏览器输入时出现多次 `/multi-search`
  - public 不再调用 `AppSearch`
  - legacy fallback 关闭开关后仍可搜索和下载
- 所有网络场景使用测试替身或 route mock 隔离外部依赖

## 10. 文档要同步修正旧设计表述

- 现有旧设计中“首批不复刻 preprocessQuery”的约束已不再适用本轮目标
- 新 design / plan / review 必须统一更新为“官方行为 + 关键对齐”
- 避免执行时同时存在两套相互冲突的范围定义
