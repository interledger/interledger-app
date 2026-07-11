package entityconf

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// postgresStore is a Store backed by the entity_confs / entity_conf_values
// tables (see go/backend/db/schema.hcl). Construct one with
// NewPostgresStore.
type postgresStore struct {
	db *sqlx.DB
}

// NewPostgresStore returns a Store backed by db. The caller owns db's
// lifecycle (opening and closing it) — this Store never closes it.
func NewPostgresStore(db *sqlx.DB) Store {
	return &postgresStore{db: db}
}

type definitionRow struct {
	Key              string     `db:"key"`
	EntityType       string     `db:"entity_type"`
	Type             string     `db:"type"`
	DisplayName      string     `db:"display_name"`
	Description      string     `db:"description"`
	CodeDefault      []byte     `db:"code_default"`
	EffectiveDefault []byte     `db:"effective_default"`
	DeprecatedAt     *time.Time `db:"deprecated_at"`
}

func (row definitionRow) toStoredDefinition() (StoredDefinition, error) {
	codeDefault, err := decodeValue(ValueType(row.Type), row.CodeDefault)
	if err != nil {
		return StoredDefinition{}, fmt.Errorf("entityconf: decoding code_default for %q: %w", row.Key, err)
	}
	effectiveDefault, err := decodeValue(ValueType(row.Type), row.EffectiveDefault)
	if err != nil {
		return StoredDefinition{}, fmt.Errorf("entityconf: decoding effective_default for %q: %w", row.Key, err)
	}

	return StoredDefinition{
		Definition: Definition{
			Key:         row.Key,
			EntityType:  EntityType(row.EntityType),
			Type:        ValueType(row.Type),
			DisplayName: row.DisplayName,
			Description: row.Description,
			CodeDefault: codeDefault,
		},
		EffectiveDefault: effectiveDefault,
		DeprecatedAt:     row.DeprecatedAt,
	}, nil
}

// decodeValue unmarshals a jsonb column's raw bytes into the Go type that t
// declares (bool/int/string).
func decodeValue(t ValueType, raw []byte) (any, error) {
	switch t {
	case TypeBool:
		var v bool
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	case TypeInt:
		var v int
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	case TypeString:
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFieldKind, t)
	}
}

func (s *postgresStore) SyncDefinitions(ctx context.Context, defs []Definition) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	keys := make([]string, 0, len(defs))
	for _, d := range defs {
		raw, err := json.Marshal(d.CodeDefault)
		if err != nil {
			return fmt.Errorf("entityconf: encoding code_default for %q: %w", d.Key, err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO entity_confs (key, entity_type, type, display_name, description, code_default, effective_default, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $6, now(), now())
			ON CONFLICT (key) DO UPDATE SET
				entity_type   = excluded.entity_type,
				type          = excluded.type,
				display_name  = excluded.display_name,
				description   = excluded.description,
				code_default  = excluded.code_default,
				deprecated_at = NULL,
				updated_at    = now()`,
			d.Key, string(d.EntityType), string(d.Type), d.DisplayName, d.Description, raw)
		if err != nil {
			return fmt.Errorf("entityconf: syncing definition %q: %w", d.Key, err)
		}

		keys = append(keys, d.Key)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE entity_confs SET deprecated_at = now() WHERE deprecated_at IS NULL AND key <> ALL($1)`,
		pq.Array(keys))
	if err != nil {
		return fmt.Errorf("entityconf: marking deprecated definitions: %w", err)
	}

	return tx.Commit()
}

func (s *postgresStore) StoredDefinitions(ctx context.Context) ([]StoredDefinition, error) {
	var rows []definitionRow
	err := s.db.SelectContext(ctx, &rows, `
		SELECT key, entity_type, type, display_name, description, code_default, effective_default, deprecated_at
		FROM entity_confs
		ORDER BY key`)
	if err != nil {
		return nil, err
	}

	out := make([]StoredDefinition, 0, len(rows))
	for _, row := range rows {
		def, err := row.toStoredDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, def)
	}
	return out, nil
}

func (s *postgresStore) StoredDefinition(ctx context.Context, key string) (StoredDefinition, error) {
	var row definitionRow
	err := s.db.GetContext(ctx, &row, `
		SELECT key, entity_type, type, display_name, description, code_default, effective_default, deprecated_at
		FROM entity_confs
		WHERE key=$1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return StoredDefinition{}, fmt.Errorf("%w: %q", ErrDefinitionNotFound, key)
	}
	if err != nil {
		return StoredDefinition{}, err
	}
	return row.toStoredDefinition()
}

func (s *postgresStore) SetEffectiveDefault(ctx context.Context, key string, value any) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	var typ string
	err = tx.GetContext(ctx, &typ, `SELECT type FROM entity_confs WHERE key=$1 FOR UPDATE`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %q", ErrDefinitionNotFound, key)
	}
	if err != nil {
		return err
	}
	if !valueMatchesType(value, ValueType(typ)) {
		return fmt.Errorf("%w: key %q wants %s, got %T", ErrTypeMismatch, key, typ, value)
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("entityconf: encoding effective_default for %q: %w", key, err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE entity_confs SET effective_default=$1, updated_at=now() WHERE key=$2`,
		raw, key); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *postgresStore) ResolveAll(ctx context.Context, entityType EntityType, entityID string, keys []string) (map[string]any, error) {
	type resolvedRow struct {
		Key   string `db:"key"`
		Type  string `db:"type"`
		Value []byte `db:"value"`
	}

	var rows []resolvedRow
	err := s.db.SelectContext(ctx, &rows, `
		SELECT ec.key AS key, ec.type AS type, COALESCE(ecv.value, ec.effective_default) AS value
		FROM entity_confs ec
		LEFT JOIN entity_conf_values ecv
			ON ecv.conf_key = ec.key AND ecv.entity_type = $1 AND ecv.entity_id = $2
		WHERE ec.key = ANY($3)`,
		string(entityType), entityID, pq.Array(keys))
	if err != nil {
		return nil, err
	}

	out := make(map[string]any, len(rows))
	for _, row := range rows {
		v, err := decodeValue(ValueType(row.Type), row.Value)
		if err != nil {
			return nil, fmt.Errorf("entityconf: decoding value for %q: %w", row.Key, err)
		}
		out[row.Key] = v
	}
	return out, nil
}

func (s *postgresStore) SetValue(ctx context.Context, entityType EntityType, entityID, key string, value any) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	var def struct {
		EntityType string `db:"entity_type"`
		Type       string `db:"type"`
	}
	err = tx.GetContext(ctx, &def, `SELECT entity_type, type FROM entity_confs WHERE key=$1 FOR UPDATE`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %q", ErrDefinitionNotFound, key)
	}
	if err != nil {
		return err
	}
	if EntityType(def.EntityType) != entityType {
		return fmt.Errorf("%w: key %q belongs to %q, not %q", ErrEntityTypeMismatch, key, def.EntityType, entityType)
	}
	if !valueMatchesType(value, ValueType(def.Type)) {
		return fmt.Errorf("%w: key %q wants %s, got %T", ErrTypeMismatch, key, def.Type, value)
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("entityconf: encoding value for %q: %w", key, err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO entity_conf_values (entity_type, entity_id, conf_key, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (entity_type, entity_id, conf_key) DO UPDATE SET value=excluded.value, updated_at=now()`,
		string(entityType), entityID, key, raw); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *postgresStore) ClearValue(ctx context.Context, entityType EntityType, entityID, key string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM entity_conf_values WHERE entity_type=$1 AND entity_id=$2 AND conf_key=$3`,
		string(entityType), entityID, key)
	return err
}
