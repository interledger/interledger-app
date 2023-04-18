package workflows_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/sdk/testsuite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOutgoingTransactionWorkflow(t *testing.T) {
	t.Setenv("TABAPAY_CLIENT_ID", "test")
	t.Setenv("TABAPAY_BEARER_TOKEN", "test")
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	b := workflows.NewTestBackends()
	a := workflows.NewActivity(b)

	var linkedAccountID string
	providerID, walletID := uuid.NewString(), uuid.NewString()
	tokenizedLast4, cardName := "mber", "test"
	env.OnActivity(a.CreateExternalCard, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, arg workflows.CreateExternalCardArgs) (*external.CreateAccountResponse, error) {
			linkedAccountID = arg.LinkedAccountID
			return &external.CreateAccountResponse{
				AccountID: providerID,
			}, nil
		},
	)

	env.OnActivity(a.CreateLinkedCard, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, arg workflows.CreateLinkedCardArgs) (*linkedaccounts.LinkedAccount, error) {
			require.Equal(t, idempotencyKey, arg.ID)
			require.Equal(t, tokenizedLast4, arg.Mask)
			require.Equal(t, cardName, arg.Name)
			require.Equal(t, cardName, arg.Nickname)
			require.Equal(t, providerID, arg.ProviderID)
			return &linkedaccounts.LinkedAccount{
				ID:         linkedAccountID,
				ProviderID: providerID,
				WalletID:   walletID,
				Mask:       tokenizedLast4,
				Name:       cardName,
				Nickname:   cardName,
			}, nil
		},
	)

	env.ExecuteWorkflow(workflows.CreateTabapayCardWorkflow, tabapay.CreateCardArgs{
		WalletID:       walletID,
		Name:           cardName,
		CardNumber:     "tokenized card number",
		CVV:            "tokenized cvv",
		ExpirationDate: "200601",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result linkedaccounts.LinkedAccount
	require.NoError(t, env.GetWorkflowResult(&result))
}
