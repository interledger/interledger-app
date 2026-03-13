package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/agreements"

	"github.com/lib/pq"
)

func Get(ctx context.Context, b Backends, id string) (*agreements.Agreement, error) {
	var agreement agreements.Agreement

	err := b.DB().GetContext(ctx, &agreement, "SELECT * FROM agreements WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w %s", agreements.ErrNotFound, err.Error())
		}
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return &agreement, nil
}

func Sign(ctx context.Context, b Backends, args *agreements.SignArgs) error {
	if err := b.Validator().Struct(args); err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInvalidArgument, err)
	}

	tx, err := b.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	txStmt, err := tx.PrepareContext(ctx, "INSERT INTO agreement_signatures (agreement_id, user_id) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}
	defer txStmt.Close()

	for _, id := range args.AgreementIDs {
		_, err := txStmt.ExecContext(ctx, id, args.UserID)
		if err != nil {
			if pgErr, isPGErr := err.(*pq.Error); isPGErr {
				if pgErr.Code != "23503" {
					return fmt.Errorf("%w %s", agreements.ErrNotFound, err.Error())
				}
			}
			return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return nil
}

func GetSignatures(ctx context.Context, b Backends, userID string) ([]agreements.Signature, error) {
	var signatures []agreements.Signature

	err := b.DB().SelectContext(ctx, &signatures, "SELECT id, agreement_id, user_id, created_at, updated_at FROM agreement_signatures WHERE user_id = $1", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, agreements.ErrNotFound
		}
		return nil, fmt.Errorf("%w %s", agreements.ErrInternal, err.Error())
	}

	return signatures, nil
}
