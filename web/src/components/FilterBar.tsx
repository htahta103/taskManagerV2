import type { TaskPriority, TaskStatus } from "../types/task";

export type StatusTab = "all" | TaskStatus;
export type PriorityFilter = "all" | TaskPriority;

const statusTabs: { id: StatusTab; label: string }[] = [
  { id: "all", label: "All" },
  { id: "todo", label: "To do" },
  { id: "doing", label: "Doing" },
  { id: "done", label: "Done" },
];

const priorityOptions: { id: PriorityFilter; label: string }[] = [
  { id: "all", label: "All priorities" },
  { id: "high", label: "High" },
  { id: "medium", label: "Medium" },
  { id: "low", label: "Low" },
];

type Props = {
  status: StatusTab;
  onStatusChange: (s: StatusTab) => void;
  priority: PriorityFilter;
  onPriorityChange: (p: PriorityFilter) => void;
};

export function FilterBar({ status, onStatusChange, priority, onPriorityChange }: Props) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="flex flex-col gap-2">
        <span className="text-xs font-medium uppercase tracking-wide text-slate-500">Status</span>
        <div className="flex flex-wrap gap-1 rounded-lg bg-slate-100 p-1">
          {statusTabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => onStatusChange(tab.id)}
              className={
                status === tab.id
                  ? "rounded-md bg-white px-3 py-1.5 text-sm font-medium text-slate-900 shadow-sm"
                  : "rounded-md px-3 py-1.5 text-sm font-medium text-slate-600 hover:text-slate-900"
              }
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>
      <div className="min-w-[11rem]">
        <label className="flex flex-col gap-1">
          <span className="text-xs font-medium uppercase tracking-wide text-slate-500">Priority</span>
          <select
            value={priority}
            onChange={(e) => onPriorityChange(e.target.value as PriorityFilter)}
            className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none focus:border-sky-500 focus:ring-2 focus:ring-sky-200"
          >
            {priorityOptions.map((o) => (
              <option key={o.id} value={o.id}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
      </div>
    </div>
  );
}
