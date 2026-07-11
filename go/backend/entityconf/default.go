package entityconf

import "context"

// defaultRegistry is the package-level registry that MustRegister and Load
// operate against. Real conf-declaring code registers into this via an
// init() function; tests that want isolation from other packages'
// registrations should construct their own Registry with NewRegistry
// instead.
var defaultRegistry = NewRegistry()

// MustRegister registers example's tagged fields against the package-level
// default registry. See (*Registry).MustRegister.
func MustRegister(entityType EntityType, example any) {
	defaultRegistry.MustRegister(entityType, example)
}

// Load resolves entityID's confs using the package-level default registry.
// See (*Registry).Load.
func Load(ctx context.Context, store Store, entityID string, dest any) error {
	return defaultRegistry.Load(ctx, store, entityID, dest)
}

// Definitions returns every conf registered against the package-level
// default registry, sorted by key.
func Definitions() []Definition {
	return defaultRegistry.Definitions()
}

// DefinitionsFor returns every conf for entityType registered against the
// package-level default registry, sorted by key.
func DefinitionsFor(entityType EntityType) []Definition {
	return defaultRegistry.DefinitionsFor(entityType)
}
