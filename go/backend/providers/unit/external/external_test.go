package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateApplicationForm(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/application-forms" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		data := ApplicationForm{
			ID:   "411479",
			Type: "applicationForm",
			Attributes: ApplicationFormAttributes{
				Tags: ApplicationTags{
					FynbosUserID: "9fe19d6a-ce2e-4401-85f5-442dec6bf242",
				},
				Url:   "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6",
				Stage: "EnterIndividualInformation",
				ApplicationDetails: ApplicationFormPrefill{
					ApplicationType: "Individual",
					Nationality:     "US",
					Email:           "peter@oscorp.com",
				},
				AllowedApplicationTypes: []string{"Individual"},
			},
		}
		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		formResponse := &Response{
			Data: rawData,
		}
		payload, err := json.Marshal(formResponse)
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

	client := NewClient(server.URL, "test")
	form, err := client.CreateApplicationForm(context.Background(), &CreateApplicationFormArgs{
		ID:      uuid.NewString(),
		Email:   faker.Email(),
		Country: "US",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, form)
	assert.Equal(t, "411479", form.ID)
	assert.Equal(t, "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6", form.Attributes.Url)
}

func TestGetStatementPDF(t *testing.T) {
	t.Parallel()
	pdfContent := []byte("Statement Content PDF")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != fmt.Sprintf(`/statements/%s/pdf`, "411479") {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write(pdfContent)
		if err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})

	client := NewClient(server.URL, "test")
	pdf, err := client.GetStatementPDF(context.Background(), &GetStatementPDFArgs{
		ID: "411479",
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, pdfContent, pdf)
}

func TestGetStatements(t *testing.T) {
	t.Parallel()
	customerID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/statements" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}
		value, present := r.URL.Query()["filter[customerId]"]
		if !present {
			t.Fatal("customerId not found in query string")
		}
		if value[0] != customerID {
			t.Fatal("customerId in query string does not match")
		}

		data := []Statement{
			{
				ID:   "21",
				Type: "statement",
				Attributes: StatementAttributes{
					Period: "2022-07",
				},
				Relationships: StatementRelationships{
					Customer: Relationship{
						Data: TypeData{
							ID:   customerID,
							Type: "customer",
						},
					},
				},
			},
		}

		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		formResponse := &Response{
			Data: rawData,
		}
		payload, err := json.Marshal(formResponse)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(payload))
		if err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})

	client := NewClient(server.URL, "test")
	statements, err := client.GetStatements(context.Background(), customerID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, statements, 1)
	assert.Equal(t, customerID, statements[0].Relationships.Customer.Data.ID)
	assert.Equal(t, "2022-07", statements[0].Attributes.Period)
	assert.Equal(t, "statement", statements[0].Type)
}

func TestGetApplicationForm(t *testing.T) {
	t.Parallel()
	userID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/application-forms" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}
		value, present := r.URL.Query()["filter[tags]"]
		if !present {
			t.Fatal("Query parameter to filter by userId tag is missing.")
		}
		if value[0] != fmt.Sprintf(`{"userId":"%s"}`, userID) {
			t.Fatal("userID in tag does not match the userID specified.")
		}

		data := []ApplicationForm{
			{
				ID:   "411479",
				Type: "applicationForm",
				Attributes: ApplicationFormAttributes{
					Tags: ApplicationTags{
						FynbosUserID: userID,
					},
					Url:   "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6",
					Stage: "EnterIndividualInformation",
					ApplicationDetails: ApplicationFormPrefill{
						ApplicationType: "Individual",
						Nationality:     "US",
						Email:           "peter@oscorp.com",
					},
					AllowedApplicationTypes: []string{"Individual"},
				},
			},
		}
		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		formResponse := &Response{
			Data: rawData,
		}
		payload, err := json.Marshal(formResponse)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(payload))
		if err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})

	client := NewClient(server.URL, "test")
	forms, err := client.FilterApplicationFormsByUserID(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, forms, 1)
	assert.Equal(t, userID, forms[0].Attributes.Tags.FynbosUserID)
	assert.Equal(t, "411479", forms[0].ID)
	assert.Equal(t, "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6", forms[0].Attributes.Url)
}

func TestCreateApplication(t *testing.T) {
	t.Parallel()
	userID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/applications" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		data := Application{
			Type: "individualApplication",
			ID:   "53",
			Attributes: ApplicationAttributes{
				CreatedAt: "2020-01-14T14:05:04.718Z",
				FullName: &FullName{
					First: "Peter",
					Last:  "Parker",
				},
				Ssn: "721074426",
				Address: &Address{
					Street:     "20 Ingram St",
					State:      "NY",
					City:       "Forest Hills",
					PostalCode: "11375",
					Country:    "US",
				},
				DateOfBirth: "2001-08-10",
				Email:       "peter@oscorp.com",
				Phone: &Phone{
					CountryCode: "1",
					Number:      "5555555555",
				},
				Status: "AwaitingDocuments",
				IP:     "127.0.0.1",
				Tags: &ApplicationTags{
					FynbosUserID: userID,
				},
			},
		}
		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		applicationResponse := &Response{
			Data: rawData,
		}
		payload, err := json.Marshal(applicationResponse)
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

	client := NewClient(server.URL, "test")
	application, err := client.CreateApplication(context.Background(), &CreateApplicationArgs{
		UserID:      userID,
		Email:       faker.Email(),
		IpAddress:   "127.0.0.1",
		DateOfBirth: "2001-08-10",
		FirstName:   "Peter",
		LastName:    "Parker",
		Ssn:         "721074426",
		Address: Address{
			Street:     "20 Ingram St",
			State:      "NY",
			City:       "Forest Hills",
			PostalCode: "11375",
			Country:    "US",
		},
		Phone: Phone{
			CountryCode: "1",
			Number:      "5555555555",
		},
		DeviceFingerprints: nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, application)
	assert.Equal(t, "127.0.0.1", application.Attributes.IP)
	assert.Equal(t, userID, application.Attributes.Tags.FynbosUserID)
	assert.Equal(t, "AwaitingDocuments", application.Attributes.Status)
}

func TestCreateCounterparty(t *testing.T) {
	t.Parallel()
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

		data := Counterparty{
			ID:   counterpartyID,
			Type: "achCounterparty",
			Relationships: CounterpartyRelationships{
				Customer: Relationship{
					Data: TypeData{
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

		counterpartyResponse := &Response{
			Data: rawData,
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

	client := NewClient(server.URL, "test")
	counterparty, err := client.CreateCounterparty(context.Background(), &CreateCounterpartyArgs{
		Name:           faker.FirstName(),
		UnitCustomerID: unitCustomerID,
		RoutingNumber:  faker.CCNumber(),
		AccountNumber:  faker.CCNumber(),
		AccountType:    faker.CCType(),
		Type:           "person",
		IdempotencyKey: "test-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, counterparty)
	assert.Equal(t, unitCustomerID, counterparty.Relationships.Customer.Data.ID)
	assert.Equal(t, counterpartyID, counterparty.ID)
}

func TestOriginateAch(t *testing.T) {
	t.Parallel()
	depositID := uuid.NewString()
	achPaymentID := uuid.NewString()
	unitCustomerID := uuid.NewString()
	unitAccountID := uuid.NewString()
	counterpartyID := uuid.NewString()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/payments" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		data := AchPayment{
			ID:   achPaymentID,
			Type: "achPayment",
			Attributes: AchPaymentAttributes{
				CreatedAt:   "2020-01-13T16:01:19.346Z",
				Status:      AchStatusPending,
				Description: "Funding",
				Direction:   "Debit",
				Amount:      10000,
				Tags: DepositTags{
					DepositID: depositID,
				},
			},
			Relationships: AchPaymentRelationships{
				Customer: &Relationship{
					Data: TypeData{
						ID:   unitCustomerID,
						Type: "customer",
					},
				},
				Counterparty: Relationship{
					Data: TypeData{
						ID:   counterpartyID,
						Type: "counterparty",
					},
				},
				Account: Relationship{
					Data: TypeData{
						ID:   unitAccountID,
						Type: "depositAccount",
					},
				},
			},
		}
		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		response := &Response{
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

	client := NewClient(server.URL, "test")
	ach, err := client.OriginateAch(context.Background(), &OriginateAchArgs{
		IdempotencyKey:   "test-key",
		Amount:           10000,
		Direction:        "Debit",
		CounterpartyID:   counterpartyID,
		DepositAccountID: unitAccountID,
		Description:      "Funding",
		Tags: map[string]string{
			"DepositID": depositID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, ach)
	assert.Equal(t, depositID, ach.Attributes.Tags.DepositID)
	assert.Equal(t, unitCustomerID, ach.Relationships.Customer.Data.ID)
	assert.Equal(t, counterpartyID, ach.Relationships.Counterparty.Data.ID)
	assert.Equal(t, unitAccountID, ach.Relationships.Account.Data.ID)
}

func TestCreateDepositAccount(t *testing.T) {
	t.Parallel()
	depositAccountID := uuid.NewString()
	args := &CreateDepositAccountArgs{
		CustomerID:     uuid.NewString(),
		DepositProduct: "checking",
		Type:           "depositAccount",
		IdempotencyKey: "test",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/accounts" {
			http.Error(w, "Not found.", http.StatusNotFound)
			return
		}

		data := DepositAccount{
			ID:   depositAccountID,
			Type: "depositAccount",
			Attributes: DepositAccountAttributes{
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
			Relationships: &DepositAccountRelationships{
				Customer: Relationship{
					Data: TypeData{
						ID:   args.CustomerID,
						Type: "customer",
					},
				},
			},
		}
		rawData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		depositAccountResponse := &Response{
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

	client := NewClient(server.URL, "test")
	depositAccount, err := client.CreateDepositAccount(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, depositAccount)
	assert.NotNil(t, depositAccount.Relationships)
	assert.Equal(t, args.CustomerID, depositAccount.Relationships.Customer.Data.ID)
	assert.Equal(t, depositAccountID, depositAccount.ID)
	assert.Equal(t, "USD", depositAccount.Attributes.Currency)
	assert.Equal(t, int64(10000), depositAccount.Attributes.BalanceInCents)
	assert.Equal(t, int64(1000), depositAccount.Attributes.HoldInCents)
	assert.Equal(t, int64(9000), depositAccount.Attributes.AvailableInCents)
}
