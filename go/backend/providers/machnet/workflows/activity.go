package workflows

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	inmemory_external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	b ops.Backends
}

func NewActivity(b Backends) *Activity {
	ob := opsBackends{
		Backends: b,
		external: inmemory_external_client.New(), // TODO: This will probably need to change
	}

	return &Activity{b: ob}
}

func (a *Activity) CreateExternalSendUser(ctx context.Context, walletID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateExternalSendUser_Activity", "walletID", walletID)

	mu, err := ops.GetUserByWalletID(ctx, a.b, walletID)
	if err != nil && !errors.Is(err, machnet.ErrNotFound) {
		return "", err
	}
	if mu != nil {
		return mu.ID, nil
	}

	users, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", err
	}
	if len(users) != 1 {
		return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet(%s) has multiple users (%d) not an individual account", walletID, len(users)), "machnet.ErrInternal", machnet.ErrInternal)
	}

	userData := users[0]

	kycData, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return "", err
	}

	if kycData.DateOfBirth.IsZero() || kycData.Address == nil {
		return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet(%s) does not have required KYC information", walletID), "machnet.ErrIncompleteKYC", machnet.ErrIncompleteKYC)
	}

	var gender string
	switch kycData.Gender {
	case kyc.GenderMale:
		gender = "Male"
	case kyc.GenderFemale:
		gender = "Female"
	default:
		gender = "Other"
	}

	emu, err := a.b.External().RegisterUser(ctx, external.User{
		FirstName:    kycData.FirstName,
		LastName:     kycData.LastName,
		Email:        userData.Email,
		Gender:       gender,
		DateOfBirth:  kycData.DateOfBirth.Format("yyyy-MM-dd"),
		AddressLine1: kycData.Address.Line1,
		AddressLine2: kycData.Address.Line2,
		MobilePhone:  userData.PhoneNumber,
		City:         kycData.Address.City,
		Zipcode:      kycData.Address.ZipCode,
		State:        kycData.Address.State,
		Country:      kycData.Address.CountryCode,
		IPAddress:    kycData.IPAddress,
		Type:         external.TypeSendUser,
	})

	if err != nil {
		return "", err
	}

	return emu.ID, nil
}

func (a *Activity) CreateUser(ctx context.Context, walletID, externalID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateUser_Activity", "walletID", walletID, "externalID", externalID)

	mu, err := ops.GetUserByWalletID(ctx, a.b, walletID)
	if err != nil && !errors.Is(err, machnet.ErrNotFound) {
		return "", err
	}
	if mu != nil {
		return mu.ID, nil
	}

	u, err := ops.CreateUser(ctx, a.b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalID,
	})
	if err != nil {
		return "", err
	}

	return u.ID, nil
}

func (a *Activity) StartExternalKYC(ctx context.Context, externalID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("StartExternalKYC_Activity", "externalID", externalID)

	mu, err := a.b.External().GetUserByID(ctx, externalID)
	if errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user id (%s) not found", externalID), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if mu.Status == external.StatusVerified {
		return nil
	}

	_, err = a.b.External().InitiateKYC(ctx, externalID)

	return err
}

func (a *Activity) CreateTransaction(ctx context.Context, trxID string, trx machnet.CreateTransactionArgs) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateTransaction_Activity", "walletID", trx.FromWalletID)

	mu, err := ops.GetUserByWalletID(ctx, a.b, trx.FromWalletID)
	if errors.Is(err, machnet.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user id (%s) not found", trx.FromWalletID), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	la, err := a.b.LinkedAccounts().Get(ctx, trx.FromLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("linked account id (%s) not found", trx.FromLinkedAccountID), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}
	if la.Provider != machnet.ProviderName {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("linked account id (%s) not a machnet account", trx.FromLinkedAccountID), "ErrInternal", machnet.ErrInternal)
	}
	if la.WalletId != trx.FromWalletID {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("Wallet id (%s) not associated with linked account (%s)", trx.FromWalletID, trx.FromLinkedAccountID), "ErrInternal", machnet.ErrInternal)
	}

	_, err = a.b.External().CreateTransaction(ctx, external.CreateTransactionArgs{
		ID:                trxID,
		FromUserID:        mu.ID,
		FromFundID:        la.ProviderID,
		FundingSourceType: external.FundingSourceTypeCard,
		ToPayoutMethod:    external.PayoutMethodBankDeposit,
		FromAmount:        trx.Amount,
		FromCurrency:      trx.Currency,
		ToCurrency:        trx.Currency,
		ExchangeRate:      1,
		Purpose:           external.PurposePersonalTransfer,
		CalculationMode:   external.CalculationModeSenderAmount,
	})
	if errors.Is(err, external.ErrInvalidArgument) || errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "External", err)
	}

	return err
}

func (a *Activity) DeliverTransaction(ctx context.Context, walletID, transactionID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("DeliverTransaction_Activity", "wallet", walletID, "transactions_id", transactionID)

	mu, err := ops.GetUserByWalletID(ctx, a.b, walletID)
	if errors.Is(err, machnet.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user id (%s) not found", walletID), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	err = a.b.External().UpdateDeliveryRequest(ctx, external.DeliveryRequest{
		Status:        "DELIVERY_REQUESTED",
		TransactionID: transactionID,
		UserID:        mu.ID,
	})
	if errors.Is(err, external.ErrInvalidArgument) || errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "External", err)
	}

	return err
}
