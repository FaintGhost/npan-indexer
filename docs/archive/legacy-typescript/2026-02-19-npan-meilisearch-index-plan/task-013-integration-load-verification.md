# Task 013: 联调与压测验证（限流阈值与稳定性）

**depends-on**: task-002-green-full-crawl-rate-limit, task-004-green-retry-checkpoint, task-006-green-meili-schema-searchability, task-008-green-incremental-delete-sync, task-010-green-meili-query-api, task-012-green-download-url-proxy

## Description

对全链路进行联调验证，确认限流参数、重试策略、吞吐和稳定性满足上线基线。

## Execution Context

**Task Number**: 013 of 013  
**Phase**: Testing  
**Prerequisites**: 关键 Green 任务均已完成。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 1~6 综合验收  

## Files to Modify/Create

- Create: `tests/integration/index-sync-flow.test.ts`
- Create: `scripts/run-load-check.sh`
- Create: `docs/runbooks/index-sync-operations.md`

## Steps

### Step 1: Verify Scenario Coverage
- 列出场景 1~6 的覆盖矩阵，确认没有遗漏。

### Step 2: Integration Verification
- 运行全链路集成测试：全量同步、增量同步、搜索、下载代理。

### Step 3: Load Verification
- 在受控并发下进行压测，记录 QPS、错误率、重试次数、平均延迟。
- 调整默认限流参数并回归验证。

### Step 4: Operational Documentation
- 输出运行手册：首次全量、增量调度、故障恢复、告警阈值。

## Verification Commands

```bash
bun test
bash scripts/run-load-check.sh
```

## Success Criteria

- 场景 1~6 均有自动化验证结果。
- 限流参数在压测下稳定，无持续 429 风暴。
- 运行手册可支撑值班与故障恢复。
