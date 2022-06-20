package rafiki

//go:generate mockgen -destination=./mock.go -package=rafiki -source=./service.go

import (
	"context"
	"errors"
	"fmt"

	_validator "github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/machinebox/graphql"
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
	}

	service struct {
		validator     *_validator.Validate
		db            *sqlx.DB
		graphqlClient *graphql.Client
	}

	Identifier struct {
		ID         string `db:"id"`
		AccountID  string `db:"account_id"`
		AssetCode  string `db:"asset_code"`
		AssetScale uint8  `db:"asset_scale"`
		CreatedAt  string `db:"created_at"`
		UpdatedAt  string `db:"updated_at"`
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
		graphqlClient: graphql.NewClient(args.Url),
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

	req := graphql.NewRequest(`
		mutation CreateAccount ($input: CreateAccountInput!) {
			createAccount (input: $input) {
				message
				success
				code
				account {
					id
					asset {
						code
						scale
					}
				}
			}
		}
	`)

	type asset struct {
		Code  string `json:"code"`
		Scale int    `json:"scale"`
	}
	req.Var("input", struct {
		Asset      asset  `json:"asset"`
		PublicName string `json:"publicName"`
	}{
		Asset: asset{
			Code:  args.AssetCode,
			Scale: int(args.AssetScale),
		},
	})

	var resp map[string]struct {
		Message string
		Code    string
		Success bool
		Account struct {
			ID    string `json:"id"`
			Asset asset
		}
	}
	if err := s.graphqlClient.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	mutationResponse := resp["createAccount"]
	if !mutationResponse.Success {
		return nil, fmt.Errorf("%w %s", ErrInternal, mutationResponse.Message)
	}

	ret := &Identifier{}
	err := s.db.Get(
		ret,
		`
		INSERT INTO rafiki_identifiers (id, account_id, asset_code, asset_scale)
		VALUES ($1, $2, $3, $4) RETURNING *;`,
		mutationResponse.Account.ID,
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
		return nil, fmt.Errorf("%w %s", ErrInternal, err) // TODO: handle not found by code.
	}

	return ret, nil
}
