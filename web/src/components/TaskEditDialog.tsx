import { useEffect, useRef, useState } from "react";
import { draftToPatch, taskToDraft, type TaskDraft, useTasks } from "../tasks/TaskContext";
import type { Task } from "../tasks/types";

type Props = {
  task: Task;
  onClose: () => void;
};

export function TaskEditDialog({ task, onClose }: Props) {
  const { updateTask, removeTask } = useTasks();
  const ref = useRef<HTMLDialogElement>(null);
  const [draft, setDraft] = useState<TaskDraft>(() => taskToDraft(task));

  useEffect(() => {
    setDraft(taskToDraft(task));
  }, [task]);

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
  }, [task.id, onClose]);

  function save() {
    const t = draft.title.trim();
    if (!t) return;
    updateTask(task.id, draftToPatch({ ...draft, title: t }));
    onClose();
  }

  function del() {
    if (window.confirm("Delete this task?")) {
      removeTask(task.id);
      onClose();
    }
  }

  return (
    <dialog ref={ref} className="dialog" onClose={onClose}>
      <div className="dialog__panel">
        <header className="dialog__header">
          <h2 className="dialog__title">Edit task</h2>
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
            />
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
        </div>
        <footer className="dialog__footer">
          <button type="button" className="btn btn--ghost" onClick={del}>
            Delete
          </button>
          <div className="dialog__footer-right">
            <button type="button" className="btn btn--ghost" onClick={onClose}>
              Cancel
            </button>
            <button type="button" className="btn btn--primary" onClick={save} disabled={!draft.title.trim()}>
              Save
            </button>
          </div>
        </footer>
      </div>
    </dialog>
  );
}
