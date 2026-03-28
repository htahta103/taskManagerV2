import { taskApiPathPrefix } from "./cors.js";
import type { ApiErrorBody } from "./response.js";

export const TASK_TITLE_MAX_LEN = 200;
export const TASK_DESCRIPTION_MAX_LEN = 10_000;
export const TASK_LIST_LIMIT_DEFAULT = 50;
export const TASK_LIST_LIMIT_MIN = 1;
export const TASK_LIST_LIMIT_MAX = 100;

export const TASK_STATUSES = ["todo", "doing", "done"] as const;
export type TaskStatus = (typeof TASK_STATUSES)[number];

export const TASK_PRIORITIES = ["low", "medium", "high"] as const;
export type TaskPriority = (typeof TASK_PRIORITIES)[number];

export const FOCUS_BUCKETS = ["none", "today", "next", "later"] as const;
export type FocusBucket = (typeof FOCUS_BUCKETS)[number];

export const TASK_VIEWS = ["inbox", "today", "next", "later"] as const;
export type TaskView = (typeof TASK_VIEWS)[number];

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export function isUuid(value: string): boolean {
  return UUID_RE.test(value);
}

export type TaskCreateInput = {
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

export type TaskPatchInput = {
  title?: string;
  description?: string | null;
  status?: TaskStatus;
  priority?: TaskPriority | null;
  due_date?: string | null;
  focus_bucket?: FocusBucket;
  project_id?: string | null;
  assignee_id?: string | null;
};

export type FieldIssue = { field: string; message: string };

export type ValidationFailure = {
  ok: false;
  issues: FieldIssue[];
  errorBody: ApiErrorBody;
};

export type ValidationSuccess<T> = { ok: true; value: T };

function isPlainObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

function asStringField(
  raw: unknown,
  field: string,
  maxLen: number,
  required: boolean,
): { value?: string; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined || raw === null) {
    if (required) {
      issues.push({ field, message: "required" });
    }
    return { issues };
  }
  if (typeof raw !== "string") {
    issues.push({ field, message: "must be a string" });
    return { issues };
  }
  const t = raw.trim();
  if (required && t.length === 0) {
    issues.push({ field, message: "must not be empty" });
    return { issues };
  }
  if (t.length > maxLen) {
    issues.push({ field, message: `max length is ${maxLen}` });
    return { issues };
  }
  return { value: t, issues };
}

function optionalNullableString(
  raw: unknown,
  field: string,
  maxLen: number,
): { value?: string | null; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (raw === null) {
    return { value: null, issues };
  }
  if (typeof raw !== "string") {
    issues.push({ field, message: "must be a string or null" });
    return { issues };
  }
  if (raw.length > maxLen) {
    issues.push({ field, message: `max length is ${maxLen}` });
    return { issues };
  }
  return { value: raw, issues };
}

function optionalString(raw: unknown, field: string, maxLen: number): { value?: string; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (raw === null) {
    issues.push({ field, message: "must not be null" });
    return { issues };
  }
  if (typeof raw !== "string") {
    issues.push({ field, message: "must be a string" });
    return { issues };
  }
  if (raw.length > maxLen) {
    issues.push({ field, message: `max length is ${maxLen}` });
    return { issues };
  }
  return { value: raw, issues };
}

function optionalEnum<T extends string>(
  raw: unknown,
  field: string,
  allowed: readonly T[],
): { value?: T; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (typeof raw !== "string" || !allowed.includes(raw as T)) {
    issues.push({ field, message: `must be one of: ${allowed.join(", ")}` });
    return { issues };
  }
  return { value: raw as T, issues };
}

function optionalNullableEnum<T extends string>(
  raw: unknown,
  field: string,
  allowed: readonly T[],
): { value?: T | null; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (raw === null) {
    return { value: null, issues };
  }
  if (typeof raw !== "string" || !allowed.includes(raw as T)) {
    issues.push({ field, message: `must be one of: ${allowed.join(", ")} or null` });
    return { issues };
  }
  return { value: raw as T, issues };
}

function optionalUuid(raw: unknown, field: string): { value?: string; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (typeof raw !== "string" || !isUuid(raw)) {
    issues.push({ field, message: "must be a UUID" });
    return { issues };
  }
  return { value: raw, issues };
}

function optionalNullableUuid(
  raw: unknown,
  field: string,
): { value?: string | null; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (raw === null) {
    return { value: null, issues };
  }
  if (typeof raw !== "string" || !isUuid(raw)) {
    issues.push({ field, message: "must be a UUID or null" });
    return { issues };
  }
  return { value: raw, issues };
}

function optionalIsoDate(raw: unknown, field: string): { value?: string; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (typeof raw !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(raw)) {
    issues.push({ field, message: "must be an ISO date (YYYY-MM-DD)" });
    return { issues };
  }
  return { value: raw, issues };
}

function optionalNullableIsoDate(
  raw: unknown,
  field: string,
): { value?: string | null; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (raw === null) {
    return { value: null, issues };
  }
  if (typeof raw !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(raw)) {
    issues.push({ field, message: "must be an ISO date (YYYY-MM-DD) or null" });
    return { issues };
  }
  return { value: raw, issues };
}

function optionalUuidArray(raw: unknown, field: string): { value?: string[]; issues: FieldIssue[] } {
  const issues: FieldIssue[] = [];
  if (raw === undefined) {
    return { issues };
  }
  if (!Array.isArray(raw)) {
    issues.push({ field, message: "must be an array of UUIDs" });
    return { issues };
  }
  const ids: string[] = [];
  for (let i = 0; i < raw.length; i++) {
    const el = raw[i];
    if (typeof el !== "string" || !isUuid(el)) {
      issues.push({ field: `${field}[${i}]`, message: "must be a UUID" });
    } else {
      ids.push(el);
    }
  }
  return { value: ids, issues };
}

function issuesToErrorBody(issues: FieldIssue[]): ApiErrorBody {
  return {
    error: "Validation failed",
    code: "validation_error",
    details: { fields: issues },
  };
}

export function validateTaskCreate(
  body: unknown,
): ValidationSuccess<TaskCreateInput> | ValidationFailure {
  if (!isPlainObject(body)) {
    const issues: FieldIssue[] = [{ field: "body", message: "must be a JSON object" }];
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }

  const issues: FieldIssue[] = [];
  const title = asStringField(body.title, "title", TASK_TITLE_MAX_LEN, true);
  issues.push(...title.issues);

  const description = optionalString(body.description, "description", TASK_DESCRIPTION_MAX_LEN);
  issues.push(...description.issues);

  const status = optionalEnum(body.status, "status", TASK_STATUSES);
  issues.push(...status.issues);

  const priority = optionalEnum(body.priority, "priority", TASK_PRIORITIES);
  issues.push(...priority.issues);

  const due_date = optionalIsoDate(body.due_date, "due_date");
  issues.push(...due_date.issues);

  const focus_bucket = optionalEnum(body.focus_bucket, "focus_bucket", FOCUS_BUCKETS);
  issues.push(...focus_bucket.issues);

  const project_id = optionalUuid(body.project_id, "project_id");
  issues.push(...project_id.issues);

  const assignee_id = optionalUuid(body.assignee_id, "assignee_id");
  issues.push(...assignee_id.issues);

  const tag_ids = optionalUuidArray(body.tag_ids, "tag_ids");
  issues.push(...tag_ids.issues);

  if (issues.length > 0) {
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }

  if (title.value === undefined) {
    const iss: FieldIssue[] = [{ field: "title", message: "required" }];
    return { ok: false, issues: iss, errorBody: issuesToErrorBody(iss) };
  }

  const value: TaskCreateInput = {
    title: title.value,
  };
  if (description.value !== undefined) {
    value.description = description.value;
  }
  if (status.value !== undefined) {
    value.status = status.value;
  }
  if (priority.value !== undefined) {
    value.priority = priority.value;
  }
  if (due_date.value !== undefined) {
    value.due_date = due_date.value;
  }
  if (focus_bucket.value !== undefined) {
    value.focus_bucket = focus_bucket.value;
  }
  if (project_id.value !== undefined) {
    value.project_id = project_id.value;
  }
  if (assignee_id.value !== undefined) {
    value.assignee_id = assignee_id.value;
  }
  if (tag_ids.value !== undefined) {
    value.tag_ids = tag_ids.value;
  }

  return { ok: true, value };
}

export function validateTaskPatch(
  body: unknown,
): ValidationSuccess<TaskPatchInput> | ValidationFailure {
  if (!isPlainObject(body)) {
    const issues: FieldIssue[] = [{ field: "body", message: "must be a JSON object" }];
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }

  const keys = Object.keys(body);
  if (keys.length === 0) {
    const issues: FieldIssue[] = [{ field: "body", message: "must include at least one field" }];
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }

  const issues: FieldIssue[] = [];
  const value: TaskPatchInput = {};

  if ("title" in body) {
    const title = asStringField(body.title, "title", TASK_TITLE_MAX_LEN, true);
    issues.push(...title.issues);
    if (title.value !== undefined) {
      value.title = title.value;
    }
  }

  if ("description" in body) {
    const d = optionalNullableString(body.description, "description", TASK_DESCRIPTION_MAX_LEN);
    issues.push(...d.issues);
    if (d.value !== undefined) {
      value.description = d.value;
    }
  }

  if ("status" in body) {
    const s = optionalEnum(body.status, "status", TASK_STATUSES);
    issues.push(...s.issues);
    if (s.value !== undefined) {
      value.status = s.value;
    }
  }

  if ("priority" in body) {
    const p = optionalNullableEnum(body.priority, "priority", TASK_PRIORITIES);
    issues.push(...p.issues);
    if (p.value !== undefined) {
      value.priority = p.value;
    }
  }

  if ("due_date" in body) {
    const d = optionalNullableIsoDate(body.due_date, "due_date");
    issues.push(...d.issues);
    if (d.value !== undefined) {
      value.due_date = d.value;
    }
  }

  if ("focus_bucket" in body) {
    const f = optionalEnum(body.focus_bucket, "focus_bucket", FOCUS_BUCKETS);
    issues.push(...f.issues);
    if (f.value !== undefined) {
      value.focus_bucket = f.value;
    }
  }

  if ("project_id" in body) {
    const p = optionalNullableUuid(body.project_id, "project_id");
    issues.push(...p.issues);
    if (p.value !== undefined) {
      value.project_id = p.value;
    }
  }

  if ("assignee_id" in body) {
    const a = optionalNullableUuid(body.assignee_id, "assignee_id");
    issues.push(...a.issues);
    if (a.value !== undefined) {
      value.assignee_id = a.value;
    }
  }

  if (issues.length > 0) {
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }

  return { ok: true, value };
}

export type TaskListQuery = {
  limit: number;
  cursor?: string;
  project_id?: string;
  status?: TaskStatus;
  view?: TaskView;
  q?: string;
};

export function parseTaskListQuery(
  searchParams: URLSearchParams,
): ValidationSuccess<TaskListQuery> | ValidationFailure {
  const issues: FieldIssue[] = [];

  let limit = TASK_LIST_LIMIT_DEFAULT;
  const limitRaw = searchParams.get("limit");
  if (limitRaw !== null) {
    const n = Number(limitRaw);
    if (!Number.isInteger(n) || n < TASK_LIST_LIMIT_MIN || n > TASK_LIST_LIMIT_MAX) {
      issues.push({
        field: "limit",
        message: `must be an integer between ${TASK_LIST_LIMIT_MIN} and ${TASK_LIST_LIMIT_MAX}`,
      });
    } else {
      limit = n;
    }
  }

  const cursor = searchParams.get("cursor") ?? undefined;
  if (cursor !== undefined && cursor.length === 0) {
    issues.push({ field: "cursor", message: "must not be empty when provided" });
  }

  const projectRaw = searchParams.get("project_id");
  let project_id: string | undefined;
  if (projectRaw !== null) {
    if (!isUuid(projectRaw)) {
      issues.push({ field: "project_id", message: "must be a UUID" });
    } else {
      project_id = projectRaw;
    }
  }

  const statusRaw = searchParams.get("status");
  let status: TaskStatus | undefined;
  if (statusRaw !== null) {
    if (!TASK_STATUSES.includes(statusRaw as TaskStatus)) {
      issues.push({ field: "status", message: `must be one of: ${TASK_STATUSES.join(", ")}` });
    } else {
      status = statusRaw as TaskStatus;
    }
  }

  const viewRaw = searchParams.get("view");
  let view: TaskView | undefined;
  if (viewRaw !== null) {
    if (!TASK_VIEWS.includes(viewRaw as TaskView)) {
      issues.push({ field: "view", message: `must be one of: ${TASK_VIEWS.join(", ")}` });
    } else {
      view = viewRaw as TaskView;
    }
  }

  const qRaw = searchParams.get("q");
  let q: string | undefined;
  if (qRaw !== null) {
    if (qRaw.length > 500) {
      issues.push({ field: "q", message: "max length is 500" });
    } else {
      q = qRaw;
    }
  }

  if (issues.length > 0) {
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }

  return {
    ok: true,
    value: {
      limit,
      cursor,
      project_id,
      status,
      view,
      q,
    },
  };
}

export function parseTaskIdFromPath(
  pathname: string,
): ValidationSuccess<string> | ValidationFailure {
  const prefix = `${taskApiPathPrefix()}/`;
  if (!pathname.startsWith(prefix)) {
    const issues: FieldIssue[] = [{ field: "taskId", message: "not a task path" }];
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }
  const rest = pathname.slice(prefix.length);
  const segment = rest.split("/")[0] ?? "";
  if (!segment || !isUuid(segment)) {
    const issues: FieldIssue[] = [{ field: "taskId", message: "must be a UUID" }];
    return { ok: false, issues, errorBody: issuesToErrorBody(issues) };
  }
  return { ok: true, value: segment };
}
