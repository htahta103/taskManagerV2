/**
 * Parse and validate GET /tasks list query parameters (status, priority, search).
 */

const VALID_STATUS = new Set(["todo", "doing", "done"]);
const VALID_PRIORITY = new Set(["low", "medium", "high"]);

export type TaskListQuery = {
  status?: string;
  priority?: string;
  search?: string;
};

export type ParseResult =
  | { ok: true; value: TaskListQuery }
  | { ok: false; error: string };

export function parseTaskListQuery(url: URL): ParseResult {
  const status = url.searchParams.get("status");
  const priority = url.searchParams.get("priority");
  const searchRaw = url.searchParams.get("search");

  if (status !== null && status !== "" && !VALID_STATUS.has(status)) {
    return { ok: false, error: `invalid status: ${status}` };
  }
  if (priority !== null && priority !== "" && !VALID_PRIORITY.has(priority)) {
    return { ok: false, error: `invalid priority: ${priority}` };
  }

  const value: TaskListQuery = {};
  if (status) value.status = status;
  if (priority) value.priority = priority;
  const trimmed = searchRaw?.trim();
  if (trimmed) {
    if (trimmed.includes(",")) {
      return {
        ok: false,
        error: "search must not contain commas (reserved for filter syntax)",
      };
    }
    value.search = trimmed;
  }

  return { ok: true, value };
}

/** Escape % and _ so user input cannot broaden an ILIKE pattern. */
export function escapeILikeLiteral(s: string): string {
  return s.replace(/\\/g, "\\\\").replace(/%/g, "\\%").replace(/_/g, "\\_");
}
