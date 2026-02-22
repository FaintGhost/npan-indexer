IMPORTANT: USE BUN

## Testing

### Unit Tests

```bash
# Go backend
go test ./...

# Frontend (bun)
cd web && bun vitest run
```

### Smoke Tests (Docker)

Requires Docker. Starts meilisearch + npan containers, runs 34 API endpoint checks.

```bash
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh
docker compose -f docker-compose.ci.yml down --volumes
```

### E2E Tests (Docker + Playwright)

Requires Docker. Uses `mcr.microsoft.com/playwright` container against running services.
The `playwright` service is behind the `e2e` profile.

```bash
# Start services (if not already running from smoke tests)
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120

# Run Playwright E2E (32 tests: admin auth, sync control, search, download, edge cases)
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright

# Cleanup
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
```

Key env vars in `docker-compose.ci.yml`:
- `E2E_ADMIN_API_KEY`: admin API key for authenticated tests
- `BASE_URL`: target URL for Playwright (defaults to `http://npan:1323` inside Docker network)
- `NPA_TOKEN`: dummy token (sync will start but upstream calls fail — expected)

### CI

GitHub Actions workflow (`.github/workflows/ci.yml`) runs in order:
1. `unit-test-go` + `unit-test-frontend` + `generate-check` (parallel)
2. `smoke-test` (needs all above)
3. `e2e-test` (needs smoke-test)
