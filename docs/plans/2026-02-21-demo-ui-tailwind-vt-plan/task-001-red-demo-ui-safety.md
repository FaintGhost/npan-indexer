# Task 001: 红测（页面安全与入口文案约束）

**depends-on**: none

## 目标

- 确保 `/demo` 页面基础标识存在，并且不出现凭据输入相关文案。

## 变更范围

- 复用现有测试：`internal/httpx/server_demo_test.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1
```
