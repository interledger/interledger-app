package entityconf

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type overrideKey struct {
	entityType EntityType
	entityID   string
	key        string
}

type memoryStore struct {
	mu        sync.Mutex
	defs      map[string]StoredDefinition
	overrides map[overrideKey]any
}

// NewInMemoryStore returns a Store backed by an in-memory map. It is safe
// for concurrent use.
func NewInMemoryStore() Store {
	return &memoryStore{
		defs:      make(map[string]StoredDefinition),
		overrides: make(map[overrideKey]any),
	}
}

func (s *memoryStore) SyncDefinitions(_ context.Context, defs []Definition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	seen := make(map[string]bool, len(defs))
	for _, d := range defs {
		seen[d.Key] = true

		existing, ok := s.defs[d.Key]
		if !ok {
			s.defs[d.Key] = StoredDefinition{Definition: d, EffectiveDefault: d.CodeDefault}
			continue
		}

		existing.Definition = d // refresh entityType/type/display/desc/codeDefault
		existing.DeprecatedAt = nil
		s.defs[d.Key] = existing
	}

	now := time.Now()
	for key, existing := range s.defs {
		if !seen[key] && existing.DeprecatedAt == nil {
			t := now
			existing.DeprecatedAt = &t
			s.defs[key] = existing
		}
	}

	return nil
}

func (s *memoryStore) StoredDefinitions(_ context.Context) ([]StoredDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]StoredDefinition, 0, len(s.defs))
	for _, d := range s.defs {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

func (s *memoryStore) StoredDefinition(_ context.Context, key string) (StoredDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.defs[key]
	if !ok {
		return StoredDefinition{}, fmt.Errorf("%w: %q", ErrDefinitionNotFound, key)
	}
	return d, nil
}

func (s *memoryStore) SetEffectiveDefault(_ context.Context, key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.defs[key]
	if !ok {
		return fmt.Errorf("%w: %q", ErrDefinitionNotFound, key)
	}
	if !valueMatchesType(value, d.Type) {
		return fmt.Errorf("%w: key %q wants %s, got %T", ErrTypeMismatch, key, d.Type, value)
	}

	d.EffectiveDefault = value
	s.defs[key] = d
	return nil
}

func (s *memoryStore) ResolveAll(_ context.Context, entityType EntityType, entityID string, keys []string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]any, len(keys))
	for _, key := range keys {
		d, ok := s.defs[key]
		if !ok {
			continue
		}
		if v, ok := s.overrides[overrideKey{entityType, entityID, key}]; ok {
			out[key] = v
			continue
		}
		out[key] = d.EffectiveDefault
	}
	return out, nil
}

func (s *memoryStore) SetValue(_ context.Context, entityType EntityType, entityID, key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.defs[key]
	if !ok {
		return fmt.Errorf("%w: %q", ErrDefinitionNotFound, key)
	}
	if d.EntityType != entityType {
		return fmt.Errorf("%w: key %q belongs to %q, not %q", ErrEntityTypeMismatch, key, d.EntityType, entityType)
	}
	if !valueMatchesType(value, d.Type) {
		return fmt.Errorf("%w: key %q wants %s, got %T", ErrTypeMismatch, key, d.Type, value)
	}

	s.overrides[overrideKey{entityType, entityID, key}] = value
	return nil
}

func (s *memoryStore) ClearValue(_ context.Context, entityType EntityType, entityID, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.overrides, overrideKey{entityType, entityID, key})
	return nil
}

func valueMatchesType(value any, t ValueType) bool {
	switch t {
	case TypeBool:
		_, ok := value.(bool)
		return ok
	case TypeInt:
		_, ok := value.(int)
		return ok
	case TypeString:
		_, ok := value.(string)
		return ok
	default:
		return false
	}
}
