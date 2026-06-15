package reader

import (
	"path/filepath"
	"sync"
)

type Store struct {
	repoRoot string
	mu       sync.Mutex
	version  int64
}

func NewStore(repoRoot string) (*Store, error) {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	store := &Store{repoRoot: absRoot}
	store.RefreshSnapshot()
	return store, nil
}

func (s *Store) RepoRoot() string {
	return s.repoRoot
}

func (s *Store) RefreshSnapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.version++
	snapshot := BuildSnapshot(s.repoRoot)
	snapshot.Version = s.version
	return snapshot
}
