#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../.." && pwd)
README_FILE="$REPO_ROOT/README.md"
STRUCTURE_FILE="$REPO_ROOT/docs/STRUCTURE.md"
MAKEFILE_FILE="$REPO_ROOT/Makefile"

failures=0

pass() {
  printf 'PASS: %s\n' "$1"
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  failures=$((failures + 1))
}

assert_file_exists() {
  local file_path="$1"
  local label="$2"

  if [ -f "$file_path" ]; then
    pass "$label"
  else
    fail "$label (missing: $file_path)"
  fi
}

assert_contains() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if grep -Fq -- "$needle" "$file_path"; then
    pass "$label"
  else
    fail "$label (missing: $needle)"
  fi
}

assert_not_contains() {
  local file_path="$1"
  local needle="$2"
  local label="$3"

  if grep -Fq -- "$needle" "$file_path"; then
    fail "$label (found: $needle)"
  else
    pass "$label"
  fi
}

assert_not_exists() {
  local file_path="$1"
  local label="$2"

  if [ -e "$file_path" ]; then
    fail "$label (still exists: $file_path)"
  else
    pass "$label"
  fi
}

assert_file_exists "$README_FILE" "README.md 存在"
assert_file_exists "$STRUCTURE_FILE" "docs/STRUCTURE.md 存在"
pass "开始验证开发者活跃入口是否已切换到 task"

assert_contains "$README_FILE" "go run ./cmd/server" "README 保留后端长驻开发命令"
assert_contains "$README_FILE" "bun run dev" "README 保留前端长驻开发命令"

assert_not_contains "$README_FILE" "make rest-guard" "README 不再把 make rest-guard 作为活跃入口"
assert_not_contains "$README_FILE" "make smoke-test" "README 不再把 make smoke-test 作为活跃入口"
assert_not_contains "$README_FILE" "make e2e-test" "README 不再把 make e2e-test 作为活跃入口"

assert_contains "$README_FILE" "task guard:rest" "README 提供 task guard:rest 验证入口"
assert_contains "$README_FILE" "task verify:smoke" "README 提供 task verify:smoke 验证入口"
assert_contains "$README_FILE" "task verify:e2e" "README 提供 task verify:e2e 验证入口"

assert_not_contains "$STRUCTURE_FILE" "make test" "docs/STRUCTURE.md 发布前建议不再以 make test 为主入口"
assert_not_contains "$STRUCTURE_FILE" "make smoke-test" "docs/STRUCTURE.md 发布前建议不再以 make smoke-test 为主入口"
assert_not_contains "$STRUCTURE_FILE" "make e2e-test" "docs/STRUCTURE.md 发布前建议不再以 make e2e-test 为主入口"
assert_contains "$STRUCTURE_FILE" "task verify:smoke" "docs/STRUCTURE.md 发布前建议切换为 task verify:smoke"
assert_contains "$STRUCTURE_FILE" "task verify:e2e" "docs/STRUCTURE.md 发布前建议切换为 task verify:e2e"

assert_not_exists "$MAKEFILE_FILE" "根 Makefile 已移除"

if grep -Fq -- "go run ./cmd/server" "$README_FILE" \
  && grep -Fq -- "bun run dev" "$README_FILE" \
  && grep -Fq -- "task guard:rest" "$README_FILE" \
  && grep -Fq -- "task verify:smoke" "$README_FILE" \
  && grep -Fq -- "task verify:e2e" "$README_FILE"; then
  pass "README 已清楚区分长驻开发命令与验证命令"
else
  fail "README 尚未清楚区分长驻开发命令与验证命令（长驻命令存在，但 task 验证入口未完整出现）"
fi

if [ "$failures" -gt 0 ]; then
  printf '\nRED expected: %s assertion(s) failed because developer entrypoints are not fully migrated to task yet.\n' "$failures" >&2
  exit 1
fi

printf '\nAll developer entrypoint assertions passed.\n'
