#!/usr/bin/env node
import {
  ApiClient,
  ApiError,
  type FocusBucket,
  type TaskStatus,
} from "./client.js";
import { formatTaskDetail, formatTaskList } from "./format.js";

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function isUuid(s: string): boolean {
  return UUID_RE.test(s);
}

function die(msg: string, code = 1): never {
  console.error(msg);
  process.exit(code);
}

function usage(): void {
  console.log(`taskmanager CLI — manage tasks via HTTP API (OpenAPI /api/v1)

Environment:
  TASKMANAGER_API_URL   API root (no trailing slash), e.g. http://localhost:8080/api/v1 or https://<ref>.supabase.co/functions/v1
  TASKMANAGER_TOKEN     Optional Bearer JWT for authenticated requests

Commands:
  task add [options] <title...>     Create a task; prints new task UUID
  task list [options]               List tasks (formatted table)
  task search [options]             List tasks (GET /search; requires -q)
  task get <task-id>                Show one task
  task edit <task-id> [options]     Update fields (PATCH)
  task done <task-id>               Mark task done
  task progress <task-id>           Set status to doing (in progress)
  task delete <task-id>             Delete a task
  task clear [--done-only] [--force]   Delete many tasks (requires --force)

List options:
  --view inbox|today|next|later
  --status todo|doing|done
  --project-id <uuid>
  --limit <n>                       Page size when fetching (default 50)
  -q, --query <text>                Search substring (title)

Search options:
  -q, --query <text>                Required search string
  --limit <n>                       Page size (default 50)

Add options:
  -d, --description <text>
  --status todo|doing|done
  --priority low|medium|high
  --due-date <YYYY-MM-DD>
  --focus-bucket none|today|next|later
  --project-id <uuid>

Edit options (at least one required):
  --title <text>
  -d, --description <text>
  --clear-description
  --status todo|doing|done
  --priority low|medium|high
  --clear-priority
  --due-date <YYYY-MM-DD>
  --clear-due-date
  --focus-bucket none|today|next|later
  --project-id <uuid>
  --clear-project
  --assignee-id <uuid>
  --clear-assignee
`);
}

type ArgMap = Record<string, string | boolean>;

function parseArgs(argv: string[]): { _: string[]; flags: ArgMap } {
  const _: string[] = [];
  const flags: ArgMap = {};
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === "--") {
      _.push(...argv.slice(i + 1));
      break;
    }
    if (a === "-d" || a === "--description") {
      const v = argv[++i];
      if (!v || v.startsWith("-")) die(`${a}: expected a value`);
      flags.description = v;
      continue;
    }
    if (a === "-q" || a === "--query") {
      const v = argv[++i];
      if (!v || v.startsWith("-")) die(`${a}: expected a value`);
      flags.query = v;
      continue;
    }
    if (a.startsWith("--")) {
      const raw = a.slice(2);
      const eq = raw.indexOf("=");
      if (eq >= 0) {
        flags[raw.slice(0, eq)] = raw.slice(eq + 1);
        continue;
      }
      const next = argv[i + 1];
      const booleanOnly = [
        "force",
        "help",
        "done-only",
        "clear-description",
        "clear-priority",
        "clear-due-date",
        "clear-project",
        "clear-assignee",
      ].includes(raw);
      if (!booleanOnly && next && !next.startsWith("-")) {
        flags[raw] = next;
        i++;
      } else {
        flags[raw] = true;
      }
      continue;
    }
    if (a.startsWith("-")) {
      die(`Unknown flag: ${a}`);
    }
    _.push(a);
  }
  return { _, flags };
}

function requireUuid(label: string, id: string): void {
  if (!isUuid(id)) {
    die(`Invalid ${label}: not a UUID — ${JSON.stringify(id)}`);
  }
}

function clientFromEnv(): ApiClient {
  const base =
    process.env.TASKMANAGER_API_URL ??
    process.env.TM_API_URL ??
    "http://localhost:8080/api/v1";
  const token =
    process.env.TASKMANAGER_TOKEN ?? process.env.TM_TOKEN ?? undefined;
  return new ApiClient(base, token);
}

async function cmdAdd(c: ApiClient, args: string[], flags: ArgMap): Promise<void> {
  const title = args.join(" ").trim();
  if (!title) die("task add: missing title");
  const body: Parameters<ApiClient["createTask"]>[0] = { title };
  const d = flags.description;
  if (typeof d === "string") body.description = d;
  const st = flags.status;
  if (typeof st === "string") {
    if (!["todo", "doing", "done"].includes(st))
      die(`task add: invalid --status ${st}`);
    body.status = st as TaskStatus;
  }
  const pr = flags.priority;
  if (typeof pr === "string") {
    if (!["low", "medium", "high"].includes(pr))
      die(`task add: invalid --priority ${pr}`);
    body.priority = pr as "low" | "medium" | "high";
  }
  const pid = flags["project-id"];
  if (typeof pid === "string") {
    requireUuid("project id", pid);
    body.project_id = pid;
  }
  const due = flags["due-date"];
  if (typeof due === "string") body.due_date = due;
  const fb = flags["focus-bucket"];
  if (typeof fb === "string") {
    if (!["none", "today", "next", "later"].includes(fb))
      die(`task add: invalid --focus-bucket ${JSON.stringify(fb)}`);
    body.focus_bucket = fb as FocusBucket;
  }
  try {
    const task = await c.createTask(body);
    console.log(task.id);
  } catch (e) {
    if (e instanceof ApiError) die(`task add: ${e.message}`, e.status || 1);
    throw e;
  }
}

async function cmdSearch(c: ApiClient, flags: ArgMap): Promise<void> {
  const raw = flags.q ?? flags.query;
  const q = typeof raw === "string" ? raw.trim() : "";
  if (!q) die("task search: missing -q/--query");
  const lim = flags.limit;
  const params: Parameters<ApiClient["searchTasks"]>[0] = { q };
  if (typeof lim === "string") {
    const n = Number(lim);
    if (!Number.isFinite(n) || n < 1 || n > 100)
      die("task search: --limit must be 1–100");
    params.limit = n;
  }
  try {
    const page = await c.searchTasks(params);
    console.log(formatTaskList(page.items));
    if (page.next_cursor) {
      console.error("(more results available; pagination not shown — narrow query)");
    }
  } catch (e) {
    if (e instanceof ApiError) die(`task search: ${e.message}`, e.status || 1);
    throw e;
  }
}

async function cmdList(c: ApiClient, flags: ArgMap): Promise<void> {
  const q: Parameters<ApiClient["listTasks"]>[0] = {};
  const v = flags.view;
  if (typeof v === "string") {
    if (!["inbox", "today", "next", "later"].includes(v))
      die(`task list: invalid --view ${v}`);
    q.view = v as "inbox" | "today" | "next" | "later";
  }
  const st = flags.status;
  if (typeof st === "string") {
    if (!["todo", "doing", "done"].includes(st))
      die(`task list: invalid --status ${st}`);
    q.status = st as TaskStatus;
  }
  const pid = flags["project-id"];
  if (typeof pid === "string") {
    requireUuid("project id", pid);
    q.project_id = pid;
  }
  const lim = flags.limit;
  if (typeof lim === "string") {
    const n = Number(lim);
    if (!Number.isFinite(n) || n < 1 || n > 100)
      die("task list: --limit must be 1–100");
    q.limit = n;
  }
  const search = flags.query;
  if (typeof search === "string") q.q = search;
  try {
    const page = await c.listTasks(q);
    console.log(formatTaskList(page.items));
    if (page.next_cursor) {
      console.error("(more results available; pagination not shown — narrow filters)");
    }
  } catch (e) {
    if (e instanceof ApiError) die(`task list: ${e.message}`, e.status || 1);
    throw e;
  }
}

async function cmdEdit(c: ApiClient, id: string, flags: ArgMap): Promise<void> {
  requireUuid("task id", id);
  const patch: Record<string, unknown> = {};
  const title = flags.title;
  if (typeof title === "string") {
    if (!title.trim()) die("task edit: --title must be non-empty");
    patch.title = title;
  }
  const desc = flags.description;
  const clearDesc = Boolean(flags["clear-description"]);
  if (clearDesc && typeof desc === "string")
    die("task edit: use either --description or --clear-description, not both");
  if (typeof desc === "string") patch.description = desc;
  if (clearDesc) patch.description = null;

  const st = flags.status;
  if (typeof st === "string") {
    if (!["todo", "doing", "done"].includes(st))
      die(`task edit: invalid --status ${JSON.stringify(st)}`);
    patch.status = st;
  }

  const pr = flags.priority;
  const clearPr = Boolean(flags["clear-priority"]);
  if (clearPr && typeof pr === "string")
    die("task edit: use either --priority or --clear-priority, not both");
  if (typeof pr === "string") {
    if (!["low", "medium", "high"].includes(pr))
      die(`task edit: invalid --priority ${JSON.stringify(pr)}`);
    patch.priority = pr;
  }
  if (clearPr) patch.priority = null;

  const due = flags["due-date"];
  const clearDue = Boolean(flags["clear-due-date"]);
  if (clearDue && typeof due === "string")
    die("task edit: use either --due-date or --clear-due-date, not both");
  if (typeof due === "string") patch.due_date = due;
  if (clearDue) patch.due_date = null;

  const foc = flags["focus-bucket"];
  if (typeof foc === "string") {
    if (!["none", "today", "next", "later"].includes(foc))
      die(`task edit: invalid --focus-bucket ${JSON.stringify(foc)}`);
    patch.focus_bucket = foc;
  }

  const pid = flags["project-id"];
  const clearProj = Boolean(flags["clear-project"]);
  if (clearProj && typeof pid === "string")
    die("task edit: use either --project-id or --clear-project, not both");
  if (typeof pid === "string") {
    requireUuid("project id", pid);
    patch.project_id = pid;
  }
  if (clearProj) patch.project_id = null;

  const aid = flags["assignee-id"];
  const clearAid = Boolean(flags["clear-assignee"]);
  if (clearAid && typeof aid === "string")
    die("task edit: use either --assignee-id or --clear-assignee, not both");
  if (typeof aid === "string") {
    requireUuid("assignee id", aid);
    patch.assignee_id = aid;
  }
  if (clearAid) patch.assignee_id = null;

  if (Object.keys(patch).length === 0) {
    die("task edit: pass at least one field to change (see --help)");
  }
  try {
    const t = await c.patchTask(id, patch);
    console.log(formatTaskDetail(t));
  } catch (e) {
    if (e instanceof ApiError) {
      if (e.status === 404) die(`task edit: task not found (${id})`, 1);
      die(`task edit: ${e.message}`, e.status || 1);
    }
    throw e;
  }
}

async function cmdGet(c: ApiClient, id: string): Promise<void> {
  requireUuid("task id", id);
  try {
    const t = await c.getTask(id);
    console.log(formatTaskDetail(t));
  } catch (e) {
    if (e instanceof ApiError) {
      if (e.status === 404) die(`task get: task not found (${id})`, 1);
      die(`task get: ${e.message}`, e.status || 1);
    }
    throw e;
  }
}

async function cmdDone(c: ApiClient, id: string): Promise<void> {
  requireUuid("task id", id);
  try {
    const t = await c.patchTask(id, { status: "done" });
    console.log(formatTaskDetail(t));
  } catch (e) {
    if (e instanceof ApiError) {
      if (e.status === 404) die(`task done: task not found (${id})`, 1);
      die(`task done: ${e.message}`, e.status || 1);
    }
    throw e;
  }
}

async function cmdProgress(c: ApiClient, id: string): Promise<void> {
  requireUuid("task id", id);
  try {
    const t = await c.patchTask(id, { status: "doing" });
    console.log(formatTaskDetail(t));
  } catch (e) {
    if (e instanceof ApiError) {
      if (e.status === 404) die(`task progress: task not found (${id})`, 1);
      die(`task progress: ${e.message}`, e.status || 1);
    }
    throw e;
  }
}

async function cmdDelete(c: ApiClient, id: string): Promise<void> {
  requireUuid("task id", id);
  try {
    await c.deleteTask(id);
  } catch (e) {
    if (e instanceof ApiError) {
      if (e.status === 404) die(`task delete: task not found (${id})`, 1);
      die(`task delete: ${e.message}`, e.status || 1);
    }
    throw e;
  }
}

async function cmdClear(
  c: ApiClient,
  flags: ArgMap,
): Promise<void> {
  if (!flags.force) {
    die("task clear: refusing without --force (destructive)");
  }
  const doneOnly = Boolean(flags["done-only"]);
  try {
    const tasks = await c.listAllTasks(
      doneOnly ? { status: "done" } : {},
    );
    let n = 0;
    for (const t of tasks) {
      await c.deleteTask(t.id);
      n++;
    }
    console.error(`Deleted ${n} task(s).`);
  } catch (e) {
    if (e instanceof ApiError) die(`task clear: ${e.message}`, e.status || 1);
    throw e;
  }
}

async function main(): Promise<void> {
  const argv = process.argv.slice(2);
  if (argv.length === 0 || argv[0] === "help" || argv.includes("--help")) {
    usage();
    process.exit(0);
  }
  const cmd = argv[0];
  const rest = argv.slice(1);
  const { _, flags } = parseArgs(rest);

  if (flags.help) {
    usage();
    process.exit(0);
  }

  const c = clientFromEnv();

  switch (cmd) {
    case "add":
      await cmdAdd(c, _, flags);
      break;
    case "list":
      await cmdList(c, flags);
      break;
    case "search":
      await cmdSearch(c, flags);
      break;
    case "get":
      if (_.length < 1) die("task get: missing <task-id>");
      await cmdGet(c, _[0]);
      break;
    case "edit":
      if (_.length < 1) die("task edit: missing <task-id>");
      await cmdEdit(c, _[0], flags);
      break;
    case "done":
      if (_.length < 1) die("task done: missing <task-id>");
      await cmdDone(c, _[0]);
      break;
    case "progress":
      if (_.length < 1) die("task progress: missing <task-id>");
      await cmdProgress(c, _[0]);
      break;
    case "delete":
      if (_.length < 1) die("task delete: missing <task-id>");
      await cmdDelete(c, _[0]);
      break;
    case "clear":
      await cmdClear(c, flags);
      break;
    default:
      die(`Unknown command: ${cmd}\nRun: task help`);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
