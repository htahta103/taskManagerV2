import type { Task } from "../tasks/types";
import { focusBucketForTask } from "../tasks/buckets";
import { useTasks } from "../tasks/TaskContext";

type Props = {
  task: Task;
  onOpen: (t: Task) => void;
  showBucketHint?: boolean;
};

const statusLabel: Record<Task["status"], string> = {
  todo: "To do",
  doing: "Doing",
  done: "Done",
};

export function TaskRow({ task, onOpen, showBucketHint }: Props) {
  const { toggleDone, updateTask } = useTasks();
  const bucket = focusBucketForTask(task);

  return (
    <div className="taskrow">
      <button
        type="button"
        className={`taskrow__check${task.status === "done" ? " taskrow__check--on" : ""}`}
        onClick={(e) => {
          e.stopPropagation();
          toggleDone(task.id);
        }}
        aria-label={task.status === "done" ? "Mark not done" : "Mark done"}
      >
        {task.status === "done" ? "✓" : ""}
      </button>
      <button type="button" className="taskrow__main" onClick={() => onOpen(task)}>
        <span className={`taskrow__title${task.status === "done" ? " taskrow__title--done" : ""}`}>
          {task.title}
        </span>
        <span className="taskrow__meta">
          <span className="taskrow__pill">{statusLabel[task.status]}</span>
          {task.priority ? <span className="taskrow__pill">{task.priority}</span> : null}
          {task.dueDate ? (
            <span className="taskrow__pill" title="Due date">
              Due {task.dueDate}
            </span>
          ) : null}
          {showBucketHint ? (
            <span className="taskrow__pill taskrow__pill--muted">{bucket}</span>
          ) : null}
        </span>
      </button>
      <button
        type="button"
        className={`btn btn--ghost taskrow__pin${task.pinnedToday ? " taskrow__pin--on" : ""}`}
        title={task.pinnedToday ? "Remove from Today" : "Pin to Today"}
        onClick={(e) => {
          e.stopPropagation();
          updateTask(task.id, { pinnedToday: !task.pinnedToday });
        }}
      >
        ★
      </button>
    </div>
  );
}
