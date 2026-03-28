/**
 * Route suffix after the Edge Function name "tasks" (Supabase path: /functions/v1/tasks/...).
 */

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export type ParsedTasksRoute =
  | { kind: "collection" }
  | { kind: "by_id"; id: string }
  | { kind: "clear_done" }
  | { kind: "not_found" };

/** Extract "/..." suffix relative to the `tasks` function segment. */
export function extractTasksSuffix(pathname: string): string {
  const marker = "/tasks";
  const idx = pathname.lastIndexOf(marker);
  if (idx === -1) return "";
  const rest = pathname.slice(idx + marker.length);
  if (rest === "" || rest === "/") return "/";
  return rest.startsWith("/") ? rest : `/${rest}`;
}

export function parseTasksPathSuffix(suffix: string): ParsedTasksRoute {
  const trimmed = suffix.replace(/\/+$/, "") || "/";
  if (trimmed === "/") return { kind: "collection" };
  const parts = trimmed.split("/").filter(Boolean);
  if (parts.length === 2 && parts[0] === "clear" && parts[1] === "done") {
    return { kind: "clear_done" };
  }
  if (parts.length === 1 && UUID_RE.test(parts[0])) {
    return { kind: "by_id", id: parts[0] };
  }
  return { kind: "not_found" };
}
