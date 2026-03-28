-- Task Manager: tasks table, enums, updated_at trigger, RLS (public v1).
-- Apply via Supabase CLI (`supabase db push`) or SQL editor for project nfbquxfnsprwjsehhnvq.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
    CREATE TYPE public.task_status AS ENUM ('todo', 'doing', 'done');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_priority') THEN
    CREATE TYPE public.task_priority AS ENUM ('low', 'medium', 'high');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS public.tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(200) NOT NULL,
  description TEXT,
  status public.task_status NOT NULL DEFAULT 'todo',
  priority public.task_priority,
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

CREATE OR REPLACE FUNCTION public.tasks_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tasks_updated_at ON public.tasks;
CREATE TRIGGER tasks_updated_at
  BEFORE UPDATE ON public.tasks
  FOR EACH ROW
  EXECUTE PROCEDURE public.tasks_set_updated_at();

ALTER TABLE public.tasks ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tasks_public_v1_select ON public.tasks;
DROP POLICY IF EXISTS tasks_public_v1_insert ON public.tasks;
DROP POLICY IF EXISTS tasks_public_v1_update ON public.tasks;
DROP POLICY IF EXISTS tasks_public_v1_delete ON public.tasks;

CREATE POLICY tasks_public_v1_select
  ON public.tasks FOR SELECT
  TO anon, authenticated
  USING (true);

CREATE POLICY tasks_public_v1_insert
  ON public.tasks FOR INSERT
  TO anon, authenticated
  WITH CHECK (true);

CREATE POLICY tasks_public_v1_update
  ON public.tasks FOR UPDATE
  TO anon, authenticated
  USING (true)
  WITH CHECK (true);

CREATE POLICY tasks_public_v1_delete
  ON public.tasks FOR DELETE
  TO anon, authenticated
  USING (true);

GRANT SELECT, INSERT, UPDATE, DELETE ON public.tasks TO anon, authenticated;
GRANT USAGE ON TYPE public.task_status TO anon, authenticated;
GRANT USAGE ON TYPE public.task_priority TO anon, authenticated;
