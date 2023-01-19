package workflows

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_mock_client "gitlab.com/fynbos/backend/providers/machnet/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	machnet_external_inmem "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"

	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"

	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	user_client "gitlab.com/fynbos/backend/user/client"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateSendUserWorkflow(t *testing.T) {
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
	env := testSuite.NewTestWorkflowEnvironment()

	userID := uuid.NewString()
	externalUserID := uuid.NewString()
	// Create Signup
	_, err := b.db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)
	wallet, err := b.users.CreateNewWallet(ctx, userID, "TestWallet")
	require.NoError(t, err)

	a := NewActivity(b)

	env.OnActivity(a.UpsertExternalSendUser, mock.Anything, wallet.ID).Return(externalUserID, nil)
	env.OnActivity(a.CreateUser, mock.Anything, wallet.ID, externalUserID).Return(externalUserID, nil)
	env.OnActivity(a.StartExternalKYC, mock.Anything, externalUserID).Return(nil)
	env.OnActivity(a.CreateUserWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.UserEventsChannel, external.User{
			ID:     externalUserID,
			Status: external.StatusVerified,
		})
	}, 2*time.Minute)
	env.OnActivity(a.CompleteUserWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(a.CreateWallet, mock.Anything, externalUserID).Return(nil)

	env.ExecuteWorkflow(CreateSendUserWorkflow, wallet.ID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, externalUserID, result)
	env.AssertExpectations(t)
}

func TestCreateTransactionWorkflow(t *testing.T) {
	t.Parallel()
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
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(ExecuteWalletTopupWorkflow)

	trxID := uuid.NewString()

	a := NewActivity(b)

	fundTrx := FundWalletResponse{
		FromWalletLinkedAcc: uuid.NewString(),
		FundTX:              uuid.NewString(),
	}

	txWallets := TransactionWalletIDs{
		FromWalletID: uuid.NewString(),
		ToWalletID:   uuid.NewString(),
	}

	env.OnActivity(a.ShouldFundWallet, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.GetTransactionsWallets, mock.Anything, mock.Anything).Return(&txWallets, nil)

	env.OnWorkflow(ExecuteWalletTopupWorkflow, mock.Anything, mock.Anything).Return(&machnet.CreateTransactionResponse{
		TransactionState: transactions.StateCompleted,
		ExternalID:       fundTrx.FundTX,
	}, nil)

	env.OnActivity(a.StartWalletTransfer, mock.Anything, mock.Anything).Return(trxID, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(a.AddTransactionTransfer, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(a.AddTransaction, mock.Anything, mock.Anything).Return(trxID, nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     trxID,
			Status: external.TransactionProcessed,
		})
	}, time.Minute*2)
	env.OnActivity(a.UpdateTransferStateByType, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	env.ExecuteWorkflow(CreateTransactionWorkflow, machnet.CreateTransactionArgs{
		FromForeignID:       uuid.NewString(),
		ToForeignID:         uuid.NewString(),
		FromPaymentPointer:  uuid.NewString(),
		ToPaymentPointer:    uuid.NewString(),
		FromLinkedAccountID: uuid.NewString(),
		ToLinkedAccountID:   uuid.NewString(),
		Amount:              currency.FromFloat64(200, currency.ParseCurrency("USD")),
	}, trxID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result machnet.CreateTransactionResponse
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result.ExternalID, trxID)
	env.AssertExpectations(t)
}

func TestDeleteAccountWorkflow(t *testing.T) {
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
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivity(b)

	env.OnActivity(a.DeleteUserFundSource, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(a.DeleteLinkedAccount, mock.Anything, mock.Anything).Return(nil)

	env.ExecuteWorkflow(DeleteAccountWorkflow, uuid.NewString())

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.NoError(t, env.GetWorkflowResult(nil))
	env.AssertExpectations(t)
}

func TestExecuteWalletTopupWorkflow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      nil,
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivity(b)

	trxID := uuid.NewString()
	walletLinkedAccountID := uuid.NewString()
	fromLinkedAccountID := uuid.NewString()
	fundTrx := FundWalletResponse{
		FromWalletLinkedAcc: walletLinkedAccountID,
		FundTX:              trxID,
	}
	trx := transactions.Transaction{
		ID:          uuid.NewString(),
		ForeignID:   uuid.NewString(),
		Source:      "",
		Destination: "",
		Note:        "",
		Type:        transactions.TransactionTypeMachnetWalletTopUp,
		Timestamp:   time.Time{},
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      currency.Amount{},
		Transfers:   nil,
	}

	env.OnActivity(a.ShouldFundWallet, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.GetTransaction, mock.Anything, mock.Anything, mock.Anything).Return(trx, nil)
	env.OnActivity(a.FundUserWalletFromCard, mock.Anything, mock.Anything).Return(&fundTrx, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     fundTrx.FundTX,
			Status: external.TransactionProcessed,
		})
	}, time.Minute)
	env.OnActivity(a.UpdateTransactionState, mock.Anything, trx.ID, transactions.StateCompleted).Return(nil)

	env.ExecuteWorkflow(ExecuteWalletTopupWorkflow, ExecuteTopupArgs{
		WalletID:            uuid.NewString(),
		UpdateTransaction:   true,
		FromLinkedAccountID: fromLinkedAccountID,
		Amount:              currency.FromFloat64(200, currency.ParseCurrency("USD")),
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result machnet.CreateTransactionResponse
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result.ExternalID, trxID)
	env.AssertExpectations(t)
}

func TestExecuteWalletTopupSendsFailedEmail(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      nil,
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivity(b)

	trxID := uuid.NewString()
	walletID := uuid.NewString()
	walletLinkedAccountID := uuid.NewString()
	fromLinkedAccountID := uuid.NewString()
	fundTrx := FundWalletResponse{
		FromWalletLinkedAcc: walletLinkedAccountID,
		FundTX:              trxID,
	}
	trx := transactions.Transaction{
		ID:          uuid.NewString(),
		ForeignID:   uuid.NewString(),
		Source:      "",
		Destination: "",
		Note:        "",
		Type:        transactions.TransactionTypeMachnetWalletTopUp,
		Timestamp:   time.Time{},
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      currency.Amount{},
		Transfers:   nil,
	}

	env.OnActivity(a.ShouldFundWallet, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.GetTransaction, mock.Anything, mock.Anything, mock.Anything).Return(trx, nil)
	env.OnActivity(a.FundUserWalletFromCard, mock.Anything, mock.Anything).Return(&fundTrx, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     fundTrx.FundTX,
			Status: external.TransactionFailed,
		})
	}, time.Minute)
	env.OnActivity(a.UpdateTransactionState, mock.Anything, trx.ID, transactions.StateFailed).Return(nil)
	env.OnActivity(a.SendFailedTransactionMail, mock.Anything, walletID, transactions.TransactionTypeMachnetWalletTopUp).Return(nil)

	env.ExecuteWorkflow(ExecuteWalletTopupWorkflow, ExecuteTopupArgs{
		WalletID:            walletID,
		FromLinkedAccountID: fromLinkedAccountID,
		Amount:              currency.FromFloat64(200, currency.ParseCurrency("USD")),
		UpdateTransaction:   true,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result machnet.CreateTransactionResponse
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result.ExternalID, trxID)
	env.AssertExpectations(t)
}

func TestExecuteWalletWithdrawalSendsFailedEmail(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      nil,
		kycImpl: kyc_mock.NewMockClient(ctrl),
		linked:  linkedaccounts_mock.NewMockClient(ctrl),
		machnet: mockMachnet,
	}
	b.users = user_client.New(b, "kratosURL", "kratosAdminURL")

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivity(b)

	trxID := uuid.NewString()
	walletID := uuid.NewString()
	walletLinkedAccountID := uuid.NewString()
	toLinkedAccountID := uuid.NewString()
	withdrawTrx := machnet.WalletWithdrawal{
		ID:     uuid.NewString(),
		Status: external.TransactionFailed,
	}

	trx := transactions.Transaction{
		ID:          uuid.NewString(),
		ForeignID:   uuid.NewString(),
		Source:      "",
		Destination: "",
		Note:        "",
		Type:        transactions.TransactionTypeMachnetWalletTopUp,
		Timestamp:   time.Time{},
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      currency.Amount{},
		Transfers:   nil,
	}

	env.OnActivity(a.GetTransaction, mock.Anything, mock.Anything, mock.Anything).Return(trx, nil)
	env.OnActivity(a.WithdrawFromWallet, mock.Anything, mock.Anything, mock.Anything).Return(&withdrawTrx, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     withdrawTrx.ID,
			Status: external.TransactionFailed,
		})
	}, time.Minute)
	env.OnActivity(a.UpdateTransactionState, mock.Anything, trx.ID, transactions.StateFailed).Return(nil)
	env.OnActivity(a.SendFailedTransactionMail, mock.Anything, walletID, transactions.TransactionTypeMachnetWalletWithdrawal).Return(nil)

	env.ExecuteWorkflow(ExecuteWalletWithdrawalWorkflow, trxID, machnet.WithdrawFromWalletArgs{
		WalletID:              walletID,
		WalletLinkedAccountID: walletLinkedAccountID,
		ToLinkedAccountID:     toLinkedAccountID,
		IpAddress:             "0.0.0.0",
		Amount:                1000,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result, withdrawTrx.ID)
	env.AssertExpectations(t)
}
