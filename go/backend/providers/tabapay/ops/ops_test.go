package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
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

func Test3DS(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.Db = db.MigrateTestDB(t, ctx)
		tb.ExternalClient = external_mock.NewMockClient(ctrl)
		tb.TemporalClient = temporal_mock.NewMockClient(ctrl)
	})

	threeDSID := uuid.NewString()
	b.ExternalClient.EXPECT().Init3DS(gomock.Any(), gomock.Any()).Return(&external.Init3DSResponse{
		SC:                  200,
		ID3DS:               threeDSID,
		JWT:                 "jwt",
		DeviceCollectionURL: "http://localhost/device",
	}, nil).AnyTimes()
	b.ExternalClient.EXPECT().Lookup3DS(gomock.Any(), gomock.Any()).Return(&external.Lookup3DSResponse{
		SC:                     200,
		Version3DS:             "1",
		Enrolled:               "Y",
		ProcessorTransactionID: "ptrxid1",
		DsTransactionID:        "dstrxid1",
		Status:                 "status1",
		ChallengeURL:           "http://localhost/challenge",
		Payload:                "payload",
	}, nil).AnyTimes()
	b.ExternalClient.EXPECT().Authenticate3DS(gomock.Any(), gomock.Any()).Return(&external.Authenticate3DSResponse{
		SC:                     200,
		Version3DS:             "2",
		Enrolled:               "Y",
		ProcessorTransactionID: "ptrxid2",
		DsTransactionID:        "dstrxid2",
		Status:                 "status2",
		ECI:                    "eci",
		UCAF:                   "ucaf",
		XID:                    "xid",
	}, nil).AnyTimes()

	idempotencyKey, cardID := uuid.NewString(), uuid.NewString()
	init, err := ops.Init3DS(ctx, b, tabapay.Init3DSArgs{
		Amount: currency.Amount{
			Value:    500,
			Currency: currency.ParseCurrency("USD"),
		},
		IdempotencyKey: idempotencyKey,
		CardID:         cardID,
	})
	require.NoError(t, err)

	assert.Equal(t, threeDSID, init.ID)
	assert.Equal(t, "jwt", init.JWT)
	assert.Equal(t, "http://localhost/device", init.DeviceCollectionURL)

	session, err := ops.Get3DSSession(ctx, b, threeDSID)
	require.NoError(t, err)
	assert.Equal(t, threeDSID, session.ID)
	assert.NotEmpty(t, session.InitAt)
	assert.Empty(t, session.LookupAt)
	assert.Empty(t, session.AuthenticatedAt)
	assert.Equal(t, 1, session.Revision)

	lookup, err := ops.Lookup3DS(ctx, b, tabapay.Lookup3DSArgs{
		IdempotencyKey: idempotencyKey,
		CardID:         cardID,
		ThreeDSID:      threeDSID,
		Amount: currency.Amount{
			Value:    500,
			Currency: currency.ParseCurrency("USD"),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "ptrxid1", lookup.ProcessorTransactionID)
	assert.Equal(t, "dstrxid1", lookup.DsTransactionID)
	assert.Equal(t, "Y", lookup.Enrolled)
	assert.Equal(t, "status1", lookup.Status)
	assert.Equal(t, "http://localhost/challenge", lookup.ChallengeURL)
	assert.Equal(t, "payload", lookup.Payload)

	session, err = ops.Get3DSSession(ctx, b, threeDSID)
	require.NoError(t, err)
	assert.Equal(t, threeDSID, session.ID)
	assert.NotEmpty(t, session.InitAt)
	assert.NotEmpty(t, session.LookupAt)
	assert.Empty(t, session.AuthenticatedAt)
	assert.Equal(t, 2, session.Revision)
	assert.Equal(t, "ptrxid1", session.ProcessorTransactionID)
	assert.Equal(t, "dstrxid1", session.DsTransactionID)
	assert.Equal(t, "Y", session.Enrolled)
	assert.Equal(t, "status1", session.Status)
	assert.Equal(t, "http://localhost/challenge", session.ChallengeURL)
	assert.Equal(t, "payload", session.Payload)

	auth, err := ops.Authenticate3DS(ctx, b, tabapay.Authenticate3DSArgs{
		IdempotencyKey: idempotencyKey,
		ThreeDSID:      threeDSID,
		JWT:            "authJWT",
	})
	require.NoError(t, err)
	assert.Equal(t, "ptrxid2", auth.ProcessorTransactionID)
	assert.Equal(t, "dstrxid2", auth.DsTransactionID)
	assert.Equal(t, "Y", auth.Enrolled)
	assert.Equal(t, "status2", auth.Status)
	assert.Equal(t, "eci", auth.ECI)
	assert.Equal(t, "ucaf", auth.UCAF)
	assert.Equal(t, "xid", auth.XID)

	session, err = ops.Get3DSSession(ctx, b, threeDSID)
	require.NoError(t, err)
	assert.Equal(t, threeDSID, session.ID)
	assert.NotEmpty(t, session.InitAt)
	assert.NotEmpty(t, session.LookupAt)
	assert.NotEmpty(t, session.AuthenticatedAt)
	assert.Equal(t, 3, session.Revision)
	assert.Equal(t, "ptrxid2", session.ProcessorTransactionID)
	assert.Equal(t, "dstrxid2", session.DsTransactionID)
	assert.Equal(t, "Y", session.Enrolled)
	assert.Equal(t, "status2", session.Status)
	assert.Equal(t, "http://localhost/challenge", session.ChallengeURL)
	assert.Equal(t, "payload", session.Payload)
	assert.Equal(t, "eci", session.ECI)
	assert.Equal(t, "ucaf", session.UCAF)
	assert.Equal(t, "xid", session.XID)
}
