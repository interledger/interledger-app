// Package entityconf provides a generic, typed configuration mechanism that
// attaches named "confs" to arbitrary entities (wallets today; other entity
// types later) without requiring a schema migration to add or remove a conf.
//
// A conf is declared as a tagged field on a plain Go struct — one struct per
// entity type — and registered once via MustRegister. Reading an entity's
// confs is a single Load call that populates a struct of that same type,
// falling back to an admin-editable effective default for any conf the
// entity hasn't overridden.
//
// See plan.md at the repository root for the full design this package
// implements. This package is intentionally standalone: nothing in the rest
// of the application constructs or calls it yet.
package entityconf

// EntityType identifies the kind of entity a conf is attached to (e.g.
// "wallet"). Entity types are declared as constants alongside the structs
// that use them.
type EntityType string

// ValueType identifies the Go type backing a conf's value. It is inferred
// automatically from a struct field's Go type at registration time — it is
// never set directly by callers.
type ValueType string

const (
	TypeBool   ValueType = "bool"
	TypeInt    ValueType = "int"
	TypeString ValueType = "string"
)
