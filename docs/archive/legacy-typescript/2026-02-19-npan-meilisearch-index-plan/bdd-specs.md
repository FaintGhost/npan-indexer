# BDD 规格：云盘外部索引与下载代理

## 背景

当前云盘内置搜索效果不稳定。目标是建立一套“外部索引层”：
- 通过开放接口全量/增量同步文件与目录元数据。
- 将可搜索字段写入 Meilisearch。
- 下载时按 `file_id` 动态换取真实 `download_url`。

## 场景 1：首次全量遍历并限流入索引

**Given** 系统配置了有效用户级 token、OpenAPI 地址、Meilisearch 地址与 key  
**And** 同步起点为个人空间或指定根目录  
**When** 执行首次全量同步任务  
**Then** 任务应分页遍历所有目录与文件并写入索引  
**And** 请求速率与并发不超过配置阈值  
**And** 任务输出总处理数量与耗时统计

## 场景 2：限流/错误重试与断点续跑

**Given** 全量任务进行中出现 429 或 5xx 错误  
**When** 任务重试请求  
**Then** 应执行指数退避并带随机抖动  
**And** 达到最大重试后应记录失败并继续后续可处理分支  
**And** 任务中断后重启应从 checkpoint 继续，而不是重复全量扫描

## 场景 3：Meilisearch 文档结构与可检索性

**Given** 已抓取到文件与目录元数据  
**When** 写入 Meilisearch  
**Then** 索引文档主键应稳定（`id` + `type` 组合唯一）  
**And** `name` 与 `path_text` 可全文检索  
**And** `type`、`parent_id`、`modified_at`、`in_trash` 等字段可过滤/排序

## 场景 4：增量同步与删除同步

**Given** 已存在一次全量索引结果与 `last_sync_time`  
**When** 执行增量同步任务  
**Then** 应仅处理变更数据（新增/修改/删除）  
**And** 删除项应在索引侧正确标记或移除  
**And** 同步成功后再推进 `last_sync_time`

## 场景 5：搜索走 Meilisearch 而非平台弱检索

**Given** 索引中存在多层目录与同名文件  
**When** 用户按关键词、类型、更新时间过滤搜索  
**Then** 查询应命中 Meilisearch 并返回排序稳定的结果  
**And** 返回项应包含后续下载所需的 `file_id` 与路径信息

## 场景 6：按需获取真实下载链接

**Given** 用户选中某个 `file_id` 下载  
**When** 服务调用云盘下载接口获取 `download_url`  
**Then** 返回的链接应是最新有效的临时 URL  
**And** 不把临时 URL 持久化到搜索索引中  
**And** 下载失败时返回可诊断的错误信息
