package entityconf

import "time"

// Definition describes a single conf as declared in code: its key, the
// entity type it belongs to, its value type, its admin-portal labels, and
// its codebase-declared default.
type Definition struct {
	Key         string
	EntityType  EntityType
	Type        ValueType
	DisplayName string
	Description string
	CodeDefault any
}

// StoredDefinition is a Definition as known to a Store: it additionally
// tracks the admin-editable effective default, and — once a conf is removed
// from the code registry — when it was marked deprecated.
type StoredDefinition struct {
	Definition
	EffectiveDefault any
	DeprecatedAt     *time.Time
}
