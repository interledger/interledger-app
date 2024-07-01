package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/chimoney/ops"
	"gitlab.com/fynbos/backend/transactions"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestChimoneyWithdrawWorkflowIntegration(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	ctrl := gomock.NewController(t)
	trxMock := transactions_mock.NewMockClient(ctrl)
	trxMock.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return("test", nil)
	b := backends{
		transactions: trxMock,
	}
	env.RegisterActivity(ops.NewActivity(b))
	env.OnWorkflow(ops.ExecuteChimoneyWithdrawalWorkflow, mock.Anything, mock.Anything, "test").Return(func(ctx workflow.Context, walletID, trxID string) error {
		assert.Equal(t, "test", trxID)
		trxMock.EXPECT().GetTransaction(gomock.Any(), gomock.Any(), "test").Return(&transactions.Transaction{ID: "test", State: transactions.StateCompleted}, nil)

		return nil
	})

	env.RegisterWorkflow(ops.ExecuteChimoneyWithdrawalWorkflow)

	env.ExecuteWorkflow(ops.CreateChimoneyWithdrawalWorkflow, "testwallet", currency.Amount{
		Value:    100,
		Currency: currency.EUR,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result string
	env.GetWorkflowResult(&result)
	assert.Equal(t, "test", result)

	trx, err := trxMock.GetTransaction(context.Background(), "", "test")
	require.NoError(t, err)
	assert.Equal(t, trx.State, transactions.StateCompleted)
}
