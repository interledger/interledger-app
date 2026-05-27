// Package store provides TokenStore implementations.
//
// Production path uses Redis (see redis.go). 
// The in-memory Memory store is
// kept around for unit tests of handler logic where wiring a real Redis would
// be overkill.
package store

import (
	"context"
	"sync"

	"gitlab.com/fynbos/backend/providers/plaid"
)

// Memory is a thread-safe in-process TokenStore. Tests only.
type Memory struct {
	mu     sync.RWMutex
	tokens map[string]plaid.TokenSet
}

func NewMemory() *Memory {
	return &Memory{tokens: make(map[string]plaid.TokenSet)}
}

func (s *Memory) Get(_ context.Context, userID string) (plaid.TokenSet, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[userID]
	return t, ok, nil
}

func (s *Memory) Put(_ context.Context, userID string, t plaid.TokenSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[userID] = t
	return nil
}

func (s *Memory) Delete(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, userID)
	return nil
}
