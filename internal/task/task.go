package task

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when no task exists for the given id.
var ErrNotFound = errors.New("task not found")

// Task is the JSON shape for a single task (aligned with OpenAPI Task; tags omitted until normalized).
type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	Priority    *string    `json:"priority"`
	DueDate     *string    `json:"due_date"`
	FocusBucket string     `json:"focus_bucket"`
	ProjectID   *uuid.UUID `json:"project_id"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Tags []string `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Getter loads a task by primary key.
type Getter interface {
	GetTask(ctx context.Context, id uuid.UUID) (Task, error)
}
