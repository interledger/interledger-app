package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_mock "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay/ops"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	temporal "go.temporal.io/sdk/client"
)

type testWorkflowRun struct{}

func (t testWorkflowRun) GetID() string {
	return ""
}

func (t testWorkflowRun) GetRunID() string {
	return ""
}

func (t testWorkflowRun) Get(ctx context.Context, result interface{}) error {
	return nil
}

func (t testWorkflowRun) GetWithOptions(ctx context.Context, result interface{}, options temporal.WorkflowRunGetOptions) error {
	return nil
}

func TestCreateCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.ExternalClient = external_mock.NewMockClient(ctrl)
		tb.TemporalClient = temporal_mock.NewMockClient(ctrl)
	})

	args := tabapay.CreateCardArgs{
		IdempotencyKey: uuid.NewString(),
		WalletID:       uuid.NewString(),
		Name:           "test",
		CardNumber:     "123456789",
		CVV:            "1234",
		ExpirationDate: "202301",
	}

	b.ExternalClient.EXPECT().QueryCard(gomock.Any(), external.QueryCardArgs{
		Card: &external.Card{
			AccountNumber:  args.CardNumber,
			ExpirationDate: args.ExpirationDate,
			SecurityCode:   args.CVV,
		},
	}).Return(
		&external.QueryCardResponse{
			Card: external.CardResponse{
				Push: external.PushObject{
					Type: external.CardTypeDebit,
				},
				Pull: external.PullObject{
					Type: external.CardTypeDebit,
				},
			},
		},
		nil,
	).AnyTimes()

	b.TemporalClient.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), args).Return(
		testWorkflowRun{},
		nil,
	).AnyTimes()

	_, err := ops.CreateCard(context.Background(), b, args)
	require.NoError(t, err)
}
