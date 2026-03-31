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
import {
  createTaskApi,
  deleteTaskApi,
  fetchTasksFromApi,
  patchTaskApi,
  patchToApi,
  taskToApiUpdateBody,
} from "./api";
import { loadTasks, saveTasks } from "./storage";
import type { Task, TaskPriority, TaskStatus } from "./types";

export type TaskModalState = null | { mode: "create" } | { mode: "edit"; task: Task };

type TaskState = {
  tasks: Task[];
  loading: boolean;
  taskModal: TaskModalState;
  openCreateModal: () => void;
  openTaskEditor: (task: Task) => void;
  closeTaskModal: () => void;
  submitNewTask: (draft: TaskDraft) => Promise<void>;
  submitTaskEdit: (task: Task, draft: TaskDraft) => Promise<void>;
  addTask: (title: string) => Task | null;
  updateTask: (id: string, patch: Partial<Omit<Task, "id" | "createdAt">>) => void;
  removeTask: (id: string) => void;
  toggleDone: (id: string) => void;
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
  const [useLocalTasks, setUseLocalTasks] = useState(false);
  const [taskModal, setTaskModal] = useState<TaskModalState>(null);

  useEffect(() => {
    if (!user) {
      setTasks([]);
      setTaskModal(null);
      setLoading(false);
      setUseLocalTasks(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    fetchTasksFromApi()
      .then((items) => {
        if (cancelled) return;
        setTasks(items);
        setUseLocalTasks(false);
      })
      .catch(() => {
        if (cancelled) return;
        setTasks(loadTasks(user.id));
        setUseLocalTasks(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [user?.id]);

  useEffect(() => {
    if (!user || !useLocalTasks) return;
    saveTasks(user.id, tasks);
  }, [user?.id, tasks, useLocalTasks, user]);

  const openCreateModal = useCallback(() => {
    setTaskModal({ mode: "create" });
  }, []);

  const openTaskEditor = useCallback((task: Task) => {
    setTaskModal({ mode: "edit", task });
  }, []);

  const closeTaskModal = useCallback(() => {
    setTaskModal(null);
  }, []);

  const updateTaskLocal = useCallback((id: string, patch: Partial<Omit<Task, "id" | "createdAt">>) => {
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

  const submitNewTaskCb = useCallback(
    async (draft: TaskDraft) => {
      const title = draft.title.trim();
      if (!title) throw new Error("title required");
      if (!user) return;
      if (useLocalTasks) {
        const task = newTask(title);
        Object.assign(task, draftToPatch({ ...draft, title }));
        task.updatedAt = task.createdAt;
        setTasks((prev) => [task, ...prev]);
        closeTaskModal();
        return;
      }
      const row = await createTaskApi({
        title,
        description: draft.description === "" ? null : draft.description,
        status: draft.status,
        priority: draft.priority === "" ? null : draft.priority,
        due_date: draft.dueDate === "" ? null : draft.dueDate,
        focus_bucket: draft.pinnedToday ? "today" : "none",
      });
      setTasks((prev) => [row, ...prev]);
      closeTaskModal();
    },
    [user, useLocalTasks, closeTaskModal],
  );

  const submitTaskEditCb = useCallback(
    async (task: Task, draft: TaskDraft) => {
      const title = draft.title.trim();
      if (!title) throw new Error("title required");
      if (useLocalTasks) {
        updateTaskLocal(task.id, draftToPatch({ ...draft, title }));
        closeTaskModal();
        return;
      }
      const patch = draftToPatch({ ...draft, title });
      const merged: Task = { ...task, ...patch };
      const row = await patchTaskApi(task.id, taskToApiUpdateBody(merged));
      setTasks((prev) => prev.map((t) => (t.id === task.id ? row : t)));
      closeTaskModal();
    },
    [useLocalTasks, closeTaskModal, updateTaskLocal],
  );

  useEffect(() => {
    if (!taskModal || taskModal.mode !== "edit") return;
    if (!tasks.some((t) => t.id === taskModal.task.id)) {
      setTaskModal(null);
    }
  }, [tasks, taskModal]);

  const taskModalForUi = useMemo((): TaskModalState => {
    if (!taskModal || taskModal.mode !== "edit") return taskModal;
    const live = tasks.find((t) => t.id === taskModal.task.id);
    if (!live) return taskModal;
    return { mode: "edit", task: live };
  }, [taskModal, tasks]);

  const addTask = useCallback(
    (title: string) => {
      const trimmed = title.trim();
      if (!trimmed) return null;
      const task = newTask(trimmed);
      setTasks((prev) => [task, ...prev]);
      return task;
    },
    [],
  );

  const updateTask = useCallback(
    (id: string, patch: Partial<Omit<Task, "id" | "createdAt">>) => {
      if (useLocalTasks) {
        updateTaskLocal(id, patch);
        return;
      }
      setTasks((prev) => {
        const task = prev.find((t) => t.id === id);
        if (!task) return prev;
        const body = patchToApi(task, patch);
        if (Object.keys(body).length === 0) return prev;
        void patchTaskApi(id, body)
          .then((row) => {
            setTasks((p) => p.map((t) => (t.id === id ? row : t)));
          })
          .catch(() => {
            /* keep UI; Witness / user can refresh */
          });
        const merged = { ...task, ...patch, updatedAt: nowIso() };
        if (typeof merged.title === "string") {
          const tr = merged.title.trim();
          if (!tr) return prev;
          merged.title = tr;
        }
        return prev.map((t) => (t.id === id ? (merged as Task) : t));
      });
    },
    [useLocalTasks],
  );

  const removeTask = useCallback(
    (id: string) => {
      if (useLocalTasks) {
        setTasks((prev) => prev.filter((t) => t.id !== id));
        return;
      }
      void deleteTaskApi(id)
        .then(() => {
          setTasks((prev) => prev.filter((t) => t.id !== id));
        })
        .catch(() => {});
    },
    [useLocalTasks],
  );

  const toggleDone = useCallback(
    (id: string) => {
      if (useLocalTasks) {
        setTasks((prev) =>
          prev.map((t) => {
            if (t.id !== id) return t;
            const nextStatus: TaskStatus = t.status === "done" ? "todo" : "done";
            return { ...t, status: nextStatus, updatedAt: nowIso() };
          }),
        );
        return;
      }
      setTasks((prev) => {
        const t = prev.find((x) => x.id === id);
        if (!t) return prev;
        const nextStatus: TaskStatus = t.status === "done" ? "todo" : "done";
        void patchTaskApi(id, { status: nextStatus })
          .then((row) => {
            setTasks((p) => p.map((x) => (x.id === id ? row : x)));
          })
          .catch(() => {});
        return prev.map((x) =>
          x.id === id ? { ...x, status: nextStatus, updatedAt: nowIso() } : x,
        );
      });
    },
    [useLocalTasks],
  );

  const value = useMemo(
    () => ({
      tasks,
      loading,
      taskModal: taskModalForUi,
      openCreateModal,
      openTaskEditor,
      closeTaskModal,
      submitNewTask: submitNewTaskCb,
      submitTaskEdit: submitTaskEditCb,
      addTask,
      updateTask,
      removeTask,
      toggleDone,
    }),
    [
      tasks,
      loading,
      taskModalForUi,
      openCreateModal,
      openTaskEditor,
      closeTaskModal,
      submitNewTaskCb,
      submitTaskEditCb,
      addTask,
      updateTask,
      removeTask,
      toggleDone,
    ],
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
