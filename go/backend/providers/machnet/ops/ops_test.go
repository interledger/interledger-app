package ops_test

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestCreateAndGetUser(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)

	args := machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: uuid.NewString(),
	}
	user, err := ops.CreateUser(context.Background(), b, args)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, user.ID)
	require.Equal(t, walletID, user.WalletID)

	freshUser, err := ops.GetUserByWalletID(context.Background(), b, walletID)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, freshUser.ID)
	require.Equal(t, walletID, freshUser.WalletID)

	freshUserByID, err := ops.GetUserByID(context.Background(), b, freshUser.ID)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, freshUserByID.ID)
	require.Equal(t, walletID, freshUserByID.WalletID)

	noUser, err := ops.GetUserByWalletID(context.Background(), b, uuid.NewString())
	assert.Nil(t, noUser)
	assert.ErrorIs(t, err, machnet.ErrNotFound)

	noUserByID, err := ops.GetUserByID(context.Background(), b, uuid.NewString())
	assert.Nil(t, noUserByID)
	assert.ErrorIs(t, err, machnet.ErrNotFound)
}

func TestGetWidgetToken(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	user := NewMachnetUser(t, b, walletID)

	token, err := ops.GetWidgetToken(context.Background(), b, walletID)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(token.Value, "machnet-widget-token|"))
	assert.Equal(t, int(15), token.ExpiresInMinutes)
	assert.Equal(t, user.ID, token.UserID)

	// non-existent user
	token, err = ops.GetWidgetToken(context.Background(), b, uuid.NewString())
	require.Nil(t, token)
	assert.ErrorIs(t, err, machnet.ErrNotFound)
}

func NewWallet(t *testing.T, b backends) string {
	walletID := uuid.NewString()
	_, err := b.DB().Exec(
		"INSERT INTO wallets (id, name) VALUES ($1, $2);",
		walletID,
		"test",
	)
	require.NoError(t, err)

	return walletID
}

func NewMachnetUser(t *testing.T, b backends, walletID string) *machnet.User {
	externalUser, err := b.External().RegisterUser(context.Background(), external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)
	user, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalUser.ID,
	})
	require.NoError(t, err)

	return user
}

func TestCreateReceiveAccount(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	b.linkedaccounts.EXPECT().GetByProviderID(gomock.Any(), gomock.Any()).Return(
		nil,
		linkedaccounts.ErrNotFound,
	).Times(1)
	b.linkedaccounts.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, args *linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
			require.Equal(t, "1234", args.Mask)
			require.Equal(t, machnet.ProviderName, args.Provider)
			require.Equal(t, machnet.TypeReceiveBankAccount, args.Type)
			require.Equal(t, "test", args.Name)
			require.Equal(t, walletID, args.WalletID)

			return &linkedaccounts.LinkedAccount{
				ID: uuid.NewString(),
			}, nil
		}).Times(1)

	ra, err := ops.CreateReceiveBankAccount(context.Background(), b, machnet.CreateReceiveBankAccountArgs{
		WalletID:      walletID,
		AccountNumber: "1234",
		AccountType:   machnet.BankAccountTypeCheque,
		BankID:        1,
		BranchID:      2,
		Name:          "test",
	})
	require.NoError(t, err)
	assert.Equal(t, walletID, ra.WalletID)
	assert.Equal(t, "1234", ra.AccountNumber)
	assert.Equal(t, uint32(1), ra.BankID)
	assert.Equal(t, machnet.BankAccountTypeCheque, ra.AccountType)
	assert.Equal(t, uint32(2), ra.BranchID)

	freshRa, err := ops.GetReceiveBankAccount(context.Background(), b, ra.ID)
	require.NoError(t, err)
	assert.Equal(t, walletID, freshRa.WalletID)
	assert.Equal(t, "1234", freshRa.AccountNumber)
	assert.Equal(t, uint32(1), freshRa.BankID)
	assert.Equal(t, uint32(2), freshRa.BranchID)

	// returns not found error
	noRa, err := ops.GetReceiveBankAccount(context.Background(), b, uuid.NewString())
	assert.Nil(t, noRa)
	assert.ErrorIs(t, err, machnet.ErrNotFound)
}

func TestCreateReceiveAccountIsIdempotent(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	var existingRA machnet.ReceiveBankAccount
	insert := db.NewInsert("machnet_receive_bank_accounts").
		Value("wallet_id", walletID).
		Value("account_number", "1234").
		Value("bank_id", 1).
		Value("branch_id", 1).
		Returning("id, wallet_id, account_number, bank_id, branch_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	require.NoError(t, err)
	err = b.DB().GetContext(context.Background(), &existingRA, statement, values...)
	require.NoError(t, err)
	b.linkedaccounts.EXPECT().GetByProviderID(gomock.Any(), linkedaccounts.GetByProviderIDArgs{
		Provider:   machnet.ProviderName,
		ProviderID: existingRA.ID,
		Type:       machnet.TypeReceiveBankAccount,
		WalletID:   walletID,
	}).Return(
		&linkedaccounts.LinkedAccount{
			ID: uuid.NewString(),
		},
		nil,
	).Times(1)
	b.linkedaccounts.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	ra, err := ops.CreateReceiveBankAccount(context.Background(), b, machnet.CreateReceiveBankAccountArgs{
		WalletID:      walletID,
		AccountNumber: "1234",
		BankID:        1,
		BranchID:      1,
		Name:          "test",
	})
	require.NoError(t, err)
	assert.Equal(t, walletID, ra.WalletID)
	assert.Equal(t, "1234", ra.AccountNumber)
	assert.Equal(t, uint32(1), ra.BankID)
	assert.Equal(t, uint32(1), ra.BranchID)
}

func TestCreateReceiveUser(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	receiveWalletID := NewWallet(t, b)

	sendUser, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: uuid.NewString(),
	})
	require.NoError(t, err)

	externalID := uuid.NewString()
	ru, err := ops.CreateReceiveUser(context.Background(), b, machnet.CreateReceiveUserArgs{
		ExternalID:      externalID,
		SendUserID:      sendUser.ID,
		ReceiveWalletID: receiveWalletID,
	})
	require.NoError(t, err)
	assert.Equal(t, sendUser.ID, ru.SendUserID)
	assert.Equal(t, externalID, ru.ID)
	assert.Equal(t, receiveWalletID, ru.ReceiveWalletID)

	// can only add receive wallet to send user once
	ru, err = ops.CreateReceiveUser(context.Background(), b, machnet.CreateReceiveUserArgs{
		ExternalID:      uuid.NewString(),
		SendUserID:      sendUser.ID,
		ReceiveWalletID: receiveWalletID,
	})
	require.Error(t, err)
	require.Nil(t, ru)

	freshRu, err := ops.GetReceiveUser(context.Background(), b, machnet.GetReceiveUserArgs{
		ReceiveWalletID: receiveWalletID,
		SendUserID:      sendUser.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, sendUser.ID, freshRu.SendUserID)
	assert.Equal(t, externalID, freshRu.ID)
	assert.Equal(t, receiveWalletID, freshRu.ReceiveWalletID)
}

func TestCreateReceiveUserAccount(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	receiveWalletID := NewWallet(t, b)

	var ra machnet.ReceiveBankAccount
	insert := db.NewInsert("machnet_receive_bank_accounts").
		Value("wallet_id", walletID).
		Value("account_number", "1234").
		Value("bank_id", 1).
		Value("branch_id", 1).
		Returning("id, wallet_id, account_number, bank_id, branch_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	require.NoError(t, err)
	err = b.DB().GetContext(context.Background(), &ra, statement, values...)
	require.NoError(t, err)

	sendUser, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: uuid.NewString(),
	})
	require.NoError(t, err)

	ru, err := ops.CreateReceiveUser(context.Background(), b, machnet.CreateReceiveUserArgs{
		ExternalID:      uuid.NewString(),
		SendUserID:      sendUser.ID,
		ReceiveWalletID: receiveWalletID,
	})
	require.NoError(t, err)

	externalID := uuid.NewString()
	rua, err := ops.CreateReceiveUserBankAccount(context.Background(), b, machnet.CreateReceiveUserBankAccountArgs{
		ExternalID:           externalID,
		ReceiveUserID:        ru.ID,
		ReceiveBankAccountID: ra.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, externalID, rua.ID)
	assert.Equal(t, ru.ID, rua.ReceiveUserID)
	assert.Equal(t, ra.ID, rua.ReceiveBankAccountID)

	// can only add a receive account once to a receive user
	rua, err = ops.CreateReceiveUserBankAccount(context.Background(), b, machnet.CreateReceiveUserBankAccountArgs{
		ExternalID:           uuid.NewString(),
		ReceiveUserID:        ru.ID,
		ReceiveBankAccountID: ra.ID,
	})
	require.Error(t, err)
	require.Nil(t, rua)

	freshRua, err := ops.GetReceiveUserBankAccount(context.Background(), b, machnet.GetReceiveUserBankAccountArgs{
		ReceiveUserID:        ru.ID,
		ReceiveBankAccountID: ra.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, externalID, freshRua.ID)
	assert.Equal(t, ru.ID, freshRua.ReceiveUserID)
	assert.Equal(t, ra.ID, freshRua.ReceiveBankAccountID)
}

func TestGetBanks(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	countryCode := "US"

	banks, err := ops.GetBanks(context.Background(), b, countryCode)
	require.NoError(t, err)

	require.Len(t, banks, 1)
	assert.Equal(t, uint32(1), banks[0].ID)
	assert.Equal(t, "Test", banks[0].Name)
	assert.ElementsMatch(t, []string{countryCode}, banks[0].ReceivingCurrency)
	assert.ElementsMatch(t, []string{"C2C", "B2B", "B2C"}, banks[0].TransactionSupportedTypes)
	require.Len(t, banks[0].Branches, 1)
	assert.Equal(t, uint32(1), banks[0].Branches[0].ID)
	assert.Equal(t, "Local", banks[0].Branches[0].Name)
}

func TestCreateAndGetWallet(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)

	externalSendUser, err := b.External().RegisterUser(context.Background(), external.User{
		ID:   uuid.NewString(),
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	sendUser, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalSendUser.ID,
	})
	require.NoError(t, err)

	linkedAccountID := uuid.NewString()
	var externalWalletID string
	b.linkedaccounts.EXPECT().Create(context.Background(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, args *linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
			externalWalletID = args.ProviderID
			return &linkedaccounts.LinkedAccount{
				ID:         linkedAccountID,
				WalletID:   walletID,
				Name:       args.Name,
				Mask:       args.Mask,
				Provider:   args.Provider,
				ProviderID: args.ProviderID,
				Type:       args.Type,
			}, nil
		},
	).Times(1)

	linkedAccount, err := ops.CreateWallet(context.Background(), b, machnet.CreateWalletArgs{
		Nickname:   "fynesse",
		SendUserID: sendUser.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Fynbos Cash", linkedAccount.Mask)
	assert.Equal(t, machnet.ProviderName, linkedAccount.Provider)
	assert.NotEqual(t, "", linkedAccount.ProviderID)
	assert.Equal(t, "fynesse", linkedAccount.Name)
	assert.Equal(t, walletID, linkedAccount.WalletID)

	getWallet, err := ops.GetWallet(context.Background(), b, linkedAccount.ProviderID)
	require.NoError(t, err)
	assert.Equal(t, "fynesse", getWallet.Nickname)
	assert.Equal(t, linkedAccount.ProviderID, getWallet.ID)
	assert.Equal(t, sendUser.ID, getWallet.SendUserID)
	assert.Equal(t, uint64(0), getWallet.AvailableBalance)
	assert.Equal(t, uint64(0), getWallet.Balance)

	// is idempotent
	b.linkedaccounts.EXPECT().GetByProviderID(context.Background(), linkedaccounts.GetByProviderIDArgs{
		Provider:   machnet.ProviderName,
		ProviderID: externalWalletID,
		Type:       machnet.TypeWallet,
		WalletID:   walletID,
	}).Return(
		&linkedaccounts.LinkedAccount{ID: linkedAccountID}, nil,
	).Times(1)
	idempotentLa, err := ops.CreateWallet(context.Background(), b, machnet.CreateWalletArgs{
		Nickname:   "fynesse",
		SendUserID: sendUser.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, linkedAccount.ID, idempotentLa.ID)

	// can't create more than 1 wallet per send user
	_, err = ops.CreateWallet(context.Background(), b, machnet.CreateWalletArgs{
		Nickname:   "test",
		SendUserID: sendUser.ID,
	})
	require.Error(t, err)
}

func TestWithdrawFromWallet(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)

	externalSendUser, err := b.External().RegisterUser(context.Background(), external.User{
		ID:   uuid.NewString(),
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	sendUser, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalSendUser.ID,
	})
	require.NoError(t, err)

	externalWallet, err := b.external.CreateUserWallet(context.Background(), sendUser.ID, "default")
	require.NoError(t, err)

	insert := db.NewInsert("machnet_wallets").
		Value("id", externalWallet.ID).Returning("id").
		Value("nickname", "default").Returning("nickname").
		Value("send_user_id", externalWallet.UserID).Returning("send_user_id").
		Returning("created_at").Returning("updated_at")
	sql, values, err := insert.GetStatement()
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.DB().ExecContext(context.Background(), sql, values...)
	if err != nil {
		t.Fatal(err)
	}

	walletLinkedAccountID, toLinkedAccountID := uuid.NewString(), uuid.NewString()
	b.linkedaccounts.EXPECT().Get(gomock.Any(), walletLinkedAccountID).Return(&linkedaccounts.LinkedAccount{
		ID:         walletLinkedAccountID,
		WalletID:   walletID,
		Provider:   machnet.ProviderName,
		ProviderID: externalWallet.ID,
		Type:       machnet.TypeWallet,
	}, nil).Times(1)
	bankAccountID := uuid.NewString()
	b.linkedaccounts.EXPECT().Get(gomock.Any(), toLinkedAccountID).Return(&linkedaccounts.LinkedAccount{
		ID:         toLinkedAccountID,
		WalletID:   walletID,
		Provider:   machnet.ProviderName,
		ProviderID: bankAccountID,
		Type:       machnet.TypeReceiveBankAccount,
	}, nil).Times(1)

	_, err = ops.WithdrawFromWallet(context.Background(), b, machnet.WithdrawFromWalletArgs{
		Amount:                1000,
		WalletLinkedAccountID: walletLinkedAccountID,
		ToLinkedAccountID:     toLinkedAccountID,
		IpAddress:             "10.10.10.10",
	})
	require.ErrorIs(t, err, machnet.ErrInternal)
	assert.Contains(t, err.Error(), "Insufficient balance")
}

func NewTestBackends(t *testing.T) backends {
	ctrl := gomock.NewController(t)
	return backends{
		db:             db.MigrateTestDB(t, context.Background()),
		external:       external_client.New(),
		linkedaccounts: linkedaccounts_mock.NewMockClient(ctrl),
		temporal:       &mocks.Client{},
	}
}

type backends struct {
	db             *sqlx.DB
	external       *external_client.Client
	linkedaccounts *linkedaccounts_mock.MockClient
	users          user.Client
	kycImpl        kyc.Client
	temporal       *mocks.Client
}

func (b backends) Users() user.Client {
	return b.users
}

func (b backends) KYC() kyc.Client {
	return b.kycImpl
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) External() external.Client {
	return b.external
}

func (b backends) LinkedAccounts() linkedaccounts.Client {
	return b.linkedaccounts
}

func (b backends) Temporal() client.Client {
	return b.temporal
}
