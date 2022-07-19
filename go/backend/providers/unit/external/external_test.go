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

		formResponse := &ApplicationFormRequest{
			Data: ApplicationForm{
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
			},
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

		formResponse := &ListApplicationFormRequest{
			Data: []ApplicationForm{
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
			},
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

		applicationResponse := &ApplicationResponse{
			Data: Application{
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
			},
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

		counterpartyResponse := &CounterpartyRequest{
			Data: Counterparty{
				ID:   counterpartyID,
				Type: "achCounterparty",
				Relationships: CounterpartyRelationships{
					Customer: Customer{
						Data: TypeData{
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
