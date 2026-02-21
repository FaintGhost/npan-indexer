# Task 030: Implement Dockerfile

**depends-on**: (none)

## Description

Create a multi-stage Dockerfile that builds the Go binaries (server and CLI) in a build stage and copies them to a minimal Alpine runtime image. The runtime runs as a non-root user with health checks configured.

## Execution Context

**Task Number**: 030 of 032
**Phase**: Frontend & Deployment
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: No direct BDD scenario
**Architecture ref**: `../2026-02-21-production-hardening-design/architecture.md` Section 7

## Files to Modify/Create

- Create: `Dockerfile`
- Modify: `docker-compose.yml` — add npan service definition

## Steps

### Step 1: Create Dockerfile

- Create `Dockerfile` at project root with multi-stage build:
  - Stage 1 (builder): golang:1.25-alpine, install ca-certificates and tzdata, copy source, build both binaries with `-ldflags="-s -w" -trimpath`
  - Stage 2 (runtime): alpine:3.21, install ca-certificates and tzdata, create non-root user `npan`, copy binaries, copy web/ directory, create /app/data volume, set WORKDIR /app, USER npan, EXPOSE 1323
  - Add HEALTHCHECK using wget to /healthz
  - ENTRYPOINT npan-server

### Step 2: Update docker-compose.yml

- Add npan service that builds from Dockerfile
- Configure environment variables referencing env_file
- Map port 1323
- Depend on meilisearch service

### Step 3: Verify

- Build the Docker image: `docker build -t npan-server .`
- Verify image size is < 50MB
- Verify the image runs as non-root user

## Verification Commands

```bash
docker build -t npan-server .
docker image inspect npan-server --format '{{.Size}}'
docker run --rm npan-server whoami
```

## Success Criteria

- `docker build` succeeds
- Image size < 50MB
- Runs as non-root user `npan`
- Health check configured
- Both binaries (server + CLI) included
