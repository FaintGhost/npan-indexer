# End User Demo 改造计划（2026-02-21）

## Goal

- 将现有 `/demo` 改为面向终端用户的搜索下载页面。
- 页面不暴露 API Key 和 Token 输入。
- 支持 sticky 搜索框、输入即搜、无限滚动懒加载、点击直接下载。

## Constraints

- 继续使用纯 HTML/JS 单文件，不引入前端构建链。
- 认证与下载凭据全部在后端处理，前端只调用 demo 专用接口。
- 保持现有管理接口与同步能力不受影响。

## BDD Reference

- [BDD Specifications](./bdd-specs.md)

## Execution Plan

- [Task 001: 红测（demo 专用接口与页面行为）](./task-001-red-demo-enduser-api-and-ui.md)
- [Task 002: 绿测（实现后端 demo 接口）](./task-002-green-demo-enduser-api.md)
- [Task 003: 绿测（实现 sticky 搜索与无限滚动 UI）](./task-003-green-demo-enduser-ui.md)
- [Task 004: 验证与文档更新](./task-004-verify-and-docs.md)

## Progress

- [Progress](./progress.md)
