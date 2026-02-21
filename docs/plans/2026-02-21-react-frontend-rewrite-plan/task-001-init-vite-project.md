# Task 001: 初始化 Vite + React 19 + TypeScript 项目

**depends-on**: (none)

## Description

初始化前端项目脚手架：Vite 7 + React 19 + TypeScript strict 模式。项目目录为 `cli/`（仓库根目录下）。

## Execution Context

**Task Number**: 001 of 046
**Phase**: Setup
**Prerequisites**: Node.js >= 20, npm >= 10

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: (Infrastructure — all scenarios depend on this)

## Files to Modify/Create

- Create: `cli/package.json`
- Create: `cli/vite.config.ts`
- Create: `cli/tsconfig.json`
- Create: `cli/tsconfig.app.json`
- Create: `cli/tsconfig.node.json`
- Create: `cli/index.html`
- Create: `cli/src/main.tsx`
- Create: `cli/src/vite-env.d.ts`

## Steps

### Step 1: Scaffold Vite project

- Use `npm create vite@latest` with React + TypeScript template in `cli/` directory, or manually create the files
- Ensure `package.json` specifies React 19 (`"react": "^19.0.0"`, `"react-dom": "^19.0.0"`)
- Ensure `vite.config.ts` uses `@vitejs/plugin-react`

### Step 2: Configure TypeScript strict mode

- Set `"strict": true` in `tsconfig.json`
- Configure path alias `@/` → `src/` in both tsconfig and vite config

### Step 3: Configure build output

- Set `build.outDir` to `"../web/dist"` in `vite.config.ts`
- Set `build.emptyOutDir` to `true`
- Set `base` to `"/app/"` for the SPA base path

### Step 4: Install dev dependencies

- Install: `oxlint` for linting
- Ensure `index.html` includes Google Fonts preconnect links (Noto Sans SC + Outfit)

### Step 5: Verify

- `npm run dev` starts dev server
- `npm run build` produces output in `../web/dist/`
- `npx tsc --noEmit` passes with zero errors

## Verification Commands

```bash
cd cli && npm install && npm run dev -- --host 0.0.0.0 &
sleep 3 && curl -s http://localhost:5173/ | head -5
npm run build
npx tsc --noEmit
```

## Success Criteria

- Vite dev server starts without errors
- TypeScript strict mode compilation passes
- Build output appears in `web/dist/`
- `index.html` includes font preconnect links
