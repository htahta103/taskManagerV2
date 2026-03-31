# Architecture Requirements: Task Manager V2

> **Superseded (historical)** — This file described an early Supabase Edge–centric plan. The **current** architecture (Go API, PostgreSQL, Vite web, auth model, and deployment shape) is documented in **[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)** and **[docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md)**. Keep this file for PRD-era context only.

## Core Domain Entities and Relationships

- **Task**
  - Fields: `id`, `title`, `description`, `status`, `priority`, `due_date`, `created_at`, `updated_at`
  - Status lifecycle: `pending` -> `in_progress` -> `done`
  - Priority values: `low`, `medium`, `high`
- **Task Collection**
  - Logical aggregate returned by list APIs with optional filters (`status`, `priority`, `search`)
- **Client Channels**
  - Web app and CLI both operate on the same Task domain through shared edge endpoints

Relationship model:
- One persistent `tasks` table stores all task records.
- Edge function handlers are the application boundary between clients and database.
- Frontend and CLI are independent clients of the same API contract.

## Key User Flows

1. **Create Task**
   - User submits title (+ optional description, priority, due date).
   - API validates payload and inserts row in PostgreSQL.
   - Client refreshes and displays created task.
2. **List and Find Tasks**
   - User opens dashboard or runs CLI list command.
   - API returns tasks sorted by newest first.
   - Optional filtering by status/priority/search narrows results.
3. **Update Task**
   - User edits task fields or updates status to in progress/done.
   - API applies partial update and returns latest object.
4. **Delete Task**
   - User deletes single task by id; API removes row.
5. **Clear Completed**
   - User triggers bulk cleanup; API removes rows where status is done and returns deleted count.
6. **Health Verification**
   - Monitoring or clients call health endpoint to verify API availability.

## Non-Functional Requirements

- **Scale**
  - Use Supabase free-tier constraints as baseline (function invocation and compute limits).
  - Keep handlers stateless and lightweight to support horizontal edge execution.
- **Latency**
  - Prefer co-located compute with Supabase Postgres to minimize API-to-DB round trips.
  - Keep list filtering server-side to reduce payload size and client-side processing.
- **Availability**
  - Health endpoint must return stable 200 responses for monitoring checks.
  - All endpoints should fail with explicit JSON errors, not opaque runtime failures.
- **Security / Auth Model**
  - v1 is single-tenant with no end-user auth in scope.
  - Service role key is stored only in Supabase function secrets.
  - Public clients (web/CLI) call edge endpoints using anon key where required.
  - CORS must be explicitly managed for browser clients.
- **Reliability and Maintainability**
  - Consistent JSON response and error shapes across handlers.
  - Shared validation/CORS helpers to avoid drift.
  - `updated_at` auto-maintained via DB trigger for auditability.

## Preferred Tech Stack

- **Backend:** Supabase Edge Functions (Deno + TypeScript), Supabase PostgreSQL
- **Frontend:** React 18 + Vite + Tailwind CSS, deployed on Cloudflare Pages (`htahta103`)
- **CLI:** Node.js command-line client consuming the same edge API
- **Deployment Tooling:** Supabase CLI for function deploy and secret management
