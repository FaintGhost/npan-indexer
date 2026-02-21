# Task 001: 红测（demo 专用接口与页面行为）

**depends-on**: none

## BDD 场景关联

- `bdd-specs.md` Scenario 1
- `bdd-specs.md` Scenario 4

## 目标

- 先建立失败测试，锁定 demo 路由和页面中“无凭据输入”的要求。

## 变更范围

- Update: `internal/httpx/server_demo_test.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1
```
