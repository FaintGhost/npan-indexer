# Demo UI Tailwind + View Transition 计划（2026-02-21）

## Goal

- 使用 `https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4` 重构 `/demo` 页面视觉与布局。
- 使用 View Transition API 优化搜索切换与列表增量加载体验。
- 保持 end user 体验：不暴露 token 或 api key 输入。

## Constraints

- 不引入构建工具，维持单文件 HTML。
- 保持已有接口：`/api/v1/demo/search` 与 `/api/v1/demo/download-url`。
- 保留 sticky 搜索栏、无限滚动懒加载与点击下载。

## Execution Plan

- [Task 001: 红测（页面安全与入口文案约束）](./task-001-red-demo-ui-safety.md)
- [Task 002: 绿测（Tailwind 4 + View Transition UI 重构）](./task-002-green-demo-ui-tailwind-vt.md)
- [Task 003: 验证与文档更新](./task-003-verify-and-docs.md)

## Progress

- [Progress](./progress.md)
