// Package store provides TokenStore implementations. The POC ships only the
// in-memory variant; a Postgres-backed store with column-level encryption
// will replace it post-POC (see documentation/poc/plaid/architecture.md §7).
package store

import (
	"sync"

	"gitlab.com/fynbos/backend/providers/plaid"
)

// Memory is a thread-safe in-process TokenStore. Backend restarts wipe state —
// acceptable for the POC; callers re-link when this happens.
type Memory struct {
	mu     sync.RWMutex
	tokens map[string]plaid.TokenSet
}

// NewMemory returns a ready-to-use in-memory TokenStore.
func NewMemory() *Memory {
	return &Memory{tokens: make(map[string]plaid.TokenSet)}
}

func (s *Memory) Get(userID string) (plaid.TokenSet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[userID]
	return t, ok
}

func (s *Memory) Put(userID string, t plaid.TokenSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[userID] = t
}

func (s *Memory) Delete(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, userID)
}
