package ops_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/accounts"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/identity"
	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/external"
	external_mock "gitlab.com/fynbos/backend/providers/unit/external/client/mock"
	"gitlab.com/fynbos/backend/providers/unit/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/mocks"
)

func TestVerifyWebhook(s *testing.T) {
	s.Parallel()
	ctx := context.Background()
	s.Run("Successfully verifies incoming webhook request", func(t *testing.T) {
		body := []byte(`{"data":[{"id":"2504140","type":"customer.created","attributes":{"createdAt":"2022-05-18T14:35:00.702Z","tags":{"userID":"02242b61-a99e-4b44-bda7-cf6a4f535a5f","test":"webhook-tag","key":"another-tag","number":"111"}},"relationships":{"customer":{"data":{"id":"344063","type":"individualCustomer"}},"application":{"data":{"id":"404728","type":"individualApplication"}}}}]}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg=" // key = fynbos_local_unit_webhook_token

		err := ops.VerifyWebhook(ctx, body, signature, "fynbos_local_unit_webhook_token")

		assert.NoError(t, err)
	})

	s.Run("Fails if signature of webhook body does not match provided one", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg"

		err := ops.VerifyWebhook(ctx, body, signature, "fynbos_local_unit_webhook_token")

		assert.ErrorIs(t, err, unit.ErrUnauthorized)
	})
}

func TestCreateAndGetCustomer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockExternalClient := external_mock.NewMockClient(ctrl)
	mockIdentityService := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, ctx),
		mockIdentityService,
		accounts_mock.NewMockClient(ctrl),
		&mocks.Client{},
	)
	identityID := uuid.NewString()
	mockIdentityService.EXPECT().Get(context.Background(), identityID).
		Return(
			&identity.Identity{ID: identityID},
			nil,
		)

	customerID := uuid.NewString()
	customerType := "individual"
	customer, err := ops.CreateCustomer(ctx, b, mockExternalClient, &unit.CreateCustomerArgs{
		ID:         customerID,
		IdentityID: identityID,
		Type:       customerType,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customer.ID)
	assert.Equal(t, identityID, customer.IdentityID)
	assert.Equal(t, customerType, customer.Type)

	customerByID, err := ops.GetCustomer(ctx, b, customer.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customerByID.ID)
	assert.Equal(t, identityID, customerByID.IdentityID)
	assert.Equal(t, customerType, customerByID.Type)

	customerByIdentityID, err := ops.GetCustomerByIdentityID(ctx, b, identityID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customerByIdentityID.ID)
	assert.Equal(t, identityID, customerByIdentityID.IdentityID)
	assert.Equal(t, customerType, customerByIdentityID.Type)
}

func TestCreateAndGetCounterParty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	counterpartyID := uuid.NewString()
	args := &unit.CreateCounterPartyArgs{
		FundingsourceID: uuid.NewString(),
		Name:            "test name",
		UnitCustomerID:  uuid.NewString(),
		RoutingNumber:   faker.CCNumber(),
		AccountNumber:   faker.CCNumber(),
		AccountType:     faker.CCType(),
		Type:            "person",
		IdempotencyKey:  "test",
	}
	mockExternalClient := external_mock.NewMockClient(ctrl)
	mockExternalClient.EXPECT().CreateCounterparty(context.Background(), &external.CreateCounterpartyArgs{
		Name:           "test name",
		UnitCustomerID: args.UnitCustomerID,
		RoutingNumber:  args.RoutingNumber,
		AccountNumber:  args.AccountNumber,
		AccountType:    args.AccountType,
		Type:           args.Type,
		IdempotencyKey: args.IdempotencyKey,
	}).Return(
		&external.Counterparty{
			ID:   counterpartyID,
			Type: "achCounterparty",
			Relationships: external.CounterpartyRelationships{
				Customer: external.Relationship{
					Data: external.TypeData{
						ID:   args.UnitCustomerID,
						Type: "customer",
					},
				},
			},
		},
		nil,
	).Times(1)
	mockIdentityService := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, ctx),
		mockIdentityService,
		accounts_mock.NewMockClient(ctrl),
		&mocks.Client{},
	)

	unitCounterparty, err := ops.CreateCounterParty(ctx, b, mockExternalClient, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.FundingsourceID, args.FundingsourceID)
	assert.NotEqual(t, "", unitCounterparty.ID)

	freshCounterParty, err := ops.GetCounterPartyByFundingsourceID(ctx, b, args.FundingsourceID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.FundingsourceID, freshCounterParty.FundingsourceID)
	assert.NotEqual(t, "", freshCounterParty.ID)
}

func TestCreateAndGetDepositAccount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	unitCustomerID := uuid.NewString()
	unitDepositAccountID := uuid.NewString()
	mockAccountService := accounts_mock.NewMockClient(ctrl)
	mockExternalClient := external_mock.NewMockClient(ctrl)
	mockIdentityService := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, ctx),
		mockIdentityService,
		mockAccountService,
		&mocks.Client{},
	)
	idempotencyKey := sha256.Sum256([]byte(unitCustomerID))
	mockExternalClient.EXPECT().CreateDepositAccount(context.Background(), &external.CreateDepositAccountArgs{
		CustomerID:     unitCustomerID,
		DepositProduct: "checking",
		Type:           "depositAccount",
		IdempotencyKey: string(idempotencyKey[0:]),
	}).Return(
		&external.DepositAccount{
			ID:   unitDepositAccountID,
			Type: "depositAccount",
			Attributes: external.DepositAccountAttributes{
				CreatedAt:        "2000-05-11T10:19:30.409Z",
				Name:             "Peter parker",
				Status:           "Open",
				DepositProduct:   "checking",
				RoutingNumber:    "812345678",
				AccountNumber:    "1000000002",
				Currency:         "USD",
				BalanceInCents:   10000,
				HoldInCents:      1000,
				AvailableInCents: 9000,
			},
			Relationships: &external.DepositAccountRelationships{
				Customer: external.Relationship{
					Data: external.TypeData{
						ID:   unitCustomerID,
						Type: "customer",
					},
				},
			},
		},
		nil,
	).Times(1)

	createdAcc, err := ops.CreateDepositAccount(ctx, b, mockExternalClient, unitCustomerID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, unitCustomerID, createdAcc.CustomerID)

	freshAcc, err := ops.GetDepositAccount(ctx, b, unitDepositAccountID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, unitCustomerID, freshAcc.CustomerID)
}

func TestInitiateUserDeposit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := uuid.NewString()
	accountID := uuid.NewString()
	depositID := uuid.NewString()
	depositAccountID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	achPaymentID := uuid.NewString()
	counterpartyID := uuid.NewString()
	ctrl := gomock.NewController(t)
	mockExternalClient := external_mock.NewMockClient(ctrl)
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		&mocks.Client{},
	)

	mockIdentity.EXPECT().Get(context.Background(), userID).Return(&identity.Identity{ID: userID}, nil)
	customer, err := ops.CreateCustomer(ctx, b, mockExternalClient, &unit.CreateCustomerArgs{
		ID:         "58",
		IdentityID: userID,
		Type:       "person",
	})
	if err != nil {
		t.Fatal(err)
	}

	mockAccounts.EXPECT().Get(context.Background(), accountID).Return(
		&accounts.Account{ID: accountID, Provider: "unit", ProviderID: depositAccountID},
		nil,
	)
	mockExternalClient.EXPECT().CreateCounterparty(context.Background(), gomock.Any()).Return(
		&external.Counterparty{
			ID:   counterpartyID,
			Type: "achCounterparty",
		},
		nil,
	).Times(1)
	_, err = ops.CreateCounterParty(context.Background(), b, mockExternalClient, &unit.CreateCounterPartyArgs{
		FundingsourceID: fundingsourceID,
		Name:            "test",
		UnitCustomerID:  customer.ID,
		RoutingNumber:   "123",
		AccountNumber:   "123",
		AccountType:     "Checking",
		Type:            "person",
		IdempotencyKey:  "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	idempotencyKey := sha256.Sum256([]byte(depositID))
	mockExternalClient.EXPECT().OriginateAch(context.Background(), &external.OriginateAchArgs{
		IdempotencyKey:   string(idempotencyKey[0:]),
		Amount:           10000,
		Direction:        "Debit",
		CounterpartyID:   counterpartyID,
		DepositAccountID: depositAccountID,
		Description:      "Funding",
	}).Return(
		&external.AchPayment{
			ID:   achPaymentID,
			Type: "achPayment",
		},
		nil,
	).Times(1)
	achDeposit, err := ops.InitiateUserDeposit(context.Background(), b, mockExternalClient, &unit.InitiateUserDepositArgs{
		DepositID:       depositID,
		AccountID:       accountID,
		FundingsourceID: fundingsourceID,
		Amount:          10000,
		Description:     "Funding",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, achPaymentID, achDeposit.ID)
	assert.Equal(t, depositAccountID, achDeposit.DepositAccountID)
	assert.Equal(t, counterpartyID, achDeposit.CounterPartyID)
	assert.Equal(t, uint64(10000), achDeposit.Amount)
	assert.Equal(t, depositID, achDeposit.DepositID)
}

func TestHandleCreatedCustomerEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	temporalMockClient := &mocks.Client{}
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		temporalMockClient,
	)

	scenarios := []struct {
		Name            string
		OnboardingError error
	}{
		{
			Name:            "Succeeds if unit onboarding is initiated.",
			OnboardingError: nil,
		},
	}
	for _, scenario := range scenarios {
		customerCreatedEvent := NewCustomerCreatedEvent()

		temporalMockClient.On("SignalWorkflow", mock.Anything, "unit_onboarding_"+customerCreatedEvent.Attributes.Tags.FynbosUserID, mock.AnythingOfType("string"), "onboard-unit-customer-created", mock.Anything).Return(scenario.OnboardingError).Times(1)

		rawEvent := marshalEvent(t, customerCreatedEvent)
		err := ops.HandleEvent(ctx, b, external.Event{ID: customerCreatedEvent.ID, Type: external.EventType(customerCreatedEvent.Type)}, rawEvent)

		if scenario.OnboardingError != nil {
			assert.ErrorIs(t, err, unit.ErrInternal, scenario.Name)
		} else {
			assert.NoError(t, err, scenario.Name)
		}
	}
}

func TestHandleApplicationDeniedEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	temporalMockClient := &mocks.Client{}
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		temporalMockClient,
	)

	scenarios := []struct {
		Name            string
		OnboardingError error
	}{
		{
			Name:            "Succeeds if unit onboarding is initiated.",
			OnboardingError: nil,
		},
	}
	for _, scenario := range scenarios {
		applicationDeniedEvent := NewApplicationDeniedEvent()

		temporalMockClient.On("SignalWorkflow", mock.Anything, "unit_onboarding_"+applicationDeniedEvent.Attributes.Tags.FynbosUserID, mock.AnythingOfType("string"), "onboard-unit-application-denied", mock.Anything).Return(scenario.OnboardingError).Times(1)

		rawEvent := marshalEvent(t, applicationDeniedEvent)
		err := ops.HandleEvent(ctx, b, external.Event{ID: applicationDeniedEvent.ID, Type: external.EventType(applicationDeniedEvent.Type)}, rawEvent)

		if scenario.OnboardingError != nil {
			assert.ErrorIs(t, err, unit.ErrInternal, scenario.Name)
		} else {
			assert.NoError(t, err, scenario.Name)
		}
	}
}

func TestHandlePaymentEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	temporalMockClient := &mocks.Client{}
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		temporalMockClient,
	)

	depositID := uuid.NewString()
	paymentEvent := external.AchPayment{
		Type: string(external.PAYMENT_SENT),
		ID:   uuid.NewString(),
		Attributes: external.AchPaymentAttributes{
			Tags: external.DepositTags{
				DepositID: depositID,
			},
		},
	}

	cases := []external.EventType{
		external.PAYMENT_CANCELED,
		external.PAYMENT_CLEARING,
		external.PAYMENT_CREATED,
		external.PAYMENT_REJECTED,
		external.PAYMENT_RETURNED,
		external.PAYMENT_SENT,
		external.PAYMENT_PENDING_REVIEW,
	}

	for _, eventType := range cases {
		t.Run(string(eventType), func(st *testing.T) {
			temporalMockClient.On(
				"SignalWorkflow",
				mock.Anything,
				"deposit_"+depositID,
				mock.AnythingOfType("string"),
				"unit-user-ach-deposit",
				string(eventType),
			).Return(nil).Times(1)
			paymentEvent.ID = uuid.NewString()
			paymentEvent.Type = string(eventType)
			rawEvent := marshalEvent(t, paymentEvent)

			err := ops.HandleEvent(ctx, b, external.Event{ID: paymentEvent.ID, Type: eventType}, rawEvent)

			assert.NoError(t, err)
		})
	}
}

func TestDontFailForUnknownEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	temporalMockClient := &mocks.Client{}
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		temporalMockClient,
	)

	customerCreatedEvent := NewCustomerCreatedEvent()

	rawEvent := marshalBody(t, customerCreatedEvent)
	err := ops.HandleEvent(ctx, b, external.Event{ID: customerCreatedEvent.ID, Type: external.EventType("unknown")}, rawEvent.Bytes())
	assert.NoError(t, err)
}

func TestStoreEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	temporalMockClient := &mocks.Client{}
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		temporalMockClient,
	)

	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvent := marshalEvent(t, customerCreatedEvent)
	testEvent := external.Event{ID: customerCreatedEvent.ID, Type: external.EventType(customerCreatedEvent.Type)}

	storedEvent, err := ops.StoreEvent(ctx, b, testEvent, rawEvent)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testEvent.ID, storedEvent.ID)
	assert.Equal(t, testEvent.Type, external.EventType(storedEvent.Type))
	assert.JSONEq(t, string(rawEvent), storedEvent.RawEvent.String())
}

func TestStoreDuplicateEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	temporalMockClient := &mocks.Client{}
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(
		t,
		test_utils.MigrateCockroachDB(t, context.Background()),
		mockIdentity,
		mockAccounts,
		temporalMockClient,
	)

	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvent := marshalEvent(t, customerCreatedEvent)
	testEvent := external.Event{ID: customerCreatedEvent.ID, Type: external.EventType(customerCreatedEvent.Type)}

	_, err := ops.StoreEvent(ctx, b, testEvent, rawEvent)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ops.StoreEvent(ctx, b, testEvent, rawEvent)

	assert.ErrorIs(t, err, unit.ErrDuplicateEvent)
}

func marshalBody(t *testing.T, events ...interface{}) *bytes.Buffer {
	body := struct {
		Data []json.RawMessage `json:"data"`
	}{}
	for _, event := range events {
		rawEvent, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		body.Data = append(body.Data, rawEvent)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewBuffer(raw)
}

func marshalEvent(t *testing.T, event interface{}) json.RawMessage {
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func NewCustomerCreatedEvent() external.CustomerCreatedEvent {
	return external.CustomerCreatedEvent{
		ID:   uuid.NewString(),
		Type: "customer.created",
		Attributes: external.EventAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: external.ApplicationTags{
				FynbosUserID: uuid.NewString(),
			},
		},
		Relationships: external.EventRelationships{
			Customer: external.JsonCustomer{
				Data: external.Data{
					ID:   "52",
					Type: "individualCustomer",
				},
			},
			Application: external.JsonApplication{
				Data: external.Data{
					ID:   "52",
					Type: "individualApplication",
				},
			},
		},
	}
}

func NewApplicationDeniedEvent() external.ApplicationDeniedEvent {
	return external.ApplicationDeniedEvent{
		ID:   uuid.NewString(),
		Type: "application.denied",
		Attributes: external.EventAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: external.ApplicationTags{
				FynbosUserID: uuid.NewString(),
			},
		},
		Relationships: external.EventRelationships{
			Application: external.JsonApplication{
				Data: external.Data{
					ID:   "52",
					Type: "individualApplication",
				},
			},
		},
	}
}
