-- Task Manager V2 — users, auth, projects, tags, tasks (OpenAPI + ARCHITECTURE §5).
-- Idempotent CREATE for dev bootstrap; production should apply migrations in order once per release.

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_name_nonempty CHECK (char_length(trim(name)) > 0),
    CONSTRAINT users_name_len CHECK (char_length(name) <= 120)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_idx ON users (lower(email));

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS refresh_tokens_token_hash_uidx ON refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_user_expires_idx ON refresh_tokens (user_id, expires_at);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT projects_name_nonempty CHECK (char_length(trim(name)) > 0),
    CONSTRAINT projects_name_len CHECK (char_length(name) <= 200)
);

CREATE INDEX IF NOT EXISTS projects_user_idx ON projects (user_id);
CREATE INDEX IF NOT EXISTS projects_user_updated_idx ON projects (user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tags_name_nonempty CHECK (char_length(trim(name)) > 0),
    CONSTRAINT tags_name_len CHECK (char_length(name) <= 64)
);

CREATE UNIQUE INDEX IF NOT EXISTS tags_user_name_lower_idx ON tags (user_id, lower(name));
CREATE INDEX IF NOT EXISTS tags_user_created_id_idx ON tags (user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects (id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'doing', 'done')),
    priority TEXT CHECK (priority IN ('low', 'medium', 'high')),
    due_date DATE,
    focus_bucket TEXT NOT NULL DEFAULT 'none' CHECK (focus_bucket IN ('none', 'today', 'next', 'later')),
    assignee_id UUID REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tasks_title_nonempty CHECK (char_length(trim(title)) > 0),
    CONSTRAINT tasks_title_len CHECK (char_length(title) <= 200),
    CONSTRAINT tasks_description_len CHECK (description IS NULL OR char_length(description) <= 10000)
);

CREATE INDEX IF NOT EXISTS tasks_user_status_updated_idx ON tasks (user_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS tasks_user_due_idx ON tasks (user_id, due_date) WHERE due_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS tasks_user_updated_id_idx ON tasks (user_id, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS task_tags (
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);

CREATE INDEX IF NOT EXISTS task_tags_tag_idx ON task_tags (tag_id);
