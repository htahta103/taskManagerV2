export type TaskStatus = "todo" | "doing" | "done";
export type TaskPriority = "low" | "medium" | "high";
export type FocusBucket = "none" | "today" | "next" | "later";

export type Tag = {
  id: string;
  name: string;
  created_at: string;
};

export type Task = {
  id: string;
  title: string;
  description: string | null;
  status: TaskStatus;
  priority: TaskPriority | null;
  due_date: string | null;
  focus_bucket: FocusBucket;
  project_id: string | null;
  assignee_id: string | null;
  tags?: Tag[];
  created_at: string;
  updated_at: string;
};

export type TaskCreateBody = {
  title: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  due_date?: string;
  focus_bucket?: FocusBucket;
  project_id?: string;
  assignee_id?: string;
  tag_ids?: string[];
};

export type TaskPatchBody = {
  title?: string;
  description?: string | null;
  status?: TaskStatus;
  priority?: TaskPriority | null;
  due_date?: string | null;
  focus_bucket?: FocusBucket;
  project_id?: string | null;
  assignee_id?: string | null;
};

export type TaskListPage = {
  items: Task[];
  next_cursor?: string | null;
};
