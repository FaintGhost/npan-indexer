## Feature: End User 搜索下载 Demo

### Scenario 1: 页面不暴露凭据输入

Given 用户访问 `/demo`  
When 页面渲染完成  
Then 页面应只有搜索输入与结果区域  
And 不应出现 API Key 或 Token 输入控件

### Scenario 2: 输入关键词后动态返回文件结果

Given 用户在搜索框输入关键词  
When debounce 时间到达并触发请求  
Then 页面应调用 demo 搜索接口拉取第 1 页文件结果  
And 结果列表应渲染文件名和下载操作

### Scenario 3: 滚动到底部触发懒加载

Given 用户已经看到第一页结果  
When 列表底部进入可视区域  
Then 页面应继续请求下一页结果  
And 直到达到 total 后停止请求

### Scenario 4: 点击下载由后端代理生成链接

Given 用户点击某个文件的下载按钮  
When 页面调用 demo 下载接口  
Then 后端应使用服务端配置凭据生成下载链接  
And 前端应直接跳转或打开该下载链接
