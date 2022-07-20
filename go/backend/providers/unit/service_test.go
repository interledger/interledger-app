package unit

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/unit/external"
	test_utils "gitlab.com/fynbos/backend/utils"
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
		IdentityService: identity.NewMockService(ctrl),
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
	unitService, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         "localhost",
		Token:           "test token",
		Db:              test_utils.MigrateCockroachDB(t, ctx),
		IdentityService: identity.NewMockService(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}

	accountID := uuid.NewString()
	customerID := uuid.NewString()
	customerType := "individual"
	customer, err := unitService.CreateCustomer(ctx, &CreateCustomerArgs{
		ID:        customerID,
		AccountID: accountID,
		Type:      customerType,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customer.ID)
	assert.Equal(t, accountID, customer.AccountID)
	assert.Equal(t, customerType, customer.Type)

	customerByID, err := unitService.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customerByID.ID)
	assert.Equal(t, accountID, customerByID.AccountID)
	assert.Equal(t, customerType, customerByID.Type)

	customerByAccountID, err := unitService.GetCustomerByAccountID(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, customerID, customerByAccountID.ID)
	assert.Equal(t, customerID, customerByAccountID.ID)
	assert.Equal(t, customerType, customerByAccountID.Type)
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
					Customer: external.Customer{
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
		IdentityService: identity.NewMockService(ctrl),
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
