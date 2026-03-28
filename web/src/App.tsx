import { useEffect, useMemo, useState } from "react";
import { fetchTasks } from "./api/tasks";
import { FilterBar, type PriorityFilter, type StatusTab } from "./components/FilterBar";
import { SearchInput } from "./components/SearchInput";
import { TaskCard } from "./components/TaskCard";
import { useDebouncedValue } from "./hooks/useDebouncedValue";
import type { Task, TaskStatus } from "./types/task";

const apiBase =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") || "http://localhost:8080";

function matchesPriority(task: Task, filter: PriorityFilter): boolean {
  if (filter === "all") return true;
  return task.priority === filter;
}

export default function App() {
  const [statusTab, setStatusTab] = useState<StatusTab>("all");
  const [priorityFilter, setPriorityFilter] = useState<PriorityFilter>("all");
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search, 300);

  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    const statusArg: TaskStatus | undefined =
      statusTab === "all" ? undefined : (statusTab as TaskStatus);
    const q = debouncedSearch.trim();
    fetchTasks(apiBase, {
      status: statusArg,
      q: q || undefined,
    })
      .then((items) => {
        if (!cancelled) setTasks(items);
      })
      .catch((e: unknown) => {
        if (!cancelled) setError(e instanceof Error ? e.message : "Failed to load tasks");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [apiBase, statusTab, debouncedSearch]);

  const visible = useMemo(
    () => tasks.filter((t) => matchesPriority(t, priorityFilter)),
    [tasks, priorityFilter],
  );

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-3xl px-4 py-6">
          <h1 className="text-2xl font-bold tracking-tight">Tasks</h1>
          <p className="mt-1 text-sm text-slate-500">
            API{" "}
            <code className="rounded bg-slate-100 px-1 py-0.5 text-xs text-slate-700">{apiBase}</code>
          </p>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-8">
        <div className="flex flex-col gap-6 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <SearchInput value={search} onChange={setSearch} />
          <FilterBar
            status={statusTab}
            onStatusChange={setStatusTab}
            priority={priorityFilter}
            onPriorityChange={setPriorityFilter}
          />
        </div>

        <section className="mt-8" aria-live="polite">
          {loading ? (
            <p className="text-center text-sm text-slate-500">Loading tasks…</p>
          ) : error ? (
            <p className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
              {error}
            </p>
          ) : visible.length === 0 ? (
            <p className="text-center text-sm text-slate-500">No tasks match your filters.</p>
          ) : (
            <ul className="flex flex-col gap-3">
              {visible.map((task) => (
                <li key={task.id}>
                  <TaskCard task={task} />
                </li>
              ))}
            </ul>
          )}
        </section>
      </main>
    </div>
  );
}
