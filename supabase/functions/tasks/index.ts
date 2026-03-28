import { createClient, type SupabaseClient } from "https://esm.sh/@supabase/supabase-js@2.49.1";
import { corsHeaders } from "../_shared/cors.ts";
import { escapeILikeLiteral, parseTaskListQuery } from "../_shared/query.ts";
import {
  extractTasksSuffix,
  parseTasksPathSuffix,
} from "../_shared/path.ts";

const jsonContent = "application/json";

const VALID_STATUS = new Set(["todo", "doing", "done"]);
const VALID_PRIORITY = new Set(["low", "medium", "high"]);
const DATE_RE = /^\d{4}-\d{2}-\d{2}$/;
const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function clientJsonHeaders(req: Request): Record<string, string> {
  return {
    ...corsHeaders(req),
    "Content-Type": jsonContent,
  };
}

function json(req: Request, body: unknown, status: number) {
  return new Response(JSON.stringify(body), {
    status,
    headers: clientJsonHeaders(req),
  });
}

function getSupabase(): SupabaseClient | null {
  const url = Deno.env.get("SUPABASE_URL") ?? "";
  const key =
    Deno.env.get("SUPABASE_SERVICE_ROLE_KEY") ??
    Deno.env.get("SUPABASE_ANON_KEY") ??
    "";
  if (!url || !key) return null;
  return createClient(url, key);
}

async function readJsonBody(req: Request): Promise<unknown | null> {
  const text = await req.text();
  if (!text.trim()) return {};
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

function validateTitle(
  raw: unknown,
): { ok: true; value: string } | { ok: false; error: string } {
  if (raw === undefined || raw === null) {
    return { ok: false, error: "title is required" };
  }
  if (typeof raw !== "string") {
    return { ok: false, error: "title must be a string" };
  }
  const s = raw.trim();
  if (s.length === 0) return { ok: false, error: "title is required" };
  if (s.length > 200) return { ok: false, error: "title max length is 200" };
  return { ok: true, value: s };
}

function validateDescriptionField(
  raw: unknown,
): { ok: true; value: string | null | undefined } | { ok: false; error: string } {
  if (raw === undefined) return { ok: true, value: undefined };
  if (raw === null) return { ok: true, value: null };
  if (typeof raw !== "string") {
    return { ok: false, error: "description must be a string" };
  }
  if (raw.length > 10000) {
    return { ok: false, error: "description max length is 10000" };
  }
  return { ok: true, value: raw };
}

function validateOptionalUuid(
  raw: unknown,
  field: string,
): { ok: true; value: string | null | undefined } | { ok: false; error: string } {
  if (raw === undefined) return { ok: true, value: undefined };
  if (raw === null) return { ok: true, value: null };
  if (typeof raw !== "string" || !UUID_RE.test(raw)) {
    return { ok: false, error: `invalid ${field}` };
  }
  return { ok: true, value: raw };
}

function validateOptionalStatus(
  raw: unknown,
): { ok: true; value: string | undefined } | { ok: false; error: string } {
  if (raw === undefined) return { ok: true, value: undefined };
  if (typeof raw !== "string" || !VALID_STATUS.has(raw)) {
    return { ok: false, error: "invalid status" };
  }
  return { ok: true, value: raw };
}

function validateOptionalPriority(
  raw: unknown,
): { ok: true; value: string | null | undefined } | { ok: false; error: string } {
  if (raw === undefined) return { ok: true, value: undefined };
  if (raw === null) return { ok: true, value: null };
  if (typeof raw !== "string" || !VALID_PRIORITY.has(raw)) {
    return { ok: false, error: "invalid priority" };
  }
  return { ok: true, value: raw };
}

function validateOptionalDueDate(
  raw: unknown,
): { ok: true; value: string | null | undefined } | { ok: false; error: string } {
  if (raw === undefined) return { ok: true, value: undefined };
  if (raw === null) return { ok: true, value: null };
  if (typeof raw !== "string" || !DATE_RE.test(raw)) {
    return { ok: false, error: "invalid due_date (expected YYYY-MM-DD)" };
  }
  return { ok: true, value: raw };
}

function validateTagsArray(
  raw: unknown,
): { ok: true; value: string[] | undefined } | { ok: false; error: string } {
  if (raw === undefined) return { ok: true, value: undefined };
  if (!Array.isArray(raw)) return { ok: false, error: "tags must be an array of strings" };
  for (const t of raw) {
    if (typeof t !== "string") {
      return { ok: false, error: "tags must be an array of strings" };
    }
  }
  return { ok: true, value: raw as string[] };
}

async function listTasks(req: Request, supabase: SupabaseClient, url: URL) {
  const parsed = parseTaskListQuery(url);
  if (!parsed.ok) {
    return json(req, { error: parsed.error, code: "validation_error" }, 422);
  }

  let q = supabase
    .from("tasks")
    .select("*", { count: "exact" })
    .order("created_at", { ascending: false });

  const { status, priority, search } = parsed.value;
  if (status) q = q.eq("status", status);
  if (priority) q = q.eq("priority", priority);
  if (search) {
    const pattern = `%${escapeILikeLiteral(search)}%`;
    q = q.or(`title.ilike.${pattern},description.ilike.${pattern}`);
  }

  const { data, error, count } = await q;
  if (error) return json(req, { error: error.message }, 500);
  return json(req, { data: data ?? [], count: count ?? 0 }, 200);
}

async function createTask(req: Request, supabase: SupabaseClient) {
  const body = await readJsonBody(req);
  if (body === null) {
    return json(req, { error: "Invalid JSON body", code: "validation_error" }, 400);
  }
  if (typeof body !== "object" || body === null || Array.isArray(body)) {
    return json(req, { error: "Body must be a JSON object", code: "validation_error" }, 400);
  }
  const o = body as Record<string, unknown>;

  const titleR = validateTitle(o.title);
  if (!titleR.ok) {
    return json(req, { error: titleR.error, code: "validation_error" }, 422);
  }
  const descR = validateDescriptionField(o.description);
  if (!descR.ok) {
    return json(req, { error: descR.error, code: "validation_error" }, 422);
  }
  const statusR = validateOptionalStatus(o.status);
  if (!statusR.ok) {
    return json(req, { error: statusR.error, code: "validation_error" }, 422);
  }
  const priR = validateOptionalPriority(o.priority);
  if (!priR.ok) {
    return json(req, { error: priR.error, code: "validation_error" }, 422);
  }
  const dueR = validateOptionalDueDate(o.due_date);
  if (!dueR.ok) {
    return json(req, { error: dueR.error, code: "validation_error" }, 422);
  }
  const projR = validateOptionalUuid(o.project_id, "project_id");
  if (!projR.ok) {
    return json(req, { error: projR.error, code: "validation_error" }, 422);
  }
  const assignR = validateOptionalUuid(o.assignee_id, "assignee_id");
  if (!assignR.ok) {
    return json(req, { error: assignR.error, code: "validation_error" }, 422);
  }
  const tagsR = validateTagsArray(o.tags);
  if (!tagsR.ok) {
    return json(req, { error: tagsR.error, code: "validation_error" }, 422);
  }

  const insert: Record<string, unknown> = { title: titleR.value };
  if (descR.value !== undefined) insert.description = descR.value;
  if (statusR.value !== undefined) insert.status = statusR.value;
  if (priR.value !== undefined) insert.priority = priR.value;
  if (dueR.value !== undefined) insert.due_date = dueR.value;
  if (projR.value !== undefined) insert.project_id = projR.value;
  if (assignR.value !== undefined) insert.assignee_id = assignR.value;
  if (tagsR.value !== undefined) insert.tags = tagsR.value;

  const { data, error } = await supabase.from("tasks").insert(insert).select().single();
  if (error) {
    return json(req, { error: error.message, code: "database_error" }, 500);
  }
  return json(req, { data }, 201);
}

async function getTask(req: Request, supabase: SupabaseClient, id: string) {
  const { data, error } = await supabase.from("tasks").select("*").eq("id", id).maybeSingle();
  if (error) return json(req, { error: error.message }, 500);
  if (!data) {
    return json(req, { error: "Task not found", code: "not_found" }, 404);
  }
  return json(req, { data }, 200);
}

async function patchTask(req: Request, supabase: SupabaseClient, id: string) {
  const body = await readJsonBody(req);
  if (body === null) {
    return json(req, { error: "Invalid JSON body", code: "validation_error" }, 400);
  }
  if (typeof body !== "object" || body === null || Array.isArray(body)) {
    return json(req, { error: "Body must be a JSON object", code: "validation_error" }, 400);
  }
  const o = body as Record<string, unknown>;

  const patch: Record<string, unknown> = {};

  if ("title" in o) {
    const titleR = validateTitle(o.title);
    if (!titleR.ok) {
      return json(req, { error: titleR.error, code: "validation_error" }, 422);
    }
    patch.title = titleR.value;
  }
  if ("description" in o) {
    const descR = validateDescriptionField(o.description);
    if (!descR.ok) {
      return json(req, { error: descR.error, code: "validation_error" }, 422);
    }
    patch.description = descR.value;
  }
  if ("status" in o) {
    const statusR = validateOptionalStatus(o.status);
    if (!statusR.ok || statusR.value === undefined) {
      return json(
        req,
        { error: statusR.ok ? "status cannot be undefined" : statusR.error, code: "validation_error" },
        422,
      );
    }
    patch.status = statusR.value;
  }
  if ("priority" in o) {
    const priR = validateOptionalPriority(o.priority);
    if (!priR.ok) {
      return json(req, { error: priR.error, code: "validation_error" }, 422);
    }
    patch.priority = priR.value;
  }
  if ("due_date" in o) {
    const dueR = validateOptionalDueDate(o.due_date);
    if (!dueR.ok) {
      return json(req, { error: dueR.error, code: "validation_error" }, 422);
    }
    patch.due_date = dueR.value;
  }
  if ("project_id" in o) {
    const projR = validateOptionalUuid(o.project_id, "project_id");
    if (!projR.ok) {
      return json(req, { error: projR.error, code: "validation_error" }, 422);
    }
    patch.project_id = projR.value;
  }
  if ("assignee_id" in o) {
    const assignR = validateOptionalUuid(o.assignee_id, "assignee_id");
    if (!assignR.ok) {
      return json(req, { error: assignR.error, code: "validation_error" }, 422);
    }
    patch.assignee_id = assignR.value;
  }
  if ("tags" in o) {
    const tagsR = validateTagsArray(o.tags);
    if (!tagsR.ok || tagsR.value === undefined) {
      return json(
        req,
        { error: tagsR.ok ? "tags cannot be undefined" : tagsR.error, code: "validation_error" },
        422,
      );
    }
    patch.tags = tagsR.value;
  }

  if (Object.keys(patch).length === 0) {
    return json(req, { error: "No fields to update", code: "validation_error" }, 422);
  }

  const { data, error } = await supabase.from("tasks").update(patch).eq("id", id).select().maybeSingle();
  if (error) return json(req, { error: error.message }, 500);
  if (!data) {
    return json(req, { error: "Task not found", code: "not_found" }, 404);
  }
  return json(req, { data }, 200);
}

async function deleteTask(req: Request, supabase: SupabaseClient, id: string) {
  const { data, error } = await supabase.from("tasks").delete().eq("id", id).select("id").maybeSingle();
  if (error) return json(req, { error: error.message }, 500);
  if (!data) {
    return json(req, { error: "Task not found", code: "not_found" }, 404);
  }
  return new Response(null, { status: 204, headers: corsHeaders(req) });
}

async function clearDoneTasks(req: Request, supabase: SupabaseClient) {
  const { error, count } = await supabase
    .from("tasks")
    .delete({ count: "exact" })
    .eq("status", "done");

  if (error) return json(req, { error: error.message }, 500);
  return json(req, { data: { deleted: count ?? 0 } }, 200);
}

Deno.serve(async (req: Request) => {
  if (req.method === "OPTIONS") {
    return new Response(null, { status: 204, headers: corsHeaders(req) });
  }

  const supabase = getSupabase();
  if (!supabase) {
    return json(
      req,
      { error: "Server misconfigured: missing SUPABASE_URL and key" },
      500,
    );
  }

  const url = new URL(req.url);
  const suffix = extractTasksSuffix(url.pathname);
  const route = parseTasksPathSuffix(suffix);

  if (route.kind === "not_found") {
    return json(req, { error: "Not found", code: "not_found" }, 404);
  }

  if (route.kind === "collection") {
    if (req.method === "GET") return listTasks(req, supabase, url);
    if (req.method === "POST") return createTask(req, supabase);
    return json(req, { error: "Method not allowed" }, 405);
  }

  if (route.kind === "clear_done") {
    if (req.method === "DELETE") return clearDoneTasks(req, supabase);
    return json(req, { error: "Method not allowed" }, 405);
  }

  const id = route.id;
  if (req.method === "GET") return getTask(req, supabase, id);
  if (req.method === "PATCH") return patchTask(req, supabase, id);
  if (req.method === "DELETE") return deleteTask(req, supabase, id);
  return json(req, { error: "Method not allowed" }, 405);
});
