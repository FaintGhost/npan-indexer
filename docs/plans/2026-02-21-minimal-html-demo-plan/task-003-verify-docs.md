# Task 003: 文档与验证

**depends-on**: task-002-green-demo-route-and-page.md

## 目标

- 更新 README 使用说明，给出最小 Demo 访问与测试步骤。
- 完成构建与全量测试验证。

## 变更范围

- Update: `README.md`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...
```
