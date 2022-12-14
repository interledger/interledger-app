package workflows

import (
	"context"
	"testing"
	"time"

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

	trxID := uuid.NewString()

	a := NewActivity(b)

	fundTrx := FundWalletResponse{
		FromWalletLinkedAcc: uuid.NewString(),
		FundTX:              uuid.NewString(),
	}

	env.OnActivity(a.ShouldFundWallet, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.FundUserWalletFromCard, mock.Anything, mock.Anything).Return(&fundTrx, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     fundTrx.FundTX,
			Status: external.TransactionProcessed,
		})
	}, time.Minute)
	env.OnActivity(a.StartWalletTransfer, mock.Anything, mock.Anything).Return(trxID, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     trxID,
			Status: external.TransactionProcessed,
		})
	}, time.Minute*2)

	env.ExecuteWorkflow(CreateTransactionWorkflow, machnet.CreateTransactionArgs{
		FromForeignID:       uuid.NewString(),
		ToForeignID:         uuid.NewString(),
		FromLinkedAccountID: uuid.NewString(),
		ToLinkedAccountID:   uuid.NewString(),
		Amount:              200,
		Currency:            "USD",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result, trxID)
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
}

func TestCreateWalletTopupWorkflow(t *testing.T) {
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

	env.OnActivity(a.ShouldFundWallet, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.FundUserWalletFromCard, mock.Anything, mock.Anything).Return(&fundTrx, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Transaction{
			ID:     fundTrx.FundTX,
			Status: external.TransactionProcessed,
		})
	}, time.Minute)

	env.ExecuteWorkflow(CreateWalletTopupWorkflow, machnet.StartWalletTopupArgs{
		WalletLinkedAccountID: walletLinkedAccountID,
		FromLinkedAccountID:   fromLinkedAccountID,
		Amount:                200,
		Currency:              "USD",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result, trxID)
}
