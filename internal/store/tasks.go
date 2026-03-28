package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Task is a work item owned by a user.
type Task struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ProjectID   *uuid.UUID
	Title       string
	Description *string
	Status      string
	Priority    *string
	DueDate     *time.Time
	FocusBucket string
	AssigneeID  *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListTasksParams filters and paginates task lists.
type ListTasksParams struct {
	UserID    uuid.UUID
	ProjectID *uuid.UUID
	Status    *string
	View      *string
	Q         *string
	Limit     int
	Cursor    *PageCursor
}

func clampLimit(n int) int {
	if n <= 0 {
		return 50
	}
	if n > 100 {
		return 100
	}
	return n
}

func likePattern(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return "%" + q + "%"
}

func (s *Store) ListTasks(ctx context.Context, p ListTasksParams) ([]Task, *string, error) {
	limit := clampLimit(p.Limit)
	wheres := []string{"user_id = $1"}
	args := []any{p.UserID}
	argN := 2

	if p.ProjectID != nil {
		wheres = append(wheres, fmt.Sprintf("project_id = $%d", argN))
		args = append(args, *p.ProjectID)
		argN++
	}
	if p.Status != nil {
		st := *p.Status
		if st != "todo" && st != "doing" && st != "done" {
			return nil, nil, ErrInvalidInput
		}
		wheres = append(wheres, fmt.Sprintf("status = $%d", argN))
		args = append(args, st)
		argN++
	}
	if p.View != nil {
		switch *p.View {
		case "inbox":
			// no extra filter
		case "today":
			wheres = append(wheres, `(focus_bucket = 'today' OR (due_date IS NOT NULL AND due_date = CURRENT_DATE))`)
		case "next":
			wheres = append(wheres, `(focus_bucket = 'next' OR (due_date IS NOT NULL AND due_date > CURRENT_DATE AND due_date <= CURRENT_DATE + INTERVAL '7 day'))`)
		case "later":
			wheres = append(wheres, `(focus_bucket = 'later' OR due_date IS NULL OR due_date > CURRENT_DATE + INTERVAL '7 day')`)
		default:
			return nil, nil, ErrInvalidInput
		}
	}
	if p.Q != nil && strings.TrimSpace(*p.Q) != "" {
		pat := likePattern(strings.TrimSpace(*p.Q))
		wheres = append(wheres, fmt.Sprintf(`(title ILIKE $%d ESCAPE '\' OR description ILIKE $%d ESCAPE '\')`, argN, argN))
		args = append(args, pat)
		argN++
	}
	if p.Cursor != nil {
		wheres = append(wheres, fmt.Sprintf(
			`(updated_at < $%d OR (updated_at = $%d AND id < $%d))`,
			argN, argN, argN+1,
		))
		args = append(args, p.Cursor.UpdatedAt, p.Cursor.ID)
		argN += 2
	}

	limitArg := argN
	args = append(args, limit+1) // fetch one extra to know if there is a next page

	q := fmt.Sprintf(`
SELECT id, user_id, project_id, title, description, status, priority, due_date, focus_bucket, assignee_id, created_at, updated_at
FROM tasks
WHERE %s
ORDER BY updated_at DESC, id DESC
LIMIT $%d`, strings.Join(wheres, " AND "), limitArg)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.DueDate, &t.FocusBucket, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var next *string
	if len(tasks) > limit {
		last := tasks[limit-1]
		tasks = tasks[:limit]
		enc, err := EncodePageCursor(PageCursor{UpdatedAt: last.UpdatedAt, ID: last.ID})
		if err != nil {
			return nil, nil, err
		}
		next = &enc
	}
	return tasks, next, nil
}

type TaskCreateInput struct {
	Title        string
	Description  *string
	Status       *string
	Priority     *string
	DueDate      *time.Time
	FocusBucket  *string
	ProjectID    *uuid.UUID
	AssigneeID   *uuid.UUID
	TagIDs       []uuid.UUID
}

func (s *Store) CreateTask(ctx context.Context, userID uuid.UUID, in TaskCreateInput) (Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 200 {
		return Task{}, ErrInvalidInput
	}
	if in.Description != nil && len(*in.Description) > 10000 {
		return Task{}, ErrInvalidInput
	}
	status := "todo"
	if in.Status != nil {
		st := *in.Status
		if st != "todo" && st != "doing" && st != "done" {
			return Task{}, ErrInvalidInput
		}
		status = st
	}
	focus := "none"
	if in.FocusBucket != nil {
		fb := *in.FocusBucket
		if fb != "none" && fb != "today" && fb != "next" && fb != "later" {
			return Task{}, ErrInvalidInput
		}
		focus = fb
	}
	if in.ProjectID != nil {
		ok, err := s.ProjectOwnedBy(ctx, userID, *in.ProjectID)
		if err != nil {
			return Task{}, err
		}
		if !ok {
			return Task{}, ErrNotFound
		}
	}
	assignee := in.AssigneeID
	if assignee != nil && *assignee != userID {
		return Task{}, ErrInvalidInput
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Task{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const ins = `
INSERT INTO tasks (user_id, project_id, title, description, status, priority, due_date, focus_bucket, assignee_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, project_id, title, description, status, priority, due_date, focus_bucket, assignee_id, created_at, updated_at`
	var t Task
	err = tx.QueryRow(ctx, ins, userID, in.ProjectID, title, in.Description, status, in.Priority, in.DueDate, focus, assignee).Scan(
		&t.ID, &t.UserID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.DueDate, &t.FocusBucket, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}

	if len(in.TagIDs) > 0 {
		var n int
		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM tags WHERE user_id = $1 AND id = ANY($2::uuid[])`, userID, in.TagIDs).Scan(&n)
		if err != nil {
			return Task{}, err
		}
		if n != len(in.TagIDs) {
			return Task{}, ErrNotFound
		}
		for _, tid := range in.TagIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, t.ID, tid); err != nil {
				return Task{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *Store) GetTask(ctx context.Context, userID, taskID uuid.UUID) (Task, error) {
	const q = `
SELECT id, user_id, project_id, title, description, status, priority, due_date, focus_bucket, assignee_id, created_at, updated_at
FROM tasks WHERE id = $1 AND user_id = $2`
	var t Task
	err := s.pool.QueryRow(ctx, q, taskID, userID).Scan(
		&t.ID, &t.UserID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.DueDate, &t.FocusBucket, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

// TaskUpdateValues lists fields to change. Each Set* flag means the corresponding field is present in the JSON patch.
type TaskUpdateValues struct {
	SetTitle        bool
	Title           string
	SetDescription  bool
	Description     *string
	SetStatus       bool
	Status          string
	SetPriority     bool
	Priority        *string
	SetDueDate      bool
	DueDate         *time.Time
	SetFocusBucket  bool
	FocusBucket     string
	SetProjectID    bool
	ProjectID       *uuid.UUID
	SetAssigneeID   bool
	AssigneeID      *uuid.UUID
}

func (s *Store) UpdateTask(ctx context.Context, userID, taskID uuid.UUID, patch TaskUpdateValues) (Task, error) {
	cur, err := s.GetTask(ctx, userID, taskID)
	if err != nil {
		return Task{}, err
	}
	title := cur.Title
	desc := cur.Description
	status := cur.Status
	priority := cur.Priority
	due := cur.DueDate
	focus := cur.FocusBucket
	proj := cur.ProjectID
	assignee := cur.AssigneeID

	if patch.SetTitle {
		v := strings.TrimSpace(patch.Title)
		if v == "" || len(v) > 200 {
			return Task{}, ErrInvalidInput
		}
		title = v
	}
	if patch.SetDescription {
		if patch.Description != nil && len(*patch.Description) > 10000 {
			return Task{}, ErrInvalidInput
		}
		desc = patch.Description
	}
	if patch.SetStatus {
		st := patch.Status
		if st != "todo" && st != "doing" && st != "done" {
			return Task{}, ErrInvalidInput
		}
		status = st
	}
	if patch.SetPriority {
		priority = patch.Priority
		if priority != nil {
			p := *priority
			if p != "low" && p != "medium" && p != "high" {
				return Task{}, ErrInvalidInput
			}
		}
	}
	if patch.SetDueDate {
		due = patch.DueDate
	}
	if patch.SetFocusBucket {
		fb := patch.FocusBucket
		if fb != "none" && fb != "today" && fb != "next" && fb != "later" {
			return Task{}, ErrInvalidInput
		}
		focus = fb
	}
	if patch.SetProjectID {
		p := patch.ProjectID
		proj = p
		if p != nil {
			ok, err := s.ProjectOwnedBy(ctx, userID, *p)
			if err != nil {
				return Task{}, err
			}
			if !ok {
				return Task{}, ErrNotFound
			}
		}
	}
	if patch.SetAssigneeID {
		a := patch.AssigneeID
		assignee = a
		if a != nil && *a != userID {
			return Task{}, ErrInvalidInput
		}
	}

	const q = `
UPDATE tasks SET
  title = $3,
  description = $4,
  status = $5,
  priority = $6,
  due_date = $7,
  focus_bucket = $8,
  project_id = $9,
  assignee_id = $10,
  updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, project_id, title, description, status, priority, due_date, focus_bucket, assignee_id, created_at, updated_at`
	var t Task
	err = s.pool.QueryRow(ctx, q, taskID, userID, title, desc, status, priority, due, focus, proj, assignee).Scan(
		&t.ID, &t.UserID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.DueDate, &t.FocusBucket, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *Store) DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND user_id = $2`, taskID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
