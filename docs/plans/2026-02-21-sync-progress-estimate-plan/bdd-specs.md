## Feature: sync-full 估算进度

### Scenario 1: 可获取根目录总量时输出估算百分比

Given 根目录进度包含可用的 `estimatedTotalDocs`  
When CLI 以 `human` 模式渲染 `sync-full` 进度  
Then 输出应包含估算进度百分比与 `done/total` 文档数  
And 输出应包含估算覆盖根目录数量（`known/total roots`）

### Scenario 2: 无法获取总量时回退 n/a

Given 根目录进度不包含可用的 `estimatedTotalDocs`  
When CLI 以 `human` 模式渲染 `sync-full` 进度  
Then 输出应显示 `est=n/a`  
And 不影响原有统计字段输出

### Scenario 3: 部门根目录自动注入估算总量

Given 同步入口启用了部门根目录发现  
When `discoverRootFolders` 返回部门目录列表（包含 `item_count`）  
Then 创建/恢复的 root progress 应写入 `estimatedTotalDocs=item_count+1`  
And 断点续跑恢复后仍保留该估算字段
