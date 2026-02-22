#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Smoke tests for npan CI environment
# Expects the stack from docker-compose.ci.yml to be running.
# ---------------------------------------------------------------------------

BASE_URL="${BASE_URL:-http://localhost:1323}"
METRICS_URL="${METRICS_URL:-http://localhost:9091}"
API_KEY="${API_KEY:-ci-test-admin-api-key-1234}"

passed=0
failed=0
results=()

RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
RESET='\033[0m'

# run_test NAME CURL_ARGS... — captures status code + body, evaluates assertion.
run_test() {
  local name="$1"; shift
  local assert_fn="$1"; shift

  local tmp
  tmp=$(mktemp)
  local http_code
  http_code=$(curl -s -o "$tmp" -w "%{http_code}" "$@") || true

  local body
  body=$(cat "$tmp")
  rm -f "$tmp"

  if $assert_fn "$http_code" "$body"; then
    results+=("${GREEN}PASS${RESET}  $name")
    ((passed++))
  else
    results+=("${RED}FAIL${RESET}  $name  (HTTP $http_code)")
    ((failed++))
  fi
}

# ---------------------------------------------------------------------------
# Assertion helpers
# ---------------------------------------------------------------------------

assert_health() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] && echo "$body" | jq -e '.status == "ok"' > /dev/null 2>&1
}

assert_200() {
  local code="$1"
  [[ "$code" == "200" ]]
}

assert_401() {
  local code="$1"
  [[ "$code" == "401" ]]
}

assert_200_or_404() {
  local code="$1"
  [[ "$code" == "200" || "$code" == "404" ]]
}

assert_search() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] \
    && echo "$body" | jq -e 'has("items") and has("total")' > /dev/null 2>&1
}

# ---------------------------------------------------------------------------
# Test cases
# ---------------------------------------------------------------------------

run_test "GET /healthz returns 200 with status ok" \
  assert_health \
  "${BASE_URL}/healthz"

run_test "GET /readyz returns 200" \
  assert_200 \
  "${BASE_URL}/readyz"

run_test "GET /api/v1/admin/sync without key returns 401" \
  assert_401 \
  "${BASE_URL}/api/v1/admin/sync"

run_test "GET /api/v1/admin/sync with key returns 200 or 404" \
  assert_200_or_404 \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/admin/sync"

run_test "GET /api/v1/app/search?q=test returns 200 with items and total" \
  assert_search \
  "${BASE_URL}/api/v1/app/search?q=test"

run_test "GET /metrics returns 200" \
  assert_200 \
  "${METRICS_URL}/metrics"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo ""
echo -e "${BOLD}=== Smoke Test Results ===${RESET}"
for r in "${results[@]}"; do
  echo -e "  $r"
done
echo ""
echo -e "  ${GREEN}Passed: ${passed}${RESET}  ${RED}Failed: ${failed}${RESET}"
echo ""

if [[ "$failed" -gt 0 ]]; then
  exit 1
fi
