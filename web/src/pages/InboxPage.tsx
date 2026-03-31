import { TaskList } from "../components/TaskList";
import { useTasks } from "../tasks/TaskContext";

export function InboxPage() {
  const { tasks, loading, openTaskEditor } = useTasks();
  const open = tasks.filter((t) => t.status !== "done").sort(byUpdatedDesc);

  return (
    <div className="page">
      <header className="page__header">
        <h1>Inbox</h1>
        <p className="muted">Everything that still needs attention.</p>
      </header>
      {loading ? (
        <p className="muted">Loading tasks…</p>
      ) : (
        <TaskList
          tasks={open}
          empty="No open tasks — add one above."
          onOpen={openTaskEditor}
          showBucketHint
        />
      )}
    </div>
  );
}

function byUpdatedDesc(a: { updatedAt: string }, b: { updatedAt: string }): number {
  return b.updatedAt.localeCompare(a.updatedAt);
}
