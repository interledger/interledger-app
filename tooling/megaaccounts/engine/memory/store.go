// Package memory provides an in-memory implementation of engine.Store.
package memory

import (
	"fmt"
	"sync"

	"megaaccounts/engine"
)

// Store is a thread-safe, in-memory implementation of engine.Store.
type Store struct {
	mu        sync.RWMutex
	providers map[string]engine.Provider
	accounts  map[string]engine.Account
	lines     []engine.JournalLine
	legacy    []engine.LedgerEntry
}

// New returns an empty in-memory Store.
func New() *Store {
	return &Store{
		providers: make(map[string]engine.Provider),
		accounts:  make(map[string]engine.Account),
	}
}

func (s *Store) SaveProvider(p engine.Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[p.ID] = p
	return nil
}

func (s *Store) GetProvider(id string) (engine.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.providers[id]
	if !ok {
		return engine.Provider{}, fmt.Errorf("provider %q not found", id)
	}
	return p, nil
}

func (s *Store) ListProviders() ([]engine.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]engine.Provider, 0, len(s.providers))
	for _, p := range s.providers {
		result = append(result, p)
	}
	return result, nil
}

func (s *Store) SaveAccount(a engine.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[a.ID] = a
	return nil
}

func (s *Store) GetAccount(id string) (engine.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accounts[id]
	if !ok {
		return engine.Account{}, fmt.Errorf("account %q not found", id)
	}
	return a, nil
}

func (s *Store) ListAccounts() ([]engine.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]engine.Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		result = append(result, a)
	}
	return result, nil
}

func (s *Store) FindSystemAccount(providerID string, currency engine.Currency) (engine.Account, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.Type == engine.AccountTypeSystem && a.ProviderID == providerID && a.Currency == currency {
			return a, true, nil
		}
	}
	return engine.Account{}, false, nil
}

func (s *Store) FindLiquidityAccount(providerID string, currency engine.Currency) (engine.Account, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.Type == engine.AccountTypeLiquidity && a.ProviderID == providerID && a.Currency == currency {
			return a, true, nil
		}
	}
	return engine.Account{}, false, nil
}

func (s *Store) FindPositionAccount(liquidityAccountID, counterpartyID string) (engine.Account, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.Type == engine.AccountTypePosition && a.LiquidityAccountID == liquidityAccountID && a.CounterpartyID == counterpartyID {
			return a, true, nil
		}
	}
	return engine.Account{}, false, nil
}

func (s *Store) FindUserAccount(userID, providerID string, currency engine.Currency) (engine.Account, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.Type == engine.AccountTypeUser && a.UserID == userID && a.ProviderID == providerID && a.Currency == currency {
			return a, true, nil
		}
	}
	return engine.Account{}, false, nil
}

func (s *Store) PostEntries(entries []engine.LedgerEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.legacy = append(s.legacy, entries...)
	return nil
}

func (s *Store) GetEntriesByAccount(accountID string) ([]engine.LedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]engine.LedgerEntry, 0, len(s.lines)*2)
	result = append(result, engine.ExpandJournalLines(s.lines, accountID)...)
	for _, e := range s.legacy {
		if e.AccountID == accountID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Store) GetEntriesByEvent(eventID string) ([]engine.LedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]engine.LedgerEntry, 0, len(s.lines)*2)
	for _, e := range engine.ExpandJournalLines(s.lines, "") {
		if e.EventID == eventID {
			result = append(result, e)
		}
	}
	for _, e := range s.legacy {
		if e.EventID == eventID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Store) GetAllEntries() ([]engine.LedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]engine.LedgerEntry, 0, len(s.lines)*2+len(s.legacy))
	result = append(result, engine.ExpandJournalLines(s.lines, "")...)
	result = append(result, s.legacy...)
	return result, nil
}

func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers = make(map[string]engine.Provider)
	s.accounts = make(map[string]engine.Account)
	s.lines = nil
	s.legacy = nil
	return nil
}
