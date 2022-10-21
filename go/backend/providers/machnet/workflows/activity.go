package workflows

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	b ops.Backends
}

func NewActivity(b Backends) *Activity {
	ob := opsBackends{
		Backends: b,
		external: b.Machnet().External(),
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

type TransactionTo struct {
	ReceiveUserID string
	ReceiveFundID string
}

func (a *Activity) GetOrCreateReceiveUser(ctx context.Context, trx machnet.CreateTransactionArgs) (*TransactionTo, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GetOrCreateReceiveUser_Activity", "from", trx.FromLinkedAccountID, "to", trx.ToLinkedAccountID)

	fromLinkedAcc, err := getLinkedAccount(ctx, a.b, trx.FromLinkedAccountID)
	if err != nil {
		return nil, err
	}

	toLinkedAcc, err := getLinkedAccount(ctx, a.b, trx.ToLinkedAccountID)
	if err != nil {
		return nil, err
	}

	sendUser, err := ops.GetUserByWalletID(ctx, a.b, fromLinkedAcc.WalletId)
	if err != nil {
		return nil, err
	}

	receiveUsers, err := a.b.External().GetReceiveUserList(ctx, sendUser.ID)
	if err != nil {
		return nil, err
	}

	fynbosReceives, err := a.b.Users().ListUsers(ctx, toLinkedAcc.WalletId)
	if err != nil {
		return nil, err
	}
	if len(fynbosReceives) != 1 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet id (%s) has multiple accounts linked (%d)", toLinkedAcc.WalletId, len(fynbosReceives)), "ErrInternal", machnet.ErrInternal)
	}

	var recvUserID string
	for _, ru := range receiveUsers {
		for _, fr := range fynbosReceives {
			// User email matches, assume we have added the user before.
			if strings.EqualFold(ru.Email, fr.Email) {
				recvUserID = ru.ID
			}
		}
	}

	// We didn't find the user, register a new one.
	if recvUserID == "" {
		recvUserID, err = addReceiveUser(ctx, a.b, toLinkedAcc.WalletId, sendUser.ID, fynbosReceives[0])
		if err != nil {
			return nil, err
		}
	}

	accs, err := a.b.External().ListReceiveUserBankAccounts(ctx, sendUser.ID, recvUserID)
	if err != nil {
		return nil, err
	}

	ba, err := ops.GetReceiveBankAccount(ctx, a.b, toLinkedAcc.ProviderID)
	if err != nil {
		return nil, err
	}

	var recvAccID string
	for _, acc := range accs {
		if strings.EqualFold(acc.AccountNumber, ba.AccountNumber) {
			recvAccID = acc.ID
		}
	}

	// The account we are trying to send to doesn't exist on the receive user
	if recvAccID == "" {
		recvAccID, err = addReceiveUserBankAccount(ctx, a.b, sendUser.ID, recvUserID, ba)
		if err != nil {
			return nil, err
		}
	}

	return &TransactionTo{
		ReceiveUserID: recvUserID,
		ReceiveFundID: recvAccID,
	}, nil
}

func getLinkedAccount(ctx context.Context, b ops.Backends, id string) (*linkedaccounts.LinkedAccount, error) {
	la, err := b.LinkedAccounts().Get(ctx, id)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("linked account id (%s) not found", id), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}
	if la.Provider != machnet.ProviderName {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("linked account id (%s) not a machnet account", id), "ErrInternal", machnet.ErrInternal)
	}

	return la, nil
}

func addReceiveUser(ctx context.Context, b ops.Backends, recvWalletID, extSendUserID string, recvUser user.User) (string, error) {

	indvKYC, err := b.KYC().GetIndividualDetails(ctx, recvWalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return "", temporal.NewNonRetryableApplicationError("no KYC information for receive", "KYC", err)
	}
	if err != nil {
		return "", err
	}
	if indvKYC.DateOfBirth.IsZero() || indvKYC.Address == nil {
		return "", temporal.NewNonRetryableApplicationError("incomplete KYC information for receive user", "KYC", kyc.ErrNoKYCInfo)
	}

	var gender string
	switch indvKYC.Gender {
	case kyc.GenderMale:
		gender = "Male"
	case kyc.GenderFemale:
		gender = "Female"
	default:
		gender = "Other"
	}

	resp, err := b.External().RegisterUser(ctx, external.User{
		FirstName:    indvKYC.FirstName,
		LastName:     indvKYC.LastName,
		Email:        recvUser.Email,
		Gender:       gender,
		DateOfBirth:  indvKYC.DateOfBirth.Format("yyyy-MM-dd"),
		AddressLine1: indvKYC.Address.Line1,
		AddressLine2: indvKYC.Address.Line2,
		MobilePhone:  recvUser.PhoneNumber,
		City:         indvKYC.Address.City,
		State:        indvKYC.Address.State,
		Zipcode:      indvKYC.Address.ZipCode,
		Country:      indvKYC.CountryCode,
		Type:         external.TypeReceiveUser,
		SendUserID:   extSendUserID,
	})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func addReceiveUserBankAccount(ctx context.Context, b ops.Backends, extSendUserID, extRecvUserID string, bankAccount *machnet.ReceiveBankAccount) (string, error) {

	accType := external.AccountTypeCheque
	if bankAccount.AccountType == machnet.BankAccountTypeSavings {
		accType = external.AccountTypeSavings
	}

	ba, err := b.External().CreateReceiveUserBankAccount(ctx, extSendUserID, extRecvUserID, external.ReceiveUserBankAccount{
		AccountNumber: bankAccount.AccountNumber,
		AccountType:   accType,
		BankID:        int(bankAccount.BankID),
		BranchID:      int(bankAccount.BranchID),
		PayoutMethod:  external.TypeBankDeposit,
	})
	if err != nil {
		return "", err
	}

	return ba.ID, nil
}

func (a *Activity) CreateTransaction(ctx context.Context, trx machnet.CreateTransactionArgs, to TransactionTo) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateTransaction_Activity", "from", trx.FromLinkedAccountID, "to", trx.ToLinkedAccountID)

	la, err := getLinkedAccount(ctx, a.b, trx.FromLinkedAccountID)
	if err != nil {
		return "", err
	}

	mu, err := ops.GetUserByWalletID(ctx, a.b, la.WalletId)
	if errors.Is(err, machnet.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user not found for linked acc (%s) wallet id (%s)", trx.FromLinkedAccountID, la.WalletId), "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	extTrx, err := a.b.External().CreateTransaction(ctx, external.CreateTransactionArgs{
		FromUserID:        mu.ID,
		FromFundID:        la.ProviderID,
		FundingSourceType: external.FundingSourceTypeCard,
		FromAmount:        trx.Amount,
		FromCurrency:      trx.Currency,
		ToCurrency:        trx.Currency,
		ExchangeRate:      1,
		Purpose:           external.PurposePersonalTransfer,
		To: external.TransactionTo{
			CalculationMode: external.CalculationModeSenderAmount,
			ID:              to.ReceiveUserID,
			FundID:          to.ReceiveFundID,
			PayoutMethod:    external.PayoutMethodBankDeposit,
		},
	})
	if errors.Is(err, external.ErrInvalidArgument) || errors.Is(err, external.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError(err.Error(), "External", err)
	}
	if err != nil {
		return "", err
	}

	return extTrx.ID, nil
}

func (a *Activity) DeliverTransaction(ctx context.Context, fromLinkedAccID, transactionID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("DeliverTransaction_Activity", "from", fromLinkedAccID, "transactions_id", transactionID)

	la, err := getLinkedAccount(ctx, a.b, fromLinkedAccID)
	if err != nil {
		return err
	}

	mu, err := ops.GetUserByWalletID(ctx, a.b, la.WalletId)
	if errors.Is(err, machnet.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user id (%s) not found", la.WalletId), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	err = a.b.External().UpdateDeliveryRequest(ctx, external.DeliveryRequest{
		Status:        external.DeliveryStatusRequested,
		TransactionID: transactionID,
		UserID:        mu.ID,
	})
	if errors.Is(err, external.ErrInvalidArgument) || errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "External", err)
	}

	return err
}
