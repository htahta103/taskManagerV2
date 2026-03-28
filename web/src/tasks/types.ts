export type TaskStatus = "todo" | "doing" | "done";

export type TaskPriority = "low" | "medium" | "high";

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
  tags: string[];
  createdAt: string;
  updatedAt: string;
};

export type FocusBucket = "today" | "next" | "later";
