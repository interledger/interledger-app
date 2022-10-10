package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/db"
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
		"INSERT INTO machnet_users (id, wallet_id, kyc_status) VALUES ($1, $2, $3) RETURNING id, wallet_id, kyc_status, created_at, updated_at;",
		args.ExternalID,
		args.WalletID,
		machnet.KYCStatusUnknown,
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
		"SELECT id, wallet_id, kyc_status, created_at, updated_at from machnet_users WHERE wallet_id = $1;",
		walletID,
	)
	if errors.Is(err, sql.ErrNoRows) {
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
		"SELECT id, wallet_id, kyc_status, created_at, updated_at from machnet_users WHERE id = $1;",
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
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

func CreateReceiveBankAccount(ctx context.Context, b Backends, args machnet.CreateReceiveBankAccountArgs) (*machnet.ReceiveBankAccount, error) {
	insert := db.NewInsert("machnet_receive_bank_accounts").
		Value("wallet_id", args.WalletID).
		Value("account_number", args.AccountNumber).
		Value("bank_id", args.BankID).
		Value("branch_id", args.BranchID).
		Returning("id, wallet_id, account_number, bank_id, branch_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var ra machnet.ReceiveBankAccount
	err = b.DB().GetContext(ctx, &ra, statement, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ra, nil
}

func GetReceiveBankAccount(ctx context.Context, b Backends, id string) (*machnet.ReceiveBankAccount, error) {
	var ra machnet.ReceiveBankAccount
	err := b.DB().GetContext(
		ctx,
		&ra,
		"SELECT id, wallet_id, account_number, bank_id, branch_id, created_at, updated_at FROM machnet_receive_bank_accounts WHERE id=$1;",
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ra, nil
}

func CreateReceiveUser(ctx context.Context, b Backends, args machnet.CreateReceiveUserArgs) (*machnet.ReceiveUser, error) {
	insert := db.NewInsert("machnet_receive_users").
		Value("id", args.ExternalID).
		Value("send_user_id", args.SendUserID).
		Value("receive_wallet_id", args.ReceiveWalletID).
		Returning("id, send_user_id, receive_wallet_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var ru machnet.ReceiveUser
	err = b.DB().GetContext(ctx, &ru, statement, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ru, nil
}

func GetReceiveUser(ctx context.Context, b Backends, args machnet.GetReceiveUserArgs) (*machnet.ReceiveUser, error) {
	var ru machnet.ReceiveUser
	err := b.DB().GetContext(
		ctx,
		&ru,
		"SELECT id, send_user_id, receive_wallet_id, created_at, updated_at FROM machnet_receive_users WHERE receive_wallet_id=$1 and send_user_id=$2;",
		args.ReceiveWalletID,
		args.SendUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ru, nil
}

func CreateReceiveUserBankAccount(ctx context.Context, b Backends, args machnet.CreateReceiveUserBankAccountArgs) (*machnet.ReceiveUserBankAccount, error) {
	insert := db.NewInsert("machnet_receive_user_bank_accounts").
		Value("id", args.ExternalID).
		Value("receive_user_id", args.ReceiveUserID).
		Value("receive_bank_account_id", args.ReceiveBankAccountID).
		Returning("id, receive_user_id, receive_bank_account_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var rua machnet.ReceiveUserBankAccount
	err = b.DB().GetContext(ctx, &rua, statement, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &rua, nil
}

func GetReceiveUserBankAccount(ctx context.Context, b Backends, args machnet.GetReceiveUserBankAccountArgs) (*machnet.ReceiveUserBankAccount, error) {
	var rua machnet.ReceiveUserBankAccount
	err := b.DB().GetContext(
		ctx,
		&rua,
		"SELECT id, receive_user_id, receive_bank_account_id, created_at, updated_at FROM machnet_receive_user_bank_accounts WHERE receive_bank_account_id=$1 AND receive_user_id=$2;",
		args.ReceiveBankAccountID,
		args.ReceiveUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &rua, nil
}

func GetBanks(ctx context.Context, b Backends, countryCode string) ([]machnet.Bank, error) {
	externalList, err := b.External().GetBanks(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	banks := make([]machnet.Bank, len(externalList))
	for i, externalBank := range externalList {
		branches := make([]machnet.Branch, len(externalBank.Branches))
		for j, externalBranch := range externalBank.Branches {
			branches[j] = machnet.Branch{
				ID:   externalBranch.ID,
				Name: externalBranch.Name,
			}
		}
		banks[i] = machnet.Bank{
			ID:                        externalBank.ID,
			Name:                      externalBank.Name,
			Branches:                  branches,
			Country:                   externalBank.Country,
			TransactionSupportedTypes: externalBank.TransactionSupportedTypes,
			ReceivingCurrency:         externalBank.ReceivingCurrency,
		}
	}

	return banks, nil
}

func HandleEvent(ctx context.Context, b Backends, event external.Event) error {
	// TODO: validate payload
	var err error
	switch event.EventName {
	case external.UserCardAdded:
		err = HandleUserCardAddedEvent(ctx, b, event)
	case external.UserKYCInProgress, external.UserKYCSuspended, external.UserKYCRetry, external.UserKYCVerified, external.UserKYCReviewPending:
		err = HandleUserKYCEvent(ctx, b, event)
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

func HandleUserKYCEvent(ctx context.Context, b Backends, event external.Event) error {
	_, err := GetUserByID(ctx, b, event.UserID)
	if err != nil {
		return err
	}

	var newStatus machnet.KYCStatus
	switch event.EventName {
	case external.UserKYCInProgress:
		newStatus = machnet.KYCStatusInProgress
	case external.UserKYCSuspended:
		newStatus = machnet.KYCStatusSuspended
	case external.UserKYCRetry:
		newStatus = machnet.KYCStatusRetry
	case external.UserKYCVerified:
		newStatus = machnet.KYCStatusVerified
	case external.UserKYCReviewPending:
		newStatus = machnet.KYCStatusReviewPending
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE machnet_users SET updated_at=now(), kyc_status=$1 WHERE id=$2", newStatus, event.UserID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
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
