# 进度记录

## 2026-02-21

- 已完成：创建 Demo 动态搜索卡住修复计划。
- 已完成：在 `web/demo/index.html` 引入 `AbortController`，实现 in-flight 请求中止。
- 已完成：调整 replace 搜索路径并发门控，允许新查询抢占旧查询，避免输入后“卡住”。
- 已完成：处理 `AbortError`，避免中止请求污染状态提示与列表渲染。

## 验证结果

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1` 通过。
