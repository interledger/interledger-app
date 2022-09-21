package ops

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/fundingsources"

	"github.com/google/uuid"
)

func Create(ctx context.Context, b Backends, args *fundingsources.CreateArgs) (*fundingsources.FundingSource, error) {
	// TODO: refactor errors
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInvalidArgument, err.Error())
	}

	// TODO: add back ACL

	fundingsourceID := args.ID
	if fundingsourceID == "" {
		fundingsourceID = uuid.NewString()
	}
	var fs fundingsources.FundingSource
	err = b.DB().GetContext(
		ctx,
		&fs,
		`
			INSERT INTO funding_sources (
				id, wallet_id, name, mask, provider, type
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING *;
		`,
		fundingsourceID,
		args.WalletID,
		args.Name,
		args.Mask,
		args.Provider,
		args.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}

	return &fs, nil
}

func Get(ctx context.Context, b Backends, id string) (*fundingsources.FundingSource, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", fundingsources.ErrInvalidArgument)
	}

	var fundingsource fundingsources.FundingSource
	err := b.DB().GetContext(ctx, &fundingsource, "SELECT * FROM funding_sources where id=$1 LIMIT 1;", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fundingsources.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}

	return &fundingsource, nil
}

func ListByWalletId(ctx context.Context, b Backends, walletId string) ([]fundingsources.FundingSource, error) {

	fundingSources := []fundingsources.FundingSource{}
	err := b.DB().SelectContext(ctx, &fundingSources, "SELECT * FROM funding_sources WHERE wallet_id=$1;", walletId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fundingsources.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}

	return fundingSources, nil
}
