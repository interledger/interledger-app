package ops

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/providers/machnet"
)

type user struct {
	ID        string `db:"id"`
	WalletID  string `db:"wallet_id"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func CreateUser(ctx context.Context, b Backends, args machnet.CreateArgs) (*machnet.User, error) {
	var user user
	err := b.DB().GetContext(
		ctx,
		&user,
		"INSERT INTO machnet_users (id, wallet_id) VALUES ($1, $2) RETURNING id, wallet_id, created_at, updated_at;",
		args.ExternalID,
		args.WalletID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &machnet.User{
		ID: user.ID,
	}, nil
}

func GetUser(ctx context.Context, b Backends, walletID string) (*machnet.User, error) {
	var user user
	err := b.DB().GetContext(
		ctx,
		&user,
		"SELECT id, wallet_id, created_at, updated_at from machnet_users WHERE wallet_id = $1;",
		walletID,
	)
	if err == sql.ErrNoRows {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w, %s", machnet.ErrInternal, err)
	}

	return &machnet.User{
		ID: user.ID,
	}, nil
}

func GetWidgetToken(ctx context.Context, b Backends, walletID string) (*machnet.WidgetToken, error) {
	user, err := GetUser(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	token, err := b.External().GetFundingAccountWidgetToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &machnet.WidgetToken{
		Value:            token.Token,
		ExpiresInMinutes: token.ExpiryMinutes,
	}, nil
}
