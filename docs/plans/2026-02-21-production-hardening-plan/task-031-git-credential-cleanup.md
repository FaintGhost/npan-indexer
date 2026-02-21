# Task 031: Git credential cleanup

**depends-on**: (none)

## Description

Remove tracked secret files (.env, .env.meilisearch) from git, ensure .gitignore covers them, create .env.example templates, and document the credential rotation requirement. This is a manual operations task.

## Execution Context

**Task Number**: 031 of 032
**Phase**: Cleanup & Verification
**Prerequisites**: None — can be done at any time

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- Feature 5 — ".env 文件不在 git 中"
- Feature 5 — "API 响应不包含服务端凭据"

## Files to Modify/Create

- Modify: `.gitignore` — ensure .env and .env.meilisearch are listed
- Create: `.env.example` — template with placeholder values
- Create: `.env.meilisearch.example` — template with placeholder values
- Git operations: `git rm --cached .env .env.meilisearch`

## Steps

### Step 1: Check current git tracking

- Run `git ls-files --error-unmatch .env .env.meilisearch` to check if files are tracked
- If tracked, proceed with removal

### Step 2: Remove from git tracking

- `git rm --cached .env .env.meilisearch`
- Do NOT delete the actual files (they're needed locally)

### Step 3: Verify .gitignore

- Ensure `.gitignore` contains entries for `.env`, `.env.*`, `!.env.example`, `!.env.meilisearch.example`

### Step 4: Create example files

- Create `.env.example` with all required environment variables and placeholder values (e.g., `NPA_ADMIN_API_KEY=your-admin-key-here-minimum-16-chars`)
- Create `.env.meilisearch.example` with `MEILI_MASTER_KEY=your-master-key-here`

### Step 5: Document credential rotation

- Note that ALL credentials currently in .env and .env.meilisearch have been exposed in git history
- After the commit, all credentials must be rotated:
  - OAuth client_id/client_secret
  - Meilisearch master key / API key
  - Admin API key
- Consider running `git filter-repo` to clean history (optional, requires force push)

### Step 6: Verify

- `git ls-files .env .env.meilisearch` should return empty
- `.env.example` exists with placeholder values

## Verification Commands

```bash
git ls-files .env .env.meilisearch
# Should return nothing
cat .env.example
cat .env.meilisearch.example
```

## Success Criteria

- .env and .env.meilisearch not tracked by git
- .env.example and .env.meilisearch.example exist with safe placeholder values
- .gitignore properly configured
- Credential rotation documented/planned
