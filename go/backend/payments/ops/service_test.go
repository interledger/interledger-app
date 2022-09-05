package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"gitlab.com/fynbos/backend/twilio"

	"github.com/bxcodec/faker/v3"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/accounts"
	_accounts "gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"
	_country "gitlab.com/fynbos/backend/country"
	country_client "gitlab.com/fynbos/backend/country/client"
	"gitlab.com/fynbos/backend/identity"
	_identity "gitlab.com/fynbos/backend/identity"
	identity_client "gitlab.com/fynbos/backend/identity/client"
	"gitlab.com/fynbos/backend/onboarding"
	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"
	"gitlab.com/fynbos/backend/payments"
	payments_client "gitlab.com/fynbos/backend/payments/client"
	"gitlab.com/fynbos/backend/providers/noop"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
)

func TestPayments(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("initiating an outgoing payment adds a workflow", func(t *testing.T) {
		amount := uint64(100)
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		id, err := container.IdentityService.Create(container.Ctx, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		acc, err := container.AccountService.Create(
			ctx,
			&accounts.CreateAccountArgs{
				IdentityID: id.ID,
				Provider:   "unit",
				ProviderID: "test",
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = NewDeposit(container, &account_transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Test transaction",
			Type:        "deposit",
			NetAmount:   1000,
			LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.NoopService.GetLedgerID(),
					CreditAccountID: container.NoopService.GetEquityAccountID(),
					DebitAccountID:  acc.LedgerAccountID,
					Amount:          1000,
					Code:            1,
					Flags:           account_transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		container.TemporalMock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(
			func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
				testWorkflowID := opts.ID
				testRunID := "test-runid"

				mockWorkflowRun := &mocks.WorkflowRun{}
				mockWorkflowRun.On("GetID").Return(testWorkflowID)
				mockWorkflowRun.On("GetRunID").Return(testRunID)
				mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
				return mockWorkflowRun
			}, nil,
		).Times(1)

		container.TwilioImpl.EXPECT().CheckVerificationCode(gomock.Any(), &twilio.CheckVerificationCodeArgs{
			PhoneNumber: id.MobileNumber,
			Code:        "103",
		}).Return(&twilio.Verification{
			Status:      "approved",
			PhoneNumber: id.MobileNumber,
			Sid:         "",
		}, nil).Times(1)

		p, err := container.PaymentService.InitiateOutgoingPayment(context.Background(), payments.InitiateOutgoingPaymentArgs{
			UserID: acc.IdentityID,
			Amount: amount,
			To:     "$test.fynbos.test/alice",
			OTP:    "103",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, payments.Created, p.State)
		assert.Equal(t, acc.ID, p.AccountID)
		assert.Equal(t, amount, p.Amount)
	})

	s.Run("initiating an outgoing payment with insufficient balance fails", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		id, err := container.IdentityService.Create(container.Ctx, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		acc, err := container.AccountService.Create(
			ctx,
			&accounts.CreateAccountArgs{
				IdentityID: id.ID,
				Provider:   "unit",
				ProviderID: "test",
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		container.TemporalMock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(
			func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
				testWorkflowID := opts.ID
				testRunID := "test-runid"

				mockWorkflowRun := &mocks.WorkflowRun{}
				mockWorkflowRun.On("GetID").Return(testWorkflowID)
				mockWorkflowRun.On("GetRunID").Return(testRunID)
				mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
				return mockWorkflowRun
			}, nil,
		).Times(1)

		p, err := container.PaymentService.InitiateOutgoingPayment(context.Background(), payments.InitiateOutgoingPaymentArgs{
			UserID: acc.IdentityID,
			Amount: 100,
			To:     "$test.fynbos.test/alice",
		})

		assert.Nil(t, p)
		assert.ErrorIs(t, err, payments.ErrInsufficientBalance)
	})
}

type TestContainer struct {
	Ctrl               *gomock.Controller
	IdentityService    _identity.Client
	AccountService     _accounts.Client
	CountryService     _country.Client
	NoopService        noop.Service
	OnboardService     onboarding.Client
	TransactionService account_transactions.Client
	PaymentService     payments.Client
	TemporalMock       *mocks.Client
	PacioliClient      pacioli.Client
	PacioliLedgerID    uint32
	Db                 *sqlx.DB
	Logger             *zap.Logger
	Ctx                context.Context
	ValidatorImpl      *validator.Validate
	TwilioImpl         *twilio.MockService
}

func (t TestContainer) Payments() payments.Client {
	return t.PaymentService
}

func (t TestContainer) Transactions() account_transactions.Client {
	return t.TransactionService
}

func (t TestContainer) Noop() noop.Service {
	return t.NoopService
}

func (t TestContainer) Temporal() client.Client {
	return t.TemporalMock
}

func (t TestContainer) Accounts() _accounts.Client {
	return t.AccountService
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Identity() _identity.Client {
	return t.IdentityService
}

func (t TestContainer) Countries() _country.Client {
	return t.CountryService
}

func (t TestContainer) Pacioli() pacioli.Client {
	return t.PacioliClient
}

func (t TestContainer) Twilio() twilio.Service {
	return t.TwilioImpl
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	db := test_utils.MigrateCockroachDB(s, ctx)
	c.Db = db

	c.PacioliLedgerID = 1
	pClient := test_utils.SetupPacioli(s, ctx)
	c.PacioliClient = pClient

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	c.Ctrl = gomock.NewController(s)
	c.TwilioImpl = twilio.NewMockService(c.Ctrl)

	cs := country_client.New(c)
	c.CountryService = cs

	is := identity_client.New(c, logger)
	c.IdentityService = is

	as := accounts_client.New(c, c.PacioliLedgerID, logger)

	c.AccountService = as

	np, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   "46d4b2bd-e29b-4a63-9aa8-7990776c714e",
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}

	c.NoopService = np

	temporal := &mocks.Client{}
	c.TemporalMock = temporal

	os := onboarding_client.New(c)
	c.OnboardService = os

	ts := transactions_client.New(c, logger)
	c.TransactionService = ts

	ps := payments_client.New(c, logger)
	c.PaymentService = ps

	return c, nil
}

func NewDeposit(
	container *TestContainer,
	args *account_transactions.CreateTransactionArgs,
) (*account_transactions.AccountTransaction, error) {
	trx, err := container.TransactionService.Create(container.Ctx, args)
	if err != nil {
		return nil, err
	}

	return trx, nil
}
