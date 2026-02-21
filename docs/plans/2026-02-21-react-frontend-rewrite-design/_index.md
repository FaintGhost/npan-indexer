# Npan React 前端重写设计

## 项目背景与动机

npan 是一个 Go 语言服务（Echo v5 + Meilisearch），作为 Novastar 内网云盘（Npan）的外部索引服务。服务将 Npan 云盘文件元数据同步到 Meilisearch，为用户提供高性能的本地全文搜索与直接下载功能。

### 当前前端状态

当前前端是单个 HTML 文件（`web/app/index.html`，约 780 行），采用内联 vanilla JS + Tailwind CSS v4 CDN 方案。具备以下功能：

- **搜索**：280ms debounce 输入即搜、Enter 键搜索、搜索按钮点击搜索
- **无限滚动**：IntersectionObserver 实现懒加载，180px rootMargin 预触发
- **文件卡片**：根据文件扩展名分类显示 5 类图标（压缩包/安装包/固件/文档/通用文件）
- **下载流程**：点击获取下载链接 → 新窗口打开，含 loading/success/error 状态反馈
- **View Transition**：hero（居中搜索框）与 docked（顶部固定搜索栏 + 结果列表）两种模式间平滑过渡，使用 CSS View Transition API，650ms cubic-bezier 动画
- **Cmd/Ctrl+K 快捷键**：全局聚焦搜索框，自动检测 Mac/非 Mac 平台
- **骨架屏**：首次搜索显示 5 个 skeleton cards，带交错 pulse 动画
- **搜索高亮**：Meilisearch 返回 `<mark>` 标签高亮匹配关键词
- **请求竞态控制**：AbortController + requestSeq 序列号双重保护
- **无管理页面**：管理操作（同步管理）仅能通过 curl 或 API 客户端完成

### 重写动机

1. **可维护性**：780 行单文件无模块化，HTML 字符串拼接（`cardHTML()` 等函数），新功能增加困难
2. **类型安全**：vanilla JS 无类型检查，API 响应结构变化时无编译期提示
3. **新功能需求**：需新增 Admin 管理页面（同步管理），单文件架构无法支撑多页面路由
4. **开发体验**：无 HMR、无组件复用、无 IDE 自动补全/重构支持
5. **部署升级**：从服务端运行时读取单个 HTML 文件，升级为 Go `embed.FS` 嵌入 Vite 构建产物的单二进制部署

### 目标状态

- 使用 React 19 + TypeScript + Vite 7 重写为现代 SPA
- 保留所有现有搜索/下载功能与 UI 效果（视觉一致性）
- 新增 Admin 管理页面（API Key 认证、同步管理）
- Tailwind CSS v4 原生配置（CSS-first，非 CDN）
- TanStack Router 实现类型安全的客户端路由
- Zod 进行运行时 API 响应校验
- oxlint + oxfmt + tsgolint 代码质量保证
- 构建产物嵌入 Go 二进制（`embed.FS`），保持单二进制部署

## 用户决策

1. **技术栈**: React 19 + Vite 7 + TypeScript + Tailwind CSS v4 + TanStack Router + Zod
2. **工具链**: oxlint + oxfmt（代码检查与格式化）+ tsgolint（TypeScript 规范）
3. **构建集成**: Vite 构建产物输出至 `web/dist/`，Go 使用 `embed.FS` 嵌入二进制
4. **路由**: `/app` → 搜索页，`/app/admin` → 管理页
5. **认证模型**: 搜索页无需认证（EmbeddedAuth 中间件自动处理）；管理页前端侧 API Key 存储于 localStorage，通过 `X-API-Key` Header 传递
6. **管理功能**: 启动全量同步、查看同步进度（轮询）、取消同步
7. **状态管理**: 不引入全局状态管理库（无 Redux/Zustand），使用 React hooks + Context
8. **部署目标**: 2C2G 服务器，Go 服务与 Meilisearch 同机部署

## 需求列表

### 功能需求

#### Search（搜索功能 —— 现有功能 1:1 重写）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| S-01 | 搜索输入框与 debounce | P0 | 280ms 防抖，支持 Enter 键和搜索按钮触发 |
| S-02 | 搜索结果卡片 | P0 | 文件卡片：扩展名图标、名称（含高亮）、大小、修改时间、ID |
| S-03 | 搜索高亮 | P0 | 使用 `highlighted_name` 字段渲染 `<mark>` 标签高亮匹配关键词 |
| S-04 | 无限滚动分页 | P0 | IntersectionObserver + sentinel 元素，180px rootMargin 预加载 |
| S-05 | 空状态/无结果/错误状态 | P0 | 初始等待、无结果、网络错误三种状态卡片 |
| S-06 | 清空输入 | P1 | 清空按钮回到初始 Hero 状态，隐藏清空/显示快捷键提示 |
| S-07 | 请求竞态处理 | P0 | 新搜索取消旧请求（AbortController），requestSeq 序列号校验 |
| S-08 | 骨架屏加载态 | P0 | 首次搜索显示 5 个 skeleton cards，带交错延时 pulse 动画 |
| S-09 | 状态栏文案 | P1 | 检索中/已加载 N / M 个文件/生成下载链接中等实时文案 |
| S-10 | 结果计数器 | P1 | 显示 "已加载数 / 总数" |
| S-11 | 文件图标分类 | P0 | 5 类图标：压缩包(zip/rar/7z/tar/gz)、安装包(apk/ipa/exe/dmg/含"安装包")、固件(bin/iso/img/rom/含"固件")、文档(pdf/doc/docx/txt/md)、通用 |

#### Download（下载功能）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| D-01 | 下载按钮 | P0 | 获取下载链接后 `window.open` 新标签页打开 |
| D-02 | 下载状态反馈 | P0 | 四态按钮：默认/加载中(spinner)/成功(checkmark, 1.5s 后恢复)/失败(重试) |
| D-03 | 下载链接缓存 | P1 | 相同 file_id 的下载链接在会话内 Map 缓存，避免重复请求 |

#### UI Transition（界面过渡）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| U-01 | Hero → Docked 过渡 | P0 | 搜索触发时从全屏居中切换到吸顶搜索栏 + 显示结果区域 |
| U-02 | Docked → Hero 过渡 | P0 | 清空搜索时恢复全屏居中，隐藏结果区域 |
| U-03 | View Transition 动画 | P1 | 使用 `document.startViewTransition()` API，650ms cubic-bezier(0.22,1,0.36,1) 过渡 |
| U-04 | 搜索结果刷新动画 | P1 | 已处于 Docked 模式时新搜索的 opacity 过渡效果 |

#### Keyboard（键盘交互）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| K-01 | Cmd/Ctrl+K 聚焦 | P0 | 全局快捷键聚焦搜索输入框，Mac 显示 `Cmd K`，其他显示 `Ctrl K` |

#### Admin Auth（管理认证）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| A-01 | API Key 输入 | P0 | 管理页面顶部输入框或首次访问弹窗 |
| A-02 | localStorage 持久化 | P0 | API Key 存储在 localStorage，页面刷新后自动读取 |
| A-03 | 无效 Key 处理 | P0 | 401 响应时清除存储的 Key 并提示重新输入 |
| A-04 | X-API-Key Header | P0 | 所有管理 API 请求自动携带 `X-API-Key` Header |

#### Admin Sync（同步管理）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| AS-01 | 启动全量同步 | P0 | POST `/api/v1/admin/sync/full`，显示操作结果反馈 |
| AS-02 | 同步进度展示 | P0 | GET `/api/v1/admin/sync/full/progress` 轮询，展示进度数据 |
| AS-03 | 取消同步 | P0 | POST `/api/v1/admin/sync/full/cancel`，含确认对话框 |
| AS-04 | 进度详情卡片 | P1 | 展示 status、聚合统计（folders/files/pages/errors）、各根目录进度、预估完成百分比 |
| AS-05 | 冲突提示 | P1 | 同步已在运行时启动返回 409，显示友好提示 |

#### Routing（路由）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| R-01 | 搜索页路由 | P0 | `/app` → 搜索页 |
| R-02 | 管理页路由 | P0 | `/app/admin` → 管理页 |
| R-03 | 管理页懒加载 | P1 | Admin 页面代码分割（React.lazy） |
| R-04 | SPA fallback | P0 | Go 服务端对 `/app/*` 路径返回 `index.html`，支持刷新页面 |

### 非功能需求

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| NF-01 | 中文 UI | P0 | 所有界面文案使用中文 |
| NF-02 | 响应式布局 | P0 | 移动端/桌面端自适应（保留现有 `sm:` 断点设计） |
| NF-03 | 构建产物体积 | P1 | Vite 构建 gzip 后 < 200KB（不含字体 CDN） |
| NF-04 | Go embed 部署 | P0 | 构建产物输出到 `web/dist/`，Go 二进制 `embed.FS` 嵌入并服务 |
| NF-05 | API 响应校验 | P0 | 使用 Zod schema 在运行时校验所有 API 响应，类型从 schema 自动推导 |
| NF-06 | 类型安全路由 | P1 | TanStack Router 提供编译时路由参数类型检查 |
| NF-07 | 代码质量 | P1 | oxlint + oxfmt + tsgolint 统一检查与格式化 |
| NF-08 | 无障碍基础 | P2 | `aria-live` 区域用于搜索结果动态更新、语义化 HTML 标签 |
| NF-09 | 首屏性能 | P1 | LCP < 1.5s（局域网环境） |
| NF-10 | 浏览器兼容 | P1 | Chrome 111+、Edge 111+、Safari 17+（View Transition API 要求） |
| NF-11 | TypeScript strict | P0 | `tsconfig.json` 启用 strict 模式，编译零错误 |
| NF-12 | 字体一致性 | P1 | 保留 Noto Sans SC（正文）+ Outfit（标题）的字体组合 |

## 验收标准

1. **搜索功能完整**：所有 S-01 至 S-11、D-01 至 D-03 功能与现有 HTML 版本行为一致
2. **视觉一致性**：搜索页面视觉效果与现有版本无明显差异（配色、字体、间距、卡片样式、动画时序）
3. **View Transition**：Hero <-> Docked 模式切换带 View Transition 动画，降级方案在不支持的浏览器正常工作
4. **键盘交互**：Cmd/Ctrl+K 快捷键正常聚焦搜索框
5. **管理页面完整**：Admin 页面可完成 API Key 输入、启动同步、查看进度、取消同步完整流程
6. **API 响应校验**：所有 API 响应经过 Zod schema 校验，异常结构触发错误处理而非静默失败
7. **类型安全**：TypeScript strict mode 编译零错误
8. **构建集成**：`npm run build` 输出到 `web/dist/`，Go 服务 `embed.FS` 嵌入后正确服务静态资源
9. **SPA 路由**：`/app` 和 `/app/admin` 均可直接 URL 访问（刷新不 404），Go 服务端处理 SPA fallback
10. **开发体验**：`npm run dev` 启动 Vite dev server，HMR 正常
11. **代码质量**：`oxlint` 和 `tsgolint` 检查通过，`oxfmt` 格式化无差异
12. **Admin 代码分割**：Admin 页面实现懒加载，搜索页首屏不加载管理代码
13. **无障碍**：Lighthouse Accessibility 评分 >= 90
14. **Docker 部署**：Docker 构建成功，`/app` 和 `/app/admin` 在容器内正常可用

## 约束与假设

### 约束（不可变更）

1. **后端 API 不变**：所有 API 端点路径、请求参数、响应格式保持不变
2. **认证模型不变**：搜索页面无需认证（EmbeddedAuth 中间件自动注入服务端凭据），管理页面使用 `X-API-Key` Header
3. **部署环境**：2C2G 服务器，Go 服务与 Meilisearch 同机部署
4. **单二进制部署**：使用 Go `embed.FS` 将前端构建产物嵌入，不依赖运行时外部静态文件
5. **技术栈已选定**：React 19 + TypeScript + Vite 7 + Tailwind CSS v4 + TanStack Router + Zod
6. **现有视觉设计保留**：保留当前的配色方案（蓝色主调/slate灰色系）、字体组合（Noto Sans SC + Outfit）、卡片样式、动画效果
7. **纯 CSR SPA**：不引入 SSR/SSG
8. **中文 UI**：保持中文界面，不做 i18n 国际化

### 假设

1. 用户主要使用桌面端 Chrome/Edge 浏览器（局域网内网环境）
2. 搜索结果量级在万级别，单次分页 24-30 条足够
3. 下载链接有时效性但在单次会话内有效，Map 缓存足够
4. 管理页面使用频率低（日级别），无需复杂的状态管理或实时推送
5. 部署在内网或反向代理之后，TLS 由代理层处理
6. 字体通过 Google Fonts CDN 加载（部署环境可访问公网或配置了字体缓存）

## 关键决策

### D-01: 选择 React 19 + TypeScript

**决策**：使用 React 19 + TypeScript 作为前端框架。

**理由**：
- React 生态最成熟，组件库和工具链选择最广
- TypeScript 严格类型检查确保 API 响应结构变化可在编译期发现
- React 19 的改进（如 `use()` hook、改进的 Suspense）简化异步数据加载
- 与 Zod（`z.infer` 类型推导）、TanStack Router（类型安全路由参数）集成自然

### D-02: 选择 Vite 7 作为构建工具

**决策**：使用 Vite 7 替代 CDN 引入方式。

**理由**：
- 开发环境基于原生 ESM 的 HMR 极快，提升开发体验
- 构建产物 tree-shaking 和代码分割优化好，符合体积约束（gzip < 200KB）
- Tailwind CSS v4 原生支持 Vite 插件（`@tailwindcss/vite`），无需额外配置
- `build.outDir` 可直接配置为 `../web/dist/` 用于 Go embed
- Vite 7 的 Environment API 为未来扩展留有空间

### D-03: 选择 TanStack Router 而非 React Router

**决策**：使用 TanStack Router 进行路由管理。

**理由**：
- 100% 类型安全的路由参数和搜索参数（search params）
- 搜索页面的 query 参数可通过 URL search params 实现，支持分享链接和浏览器前进/后退
- 内置 search params 验证与 Zod 天然配合
- 文件路由约定简化路由定义

### D-04: 使用 Zod 进行运行时 API 响应校验

**决策**：所有 API 响应经过 Zod schema 解析。

**理由**：
- 后端 Go 服务的响应结构可能随版本迭代变化，运行时校验提供安全保障
- `z.infer<typeof schema>` 自动生成 TypeScript 类型，避免手动定义与实际响应不一致
- 解析失败时可提供有意义的错误信息，比静默使用错误数据更安全
- 可对部分字段设置 `.default()` 值增强容错性

### D-05: 选择 oxlint + oxfmt 而非 ESLint + Prettier

**决策**：使用 oxc 工具链（oxlint + oxfmt）进行代码检查和格式化。

**理由**：
- Rust 实现，性能远优于 ESLint/Prettier（50-100x 速度提升）
- 零配置即可覆盖常见 lint 规则
- 单一工具链减少 `node_modules` 依赖复杂度
- 配合 tsgolint 覆盖 TypeScript 特定规范

### D-06: Go embed 替代文件系统静态服务

**决策**：Vite 构建输出到 `web/dist/`，Go 使用 `embed.FS` 嵌入并服务。

**理由**：
- 单二进制部署，无需额外文件拷贝或 volume mount
- 编译时嵌入，部署与运行时文件路径解耦
- SPA fallback 由 Go 路由处理（`/app/*` → `index.html`），确保刷新页面正常
- 与当前 Dockerfile 多阶段构建兼容（先 `npm run build`，再 `go build`）

### D-07: 不引入全局状态管理库

**决策**：不使用 Redux/Zustand 等状态管理库，使用 React 内置 hooks + Context。

**理由**：
- 仅两个页面，状态复杂度低
- 搜索状态通过 URL search params 持久化（TanStack Router 管理）
- Admin 页面状态通过 localStorage（API Key）和服务端（进度数据轮询）管理
- 下载链接缓存使用 `useRef(new Map())` 即可
- 避免引入额外依赖和学习成本

## 不在范围内

- Go 后端代码修改（除 embed/SPA fallback 路由调整外）
- API 接口变更
- SSR / SSG
- i18n 国际化（保持中文）
- E2E 测试框架（Playwright/Cypress）
- PWA 离线支持
- 用户管理系统（仅静态 API Key）
- 监控告警集成（Prometheus 等）
- 缓存层（Redis 等）

## API 端点规约

### 搜索页面接口（无需认证，EmbeddedAuth 自动处理）

#### GET /api/v1/app/search

请求参数（Query String）：
- `query` (string, 必填) - 搜索关键词，别名 `q`
- `page` (number, 默认 1) - 页码，必须为正整数
- `page_size` (number, 默认 30, 最大 100) - 每页数量

响应体（200）：
```json
{
  "items": [
    {
      "doc_id": "file_123",
      "source_id": 123,
      "type": "file",
      "name": "MX40固件V2.1.bin",
      "path_text": "/技术文档/固件",
      "parent_id": 456,
      "modified_at": 1708444800000,
      "created_at": 1708358400000,
      "size": 10485760,
      "highlighted_name": "MX40<mark>固件</mark>V2.1.bin"
    }
  ],
  "total": 142
}
```

#### GET /api/v1/app/download-url

请求参数（Query String）：
- `file_id` (number, 必填) - 文件 ID，必须为正整数
- `valid_period` (number, 可选) - 链接有效期

响应体（200）：
```json
{
  "file_id": 123,
  "download_url": "https://..."
}
```

### 管理页面接口（需 X-API-Key Header）

#### POST /api/v1/admin/sync/full

请求头：`X-API-Key: <admin_api_key>`

响应体（202 Accepted）：
```json
{
  "message": "全量同步任务已启动"
}
```

错误响应（409 Conflict）：
```json
{
  "code": "CONFLICT",
  "message": "启动同步失败",
  "request_id": "..."
}
```

#### GET /api/v1/admin/sync/full/progress

请求头：`X-API-Key: <admin_api_key>`

响应体（200）：
```json
{
  "status": "running | done | error | cancelled",
  "startedAt": 1708444800000,
  "updatedAt": 1708444900000,
  "roots": [100, 200],
  "completedRoots": [100],
  "activeRoot": 200,
  "aggregateStats": {
    "foldersVisited": 50,
    "filesIndexed": 1200,
    "pagesFetched": 60,
    "failedRequests": 2,
    "startedAt": 1708444800000,
    "endedAt": 1708444900000
  },
  "rootProgress": {
    "100": {
      "rootFolderId": 100,
      "status": "done",
      "estimatedTotalDocs": 800,
      "stats": { "foldersVisited": 20, "filesIndexed": 780, "..." : "..." },
      "updatedAt": 1708444850000
    },
    "200": {
      "rootFolderId": 200,
      "status": "running",
      "estimatedTotalDocs": 1500,
      "stats": { "foldersVisited": 30, "filesIndexed": 420, "..." : "..." },
      "updatedAt": 1708444900000
    }
  },
  "lastError": ""
}
```

错误响应（404）：未找到同步进度。

#### POST /api/v1/admin/sync/full/cancel

请求头：`X-API-Key: <admin_api_key>`

响应体（200）：
```json
{
  "message": "同步取消信号已发送"
}
```

错误响应（409 Conflict）：当前没有运行中的同步任务。

### 统一错误响应格式

所有 API 错误遵循以下结构：
```json
{
  "code": "BAD_REQUEST | UNAUTHORIZED | NOT_FOUND | CONFLICT | RATE_LIMITED | INTERNAL_ERROR",
  "message": "人类可读的中文错误描述",
  "request_id": "可选的请求追踪 ID"
}
```

## 详细设计

详细设计参见以下文档。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为驱动规范（Gherkin 场景）
- [Architecture](./architecture.md) - 系统架构、目录结构与组件设计
- [Best Practices](./best-practices.md) - React 19 / Vite 7 / Tailwind v4 / Zod 最佳实践
