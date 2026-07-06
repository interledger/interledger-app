package tigerroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
)

func ConfigureLedgers(
	ctx context.Context,
	b Backends,
	args []pacioli.ConfigureLedgerArgs,
) ([]pacioli.LedgerResult, error) {

	var results []pacioli.LedgerResult
	for i, lc := range args {
		err := b.Validator().Struct(lc)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err, pacioli.ErrInvalidArg)
		}
	}

	for i, lc := range args {
		code, err := configureLedger(ctx, b, lc)
		if err != nil {
			return results, fmt.Errorf("%s %d %s %w", "index: ", i, err, pacioli.ErrInternal)
		}

		if code == pacioli.LedgerOK {
			continue
		}

		results = append(results, pacioli.LedgerResult{
			Index: uint32(i),
			Code:  code,
		})
	}

	return results, nil
}

func configureLedger(
	ctx context.Context,
	b Backends,
	args pacioli.ConfigureLedgerArgs,
) (pacioli.LedgerResultCode, error) {

	ex, err := GetLedger(ctx, b, args.ID)

	if err != nil && !errors.Is(err, pacioli.ErrNotFound) {
		return 0, err
	}
	if ex != nil {
		if args.Name != ex.Name {
			return pacioli.LedgerExistsWithDifferentName, nil
		} else if args.Asset != ex.Asset {
			return pacioli.LedgerExistsWithDifferentAsset, nil
		} else if args.Scale != ex.Scale {
			return pacioli.LedgerExistsWithDifferentScale, nil
		}

		// The Ledger in the DB matches the args, do nothing,
		return pacioli.LedgerOK, nil
	}

	_, err = b.DB().ExecContext(
		ctx,
		`INSERT INTO ledgers (id, name, asset, scale) VALUES ($1, $2, $3, $4);`,
		args.ID,
		args.Name,
		args.Asset,
		args.Scale,
	)
	if err != nil {
		return pacioli.LedgerOK, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	return pacioli.LedgerOK, nil
}

func GetLedger(ctx context.Context, b Backends, id uint32) (*pacioli.Ledger, error) {
	var ledger pacioli.Ledger
	err := b.DB().GetContext(ctx, &ledger, "SELECT * FROM ledgers WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pacioli.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	return &ledger, nil
}

func ListLedgers(ctx context.Context, b Backends, ids []uint32) ([]pacioli.Ledger, error) {
	var ledgers []pacioli.Ledger
	query, args, err := sqlx.In("SELECT * FROM ledgers WHERE id IN (?);", ids)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}
	err = b.DB().SelectContext(ctx, &ledgers, b.DB().Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	return ledgers, nil
}
