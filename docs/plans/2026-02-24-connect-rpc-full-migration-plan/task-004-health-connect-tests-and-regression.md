# Task 004: Health Connect 集成测试与回归验证

## Description

新增 Connect 级别测试，校验 `Health/Readyz` RPC 可调用并返回预期字段；并执行本轮最小回归。

## Execution Context

**Phase**: Red/Green (Tests + Verify)  
**depends-on**: `task-003-health-connect-handler-and-routing.md`

## BDD Scenario Reference

- Scenario 3
- Scenario 4

## Files to Modify/Create

- Create: `internal/httpx/connect_health_test.go`

## Verification

```bash
go test ./internal/httpx -run 'Connect|Health|Routes' -count=1
GOCACHE=/tmp/go-build go test ./... -count=1
```

## Success Criteria

- 新增 Connect 测试通过。
- 全仓 Go 测试通过。
