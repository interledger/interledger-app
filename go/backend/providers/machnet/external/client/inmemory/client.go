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
		receiveUserAccounts: map[string]external.ReceiveUserAccount{},
	}
}

type Client struct {
	users               map[string]external.User
	userHasReceiveUsers map[string][]string
	fundingsources      map[string]external.FundingSource
	transactions        map[string]external.Transaction
	receiveUserAccounts map[string]external.ReceiveUserAccount
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
	ret.Status = external.StatusUnverified
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

	user.Status = external.StatusVerified
	c.users[userID] = user

	return &external.InitiateKycResponse{
		Success: true,
		Status:  external.StatusVerified,
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
		Token:         "machnet-widget-token",
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

func (c Client) CreateTransaction(
	ctx context.Context, transaction external.Transaction,
) (*external.Transaction, error) {
	_, err := c.GetUserByID(ctx, transaction.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w User not found.", external.ErrInternal)
	}

	trx := transaction
	trx.ID = uuid.NewString()
	trx.DeliveryStatus = external.DeliveryStatusNone
	c.transactions[trx.ID] = trx

	// TODO: send transaction status event to our webhook

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

func (c Client) CreateReceiveUserAccount(ctx context.Context, sendUserID, receiveUserID string, acc external.ReceiveUserAccount) (*external.ReceiveUserAccount, error) {
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

	ra := external.ReceiveUserAccount{
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
