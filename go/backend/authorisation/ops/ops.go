package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/authorisation"
)

func LookupClient(ctx context.Context, b Backends, clientURL string) (*authorisation.Client, error) {
	var client authorisation.Client
	err := b.DB().GetContext(ctx, &client, "SELECT id, url FROM authorisation_clients WHERE url=$1", clientURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", authorisation.ErrNotFound, err)
	} else if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	return &client, nil
}

func CreateGrant(ctx context.Context, b Backends, args authorisation.GrantRequest) (*authorisation.Grant, error) {
	err := b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInvalidArgument, err)
	}

	cl, err := LookupClient(ctx, b, args.Client.Display.URI)
	if err != nil {
		return nil, err
	}

	gid := uuid.NewString()
	label := args.AccessToken.Label
	if label == "" {
		label = fmt.Sprintf("auto_label_%d", rand.Int31n(1000))
	}

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

		_, err := tx.ExecContext(ctx, "INSERT INTO authorisation_grants (id, client_id, state, continue_token, wait) VALUES ($1, $2, $3, $4, $5)",
			gid, cl.ID, authorisation.GrantStateApproved, uuid.NewString(), 0)
		if err != nil {
			return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		for _, acc := range args.AccessToken.Access {
			_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_grant_access (grant_id, type, actions, identifier, locations, data_types) VALUES ($1, $2, $3, $4, $5, $6)",
				gid, acc.Type, acc.Actions, uuid.NewString(), acc.Locations, acc.Datatypes)
			if err != nil {
				return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
			}
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_access_tokens (grant_id, value, label, expires_in) VALUES ($1, $2, $3)",
			gid, uuid.NewString(), label, 60*60)
		if err != nil {
			return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return lookupGrant(ctx, b, gid)
}

type dbGrant struct {
	ID            string                   `db:"id"`
	State         authorisation.GrantState `db:"state"`
	ClientID      string                   `db:"client_id"`
	ContinueToken string                   `db:"continue_token"`
	Wait          string                   `db:"wait"`
}

type dbAccessToken struct {
	ID        string    `db:"id"`
	Value     string    `db:"value"`
	Label     string    `db:"label"`
	ExpiresIn int       `db:"expires_in"`
	CreatedAt time.Time `db:"created_at"`
}

type dbAccess struct {
	ID         string   `db:"id"`
	Type       string   `db:"type"`
	Actions    []string `db:"actions"`
	Identifier string   `db:"identifier"`
	Locations  []string `db:"locations"`
	DataTypes  []string `db:"data_types"`
}

func lookupGrant(ctx context.Context, b Backends, id string) (*authorisation.Grant, error) {
	var dbg dbGrant
	err := b.DB().GetContext(ctx, &dbg, "SELECT id, client_id, state, continue_token, wait FROM authorisation_grants WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authorisation.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	var access []dbAccess
	err = b.DB().SelectContext(ctx, &access, "SELECT id, type, actions, identifier, locations, data_types FROM authorisation_grant_access WHERE grant_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	var tokens []dbAccessToken
	err = b.DB().SelectContext(ctx, tokens, "SELECT id, value, label, expires_in, created_at FROM authorisation_access_tokens WHERE grant_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	resp := &authorisation.Grant{
		ID:            dbg.ID,
		State:         dbg.State,
		ContinueToken: dbg.ContinueToken,
		Wait:          dbg.Wait,
	}

	respAccess := make([]authorisation.Access, len(access))
	for i, ac := range access {
		respAccess[i] = authorisation.Access{
			Type:      ac.Type,
			Actions:   ac.Actions,
			Locations: ac.Locations,
			Datatypes: ac.DataTypes,
		}
	}

	respTokens := make([]authorisation.AccessToken, len(tokens))
	for i, tk := range tokens {
		respTokens[i] = authorisation.AccessToken{
			Value:     tk.Value,
			Access:    respAccess,
			ExpiresIn: int(math.Max(float64(tk.CreatedAt.Unix()+int64(tk.ExpiresIn)-time.Now().Unix()), 0)),
		}
	}

	return resp, nil
}
