# Task 002: Buf 生成链路配置 connect-go/connect-es

## Description

更新 `buf.gen.yaml`，确保 Go 端使用 `connect-go`，TS 端使用 `connect-es`，并保留消息类型生成。

## Execution Context

**Phase**: Green (Tooling)  
**depends-on**: `task-001-full-api-proto-schema.md`

## BDD Scenario Reference

- Scenario 2

## Files to Modify/Create

- Modify: `buf.gen.yaml`
- Modify: `web/package.json`（Connect 运行时依赖）

## Verification

```bash
./.bin/buf generate
find gen -type f | sort
```

## Success Criteria

- 生成目录包含 `*.connect.go` 与 `*_connect.ts`。
