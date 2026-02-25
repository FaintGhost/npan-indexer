#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Smoke tests for npan CI environment
# Expects the stack from docker-compose.ci.yml to be running.
# ---------------------------------------------------------------------------

# docker-compose.ci.yml maps container ports to 11323/19091 on host.
# Defaults are aligned so smoke tests work out-of-the-box in CI/local Docker flow.
BASE_URL="${BASE_URL:-http://localhost:11323}"
METRICS_URL="${METRICS_URL:-http://localhost:19091}"
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

assert_200_or_404_connect_not_found() {
  local code="$1" body="$2"
  if [[ "$code" == "404" ]]; then
    return 0
  fi
  if [[ "$code" == "200" ]]; then
    echo "$body" | jq -e '.code == "not_found" or .error.code == "not_found" or .message == "未找到同步进度"' > /dev/null 2>&1
    return $?
  fi
  return 1
}

assert_200_or_409_connect_abort() {
  local code="$1" body="$2"
  if [[ "$code" == "409" ]]; then
    return 0
  fi
  if [[ "$code" == "200" ]]; then
    echo "$body" | jq -e '.code == "aborted" or .error.code == "aborted" or .message == "当前没有运行中的同步任务"' > /dev/null 2>&1
    return $?
  fi
  return 1
}

assert_connect_start_sync_response() {
  local code="$1" body="$2"
  if [[ "$code" == "409" ]]; then
    return 0
  fi
  if [[ "$code" == "200" ]]; then
    if echo "$body" | jq -e 'has("message")' > /dev/null 2>&1; then
      return 0
    fi
    echo "$body" | jq -e '.code == "aborted" or .error.code == "aborted"' > /dev/null 2>&1
    return $?
  fi
  return 1
}

assert_connect_index_stats() {
  local code="$1" body="$2"
  [[ "$code" == "200" ]] \
    && echo "$body" | jq -e '.documentCount != null or . == {}' > /dev/null 2>&1
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

run_test "POST /npan.v1.AdminService/GetSyncProgress without key → 401" \
  assert_401_unauthorized \
  -X POST -H "Content-Type: application/json" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/GetSyncProgress"

run_test "POST /npan.v1.AdminService/StartSync without key → 401" \
  assert_401_unauthorized \
  -X POST -H "Content-Type: application/json" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/StartSync"

run_test "POST /npan.v1.AdminService/CancelSync without key → 401" \
  assert_401_unauthorized \
  -X POST -H "Content-Type: application/json" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/CancelSync"

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
# 8. Admin Connect lifecycle (needs API key)
# ===========================================================================
section "Admin Connect Lifecycle"

run_test "POST /npan.v1.AdminService/GetIndexStats → 200 with documentCount" \
  assert_connect_index_stats \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/GetIndexStats"

run_test "POST /npan.v1.AdminService/GetSyncProgress → 200(not_found) or 404" \
  assert_200_or_404_connect_not_found \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/GetSyncProgress"

run_test "POST /npan.v1.AdminService/CancelSync (nothing running) → 200(aborted) or 409" \
  assert_200_or_409_connect_abort \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/CancelSync"

run_test "POST /npan.v1.AdminService/StartSync → 200 or 409" \
  assert_connect_start_sync_response \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"mode":"SYNC_MODE_FULL"}' \
  "${BASE_URL}/npan.v1.AdminService/StartSync"

sleep 2

run_test "POST /npan.v1.AdminService/GetSyncProgress (after start) → 200 or 404" \
  assert_200_or_404 \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/GetSyncProgress"

run_test "POST /npan.v1.AdminService/CancelSync (after start) → 200 or 409" \
  assert_200_or_409_connect_abort \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{}' \
  "${BASE_URL}/npan.v1.AdminService/CancelSync"

sleep 2

run_test "POST /npan.v1.AdminService/InspectRoots with empty folder_ids → 400" \
  assert_400 \
  -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{"folderIds":[]}' \
  "${BASE_URL}/npan.v1.AdminService/InspectRoots"

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
