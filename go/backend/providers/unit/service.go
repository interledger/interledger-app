package unit

//go:generate mockgen -destination=./mock.go -package=unit -source=./service.go

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/nyaruka/phonenumbers"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/unit/external"
)

var (
	ErrInternal        = errors.New("unit: internal error")
	ErrUnauthorized    = errors.New("unit: unauthorized webhook request")
	ErrInvalidArgument = errors.New("unit: invalid argument")
	ErrClient          = errors.New("unit: client error")
	ErrServer          = errors.New("unit: server error")
	ErrTimeout         = errors.New("unit: timeout error")
	ErrRateLimit       = errors.New("unit: rate limit error")
	ErrNotFound        = errors.New("unit: not found")
)

const (
	SignatureHeader = "x-unit-signature"
)

type Service interface {
	// This function is idempotent for 48 hours. An idempotency key is generated based on the
	// customerID when calling out to unit. The mapping is stored in our database.
	CreateDepositAccount(ctx context.Context, customerID string) (*DepositAccount, error)
	GetDepositAccount(ctx context.Context, id string) (*DepositAccount, error)
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
		externalClient  external.Unit
		webhookToken    string
		db              *sqlx.DB
		identityService identity.Service
		accountService  accounts.Service
		logger          *zap.Logger
	}

	ServiceArgs struct {
		BaseURL         string           `validate:"required"`
		Token           string           `validate:"required"`
		WebhookToken    string           `validate:"required"`
		Db              *sqlx.DB         `validate:"required"`
		IdentityService identity.Service `validate:"required"`
		AccountService  accounts.Service `validate:"required"`
		Logger          *zap.Logger      `validate:"required"`
	}
)

func NewService(args ServiceArgs) (Service, error) {
	v := validator.New()
	err := v.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator:       v,
		webhookToken:    args.WebhookToken,
		db:              args.Db,
		identityService: args.IdentityService,
		accountService:  args.AccountService,
		externalClient:  external.NewClient(args.BaseURL, args.Token),
		logger:          args.Logger.With(zap.String("service", "unit")),
	}, nil
}

type ApplicationForm struct {
	ID  string
	URL string
}

func (s *service) GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error) {
	forms, err := s.externalClient.FilterApplicationFormsByUserID(ctx, userID)
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	if len(forms) != 1 {
		return nil, ErrNotFound
	}

	return &ApplicationForm{
		ID:  forms[0].ID,
		URL: forms[0].Attributes.Url,
	}, nil
}

type CreateApplicationFormArgs struct {
	ID      string `validate:"required"`
	Email   string `validate:"required"`
	Country string `validate:"required"`
}

func (s *service) CreateApplicationForm(
	ctx context.Context,
	args *CreateApplicationFormArgs,
) (*ApplicationForm, error) {
	form, err := s.externalClient.CreateApplicationForm(ctx, &external.CreateApplicationFormArgs{
		ID:      args.ID,
		Email:   args.Email,
		Country: args.Country,
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ApplicationForm{
		ID:  form.ID,
		URL: form.Attributes.Url,
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

func (s *service) CreateApplication(ctx context.Context, args *CreateApplicationArgs) (application *Application, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"failed to create unit application",
				zap.String("userId", args.UserID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"created unit application",
			zap.String("userId", args.UserID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	id, err := s.identityService.Get(ctx, args.UserID)
	if err != nil {
		err = fmt.Errorf("%w %s", ErrInternal, err)
		return
	}

	phoneNumber, err := phonenumbers.Parse(id.MobileNumber, "")
	if err != nil {
		err = fmt.Errorf("%w %s", ErrInternal, err)
		return
	}

	// deviceFingerprints := make([]DeviceFingerprint, len(args.DeviceFingerprints))
	// for _, fingerprint := range args.DeviceFingerprints {
	// 	deviceFingerprints = append(deviceFingerprints, DeviceFingerprint{
	// 		Provider: "iovation",
	// 		Value:    fingerprint,
	// 	})
	// }

	app, err := s.externalClient.CreateApplication(ctx, &external.CreateApplicationArgs{
		UserID:             args.UserID,
		Email:              id.Email,
		IpAddress:          args.IpAddress,
		DateOfBirth:        args.DateOfBirth,
		DeviceFingerprints: nil, // TODO: figure out why this isn't working
		FirstName:          id.FirstName,
		LastName:           id.LastName,
		Ssn:                args.Ssn,
		Address: external.Address{
			Street:     args.Street,
			Street2:    args.Street2,
			City:       args.City,
			State:      args.State,
			PostalCode: args.PostalCode,
			Country:    id.Country,
		},
		Phone: external.Phone{
			CountryCode: strconv.Itoa(int(*phoneNumber.CountryCode)),
			Number:      strconv.Itoa(int(*phoneNumber.NationalNumber)),
		},
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			err = fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
			return
		}
		if errors.Is(err, external.ErrRequest) {
			err = fmt.Errorf("%w %s", ErrClient, err)
			return
		}
		err = fmt.Errorf("%w %s", ErrInternal, err)
		return
	}

	return &Application{
		Type:         app.Type,
		ID:           app.ID,
		Status:       app.Attributes.Status,
		FynbosUserId: app.Attributes.Tags.FynbosUserID,
		Archived:     app.Attributes.Archived,
		CustomerID:   app.Relationships.Customer.Data.ID,
	}, nil
}

func (s *service) VerifyWebhook(ctx context.Context, body []byte, signature string) error {
	mac := hmac.New(sha1.New, []byte(s.webhookToken))
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
	counterparty, err := s.externalClient.CreateCounterparty(ctx, &external.CreateCounterpartyArgs{
		Name:           args.Name,
		UnitCustomerID: args.UnitCustomerID,
		RoutingNumber:  args.RoutingNumber,
		AccountNumber:  args.AccountNumber,
		AccountType:    args.AccountType,
		Type:           args.Type,
		IdempotencyKey: args.IdempotencyKey,
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	ret := &CounterParty{}
	err = s.db.GetContext(
		ctx,
		ret,
		"INSERT INTO unit_counterparties (id, fundingsource_id) VALUES ($1, $2) RETURNING *;",
		counterparty.ID,
		args.FundingsourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	return ret, nil
}

func (s *service) GetCounterPartyByFundingsourceID(ctx context.Context, fundingsourceID string) (*CounterParty, error) {
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

// maps the unit deposit account and unit customerID
type DepositAccount struct {
	ID         string `db:"id"`
	CustomerID string `db:"customer_id"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

func (s *service) CreateDepositAccount(
	ctx context.Context,
	customerID string,
) (*DepositAccount, error) {
	if customerID == "" {
		return nil, fmt.Errorf("%w CustomerID is required", ErrInvalidArgument)
	}

	idempotencyKey := sha256.Sum256([]byte(customerID))
	depositAccount, err := s.externalClient.CreateDepositAccount(ctx, &external.CreateDepositAccountArgs{
		CustomerID:     customerID,
		DepositProduct: "checking",
		IdempotencyKey: string(idempotencyKey[0:]),
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var ret DepositAccount
	err = s.db.GetContext(
		ctx,
		&ret,
		"INSERT INTO unit_deposit_accounts (id, customer_id) VALUES ($1, $2) RETURNING *;",
		depositAccount.ID,
		customerID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ret, nil
}

func (s *service) GetDepositAccount(ctx context.Context, id string) (*DepositAccount, error) {
	var ret DepositAccount
	err := s.db.GetContext(
		ctx,
		&ret,
		"SELECT * FROM unit_deposit_accounts WHERE id=$1;",
		id,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ret, nil
}

func IsRetryableError(err error) bool {
	return errors.Is(err, ErrClient) ||
		errors.Is(err, ErrRateLimit) ||
		errors.Is(err, ErrServer) ||
		errors.Is(err, ErrTimeout)
}

func statusToError(statusCode int) error {
	if statusCode == http.StatusRequestTimeout {
		return ErrTimeout
	}
	if statusCode == http.StatusTooManyRequests {
		return ErrRateLimit
	}
	if statusCode > http.StatusInternalServerError {
		return ErrServer
	} else {
		return ErrInternal
	}
}
