package ops_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	_accounts "gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"go.uber.org/zap"
)

func TestAccountTransactions(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("validates arguments", func(t *testing.T) {
		type scenario struct {
			Name            string
			Args            *account_transactions.CreateTransactionArgs
			ExpectedMessage string
			ExpectedError   error
		}
		scenarios := []scenario{
			{
				Name:            "Account must exist.",
				Args:            generateCreateTransactionArgs(withAccountID(uuid.NewString())),
				ExpectedMessage: "not found.",
				ExpectedError:   account_transactions.ErrInternal,
			},
			{
				Name:            "Amount must be greater than 0.",
				Args:            generateCreateTransactionArgs(withAmount(0)),
				ExpectedMessage: "Key: 'CreateTransactionArgs.NetAmount' Error:Field validation for 'NetAmount' failed on the 'gt' tag",
				ExpectedError:   account_transactions.ErrInvalidArgument,
			},
		}

		for _, scenario := range scenarios {
			trx, err := container.TransactionService.Create(ctx, scenario.Args)

			assert.ErrorIs(t, err, scenario.ExpectedError)
			assert.Contains(t, err.Error(), scenario.ExpectedMessage)
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
			t.Fatal(err)
		}
		equityAccID, err := createEquityAccount(container)
		if err != nil {
			t.Fatal(err)
		}

		args := generateCreateTransactionArgs(
			withAccountID(acc.ID),
			withAmount(17),
			withLedgerTransfers([]account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.PacioliLedgerID,
					DebitAccountID:  acc.LedgerAccountID,
					CreditAccountID: equityAccID,
					Amount:          100,
					Code:            1,
					Flags:           account_transactions.LedgerTransferFlags{},
				},
			}),
		)

		trx, err := container.TransactionService.Create(container.Ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.AccountID, trx.AccountID)
		assert.Equal(t, int64(args.NetAmount), trx.NetAmount)
		assert.Equal(t, args.Description, trx.Description)
		assert.Equal(t, account_transactions.Posted, trx.State)
		assert.Equal(t, args.Type, trx.Type)

		// check that ledger transfers were created
		transfers, err := container.PacioliClient.GetTransfers(ctx, trx.TransferIDs)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, transfers, 1)
		transfer := transfers[0]
		assert.Equal(t, transfer.CreditAccountID, equityAccID)
		assert.Equal(t, transfer.DebitAccountID, acc.LedgerAccountID)
		assert.Equal(t, transfer.Flags.Pending, false)

		// check that get works
		var trxs []*account_transactions.AccountTransaction
		err = crdbsqlx.ExecuteTx(container.Ctx, container.Db, nil, func(tx *sqlx.Tx) error {
			_trxs, err := container.TransactionService.GetByAccount(container.Ctx, tx, &account_transactions.GetByAccountArgs{
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
		assert.Equal(t, account_transactions.Posted, fetchedTransaction.State)
		assert.Equal(t, args.Type, fetchedTransaction.Type)
		assert.Equal(t, trx.TransferIDs, fetchedTransaction.TransferIDs)
	})

	s.Run("pending", func(tg *testing.T) {
		tg.Run("can create a pending transaction", func(t *testing.T) {
			user := _user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			_, acc, err := onboard(container, &user)
			if err != nil {
				t.Fatal(err)
			}
			equityAccID, err := createEquityAccount(container)
			if err != nil {
				t.Fatal(err)
			}

			args := &account_transactions.CreatePendingTransactionArgs{
				AccountID:   acc.ID,
				Description: "Pending trx",
				Type:        "deposit",
				NetAmount:   100,
				LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
					{
						LedgerID:        container.PacioliLedgerID,
						DebitAccountID:  acc.LedgerAccountID,
						CreditAccountID: equityAccID,
						Amount:          100,
						Code:            1,
						Flags:           account_transactions.LedgerTransferFlags{},
					},
				},
			}

			trx, err := container.TransactionService.CreatePending(container.Ctx, args)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, args.AccountID, trx.AccountID)
			assert.Equal(t, int64(args.NetAmount), trx.NetAmount)
			assert.Equal(t, args.Description, trx.Description)
			assert.Equal(t, account_transactions.Pending, trx.State)
			assert.Equal(t, args.Type, trx.Type)

			// check that ledger transfers were created
			transfers, err := container.PacioliClient.GetTransfers(ctx, trx.TransferIDs)
			if err != nil {
				t.Fatal(err)
			}
			assert.Len(t, transfers, 1)
			transfer := transfers[0]
			assert.Equal(t, transfer.CreditAccountID, equityAccID)
			assert.Equal(t, transfer.DebitAccountID, acc.LedgerAccountID)
			assert.Equal(t, transfer.Flags.Pending, true)

			// check that get works
			var trxs []*account_transactions.AccountTransaction
			err = crdbsqlx.ExecuteTx(container.Ctx, container.Db, nil, func(tx *sqlx.Tx) error {
				_trxs, err := container.TransactionService.GetByAccount(container.Ctx, tx, &account_transactions.GetByAccountArgs{
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
			assert.Equal(t, account_transactions.Pending, fetchedTransaction.State)
			assert.Equal(t, args.Type, fetchedTransaction.Type)
			assert.Equal(t, trx.TransferIDs, fetchedTransaction.TransferIDs)
		})
	})

	s.Run("post pending", func(tg *testing.T) {
		tg.Run("can post a pending transaction", func(t *testing.T) {
			user := _user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			_, acc, err := onboard(container, &user)
			if err != nil {
				t.Fatal(err)
			}
			equityAccID, err := createEquityAccount(container)
			if err != nil {
				t.Fatal(err)
			}

			args := &account_transactions.CreatePendingTransactionArgs{
				AccountID:   acc.ID,
				Description: "Pending trx",
				Type:        "deposit",
				NetAmount:   100,
				LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
					{
						LedgerID:        container.PacioliLedgerID,
						DebitAccountID:  acc.LedgerAccountID,
						CreditAccountID: equityAccID,
						Amount:          100,
						Code:            1,
						Flags:           account_transactions.LedgerTransferFlags{},
					},
				},
			}
			trx, err := container.TransactionService.CreatePending(container.Ctx, args)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, account_transactions.Pending, trx.State)

			trx, err = container.TransactionService.PostPending(container.Ctx, trx.ID)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, account_transactions.Posted, trx.State)

			// TODO check the commit transfer was created
		})

		tg.Run("cant post a non pending transaction", func(t *testing.T) {
			user := _user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			_, acc, err := onboard(container, &user)
			if err != nil {
				t.Fatal(err)
			}
			equityAccID, err := createEquityAccount(container)
			if err != nil {
				t.Fatal(err)
			}

			args := generateCreateTransactionArgs(
				withAccountID(acc.ID),
				withAmount(17),
				withLedgerTransfers([]account_transactions.CreateLedgerTransferArgs{
					{
						LedgerID:        container.PacioliLedgerID,
						DebitAccountID:  acc.LedgerAccountID,
						CreditAccountID: equityAccID,
						Amount:          100,
						Code:            1,
						Flags:           account_transactions.LedgerTransferFlags{},
					},
				}),
			)

			trx, err := container.TransactionService.Create(container.Ctx, args)
			if err != nil {
				t.Fatal(err)
			}
			assert.NotEqual(t, account_transactions.Pending, trx.State)

			_, err = container.TransactionService.PostPending(container.Ctx, trx.ID)
			if err == nil {
				t.Fatal("expected error to occur")
			}
		})
	})

	s.Run("void pending", func(tg *testing.T) {
		tg.Run("can void a pending transaction", func(t *testing.T) {
			user := _user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			_, acc, err := onboard(container, &user)
			if err != nil {
				t.Fatal(err)
			}
			equityAccID, err := createEquityAccount(container)
			if err != nil {
				t.Fatal(err)
			}

			args := &account_transactions.CreatePendingTransactionArgs{
				AccountID:   acc.ID,
				Description: "Pending trx",
				Type:        "deposit",
				NetAmount:   100,
				LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
					{
						LedgerID:        container.PacioliLedgerID,
						DebitAccountID:  acc.LedgerAccountID,
						CreditAccountID: equityAccID,
						Amount:          100,
						Code:            1,
						Flags:           account_transactions.LedgerTransferFlags{},
					},
				},
			}
			trx, err := container.TransactionService.CreatePending(container.Ctx, args)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, account_transactions.Pending, trx.State)

			trx, err = container.TransactionService.VoidPending(container.Ctx, trx.ID)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, account_transactions.Voided, trx.State)
		})

		tg.Run("cant void a non pending transaction", func(t *testing.T) {
			user := _user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			_, acc, err := onboard(container, &user)
			if err != nil {
				t.Fatal(err)
			}
			equityAccID, err := createEquityAccount(container)
			if err != nil {
				t.Fatal(err)
			}

			args := generateCreateTransactionArgs(
				withAccountID(acc.ID),
				withAmount(17),
				withLedgerTransfers([]account_transactions.CreateLedgerTransferArgs{
					{
						LedgerID:        container.PacioliLedgerID,
						DebitAccountID:  acc.LedgerAccountID,
						CreditAccountID: equityAccID,
						Amount:          100,
						Code:            1,
						Flags:           account_transactions.LedgerTransferFlags{},
					},
				}),
			)

			trx, err := container.TransactionService.Create(container.Ctx, args)
			if err != nil {
				t.Fatal(err)
			}
			assert.NotEqual(t, account_transactions.Pending, trx.State)

			_, err = container.TransactionService.VoidPending(container.Ctx, trx.ID)
			if err == nil {
				t.Fatal("expected error to occur")
			}
		})
	})
}

func generateCreateTransactionArgs(opts ...func(*account_transactions.CreateTransactionArgs)) *account_transactions.CreateTransactionArgs {
	args := &account_transactions.CreateTransactionArgs{
		AccountID:       uuid.NewString(),
		Description:     faker.Sentence(),
		Type:            "deposit",
		NetAmount:       100,
		LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{},
	}
	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withAccountID(id string) func(*account_transactions.CreateTransactionArgs) {
	return func(args *account_transactions.CreateTransactionArgs) {
		args.AccountID = id
	}
}

func withAmount(amount uint64) func(*account_transactions.CreateTransactionArgs) {
	return func(args *account_transactions.CreateTransactionArgs) {
		args.NetAmount = amount
	}
}

func withLedgerTransfers(transfers []account_transactions.CreateLedgerTransferArgs) func(*account_transactions.CreateTransactionArgs) {
	return func(args *account_transactions.CreateTransactionArgs) {
		args.LedgerTransfers = transfers
	}
}

type TestContainer struct {
	IdentityService    _identity.Service
	AccountService     _accounts.Client
	CountryService     _country.Service
	PacioliContainer   *test_utils.PacioliContainer
	PacioliClient      pacioli.Client
	PacioliLedgerID    uint32
	Ctrl               *gomock.Controller
	TransactionService account_transactions.Client
	Db                 *sqlx.DB
	Logger             *zap.Logger
	Ctx                context.Context
	ValidatorImpl      *validator.Validate
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

func (t TestContainer) Identity() _identity.Service {
	return t.IdentityService
}

func (t TestContainer) Countries() _country.Service {
	return t.CountryService
}

func (t TestContainer) Pacioli() pacioli.Client {
	return t.PacioliClient
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{
		ValidatorImpl: validator.New(),
	}
	c.Ctx = ctx
	db := test_utils.MigrateCockroachDB(s, ctx)
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	cs := _country.NewService(db)
	c.CountryService = cs

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.IdentityService = _identity.NewLoggingService(is, logger)

	c.PacioliContainer = test_utils.SetupPacioli(s, ctx)

	c.PacioliLedgerID = uint32(1)

	pClient, err := pacioli_client.New(c.PacioliContainer.PacioliUrl)
	if err != nil {
		s.Fatal(err)
	}

	c.PacioliClient = pClient
	ac := accounts_client.New(c, c.PacioliLedgerID, logger)

	c.AccountService = ac

	c.TransactionService = transactions_client.New(c, logger)

	return c, nil
}

// Helper function that creates an identity and account for the user.
func onboard(container *TestContainer, user *_user.User) (*_identity.Identity, *accounts.Account, error) {
	var identity *_identity.Identity
	err := crdbsqlx.ExecuteTx(container.Ctx, container.Db, nil, func(tx *sqlx.Tx) error {
		id, err := container.IdentityService.Create(container.Ctx, &_identity.CreateArgs{
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
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	acc, err := container.AccountService.Create(container.Ctx, &_accounts.CreateAccountArgs{
		IdentityID: user.ID,
		Provider:   "unit",
		ProviderID: uuid.NewString(),
	})
	if err != nil {
		return nil, nil, err
	}

	return identity, acc, nil
}

func createEquityAccount(container *TestContainer) (string, error) {
	equityAccID := uuid.NewString()
	accountErrs, err := container.PacioliClient.ConfigureAccounts(context.Background(), []pacioli.ConfigureAccountArgs{
		{
			ID:       equityAccID,
			LedgerID: container.PacioliLedgerID,
			Code:     1,
		}},
	)
	if err != nil {
		return "", err
	}
	if len(accountErrs) != 0 {
		return "", fmt.Errorf("equity account failed to be created")
	}
	return equityAccID, nil
}
