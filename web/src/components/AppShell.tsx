import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { TaskModal } from "./TaskModal";
import { TaskQuickAdd } from "./TaskQuickAdd";
import { useTasks } from "../tasks/TaskContext";

const nav: { to: string; label: string; end?: boolean }[] = [
  { to: "/", label: "Inbox", end: true },
  { to: "/today", label: "Today" },
  { to: "/projects", label: "Projects" },
  { to: "/search", label: "Search" },
];

export function AppShell() {
  const { user, logout } = useAuth();
  const { taskModal, closeTaskModal } = useTasks();
  const navigate = useNavigate();

  return (
    <div className="shell">
      <header className="shell__top">
        <div className="shell__brand">
          <span className="shell__logo" aria-hidden>
            ✓
          </span>
          <span className="shell__title">Task Manager</span>
        </div>
        <div className="shell__actions">
          <span className="shell__user muted" title={user?.email}>
            {user?.name || user?.email}
          </span>
          <button
            type="button"
            className="btn btn--ghost"
            onClick={async () => {
              await logout();
              navigate("/login", { replace: true });
            }}
          >
            Sign out
          </button>
        </div>
      </header>
      <div className="shell__body">
        <aside className="shell__nav" aria-label="Primary">
          <nav className="nav">
            {nav.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) => `nav__link${isActive ? " nav__link--active" : ""}`}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </aside>
        <main className="shell__main">
          <div className="shell__toolbar">
            <TaskQuickAdd />
          </div>
          <Outlet />
        </main>
      </div>
      {taskModal ? (
        <TaskModal
          key={taskModal.mode === "edit" ? taskModal.task.id : "create"}
          state={taskModal}
          onClose={closeTaskModal}
        />
      ) : null}
    </div>
  );
}
