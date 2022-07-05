package unit

//go:generate mockgen -destination=./mock.go -package=unit -source=./service.go

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/nyaruka/phonenumbers"
	"gitlab.com/fynbos/backend/identity"
)

var (
	ErrInternal        = errors.New("unit: internal error")
	ErrUnauthorized    = errors.New("unit: unauthorized webhook request")
	ErrInvalidArgument = errors.New("unit: invalid argument.")
	ErrNotFound        = errors.New("unit: not found.")
)

const (
	ApplicationUserIDTag = "fynbosUserId"
	SignatureHeader      = "x-unit-signature"
)

type Service interface {
	GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error)
	CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
	CreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error)
	VerifyWebhook(ctx context.Context, body []byte, signature string) error
	CreateCustomer(ctx context.Context, args *CreateCustomerArgs) (*Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*Customer, error)
	GetCustomerByAccountID(ctx context.Context, accountID string) (*Customer, error)
	CreateCounterParty(ctx context.Context, args *CreateCounterPartyArgs) (*CounterParty, error)
	GetCounterPartyByFundingsourceID(ctx context.Context, fundingsourceID string) (*CounterParty, error)
}

type (
	service struct {
		validator       *validator.Validate
		baseURL         string
		token           string
		webhookToken    string
		db              *sqlx.DB
		identityService identity.Service
	}

	ServiceArgs struct {
		BaseURL         string           `validate:"required"`
		Token           string           `validate:"required"`
		WebhookToken    string           `validate:"required"`
		Db              *sqlx.DB         `validate:"required"`
		IdentityService identity.Service `validate:"required"`
	}
)

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator:       validator,
		baseURL:         args.BaseURL,
		token:           args.Token,
		webhookToken:    args.WebhookToken,
		db:              args.Db,
		identityService: args.IdentityService,
	}, nil
}

type ApplicationForm struct {
	ID  string
	URL string
}

func (self *service) GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error) {

	url := fmt.Sprintf(`%s/application-forms?page[limit]=1&filter[tags]={"userId":"%s"}`, self.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", self.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Data []struct {
			ID   string `json:"id"`
			Attr struct {
				Url string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	if len(data.Data) == 0 {
		return nil, nil
	}

	return &ApplicationForm{
		ID:  data.Data[0].ID,
		URL: data.Data[0].Attr.Url,
	}, nil
}

type CreateApplicationFormArgs struct {
	ID      string `validate:"required"`
	Email   string `validate:"required"`
	Country string `validate:"required"`
}

func (self *service) CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error) {

	url := fmt.Sprintf(`%s/application-forms`, self.baseURL)

	var jsonStr = []byte(fmt.Sprintf(`{
		"data": {
			"type": "applicationForm",
			"attributes": {
				"tags": {
					"%s": "%s"
				},
				"allowedApplicationTypes": ["Individual"],
				"applicantDetails": {
					"applicationType": "Individual",
					"nationality": "%s",
					"email": "%s"
				}
			}
		}
	}`, ApplicationUserIDTag, args.ID, args.Country, args.Email))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", self.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Data struct {
			ID   string `json:"id"`
			Attr struct {
				Url string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &ApplicationForm{
		ID:  data.Data.ID,
		URL: data.Data.Attr.Url,
	}, nil
}

type CreateApplicationArgs struct {
	Ssn                string   `validate:"required"`
	DateOfBirth        string   `validate:"required"`
	Street             string   `validate:"required"`
	Street2            string   `validate:"required"`
	City               string   `validate:"required"`
	State              string   `validate:"required"`
	PostalCode         string   `validate:"required"`
	IpAddress          string   `validate:"required"`
	UserID             string   `validate:"required"`
	DeviceFingerprints []string `validate:"required"`
}

type Application struct {
	Type         string
	ID           string
	Status       string
	FynbosUserId string
	Archived     bool
	CustomerID   string
}

func (s *service) CreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error) {
	identity, err := s.identityService.Get(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	phoneNumber, err := phonenumbers.Parse(identity.MobileNumber, "")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	// https://docs.unit.co/applications/#create-individual-application
	url := fmt.Sprintf(`%s/applications`, s.baseURL)

	// deviceFingerprints := make([]DeviceFingerprint, len(args.DeviceFingerprints))
	// for _, fingerprint := range args.DeviceFingerprints {
	// 	deviceFingerprints = append(deviceFingerprints, DeviceFingerprint{
	// 		Provider: "iovation",
	// 		Value:    fingerprint,
	// 	})
	// }

	application := &CreateApplicationRequest{
		Data: CreateApplicationRequestData{
			Type: "individualApplication",
			Attributes: RequestApplicationAttributes{
				Ssn: args.Ssn,
				FullName: FullName{
					First: identity.FirstName,
					Last:  identity.LastName,
				},
				DateOfBirth: args.DateOfBirth,
				Address: Address{
					Street:     args.Street,
					Street2:    args.Street2,
					City:       args.City,
					State:      args.State,
					PostalCode: args.PostalCode,
					Country:    identity.Country,
				},
				Email: identity.Email,
				Phone: Phone{
					CountryCode: strconv.Itoa(int(*phoneNumber.CountryCode)),
					Number:      strconv.Itoa(int(*phoneNumber.NationalNumber)),
				},
				IP: args.IpAddress,
				Tags: Tags{
					FynbosUserId: args.UserID,
				},
				// DeviceFingerprints: deviceFingerprints,
			},
		},
	}

	rawApplication, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawApplication))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data CreateApplicationResponse

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &Application{
		Type:         data.Data.Type,
		ID:           data.Data.ID,
		Status:       data.Data.Attributes.Status,
		FynbosUserId: data.Data.Attributes.Tags.FynbosUserId,
		Archived:     data.Data.Attributes.Archived,
		CustomerID:   data.Data.Relationships.Customer.Data.ID,
	}, nil
}

func (self *service) VerifyWebhook(ctx context.Context, body []byte, signature string) error {
	mac := hmac.New(sha1.New, []byte(self.webhookToken))
	mac.Write(body)
	sha := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if sha != signature {
		return ErrUnauthorized
	}

	return nil
}

// maps the unit customer to the Fynbos account
type (
	Customer struct {
		ID        string `db:"id"`
		AccountID string `db:"account_id"`
		Type      string `db:"type"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	CreateCustomerArgs struct {
		ID        string
		AccountID string
		Type      string
	}
)

func (s *service) CreateCustomer(
	ctx context.Context,
	args *CreateCustomerArgs,
) (*Customer, error) {
	var customer Customer
	err := s.db.GetContext(
		ctx,
		&customer,
		"INSERT INTO unit_customers (id, account_id, type) VALUES ($1, $2, $3) RETURNING *;",
		args.ID,
		args.AccountID,
		args.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &customer, nil
}

func (s *service) GetCustomerByID(ctx context.Context, id string) (*Customer, error) {
	var customer Customer
	if err := s.db.GetContext(ctx, &customer, "SELECT * FROM unit_customers WHERE id=$1", id); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &customer, nil
}

func (s *service) GetCustomerByAccountID(ctx context.Context, accountID string) (*Customer, error) {
	var customer Customer
	if err := s.db.GetContext(ctx, &customer, "SELECT * FROM unit_customers WHERE account_id=$1", accountID); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &customer, nil
}

type (
	CreateCounterPartyArgs struct {
		FundingsourceID string `validate:"uuid4"`
		Name            string `validate:"required"`
		UnitCustomerID  string `validate:"required"`
		RoutingNumber   string `validate:"required"`
		AccountNumber   string `validate:"required"`
		AccountType     string `validate:"required"`
		Type            string `validate:"required"`

		// This idempotency key is valid for 48 hours on Unit's api.
		IdempotencyKey string `validate:"lte=255"`
	}

	CounterParty struct {
		ID              string
		FundingsourceID string `db:"fundingsource_id"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}
)

// This will create the counter party on Unit and store a record of it in our database.
func (s *service) CreateCounterParty(ctx context.Context, args *CreateCounterPartyArgs) (*CounterParty, error) {
	url := fmt.Sprintf(`%s/counterparties`, s.baseURL)
	var jsonStr = []byte(fmt.Sprintf(`{
		"data": {
			"type": "achCounterparty",
			"attributes": {
				"name": %s,
				"routingNumber": %s,
				"accountNumber": %s,
				"accountType": %s,
				"type": %s,
				"idempotencyKey": %s
			},
			"relationships": {
				"customer": {
					"data": {
						"type": "customer",
						"id": %s
					}
				}
			}
		}
	}`, args.Name, args.RoutingNumber, args.AccountNumber, args.AccountType, args.IdempotencyKey, args.Type, args.UnitCustomerID))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	createCounterPartyResponse := &struct {
		Data struct {
			Type string
			ID   string `json:"id"`
		}
	}{}
	err = json.Unmarshal(body, createCounterPartyResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	ret := &CounterParty{}
	err = s.db.GetContext(
		ctx,
		ret,
		"INSERT INTO unit_counterparties (id, fundingsource_id) VALUES ($1, $2) RETURNING *;",
		createCounterPartyResponse.Data.ID,
		args.FundingsourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	return ret, nil
}

func (s service) GetCounterPartyByFundingsourceID(ctx context.Context, fundingsourceID string) (*CounterParty, error) {
	ret := &CounterParty{}
	err := s.db.GetContext(
		ctx,
		ret,
		"SELECT * FROM unit_counterparties WHERE fundingsource_id=$1 LIMIT 1;",
		fundingsourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}
