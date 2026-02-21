# 进度记录

## 2026-02-21

- 已完成：创建 end user demo 改造计划与 BDD 规格。
- 已完成：新增 demo 专用接口 `/api/v1/demo/search` 与 `/api/v1/demo/download-url`。
- 已完成：`/demo` 改为 end user 体验，移除 token/API key 输入。
- 已完成：页面支持 sticky 搜索框、输入即搜（debounce）、无限滚动懒加载、点击直接下载。
- 已完成：更新 README 中 demo 接口与说明。

## 验证结果

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...` 通过。
