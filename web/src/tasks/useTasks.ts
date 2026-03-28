import { useCallback, useEffect, useState } from "react";
import { createTask as createTaskApi, fetchTasks, updateTask as updateTaskApi } from "./api";
import type { Task, TaskCreateBody, TaskPatchBody } from "./types";

export function useTasks(options: { view?: "inbox" | "today" | "next" | "later" } = {}) {
  const view = options.view ?? "inbox";
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setError(null);
    setLoading(true);
    try {
      const page = await fetchTasks({ view });
      setTasks(page.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load tasks");
      setTasks([]);
    } finally {
      setLoading(false);
    }
  }, [view]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const createTask = useCallback(async (body: TaskCreateBody) => {
    const task = await createTaskApi(body);
    setTasks((prev) => [task, ...prev]);
    return task;
  }, []);

  const updateTask = useCallback(async (taskId: string, patch: TaskPatchBody) => {
    const task = await updateTaskApi(taskId, patch);
    setTasks((prev) => prev.map((t) => (t.id === taskId ? task : t)));
    return task;
  }, []);

  return { tasks, loading, error, refresh, createTask, updateTask };
}
