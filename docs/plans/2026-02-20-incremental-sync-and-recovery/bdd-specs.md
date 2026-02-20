# BDD 规格

## Feature: 增量索引链路

Scenario: 基于游标拉取变更并更新状态
  Given 已存在上次增量时间 `lastSyncTime`
  When 执行一次增量同步
  Then 系统应仅拉取该时间之后的变更
  And 变更应拆分为 upsert 与 delete
  And 成功后写回新的 `lastSyncTime`

Scenario: 写入失败时不得推进游标
  Given 拉取到变更
  When upsert 或 delete 任一环节失败
  Then 增量同步返回错误
  And `lastSyncTime` 保持不变

Scenario: 历史毫秒游标可兼容迁移
  Given `lastSyncTime` 为历史毫秒值
  When 执行一次增量同步
  Then 系统应将游标按秒语义进行兼容处理
  And 同步成功后写回秒级游标

Scenario: 默认增量查询词应可返回变更集合
  Given 增量查询词默认值已配置
  When 执行一次增量同步窗口查询
  Then 不应使用会稳定返回空结果的 `*` 作为默认值

## Feature: 全量中断后的恢复可观测性

Scenario: CLI 收到中断信号
  Given 全量同步正在执行
  When 进程收到 SIGINT
  Then CLI 发送取消信号并等待同步协程结束
  And 输出最终进度/状态摘要
