package workflows

import (
	"context"
	"testing"
	"time"

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
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateSendUserWorkflow(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
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

	env.OnActivity(a.CreateExternalSendUser, mock.Anything, wallet.ID).Return(externalUserID, nil)
	env.OnActivity(a.CreateUser, mock.Anything, wallet.ID, externalUserID).Return(externalUserID, nil)
	env.OnActivity(a.StartExternalKYC, mock.Anything, externalUserID).Return(nil)
	env.OnActivity(a.CreateWallet, mock.Anything, externalUserID).Return(nil)

	env.ExecuteWorkflow(CreateSendUserWorkflow, wallet.ID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, externalUserID, result)
}

func TestCreateTransactionWorkflow(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMachnet := machnet_mock_client.NewMockClient(ctrl)
	mockMachnet.EXPECT().External().Return(machnet_external_inmem.New()).AnyTimes()
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
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

	env.OnActivity(a.FundUserWalletFromCard, mock.Anything, mock.Anything).Return(&fundTrx, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Event{
			ID:         uuid.NewString(),
			EventName:  external.TransactionProcessedEvent,
			ResourceID: fundTrx.FundTX,
		})
	}, time.Minute)
	env.OnActivity(a.StartWalletTransfer, mock.Anything, mock.Anything).Return(trxID, nil)
	env.OnActivity(a.CreateTransactionWorkflowRef, mock.Anything, mock.Anything).Return(nil)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ops.TransactionEventsChannel, external.Event{
			ID:         uuid.NewString(),
			EventName:  external.TransactionProcessedEvent,
			ResourceID: trxID,
		})
	}, 2*time.Minute)

	env.ExecuteWorkflow(CreateTransactionWorkflow, machnet.CreateTransactionArgs{
		ToLinkedAccountID:   uuid.NewString(),
		FromLinkedAccountID: uuid.NewString(),
		Amount:              200,
		Currency:            "USD",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, result, trxID)
}
