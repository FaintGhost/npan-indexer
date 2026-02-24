# Task 005: App/Auth/Search 域 Connect 接入（下一批）

## Description

把 `App`、`Auth`、`Search`（含 Download）域逐步从 REST handler 迁移到 Connect handler，优先复用现有 service 层。

## Execution Context

**Phase**: Future Batch  
**depends-on**: `task-004-health-connect-tests-and-regression.md`

## BDD Scenario Reference

- Scenario 1
- Scenario 4

## Verification

```bash
go test ./internal/httpx -count=1
cd web && bun vitest run
```

## Success Criteria

- 以上域的 Connect handler 可用，REST 保持兼容。
