# Task 003: Health Connect handler 与路由挂载

## Description

实现 `HealthService` 的 `connect-go` handler，并在 Echo 中挂载 Connect 路由，与现有 REST 共存。

## Execution Context

**Phase**: Green (Implementation)  
**depends-on**: `task-002-buf-generation-connect-go-es.md`

## BDD Scenario Reference

- Scenario 3
- Scenario 4

## Files to Modify/Create

- Create: `internal/httpx/connect_health.go`
- Modify: `internal/httpx/server.go`
- Modify: `go.mod`
- Modify: `go.sum`

## Verification

```bash
go test ./internal/httpx -run 'Health|Routes' -count=1
```

## Success Criteria

- Connect Health 路由可访问。
- REST 路由不受影响。
