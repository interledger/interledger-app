package workflows

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"

	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	user_client "gitlab.com/fynbos/backend/user/client"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/testsuite"
)

func TestActivity_CreateExternalSendUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		users:   user_mock.NewMock(),
		kycImpl: kyc_mock.NewMockClient(ctrl),
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateExternalSendUser)

	userID := uuid.NewString()
	wallet, err := b.users.CreateNewWallet(ctx, userID, "TestWallet")
	require.NoError(t, err)

	b.kycImpl.EXPECT().GetIndividualDetails(gomock.Any(), wallet.ID).Return(&kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   "FirstName",
		LastName:    "LastName",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
		DateOfBirth: time.Date(2001, time.April, 5, 0, 0, 0, 0, time.UTC),
		Address: &kyc.Address{
			Line1:       "Line1",
			Line2:       "Line2",
			Building:    "Building",
			Apartment:   "2",
			City:        "Cape Town",
			State:       "ZA-WC",
			ZipCode:     "8001",
			CountryCode: "ZA",
		},
		IPAddress: "192.8.6.12",
	}, nil).Times(1)

	val, err := env.ExecuteActivity(a.CreateExternalSendUser, wallet.ID)
	require.NoError(t, err)

	var res string
	require.NoError(t, val.Get(&res))
	require.NotEmpty(t, res)
}

func TestActivity_CreateUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
	}
	b.users = user_client.New(b, "Testing")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateUser)

	userID := uuid.NewString()
	// Create Signup
	_, err := b.db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)

	externalUserID := uuid.NewString()
	wallet, err := b.users.CreateNewWallet(ctx, userID, "TestWallet")
	require.NoError(t, err)

	_, err = env.ExecuteActivity(a.CreateUser, wallet.ID, externalUserID)
	require.NoError(t, err)

	u, err := ops.GetUserByWalletID(ctx, a.b, wallet.ID)
	require.NoError(t, err)

	assert.Equal(t, u.ID, externalUserID)
	assert.Equal(t, u.KYCStatus, machnet.KYCStatusUnknown)
}

func TestActivity_StartExternalKYC(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
	}
	b.users = user_client.New(b, "Testing")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.StartExternalKYC)

	mu, err := a.b.External().RegisterUser(ctx, external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)
	_, err = env.ExecuteActivity(a.StartExternalKYC, mu.ID)
	require.NoError(t, err)

	mu, err = a.b.External().GetUserByID(ctx, mu.ID)
	require.NoError(t, err)
	assert.Equal(t, mu.Status, external.StatusVerified)
}

func TestActivity_GetOrCreateReceiveUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := &testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		users:   user_mock.NewMock(),
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.GetOrCreateReceiveUser)

	toLinkedAccID := uuid.NewString()
	fromLinkedAccID := uuid.NewString()
	fromUserID := uuid.NewString()
	toUserID := uuid.NewString()

	// Create Signups
	/*_, err := b.db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2), ($3, $4)", uuid.NewString(), fromUserID, uuid.NewString(), toUserID)
	require.NoError(t, err)
	*/
	fromWallet, err := b.users.CreateNewWallet(ctx, fromUserID, "TestWallet")
	require.NoError(t, err)

	toWallet, err := b.users.CreateNewWallet(ctx, toUserID, "TestWallet")
	require.NoError(t, err)

	// Neeed to create wallets in the DB because of referenctial integrity and we are using a mock
	b.db.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", fromWallet.ID, "TestWallet")
	b.db.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", toWallet.ID, "TestWallet")

	sendUser, err := a.b.External().RegisterUser(ctx, external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	_, err = ops.CreateUser(ctx, a.b, machnet.CreateArgs{
		WalletID:   fromWallet.ID,
		ExternalID: sendUser.ID,
	})
	require.NoError(t, err)

	bankAcc, err := ops.CreateReceiveBankAccount(ctx, a.b, machnet.CreateReceiveBankAccountArgs{
		WalletID:      toWallet.ID,
		AccountNumber: "234",
		BankID:        1,
		BranchID:      1,
	})
	require.NoError(t, err)

	b.linked.EXPECT().Get(gomock.Any(), fromLinkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         fromLinkedAccID,
		WalletId:   fromWallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: uuid.NewString(),
	}, nil)

	b.linked.EXPECT().Get(gomock.Any(), toLinkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         toLinkedAccID,
		WalletId:   toWallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: bankAcc.ID,
	}, nil)

	b.kycImpl.EXPECT().GetIndividualDetails(gomock.Any(), toWallet.ID).Return(&kyc.IndividualDetails{
		WalletID:    toWallet.ID,
		FirstName:   "FirstName",
		LastName:    "LastName",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
		DateOfBirth: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		Address: &kyc.Address{
			Line1:       "Leidenhalm 1",
			Line2:       "",
			Building:    "Bridge House",
			Apartment:   "",
			City:        "Newlands",
			State:       "ZA-WC",
			ZipCode:     "7901",
			CountryCode: "ZA",
		},
	}, nil)

	toEnc, err := env.ExecuteActivity(a.GetOrCreateReceiveUser, machnet.CreateTransactionArgs{
		ToLinkedAccountID:   toLinkedAccID,
		FromLinkedAccountID: fromLinkedAccID,
		Amount:              200,
		Currency:            "USD",
	})
	require.NoError(t, err)

	var to TransactionTo
	err = toEnc.Get(&to)
	require.NoError(t, err)
	require.NotNil(t, to)

	rul, err := a.b.External().GetReceiveUserList(ctx, sendUser.ID)
	require.NoError(t, err)
	require.Len(t, rul, 1)
	assert.Equal(t, rul[0].ID, to.ReceiveUserID)

	rbl, err := a.b.External().ListReceiveUserBankAccounts(ctx, sendUser.ID, to.ReceiveUserID)
	require.NoError(t, err)
	require.Len(t, rbl, 1)
	assert.Equal(t, rbl[0].ID, to.ReceiveFundID)
}

func TestActivity_CreateTransaction(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		users:   user_mock.NewMock(),
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateTransaction)

	linkedAccID := uuid.NewString()
	userID := uuid.NewString()

	wallet, err := b.users.CreateNewWallet(ctx, userID, "TestWallet")
	require.NoError(t, err)

	// Neeed to create wallets in the DB because of referenctial integrity and we are using a mock
	b.db.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", wallet.ID, "TestWallet")

	mu, err := a.b.External().RegisterUser(ctx, external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	_, err = ops.CreateUser(ctx, a.b, machnet.CreateArgs{
		WalletID:   wallet.ID,
		ExternalID: mu.ID,
	})
	require.NoError(t, err)

	b.linked.EXPECT().Get(gomock.Any(), linkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         linkedAccID,
		WalletId:   wallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: uuid.NewString(),
	}, nil)

	trxIDEnc, err := env.ExecuteActivity(a.CreateTransaction, machnet.CreateTransactionArgs{
		FromLinkedAccountID: linkedAccID,
		Amount:              200,
		Currency:            "USD",
	}, TransactionTo{
		ReceiveUserID: uuid.NewString(),
		ReceiveFundID: uuid.NewString(),
	})
	require.NoError(t, err)
	var trxID string
	err = trxIDEnc.Get(&trxID)
	require.NoError(t, err)
	require.NotEmpty(t, trxID)
}

func TestActivity_CreateUserFundingsource(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
	}
	b.users = user_client.New(b, "Testing")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateTransaction)
	env.RegisterActivity(a.DeliverTransaction)

	fromLinkedAccID := uuid.NewString()
	fromUserID := uuid.NewString()
	toLinkedAccID := uuid.NewString()

	// Create Signup
	_, err := b.db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), fromUserID)
	require.NoError(t, err)

	fromWallet, err := b.users.CreateNewWallet(ctx, fromUserID, "TestWallet")
	require.NoError(t, err)

	mu, err := a.b.External().RegisterUser(ctx, external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	_, err = ops.CreateUser(ctx, a.b, machnet.CreateArgs{
		WalletID:   fromWallet.ID,
		ExternalID: mu.ID,
	})
	require.NoError(t, err)

	b.linked.EXPECT().Get(gomock.Any(), fromLinkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         fromLinkedAccID,
		WalletId:   fromWallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: uuid.NewString(),
	}, nil).Times(2)

	trxIDEnc, err := env.ExecuteActivity(a.CreateTransaction, machnet.CreateTransactionArgs{
		ToLinkedAccountID:   toLinkedAccID,
		FromLinkedAccountID: fromLinkedAccID,
		Amount:              200,
		Currency:            "USD",
	}, TransactionTo{
		ReceiveUserID: uuid.NewString(),
		ReceiveFundID: uuid.NewString(),
	})
	require.NoError(t, err)
	var trxID string
	err = trxIDEnc.Get(&trxID)
	require.NoError(t, err)
	require.NotEmpty(t, trxID)

	_, err = env.ExecuteActivity(a.DeliverTransaction, fromLinkedAccID, trxID)
	require.NoError(t, err)
}
