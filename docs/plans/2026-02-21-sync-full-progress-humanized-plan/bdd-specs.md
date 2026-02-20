# BDD 规格

## Feature: sync-full 进度输出可读性

Scenario: 默认输出为人类可读摘要
  Given 用户执行 `sync-full` 且未指定进度输出模式
  When CLI 轮询到同步进度
  Then 输出应为单行摘要文本
  And 包含状态、已完成根目录数量、活动根目录、累计统计与耗时

Scenario: 支持 JSON 进度模式
  Given 用户执行 `sync-full --progress-output json`
  When CLI 轮询到同步进度
  Then 输出应保持结构化 JSON 进度摘要

Scenario: 中断时输出友好摘要
  Given 同步任务正在运行
  When 进程收到 SIGINT 或 SIGTERM
  Then CLI 在取消并等待停止后输出最终摘要
  And 返回中断错误语义不变
