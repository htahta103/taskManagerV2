# Task Manager V2

- [Product requirements](./PRD-task-manager-v2.md)
- [Architecture](./docs/ARCHITECTURE.md)
- [API spec (OpenAPI 3)](./docs/api/openapi.yaml)

## Layout

| Path | Purpose |
|------|---------|
| `cmd/api` | HTTP API entrypoint (`/healthz`, `/api/v1` stub) |
| `internal/config` | Environment-based configuration |
| `web/` | Vite + React + TypeScript frontend |
| `docker-compose.yml` | Local PostgreSQL 17 |
| `.env.example` | Copy to `.env` and adjust for local dev |

## Prerequisites

- Go 1.23+
- Node.js 22+ (for `web/`)
- Docker (optional, for Postgres)

## Local development

1. **Database (optional for scaffold):**

   ```bash
   docker compose up -d postgres
   cp .env.example .env
   ```

2. **API:**

   ```bash
   go run ./cmd/api
   ```

   - Health: `GET http://localhost:8080/healthz`
   - API stub: `GET http://localhost:8080/api/v1`

3. **Web:**

   ```bash
   cd web && npm install && npm run dev
   ```

   Set `VITE_API_BASE_URL` in `web/.env.local` if the API is not on `http://localhost:8080`.

## CI

GitHub Actions runs `go test`, `go vet`, API build, and `web` typecheck + production build on pushes and PRs to `main`.

## Makefile

- `make api-run` / `make api-test` — run API or Go checks
- `make web-build` — install deps and build the frontend
- `make ci` — same checks as CI (Go + web build)
