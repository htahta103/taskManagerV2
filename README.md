# Task Manager V2

- [Product requirements](./PRD-task-manager-v2.md)
- [Architecture](./docs/ARCHITECTURE.md)
- [API spec (OpenAPI 3)](./docs/api/openapi.yaml)

## Layout

| Path | Purpose |
|------|---------|
| `cmd/api` | HTTP API entrypoint (`/healthz`, `/api/v1` stub) |
| `internal/config` | Environment-based configuration |
| `cli/` | Node.js CLI (`task` binary) for `/api/v1` tasks |
| `web/` | Vite + React + TypeScript frontend |
| `docker-compose.yml` | Local PostgreSQL 17 |
| `.env.example` | Copy to `.env` and adjust for local dev |

## Prerequisites

- Go 1.23+
- Node.js 22+ (for `web/` and `cli/`)
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

3. **CLI:**

   ```bash
   cd cli && npm install && npm run build
   ```

   Point at your API (Supabase Edge Function base URL or local API):

   ```bash
   export TASKMANAGER_API_URL=http://localhost:8080/api/v1
   export TASKMANAGER_TOKEN=   # optional Bearer JWT
   node dist/cli.js list
   node dist/cli.js search -q "keyword"
   node dist/cli.js add "Title" --due-date 2026-04-01 --focus-bucket today
   node dist/cli.js edit <task-uuid> --status todo --focus-bucket next
   ```

   Or install the `task` shim: `cd cli && npm link` (after `npm run build`).

4. **Web:**

   ```bash
   cd web && npm install && npm run dev
   ```

   Vite proxies `/api` to `http://127.0.0.1:8080`. Set `VITE_API_BASE_URL` in `web/.env.local` only if the API is on another origin (leave unset for the proxy).

   Optional UI-only auth (no backend): `VITE_MOCK_AUTH=true npm run dev`.

The SPA includes sign-in / sign-up, session restore via `GET /api/v1/me`, and the primary shell routes from the PRD (Inbox, Today, Projects, Search). Auth responses follow `docs/api/openapi.yaml`: `access_token` (or legacy `token`) plus `user`; the client stores the access token and sends `Authorization: Bearer …` while still using `credentials: "include"` for cookie-based sessions.

## CI

GitHub Actions runs `go test`, `go vet`, API build, `web` typecheck + build, CLI typecheck + build, and Playwright e2e on pushes and PRs to `main`.

## Makefile

- `make api-run` / `make api-test` — run API or Go checks
- `make cli-ci` — install, typecheck, and build the CLI
- `make web-build` — install deps and build the frontend
- `make ci` — same checks as default CI (Go + CLI + web build)
- `make ci-full` — adds e2e (Playwright) on top of `ci`

## Database migrations

SQL lives in `db/migrations/` (`001_*.sql`, `002_*.sql`, …), embedded at build time via `db/embed.go`. With `DATABASE_URL` set, `go run ./cmd/api` applies these files in lexicographic order on startup through `internal/db.Migrate`.

PostgreSQL **14+** is required (`CREATE TRIGGER … EXECUTE FUNCTION`). Migration `001_extensions.sql` enables `pgcrypto` (UUID defaults) and `pg_trgm` (ILIKE / search indexes); ensure your managed provider allows those extensions.

Manual apply (e.g. managed Postgres — set `DATABASE_URL` to your cluster):

```bash
set -euo pipefail
for f in db/migrations/*.sql; do
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done
```

Alternatively open `fly postgres connect -a <app>` and run the same files in order. Required API env: `.env.example` (`DATABASE_URL` for database-backed runs).
