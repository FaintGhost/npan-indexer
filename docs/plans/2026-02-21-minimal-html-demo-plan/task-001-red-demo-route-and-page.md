# Task 001: 红测（Demo 路由与页面期望）

**depends-on**: none

## 目标

- 先建立失败测试，确认 `/demo` 路由存在且可返回页面内容。

## 变更范围

- Create: `internal/httpx/server_demo_test.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1
```
