# Task 005: 全链路验证与运行文档收口

**depends-on**: task-001-sqlite-state-store-impl, task-002-sync-manager-state-impl, task-003-checkpoint-lifecycle-impl, task-004-admin-cli-sqlite-impl

## Description

在实现全部完成后，执行完整验证链并更新运行文档，证明 SQLite 状态迁移在本仓库的默认开发、测试与 Docker 场景下都可稳定运行。

## Execution Context

**Task Number**: 005 of 009
**Phase**: Verification
**Prerequisites**: SQLite store、SyncManager、checkpoint 生命周期、Admin/CLI wiring 都已完成并通过对应局部测试

## BDD Scenario

```gherkin
Scenario: 全量同步成功后进度与增量游标写入 SQLite
  Given SyncManager 使用 SQLite progress store、sync state store 与 checkpoint store factory
  And 一次全量同步成功完成
  When 管理端或 CLI 读取同步状态
  Then 应能从 SQLite 读取到 status=done 的 SyncProgressState
  And 应能从 SQLite 读取到 LastSyncTime 大于 0 的 SyncState
  And 进度中的根目录统计、verification 与 completedRoots 应保持与迁移前一致

Scenario: GetSyncProgress 在 SQLite 后端下保持当前响应语义
  Given 后端已切换到 SQLite 状态存储
  When AdminService.GetSyncProgress 被调用
  Then 返回的状态字段、根目录进度、聚合统计与错误信息应与迁移前保持兼容
  And 前端不需要因为状态存储切换而修改协议或字段语义

Scenario: CLI sync-progress 从 SQLite 读取状态而不是依赖 JSON 文件
  Given 运行环境已切换到 SQLite 状态库
  And 旧的 progress JSON 文件不存在或未更新
  When 用户执行 CLI sync-progress 命令
  Then 命令仍应返回当前同步进度
  And 数据来源应是 SQLite 中的 progress 记录
```

**Spec Source**: `../2026-03-08-sync-state-sqlite-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `README.md`
- Modify: `docs/runbooks/index-sync-operations.md`
- Modify: `npan-indexer/CLAUDE.md`（如需补充默认状态库入口与注意事项）
- Modify: `tasks/todo.md`

## Steps

### Step 1: Verify Scenario

- 确认最终验收覆盖 SQLite 作为主状态源、Admin/CLI 兼容、文档可操作三类目标。

### Step 2: Execute Verification

- 先跑局部 Go 测试，确认本轮引入的问题已修复。
- 再跑全量 Go 测试与前端 Vitest。
- 最后跑 Docker smoke 与 Playwright E2E，确认 SQLite 状态迁移未破坏长链路。

### Step 3: Update Runbook

- 更新 README 与 runbook，明确：
  - 新增 SQLite 状态库配置项
  - legacy JSON 仅作为导入来源
  - 如何排查 SQLite 状态文件、如何保留/对照旧 JSON
- 在 `tasks/todo.md` 中记录本次计划与 review 结论。

### Step 4: Final Review

- 汇总改动文件、验证结果、残余风险与后续清理项。
- 明确说明是否仍保留 legacy JSON 兼容代码以及原因。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/storage ./internal/service ./internal/httpx ./internal/cli ./internal/config -count=1
GOCACHE=/tmp/go-build go test ./...
cd web && bun vitest run
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
./tests/smoke/smoke_test.sh
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
```

## Success Criteria

- 局部与全量测试通过。
- Docker smoke 与 Playwright E2E 通过。
- README / runbook / tasks 记录已更新，足以指导后续运维与排障。
- 最终 review 明确说明 SQLite 已成为主状态源，以及 legacy JSON 的剩余角色。
