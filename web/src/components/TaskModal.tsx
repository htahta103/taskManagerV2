import { useEffect, useMemo, useRef, useState } from "react";
import {
  taskToDraft,
  type TaskDraft,
  type TaskModalState,
  useTasks,
} from "../tasks/TaskContext";
import type { Task } from "../tasks/types";

const emptyDraft: TaskDraft = {
  title: "",
  description: "",
  status: "todo",
  priority: "",
  dueDate: "",
  pinnedToday: false,
};

type Props = {
  state: Exclude<TaskModalState, null>;
  onClose: () => void;
};

export function TaskModal({ state, onClose }: Props) {
  const { submitNewTask, submitTaskEdit, removeTask } = useTasks();
  const ref = useRef<HTMLDialogElement>(null);
  const isCreate = state.mode === "create";
  const task: Task | undefined = state.mode === "edit" ? state.task : undefined;

  const initialDraft = useMemo(() => (isCreate ? emptyDraft : taskToDraft(task!)), [isCreate, task]);

  const [draft, setDraft] = useState<TaskDraft>(initialDraft);
  const [saveAttempted, setSaveAttempted] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    setDraft(initialDraft);
    setSaveAttempted(false);
    setSubmitError(null);
  }, [initialDraft, state.mode, task?.id]);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.showModal();
    const onCancel = () => onClose();
    el.addEventListener("cancel", onCancel);
    return () => {
      el.removeEventListener("cancel", onCancel);
      el.close();
    };
  }, [state.mode, task?.id, onClose]);

  const titleError = saveAttempted && !draft.title.trim();

  async function save() {
    setSaveAttempted(true);
    setSubmitError(null);
    if (!draft.title.trim()) return;
    try {
      if (isCreate) {
        await submitNewTask(draft);
      } else {
        await submitTaskEdit(task!, draft);
      }
    } catch (e) {
      setSubmitError(e instanceof Error ? e.message : "Something went wrong");
    }
  }

  function del() {
    if (!task || isCreate) return;
    if (window.confirm("Delete this task?")) {
      removeTask(task.id);
      onClose();
    }
  }

  return (
    <dialog ref={ref} className="dialog" onClose={onClose}>
      <div className="dialog__panel">
        <header className="dialog__header">
          <h2 className="dialog__title">{isCreate ? "Add task" : "Edit task"}</h2>
          <button type="button" className="btn btn--ghost dialog__close" onClick={onClose} aria-label="Close">
            ×
          </button>
        </header>
        <div className="dialog__body stack">
          <label className="field">
            <span className="field__label">Title</span>
            <input
              className="input"
              value={draft.title}
              maxLength={200}
              onChange={(e) => setDraft((d) => ({ ...d, title: e.target.value }))}
              aria-invalid={titleError ? true : undefined}
              aria-describedby={titleError ? "task-modal-title-error" : undefined}
            />
            {titleError ? (
              <span id="task-modal-title-error" className="field__error" role="alert">
                Title is required
              </span>
            ) : null}
          </label>
          <label className="field">
            <span className="field__label">Description</span>
            <textarea
              className="input input--textarea"
              value={draft.description}
              maxLength={10_000}
              rows={4}
              onChange={(e) => setDraft((d) => ({ ...d, description: e.target.value }))}
            />
          </label>
          <div className="dialog__row">
            <label className="field">
              <span className="field__label">Status</span>
              <select
                className="input"
                value={draft.status}
                onChange={(e) =>
                  setDraft((d) => ({
                    ...d,
                    status: e.target.value as TaskDraft["status"],
                  }))
                }
              >
                <option value="todo">To do</option>
                <option value="doing">Doing</option>
                <option value="done">Done</option>
              </select>
            </label>
            <label className="field">
              <span className="field__label">Priority</span>
              <select
                className="input"
                value={draft.priority}
                onChange={(e) =>
                  setDraft((d) => ({
                    ...d,
                    priority: e.target.value as TaskDraft["priority"],
                  }))
                }
              >
                <option value="">None</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </label>
          </div>
          <label className="field">
            <span className="field__label">Due date</span>
            <input
              className="input"
              type="date"
              value={draft.dueDate}
              onChange={(e) => setDraft((d) => ({ ...d, dueDate: e.target.value }))}
            />
          </label>
          <label className="field field--inline">
            <input
              type="checkbox"
              checked={draft.pinnedToday}
              onChange={(e) => setDraft((d) => ({ ...d, pinnedToday: e.target.checked }))}
            />
            <span className="field__label">Pin to Today</span>
          </label>
          {submitError ? (
            <p className="field__error" role="alert">
              {submitError}
            </p>
          ) : null}
        </div>
        <footer className="dialog__footer">
          {isCreate ? (
            <span />
          ) : (
            <button type="button" className="btn btn--ghost" onClick={del}>
              Delete
            </button>
          )}
          <div className="dialog__footer-right">
            <button type="button" className="btn btn--ghost" onClick={onClose}>
              Cancel
            </button>
            <button type="button" className="btn btn--primary" onClick={save}>
              {isCreate ? "Create" : "Save"}
            </button>
          </div>
        </footer>
      </div>
    </dialog>
  );
}
