package rafiki

//go:generate mockgen -destination=./mock.go -package=rafiki -source=./service.go
//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	_validator "github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

var (
	ErrInvalidArgument = errors.New("rafiki provider: invalid argument.")
	ErrInternal        = errors.New("rafiki provider: internal error.")
	ErrNotFound        = errors.New("rafiki provider: not found.")
)

type (
	ServiceArgs struct {
		Db  *sqlx.DB `validate:"required"`
		Url string   `validate:"required"`
	}

	Service interface {
		CreateIdentifier(ctx context.Context, args *CreateIdentifierArgs) (*Identifier, error)
		GetIdentifier(ctx context.Context, id string) (*Identifier, error)
		GetIdentifierByAccountAndCurrency(
			ctx context.Context,
			accountID string,
			currencyCode string,
		) (*Identifier, error)
		CreateQuote(ctx context.Context, args *CreateQuoteArgs) (*Quote, error)
	}

	service struct {
		validator     *_validator.Validate
		db            *sqlx.DB
		graphqlClient graphql.Client
	}

	Identifier struct {
		ID         string `db:"id"`
		AccountID  string `db:"account_id"`
		AssetCode  string `db:"asset_code"`
		AssetScale uint8  `db:"asset_scale"`
		CreatedAt  string `db:"created_at"`
		UpdatedAt  string `db:"updated_at"`
	}

	Quote struct {
		ID           string `db:"id"`
		IdentifierID string `db:"identifier_id"`
		ExpiresAt    string `db:"expired_at"`

		// Address where payment is to be made.
		Receiver                  string  `db:"receiver"`
		SendAssetCode             string  `db:"send_asset_code"`
		SendAssetScale            uint8   `db:"send_asset_scale"`
		SendAmount                uint64  `db:"send_amount"`
		ReceiveAmount             uint64  `db:"receive_amount"`
		ReceiveAssetCode          string  `db:"receive_asset_code"`
		ReceiveAssetScale         uint8   `db:"receive_asset_scale"`
		MinExchangeRate           float64 `db:"min_exchange_rate"`
		LowEstimatedExchangeRate  float64 `db:"low_estimated_exchange_rate"`
		HighEstimatedExchangeRate float64 `db:"high_estimated_exchange_rate"`
	}
)

func NewService(args *ServiceArgs) (Service, error) {
	validator := _validator.New()
	if err := validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &service{
		validator:     validator,
		db:            args.Db,
		graphqlClient: graphql.NewClient(args.Url, http.DefaultClient),
	}, nil
}

type CreateIdentifierArgs struct {
	AccountID  string `validate:"uuid4"`
	AssetCode  string `validate:"required"`
	AssetScale uint8  `validate:"required"`
}

// This will call out to Rafiki to create an identifier and then map it to the specified account id.
func (s *service) CreateIdentifier(ctx context.Context, args *CreateIdentifierArgs) (*Identifier, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	response, err := CreateAccount(ctx, s.graphqlClient, CreateAccountInput{
		Asset: AssetInput{
			Code:  args.AssetCode,
			Scale: int(args.AssetScale),
		},
		// TODO: public name
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !response.CreateAccount.Success {
		return nil, fmt.Errorf("%w %s", ErrInternal, response.CreateAccount.Message)
	}
	account := response.CreateAccount.Account

	ret := &Identifier{}
	err = s.db.Get(
		ret,
		`
		INSERT INTO rafiki_identifiers (id, account_id, asset_code, asset_scale)
		VALUES ($1, $2, $3, $4) RETURNING *;`,
		account.Id,
		args.AccountID,
		args.AssetCode,
		args.AssetScale,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (s *service) GetIdentifier(ctx context.Context, id string) (*Identifier, error) {
	ret := &Identifier{}
	err := s.db.Get(
		ret,
		"SELECT * FROM rafiki_identifiers WHERE id=$1;",
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

// This will return the most recently created Identifier that matches the accountID and currencyCode.
func (s *service) GetIdentifierByAccountAndCurrency(
	ctx context.Context,
	accountID string,
	currencyCode string,
) (*Identifier, error) {
	ret := &Identifier{}
	err := s.db.Get(
		ret,
		"SELECT * FROM rafiki_identifiers WHERE account_id=$1 AND asset_code=$2 ORDER BY created_at DESC LIMIT 1;",
		accountID,
		currencyCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

type CreateQuoteArgs struct {
	IdentifierID           string `validate:"required"`
	ReceiverPaymentPointer string `validate:"required"`
	SendAssetCode          string `validate:"required_with=SendAmount"`
	SendAssetScale         uint8  `validate:"required_with=SendAmount"`
	SendAmount             uint64 `validate:"required_if=ReceiveAmount 0"`
	ReceiveAssetCode       string `validate:"required_with=ReceiveAmount"`
	ReceiveAssetScale      uint8  `validate:"required_with=ReceiveAmount"`
	ReceiveAmount          uint64 `validate:"required_if=SendAmount 0"`
}

// Creates a quote to be sent from the specified `IdentifierID` to the `ReceiverPaymentPointer`.
// The `AssetCode` and `AssetScale` describe the currency in which the payment will be made. Either
// `SendAmount`
// or `ReceiveAmount` must be specified.
//
// `SendAmount` specifies what will leave the account attached to the `Identifier`.
//
// `ReceiveAmount` specifies what the receiver will get.
func (s service) CreateQuote(ctx context.Context, args *CreateQuoteArgs) (*Quote, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	input := CreateQuoteInput{
		// Our identifier is mapped to Rafiki's internal account.
		AccountId: args.IdentifierID,
		Receiver:  args.ReceiverPaymentPointer,
	}

	if args.SendAmount != 0 {
		input.SendAmount = &AmountInput{
			Value:      args.SendAmount,
			AssetCode:  args.SendAssetCode,
			AssetScale: int(args.SendAssetScale),
		}
	}
	if args.ReceiveAmount != 0 {
		input.ReceiveAmount = &AmountInput{
			Value:      args.ReceiveAmount,
			AssetCode:  args.ReceiveAssetCode,
			AssetScale: int(args.ReceiveAssetScale),
		}
	}
	response, err := CreateQuote(ctx, s.graphqlClient, input)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !response.CreateQuote.Success {
		return nil, fmt.Errorf("%w %s", ErrInternal, response.CreateQuote.Message)
	}
	quote := response.CreateQuote.Quote

	return &Quote{
		ID:                        quote.Id,
		IdentifierID:              quote.AccountId,
		ExpiresAt:                 quote.ExpiresAt,
		Receiver:                  quote.Receiver,
		SendAssetCode:             quote.SendAmount.AssetCode,
		SendAssetScale:            uint8(quote.SendAmount.AssetScale),
		SendAmount:                quote.SendAmount.Value,
		ReceiveAssetCode:          quote.ReceiveAmount.AssetCode,
		ReceiveAssetScale:         uint8(quote.ReceiveAmount.AssetScale),
		ReceiveAmount:             quote.ReceiveAmount.Value,
		MinExchangeRate:           quote.MinExchangeRate,
		LowEstimatedExchangeRate:  quote.LowEstimatedExchangeRate,
		HighEstimatedExchangeRate: quote.HighEstimatedExchangeRate,
	}, nil
}
