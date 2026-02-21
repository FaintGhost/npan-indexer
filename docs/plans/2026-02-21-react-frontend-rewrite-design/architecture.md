# Architecture

## 项目目录结构

```
web/
  package.json
  pnpm-lock.yaml
  tsconfig.json                 # TypeScript 主配置（Project References）
  tsconfig.app.json             # 浏览器端 TS 配置（src/）
  tsconfig.node.json            # Node 端 TS 配置（vite.config.ts）
  vite.config.ts                # Vite 配置：插件注册 + proxy
  .oxlintrc.json                # Oxlint 规则配置
  .oxfmtrc.json                 # Oxfmt 格式化配置
  index.html                    # Vite 入口 HTML（SPA 壳）
  src/
    main.tsx                    # 应用入口：createRouter + ReactDOM.createRoot
    app.css                     # 全局样式：@import "tailwindcss" + @theme + 自定义动画
    routeTree.gen.ts            # TanStack Router 自动生成（git 忽略）
    routes/
      __root.tsx                # 根路由：全局 layout + ErrorBoundary
      index.tsx                 # / 重定向到 /app
      app.tsx                   # /app 布局：SearchStage sticky header + Outlet
      app.index.lazy.tsx        # /app 搜索主页面（懒加载）
      _admin.tsx                # admin pathless layout：API key 验证层
      _admin.sync.lazy.tsx      # /sync 同步管理页面（懒加载）
    components/
      SearchCard.tsx            # 搜索卡片：输入框 + 按钮 + 快捷键提示
      ResultList.tsx            # 结果列表 + IntersectionObserver 无限加载
      ResultItem.tsx            # 单条结果卡片：图标 + 名称高亮 + 元信息 + 下载
      FileIcon.tsx              # 文件类型图标映射（扩展名 -> 颜色 + SVG）
      DownloadButton.tsx        # 下载按钮：4 状态（default/loading/success/error）
      EmptyState.tsx            # 空状态组件（initial/no-results/error 三种 variant）
      SkeletonCard.tsx          # 加载骨架屏
    lib/
      api-client.ts             # fetch 封装：baseURL、query 序列化、错误处理
      api.ts                    # API 函数：appSearch、appDownloadURL + admin 系列
      schemas.ts                # Zod schema：SearchResponse、DownloadURL、SyncProgress 等
    hooks/
      use-search.ts             # 搜索状态 hook：query/page/items/loading/hasMore + debounce + abort
      use-download.ts           # 下载链接 hook：带缓存 + 按钮状态机
      use-keyboard-shortcut.ts  # 全局快捷键 hook（Cmd/Ctrl+K）
      use-view-transition.ts    # View Transition API 封装
    utils/
      format.ts                 # formatBytes、formatTime、escapeHTML
  dist/                         # Vite 构建输出（git 忽略，Go embed 目标）
```

## 组件层级

```
__root.tsx                          <- 全局 ErrorBoundary + 字体 + 背景渐变
  +-- app.tsx                       <- /app 布局：SearchStage (sticky) + main
  |   +-- app.index.lazy.tsx        <- 搜索主页面
  |       +-- SearchCard            <- 搜索输入区
  |       +-- ResultList            <- 结果列表 + 无限滚动
  |       |   +-- ResultItem x N    <- 单条结果
  |       |   |   +-- FileIcon      <- 文件类型图标
  |       |   |   +-- DownloadButton<- 下载按钮
  |       |   +-- SkeletonCard x N  <- 加载占位
  |       +-- EmptyState            <- 初始/无结果/错误 占位
  +-- _admin.tsx                    <- admin pathless layout（API key 验证）
      +-- _admin.sync.lazy.tsx      <- 同步管理页面
```

## API 层设计

### fetch 封装（lib/api-client.ts）

- `apiGet<T>(path, params, schema, signal?)` 泛型函数
- 自动序列化 params 为 URL query（过滤 undefined/null/空字符串）
- `window.location.origin` 作为 baseURL（dev 由 Vite proxy，prod 同源）
- HTTP 非 2xx 时解析 ErrorResponse 并抛出类型化错误
- 入参 Zod schema 对响应执行 `schema.parse(data)`

### Zod Schema（lib/schemas.ts）

```
SearchResponseSchema:
  items: z.array(SearchItemSchema)
  total: z.number()

SearchItemSchema:
  doc_id, source_id, type, name, path_text,
  parent_id, modified_at, created_at, size,
  highlighted_name (optional)

DownloadURLResponseSchema:
  file_id: z.number()
  download_url: z.string()

SyncProgressSchema:
  status, startedAt, updatedAt, roots, completedRoots,
  activeRoot, aggregateStats, rootProgress, lastError

ErrorResponseSchema:
  code: z.string()
  message: z.string()
  request_id: z.string().optional()
```

### API 函数（lib/api.ts）

- `appSearch(params, signal?)` -> SearchResponse
- `appDownloadURL(params, signal?)` -> DownloadURLResponse
- `adminStartSync(apiKey, body)` -> { message: string }
- `adminGetProgress(apiKey)` -> SyncProgressState
- `adminCancelSync(apiKey)` -> { message: string }

Admin 函数额外接收 `apiKey` 参数，通过 `X-API-Key` header 传递。

## Go embed 集成

### 新增文件：web/embed.go

```go
package web

import "embed"

//go:embed dist/*
var DistFS embed.FS
```

### 修改 internal/httpx/server.go

1. 删除 `resolveAppHTMLPath()` 函数
2. `NewServer()` 新增 `distFS fs.FS` 参数
3. 使用 `fs.Sub(distFS, "dist")` 获取子文件系统
4. SPA fallback handler：
   - `/app` 及子路径先查找精确文件（JS/CSS/图片等静态资源）
   - 文件存在则通过 http.FileServer 提供
   - 不存在则返回 index.html（SPA history fallback）
   - API 路由不受影响
5. Cache-Control：
   - `assets/*`（带 hash）：`max-age=31536000, immutable`
   - `index.html`：`no-cache`

### 修改 cmd/server/main.go

```go
import "npan/web"
// 传入 web.DistFS 到 httpx.NewServer()
```

### 路由挂载

```
/healthz                              -> handlers.Health
/readyz                               -> handlers.Readyz
/app, /app/*                          -> SPA handler (embed FS + index.html fallback)
/api/v1/app/search                    -> handlers.AppSearch (EmbeddedAuth)
/api/v1/app/download-url              -> handlers.AppDownloadURL (EmbeddedAuth)
/api/v1/admin/sync/full               -> handlers.StartFullSync (APIKeyAuth)
/api/v1/admin/sync/full/progress      -> handlers.GetFullSyncProgress (APIKeyAuth)
/api/v1/admin/sync/full/cancel        -> handlers.CancelFullSync (APIKeyAuth)
```

## 构建流水线

### package.json scripts

```
"dev"        -> "vite"                    # 开发服务器 + HMR
"build"      -> "tsc -b && vite build"    # 类型检查 + 生产构建
"preview"    -> "vite preview"            # 预览生产构建
"lint"       -> "oxlint"                  # Lint
"fmt"        -> "oxfmt --write ."         # 格式化
"fmt:check"  -> "oxfmt --check ."         # CI 格式检查
"typecheck"  -> "tsc -b --noEmit"         # 仅类型检查
```

### 开发模式

vite.config.ts proxy 配置：

```ts
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:1323',
      changeOrigin: true,
    },
  },
}
```

工作流：
1. 终端 A：`go run ./cmd/server`（:1323）
2. 终端 B：`cd web && pnpm dev`（:5173）
3. 浏览器访问 `http://localhost:5173/app`
4. `/api/*` 请求由 Vite proxy 转发

### 生产构建

```bash
cd web && pnpm build          # 输出到 web/dist/
cd .. && go build ./cmd/server # embed web/dist/ 到二进制
```

### Docker 多阶段构建

```
Stage 1 (frontend): node:22-alpine -> pnpm install -> pnpm build -> web/dist/
Stage 2 (backend):  golang:1.25-alpine -> COPY --from=frontend -> go build
Stage 3 (runtime):  alpine:3.21 -> 仅复制 Go 二进制
```

## TanStack Router 配置

### Vite 插件顺序

```ts
plugins: [
  tanstackRouter({
    target: 'react',
    autoCodeSplitting: true,
    routesDirectory: './src/routes',
    generatedRouteTree: './src/routeTree.gen.ts',
  }),
  tailwindcss(),
  react(),
]
```

TanStack Router 插件必须在 React 插件之前。

### 路由文件映射

| 文件 | 路径 | 用途 |
|------|------|------|
| `__root.tsx` | (全局) | createRootRoute，全局 layout + ErrorBoundary |
| `index.tsx` | `/` | 重定向到 `/app` |
| `app.tsx` | `/app` | 搜索 layout：sticky header + Outlet |
| `app.index.lazy.tsx` | `/app` (index) | 搜索主页面（懒加载） |
| `_admin.tsx` | (pathless) | admin API key 验证层 |
| `_admin.sync.lazy.tsx` | `/sync` | 同步管理页面（懒加载） |

### 自动代码分割

`autoCodeSplitting: true` 自动将路由拆分：
- Critical（路由配置、loader、search params）-> 主 bundle
- Non-critical（component、pendingComponent、errorComponent）-> 按需懒加载

## Oxlint + Oxfmt 配置

### .oxlintrc.json

- plugins: typescript, react, unicorn, import
- correctness: error, suspicious: warn
- 关键规则：no-floating-promises, exhaustive-deps, no-console
- ignorePatterns: dist/**, routeTree.gen.ts

### .oxfmtrc.json

- printWidth: 100, tabWidth: 2, semi: false, singleQuote: true
- ignorePatterns: dist/**, routeTree.gen.ts

## 依赖清单

### 运行时

```
react                  ^19.0.0
react-dom              ^19.0.0
@tanstack/react-router ^2.x
zod                    ^3.x
```

### 开发时

```
vite                   ^7.x
@vitejs/plugin-react   ^4.x
@tanstack/router-plugin ^2.x
tailwindcss            ^4.x
@tailwindcss/vite      ^4.x
typescript             ^5.x
@types/react           latest
@types/react-dom       latest
oxlint                 latest
oxfmt                  latest
tsgolint               latest
```

## 迁移映射

| 现有 index.html | React 目标 |
|---|---|
| `<style>` CSS 变量、动画 | `src/app.css`（@theme + 自定义 CSS） |
| `UIStates` HTML 字符串 | `EmptyState.tsx`（variant prop） |
| `state` 对象 | `use-search.ts` hook |
| `requestJSON()` | `lib/api-client.ts` |
| `fetchPage()` / `triggerSearch()` | `use-search.ts` 内部逻辑 |
| `getDownloadURL()` + `linkCache` | `use-download.ts`（Map 缓存） |
| `cardHTML()` | `ResultItem.tsx` JSX |
| `renderSkeleton()` | `SkeletonCard.tsx` |
| `BtnStates` | `DownloadButton.tsx`（状态机） |
| IntersectionObserver | `ResultList.tsx` useEffect |
| `startViewTransition()` | `use-view-transition.ts` hook |
| `formatTime()`/`formatBytes()`/`getFileIcon()` | `utils/format.ts` + `FileIcon.tsx` |
| Cmd/Ctrl+K | `use-keyboard-shortcut.ts` |
