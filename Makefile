.PHONY: generate-go generate-ts generate generate-check test test-frontend smoke-test

generate-go:
	go generate ./api/...

generate-ts:
	cd web && bun run generate

generate: generate-go generate-ts

generate-check: generate
	@git diff --exit-code -- api/types.gen.go web/src/api/generated/ \
		|| (echo "ERROR: generated files are out of date. Run 'make generate' and commit the changes." && exit 1)

test:
	go test ./... -short -count=1 -race

test-frontend:
	cd web && bun run test

smoke-test:
	@cleanup() { docker compose -f docker-compose.ci.yml down --volumes; }; \
	trap cleanup EXIT; \
	docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120 && \
	./tests/smoke/smoke_test.sh
