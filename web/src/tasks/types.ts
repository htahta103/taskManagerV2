export type TaskStatus = "todo" | "doing" | "done";

export type TaskPriority = "low" | "medium" | "high";

/** Stored on tasks loaded from `/api/v1/tasks` (`focus_bucket` column). */
export type TaskFocusBucket = "none" | "today" | "next" | "later";

export type Task = {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority?: TaskPriority;
  /** Local calendar date YYYY-MM-DD */
  dueDate?: string;
  /** Manual “do today” flag (PRD: flagged for today) */
  pinnedToday: boolean;
  /** Explicit bucket from API when present (local-only tasks may omit). */
  focusBucket?: TaskFocusBucket;
  tags: string[];
  createdAt: string;
  updatedAt: string;
};

export type FocusBucket = "today" | "next" | "later";
