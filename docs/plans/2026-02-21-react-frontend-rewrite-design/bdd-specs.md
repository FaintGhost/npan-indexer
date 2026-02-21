# BDD 行为规范

## 1. 搜索功能

```gherkin
Feature: 文件搜索
  用户可以通过关键词搜索云盘中的文件，
  结果以卡片列表形式展示，支持无限滚动分页。

  Background:
    Given 用户已打开搜索页面 "/app"
    And 搜索 API 端点为 "GET /api/v1/app/search"

  # --- 空状态 ---

  Scenario: 初始空状态显示引导提示
    When 页面首次加载完成
    Then 搜索框处于 Hero 居中模式
    And 搜索输入框为空
    And 结果区域显示"等待探索"引导提示
    And 计数器显示 "0 / 0"
    And 状态栏显示"随时准备为您检索文件"

  # --- 搜索带结果 ---

  Scenario: 输入关键词后自动搜索并显示结果
    Given API 返回 3 条文件结果且 total 为 3
    When 用户在搜索框输入 "MX40"
    And 等待 debounce 时间（280ms）过后
    Then 应发送请求 "GET /api/v1/app/search?query=MX40&page=1&page_size=24"
    And 页面切换到 Docked 吸顶模式
    And 结果区域显示 3 张文件卡片
    And 计数器显示 "3 / 3"
    And 状态栏显示"已加载 3 / 3 个文件"

  Scenario: 文件卡片正确展示信息
    Given API 返回 1 条文件结果：
      | name        | size    | modified_at   | source_id | highlighted_name          |
      | MX40固件.bin | 1048576 | 1700000000000 | 42        | <mark>MX40</mark>固件.bin |
    When 搜索完成并渲染结果
    Then 文件卡片应显示高亮名称 "MX40固件.bin"（MX40 部分高亮）
    And 文件卡片应显示大小 "1 MB"
    And 文件卡片应显示修改日期（格式化后的时间）
    And 文件卡片应显示对应扩展名的图标（bin → 固件图标，紫色）
    And 文件卡片应包含下载按钮

  Scenario: 搜索框按 Enter 立即触发搜索（不等待 debounce）
    Given 搜索框中已输入 "固件"
    When 用户按下 Enter 键
    Then 应立即发送搜索请求，不等待 debounce 延迟
    And 之前排队的 debounce 定时器应被取消

  # --- 无结果 ---

  Scenario: 搜索无结果时显示空状态
    Given API 返回 0 条结果且 total 为 0
    When 用户搜索 "不存在的文件名"
    Then 结果区域显示"未找到相关文件"空状态卡片
    And 状态栏显示"未找到相关文件"

  # --- 错误状态 ---

  Scenario: 搜索 API 返回错误时显示错误状态
    Given API 返回 HTTP 500 错误
    When 用户搜索 "测试"
    Then 结果区域显示"加载出错了"错误状态卡片
    And 状态栏显示错误信息（红色文字）

  Scenario: 网络断开时显示错误状态
    Given 网络不可用（fetch 抛出 TypeError）
    When 用户搜索 "测试"
    Then 结果区域显示"加载出错了"错误状态卡片

  # --- 分页/无限滚动 ---

  Scenario: 滚动到底部自动加载下一页
    Given 首次搜索返回 24 条结果且 total 为 50
    And 页面处于 Docked 模式
    When 用户滚动到结果列表底部触发 IntersectionObserver
    Then 应发送请求 "GET /api/v1/app/search?query=...&page=2&page_size=24"
    And 新结果追加到现有列表末尾（不替换）
    And 计数器更新为 "48 / 50"

  Scenario: 所有结果已加载时不再请求
    Given 列表已显示全部 50 条结果（total 为 50）
    When 用户滚动到底部
    Then 不应发送新的 API 请求

  Scenario: 加载中不重复请求
    Given 正在加载第 2 页数据
    When 用户再次滚动到底部
    Then 不应发送重复请求

  Scenario: 去重——同一文件不重复显示
    Given 第 1 页返回 source_id 为 [1,2,3] 的结果
    And 第 2 页返回 source_id 为 [3,4,5] 的结果
    When 两页数据都已加载
    Then 列表中应有 5 张不重复的卡片

  # --- Debounce ---

  Scenario: 快速连续输入只触发一次请求
    When 用户在 200ms 内连续输入 "M"、"MX"、"MX4"、"MX40"
    Then 在最后一次输入后的 280ms debounce 窗口内只发送 1 次请求
    And 请求参数 query 为 "MX40"

  Scenario: 新搜索取消进行中的旧请求
    Given 正在请求 query="旧关键词" 的结果
    When 用户输入新关键词 "新关键词" 并触发搜索
    Then 旧请求应被中止（AbortController.abort）
    And 发送 query="新关键词" 的新请求

  # --- 清空输入 ---

  Scenario: 点击清空按钮恢复初始状态
    Given 搜索框中有文字且页面处于 Docked 模式
    When 用户点击清空按钮（X 图标）
    Then 搜索框清空
    And 页面切换回 Hero 居中模式
    And 结果区域恢复"等待探索"引导提示
    And 计数器恢复为 "0 / 0"
    And 状态栏恢复为"随时准备为您检索文件"
    And 搜索框重新获得焦点

  Scenario: 搜索框有文字时显示清空按钮，无文字时显示快捷键提示
    When 搜索框为空
    Then 显示键盘快捷键徽章 "⌘K"（Mac）或 "Ctrl K"（非 Mac）
    And 清空按钮隐藏
    When 搜索框输入任意文字
    Then 清空按钮显示
    And 快捷键徽章隐藏
```

## 2. 下载功能

```gherkin
Feature: 文件下载
  用户可以点击下载按钮获取文件下载链接，
  链接在新标签页中打开。

  Background:
    Given 用户已在搜索结果中看到文件卡片
    And 下载 API 端点为 "GET /api/v1/app/download-url"

  # --- 点击下载 ---

  Scenario: 点击下载按钮获取下载链接
    Given 文件卡片 source_id 为 42
    When 用户点击该卡片的下载按钮
    Then 应发送请求 "GET /api/v1/app/download-url?file_id=42"

  # --- 加载状态 ---

  Scenario: 下载按钮在请求期间显示加载状态
    When 用户点击下载按钮
    And 下载链接请求正在进行中
    Then 按钮应显示旋转加载图标和"获取中"文字
    And 按钮应处于禁用状态（不可重复点击）
    And 状态栏显示"正在生成下载链接..."

  # --- 成功 ---

  Scenario: 下载链接获取成功后打开新标签页
    Given API 返回 download_url 为 "https://example.com/file.bin"
    When 下载请求成功
    Then 按钮应显示绿色勾号图标和"成功"文字
    And 应调用 window.open 在新标签页打开 "https://example.com/file.bin"
    And 状态栏显示"已打开下载链接"
    And 1.5 秒后按钮恢复为默认"下载"状态

  # --- 缓存 ---

  Scenario: 重复下载同一文件使用缓存链接
    Given 文件 42 的下载链接已经获取过
    When 用户再次点击文件 42 的下载按钮
    Then 不应发送新的 API 请求
    And 应直接打开缓存的下载链接

  # --- 错误/重试 ---

  Scenario: 下载链接获取失败显示重试按钮
    Given API 返回 HTTP 502 错误
    When 下载请求失败
    Then 按钮应显示红色"重试"文字
    And 按钮恢复为可点击状态
    And 状态栏显示错误信息（红色文字）

  Scenario: 点击重试按钮重新请求下载链接
    Given 下载按钮当前显示"重试"状态
    When 用户点击重试按钮
    Then 应重新发送下载链接请求
    And 按钮切换到加载状态

  Scenario: 下载链接为空视为错误
    Given API 返回 download_url 为空字符串
    When 下载请求完成
    Then 应视为错误，显示"重试"按钮
    And 状态栏显示"下载链接为空"
```

## 3. UI 模式过渡

```gherkin
Feature: Hero / Docked 模式过渡
  搜索框在未搜索时居中显示（Hero），
  搜索开始后吸顶显示（Docked），
  通过 View Transition API 实现动画过渡。

  Background:
    Given 浏览器支持 View Transition API（document.startViewTransition 可用）

  # --- Hero → Docked ---

  Scenario: 搜索触发 Hero → Docked 过渡
    Given 页面处于 Hero 模式（搜索框居中，body 有 mode-hero 类）
    When 用户触发搜索（输入文字 debounce 后或按 Enter）
    Then 应通过 View Transition API 执行模式切换
    And body 类从 "mode-hero" 切换为 "mode-docked"
    And 搜索框区域吸顶（sticky top:0）并添加半透明背景和毛玻璃效果
    And 结果区域从隐藏（opacity:0, translateY:60px）过渡到可见
    And 过渡动画时长约 650ms，使用 cubic-bezier(0.22, 1, 0.36, 1) 缓动

  # --- Docked → Hero ---

  Scenario: 清空搜索触发 Docked → Hero 过渡
    Given 页面处于 Docked 模式
    When 用户点击清空按钮
    Then 应通过 View Transition API 执行模式切换
    And body 类从 "mode-docked" 切换为 "mode-hero"
    And 搜索框区域回到居中全屏布局
    And 结果区域淡出并隐藏

  # --- 浏览器不支持 View Transition ---

  Scenario: 浏览器不支持 View Transition 时直接切换（无动画降级）
    Given 浏览器不支持 View Transition API（document.startViewTransition 为 undefined）
    When 用户触发搜索
    Then 模式切换立即生效（无动画）
    And 功能不受影响

  # --- 骨架屏 ---

  Scenario: 首次搜索从 Hero 切换时显示骨架屏
    Given 页面处于 Hero 模式
    When 用户触发搜索
    Then 在 API 响应到达之前，结果区域应显示 5 个骨架屏卡片
    And 骨架屏有脉冲动画效果

  Scenario: Docked 模式下重新搜索显示半透明遮罩
    Given 页面处于 Docked 模式且已有搜索结果
    When 用户输入新关键词触发搜索
    Then 现有结果列表应添加半透明遮罩（opacity:0.5, pointer-events:none）
    And API 响应后遮罩移除并更新结果
```

## 4. 键盘快捷键

```gherkin
Feature: 键盘快捷键
  用户可以通过键盘快捷键快速操作。

  Scenario: Mac 用户按 Cmd+K 聚焦搜索框
    Given 用户的操作系统为 macOS
    And 搜索框未获得焦点
    When 用户按下 Cmd+K
    Then 搜索框应获得焦点
    And 浏览器默认行为应被阻止（preventDefault）

  Scenario: 非 Mac 用户按 Ctrl+K 聚焦搜索框
    Given 用户的操作系统为 Windows/Linux
    And 搜索框未获得焦点
    When 用户按下 Ctrl+K
    Then 搜索框应获得焦点
    And 浏览器默认行为应被阻止

  Scenario: 快捷键徽章根据平台显示正确文字
    Given 用户的操作系统为 macOS
    Then 搜索框右侧徽章显示 "⌘K"
    Given 用户的操作系统为 Windows/Linux
    Then 搜索框右侧徽章显示 "Ctrl K"

  Scenario: 焦点已在搜索框时快捷键仍然正常响应
    Given 搜索框已获得焦点
    When 用户按下 Cmd/Ctrl+K
    Then 搜索框保持焦点（不出错）
```

## 5. 管理页认证

```gherkin
Feature: Admin 页面 API Key 认证
  管理页面需要 API Key 才能访问后端管理接口，
  Key 通过输入对话框获取并持久化到 localStorage。

  Background:
    Given 用户导航到管理页面 "/app/admin"

  # --- 输入 API Key ---

  Scenario: 首次访问管理页且无存储的 Key 时弹出输入对话框
    Given localStorage 中没有存储 API Key
    When 管理页面加载完成
    Then 应显示 API Key 输入对话框
    And 对话框包含密码类型输入框和"确认"按钮
    And 管理功能区域不可见或不可操作

  Scenario: 输入有效 API Key 后存储并关闭对话框
    Given API Key 输入对话框已显示
    When 用户输入 "valid-admin-api-key-32chars-min!!"
    And 用户点击"确认"按钮
    Then 应发送验证请求（如 GET /api/v1/admin/sync/full/progress）携带 X-API-Key Header
    And 如果 API 响应非 401，则视为 Key 有效
    And API Key 应存储到 localStorage（键名 "npan_admin_api_key"）
    And 输入对话框关闭
    And 管理功能区域正常显示

  # --- localStorage 持久化 ---

  Scenario: 已存储的 API Key 自动使用
    Given localStorage 中已存储有效 API Key
    When 管理页面加载完成
    Then 不应弹出输入对话框
    And 管理功能区域直接显示
    And 后续 API 请求自动携带存储的 X-API-Key Header

  Scenario: 刷新页面后仍可使用存储的 Key
    Given localStorage 中已存储 API Key
    When 用户刷新浏览器
    Then 管理页面正常加载，无需重新输入 Key

  # --- 无效 Key 处理 ---

  Scenario: 输入无效 API Key 显示错误提示
    Given API Key 输入对话框已显示
    When 用户输入 "wrong-key" 并确认
    And 验证请求返回 HTTP 401
    Then 对话框应显示错误提示"API Key 无效，请重新输入"
    And 输入框清空，等待用户重新输入
    And Key 不应被存储到 localStorage

  Scenario: 已存储的 Key 失效时重新提示输入
    Given localStorage 中存储的 API Key 已过期或被服务端更改
    When 管理页面发送请求收到 HTTP 401 响应
    Then 应清除 localStorage 中存储的 Key
    And 重新弹出 API Key 输入对话框
    And 显示提示"API Key 已失效，请重新输入"

  Scenario: 空输入提交不发送验证请求
    Given API Key 输入对话框已显示
    When 用户不输入内容直接点击"确认"
    Then 不应发送任何请求
    And 输入框应显示验证错误"请输入 API Key"
```

## 6. 管理页同步功能

```gherkin
Feature: 全量同步管理
  管理员可以启动、监控和取消全量同步任务。

  Background:
    Given 用户已在管理页面 "/app/admin"
    And 用户已通过 API Key 认证
    And 同步相关 API 端点使用 "X-API-Key" Header 认证

  # --- 启动同步 ---

  Scenario: 点击启动全量同步
    Given 当前没有运行中的同步任务
    When 用户点击"启动全量同步"按钮
    Then 应发送 "POST /api/v1/admin/sync/full" 请求
    And 请求体为 JSON（可选包含 root_folder_ids 等参数）
    And 请求携带 X-API-Key Header

  Scenario: 启动同步成功
    Given API 返回 HTTP 202 和 { "message": "全量同步任务已启动" }
    When 同步启动请求成功
    Then 应显示成功提示"全量同步任务已启动"
    And 自动开始轮询同步进度
    And "启动全量同步"按钮变为禁用状态

  Scenario: 已有同步任务运行时启动返回冲突错误
    Given API 返回 HTTP 409（已有同步任务运行）
    When 用户点击"启动全量同步"按钮
    Then 应显示错误提示"已有同步任务在运行中"
    And 按钮保持可用状态

  # --- 查看进度（轮询） ---

  Scenario: 同步运行中自动轮询进度
    Given 同步任务状态为 "running"
    When 管理页面检测到运行中的同步任务
    Then 应每隔 3 秒发送 "GET /api/v1/admin/sync/full/progress" 请求
    And 轮询持续直到状态变为 "done"、"error" 或 "cancelled"

  Scenario: 进度信息正确展示
    Given API 返回同步进度 JSON：
      | 字段                           | 值         |
      | status                         | running    |
      | roots                          | [101, 102] |
      | completedRoots                 | [101]      |
      | activeRoot                     | 102        |
      | aggregateStats.filesIndexed    | 1234       |
      | aggregateStats.pagesFetched    | 56         |
      | aggregateStats.failedRequests  | 2          |
    When 进度数据渲染完成
    Then 应显示状态为"运行中"（带运行指示器）
    And 应显示根目录完成进度 "1 / 2"
    And 应显示当前活跃根目录 ID 为 102
    And 应显示已索引文件数 1234
    And 应显示已抓取页面数 56
    And 应显示失败请求数 2
    And 应显示估算百分比进度条

  Scenario: 同步完成后停止轮询
    Given 轮询中 API 返回 status 为 "done"
    When 进度更新完成
    Then 应停止轮询
    And 应显示状态为"已完成"（带完成图标）
    And "启动全量同步"按钮恢复为可用状态

  Scenario: 同步出错时展示错误信息
    Given API 返回 status 为 "error" 且 lastError 不为空
    When 进度更新完成
    Then 应停止轮询
    And 应显示状态为"出错"（带错误图标）
    And 应显示 lastError 错误信息

  Scenario: 无同步进度记录时显示提示
    Given API 返回 HTTP 404（未找到同步进度）
    When 管理页面加载同步进度
    Then 应显示"暂无同步记录"提示
    And "启动全量同步"按钮可用

  # --- 取消同步 ---

  Scenario: 点击取消正在运行的同步
    Given 同步任务状态为 "running"
    When 用户点击"取消同步"按钮
    Then 应弹出确认对话框"确认取消当前同步任务？"

  Scenario: 确认取消同步
    Given 确认对话框已显示
    When 用户点击"确认"
    Then 应发送 "POST /api/v1/admin/sync/full/cancel" 请求
    And 请求携带 X-API-Key Header

  Scenario: 取消同步成功
    Given API 返回 HTTP 200 和 { "message": "同步取消信号已发送" }
    When 取消请求成功
    Then 应显示提示"同步取消信号已发送"
    And 继续轮询直到状态变为 "cancelled"

  Scenario: 无运行中任务时取消返回冲突错误
    Given API 返回 HTTP 409（当前没有运行中的同步任务）
    When 取消请求失败
    Then 应显示错误提示"当前没有运行中的同步任务"

  # --- 进度百分比估算 ---

  Scenario: 根据 estimatedTotalDocs 计算进度百分比
    Given 同步进度中某 root 的 estimatedTotalDocs 为 1000
    And 该 root 的 filesIndexed 为 500
    When 进度渲染完成
    Then 该 root 应显示约 50% 的进度

  Scenario: 无 estimatedTotalDocs 时不显示百分比
    Given 同步进度中某 root 的 estimatedTotalDocs 为 null
    When 进度渲染完成
    Then 该 root 应显示"进度未知"或仅显示已索引数量
```

## 7. 路由导航

```gherkin
Feature: 客户端路由
  应用使用 TanStack Router 实现 SPA 路由，
  搜索页和管理页通过客户端导航切换。

  Background:
    Given 应用使用 TanStack Router 配置文件路由
    And 基础路径为 "/app"

  # --- 路由定义 ---

  Scenario: 访问 /app 渲染搜索页
    When 用户导航到 "/app"
    Then 应渲染搜索页组件
    And 页面标题为 "Npan Search"

  Scenario: 访问 /app/admin 渲染管理页
    When 用户导航到 "/app/admin"
    Then 应渲染管理页组件

  Scenario: 访问未定义路由显示 404
    When 用户导航到 "/app/unknown-path"
    Then 应渲染 404 未找到页面
    And 页面提供返回搜索页的链接

  # --- 导航 ---

  Scenario: 搜索页导航到管理页
    Given 用户当前在搜索页 "/app"
    When 用户点击管理页入口链接
    Then 应客户端导航到 "/app/admin"（不刷新页面）
    And URL 更新为 "/app/admin"

  Scenario: 管理页导航回搜索页
    Given 用户当前在管理页 "/app/admin"
    When 用户点击返回搜索页链接
    Then 应客户端导航到 "/app"
    And 搜索页恢复初始 Hero 模式

  # --- 代码分割 ---

  Scenario: 管理页懒加载
    Given 用户首次加载搜索页
    Then 管理页的 JavaScript 代码不应包含在初始 bundle 中
    When 用户首次导航到 "/app/admin"
    Then 应异步加载管理页代码块（dynamic import）
    And 加载期间显示加载指示器

  # --- 浏览器直接访问 ---

  Scenario: 浏览器直接访问 /app/admin
    When 用户在地址栏直接输入 "/app/admin" 并回车
    Then Go 服务应返回前端 index.html（SPA fallback）
    And TanStack Router 在客户端解析路由为管理页
    And 管理页正常渲染

  # --- 搜索参数持久化 ---

  Scenario: 搜索关键词通过 URL search params 持久化
    When 用户搜索 "MX40"
    Then URL 应更新为 "/app?query=MX40"
    When 用户刷新浏览器
    Then 搜索框应预填充 "MX40"
    And 自动触发搜索

  Scenario: 分享搜索 URL 可直接显示结果
    Given URL 为 "/app?query=固件"
    When 新用户在浏览器中打开该 URL
    Then 搜索框应显示"固件"
    And 自动执行搜索并显示结果
```

## 8. API 响应校验

```gherkin
Feature: API 响应 Zod Schema 校验
  所有 API 响应在客户端经过 Zod schema 校验，
  确保数据结构符合预期。

  Scenario: 搜索 API 响应通过 schema 校验
    Given 搜索 API 返回符合预期结构的 JSON
    When 响应经过 Zod schema 解析
    Then 应成功解析出 items 数组和 total 数值
    And items 中每个元素包含 doc_id, source_id, name, size 等必填字段

  Scenario: 搜索 API 响应结构异常时优雅降级
    Given 搜索 API 返回不符合预期的 JSON（如缺少 total 字段）
    When 响应经过 Zod schema 解析
    Then 应捕获 ZodError
    And 向用户显示友好的错误提示（非原始错误信息）

  Scenario: 下载 URL API 响应通过 schema 校验
    Given 下载 API 返回 { "file_id": 42, "download_url": "https://..." }
    When 响应经过 Zod schema 解析
    Then 应成功解析出 file_id 和 download_url
    And download_url 为非空字符串

  Scenario: 同步进度 API 响应通过 schema 校验
    Given 同步进度 API 返回完整的 SyncProgressState JSON
    When 响应经过 Zod schema 解析
    Then 应成功解析出 status, roots, completedRoots, aggregateStats 等字段
    And aggregateStats 包含 filesIndexed, pagesFetched, failedRequests 数值字段

  Scenario: 错误响应通过 schema 校验
    Given API 返回错误响应 { "code": "BAD_REQUEST", "message": "缺少 query 参数" }
    When 响应经过 Zod error schema 解析
    Then 应成功解析出 code 和 message 字段
```

## 9. 可访问性

```gherkin
Feature: 可访问性（Accessibility）
  应用符合 WCAG 2.1 AA 级标准的基本要求。

  Scenario: 搜索结果区域为 ARIA live region
    Given 搜索结果区域存在
    Then 结果列表容器应设置 aria-live="polite"
    When 搜索结果更新时
    Then 屏幕阅读器应自动播报变化

  Scenario: 搜索输入框有正确的 ARIA 标签
    Then 搜索输入框应有 aria-label 或关联的 label 元素
    And placeholder 文字描述搜索功能

  Scenario: 下载按钮有描述性标签
    Given 文件 "MX40固件.bin" 的下载按钮存在
    Then 按钮应有 aria-label 如 "下载 MX40固件.bin"
    And 按钮加载状态变化时 aria-label 应同步更新

  Scenario: 键盘可完整操作
    Then 所有交互元素可通过 Tab 键依次聚焦
    And 按钮可通过 Enter 或 Space 激活
    And 对话框可通过 Escape 关闭

  Scenario: 颜色对比度满足 AA 标准
    Then 所有文本颜色与背景颜色的对比度 >= 4.5:1（正常文字）
    And 大号文字对比度 >= 3:1

  Scenario: 加载状态有 ARIA 反馈
    When 搜索结果正在加载
    Then 应有 aria-busy="true" 属性
    And 骨架屏应有 aria-hidden="true"
    And 应有视觉隐藏的状态文本供屏幕阅读器读取
```
