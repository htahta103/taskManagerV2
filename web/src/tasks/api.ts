import { ApiError, apiFetch } from "../auth/api";
import type { Tag, Task, TaskCreateBody, TaskListPage, TaskPatchBody } from "./types";

type Json = Record<string, unknown>;

async function parseJson(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function isTag(x: unknown): x is Tag {
  if (!x || typeof x !== "object") return false;
  const t = x as Json;
  return typeof t.id === "string" && typeof t.name === "string" && typeof t.created_at === "string";
}

function messageFromBody(body: unknown, fallback: string): string {
  if (body && typeof body === "object" && "error" in body) {
    const err = (body as Json).error;
    if (typeof err === "string") return err;
  }
  if (body && typeof body === "object" && "message" in body) {
    const msg = (body as Json).message;
    if (typeof msg === "string") return msg;
  }
  return fallback;
}

function parseTask(raw: unknown): Task | null {
  if (!raw || typeof raw !== "object") return null;
  const o = raw as Json;
  const id = o.id;
  const title = o.title;
  const status = o.status;
  const focus_bucket = o.focus_bucket;
  const created_at = o.created_at;
  const updated_at = o.updated_at;
  if (
    typeof id !== "string" ||
    typeof title !== "string" ||
    typeof status !== "string" ||
    typeof focus_bucket !== "string" ||
    typeof created_at !== "string" ||
    typeof updated_at !== "string"
  ) {
    return null;
  }
  return {
    id,
    title,
    description: typeof o.description === "string" ? o.description : null,
    status: status as Task["status"],
    priority:
      o.priority === null || o.priority === undefined
        ? null
        : typeof o.priority === "string"
          ? (o.priority as Task["priority"])
          : null,
    due_date:
      o.due_date === null || o.due_date === undefined
        ? null
        : typeof o.due_date === "string"
          ? o.due_date
          : null,
    focus_bucket: focus_bucket as Task["focus_bucket"],
    project_id:
      o.project_id === null || o.project_id === undefined
        ? null
        : typeof o.project_id === "string"
          ? o.project_id
          : null,
    assignee_id:
      o.assignee_id === null || o.assignee_id === undefined
        ? null
        : typeof o.assignee_id === "string"
          ? o.assignee_id
          : null,
    tags: Array.isArray(o.tags) ? (o.tags as unknown[]).filter(isTag) : undefined,
    created_at,
    updated_at,
  };
}

export async function fetchTasks(params: {
  view?: "inbox" | "today" | "next" | "later";
  project_id?: string;
  status?: string;
  q?: string;
}): Promise<TaskListPage> {
  const sp = new URLSearchParams();
  if (params.view) sp.set("view", params.view);
  if (params.project_id) sp.set("project_id", params.project_id);
  if (params.status) sp.set("status", params.status);
  if (params.q) sp.set("q", params.q);
  const qs = sp.toString();
  const path = qs ? `/tasks?${qs}` : "/tasks";
  const res = await apiFetch(path, { method: "GET" });
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Could not load tasks"), res.status, body);
  }
  if (!body || typeof body !== "object" || !Array.isArray((body as Json).items)) {
    throw new ApiError("Invalid task list response", res.status, body);
  }
  const items = (body as Json).items as unknown[];
  const tasks: Task[] = [];
  for (const item of items) {
    const t = parseTask(item);
    if (t) tasks.push(t);
  }
  const next = (body as Json).next_cursor;
  return {
    items: tasks,
    next_cursor: typeof next === "string" || next === null ? next : undefined,
  };
}

export async function createTask(body: TaskCreateBody): Promise<Task> {
  const res = await apiFetch("/tasks", { method: "POST", json: body });
  const raw = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(raw, "Could not create task"), res.status, raw);
  }
  const task = parseTask(raw);
  if (!task) throw new ApiError("Invalid task response", res.status, raw);
  return task;
}

export async function updateTask(taskId: string, patch: TaskPatchBody): Promise<Task> {
  const res = await apiFetch(`/tasks/${encodeURIComponent(taskId)}`, {
    method: "PATCH",
    json: patch,
  });
  const raw = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(raw, "Could not update task"), res.status, raw);
  }
  const task = parseTask(raw);
  if (!task) throw new ApiError("Invalid task response", res.status, raw);
  return task;
}
