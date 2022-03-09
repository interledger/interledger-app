package account_transactions

import (
	"context"
	"log"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	_accounts "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/proto/pacioli/v1"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestAccountTransactions(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err := container.Cleanup()
		if err != nil {
			return
		}
	})

	s.Run("validates arguments", func(t *testing.T) {
		type scenario struct {
			Name          string
			Args          *CreateTransactionArgs
			ExpectedError string
		}
		scenarios := []scenario{
			{
				Name:          "Account must exist.",
				Args:          generateCreateTransactionArgs(withAccountID(uuid.NewString())),
				ExpectedError: "Transaction service: Accounts service: Not found.",
			},
			{
				Name:          "Amount must be greater than 0.",
				Args:          generateCreateTransactionArgs(withAmount(0)),
				ExpectedError: "Transaction service: Key: 'CreateTransactionArgs.NetAmount' Error:Field validation for 'NetAmount' failed on the 'gt' tag",
			},
			{
				Name:          "State is required.",
				Args:          generateCreateTransactionArgs(withState("")),
				ExpectedError: "Transaction service: Key: 'CreateTransactionArgs.State' Error:Field validation for 'State' failed on the 'required' tag",
			},
		}

		for _, scenario := range scenarios {
			var trx *AccountTransaction
			err := crdbsqlx.ExecuteTx(ctx, container.Db, nil, func(tx *sqlx.Tx) error {
				_trx, err := container.TransactionService.Create(ctx, tx, scenario.Args)
				if err != nil {
					return err
				}

				trx = _trx
				return nil
			})

			assert.Equal(t, scenario.ExpectedError, err.Error())
			assert.Nil(t, trx, scenario.Name)
		}
	})

	s.Run("can create a transaction", func(t *testing.T) {
		user := _user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, acc, err := onboard(container, &user)
		if err != nil {
			log.Fatalln(err)
		}
		ledgerID := uint16(1)
		equityAccID := uuid.NewString()
		args := generateCreateTransactionArgs(
			withAccountID(acc.ID),
			withAmount(17),
			withLedgerTransfers([]CreateLedgerTransferArgs{
				{
					LedgerID:        ledgerID,
					DebitAccountID:  acc.ID,
					CreditAccountID: equityAccID,
					Amount:          100,
					Code:            0,
					Flags:           LedgerTransferFlags{},
				},
			}),
		)

		// Create function calls out to Pacioli to check if the ledger account exists and then creates the transfers.
		expectedTransferIDs := []string{}
		container.MockPacioliClient.EXPECT().GetAccounts(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioli.GetAccountsResponse, error) {
				return &pacioli.GetAccountsResponse{
					Accounts: []*pacioli.Account{
						{
							Id: args.Ids[0],
						},
					},
				}, nil
			}).Times(1)
		container.MockPacioliClient.EXPECT().CreateTransfers(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.CreateTransfersRequest, opts ...grpc.CallOption) (*pacioliv1.CreateTransfersResponse, error) {
				assert.Len(t, args.Transfers, 1)
				expectedTransferIDs = append(expectedTransferIDs, args.Transfers[0].Id)
				return &pacioliv1.CreateTransfersResponse{
					Errors: []*pacioliv1.EventError{},
				}, nil
			}).Times(1)
		var trx *AccountTransaction
		err = crdbsqlx.ExecuteTx(container.Ctx, container.Db, nil, func(tx *sqlx.Tx) error {
			_trx, err := container.TransactionService.Create(container.Ctx, tx, args)
			if err != nil {
				return err
			}
			trx = _trx

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.AccountID, trx.AccountID)
		assert.Equal(t, int64(args.NetAmount), trx.NetAmount)
		assert.Equal(t, args.Description, trx.Description)
		assert.Equal(t, args.State, trx.State)
		assert.Equal(t, args.Type, trx.Type)
		assert.Equal(t, expectedTransferIDs, trx.TransferIDs)

		var trxs []*AccountTransaction
		err = crdbsqlx.ExecuteTx(container.Ctx, container.Db, nil, func(tx *sqlx.Tx) error {
			_trxs, err := container.TransactionService.GetByAccount(container.Ctx, tx, &GetByAccountArgs{
				AccountID: acc.ID,
				Limit:     10,
				OrderBy:   "ASC",
			})
			if err != nil {
				return err
			}
			trxs = _trxs

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, trxs, 1)
		fetchedTransaction := trxs[0]
		assert.Equal(t, args.AccountID, fetchedTransaction.AccountID)
		assert.Equal(t, int64(args.NetAmount), fetchedTransaction.NetAmount)
		assert.Equal(t, args.Description, fetchedTransaction.Description)
		assert.Equal(t, args.State, fetchedTransaction.State)
		assert.Equal(t, args.Type, fetchedTransaction.Type)
		assert.Equal(t, expectedTransferIDs, fetchedTransaction.TransferIDs)
	})
}

func generateCreateTransactionArgs(opts ...func(*CreateTransactionArgs)) *CreateTransactionArgs {
	args := &CreateTransactionArgs{
		AccountID:       uuid.NewString(),
		Description:     faker.Sentence(),
		Type:            "deposit",
		NetAmount:       100,
		State:           "pending",
		LedgerTransfers: []CreateLedgerTransferArgs{},
	}
	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withAccountID(id string) func(*CreateTransactionArgs) {
	return func(args *CreateTransactionArgs) {
		args.AccountID = id
	}
}

func withState(state string) func(*CreateTransactionArgs) {
	return func(args *CreateTransactionArgs) {
		args.State = state
	}
}

func withAmount(amount uint64) func(*CreateTransactionArgs) {
	return func(args *CreateTransactionArgs) {
		args.NetAmount = amount
	}
}

func withLedgerTransfers(transfers []CreateLedgerTransferArgs) func(*CreateTransactionArgs) {
	return func(args *CreateTransactionArgs) {
		args.LedgerTransfers = transfers
	}
}

type TestContainer struct {
	IdentityService    _identity.Service
	AccountService     _accounts.Service
	CountryService     _country.Service
	MockPacioliClient  *mockPacioliV1.MockPacioliServiceClient
	Ctrl               *gomock.Controller
	TransactionService Service
	Db                 *sqlx.DB
	Logger             *zap.Logger
	Crdb               *test_utils.CockroachDBContainer
	Ctx                context.Context
}

func (c *TestContainer) Cleanup() error {
	err := c.Db.Close()
	if err != nil {
		return err
	}
	c.Ctrl.Finish()

	return nil
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		return nil, err
	}
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	ctrl := gomock.NewController(s)
	c.Ctrl = ctrl

	cs := _country.NewService(db)
	c.CountryService = cs

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.IdentityService = _identity.NewLoggingService(is, logger)

	pClient := mockPacioliV1.NewMockPacioliServiceClient(ctrl)
	c.MockPacioliClient = pClient
	pacioliLedgerID := uint16(1)
	pClient.EXPECT().ConfigureLedgers(ctx, &pacioli.ConfigureLedgersRequest{
		Args: []*pacioli.Ledger{
			{
				Id:    uint32(pacioliLedgerID),
				Name:  "Fynbos ledger",
				Asset: "840", // US dollars
				Scale: 2,
			},
		},
	}).Return(&pacioli.ConfigureLedgersResponse{}, nil).Times(1)
	as, err := _accounts.NewService(&accounts.ServiceArgs{
		Is:              is,
		Cs:              cs,
		PacioliLedgerID: pacioliLedgerID,
		PacioliClient:   pClient,
		Db:              db,
	})
	if err != nil {
		return nil, err
	}
	err = as.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.AccountService = _accounts.NewLoggingService(as, logger)

	ts, err := NewService(&ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
	})
	if err != nil {
		return nil, err
	}
	c.TransactionService = NewLoggingService(ts, logger)
	return c, nil
}

// Helper function that creates an identity and account for the user.
func onboard(container *TestContainer, user *_user.User) (*_identity.Identity, *accounts.Account, error) {
	var acc *accounts.Account
	var identity *_identity.Identity
	err := crdbsqlx.ExecuteTx(container.Ctx, container.Db, nil, func(tx *sqlx.Tx) error {
		id, err := container.IdentityService.Create(container.Ctx, tx, _identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			return err
		}
		identity = id

		container.MockPacioliClient.EXPECT().ConfigureAccounts(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.ConfigureAccountsResponse{}, nil,
		).Times(1)
		_acc, err := container.AccountService.Create(container.Ctx, tx, &_accounts.CreateAccountArgs{
			IdentityID: user.ID,
			Country:    "US",
		})
		if err != nil {
			return err
		}
		acc = _acc

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return identity, acc, nil
}
