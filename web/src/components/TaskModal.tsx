import { useEffect, useId, useState } from "react";
import type { FocusBucket, Task, TaskCreateBody, TaskPatchBody, TaskPriority, TaskStatus } from "../tasks/types";

type Props = {
  open: boolean;
  mode: "create" | "edit";
  task: Task | null;
  onClose: () => void;
  onCreate: (body: TaskCreateBody) => Promise<void>;
  onUpdate: (id: string, patch: TaskPatchBody) => Promise<void>;
};

const emptyForm = {
  title: "",
  description: "",
  status: "todo" as TaskStatus,
  priority: "" as "" | TaskPriority,
  due_date: "",
  focus_bucket: "none" as FocusBucket,
};

export function TaskModal({ open, mode, task, onClose, onCreate, onUpdate }: Props) {
  const titleId = useId();
  const [form, setForm] = useState(emptyForm);
  const [titleError, setTitleError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) return;
    setTitleError(null);
    setSubmitError(null);
    if (mode === "edit" && task) {
      setForm({
        title: task.title,
        description: task.description ?? "",
        status: task.status,
        priority: task.priority ?? "",
        due_date: task.due_date ?? "",
        focus_bucket: task.focus_bucket,
      });
    } else {
      setForm(emptyForm);
    }
  }, [open, mode, task]);

  if (!open) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = form.title.trim();
    if (!trimmed) {
      setTitleError("Title is required.");
      return;
    }
    setTitleError(null);
    setSubmitError(null);
    setSubmitting(true);
    try {
      if (mode === "create") {
        const body: TaskCreateBody = {
          title: trimmed,
          ...(form.description.trim() ? { description: form.description.trim() } : {}),
          status: form.status,
          ...(form.priority ? { priority: form.priority } : {}),
          ...(form.due_date ? { due_date: form.due_date } : {}),
          focus_bucket: form.focus_bucket,
        };
        await onCreate(body);
      } else if (task) {
        const patch: TaskPatchBody = {
          title: trimmed,
          description: form.description.trim() ? form.description.trim() : null,
          status: form.status,
          priority: form.priority || null,
          due_date: form.due_date || null,
          focus_bucket: form.focus_bucket,
        };
        await onUpdate(task.id, patch);
      }
      onClose();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Something went wrong.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      className="modal-backdrop"
      role="presentation"
      onMouseDown={(ev) => {
        if (ev.target === ev.currentTarget) onClose();
      }}
    >
      <div
        className="modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onMouseDown={(ev) => ev.stopPropagation()}
      >
        <div className="modal__head">
          <h2 id={titleId} className="modal__title">
            {mode === "create" ? "New task" : "Edit task"}
          </h2>
          <button type="button" className="btn btn--ghost modal__close" onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>
        <form className="modal__body stack" onSubmit={handleSubmit}>
          <div className="field">
            <label className="field__label" htmlFor={`${titleId}-title`}>
              Title
            </label>
            <input
              id={`${titleId}-title`}
              className="input"
              value={form.title}
              onChange={(ev) => {
                setTitleError(null);
                setForm((f) => ({ ...f, title: ev.target.value }));
              }}
              autoComplete="off"
              aria-invalid={Boolean(titleError)}
              aria-describedby={titleError ? `${titleId}-title-err` : undefined}
            />
            {titleError ? (
              <p id={`${titleId}-title-err`} className="error" role="alert">
                {titleError}
              </p>
            ) : null}
          </div>
          <div className="field">
            <label className="field__label" htmlFor={`${titleId}-desc`}>
              Description
            </label>
            <textarea
              id={`${titleId}-desc`}
              className="input input--textarea"
              rows={4}
              value={form.description}
              onChange={(ev) => setForm((f) => ({ ...f, description: ev.target.value }))}
            />
          </div>
          <div className="modal__row">
            <div className="field">
              <label className="field__label" htmlFor={`${titleId}-status`}>
                Status
              </label>
              <select
                id={`${titleId}-status`}
                className="input"
                value={form.status}
                onChange={(ev) => setForm((f) => ({ ...f, status: ev.target.value as TaskStatus }))}
              >
                <option value="todo">To do</option>
                <option value="doing">Doing</option>
                <option value="done">Done</option>
              </select>
            </div>
            <div className="field">
              <label className="field__label" htmlFor={`${titleId}-priority`}>
                Priority
              </label>
              <select
                id={`${titleId}-priority`}
                className="input"
                value={form.priority}
                onChange={(ev) =>
                  setForm((f) => ({ ...f, priority: ev.target.value as "" | TaskPriority }))
                }
              >
                <option value="">—</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>
          <div className="modal__row">
            <div className="field">
              <label className="field__label" htmlFor={`${titleId}-due`}>
                Due date
              </label>
              <input
                id={`${titleId}-due`}
                type="date"
                className="input"
                value={form.due_date}
                onChange={(ev) => setForm((f) => ({ ...f, due_date: ev.target.value }))}
              />
            </div>
            <div className="field">
              <label className="field__label" htmlFor={`${titleId}-focus`}>
                Focus
              </label>
              <select
                id={`${titleId}-focus`}
                className="input"
                value={form.focus_bucket}
                onChange={(ev) => setForm((f) => ({ ...f, focus_bucket: ev.target.value as FocusBucket }))}
              >
                <option value="none">None</option>
                <option value="today">Today</option>
                <option value="next">Next</option>
                <option value="later">Later</option>
              </select>
            </div>
          </div>
          {submitError ? (
            <p className="error" role="alert">
              {submitError}
            </p>
          ) : null}
          <div className="modal__actions">
            <button type="button" className="btn btn--ghost" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn--primary" disabled={submitting}>
              {submitting ? "Saving…" : mode === "create" ? "Create" : "Save"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
