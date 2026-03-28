package tasks

import (
	"sync"

	"github.com/google/uuid"
)

// MemoryStore is an in-memory task store (until persistence lands).
type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]Task
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[uuid.UUID]Task)}
}

// Put inserts or replaces a task (used by tests and future create handler).
func (s *MemoryStore) Put(t Task) error {
	id, err := uuid.Parse(t.ID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[id] = t
	return nil
}

// Get returns a task by id.
func (s *MemoryStore) Get(id uuid.UUID) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

// Replace updates a task in place (caller holds merged Task).
func (s *MemoryStore) Replace(t Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.MustParse(t.ID)
	s.tasks[id] = t
}
