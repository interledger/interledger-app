package entityconf

import "errors"

var (
	// Registration errors (Register / MustRegister).
	ErrNotAStruct            = errors.New("entityconf: example must be a struct value, not a pointer or other kind")
	ErrTypeAlreadyRegistered = errors.New("entityconf: this Go type is already registered")
	ErrMissingConfTag        = errors.New(`entityconf: exported field is missing a "conf" tag (use conf:"-" to skip it)`)
	ErrEmptyConfKey          = errors.New("entityconf: conf tag key must not be empty")
	ErrInvalidKeyPrefix      = errors.New(`entityconf: conf key must be prefixed with "<entityType>."`)
	ErrDuplicateKey          = errors.New("entityconf: conf key is already registered")
	ErrUnsupportedFieldKind  = errors.New("entityconf: unsupported field kind (only bool, int, and string are supported)")
	ErrInvalidDefaultTag     = errors.New(`entityconf: "default" tag could not be parsed for the field's type`)

	// Load errors.
	ErrNotAPointer        = errors.New("entityconf: destination must be a non-nil pointer to a registered struct")
	ErrTypeNotRegistered  = errors.New("entityconf: destination type was never registered")
	ErrDefinitionNotFound = errors.New("entityconf: conf definition not found (has SyncDefinitions run for this store?)")
	ErrTypeMismatch       = errors.New("entityconf: value's Go type does not match the conf's declared type")

	// Store errors.
	ErrEntityTypeMismatch = errors.New("entityconf: key belongs to a different entity type")
)
