# Task 003: RED backend catalog preserve progress tests

## Description

先用失败测试锁定“局部全量后历史根目录详情不被覆盖”的服务端语义，避免前端靠缓存掩盖问题。

## Execution Context

**Task Number**: 003 of 011  
**Phase**: Testing (Red)  
**Prerequisites**: 无

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `局部补同步后根目录详情列表仍保留历史条目`

## Files to Modify/Create

- Modify: `internal/service/sync_manager_progress_test.go`
- Modify: `internal/service/sync_manager_routing_test.go`（若语义放在 routing 流程验证）

## Steps

### Step 1: Verify Scenario

- 确认场景要求区分“本次执行 roots”与“展示目录册”。

### Step 2: Implement Test (Red)

- 构造 existing progress（含多个历史 roots）。
- 发起仅单目录 scoped full 请求（带保留目录册语义开关/字段）。
- 断言运行后：
  - 历史目录册仍包含全部历史项
  - 本次执行目录统计被更新
  - 非执行目录未被清空
- 使用 fake store / fake npan API / fake index，隔离网络与外部服务。

### Step 3: Verify Failure

- 在未实现 catalog 保留逻辑前，确认测试失败。

## Verification Commands

```bash
go test ./internal/service -run CatalogPreserve -count=1
```

## Success Criteria

- 红测能稳定复现“历史目录被覆盖”问题。
- 失败断言直指 progress 语义，而非偶发时序问题。
