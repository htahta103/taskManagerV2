import { TaskList } from "../components/TaskList";
import { tasksForBucket } from "../tasks/buckets";
import { useTasks } from "../tasks/TaskContext";

export function TodayPage() {
  const { tasks, loading, openTaskEditor } = useTasks();
  const now = new Date();
  const today = tasksForBucket(tasks, "today", now);
  const next = tasksForBucket(tasks, "next", now);
  const later = tasksForBucket(tasks, "later", now);

  return (
    <div className="page">
      <header className="page__header">
        <h1>Today</h1>
        <p className="muted">
          Three columns: Today (due / overdue / pinned / high), Next (due within a week or medium priority), and Later
          (everything else that is still open).
        </p>
      </header>
      {loading ? (
        <p className="muted">Loading tasks…</p>
      ) : (
        <div className="plan">
          <section className="plan__col" aria-labelledby="plan-today">
            <h2 id="plan-today" className="plan__heading">
              Today
            </h2>
            <p className="plan__hint muted">Due today, overdue, pinned, or high priority.</p>
            <TaskList
              tasks={today}
              empty="Nothing scheduled for today."
              onOpen={openTaskEditor}
            />
          </section>
          <section className="plan__col" aria-labelledby="plan-next">
            <h2 id="plan-next" className="plan__heading">
              Next
            </h2>
            <p className="plan__hint muted">Due in the next week or medium priority.</p>
            <TaskList
              tasks={next}
              empty="No upcoming items in this window."
              onOpen={openTaskEditor}
            />
          </section>
          <section className="plan__col" aria-labelledby="plan-later">
            <h2 id="plan-later" className="plan__heading">
              Later
            </h2>
            <p className="plan__hint muted">Everything else that is still open.</p>
            <TaskList
              tasks={later}
              empty="Your backlog is clear."
              onOpen={openTaskEditor}
            />
          </section>
        </div>
      )}
    </div>
  );
}
