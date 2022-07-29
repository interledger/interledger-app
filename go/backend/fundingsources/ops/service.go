package ops

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/fundingsources"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, b Backends, args *fundingsources.CreateArgs) (*fundingsources.FundingSource, error) {
	// TODO: refactor errors
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInvalidArgument, err.Error())
	}

	identity, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}
	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}
	if !b.Accounts().CanCreateFundingSource(acc, identity.ID) {
		return nil, fundingsources.ErrUnauthorized
	}

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
				id, account_id, name, mask, verification_state, type, subtype
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING *;
		`,
		fundingsourceID,
		acc.ID,
		args.Name,
		args.Mask,
		args.VerificationState,
		args.Type,
		args.SubType,
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

func GetByAccountId(ctx context.Context, b Backends, identityId string) ([]fundingsources.FundingSource, error) {
	if identityId == "" {
		return nil, fmt.Errorf("%w IdentityID is required.", fundingsources.ErrInvalidArgument)
	}

	fundingSources := []fundingsources.FundingSource{}
	err := b.DB().SelectContext(ctx, &fundingSources, "SELECT * FROM funding_sources WHERE account_id=$1;", identityId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fundingsources.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}

	return fundingSources, nil
}

func Verify(ctx context.Context, b Backends, args *fundingsources.VerifyArgs) (*fundingsources.FundingSource, error) {
	// TODO: refactor errors
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInvalidArgument, err.Error())
	}

	id, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}
	fs, err := Get(ctx, b, args.FundingSourceID)
	if err != nil {
		return nil, err
	}
	acc, err := b.Accounts().Get(ctx, fs.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
	}
	if !b.Accounts().CanVerifyFundingSource(acc, id.ID) {
		return nil, fundingsources.ErrUnauthorized
	}

	var verifiedFs fundingsources.FundingSource
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed(`
			UPDATE funding_sources SET verification_state=$1 where id=$2 RETURNING *;
		`)
		if err != nil {
			return fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
		}

		err = stmt.Stmt.Get(&verifiedFs,
			"verified",
			args.FundingSourceID,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return fundingsources.ErrNotFound
			}

			return fmt.Errorf("%w %s", fundingsources.ErrInternal, err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &verifiedFs, nil
}

func CreateBankAccount(
	ctx context.Context,
	b Backends,
	args *fundingsources.CreateBankAccountArgs,
) (*fundingsources.FundingSource, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", fundingsources.ErrInvalidArgument, err.Error())
	}

	fundingsource, err := Create(ctx, b, &fundingsources.CreateArgs{
		IdentityID:        args.IdentityID,
		AccountID:         args.AccountID,
		Name:              args.Name,
		Mask:              args.AccountNumber[:4],
		VerificationState: "required",
		Type:              "noop",
		SubType:           args.Type,
	})
	if err != nil {
		return nil, err
	}

	return fundingsource, nil
}

func IsVerified(fs *fundingsources.FundingSource) bool {
	return fs.VerificationState == "verified"
}
