package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/backend/accounts"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/identity"
	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"
	"gitlab.com/fynbos/backend/providers/unit/external"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

func TestVerifyWebhook(s *testing.T) {
	s.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(s)
	unitService, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         "localhost",
		Token:           "test token",
		Db:              &sqlx.DB{},
		IdentityService: identity_mock.NewMockClient(ctrl),
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		s.Fatal(err)
	}

	s.Run("Successfully verifies incoming webhook request", func(t *testing.T) {
		body := []byte(`{"data":[{"id":"2504140","type":"customer.created","attributes":{"createdAt":"2022-05-18T14:35:00.702Z","tags":{"userID":"02242b61-a99e-4b44-bda7-cf6a4f535a5f","test":"webhook-tag","key":"another-tag","number":"111"}},"relationships":{"customer":{"data":{"id":"344063","type":"individualCustomer"}},"application":{"data":{"id":"404728","type":"individualApplication"}}}}]}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg=" // key = fynbos_local_unit_webhook_token

		err := unitService.VerifyWebhook(ctx, body, signature)

		assert.NoError(t, err)
	})

	s.Run("Fails if signature of webhook body does not match provided one", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg"

		err := unitService.VerifyWebhook(ctx, body, signature)

		assert.ErrorIs(t, err, ErrUnauthorized)
	})
}

func TestCreateAndGetCustomer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockIdentityService := identity_mock.NewMockClient(ctrl)
	unitService, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         "localhost",
		Token:           "test token",
		Db:              test_utils.MigrateCockroachDB(t, ctx),
		IdentityService: mockIdentityService,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}
	identityID := uuid.NewString()
	mockIdentityService.EXPECT().Get(context.Background(), identityID).
		Return(
			&identity.Identity{ID: identityID},
			nil,
		)

	customerID := uuid.NewString()
	customerType := "individual"
	customer, err := unitService.CreateCustomer(ctx, &CreateCustomerArgs{
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

	customerByID, err := unitService.GetCustomer(ctx, customer.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customerByID.ID)
	assert.Equal(t, identityID, customerByID.IdentityID)
	assert.Equal(t, customerType, customerByID.Type)

	customerByIdentityID, err := unitService.GetCustomerByIdentityID(ctx, identityID)
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
	unitCustomerID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/counterparties" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		counterpartyResponse := &external.CounterpartyRequest{
			Data: external.Counterparty{
				ID:   counterpartyID,
				Type: "achCounterparty",
				Relationships: external.CounterpartyRelationships{
					Customer: external.Relationship{
						Data: external.TypeData{
							ID:   unitCustomerID,
							Type: "customer",
						},
					},
				},
			},
		}
		payload, err := json.Marshal(counterpartyResponse)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(payload))
		if err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})
	unitService, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         server.URL,
		Token:           "test token",
		Db:              test_utils.MigrateCockroachDB(t, ctx),
		IdentityService: identity_mock.NewMockClient(ctrl),
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}

	args := &CreateCounterPartyArgs{
		FundingsourceID: uuid.NewString(),
		Name:            "test name",
		UnitCustomerID:  uuid.NewString(),
		RoutingNumber:   faker.CCNumber(),
		AccountNumber:   faker.CCNumber(),
		AccountType:     faker.CCType(),
		Type:            "person",
		IdempotencyKey:  "test",
	}

	unitCounterparty, err := unitService.CreateCounterParty(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.FundingsourceID, args.FundingsourceID)
	assert.NotEqual(t, "", unitCounterparty.ID)

	freshCounterParty, err := unitService.GetCounterPartyByFundingsourceID(ctx, args.FundingsourceID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.FundingsourceID, freshCounterParty.FundingsourceID)
	assert.NotEqual(t, "", freshCounterParty.ID)
}

func TestCreateAndGetDepositAccount(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	unitCustomerID := uuid.NewString()
	unitDepositAccountID := uuid.NewString()
	mockAccountService := accounts_mock.NewMockClient(ctrl)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/accounts" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		data := external.DepositAccount{
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
		}
		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		depositAccountResponse := &external.Response{
			Data: rawData,
		}
		payload, err := json.Marshal(depositAccountResponse)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(payload))
		if err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})
	unitService, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         server.URL,
		Token:           "test token",
		Db:              test_utils.MigrateCockroachDB(t, ctx),
		IdentityService: identity_mock.NewMockClient(ctrl),
		AccountClient:   mockAccountService,
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}

	createdAcc, err := unitService.CreateDepositAccount(ctx, unitCustomerID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, unitCustomerID, createdAcc.CustomerID)

	freshAcc, err := unitService.GetDepositAccount(ctx, unitDepositAccountID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, unitCustomerID, freshAcc.CustomerID)
}

func TestInitiateUserDeposit(t *testing.T) {
	t.Parallel()
	userID := uuid.NewString()
	accountID := uuid.NewString()
	depositID := uuid.NewString()
	depositAccountID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	achPaymentID := uuid.NewString()
	counterpartyID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		var data any
		if r.URL.Path == "/payments" {
			data = external.AchPayment{
				ID:   achPaymentID,
				Type: "achPayment",
			}
		} else if r.URL.Path == "/counterparties" {
			data = external.Counterparty{
				Type: "counterparty",
				ID:   counterpartyID,
			}
		} else {
			http.NotFound(w, r)
			return
		}

		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		response := &external.Response{
			Data: rawData,
		}
		payload, err := json.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(payload))
		if err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})
	ctrl := gomock.NewController(t)
	mockAccounts := accounts_mock.NewMockClient(ctrl)
	mockIdentity := identity_mock.NewMockClient(ctrl)
	unitService, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         server.URL,
		Token:           "test token",
		Db:              test_utils.MigrateCockroachDB(t, context.Background()),
		IdentityService: mockIdentity,
		Logger:          zap.NewNop(),
		AccountClient:   mockAccounts,
	})
	if err != nil {
		t.Fatal(err)
	}

	mockIdentity.EXPECT().Get(context.Background(), userID).Return(&identity.Identity{ID: userID}, nil)
	customer, err := unitService.CreateCustomer(context.Background(), &CreateCustomerArgs{
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

	_, err = unitService.CreateCounterParty(context.Background(), &CreateCounterPartyArgs{
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

	achDeposit, err := unitService.InitiateUserDeposit(context.Background(), &InitiateUserDepositArgs{
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
