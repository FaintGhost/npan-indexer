# Task 001: 场景红测（进度格式化与输出模式）

**depends-on**: none

## BDD 场景关联

- `bdd-specs.md` Scenario: 默认输出为人类可读摘要
- `bdd-specs.md` Scenario: 支持 JSON 进度模式

## 目标

- 为进度摘要格式化与输出模式判定建立测试桩，先让测试失败（Red）。

## 变更范围

- Create: `internal/cli/root_progress_test.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/cli -run Progress -count=1
```
