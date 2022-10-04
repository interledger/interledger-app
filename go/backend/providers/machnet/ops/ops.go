package ops

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func CreateUser(ctx context.Context, b Backends, args machnet.CreateArgs) (*machnet.User, error) {
	var user machnet.User
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

	return &user, nil
}

func GetUserByWalletID(ctx context.Context, b Backends, walletID string) (*machnet.User, error) {
	var user machnet.User
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

	return &user, nil
}

func GetUserByID(ctx context.Context, b Backends, id string) (*machnet.User, error) {
	var user machnet.User
	err := b.DB().GetContext(
		ctx,
		&user,
		"SELECT id, wallet_id, created_at, updated_at from machnet_users WHERE id = $1;",
		id,
	)
	if err == sql.ErrNoRows {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w, %s", machnet.ErrInternal, err)
	}

	return &user, nil
}

func GetWidgetToken(ctx context.Context, b Backends, walletID string) (*machnet.WidgetToken, error) {
	user, err := GetUserByWalletID(ctx, b, walletID)
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
		UserID:           token.UserID,
	}, nil
}

func HandleEvent(ctx context.Context, b Backends, event external.Event) error {
	// TODO: validate payload
	var err error
	switch event.EventName {
	case external.UserCardAdded:
		err = HandleUserCardAddedEvent(ctx, b, event)
	default:
		log.Warn(
			"Unhandled machnet event",
			zap.String("eventName", event.EventName),
			zap.String("externalUserID", event.UserID),
			zap.String("externalResourceID", event.ResourceID),
		)
	}
	if err != nil {
		return err
	}

	return nil
}

func HandleUserCardAddedEvent(ctx context.Context, b Backends, event external.Event) error {
	user, err := GetUserByID(ctx, b, event.UserID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	// TODO: find out if these details are in the event payload
	card, err := b.External().GetUserFundingsource(ctx, user.ID, event.ResourceID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	_, err = b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:   user.WalletID,
		Name:       card.FundingsourceName,
		Mask:       card.AccountNumber,
		Provider:   machnet.ProviderName,
		ProviderID: card.ID,
		Type:       external.TypeCard,
	})
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}
