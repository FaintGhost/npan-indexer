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
YELLOW='\033[0;33m'
BOLD='\033[1m'
RESET='\033[0m'

# run_test NAME ASSERT_FN CURL_ARGS...
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
    ((passed++)) || true
  else
    results+=("${RED}FAIL${RESET}  $name  (HTTP $http_code) $body")
    ((failed++)) || true
  fi
}

# ---------------------------------------------------------------------------
# Assertion helpers
# ---------------------------------------------------------------------------

assert_200() {
  local code="$1"
  [[ "$code" == "200" ]]
}

assert_202() {
  local code="$1"
  [[ "$code" == "202" ]]
}

assert_400() {
  local code="$1"
  [[ "$code" == "400" ]]
}

assert_401() {
  local code="$1"
  [[ "$code" == "401" ]]
}

assert_409() {
  local code="$1"
  [[ "$code" == "409" ]]
}

assert_200_or_404() {
  local code="$1"
  [[ "$code" == "200" || "$code" == "404" ]]
}

assert_200_or_409() {
  local code="$1"
  [[ "$code" == "200" || "$code" == "409" ]]
}

assert_202_or_409() {
  local code="$1"
  [[ "$code" == "202" || "$code" == "409" ]]
}

assert_health() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] && echo "$body" | jq -e '.status == "ok"' > /dev/null 2>&1
}

assert_ready() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] && echo "$body" | jq -e '.status == "ready"' > /dev/null 2>&1
}

assert_search_result() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] \
    && echo "$body" | jq -e 'has("items") and has("total")' > /dev/null 2>&1
}

assert_error_code() {
  local expected_code="$1"
  shift
  local code="$1" body="$2"
  echo "$body" | jq -e ".code == \"$expected_code\"" > /dev/null 2>&1
}

assert_401_unauthorized() {
  local code="$1" body="$2"
  [[ "$code" == "401" ]] \
    && echo "$body" | jq -e '.code == "UNAUTHORIZED"' > /dev/null 2>&1
}

assert_400_bad_request() {
  local code="$1" body="$2"
  [[ "$code" == "400" ]] \
    && echo "$body" | jq -e '.code == "BAD_REQUEST"' > /dev/null 2>&1
}

assert_sync_progress() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] \
    && echo "$body" | jq -e 'has("status") and has("aggregateStats")' > /dev/null 2>&1
}

assert_message() {
  local code="$1" body="$2"
  echo "$body" | jq -e 'has("message")' > /dev/null 2>&1
}

assert_download_url_error() {
  local code="$1" body="$2"
  # App download-url with dummy token returns 503 (token parse fail) or 502 (upstream fail)
  [[ "$code" == "502" || "$code" == "503" ]]
}

assert_admin_download_url_error() {
  local code="$1" body="$2"
  # Admin download-url with dummy token returns 400
  [[ "$code" == "400" ]]
}

# ---------------------------------------------------------------------------
# Section helper
# ---------------------------------------------------------------------------

section() {
  results+=("" "${YELLOW}${BOLD}--- $1 ---${RESET}")
}

# ===========================================================================
# 1. Health & Readiness
# ===========================================================================
section "Health & Readiness"

run_test "GET /healthz → 200 with status=ok" \
  assert_health \
  "${BASE_URL}/healthz"

run_test "GET /readyz → 200 with status=ready" \
  assert_ready \
  "${BASE_URL}/readyz"

# ===========================================================================
# 2. App endpoints (no API key required, EmbeddedAuth)
# ===========================================================================
section "App Endpoints (public)"

run_test "GET /api/v1/app/search?q=test → 200 with items+total" \
  assert_search_result \
  "${BASE_URL}/api/v1/app/search?q=test"

run_test "GET /api/v1/app/search without query → 400" \
  assert_400 \
  "${BASE_URL}/api/v1/app/search"

run_test "GET /api/v1/app/search with pagination → 200" \
  assert_search_result \
  "${BASE_URL}/api/v1/app/search?q=test&page=1&page_size=5"

run_test "GET /api/v1/app/download-url without file_id → 400" \
  assert_400 \
  "${BASE_URL}/api/v1/app/download-url"

run_test "GET /api/v1/app/download-url?file_id=999 → 502 or 503 (dummy token)" \
  assert_download_url_error \
  "${BASE_URL}/api/v1/app/download-url?file_id=999"

# ===========================================================================
# 3. Auth boundary tests (401 without API key)
# ===========================================================================
section "Auth Boundary (401 without API key)"

run_test "POST /api/v1/token without key → 401" \
  assert_401_unauthorized \
  -X POST -H "Content-Type: application/json" \
  -d '{}' \
  "${BASE_URL}/api/v1/token"

run_test "GET /api/v1/search/remote without key → 401" \
  assert_401_unauthorized \
  "${BASE_URL}/api/v1/search/remote?query=test"

run_test "GET /api/v1/search/local without key → 401" \
  assert_401_unauthorized \
  "${BASE_URL}/api/v1/search/local?query=test"

run_test "GET /api/v1/download-url without key → 401" \
  assert_401_unauthorized \
  "${BASE_URL}/api/v1/download-url?file_id=1"

run_test "GET /api/v1/admin/sync without key → 401" \
  assert_401_unauthorized \
  "${BASE_URL}/api/v1/admin/sync"

run_test "POST /api/v1/admin/sync without key → 401" \
  assert_401_unauthorized \
  -X POST -H "Content-Type: application/json" \
  -d '{}' \
  "${BASE_URL}/api/v1/admin/sync"

run_test "DELETE /api/v1/admin/sync without key → 401" \
  assert_401_unauthorized \
  -X DELETE \
  "${BASE_URL}/api/v1/admin/sync"

# ===========================================================================
# 4. Token endpoint
# ===========================================================================
section "Token (POST /api/v1/token)"

run_test "POST /api/v1/token with key but empty body → 400 missing params" \
  assert_400_bad_request \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/api/v1/token"

run_test "POST /api/v1/token missing sub_id → 400" \
  assert_400_bad_request \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"client_id":"test","client_secret":"test"}' \
  "${BASE_URL}/api/v1/token"

run_test "POST /api/v1/token missing client_secret → 400" \
  assert_400_bad_request \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"client_id":"test","sub_id":1}' \
  "${BASE_URL}/api/v1/token"

run_test "POST /api/v1/token with all params → 400 (upstream unreachable)" \
  assert_400_bad_request \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"client_id":"ci-id","client_secret":"ci-secret","sub_id":1}' \
  "${BASE_URL}/api/v1/token"

# Also test Bearer auth header
run_test "POST /api/v1/token via Bearer auth → 400 (param validation works)" \
  assert_400_bad_request \
  -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/api/v1/token"

# ===========================================================================
# 5. Search — remote (needs API key + upstream token)
# ===========================================================================
section "Remote Search (GET /api/v1/search/remote)"

run_test "GET /api/v1/search/remote without query → 400" \
  assert_400_bad_request \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/search/remote"

run_test "GET /api/v1/search/remote?query=test → 400 (dummy token fails upstream)" \
  assert_400 \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/search/remote?query=test"

# ===========================================================================
# 6. Search — local (needs API key, queries Meilisearch directly)
# ===========================================================================
section "Local Search (GET /api/v1/search/local)"

run_test "GET /api/v1/search/local without query → 400" \
  assert_400_bad_request \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/search/local"

run_test "GET /api/v1/search/local?query=test → 200 with items+total" \
  assert_search_result \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/search/local?query=test"

run_test "GET /api/v1/search/local with filters → 200" \
  assert_search_result \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/search/local?query=test&page=1&page_size=10&type=file"

run_test "GET /api/v1/search/local?q=test (alias) → 200" \
  assert_search_result \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/search/local?q=test"

# ===========================================================================
# 7. Download URL — admin (needs API key)
# ===========================================================================
section "Download URL (GET /api/v1/download-url)"

run_test "GET /api/v1/download-url without file_id → 400" \
  assert_400 \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/download-url"

run_test "GET /api/v1/download-url?file_id=999 → 400 (dummy token)" \
  assert_admin_download_url_error \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/download-url?file_id=999"

# ===========================================================================
# 8. Sync lifecycle (needs API key)
# ===========================================================================
section "Sync Lifecycle (POST/GET/DELETE /api/v1/admin/sync)"

# GET progress — initially no sync has run
run_test "GET /api/v1/admin/sync → 200 or 404 (no prior sync)" \
  assert_200_or_404 \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/admin/sync"

# DELETE when nothing is running → 409
run_test "DELETE /api/v1/admin/sync (nothing running) → 409 conflict" \
  assert_409 \
  -X DELETE \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/admin/sync"

# Start sync — will be accepted (202) even though upstream is dummy
run_test "POST /api/v1/admin/sync → 202 accepted" \
  assert_202 \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"mode":"full"}' \
  "${BASE_URL}/api/v1/admin/sync"

# Brief pause to let sync register progress
sleep 2

# GET progress — with dummy token sync may fail before writing progress (404) or succeed (200)
run_test "GET /api/v1/admin/sync (after start) → 200 or 404" \
  assert_200_or_404 \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/admin/sync"

# Cancel sync — 200 if still running, 409 if already finished/failed
run_test "DELETE /api/v1/admin/sync (cancel) → 200 or 409" \
  assert_200_or_409 \
  -X DELETE \
  -H "X-API-Key: ${API_KEY}" \
  "${BASE_URL}/api/v1/admin/sync"

# Wait for previous sync to fully stop
sleep 2

# Start again — 202 if accepted, 409 if previous sync still winding down
run_test "POST /api/v1/admin/sync again → 202 or 409" \
  assert_202_or_409 \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"mode":"full"}' \
  "${BASE_URL}/api/v1/admin/sync"

# ===========================================================================
# 9. Metrics
# ===========================================================================
section "Metrics"

run_test "GET /metrics → 200" \
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
