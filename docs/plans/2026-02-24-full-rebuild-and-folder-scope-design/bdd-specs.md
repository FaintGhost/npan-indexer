# BDD Specifications

## Feature: Force Rebuild Must Reset Crawl Checkpoints

### Scenario: 强制重建忽略残留 checkpoint

```gherkin
Given 存在某个 root 的 checkpoint 文件，内容指向中间页
And 用户发起 mode=full 且 force_rebuild=true 的同步
When 同步任务开始执行该 root
Then 程序应先清理该 root 的 checkpoint 文件
And RunFullCrawl 应从 page 0 开始抓取
And 最终统计不应仅覆盖 checkpoint 之后的尾部数据
```

### Scenario: 非强制重建且 resume=false 也应重置 checkpoint

```gherkin
Given 存在某个 root 的 checkpoint 文件
And 用户发起全量同步且 resume_progress=false
When 同步任务启动
Then 程序应清理该 root 的 checkpoint 文件
And 不应从旧断点续跑
```

### Scenario: resume=true 保留当前断点行为

```gherkin
Given 存在某个 root 的 checkpoint 文件
And 用户发起全量同步且 resume_progress=true
When 同步任务启动
Then 程序不应清理 checkpoint
And 应继续沿用断点续跑
```

## Feature: Folder-Scoped Full Sync From Admin UI

### Scenario: 指定单个目录 ID 启动范围索引

```gherkin
Given 用户在 Admin 页面输入目录 ID "123456"
And 选择 mode=full
When 用户点击启动同步
Then 前端请求体应包含 root_folder_ids=[123456]
And 前端请求体应包含 include_departments=false
And 后端应仅以该目录作为根目录执行全量索引
```

### Scenario: 输入多个目录 ID（逗号分隔）

```gherkin
Given 用户输入 "1001, 1002,1003"
When 用户启动同步
Then 前端应解析为 [1001,1002,1003]
And 请求体中的 root_folder_ids 顺序应稳定
```

### Scenario: 空输入保持现有全库行为

```gherkin
Given 用户未填写目录范围
When 用户启动全量同步
Then 前端请求体中的 root_folder_ids 应为空数组
And 不改变当前默认 root 发现行为
```

## Feature: Official Folder Stats Estimate

### Scenario: 显式 root 使用 folder info 填充估计值

```gherkin
Given 用户指定 root_folder_ids=[123456]
And Npan /api/v2/folder/{id}/info 返回 item_count=4151, name="PIXELHUE"
When SyncManager.discoverRootFolders 执行
Then RootNames[123456] 应为 "PIXELHUE"
And RootProgress[123456].estimatedTotalDocs 应为 4152
```

### Scenario: folder info 失败不阻断同步（降级）

```gherkin
Given 用户指定 root_folder_ids=[123456]
And 获取 folder info 失败（例如 404/403/网络错误）
When 同步启动
Then 同步仍应继续
And 该 root 的 estimatedTotalDocs 可为空
And 进度中可记录告警或日志
```

## Feature: Completion Warning For Estimate Mismatch

### Scenario: 实际写入显著小于官方估计时生成警告

```gherkin
Given 某 root estimatedTotalDocs=4152
And 该 root 完成后 actual(filesIndexed + foldersVisited)=512
When 全量同步完成并生成 verification
Then verification.warnings 应包含该 root 的差异信息
And UI 应显示验证警告
```

### Scenario: 差异在阈值内不告警

```gherkin
Given 某 root estimatedTotalDocs=4152
And actual=4148
When 全量同步完成
Then verification.warnings 不应因该 root 产生告警
```

## Suggested Automated Tests

### Go Unit Tests

- `internal/service/sync_manager_routing_test.go`
  - `TestRun_ForceRebuildClearsCheckpointBeforeFullCrawl`
  - `TestRun_FullResumeFalseClearsCheckpoint`
  - `TestRun_FullResumeTrueKeepsCheckpoint`
- `internal/service/sync_manager_estimate_test.go`
  - `TestDiscoverRootFolders_ExplicitRootsFetchFolderInfoEstimate`
  - `TestDiscoverRootFolders_FolderInfoFailureDegradesGracefully`
- `internal/service/sync_manager_verification_test.go`
  - `TestBuildVerification_AppendsRootEstimateMismatchWarnings`
- `internal/npan/client_*_test.go`
  - `TestGetFolderInfo_MapsNameAndItemCount`

### Frontend Tests (Vitest)

- `web/src/hooks/use-sync-progress.test.ts`
  - 指定目录时请求体包含 `root_folder_ids` 和 `include_departments=false`
- `web/src/components/admin-page.test.tsx`
  - 目录 ID 输入解析、非法输入校验、空输入兼容
- `web/src/components/sync-progress-display.test.tsx`
  - 显示 estimate/actual 差异告警
