import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useAuth } from "../auth/AuthContext";
import { loadTasks, saveTasks } from "./storage";
import type { Task, TaskPriority, TaskStatus } from "./types";

type TaskState = {
  tasks: Task[];
  loading: boolean;
  addTask: (title: string) => Task | null;
  updateTask: (id: string, patch: Partial<Omit<Task, "id" | "createdAt">>) => void;
  removeTask: (id: string) => void;
  toggleDone: (id: string) => void;
  editing: Task | null;
  openEditor: (task: Task) => void;
  closeEditor: () => void;
};

const TaskContext = createContext<TaskState | null>(null);

function nowIso(): string {
  return new Date().toISOString();
}

function newTask(title: string): Task {
  const t = nowIso();
  return {
    id: crypto.randomUUID(),
    title,
    description: "",
    status: "todo",
    pinnedToday: false,
    tags: [],
    createdAt: t,
    updatedAt: t,
  };
}

export function TaskProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [editId, setEditId] = useState<string | null>(null);

  useEffect(() => {
    if (!user) {
      setTasks([]);
      setEditId(null);
      setLoading(false);
      return;
    }
    setLoading(true);
    setTasks(loadTasks(user.id));
    setLoading(false);
  }, [user?.id]);

  useEffect(() => {
    if (!user) return;
    saveTasks(user.id, tasks);
  }, [user?.id, tasks]);

  const addTask = useCallback(
    (title: string) => {
      const trimmed = title.trim();
      if (!trimmed) return null;
      const task = newTask(trimmed);
      setTasks((prev) => [task, ...prev]);
      return task;
    },
    [setTasks],
  );

  const updateTask = useCallback((id: string, patch: Partial<Omit<Task, "id" | "createdAt">>) => {
    setTasks((prev) =>
      prev.map((t) => {
        if (t.id !== id) return t;
        const next = { ...t, ...patch, updatedAt: nowIso() };
        if (typeof next.title === "string") {
          const tr = next.title.trim();
          if (!tr) return t;
          next.title = tr;
        }
        return next;
      }),
    );
  }, []);

  const removeTask = useCallback((id: string) => {
    setTasks((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toggleDone = useCallback((id: string) => {
    setTasks((prev) =>
      prev.map((t) => {
        if (t.id !== id) return t;
        const nextStatus: TaskStatus = t.status === "done" ? "todo" : "done";
        return { ...t, status: nextStatus, updatedAt: nowIso() };
      }),
    );
  }, []);

  const openEditor = useCallback((task: Task) => {
    setEditId(task.id);
  }, []);

  const closeEditor = useCallback(() => {
    setEditId(null);
  }, []);

  const editing = useMemo(() => {
    if (!editId) return null;
    return tasks.find((t) => t.id === editId) ?? null;
  }, [editId, tasks]);

  const value = useMemo(
    () => ({
      tasks,
      loading,
      addTask,
      updateTask,
      removeTask,
      toggleDone,
      editing,
      openEditor,
      closeEditor,
    }),
    [tasks, loading, addTask, updateTask, removeTask, toggleDone, editing, openEditor, closeEditor],
  );

  return <TaskContext.Provider value={value}>{children}</TaskContext.Provider>;
}

export function useTasks(): TaskState {
  const ctx = useContext(TaskContext);
  if (!ctx) throw new Error("useTasks must be used within TaskProvider");
  return ctx;
}

export type TaskDraft = {
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority | "";
  dueDate: string;
  pinnedToday: boolean;
};

export function taskToDraft(task: Task): TaskDraft {
  return {
    title: task.title,
    description: task.description,
    status: task.status,
    priority: task.priority ?? "",
    dueDate: task.dueDate ?? "",
    pinnedToday: task.pinnedToday,
  };
}

export function draftToPatch(d: TaskDraft): Partial<Omit<Task, "id" | "createdAt">> {
  return {
    title: d.title.trim(),
    description: d.description,
    status: d.status,
    priority: d.priority === "" ? undefined : d.priority,
    dueDate: d.dueDate === "" ? undefined : d.dueDate,
    pinnedToday: d.pinnedToday,
  };
}
