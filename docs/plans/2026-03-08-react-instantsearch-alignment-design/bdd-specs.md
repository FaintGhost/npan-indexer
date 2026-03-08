# BDD Specifications: React InstantSearch 纠偏

## Feature: 输入即搜恢复为官方行为

### Scenario 1: 输入时自动触发 public 搜索
```gherkin
Given 搜索页已成功初始化 InstantSearch 且启用 public 搜索
When 用户输入 "report" 并停止输入约 280ms
Then 浏览器应在未点击搜索按钮且未按 Enter 的情况下向 Meilisearch 发起搜索请求
And 结果列表应展示 hits
And 状态文案应基于 InstantSearch 返回的结果数量更新
And URL 中的 query 应与当前搜索状态一致
```

### Scenario 2: Enter 或搜索按钮可立即触发当前查询
```gherkin
Given 搜索页已成功初始化 InstantSearch 且启用 public 搜索
When 用户输入 "report" 后立即按 Enter 或点击搜索按钮
Then 浏览器应立即向 Meilisearch 发起当前查询请求
And 不必等待输入 debounce 到期
```

## Feature: query 预处理与 legacy 关键语义对齐

### Scenario 3: public 搜索应对 query 应用最小 legacy 预处理
```gherkin
Given 搜索页已成功初始化 InstantSearch 且启用 public 搜索
When 用户输入带扩展名、版本号或多词组合的查询并触发搜索
Then 发往 Meilisearch 的 query 应与 legacy preprocess 规则等价
And 搜索框展示值仍应保留用户原始输入
And URL 中的 query 仍应保留用户原始输入
```

## Feature: 默认过滤基线与 refinement 叠加

### Scenario 4: public 搜索始终带公开默认过滤
```gherkin
Given 索引中同时存在 file、folder、in_trash=true 和 is_deleted=true 的匹配文档
When 用户搜索 "report"
Then 发往 Meilisearch 的请求应始终包含 type=file、in_trash=false 和 is_deleted=false
And 返回结果中不应出现 folder、回收站或已删除文档
And 结果总数应基于默认过滤后的结果
```

### Scenario 5: file_category refinement 应叠加在默认过滤之上
```gherkin
Given public 搜索默认过滤已经生效
And 索引文档包含 file_category 字段并配置为 filterable
When 用户选择 "文档" 分类筛选
Then 搜索请求应同时携带默认过滤和 file_category refinement
And 结果总数应与筛选后的命中数一致
And 页面不应在结果渲染层再次做本地分类裁剪
```

## Feature: 新旧链路结果对比与发布闸门

### Scenario 6: public 与 legacy 结果对比形成可发布结论
```gherkin
Given 已准备代表性查询集，覆盖普通查询、扩展名查询、版本号查询和多词组合查询
When 分别执行 public 搜索与 legacy AppSearch
Then 应记录每个查询的命中总数、前 10 条结果、高亮输出、空态与错误态
And 应将差异标记为可接受、待解释或阻塞级差异
```

### Scenario 7: 阻塞级差异会阻止默认开启 public 搜索
```gherkin
Given 已完成 public 与 legacy 的结果对比
When 存在 folder 或 deleted 或 trash 泄漏
Or 存在关键 preprocess 查询 legacy 有结果而 public 无结果
Or public 仍然不是 search-as-you-type
Then 本次发布不得默认开启 public 搜索
And 必须保留 legacy fallback 作为主链路或回退链路
```

### Scenario 8: 达到阻塞级差异或故障门槛时可立即回滚到 AppSearch
```gherkin
Given public 搜索已在预发或线上灰度开启
And 已出现阻塞级差异或不可接受故障
When 运维关闭 instantsearchEnabled 开关
Then 前端应切回 AppSearch 搜索链路
And 下载与页面其他功能保持可用
And 用户无需变更访问入口
```
