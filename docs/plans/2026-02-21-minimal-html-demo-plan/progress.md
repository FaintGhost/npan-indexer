# 进度记录

## 2026-02-21

- 已完成：建立最小 HTML Demo 实施计划。
- 已完成：新增 `/demo` 与 `/demo/` 路由，服务端可直接返回 Demo 页面。
- 已完成：新增纯 HTML 单文件页面 `web/demo/index.html`，支持搜索、选择文件、批量生成下载链接、复制链接、直接下载。
- 已完成：新增测试 `internal/httpx/server_demo_test.go`，覆盖路由注册与页面返回。
- 已完成：README 增加最小 Demo 访问说明。

## 验证结果

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1` 通过。
