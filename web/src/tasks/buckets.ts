import type { FocusBucket, Task } from "./types";

function pad2(n: number): string {
  return String(n).padStart(2, "0");
}

/** Local calendar YYYY-MM-DD for `d`. */
export function formatLocalYMD(d: Date): string {
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;
}

/** Calendar-day offset from local “today” to `dueDate` (YYYY-MM-DD). */
export function calendarDaysFromToday(dueDate: string, now: Date): number {
  const [y, m, d] = dueDate.split("-").map(Number);
  if (!y || !m || !d) return NaN;
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const due = new Date(y, m - 1, d);
  return Math.round((due.getTime() - start.getTime()) / 86_400_000);
}

/**
 * Derives Today / Next / Later (PRD §6) for open tasks.
 * Precedence: pinned today → due / overdue → priority heuristics.
 */
export function focusBucketForTask(task: Task, now: Date = new Date()): FocusBucket {
  if (task.status === "done") return "later";

  if (task.pinnedToday) return "today";

  const fromApi = task.focusBucket;
  if (fromApi && fromApi !== "none") {
    return fromApi;
  }

  if (task.dueDate) {
    const diff = calendarDaysFromToday(task.dueDate, now);
    if (!Number.isFinite(diff)) {
      // fall through to priority
    } else {
      if (diff < 0 || diff === 0) return "today";
      if (diff > 0 && diff <= 7) return "next";
    }
  }

  if (task.priority === "high") return "today";
  if (task.priority === "medium") return "next";
  return "later";
}

export function tasksForBucket(tasks: Task[], bucket: FocusBucket, now?: Date): Task[] {
  const t = now ?? new Date();
  return tasks.filter((x) => x.status !== "done" && focusBucketForTask(x, t) === bucket);
}
