package workflows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"

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
	b := testBackends{
		db:      test_utils.MigrateCockroachDB(t, context.Background()),
		kycImpl: kyc_mock.NewMockClient(ctrl),
	}
	b.users = user_client.New(b, "Testing")

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

	env.ExecuteWorkflow(CreateSendUserWorkflow, wallet.ID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, externalUserID, result)
}
