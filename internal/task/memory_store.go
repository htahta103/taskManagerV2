package task

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore is an in-memory task store for MVP / tests.
type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[string]*Task)}
}

// Insert adds a new task (caller sets ID and timestamps).
func (s *MemoryStore) Insert(_ context.Context, t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
	return nil
}

// NewTaskFromPayload builds a Task from validated create input with server-side fields.
func NewTaskFromPayload(p *CreatePayload) *Task {
	now := time.Now().UTC()
	id := uuid.NewString()

	status := StatusTodo
	if p.Status != nil {
		status = *p.Status
	}
	fb := FocusNone
	if p.FocusBucket != nil {
		fb = *p.FocusBucket
	}

	var desc *string
	if p.Description != nil {
		d := strings.TrimSpace(*p.Description)
		if d == "" {
			desc = nil
		} else {
			desc = &d
		}
	}

	title := strings.TrimSpace(p.Title)

	t := &Task{
		ID:          id,
		Title:       title,
		Description: desc,
		Status:      status,
		Priority:    p.Priority,
		DueDate:     p.DueDate,
		FocusBucket: fb,
		ProjectID:   p.ProjectID,
		AssigneeID:  p.AssigneeID,
		Tags:        []Tag{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return t
}
