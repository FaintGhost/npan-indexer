# Task 001: 全量 API proto 契约映射（Schema）

## Description

将当前 `api/openapi.yaml` 已公开的所有接口映射到 `proto/npan/v1/*.proto`，确保 RPC 与消息可覆盖现有对外能力。

## Execution Context

**Phase**: Green (Schema)  
**depends-on**: 无

## BDD Scenario Reference

- Scenario 1

## Files to Modify/Create

- Create/Modify: `proto/npan/v1/api.proto`

## Verification

```bash
./.bin/buf lint
```

## Success Criteria

- 所有公开 operation 都有对应 RPC。
- `buf lint` 通过。
