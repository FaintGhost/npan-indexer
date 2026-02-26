# Search Results Filter Design

## Context

当前搜索页基于 Connect `AppSearch` 获取结果，并在前端按页合并去重后渲染。

- 入口页面：`web/src/routes/index.lazy.tsx`
- 请求参数目前仅包含 `query/page/pageSize`，未支持筛选参数
- 结果展示当前直接使用 `items`，状态文案与计数也基于同一数据源

已确认本次目标是：

- 仅做前端筛选，不改后端契约
- 筛选维度固定 6 类：`全部/文档/图片/视频/压缩包/其他`
- 交互模式选择方案 B：单选 + URL 参数持久化

## Requirements

### Must

- 在搜索结果页新增扩展名单选筛选控件（6 类固定枚举）
- 筛选状态与 URL 参数双向同步，支持刷新恢复、可分享链接、前进后退
- 不修改 `proto`、后端服务、Connect 请求体结构
- 筛选逻辑作用于“前端已加载并去重后的结果”
- 筛选后结果计数、空态、状态文案与展示列表保持一致

### Should

- URL 参数非法时回退到 `all`
- “全部”默认不强制写 URL 参数，保持链接简洁
- 扩展名分类规则集中管理并可单测

## Option Analysis

### Option A (Recommended): 路由原生 search params

使用 TanStack Router 的 search 参数能力维护 `ext`（读取 + 更新）。

- 优点：与现有路由栈一致，历史导航语义自然，类型边界清晰
- 缺点：需要补充/调整搜索页单测的路由上下文

### Option B: 手工 URLSearchParams + history API

在组件内手动读写 URL。

- 优点：表面改动少
- 缺点：易产生边界不一致（回退、合并参数、非法值兜底）

### Option C: 自定义 URL 状态 Hook

封装 B 的细节形成复用 Hook。

- 优点：复用性更好
- 缺点：在本场景属于提前抽象，YAGNI 风险高

**结论**：采用 Option A，在不改后端前提下达到最稳妥的一致性。

## Rationale

- 后端 `AppSearch` 当前固定面向文件检索，本次不做契约扩展可显著降低回归风险
- 前端已有去重聚合路径，筛选应接在去重后以保证计数与展示一致
- URL 持久化能满足“刷新不丢状态”和“复制链接复现视图”的核心体验

## Detailed Design

### 1. URL 参数模型

- 参数名：`ext`
- 允许值：`all | doc | image | video | archive | other`
- 兼容策略：
  - 缺失或非法值 => `all`
  - 选择 `all` 时可删除 `ext` 参数

### 2. 前端分类规则

建议新增集中规则模块（例如 `web/src/lib/file-category.ts`）：

- `doc`: `pdf,doc,docx,xls,xlsx,ppt,pptx,txt,md,markdown,rtf,odt,ods,odp,csv,tsv,epub`
- `image`: `jpg,jpeg,png,gif,webp,bmp,svg,tif,tiff,heic,heif,avif,ico`
- `video`: `mp4,mkv,mov,avi,wmv,flv,webm,m4v,mpg,mpeg,ts,rmvb`
- `archive`: `zip,rar,7z,tar,gz,tgz,bz2,xz,zst,tar.gz,tar.bz2,tar.xz`
- `other`: 未命中上述分类，或无扩展名

### 3. 过滤流水线

在 `SearchPage` 形成单向数据流：

1) Connect 返回分页数据
2) `mergePages` 去重得到 `items`
3) 基于 `items + activeFilter` 计算 `filteredItems`
4) 用 `filteredItems` 驱动列表、空态、计数和状态文案

约束：不引入“过滤后结果副本 state”，使用派生计算（必要时 `useMemo`）。

### 4. 交互与可访问性

- UI 形态：单选分段控件（6 项）
- 语义建议：`radiogroup/radio`（更贴合“互斥筛选”）
- 键盘要求：Tab 进入，方向键切换，Space/Enter 选择
- 屏幕阅读器：组和选项都提供可感知名称

### 5. 测试与验证范围

- 单测：分类规则、URL 参数同步、筛选行为、非法值兜底、清空搜索联动
- 可访问性测试：筛选控件语义和键盘行为
- 回归：确认 Connect 请求体不受筛选切换影响

## Success Criteria

- 用户可通过 URL 直接打开任意筛选视图
- 切换筛选不触发新的后端契约改动或协议路径变化
- 筛选后的数量、列表、空态与状态文案一致
- 单测覆盖核心行为并通过

## Risks and Mitigations

- 风险：筛选后列表较短，可能更频繁触发无限滚动 sentinel
  - 缓解：保持现有 `hasNextPage` 守卫，必要时增加“最小展示高度”策略
- 风险：扩展名规则分散导致行为漂移
  - 缓解：集中映射表 + 独立单测
- 风险：自定义控件可访问性不完整
  - 缓解：优先原生 radio 语义，再做样式增强

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
