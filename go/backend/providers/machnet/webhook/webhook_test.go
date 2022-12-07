package webhook_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"gitlab.com/fynbos/backend/providers/machnet/webhook"
	"gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestHandleUserKYCEvent(t *testing.T) {
	t.Parallel()
	b := newTestBackends(t)
	walletID := newWallet(t, b)
	mu := newMachnetUser(t, b, walletID)

	kycEvent := external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserKYCInProgress,
		UserID:         mu.ID,
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}

	workflowID := uuid.NewString()
	worklflowRunID := uuid.NewString()
	_, err := ops.CreateUserWorkflowRef(context.Background(), b, machnet.CreateUserWorkflowRefArgs{
		UserID:        mu.ID,
		WorkflowID:    workflowID,
		WorkflowRunID: worklflowRunID,
		ActivityName:  "ActivityName",
	})
	require.NoError(t, err)
	b.temporal.On("SignalWorkflow", context.Background(), workflowID, worklflowRunID, ops.UserEventsChannel, kycEvent).Return(nil)

	err = webhook.HandleUserKYCEvent(context.Background(), b, kycEvent)
	require.NoError(t, err)

	u, err := ops.GetUserByID(context.Background(), b, mu.ID)
	require.NoError(t, err)
	assert.Equal(t, machnet.KYCStatusInProgress, u.KYCStatus)

	// User Not Found
	kycEvent = external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserKYCInProgress,
		UserID:         uuid.NewString(),
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}

	err = webhook.HandleUserKYCEvent(context.Background(), b, kycEvent)
	require.ErrorIs(t, err, machnet.ErrNotFound)

	// User Verified
	kycEvent = external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserKYCVerified,
		UserID:         mu.ID,
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}

	b.temporal.On("SignalWorkflow", context.Background(), workflowID, worklflowRunID, ops.UserEventsChannel, kycEvent).Return(nil)

	err = webhook.HandleUserKYCEvent(context.Background(), b, kycEvent)
	require.NoError(t, err)

	u, err = ops.GetUserByID(context.Background(), b, mu.ID)
	require.NoError(t, err)
	assert.Equal(t, machnet.KYCStatusVerified, u.KYCStatus)
}

func TestHandleUserCardAddedEvent(t *testing.T) {
	t.Parallel()
	b := newTestBackends(t)
	walletID := newWallet(t, b)
	user := newMachnetUser(t, b, walletID)
	externalFundingsource, err := b.external.CreateUserFundingsource(
		context.Background(),
		external.FundingSource{
			ID:                 uuid.NewString(),
			UserID:             user.ID,
			AccountNumber:      "9991",
			FundingsourceName:  "VISA-9991",
			FundingsourceType:  "CARD",
			InstitutionName:    "VISA",
			VerificationStatus: external.StatusVerified,
		},
	)
	require.NoError(t, err)
	userCardAddedEvent := external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserCardAdded,
		ResourceID:     externalFundingsource.ID,
		UserID:         externalFundingsource.UserID,
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}

	b.linkedaccounts.EXPECT().Create(gomock.Any(), &linkedaccounts.CreateArgs{
		WalletID:   walletID,
		Name:       externalFundingsource.FundingsourceName,
		Provider:   machnet.ProviderName,
		ProviderID: externalFundingsource.ID,
		Type:       machnet.TypeSendCard,
		Mask:       externalFundingsource.AccountNumber,
	}).Return(
		&linkedaccounts.LinkedAccount{ID: uuid.NewString(), WalletID: walletID},
		nil,
	).Times(1)

	err = webhook.HandleEvent(context.Background(), b, userCardAddedEvent)
	require.NoError(t, err)
}

func TestHandleBankAccountAddedEvent(t *testing.T) {
	t.Parallel()
	b := newTestBackends(t)
	walletID := newWallet(t, b)
	user := newMachnetUser(t, b, walletID)
	externalFundingsource, err := b.external.CreateUserFundingsource(
		context.Background(),
		external.FundingSource{
			ID:                 uuid.NewString(),
			UserID:             user.ID,
			AccountNumber:      "74788569",
			FundingsourceName:  "BOI-74788569",
			FundingsourceType:  string(external.FundingSourceTypeBankAccount),
			InstitutionName:    "BANK_OF_IRELAND",
			VerificationStatus: external.StatusVerified,
		},
	)
	require.NoError(t, err)
	userCardAddedEvent := external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserBankAdded,
		ResourceID:     externalFundingsource.ID,
		UserID:         externalFundingsource.UserID,
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}

	b.linkedaccounts.EXPECT().Create(gomock.Any(), &linkedaccounts.CreateArgs{
		WalletID:   walletID,
		Name:       externalFundingsource.FundingsourceName,
		Provider:   machnet.ProviderName,
		ProviderID: externalFundingsource.ID,
		Type:       machnet.TypeBankAccount,
		Mask:       externalFundingsource.AccountNumber,
	}).Return(
		&linkedaccounts.LinkedAccount{ID: uuid.NewString(), WalletID: walletID},
		nil,
	).Times(1)

	err = webhook.HandleEvent(context.Background(), b, userCardAddedEvent)
	require.NoError(t, err)
}

func TestValidateWebhook(t *testing.T) {
	t.Parallel()
	// data from webhook received from Machnet sandbox
	signature := "b9d26b2907effc7ef0babbe86fe89741b6a02add18894fb1fca4eddf8991528c"
	payload := []byte(`{"event_name":"user_kyc_verified","persisted_object_id":"63f71c24-6ff7-46fa-98ef-2bd03a57647b","resource_id":"3995a94e-37dc-46f9-97d9-3ff7c55f5141","subscription_id":"ccbe409f-b6bc-4db3-b05e-fea1b02f956f","timestamp":"2022-10-21T08:39:37.447525","user_id":"3995a94e-37dc-46f9-97d9-3ff7c55f5141"}`)

	err := webhook.ValidateWebhook(payload, "test", signature)
	require.NoError(t, err)

	err = webhook.ValidateWebhook(payload, "fail", signature)
	assert.ErrorIs(t, err, machnet.ErrInvalidSignature)
}

func TestTransactionEventWebhook(t *testing.T) {
	t.Parallel()
	b := newTestBackends(t)
	walletID := newWallet(t, b)
	e := external.Event{
		ID:         uuid.NewString(),
		EventName:  external.TransactionProcessedEvent,
		ResourceID: uuid.NewString(),
		UserID:     uuid.NewString(),
	}
	workflowID := uuid.NewString()
	worklflowRunID := uuid.NewString()
	_, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: e.UserID,
	})
	require.NoError(t, err)
	_, err = ops.CreateTransactionWorkflowRef(context.Background(), b, machnet.CreateTransactionWorkflowRefArgs{
		ID:            e.ResourceID,
		SendUserID:    e.UserID,
		WorkflowID:    workflowID,
		WorkflowRunID: worklflowRunID,
	})
	require.NoError(t, err)
	b.temporal.On("SignalWorkflow", context.Background(), workflowID, worklflowRunID, ops.TransactionEventsChannel, e).Return(nil)

	err = webhook.HandleEvent(context.Background(), b, e)
	require.NoError(t, err)
}

func TestTransactionDeliveryEventWebhook(t *testing.T) {
	t.Parallel()
	b := newTestBackends(t)
	walletID := newWallet(t, b)
	e := external.Event{
		ID:         uuid.NewString(),
		EventName:  external.TransactionDeliveredEvent,
		ResourceID: uuid.NewString(),
		UserID:     uuid.NewString(),
	}
	workflowID := uuid.NewString()
	worklflowRunID := uuid.NewString()
	_, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: e.UserID,
	})
	require.NoError(t, err)
	_, err = ops.CreateTransactionWorkflowRef(context.Background(), b, machnet.CreateTransactionWorkflowRefArgs{
		ID:            e.ResourceID,
		SendUserID:    e.UserID,
		WorkflowID:    workflowID,
		WorkflowRunID: worklflowRunID,
	})
	require.NoError(t, err)
	b.temporal.On("SignalWorkflow", context.Background(), workflowID, worklflowRunID, ops.TransactionDeliveryEventsChannel, e).Return(nil)

	err = webhook.HandleEvent(context.Background(), b, e)
	require.NoError(t, err)
}

func TestSaveWebhook(t *testing.T) {
	t.Parallel()
	b := newTestBackends(t)
	walletID := newWallet(t, b)
	e := external.Event{
		ID:         uuid.NewString(),
		EventName:  external.TransactionDeliveredEvent,
		ResourceID: uuid.NewString(),
		UserID:     uuid.NewString(),
	}

	_, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: e.UserID,
	})
	require.NoError(t, err)

	err = webhook.SaveWebhook(context.Background(), b, e)
	require.NoError(t, err)

	// Do the same webhook as a no-op, but not an error
	err = webhook.SaveWebhook(context.Background(), b, e)
	require.NoError(t, err)
}

func newWallet(t *testing.T, b webhook.Backends) string {
	walletID := uuid.NewString()
	_, err := b.DB().Exec(
		"INSERT INTO wallets (id, name) VALUES ($1, $2);",
		walletID,
		"test",
	)
	require.NoError(t, err)

	return walletID
}

func newMachnetUser(t *testing.T, b ops.Backends, walletID string) *machnet.User {
	externalUser, err := b.External().RegisterUser(context.Background(), external.User{
		Type: external.TypeSendUser,
	})
	require.NoError(t, err)
	user, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalUser.ID,
	})
	require.NoError(t, err)

	return user
}

func newTestBackends(t *testing.T) testBackends {
	ctrl := gomock.NewController(t)
	return testBackends{
		db:             test_utils.MigrateCockroachDB(t, context.Background()),
		external:       external_client.New(),
		linkedaccounts: linkedaccounts_mock.NewMockClient(ctrl),
		temporal:       &mocks.Client{},
	}
}

type testBackends struct {
	db             *sqlx.DB
	external       *external_client.Client
	linkedaccounts *linkedaccounts_mock.MockClient
	users          user.Client
	kycImpl        kyc.Client
	temporal       *mocks.Client
}

func (b testBackends) Machnet() machnet.Client {
	return nil
}

func (b testBackends) Users() user.Client {
	return b.users
}

func (b testBackends) KYC() kyc.Client {
	return b.kycImpl
}

func (b testBackends) DB() *sqlx.DB {
	return b.db
}

func (b testBackends) External() external.Client {
	return b.external
}

func (b testBackends) LinkedAccounts() linkedaccounts.Client {
	return b.linkedaccounts
}

func (b testBackends) Temporal() client.Client {
	return b.temporal
}
