package ops_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/ops"
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

// func TestCreateAndGetCustomer(t *testing.T) {
// 	t.Parallel()
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	mockIdentityService := identity_mock.NewMockClient(ctrl)
// 	b := ops.NewTestBackends(
// 		t,
// 		&sqlx.DB{},
// 		mockIdentityService,
// 		accounts_mock.NewMockClient(ctrl),
// 	)
// 	identityID := uuid.NewString()
// 	mockIdentityService.EXPECT().Get(context.Background(), identityID).
// 		Return(
// 			&identity.Identity{ID: identityID},
// 			nil,
// 		)

// 	customerID := uuid.NewString()
// 	customerType := "individual"
// 	customer, err := ops.CreateCustomer(ctx, &CreateCustomerArgs{
// 		ID:         customerID,
// 		IdentityID: identityID,
// 		Type:       customerType,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, customerID, customer.ID)
// 	assert.Equal(t, identityID, customer.IdentityID)
// 	assert.Equal(t, customerType, customer.Type)

// 	customerByID, err := unitService.GetCustomer(ctx, customer.ID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, customerID, customerByID.ID)
// 	assert.Equal(t, identityID, customerByID.IdentityID)
// 	assert.Equal(t, customerType, customerByID.Type)

// 	customerByIdentityID, err := unitService.GetCustomerByIdentityID(ctx, identityID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, customerID, customerByIdentityID.ID)
// 	assert.Equal(t, identityID, customerByIdentityID.IdentityID)
// 	assert.Equal(t, customerType, customerByIdentityID.Type)
// }

// func TestCreateAndGetCounterParty(t *testing.T) {
// 	t.Parallel()
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	counterpartyID := uuid.NewString()
// 	unitCustomerID := uuid.NewString()
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != "POST" {
// 			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		if r.URL.Path != "/counterparties" {
// 			http.Error(w, "Not found.", http.StatusNotFound)
// 			return
// 		}

// 		counterpartyResponse := &external.CounterpartyRequest{
// 			Data: external.Counterparty{
// 				ID:   counterpartyID,
// 				Type: "achCounterparty",
// 				Relationships: external.CounterpartyRelationships{
// 					Customer: external.Relationship{
// 						Data: external.TypeData{
// 							ID:   unitCustomerID,
// 							Type: "customer",
// 						},
// 					},
// 				},
// 			},
// 		}
// 		payload, err := json.Marshal(counterpartyResponse)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		w.WriteHeader(http.StatusCreated)
// 		_, err = w.Write([]byte(payload))
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 	}))
// 	t.Cleanup(func() {
// 		server.Close()
// 	})
// 	unitService, err := NewService(ServiceArgs{
// 		WebhookToken:    "fynbos_local_unit_webhook_token",
// 		BaseURL:         server.URL,
// 		Token:           "test token",
// 		Db:              test_utils.MigrateCockroachDB(t, ctx),
// 		IdentityService: identity_mock.NewMockClient(ctrl),
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	args := &CreateCounterPartyArgs{
// 		FundingsourceID: uuid.NewString(),
// 		Name:            "test name",
// 		UnitCustomerID:  uuid.NewString(),
// 		RoutingNumber:   faker.CCNumber(),
// 		AccountNumber:   faker.CCNumber(),
// 		AccountType:     faker.CCType(),
// 		Type:            "person",
// 		IdempotencyKey:  "test",
// 	}

// 	unitCounterparty, err := unitService.CreateCounterParty(ctx, args)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, args.FundingsourceID, args.FundingsourceID)
// 	assert.NotEqual(t, "", unitCounterparty.ID)

// 	freshCounterParty, err := unitService.GetCounterPartyByFundingsourceID(ctx, args.FundingsourceID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, args.FundingsourceID, freshCounterParty.FundingsourceID)
// 	assert.NotEqual(t, "", freshCounterParty.ID)
// }

// func TestCreateAndGetDepositAccount(t *testing.T) {
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	unitCustomerID := uuid.NewString()
// 	unitDepositAccountID := uuid.NewString()
// 	mockAccountService := accounts_mock.NewMockClient(ctrl)
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != "POST" {
// 			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		if r.URL.Path != "/accounts" {
// 			http.Error(w, "Not found.", http.StatusNotFound)
// 			return
// 		}

// 		data := external.DepositAccount{
// 			ID:   unitDepositAccountID,
// 			Type: "depositAccount",
// 			Attributes: external.DepositAccountAttributes{
// 				CreatedAt:        "2000-05-11T10:19:30.409Z",
// 				Name:             "Peter parker",
// 				Status:           "Open",
// 				DepositProduct:   "checking",
// 				RoutingNumber:    "812345678",
// 				AccountNumber:    "1000000002",
// 				Currency:         "USD",
// 				BalanceInCents:   10000,
// 				HoldInCents:      1000,
// 				AvailableInCents: 9000,
// 			},
// 			Relationships: &external.DepositAccountRelationships{
// 				Customer: external.Relationship{
// 					Data: external.TypeData{
// 						ID:   unitCustomerID,
// 						Type: "customer",
// 					},
// 				},
// 			},
// 		}
// 		rawData, err := json.Marshal(data)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		depositAccountResponse := &external.Response{
// 			Data: rawData,
// 		}
// 		payload, err := json.Marshal(depositAccountResponse)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		w.WriteHeader(http.StatusCreated)
// 		_, err = w.Write([]byte(payload))
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 	}))
// 	t.Cleanup(func() {
// 		server.Close()
// 	})
// 	unitService, err := NewService(ServiceArgs{
// 		WebhookToken:    "fynbos_local_unit_webhook_token",
// 		BaseURL:         server.URL,
// 		Token:           "test token",
// 		Db:              test_utils.MigrateCockroachDB(t, ctx),
// 		IdentityService: identity_mock.NewMockClient(ctrl),
// 		AccountClient:   mockAccountService,
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	createdAcc, err := unitService.CreateDepositAccount(ctx, unitCustomerID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, unitCustomerID, createdAcc.CustomerID)

// 	freshAcc, err := unitService.GetDepositAccount(ctx, unitDepositAccountID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, unitCustomerID, freshAcc.CustomerID)
// }

// func TestInitiateUserDeposit(t *testing.T) {
// 	t.Parallel()
// 	userID := uuid.NewString()
// 	accountID := uuid.NewString()
// 	depositID := uuid.NewString()
// 	depositAccountID := uuid.NewString()
// 	fundingsourceID := uuid.NewString()
// 	achPaymentID := uuid.NewString()
// 	counterpartyID := uuid.NewString()
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != "POST" {
// 			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		var data any
// 		if r.URL.Path == "/payments" {
// 			data = external.AchPayment{
// 				ID:   achPaymentID,
// 				Type: "achPayment",
// 			}
// 		} else if r.URL.Path == "/counterparties" {
// 			data = external.Counterparty{
// 				Type: "counterparty",
// 				ID:   counterpartyID,
// 			}
// 		} else {
// 			http.NotFound(w, r)
// 			return
// 		}

// 		rawData, err := json.Marshal(data)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		response := &external.Response{
// 			Data: rawData,
// 		}
// 		payload, err := json.Marshal(response)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		w.WriteHeader(http.StatusCreated)
// 		_, err = w.Write([]byte(payload))
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 	}))
// 	t.Cleanup(func() {
// 		server.Close()
// 	})
// 	ctrl := gomock.NewController(t)
// 	mockAccounts := accounts_mock.NewMockClient(ctrl)
// 	mockIdentity := identity_mock.NewMockClient(ctrl)
// 	unitService, err := NewService(ServiceArgs{
// 		WebhookToken:    "fynbos_local_unit_webhook_token",
// 		BaseURL:         server.URL,
// 		Token:           "test token",
// 		Db:              test_utils.MigrateCockroachDB(t, context.Background()),
// 		IdentityService: mockIdentity,
// 		Logger:          zap.NewNop(),
// 		AccountClient:   mockAccounts,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	mockIdentity.EXPECT().Get(context.Background(), userID).Return(&identity.Identity{ID: userID}, nil)
// 	customer, err := unitService.CreateCustomer(context.Background(), &CreateCustomerArgs{
// 		ID:         "58",
// 		IdentityID: userID,
// 		Type:       "person",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	mockAccounts.EXPECT().Get(context.Background(), accountID).Return(
// 		&accounts.Account{ID: accountID, Provider: "unit", ProviderID: depositAccountID},
// 		nil,
// 	)

// 	_, err = unitService.CreateCounterParty(context.Background(), &CreateCounterPartyArgs{
// 		FundingsourceID: fundingsourceID,
// 		Name:            "test",
// 		UnitCustomerID:  customer.ID,
// 		RoutingNumber:   "123",
// 		AccountNumber:   "123",
// 		AccountType:     "Checking",
// 		Type:            "person",
// 		IdempotencyKey:  "test",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	achDeposit, err := unitService.InitiateUserDeposit(context.Background(), &InitiateUserDepositArgs{
// 		DepositID:       depositID,
// 		AccountID:       accountID,
// 		FundingsourceID: fundingsourceID,
// 		Amount:          10000,
// 		Description:     "Funding",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, achPaymentID, achDeposit.ID)
// 	assert.Equal(t, depositAccountID, achDeposit.DepositAccountID)
// 	assert.Equal(t, counterpartyID, achDeposit.CounterPartyID)
// 	assert.Equal(t, uint64(10000), achDeposit.Amount)
// 	assert.Equal(t, depositID, achDeposit.DepositID)
// }

// func TestWebhook(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()

// 	db := test_utils.MigrateCockroachDB(t, ctx)
// 	ctrl := gomock.NewController(t)
// 	providerMock := NewMockService(ctrl)
// 	temporalMockClient := &mocks.Client{}

// 	wh, err := NewWebhook(&WebhookArgs{
// 		Db: db,
// 		Up: providerMock,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	svr := httptest.NewServer(wh.MakeHttpHandler())

// 	t.Cleanup(func() {
// 		svr.Close()
// 	})

// 	scenarios := []struct {
// 		Name                   string
// 		VerifyError            error
// 		Payload                *bytes.Buffer
// 		ExpectedHttpStatusCode int
// 		ResponseMessage        string
// 		MockCallTimes          int
// 	}{
// 		{
// 			Name:                   "Returns 200",
// 			VerifyError:            nil,
// 			Payload:                marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent()),
// 			ExpectedHttpStatusCode: 200,
// 			MockCallTimes:          2,
// 		},
// 		{
// 			Name:                   "Returns 500 if webhook fails verification",
// 			VerifyError:            errors.New("test"),
// 			Payload:                marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent()),
// 			ExpectedHttpStatusCode: 401,
// 			ResponseMessage:        "Signature didn't match.\n",
// 			MockCallTimes:          0,
// 		},
// 		{
// 			Name:                   "Returns 500 if marshalling payload fails",
// 			VerifyError:            nil,
// 			Payload:                bytes.NewBuffer([]byte("")),
// 			ExpectedHttpStatusCode: 500,
// 			ResponseMessage:        "Failed to parse payload\n",
// 			MockCallTimes:          0,
// 		},
// 		{
// 			Name:                   "Tries to handle all events even if first one fails",
// 			VerifyError:            nil,
// 			Payload:                marshalBody(t, "", NewCustomerCreatedEvent()),
// 			ExpectedHttpStatusCode: 500,
// 			ResponseMessage:        "Failed to parse payload\n",
// 			MockCallTimes:          1,
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		t.Run(scenario.Name, func(t *testing.T) {
// 			temporalMockClient.On("SignalWorkflow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "onboard-unit-customer-created", mock.Anything).Return(nil).Times(scenario.MockCallTimes)
// 			providerMock.EXPECT().VerifyWebhook(context.Background(), gomock.Any(), gomock.Any()).Return(scenario.VerifyError)
// 			resp, err := http.Post(svr.URL, "application/json", scenario.Payload)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			body, err := io.ReadAll(resp.Body)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			assert.Equal(t, scenario.ExpectedHttpStatusCode, resp.StatusCode)
// 			assert.Equal(t, scenario.ResponseMessage, string(body))
// 		})
// 	}
// }

// func TestHandleCreatedCustomerEvent(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	identityMock := identity_mock.NewMockClient(ctrl)
// 	temporalMockClient := &mocks.Client{}

// 	db := test_utils.MigrateCockroachDB(t, ctx)

// 	provider, err := NewService(ServiceArgs{
// 		BaseURL:         "localhost:8080",
// 		Token:           "token",
// 		WebhookToken:    "webhooktoken",
// 		Db:              db,
// 		IdentityService: identityMock,
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	wh, err := NewWebhook(&WebhookArgs{
// 		Up: provider,
// 		Db: db,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	scenarios := []struct {
// 		Name            string
// 		OnboardingError error
// 	}{
// 		{
// 			Name:            "Succeeds if unit onboarding is initiated.",
// 			OnboardingError: nil,
// 		},
// 	}
// 	for _, scenario := range scenarios {
// 		customerCreatedEvent := NewCustomerCreatedEvent()

// 		temporalMockClient.On("SignalWorkflow", mock.Anything, "unit_onboarding_"+customerCreatedEvent.Attributes.Tags.FynbosUserId, mock.AnythingOfType("string"), "onboard-unit-customer-created", mock.Anything).Return(scenario.OnboardingError).Times(1)

// 		rawEvent := marshalEvent(t, customerCreatedEvent)
// 		err = wh.HandleEvent(context.Background(), Event{ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}, rawEvent)

// 		if scenario.OnboardingError != nil {
// 			assert.ErrorIs(t, err, ErrInternal, scenario.Name)
// 		} else {
// 			assert.NoError(t, err, scenario.Name)
// 		}
// 	}
// }

// func TestHandleApplicationDeniedEvent(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	identityMock := identity_mock.NewMockClient(ctrl)
// 	temporalMockClient := &mocks.Client{}

// 	db := test_utils.MigrateCockroachDB(t, ctx)

// 	provider, err := NewService(ServiceArgs{
// 		BaseURL:         "localhost:8080",
// 		Token:           "token",
// 		WebhookToken:    "webhooktoken",
// 		Db:              db,
// 		IdentityService: identityMock,
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	wh, err := NewWebhook(&WebhookArgs{
// 		Up: provider,
// 		Db: db,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	scenarios := []struct {
// 		Name            string
// 		OnboardingError error
// 	}{
// 		{
// 			Name:            "Succeeds if unit onboarding is initiated.",
// 			OnboardingError: nil,
// 		},
// 	}
// 	for _, scenario := range scenarios {
// 		applicationDeniedEvent := NewApplicationDeniedEvent()

// 		temporalMockClient.On("SignalWorkflow", mock.Anything, "unit_onboarding_"+applicationDeniedEvent.Attributes.Tags.FynbosUserId, mock.AnythingOfType("string"), "onboard-unit-application-denied", mock.Anything).Return(scenario.OnboardingError).Times(1)

// 		rawEvent := marshalEvent(t, applicationDeniedEvent)
// 		err = wh.HandleEvent(context.Background(), Event{ID: applicationDeniedEvent.ID, Type: EventType(applicationDeniedEvent.Type)}, rawEvent)

// 		if scenario.OnboardingError != nil {
// 			assert.ErrorIs(t, err, ErrInternal, scenario.Name)
// 		} else {
// 			assert.NoError(t, err, scenario.Name)
// 		}
// 	}
// }

// func TestHandlePaymentEvent(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	identityMock := identity_mock.NewMockClient(ctrl)
// 	temporalMockClient := &mocks.Client{}

// 	db := test_utils.MigrateCockroachDB(t, ctx)

// 	provider, err := NewService(ServiceArgs{
// 		BaseURL:         "localhost:8080",
// 		Token:           "token",
// 		WebhookToken:    "webhooktoken",
// 		Db:              db,
// 		IdentityService: identityMock,
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	wh, err := NewWebhook(&WebhookArgs{
// 		Up: provider,
// 		Db: db,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	depositID := uuid.NewString()
// 	paymentEvent := external.AchPayment{
// 		Type: string(PAYMENT_SENT),
// 		ID:   uuid.NewString(),
// 		Attributes: external.AchPaymentAttributes{
// 			Tags: external.DepositTags{
// 				DepositID: depositID,
// 			},
// 		},
// 	}

// 	cases := []EventType{
// 		PAYMENT_CANCELED,
// 		PAYMENT_CLEARING,
// 		PAYMENT_CREATED,
// 		PAYMENT_REJECTED,
// 		PAYMENT_RETURNED,
// 		PAYMENT_SENT,
// 		PAYMENT_PENDING_REVIEW,
// 	}

// 	for _, eventType := range cases {
// 		t.Run(string(eventType), func(st *testing.T) {
// 			temporalMockClient.On(
// 				"SignalWorkflow",
// 				mock.Anything,
// 				"deposit_"+depositID,
// 				mock.AnythingOfType("string"),
// 				"unit-user-ach-deposit",
// 				string(eventType),
// 			).Return(nil).Times(1)
// 			paymentEvent.ID = uuid.NewString()
// 			paymentEvent.Type = string(eventType)
// 			rawEvent := marshalEvent(t, paymentEvent)

// 			err = wh.HandleEvent(context.Background(), Event{ID: paymentEvent.ID, Type: eventType}, rawEvent)

// 			assert.NoError(t, err)
// 		})
// 	}
// }

// func TestDontFailForUnknownEvent(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	identityMock := identity_mock.NewMockClient(ctrl)
// 	temporalMockClient := &mocks.Client{}
// 	db := test_utils.MigrateCockroachDB(t, ctx)
// 	provider, err := NewService(ServiceArgs{
// 		BaseURL:         "localhost:8080",
// 		Token:           "token",
// 		WebhookToken:    "webhooktoken",
// 		Db:              db,
// 		IdentityService: identityMock,
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	wh, err := NewWebhook(&WebhookArgs{
// 		Db: db,
// 		Up: provider,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	customerCreatedEvent := NewCustomerCreatedEvent()

// 	rawEvent := marshalBody(t, customerCreatedEvent)
// 	err = wh.HandleEvent(context.Background(), Event{ID: customerCreatedEvent.ID, Type: EventType("unknown")}, rawEvent.Bytes())
// 	assert.NoError(t, err)
// }

// func TestStoreEvent(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	identityMock := identity_mock.NewMockClient(ctrl)
// 	temporalMockClient := &mocks.Client{}
// 	db := test_utils.MigrateCockroachDB(t, ctx)
// 	provider, err := NewService(ServiceArgs{
// 		BaseURL:         "localhost:8080",
// 		Token:           "token",
// 		WebhookToken:    "webhooktoken",
// 		Db:              db,
// 		IdentityService: identityMock,
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	wh, err := NewWebhook(&WebhookArgs{
// 		Db: db,
// 		Up: provider,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	customerCreatedEvent := NewCustomerCreatedEvent()
// 	rawEvent := marshalEvent(t, customerCreatedEvent)
// 	testEvent := Event{ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}

// 	storedEvent, err := wh.StoreEvent(ctx, testEvent, rawEvent)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, testEvent.ID, storedEvent.ID)
// 	assert.Equal(t, testEvent.Type, EventType(storedEvent.Type))
// 	assert.JSONEq(t, string(rawEvent), storedEvent.RawEvent.String())
// }

// func TestStoreDuplicateEvent(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	identityMock := identity_mock.NewMockClient(ctrl)
// 	temporalMockClient := &mocks.Client{}
// 	db := test_utils.MigrateCockroachDB(t, ctx)
// 	provider, err := NewService(ServiceArgs{
// 		BaseURL:         "localhost:8080",
// 		Token:           "token",
// 		WebhookToken:    "webhooktoken",
// 		Db:              db,
// 		IdentityService: identityMock,
// 		AccountClient:   accounts_mock.NewMockClient(ctrl),
// 		Logger:          zap.NewNop(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	wh, err := NewWebhook(&WebhookArgs{
// 		Db: db,
// 		Up: provider,
// 		Tp: temporalMockClient,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	customerCreatedEvent := NewCustomerCreatedEvent()
// 	rawEvent := marshalEvent(t, customerCreatedEvent)
// 	testEvent := Event{ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}

// 	_, err = wh.StoreEvent(ctx, testEvent, rawEvent)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	_, err = wh.StoreEvent(ctx, testEvent, rawEvent)

// 	assert.ErrorIs(t, err, ErrDuplicateEvent)
// }

// func marshalBody(t *testing.T, events ...interface{}) *bytes.Buffer {
// 	body := ResponseBody{}
// 	for _, event := range events {
// 		rawEvent, err := json.Marshal(event)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		body.Data = append(body.Data, rawEvent)
// 	}

// 	raw, err := json.Marshal(body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	return bytes.NewBuffer(raw)
// }

// func marshalEvent(t *testing.T, event interface{}) json.RawMessage {
// 	raw, err := json.Marshal(event)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	return raw
// }

// func NewCustomerCreatedEvent() CustomerCreatedEvent {
// 	return CustomerCreatedEvent{
// 		ID:   uuid.NewString(),
// 		Type: "customer.created",
// 		Attributes: EventAttributes{
// 			CreatedAt: "2020-07-29T12:53:05.882Z",
// 			Tags: Tags{
// 				FynbosUserId: uuid.NewString(),
// 			},
// 		},
// 		Relationships: EventRelationships{
// 			Customer: JsonCustomer{
// 				Data: Data{
// 					ID:   "52",
// 					Type: "individualCustomer",
// 				},
// 			},
// 			Application: JsonApplication{
// 				Data: Data{
// 					ID:   "52",
// 					Type: "individualApplication",
// 				},
// 			},
// 		},
// 	}
// }

// func NewApplicationDeniedEvent() ApplicationDeniedEvent {
// 	return ApplicationDeniedEvent{
// 		ID:   uuid.NewString(),
// 		Type: "application.denied",
// 		Attributes: EventAttributes{
// 			CreatedAt: "2020-07-29T12:53:05.882Z",
// 			Tags: Tags{
// 				FynbosUserId: uuid.NewString(),
// 			},
// 		},
// 		Relationships: EventRelationships{
// 			Application: JsonApplication{
// 				Data: Data{
// 					ID:   "52",
// 					Type: "individualApplication",
// 				},
// 			},
// 		},
// 	}
// }
