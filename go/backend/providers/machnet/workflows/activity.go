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
		gender = "male"
	case kyc.GenderFemale:
		gender = "female"
	default:
		gender = "other"
	}

	// we store state in iso3166-2 format which will fail Machnet validation.
	state := kycData.Address.State
	stateParts := strings.Split(state, "-")
	if len(stateParts) > 1 {
		state = strings.TrimSpace(stateParts[1])
	}

	emu, err := a.b.External().RegisterUser(ctx, external.User{
		FirstName:    kycData.FirstName,
		LastName:     kycData.LastName,
		Email:        userData.Email,
		Gender:       gender,
		DateOfBirth:  kycData.DateOfBirth.Format("2006-01-02"),
		AddressLine1: kycData.Address.Line1,
		AddressLine2: kycData.Address.Line2,
		MobilePhone:  strings.Trim(userData.PhoneNumber, "+"),
		City:         kycData.Address.City,
		Zipcode:      kycData.Address.ZipCode,
		State:        state,
		Country:      kycData.Address.CountryCode,
		IPAddress:    kycData.IPAddress,
		Type:         external.TypeSendUser,
	})
	if errors.Is(err, external.ErrInvalidArgument) {
		return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("invalid arguments to create send user wallet(%s) error (%s)", walletID, err), "external.ErrInvalidArgument", external.ErrInvalidArgument)
	}

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

func (a *Activity) CreateWallet(ctx context.Context, externalID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateWallet_Activity", "externalID", externalID)

	mu, err := a.b.External().GetUserByID(ctx, externalID)
	if errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user id (%s) not found", externalID), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	_, err = a.b.External().CreateUserWallet(ctx, mu.ID, "default")

	return err
}

type FundWalletResponse struct {
	FromWalletLinkedAcc string
	FundTX              string
}

type FundWalletArgs struct {
	machnet.CreateTransactionArgs
	WorkflowID string
}

func (a *Activity) FundUserWalletFromCard(ctx context.Context, args FundWalletArgs) (*FundWalletResponse, error) {
	fromCard, err := getLinkedAccount(ctx, a.b, args.FromLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("failed to find linked account (%s)", args.FromLinkedAccountID), "ErrInternal", err)
	}
	if err != nil {
		return nil, err
	}

	if fromCard.Type != machnet.TypeSendCard {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("from linked account (%s) invalid type (%s)", args.FromLinkedAccountID, fromCard.Type), "ErrInvalidArgument", machnet.ErrInvalidArgument)
	}

	accs, err := a.b.LinkedAccounts().ListByWalletId(ctx, fromCard.WalletID)
	if err != nil {
		return nil, err
	}

	var found *linkedaccounts.LinkedAccount
	for _, la := range accs {
		if la.Provider != machnet.ProviderName || la.Type != machnet.TypeWallet {
			continue
		}
		found = &la
	}

	if found == nil {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("could not find linked account for type (%s) on walletID (%s)", machnet.TypeWallet, fromCard.WalletID), "ErrNotFound", machnet.ErrNotFound)
	}

	existingWorkflowRef, err := ops.GetWorkflowRef(ctx, a.b, args.WorkflowID, "FundUserWalletFromCard")
	if err != nil && !errors.Is(err, machnet.ErrNotFound) {
		return nil, err
	}
	if existingWorkflowRef != nil { // this activity has been run sucessfully before from the same or different workflow run
		return &FundWalletResponse{
			FromWalletLinkedAcc: found.ID,
			FundTX:              existingWorkflowRef.ID,
		}, nil
	}

	mu, err := ops.GetUserByWalletID(ctx, a.b, fromCard.WalletID)
	if err != nil {
		return nil, err
	}

	fundResp, err := a.b.External().FundUserWallet(ctx, external.FundWalletArgs{
		UserID:       mu.ID,
		SourceFundID: fromCard.ProviderID,
		WalletID:     found.ProviderID,
		Amount:       args.Amount,
		Currency:     args.Currency,
		IPAddress:    args.IPAddress,
	})
	if errors.Is(err, external.ErrNotFound) || errors.Is(err, external.ErrInvalidArgument) {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("failed to fund machnet user wallet (%s) from card (%s)", found.ProviderID, fromCard.ProviderID), "ErrInternal", err)
	}
	if err != nil {
		return nil, err
	}

	return &FundWalletResponse{
		FromWalletLinkedAcc: found.ID,
		FundTX:              fundResp.ID,
	}, nil
}

type StartWalletTransferArgs struct {
	machnet.CreateTransactionArgs
	WorkflowID string
}

func (a *Activity) StartWalletTransfer(ctx context.Context, args StartWalletTransferArgs, fundTX FundWalletResponse) (string, error) {
	existingWorkflowRef, err := ops.GetWorkflowRef(ctx, a.b, args.WorkflowID, "StartWalletTransfer")
	if err != nil && !errors.Is(err, machnet.ErrNotFound) {
		return "", err
	}
	if existingWorkflowRef != nil { // this activity has been run sucessfully before for the given WorkflowID
		return existingWorkflowRef.ID, nil
	}

	sendLA, err := a.b.LinkedAccounts().Get(ctx, fundTX.FromWalletLinkedAcc)
	if err != nil {
		return "", err
	}

	recvLA, err := a.b.LinkedAccounts().Get(ctx, args.ToLinkedAccountID)
	if err != nil {
		return "", err
	}

	sendUser, err := ops.GetUserByWalletID(ctx, a.b, sendLA.WalletID)
	if err != nil {
		return "", err
	}

	recvUser, err := ops.GetUserByWalletID(ctx, a.b, recvLA.WalletID)
	if err != nil {
		return "", err
	}

	transfer, err := a.b.External().CreateWalletTransfer(ctx, external.WalletTransferArgs{
		SendUserID: sendUser.ID,
		SendFundID: sendLA.ProviderID,
		RecvUserID: recvUser.ID,
		RecvFundID: recvLA.ProviderID,
		Amount:     args.Amount,
		Currency:   args.Currency,
		IPAddress:  args.IPAddress,
	})
	if errors.Is(err, external.ErrNotFound) || errors.Is(err, external.ErrInvalidArgument) {
		return "", temporal.NewNonRetryableApplicationError("failed to initiate machnet wallet transfer", "ErrInternal", err)
	}
	if err != nil {
		return "", err
	}

	return transfer.ID, nil
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

	sendUser, err := ops.GetUserByWalletID(ctx, a.b, fromLinkedAcc.WalletID)
	if errors.Is(err, machnet.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Machnet user not found for `fromLinkedAcc`", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiveUsers, err := a.b.External().GetReceiveUserList(ctx, sendUser.ID)
	if err != nil {
		return nil, err
	}

	fynbosReceives, err := a.b.Users().ListUsers(ctx, toLinkedAcc.WalletID)
	if err != nil {
		return nil, err
	}
	if len(fynbosReceives) != 1 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet id (%s) has multiple accounts linked (%d)", toLinkedAcc.WalletID, len(fynbosReceives)), "ErrInternal", machnet.ErrInternal)
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
		recvUserID, err = addReceiveUser(ctx, a.b, toLinkedAcc.WalletID, sendUser.ID, fynbosReceives[0])
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
		gender = "male"
	case kyc.GenderFemale:
		gender = "female"
	default:
		gender = "other"
	}

	// we store state in iso3166-2 format which will fail Machnet validation.
	state := indvKYC.Address.State
	stateParts := strings.Split(state, "-")
	if len(stateParts) > 1 {
		state = strings.TrimSpace(stateParts[1])
	}
	cc := indvKYC.CountryCode
	if cc == "" {
		cc = indvKYC.Address.CountryCode
	}
	resp, err := b.External().RegisterUser(ctx, external.User{
		FirstName:    indvKYC.FirstName,
		LastName:     indvKYC.LastName,
		Email:        recvUser.Email,
		Gender:       gender,
		DateOfBirth:  indvKYC.DateOfBirth.Format("2006-01-02"),
		AddressLine1: indvKYC.Address.Line1,
		AddressLine2: indvKYC.Address.Line2,
		MobilePhone:  strings.Trim(recvUser.PhoneNumber, "+"),
		City:         indvKYC.Address.City,
		State:        state,
		Zipcode:      indvKYC.Address.ZipCode,
		Country:      cc,
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
	if errors.Is(err, external.ErrInvalidArgument) {
		return "", temporal.NewNonRetryableApplicationError("invalid argument to create machnet receive account", "ErrInvalidArgument", err)
	}
	if err != nil {
		return "", err
	}

	return ba.ID, nil
}

func (a *Activity) CreateExternalTransaction(ctx context.Context, trx machnet.CreateTransactionArgs, to TransactionTo) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateExternalTransaction_Activity", "from", trx.FromLinkedAccountID, "to", trx.ToLinkedAccountID)

	la, err := getLinkedAccount(ctx, a.b, trx.FromLinkedAccountID)
	if err != nil {
		return "", err
	}

	mu, err := ops.GetUserByWalletID(ctx, a.b, la.WalletID)
	if errors.Is(err, machnet.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user not found for linked acc (%s) wallet id (%s)", trx.FromLinkedAccountID, la.WalletID), "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	indvKYC, err := a.b.KYC().GetIndividualDetails(ctx, la.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return "", temporal.NewNonRetryableApplicationError("no KYC information for sender", "KYC", err)
	}
	if err != nil {
		return "", err
	}
	if indvKYC.IPAddress == "" {
		return "", temporal.NewNonRetryableApplicationError("incomplete KYC information for sender. Require IPAddress.", "KYC", kyc.ErrNoKYCInfo)
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
		IPAddress: indvKYC.IPAddress,
	})
	if errors.Is(err, external.ErrInvalidArgument) || errors.Is(err, external.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError(err.Error(), "External", err)
	}
	if err != nil {
		return "", err
	}

	return extTrx.ID, nil
}

type CreateTransactionWorkflowRefArgs struct {
	ExternalTransactionID string
	FromLinkedAccountID   string
	WorkflowID            string
	WorkflowRunID         string
	AcitivityName         string
}

func (a *Activity) CreateTransactionWorkflowRef(ctx context.Context, args CreateTransactionWorkflowRefArgs) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflowRef_Activity", "external transactionID", args.ExternalTransactionID)

	la, err := getLinkedAccount(ctx, a.b, args.FromLinkedAccountID)
	if err != nil {
		return err
	}

	mu, err := ops.GetUserByWalletID(ctx, a.b, la.WalletID)
	if errors.Is(err, machnet.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user not found for linked acc (%s) wallet id (%s)", args.FromLinkedAccountID, la.WalletID), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	_, err = ops.CreateTransactionWorkflowRef(ctx, a.b, machnet.CreateTransactionWorkflowRefArgs{
		ID:            args.ExternalTransactionID,
		WorkflowRunID: args.WorkflowRunID,
		WorkflowID:    args.WorkflowID,
		SendUserID:    mu.ID,
		ActivityName:  args.AcitivityName,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) DeliverTransaction(ctx context.Context, fromLinkedAccID, transactionID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("DeliverTransaction_Activity", "from", fromLinkedAccID, "transactions_id", transactionID)

	la, err := getLinkedAccount(ctx, a.b, fromLinkedAccID)
	if err != nil {
		return err
	}

	mu, err := ops.GetUserByWalletID(ctx, a.b, la.WalletID)
	if errors.Is(err, machnet.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("external user id (%s) not found", la.WalletID), "ErrNotFound", err)
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
