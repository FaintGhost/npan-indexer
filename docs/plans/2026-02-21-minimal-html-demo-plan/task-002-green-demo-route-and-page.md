# Task 002: 绿测（实现纯 HTML Demo 与服务路由）

**depends-on**: task-001-red-demo-route-and-page.md

## 目标

- 新增 `/demo` 页面路由。
- 新增纯 HTML 单文件页面，支持搜索、选择、批量生成下载链接、复制链接、直接下载。

## 变更范围

- Update: `internal/httpx/server.go`
- Add: `web/demo/index.html`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1
```
