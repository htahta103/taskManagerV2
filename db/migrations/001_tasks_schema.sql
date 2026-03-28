-- Task Manager V2: core tasks schema (Fly Postgres / PostgreSQL 14+)
-- Applies: pgcrypto, task_status / task_priority enums, tasks table, updated_at trigger.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE task_status AS ENUM ('todo', 'doing', 'done');

CREATE TYPE task_priority AS ENUM ('low', 'medium', 'high');

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status task_status NOT NULL DEFAULT 'todo',
    priority task_priority,
    due_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id UUID,
    tags TEXT[] NOT NULL DEFAULT '{}',
    assignee_id UUID,
    CONSTRAINT tasks_title_not_empty CHECK (char_length(trim(title)) > 0),
    CONSTRAINT tasks_description_len CHECK (
        description IS NULL OR char_length(description) <= 10000
    )
);

CREATE OR REPLACE FUNCTION tasks_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION tasks_set_updated_at();
