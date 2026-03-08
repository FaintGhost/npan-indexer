# Task 001: [IMPL] SQLite 状态库与 legacy 导入 (GREEN)

**depends-on**: task-001-sqlite-state-store-test

## Description

实现 SQLite 状态库、统一 state entry 存储模型与 legacy JSON 惰性导入逻辑，使 SQLite 成为 progress、sync_state、checkpoint 的主状态源。该任务只实现存储层，不改服务层 wiring。

## Execution Context

**Task Number**: 001 of 009
**Phase**: Foundation
**Prerequisites**: `task-001-sqlite-state-store-test` 已完成并稳定处于 Red

## BDD Scenario

```gherkin
Scenario: 首次读取时可从 legacy JSON 惰性导入 progress 与 sync state
  Given SQLite 状态库中还没有 progress 与 sync state 记录
  And legacy JSON progress 与 sync state 文件存在且内容有效
  When 新版本首次读取同步状态
  Then 系统应先读取 legacy JSON
  And 将对应状态写入 SQLite
  And 后续读取应优先使用 SQLite 中的记录

Scenario: 首次读取时可从 legacy checkpoint JSON 惰性导入根目录断点
  Given SQLite 中还没有某个根目录的 checkpoint 记录
  And 该根目录对应的 legacy checkpoint JSON 文件存在且内容有效
  When 该根目录以 resume_progress=true 继续执行
  Then 系统应从 legacy checkpoint JSON 导入到 SQLite
  And crawler 应按导入后的 checkpoint 恢复执行

Scenario: 并发进度保存不会产生损坏或半写入状态
  Given 同步任务在多次进度回调中频繁写入 SQLite
  When 多个保存操作在短时间内连续发生
  Then SQLite 中最终保存的 payload 应保持完整可反序列化
  And 读取到的根目录统计与队列状态不应出现损坏或非法空值
```

**Spec Source**: `../2026-03-08-sync-state-sqlite-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `internal/storage/json_store.go`
- Modify: `internal/storage/json_store_test.go`
- Create: `internal/storage/sqlite_store.go`
- Create: `internal/storage/sqlite_store_test.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/validate.go`
- Modify: `internal/config/validate_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

## Steps

### Step 1: Verify Scenario

- 对照 BDD 文档确认本任务只覆盖“状态库 + 导入 + 基础配置”，不提前改 `SyncManager`。

### Step 2: Implement Logic (Green)

- 增加 SQLite store 抽象与实现，覆盖 progress、sync_state 与 checkpoint。
- 引入统一 state entry schema，并为不同 namespace/key 提供读写入口。
- 实现 legacy JSON 惰性导入逻辑，确保 SQLite 缺失记录时才读取旧文件。
- 新增状态库配置项与基础校验，确保 SQLite 文件路径可配置且不破坏现有默认值。
- 选择兼容 `CGO_ENABLED=0` 的 SQLite 驱动，并完成依赖接入。

### Step 3: Verify Pass

- 重新运行 `task-001` 的目标测试，确认通过。
- 验证新增存储测试不会依赖真实网络。

### Step 4: Refactor & Safety Check

- 收敛公共 JSON 序列化、SQLite 读写与导入辅助逻辑，避免重复分支。
- 确认 legacy JSON 仍保留为非破坏式回退来源。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/storage ./internal/config -run 'SQLite|LegacyImport|ConcurrentSave|StateDB' -count=1
GOCACHE=/tmp/go-build go test ./internal/storage ./internal/config -count=1
```

## Success Criteria

- SQLite store 测试全部转绿。
- progress、sync_state、checkpoint 都能通过统一 SQLite 状态库读写。
- legacy JSON 惰性导入成立，且不会依赖双写作为主逻辑。
- 依赖与配置改动不破坏 `CGO_ENABLED=0` 构建前提。
