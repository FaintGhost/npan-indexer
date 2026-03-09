#!/usr/bin/env bash
set -u

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
workflow_path="${repo_root}/.github/workflows/ci.yml"

declare -a failures=()

action_pass() {
  printf 'PASS: %s\n' "$1"
}

action_fail() {
  failures+=("$1")
  printf 'FAIL: %s\n' "$1" >&2
}

count_matches() {
  local needle="$1"
  grep -F -c -- "$needle" "$workflow_path" || true
}

assert_contains() {
  local needle="$1"
  local message="$2"

  if grep -F -q -- "$needle" "$workflow_path"; then
    action_pass "$message"
  else
    action_fail "${message} (missing: ${needle})"
  fi
}

assert_not_contains() {
  local needle="$1"
  local message="$2"

  if grep -F -q -- "$needle" "$workflow_path"; then
    action_fail "${message} (unexpected legacy command: ${needle})"
  else
    action_pass "$message"
  fi
}

assert_count_at_least() {
  local needle="$1"
  local minimum="$2"
  local message="$3"
  local actual

  actual="$(count_matches "$needle")"
  if (( actual >= minimum )); then
    action_pass "$message"
  else
    action_fail "${message} (expected at least ${minimum}, found ${actual}: ${needle})"
  fi
}

if [[ ! -f "$workflow_path" ]]; then
  printf 'ERROR: workflow file not found: %s\n' "$workflow_path" >&2
  exit 2
fi

assert_contains "task guard:rest" "rest-guard job should call task guard:rest"
assert_contains "task test:go" "Go unit-test job should call task test:go"
assert_contains "task test:web" "frontend unit-test job should call task test:web"

assert_not_contains "make rest-guard" "workflow should stop calling make rest-guard"
assert_not_contains "go test ./... -short -count=1 -race" "workflow should stop calling go test directly"
assert_not_contains "cd web && bun run test" "workflow should stop calling frontend tests directly"

assert_count_at_least "      - 'Taskfile.yml'" 2 "workflow should trigger on Taskfile.yml changes for push and pull_request"
assert_not_contains "      - 'Makefile'" "workflow should stop listening to Makefile changes"

assert_count_at_least "if: failure()" 2 "smoke and e2e jobs should retain failure log branches"
assert_count_at_least "if: always()" 2 "smoke and e2e jobs should retain always-cleanup branches"
assert_count_at_least "task stack:ci:up" 2 "smoke and e2e jobs should start the CI stack via task"
assert_contains "task smoke:run" "smoke job should run smoke tests via task smoke:run"
assert_contains "task e2e:run" "e2e job should run Playwright via task e2e:run"
assert_count_at_least "task stack:ci:logs" 2 "smoke and e2e jobs should export logs via task stack:ci:logs"
assert_count_at_least "task stack:ci:down" 2 "smoke and e2e jobs should cleanup via task stack:ci:down"

assert_not_contains "docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120" "workflow should stop invoking docker compose up directly"
assert_not_contains "BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh" "workflow should stop invoking smoke_test.sh directly"
assert_not_contains "docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright" "workflow should stop invoking Playwright compose run directly"
assert_not_contains "docker compose -f docker-compose.ci.yml logs" "workflow should stop invoking docker compose logs directly"
assert_not_contains "docker compose -f docker-compose.ci.yml down --volumes" "workflow should stop invoking docker compose down directly"

if (( ${#failures[@]} > 0 )); then
  printf '\nCI workflow task migration assertions failed: %d\n' "${#failures[@]}" >&2
  for failure in "${failures[@]}"; do
    printf '  - %s\n' "$failure" >&2
  done
  exit 1
fi

printf 'All CI workflow task migration assertions passed.\n'
