# BDD Specifications: 同步状态 SQLite 迁移

## Feature: SQLite 成为同步状态的主状态源

### Scenario 1: 全量同步成功后进度与增量游标写入 SQLite
```gherkin
Given SyncManager 使用 SQLite progress store、sync state store 与 checkpoint store factory
And 一次全量同步成功完成
When 管理端或 CLI 读取同步状态
Then 应能从 SQLite 读取到 status=done 的 SyncProgressState
And 应能从 SQLite 读取到 LastSyncTime 大于 0 的 SyncState
And 进度中的根目录统计、verification 与 completedRoots 应保持与迁移前一致
```

### Scenario 2: 全量同步失败时不会错误推进增量游标
```gherkin
Given SyncManager 使用 SQLite 状态存储
And 一次全量同步在 crawl 过程中失败
When 读取 SQLite 中的 sync state
Then LastSyncTime 不应被写成新的成功时间点
And SQLite 中的 SyncProgressState 应反映 error 状态与失败原因
```

## Feature: checkpoint 在 SQLite 中支持恢复与清理语义

### Scenario 3: resume=true 时应从 SQLite checkpoint 恢复 crawl 队列
```gherkin
Given 某个根目录在 SQLite 中已有未完成的 CrawlCheckpoint
And 用户以 resume_progress=true 启动全量同步
When SyncManager 启动该根目录的 crawl
Then crawler 应从已有 checkpoint 队列恢复
And 进度中的根目录状态应继续累加而不是从零开始
```

### Scenario 4: force_rebuild 或 resume=false 时应清除 SQLite checkpoint
```gherkin
Given 某个根目录在 SQLite 中已有旧的 CrawlCheckpoint
When 用户以 force_rebuild=true 或 resume_progress=false 启动全量同步
Then SyncManager 应在 crawl 前清除该根目录的 SQLite checkpoint
And crawler 应从根目录重新开始遍历
```

### Scenario 5: crawl 完成后应清理对应的 SQLite checkpoint
```gherkin
Given 某个根目录在同步过程中持续写入 SQLite checkpoint
When 该根目录全量 crawl 成功结束
Then 对应 checkpoint 记录应被清理或标记为空状态
And 下次 resume 不应恢复到已经完成的旧队列
```

## Feature: 现有外部接口在 SQLite 后端下保持兼容

### Scenario 6: GetSyncProgress 在 SQLite 后端下保持当前响应语义
```gherkin
Given 后端已切换到 SQLite 状态存储
When AdminService.GetSyncProgress 被调用
Then 返回的状态字段、根目录进度、聚合统计与错误信息应与迁移前保持兼容
And 前端不需要因为状态存储切换而修改协议或字段语义
```

### Scenario 7: WatchSyncProgress 在 SQLite 后端下持续推送最新进度
```gherkin
Given 后端已切换到 SQLite 状态存储
And 同步任务正在运行并周期性写入进度
When AdminService.WatchSyncProgress 建立流式订阅
Then 订阅方应持续收到最新的 SyncProgressState
And 最终应收到 done、error 或 cancelled 的终态
```

### Scenario 8: CLI sync-progress 从 SQLite 读取状态而不是依赖 JSON 文件
```gherkin
Given 运行环境已切换到 SQLite 状态库
And 旧的 progress JSON 文件不存在或未更新
When 用户执行 CLI sync-progress 命令
Then 命令仍应返回当前同步进度
And 数据来源应是 SQLite 中的 progress 记录
```

## Feature: 旧 JSON 状态可平滑迁移到 SQLite

### Scenario 9: 首次读取时可从 legacy JSON 惰性导入 progress 与 sync state
```gherkin
Given SQLite 状态库中还没有 progress 与 sync state 记录
And legacy JSON progress 与 sync state 文件存在且内容有效
When 新版本首次读取同步状态
Then 系统应先读取 legacy JSON
And 将对应状态写入 SQLite
And 后续读取应优先使用 SQLite 中的记录
```

### Scenario 10: 首次读取时可从 legacy checkpoint JSON 惰性导入根目录断点
```gherkin
Given SQLite 中还没有某个根目录的 checkpoint 记录
And 该根目录对应的 legacy checkpoint JSON 文件存在且内容有效
When 该根目录以 resume_progress=true 继续执行
Then 系统应从 legacy checkpoint JSON 导入到 SQLite
And crawler 应按导入后的 checkpoint 恢复执行
```

## Feature: 重启与并发写入下的可靠性

### Scenario 11: 进程重启后 running 状态会从 SQLite 恢复为 interrupted
```gherkin
Given SQLite 中持久化了一份 status=running 的 SyncProgressState
And 当前进程内没有活跃同步 goroutine
When 管理端调用 GetSyncProgress
Then 返回状态应被修正为 interrupted
And LastError 应提示进程重启导致同步中断
And 修正后的状态应回写到 SQLite
```

### Scenario 12: 并发进度保存不会产生损坏或半写入状态
```gherkin
Given 同步任务在多次进度回调中频繁写入 SQLite
When 多个保存操作在短时间内连续发生
Then SQLite 中最终保存的 payload 应保持完整可反序列化
And 读取到的根目录统计与队列状态不应出现损坏或非法空值
```
