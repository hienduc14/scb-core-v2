package repositories

import "sync"

type InMemoryStatusStore struct {
	mu    sync.RWMutex
	state map[string]string
}

func NewInMemoryStatusStore() *InMemoryStatusStore {
	return &InMemoryStatusStore{state: make(map[string]string)}
}

func (s *InMemoryStatusStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = value
}

func (s *InMemoryStatusStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.state[key]
	return value, ok
}
