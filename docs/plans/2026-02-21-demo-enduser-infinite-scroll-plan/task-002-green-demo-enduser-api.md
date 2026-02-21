# Task 002: 绿测（实现后端 demo 接口）

**depends-on**: task-001-red-demo-enduser-api-and-ui.md

## BDD 场景关联

- `bdd-specs.md` Scenario 2
- `bdd-specs.md` Scenario 4

## 目标

- 增加 end user 专用接口：demo 搜索与下载链接代理。
- 搜索仅返回文件结果，下载接口强制使用服务端配置凭据。

## 变更范围

- Update: `internal/httpx/server.go`
- Update: `internal/httpx/handlers.go`

## 验证

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/httpx -count=1
```
