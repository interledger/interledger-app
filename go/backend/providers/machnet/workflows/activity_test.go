package workflows

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"

	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_mock_client "gitlab.com/fynbos/backend/providers/machnet/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	machnet_external_inmem "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	user_client "gitlab.com/fynbos/backend/user/client"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"go.temporal.io/sdk/testsuite"
)

func TestActivity_CreateExternalSendUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		users:   user_mock.NewMock(),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.UpsertExternalSendUser)

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

	val, err := env.ExecuteActivity(a.UpsertExternalSendUser, wallet.ID)
	require.NoError(t, err)

	var res string
	require.NoError(t, val.Get(&res))
	require.NotEmpty(t, res)

	// check that state was formatted correctly for Machnet
	usr, err := b.machnet.External().GetUserByID(context.Background(), res)
	require.NoError(t, err)
	assert.Equal(t, "WC", usr.State)
	assert.NotContains(t, usr.MobilePhone, "+")
}

func TestActivity_CreateUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

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
	assert.Equal(t, u.KYCStatus, machnet.KYCStatusInProgress)
}

func TestActivity_StartExternalKYC(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

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
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := &testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		users:   user_mock.NewMock(),
		machnet: mockMachnet,
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.GetOrCreateReceiveUser)

	toLinkedAccID := uuid.NewString()
	fromLinkedAccID := uuid.NewString()
	fromUserID := uuid.NewString()
	toUserID := uuid.NewString()

	fromWallet, err := b.users.CreateNewWallet(ctx, fromUserID, "TestWallet")
	require.NoError(t, err)

	toWallet, err := b.users.CreateNewWallet(ctx, toUserID, "TestWallet")
	require.NoError(t, err)

	// Neeed to create wallets in the DB because of referenctial integrity and we are using a mock
	_, err = b.db.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", fromWallet.ID, "TestWallet")
	require.NoError(t, err)
	_, err = b.db.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", toWallet.ID, "TestWallet")
	require.NoError(t, err)

	sendUser, err := a.b.External().RegisterUser(ctx, external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	_, err = ops.CreateUser(ctx, a.b, machnet.CreateArgs{
		WalletID:   fromWallet.ID,
		ExternalID: sendUser.ID,
	})
	require.NoError(t, err)

	b.linked.EXPECT().GetByProviderID(gomock.Any(), gomock.Any()).Return(nil, linkedaccounts.ErrNotFound).Times(1)
	b.linked.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

	bankAcc, err := ops.CreateReceiveBankAccount(ctx, a.b, machnet.CreateReceiveBankAccountArgs{
		WalletID:      toWallet.ID,
		AccountNumber: "234",
		AccountType:   machnet.BankAccountTypeCheque,
		BankID:        1,
		BranchID:      1,
	})
	require.NoError(t, err)

	b.linked.EXPECT().Get(gomock.Any(), fromLinkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         fromLinkedAccID,
		WalletID:   fromWallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: uuid.NewString(),
	}, nil)

	b.linked.EXPECT().Get(gomock.Any(), toLinkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         toLinkedAccID,
		WalletID:   toWallet.ID,
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
	}, nil).Times(1)

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

func TestActivity_CreateExternalTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		users:   user_mock.NewMock(),
		machnet: mockMachnet,
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateExternalTransaction)

	linkedAccID := uuid.NewString()
	userID := uuid.NewString()

	wallet, err := b.users.CreateNewWallet(ctx, userID, "TestWallet")
	require.NoError(t, err)

	// Neeed to create wallets in the DB because of referenctial integrity and we are using a mock
	_, err = b.db.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", wallet.ID, "TestWallet")
	require.NoError(t, err)

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
		WalletID:   wallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: uuid.NewString(),
	}, nil)

	b.kycImpl.EXPECT().GetIndividualDetails(gomock.Any(), wallet.ID).Return(&kyc.IndividualDetails{
		WalletID:    wallet.ID,
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
		IPAddress: "10.10.10.10",
	}, nil).Times(1)

	trxIDEnc, err := env.ExecuteActivity(a.CreateExternalTransaction, machnet.CreateTransactionArgs{
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
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateExternalTransaction)
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
		WalletID:   fromWallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: uuid.NewString(),
	}, nil).Times(2)

	b.kycImpl.EXPECT().GetIndividualDetails(gomock.Any(), fromWallet.ID).Return(&kyc.IndividualDetails{
		WalletID:    fromWallet.ID,
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
		IPAddress: "10.10.10.10",
	}, nil).Times(1)

	trxIDEnc, err := env.ExecuteActivity(a.CreateExternalTransaction, machnet.CreateTransactionArgs{
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

func TestActivity_FundUserWalletFromCard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	machnetExt := machnet_external_inmem.New()
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnetExt).AnyTimes()
	b := testBackends{
		db:      db.MigrateTestDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.FundUserWalletFromCard)

	fromLinkedAccID := uuid.NewString()
	fromUserID := uuid.NewString()
	toLinkedAccID := uuid.NewString()

	// Create Signup
	_, err := b.db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), fromUserID)
	require.NoError(t, err)

	fromWallet, err := b.users.CreateNewWallet(ctx, fromUserID, "TestWallet")
	require.NoError(t, err)

	externalMachnetUser, err := a.b.External().RegisterUser(ctx, external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	widget, err := a.b.External().GetFundingAccountWidgetToken(ctx, externalMachnetUser.ID)
	require.NoError(t, err)
	cardExtID := strings.Split(widget.Token, "|")[1]

	mw, err := a.b.External().CreateUserWallet(ctx, externalMachnetUser.ID, "testWallet")
	require.NoError(t, err)

	_, err = ops.CreateUser(ctx, a.b, machnet.CreateArgs{
		WalletID:   fromWallet.ID,
		ExternalID: externalMachnetUser.ID,
	})
	require.NoError(t, err)

	b.linked.EXPECT().Get(gomock.Any(), fromLinkedAccID).Return(&linkedaccounts.LinkedAccount{
		ID:         fromLinkedAccID,
		WalletID:   fromWallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: cardExtID,
		Type:       machnet.TypeSendCard,
	}, nil).AnyTimes()

	linkedAccID := uuid.NewString()
	b.linked.EXPECT().ListByWalletId(gomock.Any(), fromWallet.ID).Return([]linkedaccounts.LinkedAccount{
		{
			ID:         linkedAccID,
			WalletID:   fromWallet.ID,
			Provider:   machnet.ProviderName,
			ProviderID: mw.ID,
			Type:       machnet.TypeWallet,
		},
	}, nil).AnyTimes()

	trxIDEnc, err := env.ExecuteActivity(a.FundUserWalletFromCard, FundWalletArgs{
		CreateTransactionArgs: machnet.CreateTransactionArgs{
			ToLinkedAccountID:   toLinkedAccID,
			FromLinkedAccountID: fromLinkedAccID,
			Amount:              20,
			Currency:            "USD",
			IPAddress:           "197.0.2.8",
		},
		WorkflowID: "",
	})
	require.NoError(t, err)
	var fundResp FundWalletResponse
	err = trxIDEnc.Get(&fundResp)
	require.NoError(t, err)
	require.Equal(t, fundResp.FromWalletLinkedAcc, linkedAccID)

	// Lookup user funds
	wallet, err := machnetExt.GetUserWallet(ctx, externalMachnetUser.ID, mw.ID)
	require.NoError(t, err)

	assert.Equal(t, 20.0, wallet.Balance.Balance)
	assert.Equal(t, 20.0, wallet.Balance.AvailableBalance)

	// check idempotency
	workflowID := uuid.NewString()
	_ = b.db.MustExec("INSERT INTO machnet_transactions_workflow_ref (id, send_user_id, workflow_id, workflow_run_id, activity_name) VALUES ($1, $2, $3, $4, $5)", uuid.NewString(), externalMachnetUser.ID, workflowID, uuid.NewString(), "FundUserWalletFromCard")

	trxIDEnc, err = env.ExecuteActivity(a.FundUserWalletFromCard, FundWalletArgs{
		CreateTransactionArgs: machnet.CreateTransactionArgs{
			ToLinkedAccountID:   toLinkedAccID,
			FromLinkedAccountID: fromLinkedAccID,
			Amount:              20,
			Currency:            "USD",
			IPAddress:           "197.0.2.8",
		},
		WorkflowID: workflowID,
	})
	require.NoError(t, err)
	err = trxIDEnc.Get(&fundResp)
	require.NoError(t, err)
	require.Equal(t, fundResp.FromWalletLinkedAcc, linkedAccID)

	wallet, err = machnetExt.GetUserWallet(ctx, externalMachnetUser.ID, mw.ID)
	require.NoError(t, err)

	assert.Equal(t, 20.0, wallet.Balance.Balance)
	assert.Equal(t, 20.0, wallet.Balance.AvailableBalance)
}

func TestStripEmailPlus(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: "alice@fynbos.test", want: "alice@fynbos.test"},
		{input: "alice+bob@fynbos.test", want: "alice@fynbos.test"},
		{input: "bob1234+bob1234@fynbos.test", want: "bob1234@fynbos.test"},
	}

	for _, tc := range tests {
		got := StripEmailPlus(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestActivity_WithdrawFromWallet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	machnetExt := machnet_external_inmem.New()
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnetExt).AnyTimes()
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

	userID := uuid.NewString()
	// Create Signup
	_, err := b.db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)

	wallet, err := b.users.CreateNewWallet(ctx, userID, "TestWallet")
	require.NoError(t, err)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.WithdrawFromWallet)

	externalSendUser, err := machnetExt.RegisterUser(ctx, external.User{
		ID:   uuid.NewString(),
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)

	sendUser, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   wallet.ID,
		ExternalID: externalSendUser.ID,
	})
	require.NoError(t, err)

	externalWallet, err := machnetExt.CreateUserWallet(context.Background(), sendUser.ID, "default")
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
	b.linked.EXPECT().Get(gomock.Any(), walletLinkedAccountID).Return(&linkedaccounts.LinkedAccount{
		ID:         walletLinkedAccountID,
		WalletID:   wallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: externalWallet.ID,
		Type:       machnet.TypeWallet,
	}, nil).Times(1)
	bankAccountID := uuid.NewString()
	b.linked.EXPECT().Get(gomock.Any(), toLinkedAccountID).Return(&linkedaccounts.LinkedAccount{
		ID:         toLinkedAccountID,
		WalletID:   wallet.ID,
		Provider:   machnet.ProviderName,
		ProviderID: bankAccountID,
		Type:       machnet.TypeReceiveBankAccount,
	}, nil).Times(1)

	_, err = env.ExecuteActivity(a.WithdrawFromWallet, machnet.WithdrawFromWalletArgs{
		Amount:                1000,
		WalletID:              wallet.ID,
		WalletLinkedAccountID: walletLinkedAccountID,
		ToLinkedAccountID:     toLinkedAccountID,
		IpAddress:             "10.10.10.10",
	})
	assert.Contains(t, err.Error(), "Insufficient balance")
}
