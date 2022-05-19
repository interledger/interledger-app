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

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

var (
	ErrInternal     = errors.New("unit: internal error")
	ErrUnauthorized = errors.New("unit: unauthorized webhook request")
)

const (
	ApplicationFormUserIDTag = "userID"
	SignatureHeader          = "x-unit-signature"
)

type Service interface {
	GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error)
	CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
	VerifyWebhook(ctx context.Context, request *http.Request) error
	CreateCustomer(ctx context.Context, args *CreateCustomerArgs) (*Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*Customer, error)
	GetCustomerByAccountID(ctx context.Context, accountID string) (*Customer, error)
}

type (
	service struct {
		validator    *validator.Validate
		baseURL      string
		token        string
		webhookToken string
		db           *sqlx.DB
	}

	ServiceArgs struct {
		BaseURL      string   `validate:"required"`
		Token        string   `validate:"required"`
		WebhookToken string   `validate:"required"`
		Db           *sqlx.DB `validate:"required"`
	}
)

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator:    validator,
		baseURL:      args.BaseURL,
		token:        args.Token,
		webhookToken: args.WebhookToken,
		db:           args.Db,
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
	}`, ApplicationFormUserIDTag, args.ID, args.Country, args.Email))

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

func (self *service) VerifyWebhook(ctx context.Context, request *http.Request) error {
	signature := request.Header.Get(SignatureHeader)
	if signature == "" {
		return ErrInternal
	}

	mac := hmac.New(sha1.New, []byte(self.webhookToken))

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return ErrInternal
	}

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
