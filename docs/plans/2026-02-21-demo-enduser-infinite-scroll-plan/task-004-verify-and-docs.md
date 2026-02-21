# Task 004: 验证与文档更新

**depends-on**: task-003-green-demo-enduser-ui.md

## BDD 场景关联

- `bdd-specs.md` Scenario 1
- `bdd-specs.md` Scenario 2
- `bdd-specs.md` Scenario 3
- `bdd-specs.md` Scenario 4

## 目标

- 更新 README 的 demo 使用说明。
- 运行全量测试与构建验证。

## 变更范围

- Update: `README.md`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...
```
