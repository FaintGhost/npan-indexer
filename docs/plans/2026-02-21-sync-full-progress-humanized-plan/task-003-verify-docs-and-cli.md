# Task 003: 文档与验证（运行手册与命令校验）

**depends-on**: task-002-green-progress-format

## BDD 场景关联

- `bdd-specs.md` Scenario: 默认输出为人类可读摘要
- `bdd-specs.md` Scenario: 支持 JSON 进度模式

## 目标

- 更新运行手册，补充 `--progress-output` 的使用说明。
- 进行全量测试与构建验证，确认无回归。

## 变更范围

- Update: `docs/runbooks/index-sync-operations.md`
- Update: `README.md`（如需要）

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./...
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test -race ./...
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...
```
