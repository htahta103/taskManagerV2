import { useState } from "react";
import { TaskModal } from "../components/TaskModal";
import { useTasks } from "../tasks/useTasks";
import type { Task } from "../tasks/types";

export function InboxPage() {
  const { tasks, loading, error, refresh, createTask, updateTask } = useTasks({ view: "inbox" });
  const [modalOpen, setModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<"create" | "edit">("create");
  const [editing, setEditing] = useState<Task | null>(null);

  function openCreate() {
    setModalMode("create");
    setEditing(null);
    setModalOpen(true);
  }

  function openEdit(task: Task) {
    setModalMode("edit");
    setEditing(task);
    setModalOpen(true);
  }

  function closeModal() {
    setModalOpen(false);
    setEditing(null);
  }

  return (
    <div className="page">
      <header className="page__header page__header--row">
        <div>
          <h1>Inbox</h1>
          <p className="muted">All open tasks land here.</p>
        </div>
        <button type="button" className="btn btn--primary" onClick={openCreate}>
          Add task
        </button>
      </header>

      {loading ? (
        <p className="muted">Loading tasks…</p>
      ) : error ? (
        <div className="task-panel task-panel--error">
          <p className="error">{error}</p>
          <button type="button" className="btn" onClick={() => void refresh()}>
            Retry
          </button>
        </div>
      ) : tasks.length === 0 ? (
        <section className="empty">
          <p>No tasks yet. Use <strong>Add task</strong> to create one.</p>
        </section>
      ) : (
        <ul className="task-list" aria-label="Tasks">
          {tasks.map((task) => (
            <li key={task.id}>
              <button type="button" className="task-row" onClick={() => openEdit(task)}>
                <span className="task-row__title">{task.title}</span>
                <span className="task-row__meta muted">
                  {task.status}
                  {task.priority ? ` · ${task.priority}` : ""}
                  {task.due_date ? ` · due ${task.due_date}` : ""}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}

      <TaskModal
        open={modalOpen}
        mode={modalMode}
        task={editing}
        onClose={closeModal}
        onCreate={async (body) => {
          await createTask(body);
        }}
        onUpdate={async (id, patch) => {
          await updateTask(id, patch);
        }}
      />
    </div>
  );
}
