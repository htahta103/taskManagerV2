import type { Task, TaskStatus } from "../types/task";

const listPath = "/api/v1/tasks";

export async function fetchTasks(
  apiBase: string,
  opts: { status?: TaskStatus; q?: string },
): Promise<Task[]> {
  const root = apiBase.replace(/\/$/, "");
  const url = new URL(listPath, root.endsWith("/") ? root : root + "/");
  if (opts.status) url.searchParams.set("status", opts.status);
  if (opts.q) url.searchParams.set("q", opts.q);
  const res = await fetch(url.toString());
  if (!res.ok) {
    throw new Error(`Failed to load tasks (${res.status})`);
  }
  const data = (await res.json()) as { items: Task[] };
  return data.items ?? [];
}
