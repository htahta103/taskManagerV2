import { ApiError, apiFetch } from "../auth/api";
import type { Task, TaskPriority, TaskStatus } from "./types";

type ApiTag = { id: string; name: string };

export type ApiTaskRow = {
  id: string;
  title: string;
  description: string | null;
  status: string;
  priority: string | null;
  due_date: string | null;
  focus_bucket: string;
  created_at: string;
  updated_at: string;
  tags: ApiTag[];
};

function parseJson(res: Response): Promise<unknown> {
  return res.text().then((t) => {
    if (!t) return null;
    try {
      return JSON.parse(t) as unknown;
    } catch {
      return t;
    }
  });
}

function messageFromBody(body: unknown, fallback: string): string {
  if (body && typeof body === "object" && "error" in body) {
    const err = (body as Record<string, unknown>).error;
    if (typeof err === "string") return err;
  }
  return fallback;
}

function asTaskStatus(s: string): TaskStatus {
  if (s === "todo" || s === "doing" || s === "done") return s;
  return "todo";
}

function asTaskPriority(s: string | null | undefined): TaskPriority | undefined {
  if (s === "low" || s === "medium" || s === "high") return s;
  return undefined;
}

function asFocusBucket(s: string): NonNullable<Task["focusBucket"]> {
  if (s === "none" || s === "today" || s === "next" || s === "later") return s;
  return "none";
}

/** `pinnedToday` draft + prior `focus_bucket` → PATCH `focus_bucket`. */
export function resolvedFocusBucketForApi(merged: Task): string {
  if (merged.pinnedToday) return "today";
  if (merged.focusBucket && merged.focusBucket !== "today") return merged.focusBucket;
  return "none";
}

export function taskToApiUpdateBody(merged: Task): Record<string, unknown> {
  return {
    title: merged.title.trim(),
    description: merged.description === "" ? null : merged.description,
    status: merged.status,
    priority: merged.priority ?? null,
    due_date: merged.dueDate ?? null,
    focus_bucket: resolvedFocusBucketForApi(merged),
  };
}

/** Build a PATCH body from a task plus only the fields in `partial` (for row actions). */
export function patchToApi(
  task: Task,
  partial: Partial<Omit<Task, "id" | "createdAt">>,
): Record<string, unknown> {
  const merged: Task = { ...task, ...partial };
  const body: Record<string, unknown> = {};
  if (partial.title !== undefined) {
    const t = merged.title.trim();
    if (t) body.title = t;
  }
  if (partial.description !== undefined) {
    body.description = merged.description === "" ? null : merged.description;
  }
  if (partial.status !== undefined) body.status = merged.status;
  if (partial.priority !== undefined) body.priority = merged.priority ?? null;
  if (partial.dueDate !== undefined) body.due_date = merged.dueDate ?? null;
  if (partial.pinnedToday !== undefined || partial.focusBucket !== undefined) {
    body.focus_bucket = resolvedFocusBucketForApi(merged);
  }
  return body;
}

export function apiTaskToTask(row: ApiTaskRow): Task {
  const fb = asFocusBucket(row.focus_bucket);
  return {
    id: row.id,
    title: row.title,
    description: row.description ?? "",
    status: asTaskStatus(row.status),
    priority: asTaskPriority(row.priority),
    dueDate: row.due_date ?? undefined,
    pinnedToday: fb === "today",
    focusBucket: fb,
    tags: Array.isArray(row.tags) ? row.tags.map((t) => t.name) : [],
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
}

export async function fetchTasksFromApi(): Promise<Task[]> {
  const res = await apiFetch("/tasks", { method: "GET" });
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Could not load tasks"), res.status, body);
  }
  if (!body || typeof body !== "object" || !("items" in body)) {
    throw new ApiError("Invalid tasks response", res.status, body);
  }
  const items = (body as { items: unknown }).items;
  if (!Array.isArray(items)) throw new ApiError("Invalid tasks list", res.status, body);
  return items.map((x) => apiTaskToTask(x as ApiTaskRow));
}

export type TaskCreatePayload = {
  title: string;
  description?: string | null;
  status?: TaskStatus;
  priority?: TaskPriority | null;
  due_date?: string | null;
  focus_bucket?: string;
};

export async function createTaskApi(payload: TaskCreatePayload): Promise<Task> {
  const res = await apiFetch("/tasks", { method: "POST", json: payload });
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Could not create task"), res.status, body);
  }
  if (!body || typeof body !== "object" || !("id" in body)) {
    throw new ApiError("Invalid create task response", res.status, body);
  }
  return apiTaskToTask(body as ApiTaskRow);
}

export async function patchTaskApi(taskId: string, patch: Record<string, unknown>): Promise<Task> {
  const res = await apiFetch(`/tasks/${encodeURIComponent(taskId)}`, {
    method: "PATCH",
    json: patch,
  });
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Could not update task"), res.status, body);
  }
  if (!body || typeof body !== "object" || !("id" in body)) {
    throw new ApiError("Invalid update task response", res.status, body);
  }
  return apiTaskToTask(body as ApiTaskRow);
}

export async function deleteTaskApi(taskId: string): Promise<void> {
  const res = await apiFetch(`/tasks/${encodeURIComponent(taskId)}`, { method: "DELETE" });
  if (res.status === 204) return;
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Could not delete task"), res.status, body);
  }
}
