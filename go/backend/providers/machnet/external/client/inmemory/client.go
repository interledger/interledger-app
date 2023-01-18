package inmemory

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

var _ external.Client = Client{}

func New() *Client {
	return &Client{
		users:               map[string]external.User{},
		userHasReceiveUsers: map[string][]string{},
		fundingsources:      map[string]external.FundingSource{},
		transactions:        map[string]external.Transaction{},
		receiveUserAccounts: map[string]external.ReceiveUserBankAccount{},
		wallets:             map[string]external.Wallet{},
	}
}

type Client struct {
	users               map[string]external.User
	userHasReceiveUsers map[string][]string
	fundingsources      map[string]external.FundingSource
	transactions        map[string]external.Transaction
	receiveUserAccounts map[string]external.ReceiveUserBankAccount
	wallets             map[string]external.Wallet
}

func (c Client) RegisterUser(ctx context.Context, user external.User) (*external.User, error) {
	if user.Type != external.TypeSendUser && user.Type != external.TypeReceiveUser {
		return nil, fmt.Errorf("%w Type must be SEND/RECEIVE", external.ErrInvalidArgument)
	}
	if user.Type == external.TypeReceiveUser && user.SendUserID == "" {
		return nil, fmt.Errorf("%w SendUserID is required for a RECEIVE user.", external.ErrInvalidArgument)
	}

	ret := user
	ret.ID = uuid.NewString()
	ret.Status = external.UserKYCUnverified
	if user.Type == external.TypeReceiveUser {
		sendUser, err := c.GetUserByID(ctx, user.SendUserID)
		if err != nil {
			return nil, fmt.Errorf("%w Send user not found.", external.ErrInternal)
		}

		c.userHasReceiveUsers[sendUser.ID] = append(c.userHasReceiveUsers[sendUser.ID], ret.ID)
	}
	c.users[ret.ID] = ret

	return &ret, nil
}

func (c Client) UpdateUser(ctx context.Context, id string, newValues external.User) (*external.User, error) {
	user, found := c.users[id]
	if !found {
		return nil, external.ErrNotFound
	}
	if newValues.Type != "" {
		return nil, fmt.Errorf("%w Cannot change user type.", external.ErrInvalidArgument)
	}
	if newValues.SendUserID != "" {
		return nil, fmt.Errorf("%w Cannot change SendUserID.", external.ErrInvalidArgument)
	}

	v := reflect.ValueOf(newValues)
	merged := reflect.ValueOf(&user).Elem()
	for i, n := 0, v.NumField(); i < n; i++ {
		val := v.Field(i)
		// Update if field is not empty
		if !reflect.DeepEqual(val.Interface(), reflect.Zero(v.Field(i).Type()).Interface()) {
			merged.Field(i).Set(v.Field(i))
		}
	}

	c.users[id] = user

	return &user, nil
}

func (c Client) GetUserByID(ctx context.Context, id string) (*external.User, error) {
	user, found := c.users[id]
	if !found {
		return nil, external.ErrNotFound
	}

	return &user, nil
}

func (c Client) InitiateKYC(ctx context.Context, userID string) (*external.InitiateKycResponse, error) {
	user, found := c.users[userID]
	if !found {
		return nil, external.ErrNotFound
	}

	user.Status = external.UserKYCVerified
	c.users[userID] = user

	return &external.InitiateKycResponse{
		Success: true,
		Status:  external.UserKYCVerified,
	}, nil
}

func (c Client) GetVerificationStatus(
	ctx context.Context, userID string,
) (*external.VerificationStatus, error) {
	user, found := c.users[userID]
	if !found {
		return nil, external.ErrNotFound
	}
	return &external.VerificationStatus{
		UserID:    userID,
		KycStatus: user.Status,
		CipInfo: external.CipInfo{
			Gender:       user.Status,
			Country:      user.Status,
			FirstName:    user.Status,
			LastName:     user.Status,
			Email:        user.Status,
			City:         user.Status,
			State:        user.Status,
			AddressLine1: user.Status,
		},
	}, nil
}

func (c Client) GetReceiveUserList(ctx context.Context, userID string) ([]external.User, error) {
	user, found := c.users[userID]
	if !found {
		return nil, external.ErrNotFound
	}
	receiveUsers, found := c.userHasReceiveUsers[user.ID]
	if !found {
		return nil, nil
	}

	ret := make([]external.User, len(receiveUsers))
	for i, id := range receiveUsers {
		ret[i] = c.users[id]
	}

	return ret, nil
}

func (c Client) GetFundingAccountWidgetToken(
	ctx context.Context, userID string,
) (*external.WidgetTokenResponse, error) {
	user, found := c.users[userID]
	if !found {
		return nil, external.ErrNotFound
	}

	// automatically add a funding source
	fs := external.FundingSource{
		ID:                 uuid.NewString(),
		UserID:             user.ID,
		FundingsourceName:  "VISA-1234",
		FundingsourceType:  external.TypeCard,
		InstitutionName:    "VISA",
		VerificationStatus: external.StatusVerified,
	}
	c.fundingsources[fs.ID] = fs

	// TODO: send user_card_added event to our webhook

	return &external.WidgetTokenResponse{
		ExpiryMinutes: 15,
		UserID:        user.ID,
		Token:         "machnet-widget-token|" + fs.ID,
	}, nil
}

func (c Client) GetUserFundingsource(
	ctx context.Context, userID, fundingsourceID string,
) (*external.FundingSource, error) {
	user, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	fs, found := c.fundingsources[fundingsourceID]
	if !found {
		return nil, external.ErrNotFound
	}
	if user.ID != fs.UserID {
		return nil, external.ErrNotFound
	}

	return &fs, nil
}

func (c Client) DeleteFundingSource(
	ctx context.Context, userID, fundingSourceID string) error {

	user, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	fs, found := c.fundingsources[fundingSourceID]
	if !found {
		return external.ErrNotFound
	}
	if user.ID != fs.UserID {
		return external.ErrNotFound
	}
	delete(c.fundingsources, fundingSourceID)

	return nil
}

func (c Client) CreateTransaction(
	ctx context.Context, args external.CreateTransactionArgs,
) (*external.Transaction, error) {
	_, err := c.GetUserByID(ctx, args.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("%w User not found.", external.ErrInternal)
	}

	trx := external.Transaction{
		ID:                uuid.NewString(),
		UserID:            args.FromUserID,
		FromAmount:        args.FromAmount,
		FromCurrency:      args.FromCurrency,
		ToCurrency:        args.ToCurrency,
		FeeAmount:         args.FeeAmount,
		ExchangeRate:      args.ExchangeRate,
		FromFundID:        args.FromFundID,
		FundingsourceType: args.FundingSourceType,
		DeliveryStatus:    external.DeliveryStatusNone,
		To:                args.To,
		Status:            external.TransactionProcessed,
	}

	c.transactions[trx.ID] = trx

	// TODO: send transaction status event to our webhook

	return &trx, nil
}

func (c Client) GetUserTransaction(
	ctx context.Context, userID, transactionID string,
) (*external.Transaction, error) {
	_, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w User not found.", external.ErrInternal)
	}

	trx, exists := c.transactions[transactionID]
	if !exists || trx.UserID != userID {
		return nil, fmt.Errorf("%w Transaction not found.", external.ErrInternal)
	}

	return &trx, nil
}

func (c Client) UpdateDeliveryRequest(ctx context.Context, request external.DeliveryRequest) error {
	trx, found := c.transactions[request.TransactionID]
	if !found {
		return fmt.Errorf("%w Transaction not found.", external.ErrNotFound)
	}

	trx.DeliveryStatus = request.Status
	c.transactions[trx.ID] = trx

	return nil
}

func (c Client) CreateUserFundingsource(ctx context.Context, fs external.FundingSource) (*external.FundingSource, error) {
	_, err := c.GetUserByID(ctx, fs.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w User not found.", external.ErrInternal)
	}
	if fs.ID == "" {
		return nil, fmt.Errorf("%w ID is required.", external.ErrInvalidArgument)
	}

	c.fundingsources[fs.ID] = fs
	ret := c.fundingsources[fs.ID]
	return &ret, nil
}

func (c Client) CreateReceiveUserBankAccount(ctx context.Context, sendUserID, receiveUserID string, acc external.ReceiveUserBankAccount) (*external.ReceiveUserBankAccount, error) {
	if acc.PayoutMethod != external.TypeBankDeposit {
		return nil, fmt.Errorf("%w Can only create receive user bank accounts.", external.ErrInvalidArgument)
	}

	_, err := c.GetUserByID(ctx, sendUserID)
	if err != nil {
		return nil, fmt.Errorf("%w Send user not found.", external.ErrNotFound)
	}
	receiveUser, err := c.GetUserByID(ctx, receiveUserID)
	if err != nil {
		return nil, fmt.Errorf("%w Receive user not found.", external.ErrNotFound)
	}

	ra := external.ReceiveUserBankAccount{
		ID:            uuid.NewString(),
		UserID:        receiveUser.ID,
		AccountNumber: acc.AccountNumber,
		AccountType:   acc.AccountType,
		BankID:        acc.BankID,
		BranchID:      acc.BranchID,
		PayoutMethod:  acc.PayoutMethod,
	}
	c.receiveUserAccounts[ra.ID] = ra

	return &ra, nil
}

func (c Client) ListReceiveUserBankAccounts(ctx context.Context, sendUserID, receiveUserID string) ([]external.ReceiveUserBankAccount, error) {
	_, err := c.GetUserByID(ctx, sendUserID)
	if err != nil {
		return nil, fmt.Errorf("%w Send user not found.", external.ErrNotFound)
	}
	_, err = c.GetUserByID(ctx, receiveUserID)
	if err != nil {
		return nil, fmt.Errorf("%w Receive user not found.", external.ErrNotFound)
	}

	var resp []external.ReceiveUserBankAccount
	for _, ra := range c.receiveUserAccounts {
		if ra.UserID != receiveUserID {
			continue
		}

		resp = append(resp, ra)
	}

	return resp, nil
}

func (c Client) GetBanks(ctx context.Context, countryCode string) ([]external.Bank, error) {
	return []external.Bank{
		{
			ID:                        1,
			Name:                      "Test",
			Country:                   countryCode,
			ReceivingCurrency:         []string{countryCode},
			TransactionSupportedTypes: []string{"C2C", "B2B", "B2C"},
			Branches: []external.Branch{
				{
					ID:   1,
					Name: "Local",
				},
			},
		},
	}, nil
}

func (c Client) CreateUserWallet(ctx context.Context, userID, nickName string) (*external.Wallet, error) {
	usr, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w Send user not found.", external.ErrNotFound)
	}

	ret := external.Wallet{
		ID:                 uuid.NewString(),
		UserID:             usr.ID,
		NickName:           nickName,
		FundingSourceType:  external.TypeWallet,
		VerificationStatus: external.StatusVerified,
	}

	c.wallets[ret.ID] = ret

	return &ret, nil
}

func (c Client) GetUserWallet(ctx context.Context, userID, walletID string) (*external.Wallet, error) {
	usr, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w Send user not found.", external.ErrNotFound)
	}

	ret, exists := c.wallets[walletID]
	if !exists {
		return nil, fmt.Errorf("%w Wallet not found.", external.ErrNotFound)
	}

	if ret.UserID != usr.ID {
		return nil, fmt.Errorf("%w Wallet not found.", external.ErrNotFound)
	}

	return &ret, nil
}

func (c Client) FundUserWallet(ctx context.Context, args external.FundWalletArgs) (*external.FundWalletResponse, error) {
	wallet, err := c.GetUserWallet(ctx, args.UserID, args.WalletID)
	if err != nil {
		return nil, err
	}

	fs, err := c.GetUserFundingsource(ctx, args.UserID, args.SourceFundID)
	if err != nil {
		return nil, err
	}

	resp := external.FundWalletResponse{
		ID:           uuid.NewString(),
		UserID:       wallet.UserID,
		SourceFundID: fs.ID,
		Status:       external.TransactionProcessed,
		Amount:       args.Amount,
		Currency:     args.Currency,
		IPAddress:    args.IPAddress,
		Type:         "LOAD",
	}

	wallet.Balance.Balance += args.Amount
	wallet.Balance.AvailableBalance += args.Amount
	c.wallets[wallet.ID] = *wallet

	return &resp, nil
}

func (c Client) CreateWalletTransfer(ctx context.Context, args external.WalletTransferArgs) (*external.WalletTransfer, error) {
	sendWallet, err := c.GetUserWallet(ctx, args.SendUserID, args.SendFundID)
	if err != nil {
		return nil, err
	}

	recvWallet, err := c.GetUserWallet(ctx, args.RecvUserID, args.RecvFundID)
	if err != nil {
		return nil, err
	}

	resp := external.WalletTransfer{
		ID:         uuid.NewString(),
		UserID:     sendWallet.UserID,
		Amount:     args.Amount.Float64(),
		Currency:   args.Amount.Currency.String(),
		FromFundID: recvWallet.ID,
		Status:     external.TransactionProcessed,
		IPAddress:  args.IPAddress,
		To: external.TransactionTo{
			FundID: recvWallet.ID,
			ID:     recvWallet.UserID,
		},
		Type: "TRANSFER",
	}

	recvWallet.Balance.AvailableBalance += args.Amount.Float64()
	recvWallet.Balance.Balance += args.Amount.Float64()
	c.wallets[recvWallet.ID] = *recvWallet

	sendWallet.Balance.Balance -= args.Amount.Float64()
	sendWallet.Balance.AvailableBalance -= args.Amount.Float64()
	c.wallets[sendWallet.ID] = *sendWallet

	return &resp, nil
}

func (c Client) WithdrawFromUserWallet(ctx context.Context, args external.WithdrawFromUserWalletArgs) (*external.WalletWithdrawal, error) {
	usr, err := c.GetUserByID(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w Send user not found.", external.ErrNotFound)
	}

	wallet, exists := c.wallets[args.WalletID]
	if !exists || wallet.UserID != usr.ID {
		return nil, fmt.Errorf("%w %s", external.ErrNotFound, "User wallet not found.")
	}

	toUserID := uuid.NewString()
	toFundingsource, exists := c.fundingsources[args.ToFundID]
	if exists {
		toUserID = toFundingsource.UserID
	}

	wallet.Balance.AvailableBalance -= (args.Amount + args.FeeAmount)
	wallet.Balance.Balance -= (args.Amount + args.FeeAmount)
	if wallet.Balance.AvailableBalance < 0 || wallet.Balance.Balance < 0 {
		return nil, fmt.Errorf("%w Insufficient balance", external.ErrInternal)
	}

	c.wallets[wallet.ID] = wallet

	return &external.WalletWithdrawal{
		ID:           uuid.NewString(),
		UserID:       usr.ID,
		SourceFundID: wallet.ID,
		Status:       "PROCESSED",
		Amount:       args.Amount,
		FeeAmount:    args.FeeAmount,
		Currency:     args.Currency,
		IPAddress:    args.IPAddress,
		Type:         "UNLOAD",
		To: external.WalletWithdrawalTo{
			UserID: toUserID,
			FundID: args.ToFundID,
		},
	}, nil
}
