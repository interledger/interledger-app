package external

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"

	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/providers/xago/external/domain/dto"
)

type Client interface {
	CreateSubAccount(ctx context.Context, user user.User, details kyc.IndividualDetails, idNumber, personaInquiryURL string) (dto.SubAccount, error)
	AddBeneficiary(ctx context.Context, reqStruct dto.CreateBeneficiaryRequest) (dto.AccountBeneficiaries, error)
	ListBeneficiaries(ctx context.Context, limit, page uint) (dto.ListBeneficiariesResponse, error)
	CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID, reference string) (string, error)
	ListDeposits(ctx context.Context, page int) ([]dto.Deposit, error)
	GetWithdrawal(ctx context.Context, id string) (dto.Withdrawal, error)
	TestDeposit(ctx context.Context, reqStruct dto.TestDepositRequest) error
	UpdateSubAccount(ctx context.Context, accountID string, reqStruct dto.UpdateSubAccountRequest) error
	BankAccounts(ctx context.Context) ([]dto.Currency, error)
	GetDeposit(ctx context.Context, id string) (dto.Deposit, error)

	EstimateCurrencyConvert(ctx context.Context, pair dto.ConvertCurrencyPairEnum, amount float64) (dto.ConvertCurrencyResponse, error)
	CurrencyConvert(ctx context.Context, pair dto.ConvertCurrencyPairEnum, amount float64) (string, error)
}

// Config holds the configuration for the Xago external client.
type Config struct {
	APIBaseURL      string
	IdentityBaseURL string
	PublicKey       string
	Secret          string
	PolicyID        string
}

type client struct {
	baseURL     string
	identityURL string

	http *http.Client

	dbc *sqlx.DB

	identityService identityService
	baseService     baseService
}

func New(httpClient *http.Client, dbc *sqlx.DB, cfg Config) Client {
	transport := http.RoundTripper(http.DefaultTransport)
	if httpClient != nil && httpClient.Transport != nil {
		transport = httpClient.Transport
	}

	c := &client{
		dbc:         dbc,
		baseURL:     cfg.APIBaseURL,
		identityURL: cfg.IdentityBaseURL,
		http: &http.Client{
			Transport: &apiRoundTripper{
				loginURL:         cfg.IdentityBaseURL + "login",
				tokenProvider:    newTokenProvider(dbc, cfg, transport),
				defaultTransport: transport,
				headers: map[string]string{
					acceptHeader:      acceptValue,
					contentTypeHeader: contentTypeValue,
				},
			},
		},
	}

	c.identityService = identityService{client: c}
	c.baseService = baseService{client: c}

	return c
}

func (c *client) BankAccounts(ctx context.Context) ([]dto.Currency, error) {
	return c.baseService.BankAccounts(ctx)
}

func (c *client) CreateSubAccount(ctx context.Context, user user.User, details kyc.IndividualDetails, idNumber, personaInquiryURL string) (dto.SubAccount, error) {
	return c.baseService.CreateSubAccount(ctx, user, details, idNumber, personaInquiryURL)
}

func (c *client) AddBeneficiary(ctx context.Context, reqStruct dto.CreateBeneficiaryRequest) (dto.AccountBeneficiaries, error) {
	return c.identityService.AddBeneficiary(ctx, reqStruct)
}

func (c *client) ListBeneficiaries(ctx context.Context, limit, page uint) (dto.ListBeneficiariesResponse, error) {
	return c.identityService.ListBeneficiaries(ctx, limit, page)
}

func (c *client) CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID, reference string) (string, error) {
	return c.baseService.CreateTransaction(ctx, amt, idempotencyKey, beneficiaryID, reference)
}

func (c *client) ListDeposits(ctx context.Context, page int) ([]dto.Deposit, error) {
	return c.baseService.ListDeposits(ctx, page)
}

func (c *client) GetDeposit(ctx context.Context, id string) (dto.Deposit, error) {
	return c.baseService.GetDeposit(ctx, id)
}

func (c *client) GetWithdrawal(ctx context.Context, id string) (dto.Withdrawal, error) {
	return c.baseService.GetWithdrawal(ctx, id)
}

func (c *client) TestDeposit(ctx context.Context, reqStruct dto.TestDepositRequest) error {
	return c.baseService.TestDeposit(ctx, reqStruct)
}

func (c *client) UpdateSubAccount(ctx context.Context, accountID string, reqStruct dto.UpdateSubAccountRequest) error {
	return c.identityService.UpdateSubAccount(ctx, accountID, reqStruct)
}

func (c *client) EstimateCurrencyConvert(ctx context.Context, pair dto.ConvertCurrencyPairEnum, amount float64) (dto.ConvertCurrencyResponse, error) {
	return c.baseService.EstimateCurrencyConvert(ctx, pair, amount)
}

func (c *client) CurrencyConvert(ctx context.Context, pair dto.ConvertCurrencyPairEnum, amount float64) (string, error) {
	return c.baseService.CurrencyConvert(ctx, pair, amount)
}
