# Task 002: 场景绿测（接入 item_count 并输出估算进度）

**depends-on**: task-001-red-estimate-progress.md

## BDD 场景关联

- `bdd-specs.md` Scenario 1: 可获取根目录总量时输出估算百分比
- `bdd-specs.md` Scenario 2: 无法获取总量时回退 n/a
- `bdd-specs.md` Scenario 3: 部门根目录自动注入估算总量

## 目标

- 在模型层新增目录 `item_count` 映射并透传到同步进度。
- 在 `sync-full` human 进度摘要中增加估算进度维度。

## 变更范围

- Update: `internal/models/models.go`
- Update: `internal/npan/client.go`
- Update: `internal/service/sync_manager.go`
- Update: `internal/cli/root.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/cli ./internal/service ./internal/npan -count=1
```
