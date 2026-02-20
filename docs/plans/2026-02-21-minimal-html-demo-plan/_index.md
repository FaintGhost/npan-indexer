# 最小 HTML 搜索下载 Demo 计划（2026-02-21）

## Goal

- 在现有 Go 服务上提供一个纯 HTML 的最小可用 Web Demo。
- 让用户可直接验证：
  - 使用 Meilisearch 本地索引搜索。
  - 选择文件后批量生成下载链接。
  - 复制下载链接或直接下载。

## Constraints

- 不引入前端构建工具，不使用 JS 框架。
- 复用现有 API：`/api/v1/search/local` 与 `/api/v1/download-url`。
- 仅做最小演示能力，不改动现有同步逻辑。

## Execution Plan

- [Task 001: 红测（Demo 路由与页面期望）](./task-001-red-demo-route-and-page.md)
- [Task 002: 绿测（实现纯 HTML Demo 与服务路由）](./task-002-green-demo-route-and-page.md)
- [Task 003: 文档与验证](./task-003-verify-docs.md)

## Progress

- [Progress](./progress.md)
