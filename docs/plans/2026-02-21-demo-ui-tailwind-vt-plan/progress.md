# 进度记录

## 2026-02-21

- 已完成：创建 Demo UI Tailwind + View Transition 改造计划。
- 已完成：`web/demo/index.html` 引入 `@tailwindcss/browser@4`，完成页面重构。
- 已完成：在列表重绘与增量加载时接入 View Transition API（可降级）。
- 已完成：保留 end user 下载链路（`/api/v1/demo/search` + `/api/v1/demo/download-url`）。
- 已完成：搜索框交互改为“初始居中 -> 有结果后吸顶 sticky”。
- 已完成：通过不透明吸顶层和分层布局，避免滚动内容视觉越过搜索框。
- 已完成：README 更新 demo 交互说明。

## 验证结果

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...` 通过。
