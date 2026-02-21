# Task 003: 验证与文档更新

**depends-on**: task-002-green-demo-ui-tailwind-vt.md

## 目标

- 更新 README 中 demo 交互描述。
- 执行全量测试和构建验证。

## 变更范围

- Update: `README.md`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...
```
