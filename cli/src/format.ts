import type { Task } from "./client.js";

function pad(s: string, w: number): string {
  const t = s.length > w ? `${s.slice(0, w - 1)}…` : s;
  return t.padEnd(w);
}

function formatTagNames(t: Task): string {
  const names = t.tags?.map((x) => x.name) ?? [];
  return names.join(", ");
}

/** Fixed-width table for terminal readability. */
export function formatTaskList(tasks: Task[]): string {
  if (tasks.length === 0) {
    return "(no tasks)";
  }
  const idW = 36;
  const titleW = 32;
  const statusW = 8;
  const bucketW = 8;
  const tagsW = 18;
  const header = `${pad("ID", idW)} ${pad("TITLE", titleW)} ${pad("STATUS", statusW)} ${pad("BUCKET", bucketW)} ${pad("TAGS", tagsW)}`;
  const rule = "-".repeat(header.length);
  const lines = tasks.map(
    (t) =>
      `${pad(t.id, idW)} ${pad(t.title, titleW)} ${pad(t.status, statusW)} ${pad(t.focus_bucket, bucketW)} ${pad(formatTagNames(t), tagsW)}`,
  );
  return [header, rule, ...lines].join("\n");
}

export function formatTaskDetail(t: Task): string {
  const lines = [
    `id:           ${t.id}`,
    `title:        ${t.title}`,
    `status:       ${t.status}`,
    `focus_bucket: ${t.focus_bucket}`,
    `created_at:   ${t.created_at}`,
    `updated_at:   ${t.updated_at}`,
  ];
  if (t.description) lines.splice(3, 0, `description:  ${t.description}`);
  if (t.priority) lines.splice(3, 0, `priority:     ${t.priority}`);
  if (t.due_date) lines.splice(3, 0, `due_date:     ${t.due_date}`);
  if (t.project_id) lines.push(`project_id:   ${t.project_id}`);
  if (t.assignee_id) lines.push(`assignee_id:  ${t.assignee_id}`);
  if (t.tags?.length) {
    lines.push(`tags:         ${t.tags.map((x) => x.name).join(", ")}`);
  }
  return lines.join("\n");
}
