package tasks

import (
	"sync"
)

// Status is the task workflow state (see db task_status enum).
type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

// Store is an in-memory task registry (MVP until persistence lands).
type Store struct {
	mu sync.RWMutex
	m  map[string]Status
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{m: make(map[string]Status)}
}

// Put inserts a task id with status todo if it is not yet present (idempotent).
func (s *Store) Put(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[id]; !ok {
		s.m[id] = StatusTodo
	}
}

// SetStatus updates a task's status. It reports whether the id existed.
func (s *Store) SetStatus(id string, st Status) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[id]; !ok {
		return false
	}
	s.m[id] = st
	return true
}

// ClearDone removes every task whose status is done. It returns how many rows were removed.
func (s *Store) ClearDone() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id, st := range s.m {
		if st == StatusDone {
			delete(s.m, id)
			n++
		}
	}
	return n
}

// Delete removes a task by id. It reports whether the id existed.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[id]; !ok {
		return false
	}
	delete(s.m, id)
	return true
}
