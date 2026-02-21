# Task 002: 绿测（Tailwind 4 + View Transition UI 重构）

**depends-on**: task-001-red-demo-ui-safety.md

## 目标

- 引入 Tailwind Browser v4，重构页面视觉层级与交互细节。
- 用 View Transition API 包裹列表切换和增量渲染，减少跳变。

## 变更范围

- Update: `web/demo/index.html`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1
```
