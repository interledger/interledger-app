package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func createClientTx(ctx context.Context, tx *sqlx.Tx, clientURL string) (string, error) {
	id := uuid.NewString()
	_, err := tx.ExecContext(ctx, "INSERT INTO authorisation_clients (id, url) VALUES ($1, $2)", id, clientURL)
	if err != nil {
		return "", fmt.Errorf("%w %s")
	}

	return id, nil
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

		for _, acc := range args.AccessToken.Access {
			_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_grant_access (grant_id, type, actions, identifier, locations, data_types) VALUES ($1, $2, $3, $4, $5, $6)",
				gid, acc.Type, acc.Actions, uuid.NewString(), acc.Locations, acc.Datatypes)
			if err != nil {
				return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
			}
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO authorisation_access_tokens (grant_id, value, expires_in) VALUES ($1, $2, $3)",
			gid, uuid.NewString(), 60*60)
		if err != nil {
			return fmt.Errorf("%w %s", authorisation.ErrInternal, err)
		}

		return nil
	})

}
