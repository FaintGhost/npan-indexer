# Task 001: [TEST] SQLite 状态库与 legacy 导入 (RED)

**depends-on**: (none)

## Description

先用失败测试锁定“SQLite 成为主状态源 + legacy JSON 惰性导入”的核心语义，确保后续实现不会退化成继续依赖 JSON 文件的双主路径。该任务只负责编写与验证失败测试，不实现 SQLite store 本身。

## Execution Context

**Task Number**: 001 of 009
**Phase**: Foundation
**Prerequisites**: 已阅读 `internal/storage/json_store.go`、`internal/models/models.go` 与设计文档中的 SQLite state entry 结构

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

- Modify: `internal/storage/json_store_test.go`
- Create: `internal/storage/sqlite_store_test.go`
- Modify: `internal/config/validate_test.go`（如需锁定新 SQLite 配置约束）

## Steps

### Step 1: Verify Scenario

- 确认以上 3 个场景在 BDD 文档中存在，且覆盖 progress、sync_state、checkpoint 与并发写语义。

### Step 2: Implement Test (Red)

- 为新的 SQLite store 补充失败测试，覆盖：
  - schema 初始化后可区分 `progress/default`、`sync_state/default`、`checkpoint/<key>`
  - 首次 `Load()` 在 SQLite 为空时会读取 legacy JSON 并导入
  - 导入完成后，即使 legacy JSON 被删除，后续仍可从 SQLite 读到相同状态
  - 并发 `Save()` 后最终 payload 仍可反序列化且字段完整
- 使用临时目录、临时 SQLite 文件与测试替身；禁止依赖真实网络与外部服务。

### Step 3: Verify Red Failure

- 运行目标 Go 测试并确认失败。
- 失败原因必须指向“缺少 SQLite store / 缺少惰性导入 / 缺少并发保存保证”，不能是环境或网络错误。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/storage ./internal/config -run 'SQLite|LegacyImport|ConcurrentSave|StateDB' -count=1
```

## Success Criteria

- 新增测试可稳定处于 Red。
- 失败直接指向 SQLite 状态库与迁移语义缺失。
- 测试只依赖本地临时文件，不依赖真实外部服务。
