# BDD Specifications: React InstantSearch 直连 Meilisearch

## Feature: 公开搜索配置初始化

### Scenario 1: 页面启动时成功获取公开搜索配置
```gherkin
Given 后端返回公开搜索配置 host、indexName、searchApiKey 和 instantsearchEnabled=true
When 用户打开搜索页
Then 前端应初始化 InstantSearch search client
And 浏览器不应调用 AppService.AppSearch
```

### Scenario 2: 公开搜索配置不可用时回退旧链路
```gherkin
Given 后端返回 instantsearchEnabled=false 或公开搜索配置缺失
When 用户打开搜索页
Then 前端应回退到现有 AppSearch 链路
And 页面仍可完成搜索与下载
```

## Feature: 搜索交互由 InstantSearch 驱动

### Scenario 3: 输入关键字后触发直连 Meilisearch 搜索
```gherkin
Given 搜索页已成功初始化 InstantSearch
When 用户输入 "report" 并提交搜索
Then 浏览器应向 Meilisearch 发起搜索请求
And 结果列表应展示 hits
And 状态文案应基于 InstantSearch 返回的结果数量更新
```

### Scenario 4: InfiniteHits 驱动无限滚动
```gherkin
Given 当前查询已有第一页结果且存在下一页
When 用户滚动到结果列表底部
Then 前端应通过 InfiniteHits 继续加载下一批 hits
And 已加载结果应继续可见
```

## Feature: facet/filter 取代前端本地扩展名过滤

### Scenario 5: 文件分类筛选使用 file_category refinement
```gherkin
Given 索引文档包含 file_category 字段并配置为 filterable
When 用户选择 "文档" 分类筛选
Then 搜索请求应携带对应 refinement
And 结果总数应与筛选后的命中数一致
And 页面不应再使用本地 items.filter 进行分类裁剪
```

### Scenario 6: 非法 URL 筛选值应回退默认分类
```gherkin
Given 用户直接访问带有非法筛选参数的搜索 URL
When 搜索页初始化 routing 状态
Then 当前 refinement 应回退到默认分类
And 页面不应抛出异常
```

## Feature: 搜索状态与 URL 路由同步

### Scenario 7: query、page 和分类筛选可从 URL 恢复
```gherkin
Given 用户已经在搜索页产生 query、page 和 file_category 状态
When 用户刷新页面或通过分享链接重新打开
Then 搜索页应从 URL 恢复相同的 InstantSearch 状态
And 用户无需再次手动输入
```

### Scenario 8: 浏览器前进后退可恢复搜索视图
```gherkin
Given 用户依次切换了不同 query 或分类筛选
When 用户使用浏览器后退或前进
Then 搜索页应恢复到对应的 InstantSearch routing 状态
And 结果列表应与 URL 保持一致
```

## Feature: 高亮与结果渲染

### Scenario 9: 命中结果名称显示高亮
```gherkin
Given Meilisearch 返回带有 _formatted.name 的 hits
When 结果卡片渲染文件名称
Then 页面应展示高亮后的名称
And 未命中高亮的结果应展示原始名称
```

## Feature: 下载链路保持服务端受控

### Scenario 10: 搜索结果下载仍通过 AppDownloadURL
```gherkin
Given 用户在搜索结果中点击下载按钮
When 前端发起下载动作
Then 前端应调用 AppService.AppDownloadURL
And 不应尝试从 Meilisearch 响应中直接获取下载地址
```

## Feature: 安全边界

### Scenario 11: 浏览器只暴露 search-only key
```gherkin
Given 搜索页运行在浏览器环境中
When 页面初始化公开搜索客户端
Then 使用的凭证必须是 search-only key
And 页面中不得出现 Meilisearch admin 或 master key
```

## Feature: 回滚与兼容

### Scenario 12: 直连链路异常时可切回 AppSearch
```gherkin
Given 预发或线上发现直连 Meilisearch 存在不可接受的问题
When 运维关闭 instantsearchEnabled 开关
Then 前端应切回 AppSearch 搜索链路
And 下载与页面其他功能保持可用
```
