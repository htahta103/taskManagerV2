-- updated_at triggers, cross-table tenant guards, and search indexes (PRD list/search performance).

CREATE OR REPLACE FUNCTION tm_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION tm_set_updated_at();

DROP TRIGGER IF EXISTS projects_updated_at ON projects;
CREATE TRIGGER projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION tm_set_updated_at();

DROP TRIGGER IF EXISTS tasks_updated_at ON tasks;
CREATE TRIGGER tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION tm_set_updated_at();

CREATE OR REPLACE FUNCTION tasks_enforce_project_owner()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.project_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = NEW.project_id AND p.user_id = NEW.user_id
        ) THEN
            RAISE EXCEPTION 'tasks: project_id must reference a project owned by user_id';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tasks_project_owner ON tasks;
CREATE TRIGGER tasks_project_owner
    BEFORE INSERT OR UPDATE OF project_id, user_id ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION tasks_enforce_project_owner();

CREATE OR REPLACE FUNCTION tasks_enforce_assignee_self()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.assignee_id IS NOT NULL AND NEW.assignee_id IS DISTINCT FROM NEW.user_id THEN
        RAISE EXCEPTION 'tasks: assignee_id must equal user_id (MVP single-user scope)';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tasks_assignee_self ON tasks;
CREATE TRIGGER tasks_assignee_self
    BEFORE INSERT OR UPDATE OF assignee_id, user_id ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION tasks_enforce_assignee_self();

CREATE OR REPLACE FUNCTION task_tags_enforce_same_user()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM tasks t
        INNER JOIN tags g ON g.id = NEW.tag_id
        WHERE t.id = NEW.task_id AND t.user_id = g.user_id
    ) THEN
        RAISE EXCEPTION 'task_tags: task and tag must belong to the same user';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS task_tags_same_user ON task_tags;
CREATE TRIGGER task_tags_same_user
    BEFORE INSERT OR UPDATE ON task_tags
    FOR EACH ROW
    EXECUTE FUNCTION task_tags_enforce_same_user();

-- ILIKE search on title / description (store.ListTasks with Q); gin_trgm assists ~1k rows/user (ARCHITECTURE §5.3).
CREATE INDEX IF NOT EXISTS tasks_title_trgm_idx ON tasks USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS tasks_description_trgm_idx ON tasks USING gin (description gin_trgm_ops);
