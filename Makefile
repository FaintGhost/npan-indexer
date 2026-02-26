.PHONY: rest-guard test test-frontend smoke-test e2e-test

rest-guard:
	@if rg -n "/api/v1/" \
		--glob '!docs/archive/**' \
		--glob '!docs/plans/**' \
		--glob '!tasks/**' \
		--glob '!Makefile' \
		--glob '!web/dist/**' \
		--glob '!**/*.md' \
		--glob '!**/*.yaml' \
		--glob '!**/*.yml' \
		.; then \
		echo "ERROR: runtime code must not contain REST /api/v1 paths"; \
		exit 1; \
	fi

test:
	go test ./... -short -count=1 -race

test-frontend:
	cd web && bun run test

smoke-test:
	@cleanup() { docker compose -f docker-compose.ci.yml down --volumes; }; \
	trap cleanup EXIT; \
	docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120 && \
	BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh

e2e-test:
	@cleanup() { docker compose -f docker-compose.ci.yml down --volumes; }; \
	trap cleanup EXIT; \
	docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120 && \
	BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh && \
	docker compose -f docker-compose.ci.yml run --rm --profile e2e playwright
