# Task 002: 配置 Tailwind CSS v4 + 设计 Token

**depends-on**: task-001

## Description

安装并配置 Tailwind CSS v4（`@tailwindcss/vite` 插件），将现有 HTML 中的设计 token（颜色、字体、动画、缓动曲线）迁移到 CSS `@theme` 指令中。

## Execution Context

**Task Number**: 002 of 046
**Phase**: Setup
**Prerequisites**: Task 001 项目已初始化

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: (Infrastructure — 所有 UI 场景依赖 Tailwind 配置)

## Files to Modify/Create

- Create: `cli/src/app.css` — Tailwind 导入 + `@theme` 设计 token
- Modify: `cli/vite.config.ts` — 添加 `@tailwindcss/vite` 插件
- Modify: `cli/src/main.tsx` — 导入 `app.css`

## Steps

### Step 1: Install Tailwind CSS v4

- Install `tailwindcss` and `@tailwindcss/vite` as dev dependencies
- Add `tailwindcss()` plugin to `vite.config.ts`（在 `react()` 之后）

### Step 2: Create app.css with @theme

- `@import "tailwindcss"` 作为第一行
- 在 `@theme` 中定义以下 token：
  - `--color-paper: #f3f6fd`、`--color-paper-dark: #e9eefb`
  - `--font-display: "Outfit", system-ui, sans-serif`
  - `--font-body: "Noto Sans SC", sans-serif`
  - `--ease-smooth: cubic-bezier(0.22, 1, 0.36, 1)`
  - `--animate-soft-pulse` 动画（1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite）
- 定义 View Transition 相关的 `@layer components` 样式
- 定义 `thin-scrollbar` 工具类

### Step 3: Import app.css in main.tsx

- 在 `main.tsx` 顶部导入 `./app.css`

### Step 4: Verify

- Dev server 启动无错误
- Tailwind 工具类（如 `bg-paper`、`font-display`）可在组件中使用
- Build 产物中包含生成的 CSS

## Verification Commands

```bash
cd cli && npm run dev -- --host 0.0.0.0 &
sleep 3 && curl -s http://localhost:5173/ | grep -c "tailwind"
npm run build && ls -la ../web/dist/assets/*.css
```

## Success Criteria

- `@theme` 中的设计 token 生成对应的 Tailwind 工具类
- `bg-paper`、`font-display`、`animate-soft-pulse` 等自定义类可用
- View Transition CSS 正确包含在 `@layer components` 中
- 构建产物包含 CSS 文件
