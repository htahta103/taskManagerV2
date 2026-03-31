import { useTasks } from "../tasks/TaskContext";

export function TaskQuickAdd() {
  const { openCreateModal } = useTasks();

  return (
    <div className="quickadd">
      <button type="button" className="btn btn--primary quickadd__cta" onClick={openCreateModal}>
        Add task
      </button>
    </div>
  );
}
