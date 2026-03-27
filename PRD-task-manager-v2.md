# PRD: Task Manager V2

## 1) Summary

Task Manager V2 is a lightweight, fast, personal task manager for individuals and small teams. It focuses on capturing tasks quickly, organizing them into projects, and getting a clear “what should I do next?” view.

This PRD is intentionally scoped for an MVP that can ship end-to-end with tests and deploys, and can be extended later.

## 2) Goals

- Capture tasks in under 5 seconds.
- Organize tasks into projects and simple labels/tags.
- Provide a focused “Today / Next / Later” workflow.
- Support due dates, reminders (at least via UI indicators), and basic search.
- Make it feel responsive on mobile and desktop.

## 3) Non-goals (MVP)

- Real-time collaboration, live cursors, or complex multi-user permissions.
- Calendar sync, email ingest, or external integrations.
- Offline-first sync.
- AI planning features.

## 4) Personas

- Individual user: wants a clear daily plan, minimal overhead.
- Team lead (small team): wants shared projects and assignment, but can accept “simple rules”.

## 5) Core Concepts / Entities

- User
- Project
- Task
- Tag (optional; can be implemented as free-form labels)

### Task fields (MVP)

- title (required)
- description (optional)
- status: todo | doing | done (or todo | done if you want simpler)
- priority: low | medium | high (optional)
- due_date (optional)
- created_at, updated_at
- project_id (optional)
- tags (optional; can be a many-to-many or a string array depending on stack)
- assignee_id (optional; only relevant if multi-user is included in MVP)

## 6) User Stories (MVP)

### Auth & accounts

- As a user, I can sign up, sign in, and sign out.
- As a user, I can view and edit my profile basics (name, email).

### Projects

- As a user, I can create, rename, and archive projects.
- As a user, I can view a project and its tasks.

### Tasks

- As a user, I can create a task (title required) from anywhere.
- As a user, I can edit a task (title, description, status, due date, priority, project).
- As a user, I can mark a task done and undo it.
- As a user, I can delete a task.
- As a user, I can move a task between statuses (drag/drop is optional; buttons acceptable).
- As a user, I can filter tasks by project and status.
- As a user, I can search tasks by title/description.

### Today / Next / Later workflow

- As a user, I can see a “Today” view that shows tasks due today and tasks I’ve manually flagged for today.
- As a user, I can see a “Next” view that shows upcoming tasks (no due date or due later).
- As a user, I can see a “Later” view for deprioritized tasks.

Note: If implementing manual “bucket” assignment is too much, you may derive:
- Today: due today OR priority=high
- Next: due within 7 days OR priority=medium
- Later: everything else

## 7) Acceptance Criteria (MVP)

- Authenticated users only see their own data (or their team’s data if teams are implemented).
- Creating a task with only a title works; it appears immediately in lists.
- Task list pages load in under 1s on a typical dev machine with 1k tasks.
- Search returns results in under 250ms for 1k tasks (local dev environment).
- Basic validation:
  - title required, max length 200
  - description max length 10k
- Error states are human-readable; empty states are clear.

## 8) Admin / Team Scope (Optional MVP extension)

If time allows, support a “workspace/team” concept:
- Invite user by email to a workspace
- Assign tasks to members

If not included, keep everything single-user.

## 9) UI Requirements (MVP)

- Responsive layout.
- Primary navigation:
  - Inbox (all open tasks)
  - Today
  - Projects
  - Search
- Quick add task input accessible from all main pages.
- Task detail drawer/modal for editing.

## 10) Technical Constraints / Preferences

- Prefer a conventional, boring stack.
- API should be versioned (e.g. `/api/v1`).
- Use server-side auth (sessions or JWT) with secure defaults.
- Provide migrations and seed/dev data helpers.

## 11) Out of Scope (Explicit)

- Push notifications
- Native mobile apps
- Complex recurring tasks

