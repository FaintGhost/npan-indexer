# Task 003: 绿测（实现 sticky 搜索与无限滚动 UI）

**depends-on**: task-002-green-demo-enduser-api.md

## BDD 场景关联

- `bdd-specs.md` Scenario 1
- `bdd-specs.md` Scenario 2
- `bdd-specs.md` Scenario 3
- `bdd-specs.md` Scenario 4

## 目标

- 页面改为 end user 模式：仅展示搜索与结果，不展示凭据输入。
- 支持输入即搜、懒加载分页、点击直接下载。

## 变更范围

- Update: `web/demo/index.html`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -run Demo -count=1
```
