package workflows_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
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

func TestCreateCardWorkflow(t *testing.T) {
	t.Setenv("TABAPAY_CLIENT_ID", "test")
	t.Setenv("TABAPAY_BEARER_TOKEN", "test")
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	b := workflows.NewTestBackends()
	a := workflows.NewActivity(b)

	basisTheoryCardID, walletID, fingerprint := uuid.NewString(), uuid.NewString(), uuid.NewString()
	last4 := "1234"
	env.OnActivity(a.QueryCard, mock.Anything, mock.Anything).Return(
		&external.QueryCardResponse{
			SC: 200,
			Card: external.CardResponse{
				Last4: last4,
				Bin:   "5678",
				Push: external.PushObject{
					Network:      "VISA",
					Enabled:      true,
					Type:         external.CardTypeCredit,
					Availability: "Immediate",
					Country:      "US",
				},
				Pull: external.PullObject{
					Enabled: true,
					Network: "Mastercard",
					Country: "US",
					Type:    external.CardTypeDebit,
				},
			},
		},
		nil,
	)
	env.OnActivity(a.CreateBasisTheoryCard, mock.Anything, basistheory.CreateCardArgs{
		WalletID:         walletID,
		TokenID:          basisTheoryCardID,
		Bin:              "5678",
		PushNetwork:      "VISA",
		PushEnabled:      true,
		PushType:         "Credit",
		PushAvailability: "Immediate",
		PushCountry:      "US",
		PullEnabled:      true,
		PullNetwork:      "Mastercard",
		PullType:         "Debit",
		PullCountry:      "US",
	}).Return(
		&basistheory.Card{
			ID:              basisTheoryCardID,
			TokenizedNumber: "1234",
			WalletID:        walletID,
			Fingerprint:     fingerprint,
		}, nil,
	)

	env.OnActivity(a.MarkCardNotDeleted, mock.Anything, basisTheoryCardID).Return(
		func(ctx context.Context, idempotencyKey string) (*linkedaccounts.LinkedAccount, error) {
			return nil, temporal.NewNonRetryableApplicationError(linkedaccounts.ErrNotFound.Error(), "NotFound", linkedaccounts.ErrNotFound)
		},
	)

	providerID := uuid.NewString()
	env.OnActivity(a.CreateExternalCard, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, arg workflows.CreateExternalCardArgs) (*external.CreateAccountResponse, error) {
			return &external.CreateAccountResponse{
				AccountID: providerID,
			}, nil
		},
	)

	env.OnActivity(a.CreateLinkedCard, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, arg workflows.CreateLinkedCardArgs) (*linkedaccounts.LinkedAccount, error) {
			require.Equal(t, basisTheoryCardID, arg.ID)
			require.Equal(t, "1234", arg.Mask)
			require.Equal(t, "VISA 1234", arg.Name)
			require.Equal(t, "VISA 1234", arg.Nickname)
			require.Equal(t, providerID, arg.ProviderID)
			return &linkedaccounts.LinkedAccount{
				ID:         basisTheoryCardID,
				ProviderID: providerID,
				WalletID:   walletID,
				Mask:       "1234",
				Name:       "VISA 1234",
				Nickname:   "VISA 1234",
				CanReceive: true,
			}, nil
		},
	)

	env.ExecuteWorkflow(workflows.CreateTabapayCardWorkflow, tabapay.CreateCardArgs{
		WalletID:           walletID,
		BasisTheoryTokenID: basisTheoryCardID,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result linkedaccounts.LinkedAccount
	require.NoError(t, env.GetWorkflowResult(&result))
}

func TestCreateCardWorkflowExistingLinkedAccount(t *testing.T) {
	t.Setenv("TABAPAY_CLIENT_ID", "test")
	t.Setenv("TABAPAY_BEARER_TOKEN", "test")
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	b := workflows.NewTestBackends()
	a := workflows.NewActivity(b)

	providerID, walletID, basisTheoryCardID, fingerprint := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	last4 := "1234"
	env.OnActivity(a.QueryCard, mock.Anything, mock.Anything).Return(
		&external.QueryCardResponse{
			SC: 200,
			Card: external.CardResponse{
				Last4: last4,
				Bin:   "5678",
				Push: external.PushObject{
					Network:      "VISA",
					Enabled:      true,
					Type:         external.CardTypeCredit,
					Availability: "Immediate",
					Country:      "US",
				},
				Pull: external.PullObject{
					Enabled: true,
					Network: "Mastercard",
					Country: "US",
					Type:    external.CardTypeDebit,
				},
			},
			AVS: external.AVSResponse{
				CodeAVS: external.AVSResponseCodeY,
			},
		},
		nil,
	)
	env.OnActivity(a.CreateBasisTheoryCard, mock.Anything, basistheory.CreateCardArgs{
		WalletID:         walletID,
		TokenID:          basisTheoryCardID,
		Bin:              "5678",
		PushNetwork:      "VISA",
		PushEnabled:      true,
		PushType:         "Credit",
		PushAvailability: "Immediate",
		PushCountry:      "US",
		PullEnabled:      true,
		PullNetwork:      "Mastercard",
		PullType:         "Debit",
		PullCountry:      "US",
	}).Return(
		&basistheory.Card{
			ID:              basisTheoryCardID,
			TokenizedNumber: "VISA 1234",
			WalletID:        walletID,
			Fingerprint:     fingerprint,
		}, nil,
	)

	env.OnActivity(a.MarkCardNotDeleted, mock.Anything, basisTheoryCardID).Return(
		func(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
			return &linkedaccounts.LinkedAccount{
				ID:         basisTheoryCardID,
				WalletID:   walletID,
				Name:       "VISA 1234",
				Nickname:   "VISA 1234",
				Provider:   tabapay.ProviderName,
				ProviderID: providerID,
				Mask:       last4,
				CanReceive: true,
			}, nil
		},
	)

	env.ExecuteWorkflow(workflows.CreateTabapayCardWorkflow, tabapay.CreateCardArgs{
		WalletID:           walletID,
		BasisTheoryTokenID: basisTheoryCardID,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result linkedaccounts.LinkedAccount
	require.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, basisTheoryCardID, result.ID)
	assert.Equal(t, providerID, result.ProviderID)
	assert.Equal(t, tabapay.ProviderName, result.Provider)
}

func TestCreateUnsupportedCard(t *testing.T) {
	t.Setenv("TABAPAY_CLIENT_ID", "test")
	t.Setenv("TABAPAY_BEARER_TOKEN", "test")
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	b := workflows.NewTestBackends()
	a := workflows.NewActivity(b)

	walletID, basisTheoryCardID := uuid.NewString(), uuid.NewString()
	last4 := "1234"
	env.OnActivity(a.QueryCard, mock.Anything, mock.Anything).Return(
		&external.QueryCardResponse{
			SC: 200,
			Card: external.CardResponse{
				Last4: last4,
				Push: external.PushObject{
					Enabled: false,
				},
				Pull: external.PullObject{
					Enabled: false,
				},
			},
			AVS: external.AVSResponse{
				CodeAVS: external.AVSResponseCodeY,
			},
		},
		nil,
	)

	env.ExecuteWorkflow(workflows.CreateTabapayCardWorkflow, tabapay.CreateCardArgs{
		WalletID:           walletID,
		BasisTheoryTokenID: basisTheoryCardID,
	})

	require.True(t, env.IsWorkflowCompleted())
	workflowError := env.GetWorkflowError()
	var applicationError *temporal.ApplicationError
	require.ErrorAs(t, workflowError, &applicationError)
	require.Equal(t, "ErrUnsupportedCard", applicationError.Type())
}
