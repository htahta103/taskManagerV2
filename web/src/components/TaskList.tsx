import type { Task } from "../tasks/types";
import { TaskRow } from "./TaskRow";

type Props = {
  tasks: Task[];
  empty: string;
  onOpen: (t: Task) => void;
  showBucketHint?: boolean;
};

export function TaskList({ tasks, empty, onOpen, showBucketHint }: Props) {
  if (tasks.length === 0) {
    return (
      <section className="tasklist tasklist--empty">
        <p className="muted">{empty}</p>
      </section>
    );
  }

  return (
    <ul className="tasklist" aria-label="Tasks">
      {tasks.map((t) => (
        <li key={t.id} className="tasklist__item">
          <TaskRow task={t} onOpen={onOpen} showBucketHint={showBucketHint} />
        </li>
      ))}
    </ul>
  );
}
