-- Task Manager V2 — extensions (PostgreSQL 17+).
-- Order: apply before domain DDL. Used by cmd/api via internal/db.Migrate and by manual psql.

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
