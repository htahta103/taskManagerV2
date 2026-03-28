export type TaskStatus = "todo" | "doing" | "done";

export type TaskPriority = "low" | "medium" | "high";

export interface Task {
  id: string;
  title: string;
  description?: string | null;
  status: TaskStatus;
  priority?: TaskPriority | null;
  due_date?: string | null;
  focus_bucket: string;
  project_id?: string | null;
  assignee_id?: string | null;
  created_at: string;
  updated_at: string;
}
