# PRD: Task Manager V2

## Overview
Task Manager V2 is a fast, opinionated task management app for individuals and small teams.
It supports projects, tasks, comments, and basic collaboration, with a clean UI and strong
keyboard ergonomics.

## Goals
- Create and manage tasks quickly (keyboard-first flows).
- Organize work by project and status.
- Support lightweight collaboration (sharing a project, comments, activity).
- Be reliable and responsive on desktop and mobile.

## Non-goals (for v2)
- Complex enterprise permission matrices (beyond owner/member + read/write).
- Native mobile apps.
- Offline-first sync.

## Personas
- **Individual**: manages personal tasks across a few projects.
- **Team member**: collaborates on a shared project with 2–10 people.

## Core Concepts
- **User**: an account that owns or collaborates on projects.
- **Project**: container for tasks and membership.
- **Task**: actionable item with status, due date, priority.
- **Comment**: text attached to a task.

## User Stories (MVP)
### Authentication
- As a user, I can sign up with email + password.
- As a user, I can sign in/out.
- As a signed-in user, my session persists across refresh (until sign out).

### Projects
- As a user, I can create a project with a name.
- As a user, I can rename and archive a project.
- As a user, I can see a list of my projects.
- As a user, I can invite another user to a project by email.
- As a project owner, I can remove a member.

### Tasks
- As a user, I can create a task in a project with: title (required), optional description.
- As a user, I can update task fields:
  - status: todo / in_progress / done
  - priority: low / medium / high
  - due date (optional)
- As a user, I can reorder tasks within a status column (simple ordering).
- As a user, I can archive (soft-delete) a task.

### Task Views
- As a user, I can view tasks for a project grouped by status.
- As a user, I can filter tasks by status and search by title.
- As a user, I can view “My Tasks” across all projects (assigned to me).

### Comments & Activity
- As a user, I can add comments to a task.
- As a user, I can see an activity feed for a task (created/updated/commented).

## Acceptance Criteria (high-level)
- Unauthorized requests are rejected; all task/project endpoints require auth.
- Project membership is enforced for reads/writes.
- All create/update operations validate input (lengths, enums, required fields).
- UI works on desktop and mobile widths; key screens are usable without horizontal scrolling.

## Basic Screens
- Sign in / Sign up
- Project list
- Project board (tasks grouped by status)
- Task detail (description, comments, activity)

## Technical Notes / Constraints
- This PRD is intended to be used by the `it-company` workflow to drive architecture,
  backend API, and frontend implementation.

