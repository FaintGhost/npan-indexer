# Demo 动态搜索卡住修复计划（2026-02-21）

## Goal

- 修复 demo 页面“输入后偶发无结果，需手动点击搜索”的问题。
- 保持现有交互（防抖输入、按钮搜索、无限滚动）不回退。

## Root Cause

- 新搜索触发时会重置 `requestSeq`，但若旧请求尚未结束，`state.loading` 会阻止新请求发出。
- 旧请求返回后因 `seq` 不匹配被丢弃，导致本轮搜索没有任何请求真正生效。

## Execution Plan

- [Task 001] 增加前端 in-flight 请求中止机制（AbortController）。
- [Task 002] 调整 replace 模式下的并发门控，确保新查询可抢占旧查询。
- [Task 003] 验证动态输入、回车、点击搜索与无限滚动路径。

## Progress

- [Progress](./progress.md)
