# Deployment: API, database, and web

This rig ships a single Go HTTP process (`cmd/api`), PostgreSQL as the system of record, and a Vite + React SPA in `web/`. Use this page for **production** secrets, **TLS**, and **health checks**. For system design, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Environment variables (API)

| Variable | Required | Purpose |
|----------|----------|---------|
| `PORT` | No | Listen port (default `8080`). |
| `APP_ENV` | No | `development` (relaxed defaults) vs production-style settings; set `production` in prod. |
| `DATABASE_URL` | Yes (full stack) | PostgreSQL DSN, e.g. `postgres://user:pass@host:5432/dbname?sslmode=require`. Migrations in `db/migrations/` run on startup. |
| `JWT_SECRET` | Yes | **HMAC-SHA256** secret for signing access/refresh tokens. Use a long, high-entropy value (32+ random bytes); rotate with a coordinated token invalidation plan. |
| `JWT_ACCESS_TTL_MINUTES` | No | Access token lifetime (default 15). |
| `JWT_REFRESH_TTL_DAYS` | No | Refresh token lifetime (default 30). |
| `CORS_ORIGIN` | Recommended in prod | Exact browser `Origin` allowed for credentialed requests (e.g. `https://app.example.com`). Must match your **HTTPS** SPA origin. |
| `AUTH_RATE_LIMIT_PER_MINUTE` | No | Per-IP cap on `POST /api/v1/auth/register`, `/login`, and `/refresh`. If unset: no limit in `development`; with `APP_ENV=production`, defaults to **60**. Set `0` to disable in production. |

Copy [`.env.example`](../.env.example) for local values; **never** commit real secrets.

### Web and CLI (not containerized by default)

- **Web**: build static assets (`web/`) and serve behind your CDN or object storage + edge. Set `VITE_API_BASE_URL` (or your project’s equivalent) to the **public HTTPS API root** at build time.
- **CLI**: `TASKMANAGER_API_URL` should point at `https://<api-host>/api/v1` with optional `TASKMANAGER_TOKEN`.

## HTTPS (production)

Terminate TLS **in front of** the API (load balancer, reverse proxy, or platform ingress). The Go server listens for plain HTTP on `PORT`; it does not terminate TLS itself.

- Redirect HTTP → HTTPS at the edge.
- Set `CORS_ORIGIN` to the **HTTPS** SPA origin only.
- Use `sslmode=require` (or stricter) in `DATABASE_URL` for managed Postgres.

## Health and readiness

Paths are on the **root** of the API host (not under `/api/v1`), per [OpenAPI](./api/openapi.yaml).

| Path | Role | Success | Notes |
|------|------|---------|--------|
| `GET /healthz` | Liveness | `200` + `{"status":"ok"}` | Process is up; use for “restart if failing” checks. |
| `GET /readyz` | Readiness | `200` empty body when DB ping succeeds; `503` if DB unreachable | Use before routing traffic; fails when Postgres is down. |

**Stub mode:** If `DATABASE_URL` is unset, the process serves a minimal mux with `/healthz` only (no DB-backed routes). In that mode there is no `/readyz`.

Configure your orchestrator (Kubernetes, Fly, ECS, etc.) with:

- **Liveness**: HTTP GET `/healthz` on the container/service port.
- **Readiness**: HTTP GET `/readyz` on the same port, expect `200` when Postgres is healthy.

## Docker

- **Image**: `Dockerfile` multi-stage build → static `api` binary on Alpine.
- **Compose**: [`docker-compose.yml`](../docker-compose.yml) runs **Postgres + API** for local smoke. Example:

  ```bash
  docker compose up --build
  curl -s http://localhost:8080/healthz
  curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/readyz
  ```

  For the SPA, run `cd web && npm run dev` on the host (default proxy to `localhost:8080`) or deploy built assets separately.

## Database migrations

With `DATABASE_URL` set, `cmd/api` applies `db/migrations/*.sql` in order at startup. For managed databases you may also run the same files with `psql` (see [README.md](../README.md)).
