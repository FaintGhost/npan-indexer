# Best Practices

## Scope and Simplicity

- 仅实现“前端筛选 + URL 持久化”，不扩展后端参数
- 保持最小改动：优先修改 `SearchPage` 与新增小型工具模块
- 不做多选组合和复杂筛选器，避免过度设计

## State Management

- 派生数据（`filteredItems`）应由 `items + activeFilter` 计算，不单独存储 state
- 仅在计算量可观时使用 `useMemo`
- 筛选切换通过事件处理更新 URL，避免额外 Effect 链式同步

## URL Handling

- 统一使用 `URLSearchParams` 或路由器提供的 search API
- 对 `ext` 做白名单校验，非法值回退到 `all`
- 默认值可省略参数，防止 URL 污染

## File Category Rules

- 分类应基于 `doc.name` 原始文件名，而非高亮 HTML 字段
- 支持大小写不敏感匹配
- 支持多段扩展名（如 `tar.gz`），优先最长后缀匹配
- 分类规则集中定义，禁止散落在多个组件中

## Accessibility

- 单选筛选器优先采用 `radiogroup/radio` 语义
- 需满足键盘操作：Tab 进入、方向键切换、Space/Enter 选择
- 组和选项都必须有可感知标签

## Security

- 不将 URL 参数直接注入 HTML
- 筛选逻辑不依赖 `dangerouslySetInnerHTML` 字段
- 对未知参数值严格兜底，避免异常分支泄露

## Performance

- 先去重再筛选，避免重复计算和数量不一致
- 保持分页加载守卫，避免筛选导致无意义高频请求
- 大结果集场景优先做 profiling，再决定是否引入进一步优化

## Testing Guidance

- 覆盖 URL 初始化、参数同步、非法值回退
- 覆盖扩展名分类边界（无后缀、未知后缀、多段后缀）
- 覆盖切换筛选后计数/空态/列表一致性
- 覆盖可访问性语义和基础键盘交互

## Decision Notes (ADR-style)

- ADR-01：筛选在前端渲染层执行，不进入后端请求契约
- ADR-02：筛选采用单选 + URL 参数持久化
- ADR-03：首版固定 6 类扩展名，后续按需求演进
