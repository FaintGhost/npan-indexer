# 任务计划（2026-02-24）

## 目标

- 设计并确认方案 B：修复 `force_rebuild` 下统计异常根因 + 增加单独索引目录能力 + 接入官方统计校验能力。

## 计划清单

- [x] 1. 复核当前全量同步、checkpoint、进度展示链路并形成根因结论
- [x] 2. 调研 FangCloud OpenAPI 可用于目录统计/校验的接口与参数约束
- [x] 3. 设计后端改造（checkpoint 清理、请求参数扩展、统计校验接口/服务）
- [x] 4. 设计前端改造（单目录输入、开关参数、结果展示与误用防护）
- [x] 5. 编写 BDD 场景与测试策略（后端单测 + 前端单测 + 集成验证）
- [x] 6. 产出设计文档包到 `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/`
- [x] 7. 回填评审结论与实现前置条件

## 评审记录（完成后填写）

- 结论：方案 B 可行，且能直接覆盖“全量+强制重建统计偏小”的核心根因。
- 根因优先级：`force_rebuild` 未清理 checkpoint > 目录范围误配置 > 上游分页异常。
- 实施建议：先做后端 checkpoint 修复与测试，再做前端目录范围输入，最后做统计告警展示。
- 设计产物：
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/_index.md`
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/architecture.md`
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/bdd-specs.md`
  - `docs/plans/2026-02-24-full-rebuild-and-folder-scope-design/best-practices.md`

## 实施计划（OpenAPI 约束）

- [x] 1. 完成 `docs/plans/2026-02-24-full-rebuild-and-folder-scope-plan/` 任务分解
- [x] 2. 按计划执行后端 checkpoint 修复与测试
- [x] 3. 按计划执行后端 folder info 估计与告警
- [x] 4. 按计划执行前端目录范围输入与请求体对齐
- [x] 5. 按计划执行前端估计/告警展示
- [x] 6. 运行生成校验与回归测试并记录结果

## Review（本轮实施结果）

- 后端：
  - 已修复 `force_rebuild` 与 `resume=false` 场景下的 checkpoint 清理，避免残留断点污染全量统计。
  - 已接入 `GetFolderInfo`（`/api/v2/folder/{id}/info`）用于显式根目录估计值与名称回填。
  - 已在 verification 阶段追加 root 级“估计 vs 实际”差异告警。
- 前端：
  - Admin 页面新增目录 ID 输入（逗号分隔），支持单目录/多目录范围索引。
  - `startSync` 在传入 root IDs 时会自动发送 `include_departments=false`（契约字段）。
  - 根目录详情新增“估计/实际”展示，便于直接定位差异。
- OpenAPI 约束：
  - `api/openapi.yaml` 未发生契约漂移。
  - 生成校验已通过（`go generate ./api/...` + `bun run generate` + `git diff --exit-code -- api/types.gen.go web/src/api/generated/`）。
- 测试验证：
  - `go test ./...` 通过。
  - `cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/admin-page.test.tsx src/components/sync-progress-display.test.tsx` 通过。
  - `cd web && bun vitest run` 全量前端单测通过（22 files / 184 tests）。
  - `./tests/smoke/smoke_test.sh` 通过（34/34）。
  - `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright` 通过（32/32）。
