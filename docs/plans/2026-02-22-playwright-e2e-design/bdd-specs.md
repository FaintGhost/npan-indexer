# BDD Specifications

## Feature 1: Search Flow

```gherkin
Feature: 搜索页面
  作为用户，我希望通过关键词搜索文件并浏览结果

  Background:
    Given Meilisearch 中已播种 38 条测试文档
    And 用户访问搜索页面 "/"

  Scenario: 初始状态显示欢迎界面
    Then 页面显示 hero 模式 (.mode-hero)
    And 标题 "Npan Search" 可见
    And 状态文本显示 "随时准备为您检索文件"
    And 搜索结果列表为空

  Scenario: 输入关键词后防抖触发搜索
    When 用户在搜索框输入 "quarterly"
    Then 等待 API 响应 GET /api/v1/app/search?query=quarterly
    And 页面切换到 docked 模式 (.mode-docked)
    And 搜索结果列表至少有 1 条结果
    And 状态文本匹配 "已加载 N 个文件"

  Scenario: 点击搜索按钮立即搜索（跳过防抖）
    When 用户在搜索框输入 "project"
    And 用户点击搜索按钮
    Then API 请求立即发出（不等待 280ms）
    And 搜索结果列表至少有 1 条结果

  Scenario: 按 Enter 键立即搜索
    When 用户在搜索框输入 "design" 并按 Enter
    Then API 请求立即发出
    And 搜索结果列表至少有 1 条结果

  Scenario: 无结果时显示空状态
    When 用户搜索 "xyzzy-nonexistent-99999"
    Then 等待搜索完成
    And 页面显示 "未找到相关文件"
    And 搜索结果列表为空

  Scenario: 清空搜索恢复初始状态
    Given 用户已搜索 "test" 并看到结果
    When 用户点击清空按钮
    Then 搜索框为空
    And 页面恢复 hero 模式
    And 状态文本显示 "随时准备为您检索文件"

  Scenario: 无限滚动加载更多
    When 用户搜索 "test-file"（已播种 35 条匹配文档）
    Then 首页加载 30 条结果
    When 用户滚动到哨兵元素
    Then 等待第二页 API 响应 (page=2)
    And 结果列表总数为 35

  Scenario: Cmd/Ctrl+K 聚焦搜索框
    When 用户按下 Cmd+K (Mac) 或 Ctrl+K (其他)
    Then 搜索框获得焦点

  Scenario: 搜索带分页参数
    When 用户搜索 "test" 并附加参数 page=1&page_size=5
    Then API 请求包含正确的分页参数
    And 返回结果数不超过 5

  Scenario: 视图模式切换动画
    When 用户在搜索框输入任意文字
    Then 页面从 .mode-hero 过渡到 .mode-docked
    When 用户清空搜索框
    Then 页面从 .mode-docked 过渡回 .mode-hero
```

## Feature 2: Download Flow

```gherkin
Feature: 文件下载
  作为用户，我希望下载搜索结果中的文件

  Background:
    Given Meilisearch 中已播种测试文档
    And 下载 API (/api/v1/app/download-url) 已被 mock 返回成功响应
    And 用户已搜索并看到结果

  Scenario: 下载按钮初始状态
    Then 每个文件卡片显示 "下载" 按钮 (idle 状态)

  Scenario: 点击下载显示加载状态
    When 用户点击第一个文件的下载按钮
    Then 按钮文本变为 "获取中"
    And 按钮显示旋转图标 (animate-spin)
    And 按钮为 disabled 状态

  Scenario: 下载成功后显示成功状态
    When 用户点击下载按钮
    Then 等待 API 响应成功
    And 按钮文本变为 "成功" (绿色)
    And window.open 被调用（参数包含 download_url）
    And 1.5 秒后按钮恢复为 "下载"

  Scenario: 下载失败显示重试状态
    Given 下载 API mock 返回 502 错误
    When 用户点击下载按钮
    Then 按钮文本变为 "重试" (红色)
    And 按钮仍可点击

  Scenario: 缓存的下载不重复请求 API
    When 用户点击同一文件的下载按钮两次（间隔足够让第一次完成）
    Then API 请求只发出 1 次
    And 第二次直接调用 window.open

  Scenario: 多个文件可同时下载
    When 用户快速点击两个不同文件的下载按钮
    Then 两个按钮都进入 loading 状态
    And 两个 API 请求并行发出
```

## Feature 3: Admin Full Flow

```gherkin
Feature: 管理后台
  作为管理员，我希望通过 API Key 认证后管理同步任务

  Background:
    Given 系统配置 NPA_ADMIN_API_KEY = "ci-test-admin-api-key-1234"

  # --- 认证 ---

  Scenario: 未认证时显示 API Key 对话框
    When 用户访问 /admin/
    Then 显示全屏 API Key 对话框
    And 密码输入框可见且自动聚焦
    And "确认" 按钮可见

  Scenario: 空 API Key 显示本地错误
    Given 用户在 /admin/ 看到 API Key 对话框
    When 用户不输入任何内容直接点击确认
    Then 显示 "请输入 API Key" 错误提示

  Scenario: 错误 API Key 显示服务端错误
    Given 用户在 /admin/ 看到 API Key 对话框
    When 用户输入 "wrong-key-00000" 并点击确认
    Then 按钮显示 "验证中..." (loading)
    And API 返回 401 后显示 "API Key 无效" 错误
    And 对话框仍然可见

  Scenario: 正确 API Key 进入管理界面
    Given 用户在 /admin/ 看到 API Key 对话框
    When 用户输入 "ci-test-admin-api-key-1234" 并点击确认
    Then 对话框消失
    And 页面显示同步管理界面
    And API Key 保存到 localStorage["npan_admin_api_key"]

  Scenario: 刷新页面保持认证状态
    Given 用户已通过 API Key 认证
    When 用户刷新页面
    Then 不显示 API Key 对话框
    And 直接显示同步管理界面

  # --- 同步控制 ---

  Scenario: 显示同步模式选择器
    Given 用户已认证进入管理界面
    Then 显示三个模式按钮：自适应、全量、增量
    And "自适应" 为默认选中状态

  Scenario: 选择全量模式
    Given 用户已认证进入管理界面
    When 用户点击 "全量" 按钮
    Then "全量" 按钮为选中状态
    And 其他模式按钮为未选中状态

  Scenario: 启动同步发送正确请求
    Given 用户已认证并选择 "全量" 模式
    When 用户点击 "启动同步" 按钮
    Then POST /api/v1/admin/sync 请求发出
    And 请求头包含 X-API-Key
    And 请求体包含 mode: "full"
    And 显示 "同步任务已启动" 成功提示

  Scenario: 同步运行中显示进度
    Given 同步已启动（POST 返回 202）
    Then 页面开始轮询 GET /api/v1/admin/sync（每 2s）
    And 显示进度条和状态徽章
    And 启动按钮变为不可用或隐藏
    And 显示取消按钮

  Scenario: 取消同步需要确认
    Given 同步正在运行
    When 用户点击 "取消同步"
    Then 浏览器弹出确认对话框 "确认取消同步？"
    When 用户点击确认
    Then DELETE /api/v1/admin/sync 请求发出
    And 轮询停止

  Scenario: 取消确认框点击取消不发请求
    Given 同步正在运行
    When 用户点击 "取消同步"
    And 用户在确认框点击取消
    Then 不发送 DELETE 请求
    And 同步继续运行

  Scenario: 状态徽章颜色正确
    Given 同步进度返回以下状态之一
    Then 徽章颜色正确：
      | status      | color   |
      | idle        | gray    |
      | running     | blue    |
      | done        | green   |
      | error       | red     |
      | interrupted | amber   |
      | cancelled   | gray    |

  # --- 导航 ---

  Scenario: 返回搜索链接
    Given 用户已认证进入管理界面
    When 用户点击 "返回搜索" 链接
    Then 页面导航到 "/"
```

## Feature 4: Edge Cases

```gherkin
Feature: 边界场景
  覆盖异常和边界条件

  Scenario: 搜索特殊字符
    Given 用户访问搜索页面
    When 用户搜索 "C++ & .NET"
    Then API 请求中 query 参数正确 URL 编码
    And 不触发前端错误

  Scenario: 非常长的搜索查询
    Given 用户访问搜索页面
    When 用户输入 200 字符的搜索词
    Then 搜索正常执行，不截断不报错

  Scenario: 快速连续搜索（防抖竞态）
    Given 用户访问搜索页面
    When 用户快速依次输入 "a", "ab", "abc"
    Then 最终只有 query="abc" 的结果被渲染
    And 前两次请求被取消或结果被丢弃

  Scenario: 网络错误时显示错误状态
    Given 搜索 API 被 mock 返回网络错误
    When 用户搜索任意关键词
    Then 页面显示错误状态 (.border-rose-200)
    And 错误文本可见

  Scenario: 搜索框纯空格不触发搜索
    Given 用户访问搜索页面
    When 用户在搜索框输入 "   "（纯空格）
    Then 不发出 API 请求
    And 保持初始状态

  Scenario: 浏览器后退/前进导航
    Given 用户在搜索页面搜索了 "test"
    When 用户导航到 /admin/ 再点击浏览器后退
    Then 返回搜索页面 "/"

  Scenario: Admin 认证过期（401 清除）
    Given 用户已认证但 localStorage 中的 key 被手动清除
    When 用户刷新 /admin/ 页面
    Then 显示 API Key 对话框
```

## Testing Strategy

### 测试数据

在 `beforeAll` 中通过 Meilisearch API 播种 38 条文档：
- 3 条具名文档 (quarterly-report, project-design, architecture-diagram)
- 35 条批量文档 (test-file-000 ~ test-file-034) 用于无限滚动

### Mock 策略

| 端点 | Mock? | 原因 |
|------|-------|------|
| GET /api/v1/app/search | 不 mock | 测试真实 Meilisearch 搜索 |
| GET /api/v1/app/download-url | mock | CI 中 NPA_TOKEN 为 dummy，真实请求会失败 |
| POST /api/v1/admin/sync | 不 mock | 测试真实同步启动/取消生命周期 |
| GET /api/v1/admin/sync | 不 mock | 测试真实进度查询 |
| DELETE /api/v1/admin/sync | 不 mock | 测试真实取消 |

### 预计测试用例分布

| 文件 | 用例数 | 说明 |
|------|--------|------|
| search.spec.ts | ~12 | 搜索 + 下载 + 边界 |
| admin.spec.ts | ~10 | 认证 + 同步 + 导航 |
| 合计 | ~22 | |
