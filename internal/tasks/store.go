package tasks

import (
	"sync"
)

// Store is an in-memory task registry (MVP until persistence lands).
type Store struct {
	mu sync.RWMutex
	m  map[string]struct{}
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{m: make(map[string]struct{})}
}

// Put inserts a task id (idempotent).
func (s *Store) Put(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[id] = struct{}{}
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
