package entityconf

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Registry holds the set of confs declared via Register/MustRegister. The
// zero value is not usable; construct one with NewRegistry. Application
// code almost always uses the package-level default registry (see the
// package-level MustRegister and Load functions) rather than constructing
// its own — a dedicated Registry is mainly useful for isolated tests.
type Registry struct {
	mu sync.Mutex

	definitions map[string]Definition       // by key
	types       map[reflect.Type]registered // by the Go struct type passed to Register
}

type registered struct {
	entityType EntityType
	fields     []registeredField
}

type registeredField struct {
	index int
	key   string
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		definitions: make(map[string]Definition),
		types:       make(map[reflect.Type]registered),
	}
}

// Register reflects over example's fields and adds one Definition per
// tagged field to the registry, associating them with entityType. example
// must be a struct value (not a pointer); its field values are ignored —
// defaults come from the "default" struct tag, not from example itself.
//
// Every exported field must carry a `conf:"<key>"` tag, or an explicit
// `conf:"-"` to opt out. This is deliberate: a forgotten tag fails loudly
// here rather than the field silently never becoming a conf. The key must
// be prefixed with "<entityType>." (see plan.md §3's naming convention).
//
// Registration is atomic: if any field is invalid, nothing is registered.
func (r *Registry) Register(entityType EntityType, example any) error {
	t := reflect.TypeOf(example)
	if t == nil || t.Kind() != reflect.Struct {
		return fmt.Errorf("%w: got %v", ErrNotAStruct, t)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.types[t]; exists {
		return fmt.Errorf("%w: %v", ErrTypeAlreadyRegistered, t)
	}

	prefix := string(entityType) + "."
	seenKeys := make(map[string]bool)
	defs := make([]Definition, 0, t.NumField())
	fields := make([]registeredField, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue // unexported, can't be read/set via reflection
		}

		tag, ok := field.Tag.Lookup("conf")
		if !ok {
			return fmt.Errorf("%w: %s.%s", ErrMissingConfTag, t.Name(), field.Name)
		}
		if tag == "-" {
			continue
		}
		if tag == "" {
			return fmt.Errorf("%w: %s.%s", ErrEmptyConfKey, t.Name(), field.Name)
		}
		if !strings.HasPrefix(tag, prefix) || tag == prefix {
			return fmt.Errorf("%w: field %s.%s has key %q, want prefix %q", ErrInvalidKeyPrefix, t.Name(), field.Name, tag, prefix)
		}
		if seenKeys[tag] {
			return fmt.Errorf("%w: %q (duplicate within %s)", ErrDuplicateKey, tag, t.Name())
		}
		if _, exists := r.definitions[tag]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateKey, tag)
		}

		valueType, def, err := fieldDefault(field)
		if err != nil {
			return fmt.Errorf("field %s.%s: %w", t.Name(), field.Name, err)
		}

		seenKeys[tag] = true
		defs = append(defs, Definition{
			Key:         tag,
			EntityType:  entityType,
			Type:        valueType,
			DisplayName: field.Tag.Get("display"),
			Description: field.Tag.Get("desc"),
			CodeDefault: def,
		})
		fields = append(fields, registeredField{index: i, key: tag})
	}

	for _, d := range defs {
		r.definitions[d.Key] = d
	}
	r.types[t] = registered{entityType: entityType, fields: fields}

	return nil
}

// MustRegister calls Register and panics if it returns an error. It is
// meant to be called from an init() function, so a malformed registration
// fails the moment the binary starts, not at some later, harder-to-trace
// point.
func (r *Registry) MustRegister(entityType EntityType, example any) {
	if err := r.Register(entityType, example); err != nil {
		panic(err)
	}
}

// fieldDefault infers the ValueType from field's Go kind and parses its
// "default" tag into a value of that type.
func fieldDefault(field reflect.StructField) (ValueType, any, error) {
	defaultTag := field.Tag.Get("default")

	switch field.Type.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(defaultTag)
		if err != nil {
			return "", nil, fmt.Errorf("%w: %q is not a valid bool: %v", ErrInvalidDefaultTag, defaultTag, err)
		}
		return TypeBool, v, nil
	case reflect.Int:
		v, err := strconv.Atoi(defaultTag)
		if err != nil {
			return "", nil, fmt.Errorf("%w: %q is not a valid int: %v", ErrInvalidDefaultTag, defaultTag, err)
		}
		return TypeInt, v, nil
	case reflect.String:
		return TypeString, defaultTag, nil
	default:
		return "", nil, fmt.Errorf("%w: %v", ErrUnsupportedFieldKind, field.Type.Kind())
	}
}

// Definitions returns every registered Definition, sorted by key.
func (r *Registry) Definitions() []Definition {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Definition, 0, len(r.definitions))
	for _, d := range r.definitions {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// DefinitionsFor returns every registered Definition for the given entity
// type, sorted by key.
func (r *Registry) DefinitionsFor(entityType EntityType) []Definition {
	all := r.Definitions()
	out := make([]Definition, 0, len(all))
	for _, d := range all {
		if d.EntityType == entityType {
			out = append(out, d)
		}
	}
	return out
}

// Load resolves every registered conf for entityID and populates dest,
// which must be a non-nil pointer to a struct type previously passed to
// Register/MustRegister. The entity type is taken from that registration,
// never from the caller. Each field's value is the entity's own override
// if one exists in store, else the conf's current effective default.
func (r *Registry) Load(ctx context.Context, store Store, entityID string, dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%w: got %T", ErrNotAPointer, dest)
	}

	r.mu.Lock()
	reg, ok := r.types[v.Elem().Type()]
	r.mu.Unlock()
	if !ok {
		return fmt.Errorf("%w: %v", ErrTypeNotRegistered, v.Elem().Type())
	}

	keys := make([]string, len(reg.fields))
	for i, f := range reg.fields {
		keys[i] = f.key
	}

	values, err := store.ResolveAll(ctx, reg.entityType, entityID, keys)
	if err != nil {
		return err
	}

	elem := v.Elem()
	for _, f := range reg.fields {
		val, ok := values[f.key]
		if !ok {
			return fmt.Errorf("%w: %q", ErrDefinitionNotFound, f.key)
		}

		fv := elem.Field(f.index)
		rv := reflect.ValueOf(val)
		if rv.Type() != fv.Type() {
			return fmt.Errorf("%w: field %s expects %v, resolved value is %v",
				ErrTypeMismatch, elem.Type().Field(f.index).Name, fv.Type(), rv.Type())
		}
		fv.Set(rv)
	}

	return nil
}
