import type { Task } from "../types/task";

const statusStyle: Record<string, string> = {
  todo: "bg-amber-50 text-amber-800 ring-amber-200",
  doing: "bg-sky-50 text-sky-800 ring-sky-200",
  done: "bg-emerald-50 text-emerald-800 ring-emerald-200",
};

const priorityStyle: Record<string, string> = {
  high: "text-rose-700",
  medium: "text-amber-700",
  low: "text-slate-500",
};

type Props = { task: Task };

export function TaskCard({ task }: Props) {
  const sClass = statusStyle[task.status] ?? "bg-slate-50 text-slate-700 ring-slate-200";
  const pClass = task.priority ? priorityStyle[task.priority] ?? "text-slate-500" : "text-slate-400";

  return (
    <article className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-slate-300 hover:shadow">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <h2 className="text-base font-semibold text-slate-900">{task.title}</h2>
        <span className={`shrink-0 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${sClass}`}>
          {task.status}
        </span>
      </div>
      {task.description ? (
        <p className="mt-2 line-clamp-2 text-sm text-slate-600">{task.description}</p>
      ) : null}
      <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-slate-500">
        <span className={pClass}>Priority: {task.priority ?? "—"}</span>
        <span>Focus: {task.focus_bucket}</span>
      </div>
    </article>
  );
}
