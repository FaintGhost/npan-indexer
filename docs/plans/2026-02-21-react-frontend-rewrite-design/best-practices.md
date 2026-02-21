# 最佳实践

## 1. React 19 模式

### 1.1 ref 作为 prop（不再需要 forwardRef）

React 19 中函数组件可以直接接收 `ref` 作为 prop，无需使用 `forwardRef`：

```tsx
// React 19: ref 直接作为 prop
function SearchInput({ placeholder, ref }: { placeholder: string; ref?: React.Ref<HTMLInputElement> }) {
  return <input placeholder={placeholder} ref={ref} />
}

// 使用方
const inputRef = useRef<HTMLInputElement>(null)
<SearchInput placeholder="搜索文件..." ref={inputRef} />
```

**规则**：本项目所有新组件禁止使用 `forwardRef`，直接在 props 中声明 `ref` 参数。

### 1.2 use() Hook 读取 Promise 和 Context

React 19 新增 `use()` Hook 可在渲染期间读取 Promise 或 Context 的值，配合 Suspense 使用：

```tsx
import { use, Suspense } from "react"

// 读取 Context（可在条件语句中调用，不像 useContext）
function AdminPanel() {
  const isAuth = checkAuth()
  if (!isAuth) {
    return <AuthDialog />
  }
  const apiKey = use(ApiKeyContext)
  return <SyncManager apiKey={apiKey} />
}
```

**注意**：`use()` 的 Promise 必须来自 Suspense 兼容的数据源，不要在组件内部创建新的 Promise 传给 `use()`。

### 1.3 useActionState 管理异步操作

使用 `useActionState` 替代手动管理 loading/error/data 三元组：

```tsx
import { useActionState, startTransition } from "react"

function SyncControls() {
  const [state, dispatch, isPending] = useActionState(
    async (prevState, action: "start" | "cancel") => {
      if (action === "start") {
        const result = await startFullSync()
        return { ...prevState, message: result.message, error: null }
      }
      // ...
    },
    { message: null, error: null }
  )

  return (
    <button
      disabled={isPending}
      onClick={() => startTransition(() => dispatch("start"))}
    >
      {isPending ? "启动中..." : "启动全量同步"}
    </button>
  )
}
```

### 1.4 useOptimistic 乐观更新

在取消同步等操作中使用 `useOptimistic` 立即反馈 UI：

```tsx
import { useOptimistic } from "react"

function CancelButton({ currentStatus, onCancel }) {
  const [optimisticStatus, setOptimisticStatus] = useOptimistic(currentStatus)

  async function handleCancel() {
    setOptimisticStatus("cancelling")
    await onCancel()
  }

  return (
    <button onClick={handleCancel} disabled={optimisticStatus === "cancelling"}>
      {optimisticStatus === "cancelling" ? "取消中..." : "取消同步"}
    </button>
  )
}
```

### 1.5 组件设计原则

- **单一职责**：每个组件只负责一件事。搜索框、文件卡片、下载按钮分别是独立组件。
- **组合优于继承**：使用 children 和 render props 模式组合组件。
- **Props 类型定义**：所有组件 props 使用 TypeScript interface 定义，从 Zod schema 推导的类型优先。
- **自定义 Hook 提取逻辑**：将副作用逻辑（搜索请求、debounce、下载链接获取）提取为自定义 Hook。

推荐的 Hook 拆分：

| Hook | 职责 |
|------|------|
| `useSearch` | 搜索状态管理、debounce、分页、请求竞态处理 |
| `useDownload` | 下载链接获取、缓存、按钮状态管理 |
| `useViewMode` | Hero/Docked 模式切换、View Transition 封装 |
| `useHotkey` | 全局键盘快捷键注册与清理 |
| `useAdminAuth` | API Key 的 localStorage 读写、验证、401 拦截 |
| `useSyncProgress` | 同步进度轮询、自动启停 |
| `useIntersectionObserver` | 通用 IntersectionObserver Hook |

## 2. Tailwind CSS v4 模式

### 2.1 CSS-first 配置（@theme 指令）

Tailwind CSS v4 使用 CSS-first 配置取代 `tailwind.config.js`。所有自定义 token 在 CSS 文件中通过 `@theme` 定义：

```css
/* app.css */
@import "tailwindcss";

@theme {
  /* 品牌色 */
  --color-paper: #f3f6fd;
  --color-paper-dark: #e9eefb;

  /* 字体 */
  --font-display: "Outfit", system-ui, sans-serif;
  --font-body: "Noto Sans SC", sans-serif;

  /* 自定义缓动曲线 */
  --ease-smooth: cubic-bezier(0.22, 1, 0.36, 1);

  /* 自定义动画 */
  --animate-soft-pulse: soft-pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;

  @keyframes soft-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
}
```

**规则**：
- 不创建 `tailwind.config.js` 或 `tailwind.config.ts`
- 所有设计 token（颜色、字体、间距、动画）在 CSS `@theme` 中定义
- 使用 CSS 变量命名空间自动生成工具类（如 `--color-paper` → `bg-paper`）

### 2.2 保持与现有设计一致

从现有 HTML 中提取的设计 token 应完整迁移到 `@theme`：

| 现有定义 | CSS 变量 | 生成的工具类 |
|---------|---------|------------|
| `--paper: #f3f6fd` | `--color-paper: #f3f6fd` | `bg-paper`, `text-paper` |
| `--paper-dark: #e9eefb` | `--color-paper-dark: #e9eefb` | `bg-paper-dark` |
| `font-family: "Outfit"` | `--font-display: "Outfit", system-ui, sans-serif` | `font-display` |
| `font-family: "Noto Sans SC"` | `--font-body: "Noto Sans SC", sans-serif` | `font-body` |

### 2.3 层叠与组件样式

对于需要复杂选择器的样式（如 View Transition），使用 `@layer` 组织：

```css
@layer components {
  .search-stage {
    view-transition-name: search-stage;
  }

  .search-card {
    view-transition-name: search-card;
  }
}

@layer utilities {
  .thin-scrollbar {
    scrollbar-width: thin;
  }
}
```

### 2.4 Vite 集成

Tailwind CSS v4 通过 `@tailwindcss/vite` 插件集成，非 PostCSS 插件：

```ts
// vite.config.ts
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [
    tanstackRouter({ target: "react", autoCodeSplitting: true }),
    react(),
    tailwindcss(),
  ],
})
```

## 3. Zod Schema 设计

### 3.1 API 响应 Schema

为每个 API 端点定义对应的 Zod schema，类型从 schema 自动推导：

```ts
import { z } from "zod"

// IndexDocument schema（匹配 Go 后端 models.IndexDocument）
export const IndexDocumentSchema = z.object({
  doc_id: z.string(),
  source_id: z.number(),
  type: z.enum(["file", "folder"]),
  name: z.string(),
  path_text: z.string(),
  parent_id: z.number(),
  modified_at: z.number(),
  created_at: z.number(),
  size: z.number(),
  highlighted_name: z.string().optional().default(""),
})

// 搜索响应
export const SearchResponseSchema = z.object({
  items: z.array(IndexDocumentSchema),
  total: z.number(),
})

// 下载链接响应
export const DownloadURLResponseSchema = z.object({
  file_id: z.number(),
  download_url: z.string().min(1),
})

// 类型从 schema 推导
export type IndexDocument = z.infer<typeof IndexDocumentSchema>
export type SearchResponse = z.infer<typeof SearchResponseSchema>
export type DownloadURLResponse = z.infer<typeof DownloadURLResponseSchema>
```

### 3.2 同步进度 Schema

```ts
const CrawlStatsSchema = z.object({
  foldersVisited: z.number(),
  filesIndexed: z.number(),
  pagesFetched: z.number(),
  failedRequests: z.number(),
  startedAt: z.number(),
  endedAt: z.number(),
})

const RootProgressSchema = z.object({
  rootFolderId: z.number(),
  status: z.string(),
  estimatedTotalDocs: z.number().nullable().optional(),
  stats: CrawlStatsSchema,
  updatedAt: z.number(),
})

export const SyncProgressSchema = z.object({
  status: z.enum(["running", "done", "error", "cancelled"]),
  startedAt: z.number(),
  updatedAt: z.number(),
  roots: z.array(z.number()),
  completedRoots: z.array(z.number()),
  activeRoot: z.number().nullable().optional(),
  aggregateStats: CrawlStatsSchema,
  rootProgress: z.record(z.string(), RootProgressSchema),
  lastError: z.string().optional().default(""),
})
```

### 3.3 错误响应 Schema

```ts
export const ErrorResponseSchema = z.object({
  code: z.string(),
  message: z.string(),
  request_id: z.string().optional(),
})
```

### 3.4 Schema 使用规则

- **所有 API 调用**必须经过对应 schema 的 `.parse()` 或 `.safeParse()`
- 使用 `safeParse()` 在 API 层，将 ZodError 转换为用户友好的错误消息
- 类型一律从 schema 推导（`z.infer<>`），禁止手动定义重复的 TypeScript interface
- Schema 文件统一放在 `src/schemas/` 目录

### 3.5 API 客户端封装

```ts
async function fetchAPI<T>(
  url: string,
  schema: z.ZodSchema<T>,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(url, options)
  if (!response.ok) {
    const error = ErrorResponseSchema.safeParse(await response.json().catch(() => ({})))
    throw new ApiError(
      response.status,
      error.success ? error.data.message : `HTTP ${response.status}`
    )
  }
  const data = await response.json()
  return schema.parse(data)
}
```

## 4. 可访问性（Accessibility）

### 4.1 语义化 HTML

| 区域 | 元素 | 说明 |
|------|------|------|
| 搜索栏 | `<header>` + `<search>` | HTML5 搜索地标 |
| 搜索输入 | `<input type="search">` | 触发搜索 UI 提示 |
| 结果列表 | `<main>` + `<section>` | 主内容区域 |
| 文件卡片 | `<article>` | 独立内容单元 |
| 结果计数 | `<output>` | 计算结果关联 |

### 4.2 ARIA 属性

```tsx
// 搜索结果区域 — live region
<section aria-live="polite" aria-busy={isLoading} aria-label="搜索结果">

// 搜索输入框
<input
  type="search"
  role="searchbox"
  aria-label="搜索文件"
  aria-describedby="search-status"
/>

// 状态文本
<p id="search-status" role="status" aria-live="polite">
  已加载 3 / 50 个文件
</p>

// 下载按钮 — 描述性标签
<button aria-label={`下载 ${fileName}`} disabled={isDownloading}>
  {/* ... */}
</button>

// 骨架屏
<div aria-hidden="true" role="presentation">
  {/* skeleton cards */}
</div>
```

### 4.3 焦点管理

- 模式切换（Hero → Docked）后不自动移动焦点，保持在搜索输入框
- 清空按钮点击后焦点回到搜索输入框
- 模态对话框（API Key 输入、确认取消同步）打开时焦点陷入对话框内（focus trap）
- 对话框关闭后焦点恢复到触发元素

### 4.4 键盘导航

- 所有交互元素可通过 Tab 键访问
- 对话框支持 Escape 键关闭
- 下载按钮支持 Enter 和 Space 触发
- 文件卡片列表不需要 arrow key 导航（非 listbox 模式）

### 4.5 颜色与对比度

保持现有设计中的颜色对比度。关键检查点：

| 文字 | 背景 | 对比度要求 |
|------|------|---------|
| `text-slate-800` 正文 | `bg-white` 卡片 | >= 4.5:1 |
| `text-slate-500` 辅助文字 | `bg-white` 卡片 | >= 4.5:1 |
| `text-white` 按钮文字 | `bg-blue-600` 按钮 | >= 4.5:1 |
| `text-white` 按钮文字 | `bg-slate-900` 按钮 | >= 4.5:1 |
| `text-rose-600` 错误文字 | `bg-white` / `bg-paper` | >= 4.5:1 |

## 5. 性能优化

### 5.1 代码分割与懒加载

管理页面使用 TanStack Router 的 `autoCodeSplitting` 自动进行代码分割：

```ts
// vite.config.ts
tanstackRouter({
  target: "react",
  autoCodeSplitting: true,  // 自动按路由拆分
})
```

TanStack Router 配合 `autoCodeSplitting` 会自动将每个路由的 `component`、`loader` 等拆分为独立 chunk。

### 5.2 搜索性能

- **debounce 280ms**：与现有实现一致，避免过度请求
- **AbortController**：新搜索取消旧请求，避免过时结果覆盖
- **requestSeq 序列号**：双重保护，确保响应与当前搜索匹配
- **IntersectionObserver rootMargin**：`180px 0px` 预加载，用户滚动到底部前已开始请求

### 5.3 渲染优化

- **React.memo**：对 `FileCard` 组件使用 memo，避免列表重新搜索时所有卡片重渲染
- **key 策略**：使用 `source_id` 作为列表 key，确保列表更新时 DOM 复用
- **虚拟化**（可选未来优化）：当结果超过 200 条时考虑引入 `@tanstack/react-virtual`

### 5.4 网络优化

- **下载链接缓存**：使用 `Map<number, string>` 在组件/Hook 中缓存已获取的下载 URL
- **同步进度轮询间隔**：3 秒一次，使用 `setInterval` + `useEffect` cleanup
- **请求去重**：加载下一页时如果上一次请求未完成则跳过

### 5.5 构建产物优化

```ts
// vite.config.ts build 配置
export default defineConfig({
  build: {
    outDir: "../web/dist",  // Go embed.FS 读取目录
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ["react", "react-dom"],
          router: ["@tanstack/react-router"],
        },
      },
    },
  },
})
```

### 5.6 字体加载

保留现有字体加载策略（Google Fonts CDN + `display=swap`），在 `index.html` 中预连接：

```html
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
```

## 6. 项目结构

### 6.1 推荐目录结构

```
cli/                          # 前端项目根目录（Vite 项目）
├── index.html                # SPA 入口
├── vite.config.ts            # Vite 配置
├── tsconfig.json             # TypeScript 配置（strict: true）
├── package.json
├── src/
│   ├── app.css               # Tailwind @theme 全局样式
│   ├── main.tsx              # React 入口，挂载 RouterProvider
│   ├── router.tsx            # TanStack Router 配置
│   ├── routes/               # 文件路由
│   │   ├── __root.tsx        # 根布局
│   │   ├── index.tsx         # /app 搜索页
│   │   └── admin.tsx         # /app/admin 管理页（懒加载）
│   ├── components/           # UI 组件
│   │   ├── search-input.tsx
│   │   ├── file-card.tsx
│   │   ├── download-button.tsx
│   │   ├── skeleton-card.tsx
│   │   ├── empty-state.tsx
│   │   ├── sync-progress.tsx
│   │   └── api-key-dialog.tsx
│   ├── hooks/                # 自定义 Hooks
│   │   ├── use-search.ts
│   │   ├── use-download.ts
│   │   ├── use-view-mode.ts
│   │   ├── use-hotkey.ts
│   │   ├── use-admin-auth.ts
│   │   ├── use-sync-progress.ts
│   │   └── use-intersection-observer.ts
│   ├── schemas/              # Zod schemas
│   │   ├── search.ts
│   │   ├── download.ts
│   │   ├── sync.ts
│   │   └── error.ts
│   ├── lib/                  # 工具函数
│   │   ├── api-client.ts     # fetchAPI 封装
│   │   ├── format.ts         # formatBytes, formatTime
│   │   └── file-icon.ts     # 文件扩展名 → 图标映射
│   └── types/                # 共享类型（仅放不属于 schema 的类型）
│       └── index.ts
└── web/
    └── dist/                 # Vite 构建输出（Go embed.FS 读取）
```

### 6.2 命名规范

| 类别 | 规范 | 示例 |
|------|------|------|
| 文件名 | kebab-case | `file-card.tsx`, `use-search.ts` |
| 组件 | PascalCase | `FileCard`, `SearchInput` |
| Hook | camelCase，use 前缀 | `useSearch`, `useDownload` |
| Schema | PascalCase + Schema 后缀 | `SearchResponseSchema` |
| 类型 | PascalCase | `IndexDocument`（从 schema 推导） |
| CSS 变量 | kebab-case | `--color-paper`, `--font-display` |

### 6.3 导入规范

- 使用 Vite path alias `@/` 映射 `src/`
- 导入顺序：React → 第三方库 → 内部模块 → 类型
- 使用 `type` 关键字导入纯类型：`import type { IndexDocument } from "@/schemas/search"`

## 7. TanStack Router 配置

### 7.1 Vite 插件配置

```ts
// vite.config.ts
import { tanstackRouter } from "@tanstack/router-plugin/vite"

export default defineConfig({
  plugins: [
    // 必须在 react() 之前
    tanstackRouter({
      target: "react",
      autoCodeSplitting: true,
    }),
    react(),
    tailwindcss(),
  ],
})
```

### 7.2 文件路由定义

```tsx
// src/routes/__root.tsx
import { createRootRoute, Outlet, Link } from "@tanstack/react-router"

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  return (
    <>
      <nav>{/* 导航链接 */}</nav>
      <Outlet />
    </>
  )
}
```

```tsx
// src/routes/index.tsx
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/")({
  component: SearchPage,
})
```

### 7.3 搜索参数类型安全

TanStack Router 支持类型安全的 search params：

```tsx
// src/routes/index.tsx
import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

const searchParamsSchema = z.object({
  query: z.string().optional().default(""),
})

export const Route = createFileRoute("/")({
  validateSearch: searchParamsSchema,
  component: SearchPage,
})

function SearchPage() {
  const { query } = Route.useSearch()
  // query 有类型推导，为 string
}
```

### 7.4 basePath 配置

由于应用挂载在 `/app` 路径下：

```tsx
// src/router.tsx
import { createRouter } from "@tanstack/react-router"
import { routeTree } from "./routeTree.gen"

export const router = createRouter({
  routeTree,
  basepath: "/app",
})
```

## 8. 测试策略

### 8.1 单元测试范围

| 测试目标 | 测试工具 | 关注点 |
|---------|---------|--------|
| Zod Schema | Vitest | 有效/无效数据解析、边界值、默认值 |
| 工具函数 | Vitest | `formatBytes`, `formatTime`, `getFileIcon` 等纯函数 |
| 自定义 Hooks | Vitest + React Testing Library | Hook 状态变化、副作用触发、清理逻辑 |
| 组件渲染 | Vitest + React Testing Library | 条件渲染、props 映射、事件处理 |

### 8.2 API Mock 策略

使用 MSW（Mock Service Worker）模拟后端 API：

```ts
// tests/mocks/handlers.ts
import { http, HttpResponse } from "msw"

export const handlers = [
  http.get("/api/v1/app/search", ({ request }) => {
    const url = new URL(request.url)
    const query = url.searchParams.get("query")
    if (!query) {
      return HttpResponse.json(
        { code: "BAD_REQUEST", message: "缺少 query 参数" },
        { status: 400 }
      )
    }
    return HttpResponse.json({ items: mockItems, total: mockItems.length })
  }),
]
```

### 8.3 测试文件组织

测试文件与源文件同目录，使用 `.test.ts` / `.test.tsx` 后缀：

```
src/
├── hooks/
│   ├── use-search.ts
│   └── use-search.test.ts
├── schemas/
│   ├── search.ts
│   └── search.test.ts
└── components/
    ├── file-card.tsx
    └── file-card.test.tsx
```
