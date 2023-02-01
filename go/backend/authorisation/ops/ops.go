package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/lib/pq"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/authorisation"
)

func CreateClient(ctx context.Context, b Backends, clientURL string) error {
	_, err := b.DB().ExecContext(ctx, "INSERT INTO authorisation_clients (url) VALUES ($1)", clientURL)
	if err != nil {
		return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	return nil
}

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

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

		_, err := tx.ExecContext(ctx, "INSERT INTO authorisation_grants (id, client_id, state, continue_token, wait) VALUES ($1, $2, $3, $4, $5)",
			gid, cl.ID, authorisation.GrantStateApproved, uuid.NewString(), 0)
		if err != nil {
			return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		for _, tkn := range args.AccessToken {
			tokenID := uuid.NewString()
			label := tkn.Label
			if label == "" {
				label = fmt.Sprintf("auto_label_%d", rand.Int31n(10000))
			}
			_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_tokens (id, grant_id, state, token, label, expires_in) VALUES ($1, $2, $3, $4, $5, $6)",
				tokenID, gid, authorisation.TokenStateEnabled, uuid.NewString(), label, 60*60)
			if err != nil {
				return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
			}

			for _, acc := range tkn.Access {
				_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_token_access (token_id, type, actions, identifier, locations, data_types) VALUES ($1, $2, $3, $4, $5, $6)",
					tokenID, acc.Type, pq.Array(acc.Actions), uuid.NewString(), pq.Array(acc.Locations), pq.Array(acc.Datatypes))
				if err != nil {
					return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
				}
			}
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
	Value     string    `db:"token"`
	Label     string    `db:"label"`
	ExpiresIn int       `db:"expires_in"`
	CreatedAt time.Time `db:"created_at"`
}

type dbAccess struct {
	ID         string         `db:"id"`
	Type       string         `db:"type"`
	Actions    pq.StringArray `db:"actions"`
	Identifier string         `db:"identifier"`
	Locations  pq.StringArray `db:"locations"`
	DataTypes  pq.StringArray `db:"data_types"`
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

	resp := &authorisation.Grant{
		ID:            dbg.ID,
		State:         dbg.State,
		ContinueToken: dbg.ContinueToken,
		Wait:          dbg.Wait,
	}

	var tokens []dbAccessToken
	err = b.DB().SelectContext(ctx, &tokens, "SELECT id, token, label, expires_in, created_at FROM authorisation_tokens WHERE grant_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
	}

	respTokens := make([]authorisation.AccessToken, len(tokens))
	for i, tk := range tokens {
		var access []dbAccess
		err = b.DB().SelectContext(ctx, &access, "SELECT id, type, actions, identifier, locations, data_types FROM authorisation_token_access WHERE token_id=$1", tk.ID)
		if err != nil {
			return nil, fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		respAccess := make([]authorisation.Access, len(access))
		for y, ac := range access {
			respAccess[y] = authorisation.Access{
				Type:      ac.Type,
				Actions:   ac.Actions,
				Locations: ac.Locations,
				Datatypes: ac.DataTypes,
			}
		}

		respTokens[i] = authorisation.AccessToken{
			Value:     tk.Value,
			Access:    respAccess,
			ExpiresIn: int(math.Max(float64(tk.CreatedAt.Unix()+int64(tk.ExpiresIn)-time.Now().Unix()), 0)),
		}
	}

	resp.Tokens = respTokens

	return resp, nil
}
