# Task 006: Admin 域 Connect 接入与切换策略（下一批）

## Description

迁移 `StartSync/GetSyncProgress/CancelSync/InspectRoots` 到 Connect，并定义 REST 下线门槛（流量/稳定性/回滚方案）。

## Execution Context

**Phase**: Future Batch  
**depends-on**: `task-005-app-auth-search-connect-migration.md`

## BDD Scenario Reference

- Scenario 1
- Scenario 4

## Verification

```bash
go test ./internal/httpx ./internal/service -count=1
./tests/smoke/smoke_test.sh
```

## Success Criteria

- Admin Connect 全量可用。
- 有明确 cutover 策略与回滚路径。
