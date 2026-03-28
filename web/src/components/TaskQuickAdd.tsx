import { useState, type FormEvent } from "react";
import { useTasks } from "../tasks/TaskContext";

export function TaskQuickAdd() {
  const { addTask } = useTasks();
  const [value, setValue] = useState("");

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    const created = addTask(value);
    if (created) setValue("");
  }

  return (
    <form className="quickadd" onSubmit={onSubmit} aria-label="Quick add task">
      <input
        className="input quickadd__input"
        name="title"
        placeholder="Add a task…"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        maxLength={200}
        autoComplete="off"
      />
      <button type="submit" className="btn btn--primary" disabled={!value.trim()}>
        Add
      </button>
    </form>
  );
}
