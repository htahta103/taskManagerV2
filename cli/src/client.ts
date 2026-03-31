export type TaskStatus = "todo" | "doing" | "done";
export type FocusBucket = "none" | "today" | "next" | "later";

export interface Tag {
  id: string;
  name: string;
  created_at: string;
}

export interface Task {
  id: string;
  title: string;
  description?: string | null;
  status: TaskStatus;
  priority?: "low" | "medium" | "high" | null;
  due_date?: string | null;
  focus_bucket: FocusBucket;
  project_id?: string | null;
  assignee_id?: string | null;
  tags?: Tag[];
  created_at: string;
  updated_at: string;
}

export interface TaskListPage {
  items: Task[];
  next_cursor?: string | null;
}

export interface TaskCreateBody {
  title: string;
  description?: string;
  status?: TaskStatus;
  priority?: "low" | "medium" | "high";
  due_date?: string;
  focus_bucket?: FocusBucket;
  project_id?: string;
  assignee_id?: string;
  tag_ids?: string[];
}

export interface TaskPatchBody {
  title?: string;
  description?: string | null;
  status?: TaskStatus;
  priority?: "low" | "medium" | "high" | null;
  due_date?: string | null;
  focus_bucket?: FocusBucket;
  project_id?: string | null;
  assignee_id?: string | null;
}

interface ApiErrorBody {
  error: string;
  code?: string;
}

export class ApiError extends Error {
  readonly status: number;
  readonly body?: ApiErrorBody;

  constructor(message: string, status: number, body?: ApiErrorBody) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

export class ApiClient {
  constructor(
    private readonly baseUrl: string,
    private readonly accessToken?: string,
  ) {}

  /** Join path to API root (must not use a leading `/` or resolution goes to the origin). */
  private url(path: string): string {
    const base = this.baseUrl.replace(/\/$/, "");
    const rel = path.replace(/^\//, "");
    return new URL(rel, `${base}/`).toString();
  }

  private headers(): Record<string, string> {
    const h: Record<string, string> = {
      Accept: "application/json",
      "Content-Type": "application/json",
    };
    if (this.accessToken) {
      h.Authorization = `Bearer ${this.accessToken}`;
    }
    return h;
  }

  private async parseError(res: Response): Promise<ApiError> {
    let body: ApiErrorBody | undefined;
    const text = await res.text();
    if (text) {
      try {
        body = JSON.parse(text) as ApiErrorBody;
      } catch {
        /* ignore */
      }
    }
    const msg =
      body?.error ?? `HTTP ${res.status} ${res.statusText}`.trim();
    return new ApiError(msg, res.status, body);
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    let res: Response;
    try {
      res = await fetch(this.url(path), {
        method,
        headers: this.headers(),
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (e) {
      const cause = e instanceof Error ? e.message : String(e);
      throw new ApiError(`Network error: ${cause}`, 0);
    }

    if (res.status === 204) {
      return undefined as T;
    }

    if (!res.ok) {
      throw await this.parseError(res);
    }

    const ct = res.headers.get("content-type") ?? "";
    if (!ct.includes("application/json")) {
      return undefined as T;
    }

    return (await res.json()) as T;
  }

  createTask(payload: TaskCreateBody): Promise<Task> {
    return this.request<Task>("POST", "tasks", payload);
  }

  listTasks(params: {
    project_id?: string;
    status?: TaskStatus;
    view?: "inbox" | "today" | "next" | "later";
    q?: string;
    limit?: number;
    cursor?: string;
  }): Promise<TaskListPage> {
    const q = new URLSearchParams();
    if (params.project_id) q.set("project_id", params.project_id);
    if (params.status) q.set("status", params.status);
    if (params.view) q.set("view", params.view);
    if (params.q) q.set("q", params.q);
    if (params.limit != null) q.set("limit", String(params.limit));
    if (params.cursor) q.set("cursor", params.cursor);
    const qs = q.toString();
    const path = qs ? `tasks?${qs}` : "tasks";
    return this.request<TaskListPage>("GET", path);
  }

  /** GET /search — `q` is required by the API. */
  searchTasks(params: {
    q: string;
    limit?: number;
    cursor?: string;
  }): Promise<TaskListPage> {
    const q = new URLSearchParams();
    q.set("q", params.q);
    if (params.limit != null) q.set("limit", String(params.limit));
    if (params.cursor) q.set("cursor", params.cursor);
    return this.request<TaskListPage>("GET", `search?${q.toString()}`);
  }

  async listAllTasks(
    filter: Omit<
      Parameters<ApiClient["listTasks"]>[0],
      "cursor" | "limit"
    > & { limit?: number },
  ): Promise<Task[]> {
    const limit = filter.limit ?? 100;
    const items: Task[] = [];
    let cursor: string | undefined;
    for (;;) {
      const page = await this.listTasks({
        ...filter,
        limit,
        cursor,
      });
      items.push(...page.items);
      if (!page.next_cursor) break;
      cursor = page.next_cursor;
    }
    return items;
  }

  getTask(id: string): Promise<Task> {
    return this.request<Task>("GET", `tasks/${encodeURIComponent(id)}`);
  }

  patchTask(id: string, patch: Record<string, unknown>): Promise<Task> {
    return this.request<Task>(
      "PATCH",
      `tasks/${encodeURIComponent(id)}`,
      patch,
    );
  }

  deleteTask(id: string): Promise<void> {
    return this.request<void>("DELETE", `tasks/${encodeURIComponent(id)}`);
  }
}
