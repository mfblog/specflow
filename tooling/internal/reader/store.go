package reader

import (
	"path/filepath"
	"sync"
)

type Store struct {
	repoRoot    string
	mu          sync.RWMutex
	snapshot    Snapshot
	version     int64
	subscribers map[chan int64]struct{}
}

func NewStore(repoRoot string) (*Store, error) {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	store := &Store{
		repoRoot:    absRoot,
		subscribers: map[chan int64]struct{}{},
	}
	store.Rebuild()
	return store, nil
}

func (s *Store) RepoRoot() string {
	return s.repoRoot
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

func (s *Store) Rebuild() {
	s.mu.Lock()
	s.version++
	snapshot := BuildSnapshot(s.repoRoot)
	snapshot.Version = s.version
	s.snapshot = snapshot
	version := s.version
	subscribers := make([]chan int64, 0, len(s.subscribers))
	for subscriber := range s.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	s.mu.Unlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- version:
		default:
		}
	}
}

func (s *Store) Subscribe() (<-chan int64, func()) {
	ch := make(chan int64, 4)
	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	current := s.version
	s.mu.Unlock()
	ch <- current
	cancel := func() {
		s.mu.Lock()
		if _, ok := s.subscribers[ch]; ok {
			delete(s.subscribers, ch)
			close(ch)
		}
		s.mu.Unlock()
	}
	return ch, cancel
}
