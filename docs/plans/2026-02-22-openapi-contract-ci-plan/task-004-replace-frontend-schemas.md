# Task 004: Replace Frontend Schemas

**depends-on**: Task 003

**ref**: BDD Scenario "从 spec 生成的 Zod schemas 能解析后端 JSON 响应"

## Description

将前端手写的 Zod schemas 替换为从 OpenAPI spec 生成的版本。

## What to do

1. 在 `web/src/lib/schemas.ts` 中：
   - 将 `IndexDocumentSchema`、`SearchResponseSchema`、`DownloadURLResponseSchema`、`ErrorResponseSchema` 替换为从 `@/api/generated/zod.gen` 导入的对应 schema
   - 保留 type export（使用 `z.infer<>` 从生成的 schema 派生）

2. 在 `web/src/lib/sync-schemas.ts` 中：
   - 将 `CrawlStatsSchema`、`RootProgressSchema`、`IncrementalSyncStatsSchema`、`SyncProgressSchema` 替换为从 `@/api/generated/zod.gen` 导入的对应 schema
   - 保留 type export

3. 更新所有引用这些 schema 的文件（hooks、components、tests），确保 import 路径仍然有效
   - 如果生成的 schema 名称不同（如 `zSyncProgressState` vs `SyncProgressSchema`），创建 re-export alias

4. 运行现有前端测试确认无回归

## Files to modify

- `web/src/lib/schemas.ts`
- `web/src/lib/sync-schemas.ts`
- 可能需要调整的 import：`web/src/hooks/use-admin-auth.ts`、`web/src/hooks/use-sync-progress.ts`、`web/src/hooks/use-search.ts` 等

## Verification

- `cd web && bun run test` 全部通过
- `cd web && bun run typecheck`（如果能跑的话）无新增错误
- 删除手写 schema 后，前端代码仍能正常编译和使用生成的 schema
