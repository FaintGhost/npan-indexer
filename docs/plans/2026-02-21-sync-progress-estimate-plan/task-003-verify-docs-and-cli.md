# Task 003: 文档与验证（运行手册与全量回归）

**depends-on**: task-002-green-estimate-progress.md

## BDD 场景关联

- `bdd-specs.md` Scenario 1: 可获取根目录总量时输出估算百分比
- `bdd-specs.md` Scenario 2: 无法获取总量时回退 n/a

## 目标

- 更新运维文档，说明估算进度字段的含义与限制。
- 运行全量测试并确认 CLI 帮助与输出符合预期。

## 变更范围

- Update: `docs/runbooks/index-sync-operations.md`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1
go run ./cmd/cli sync-full --help
```
