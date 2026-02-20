# Task 001: 场景红测（估算字段与摘要渲染）

**depends-on**: none

## BDD 场景关联

- `bdd-specs.md` Scenario 1: 可获取根目录总量时输出估算百分比
- `bdd-specs.md` Scenario 2: 无法获取总量时回退 n/a
- `bdd-specs.md` Scenario 3: 部门根目录自动注入估算总量

## 目标

- 先建立失败测试，锁定“估算字段”和“human 摘要展示”的期望行为。

## 变更范围

- Update: `internal/cli/root_progress_test.go`
- Create: `internal/service/sync_manager_estimate_test.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/cli ./internal/service -run Estimate -count=1
```
