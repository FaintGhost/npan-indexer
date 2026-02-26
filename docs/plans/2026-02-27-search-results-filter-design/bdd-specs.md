# BDD Specifications

## Feature: Search Results Extension Filter With URL Persistence

### Scenario: 默认进入页面时使用全部筛选

```gherkin
Given 用户访问搜索页且 URL 不包含 ext 参数
When 页面完成初始化
Then 筛选值应为 all
And 结果列表显示当前已加载的全部结果
```

### Scenario: URL 中 ext 合法值可恢复筛选状态

```gherkin
Given 用户访问 /?q=mx40&ext=doc
When 搜索结果加载完成
Then 文档筛选应为选中状态
And 仅展示文档类扩展名结果
```

### Scenario: URL 中 ext 非法值会回退到 all

```gherkin
Given 用户访问 /?q=mx40&ext=unknown
When 页面完成初始化
Then 筛选值应回退为 all
And 页面不应报错
```

### Scenario: 用户切换筛选会同步更新 URL

```gherkin
Given 当前 URL 为 /?q=test
And 当前筛选为 all
When 用户点击图片筛选
Then URL 应更新为包含 ext=image
And 列表应仅展示图片类结果
```

### Scenario: 选择 all 时可移除 ext 参数

```gherkin
Given 当前 URL 为 /?q=test&ext=video
When 用户切换到全部筛选
Then URL 中 ext 参数应被移除或归一化为默认
And 列表恢复展示全部结果
```

### Scenario: 切换筛选不改变后端请求契约

```gherkin
Given 用户已发起搜索并拿到结果
When 用户在文档和压缩包筛选间切换
Then 不应引入新的后端查询字段
And Connect 请求路径仍为 /npan.v1.AppService/AppSearch
```

### Scenario: 去重后再筛选保持数量一致性

```gherkin
Given 两页结果中存在相同 source_id 的重复项
When 页面合并分页并应用任一筛选
Then 重复项应只显示一次
And 计数文案应与实际显示条目一致
```

### Scenario: 筛选为空时展示可理解的空态

```gherkin
Given 搜索结果总量大于 0
And 当前筛选下没有匹配条目
When 页面渲染结果区
Then 应展示筛选后空态提示
And 不应展示错误态
```

### Scenario: 清空搜索后回到初始视图

```gherkin
Given 用户已有 query 与 ext 筛选状态
When 用户点击清空搜索
Then 输入框应清空
And 结果区应回到初始状态
And URL 中搜索相关参数应被清理或归一化
```

### Scenario: 分类规则正确识别常见扩展名

```gherkin
Given 文件名分别为 report.pdf、photo.jpg、demo.mp4、backup.tar.gz、README
When 系统执行扩展名分类
Then report.pdf 属于 doc
And photo.jpg 属于 image
And demo.mp4 属于 video
And backup.tar.gz 属于 archive
And README 属于 other
```

### Scenario: 筛选控件具备可访问语义

```gherkin
Given 搜索页已渲染筛选控件
When 用户通过键盘导航筛选项
Then 控件应具备 radiogroup/radio 语义
And 当前选中项应有明确 aria 状态
```

## Testing Strategy

### Unit Tests (Vitest + RTL)

| Test | File | What it validates |
|------|------|------------------|
| `defaults to all filter when ext missing` | `web/src/components/search-page.test.tsx` | 默认筛选值 |
| `hydrates filter from url ext` | `web/src/components/search-page.test.tsx` | URL 初始化恢复 |
| `fallbacks to all on invalid ext` | `web/src/components/search-page.test.tsx` | 非法值兜底 |
| `updates url when filter changes` | `web/src/components/search-page.test.tsx` | 筛选与 URL 同步 |
| `does not alter appsearch request contract` | `web/src/components/search-page.test.tsx` | 不改请求契约 |
| `classifies file extensions correctly` | `web/src/lib/file-category.test.ts` | 分类规则准确性 |
| `supports multi-part extension archive` | `web/src/lib/file-category.test.ts` | `tar.gz` 等边界 |
| `filter control is accessible` | `web/src/tests/accessibility.test.tsx` | 语义与可达性 |

### Regression Checks

- 现有搜索页核心用例（初始态/有结果/无结果/错误态/清空）仍需通过
- 计数文案与列表长度一致
- Connect path 与请求结构保持不变
