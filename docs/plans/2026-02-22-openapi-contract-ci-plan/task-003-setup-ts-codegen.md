# Task 003: Setup TS Codegen

**depends-on**: Task 001

**ref**: BDD Scenario "从 spec 生成的 Zod schemas 能解析后端 JSON 响应"

## Description

安装 @hey-api/openapi-ts，配置 Zod v3 插件，从 OpenAPI spec 生成 TypeScript types 和 Zod schemas。

## What to do

1. 在 `web/` 目录安装 `@hey-api/openapi-ts` 作为 devDependency

2. 创建 `web/openapi-ts.config.ts` 配置文件：
   - `input: "../api/openapi.yaml"`
   - `output: "src/api/generated"`
   - 插件：`@hey-api/typescript` + `zod`（compatibilityVersion: 3）

3. 在 `web/package.json` 的 scripts 中添加 `"generate": "openapi-ts"`

4. 运行生成命令，产出 `web/src/api/generated/` 目录

5. 将 `web/src/api/generated/` 加入 `.gitignore` 的排除列表（不忽略，需要提交）

## Files to create/modify

- `web/openapi-ts.config.ts`（新建）
- `web/src/api/generated/`（生成产物目录）
- `web/package.json`（添加 generate script 和 devDependency）

## Verification

- `cd web && bun run generate` 成功执行
- `web/src/api/generated/zod.gen.ts` 包含 `SyncProgressState`、`SearchResponse` 等 Zod schemas
- 生成的 Zod schema 中 `SyncProgressState.status` 是 `z.enum([...])` 包含所有 6 个状态值
- 生成的 schema 字段名与手写的 `schemas.ts` 和 `sync-schemas.ts` 一致
