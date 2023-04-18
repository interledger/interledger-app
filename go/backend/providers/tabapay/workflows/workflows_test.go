package workflows_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"gotest.tools/assert"

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

	idempotencyKey := uuid.NewString()
	env.OnActivity(a.MarkCardNotDeleted, mock.Anything, idempotencyKey).Return(
		func(ctx context.Context, arg workflows.CreateExternalCardArgs) (*external.CreateAccountResponse, error) {
			return nil, temporal.NewNonRetryableApplicationError(linkedaccounts.ErrNotFound.Error(), "NotFound", linkedaccounts.ErrNotFound)
		},
	)

	providerID, walletID := uuid.NewString(), uuid.NewString()
	last4, cardName := "1234", "test"
	env.OnActivity(a.CreateExternalCard, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, arg workflows.CreateExternalCardArgs) (*external.CreateAccountResponse, error) {
			require.Equal(t, idempotencyKey, arg.LinkedAccountID)
			return &external.CreateAccountResponse{
				AccountID: providerID,
				Card: &external.CardResponse{
					Last4: last4,
				},
			}, nil
		},
	)

	env.OnActivity(a.CreateLinkedCard, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, arg workflows.CreateLinkedCardArgs) (*linkedaccounts.LinkedAccount, error) {
			require.Equal(t, idempotencyKey, arg.ID)
			require.Equal(t, last4, arg.Mask)
			require.Equal(t, cardName, arg.Name)
			require.Equal(t, cardName, arg.Nickname)
			require.Equal(t, providerID, arg.ProviderID)
			return &linkedaccounts.LinkedAccount{
				ID:         idempotencyKey,
				ProviderID: providerID,
				WalletID:   walletID,
				Mask:       last4,
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
		IdempotencyKey: idempotencyKey,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result linkedaccounts.LinkedAccount
	require.NoError(t, env.GetWorkflowResult(&result))
}

func TestOutgoingTransactionWorkflowExistingLinkedAccount(t *testing.T) {
	t.Setenv("TABAPAY_CLIENT_ID", "test")
	t.Setenv("TABAPAY_BEARER_TOKEN", "test")
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	b := workflows.NewTestBackends()
	a := workflows.NewActivity(b)

	idempotencyKey, providerID, walletID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	last4, cardName := "1234", "test"
	env.OnActivity(a.MarkCardNotDeleted, mock.Anything, idempotencyKey).Return(
		func(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
			return &linkedaccounts.LinkedAccount{
				ID:         id,
				WalletID:   walletID,
				Name:       cardName,
				Nickname:   cardName,
				Provider:   tabapay.ProviderName,
				ProviderID: providerID,
				Mask:       last4,
			}, nil
		},
	)

	env.ExecuteWorkflow(workflows.CreateTabapayCardWorkflow, tabapay.CreateCardArgs{
		WalletID:       walletID,
		Name:           cardName,
		CardNumber:     "tokenized card number",
		CVV:            "tokenized cvv",
		ExpirationDate: "200601",
		IdempotencyKey: idempotencyKey,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result linkedaccounts.LinkedAccount
	require.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, idempotencyKey, result.ID)
	assert.Equal(t, providerID, result.ProviderID)
	assert.Equal(t, tabapay.ProviderName, result.Provider)
}
