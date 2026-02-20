# Task 002: 场景绿测（实现人类可读进度输出）

**depends-on**: task-001-red-progress-format

## BDD 场景关联

- `bdd-specs.md` Scenario: 默认输出为人类可读摘要
- `bdd-specs.md` Scenario: 支持 JSON 进度模式
- `bdd-specs.md` Scenario: 中断时输出友好摘要

## 目标

- 新增进度渲染器与模式开关，实现 `human/json` 两种输出模式。
- 将 `sync-full` 轮询输出与中断输出接入新渲染逻辑。

## 变更范围

- Update: `internal/cli/root.go`
- Update/Create: `internal/cli/root_progress_test.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/cli -count=1
```
