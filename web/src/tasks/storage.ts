import type { Task } from "./types";

const STORAGE_PREFIX = "taskmanagerv2.tasks.v1:";

export function storageKeyForUser(userId: string): string {
  return `${STORAGE_PREFIX}${userId}`;
}

export function loadTasks(userId: string): Task[] {
  try {
    const raw = localStorage.getItem(storageKeyForUser(userId));
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(isTask);
  } catch {
    return [];
  }
}

export function saveTasks(userId: string, tasks: Task[]): void {
  localStorage.setItem(storageKeyForUser(userId), JSON.stringify(tasks));
}

function isTask(x: unknown): x is Task {
  if (!x || typeof x !== "object") return false;
  const o = x as Record<string, unknown>;
  return (
    typeof o.id === "string" &&
    typeof o.title === "string" &&
    typeof o.description === "string" &&
    (o.status === "todo" || o.status === "doing" || o.status === "done") &&
    typeof o.pinnedToday === "boolean" &&
    Array.isArray(o.tags) &&
    typeof o.createdAt === "string" &&
    typeof o.updatedAt === "string"
  );
}
