package task

import (
	"context"
	"errors"
)

// ErrNotFound is returned when no task exists for the given id.
var ErrNotFound = errors.New("task not found")

// Getter loads a task by primary key.
type Getter interface {
	GetTask(ctx context.Context, id string) (Task, error)
}
