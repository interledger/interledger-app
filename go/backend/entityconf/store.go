package entityconf

import "context"

// Store persists conf definitions and per-entity override values. Two
// implementations exist: NewInMemoryStore (in-memory, no database) and
// NewPostgresStore (backed by the entity_confs/entity_conf_values tables).
type Store interface {
	// SyncDefinitions upserts the given definitions: EntityType, Type,
	// DisplayName, Description, and CodeDefault are always refreshed for a
	// known key. EffectiveDefault is seeded from CodeDefault only the first
	// time a key is seen — it is never overwritten by a later sync. Any
	// previously-synced key absent from defs is marked deprecated; a
	// previously-deprecated key present again in defs is un-deprecated.
	SyncDefinitions(ctx context.Context, defs []Definition) error

	// StoredDefinitions returns every definition known to the store,
	// including deprecated ones, sorted by key.
	StoredDefinitions(ctx context.Context) ([]StoredDefinition, error)

	// StoredDefinition returns a single definition by key, or
	// ErrDefinitionNotFound.
	StoredDefinition(ctx context.Context, key string) (StoredDefinition, error)

	// SetEffectiveDefault overrides a conf's effective default. Returns
	// ErrDefinitionNotFound if key hasn't been synced, or ErrTypeMismatch if
	// value's Go type doesn't match the conf's declared type.
	SetEffectiveDefault(ctx context.Context, key string, value any) error

	// ResolveAll returns the resolved value for each of keys, scoped to
	// (entityType, entityID): the entity's own override if one exists, else
	// the conf's current effective default. A key with no synced
	// definition is simply omitted from the result.
	ResolveAll(ctx context.Context, entityType EntityType, entityID string, keys []string) (map[string]any, error)

	// SetValue sets a per-entity override. Returns ErrDefinitionNotFound if
	// key hasn't been synced, ErrEntityTypeMismatch if key belongs to a
	// different entity type, or ErrTypeMismatch if value's Go type doesn't
	// match the conf's declared type.
	SetValue(ctx context.Context, entityType EntityType, entityID, key string, value any) error

	// ClearValue removes a per-entity override, reverting that entity to
	// the effective default. Clearing a value that has no override is a
	// no-op.
	ClearValue(ctx context.Context, entityType EntityType, entityID, key string) error
}
