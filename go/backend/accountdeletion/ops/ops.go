package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/accountdeletion"
)

func Request(ctx context.Context, b Backends, userID string) error {
	// ON CONFLICT DO NOTHING surfaces duplicates as 0 rows affected, not a SQLSTATE error.
	r, err := b.DB().ExecContext(ctx,
		"INSERT INTO account_deletion_requests (user_id, status) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING",
		userID, accountdeletion.StatusPending,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", accountdeletion.ErrInternal, err)
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", accountdeletion.ErrInternal, err)
	}
	if affected == 0 {
		return accountdeletion.ErrAlreadyRequested
	}
	return nil
}

func Delete(ctx context.Context, b Backends, userID string) error {
	_, err := b.DB().ExecContext(ctx,
		"DELETE FROM account_deletion_requests WHERE user_id = $1",
		userID)
	if err != nil {
		return fmt.Errorf("%w: %w", accountdeletion.ErrInternal, err)
	}
	return nil
}

func GetForUser(ctx context.Context, b Backends, userID string) (*accountdeletion.Request, error) {
	var req accountdeletion.Request
	err := b.DB().GetContext(ctx, &req,
		"SELECT id, user_id, status, created_at, updated_at FROM account_deletion_requests WHERE user_id = $1 LIMIT 1;",
		userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", accountdeletion.ErrInternal, err)
	}
	return &req, nil
}
