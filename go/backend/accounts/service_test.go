package accounts

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/identity/noop"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccountsService(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}
	defer logger.Sync()

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()

	cs := _country.NewService()
	provider := noop.NewMockProvider(ctrl)
	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		NoopProvider:   provider,
	})
	if err != nil {
		s.Fatal(err)
	}
	is = _identity.NewLoggingService(is, logger)

	pacioliLedgerID := uuid.NewString()
	ledgerCode := uint16(1)
	pClient := mockPacioliV1.NewMockPacioliServiceClient(ctrl)
	pClient.EXPECT().GetLedgerByCode(gomock.Any(), gomock.Any()).Return(&pacioliv1.Ledger{
		Id:   pacioliLedgerID,
		Code: uint32(ledgerCode),
	}, nil).Times(1)
	as, err := NewService(&ServiceArgs{
		Is:                is,
		Cs:                cs,
		PacioliLedgerCode: ledgerCode,
		PacioliClient:     pClient,
		Db:                db,
	})
	if err != nil {
		s.Fatal(err)
	}
	as = NewLoggingService(as, logger)

	s.Run("create account", func(t *testing.T) {
		var identity *_identity.Identity
		err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Create(ctx, tx, _identity.CreateArgs{
				ID:           uuid.NewString(),
				Email:        faker.Email(),
				FirstName:    faker.Name(),
				LastName:     faker.LastName(),
				MobileNumber: faker.Phonenumber(),
				Country:      "US",
			})
			if err != nil {
				return err
			}

			identity = _identity
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Run("writes to db if written to pacioli", func(tt *testing.T) {
			ledgerAccountID := uuid.NewString()
			pClient.EXPECT().CreateAccount(ctx, gomock.Any()).Return(&pacioliv1.Account{
				Id: ledgerAccountID,
			}, nil).Times(1)

			var acc *Account
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_acc, err := as.Create(ctx, tx, &CreateAccountArgs{
					IdentityID: identity.ID,
					Country:    identity.Country,
				})
				if err != nil {
					return err
				}

				acc = _acc
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, ledgerAccountID, acc.LedgerAccountID)
			assert.Equal(tt, uint64(0), acc.DebitsAccepted)
			assert.Equal(tt, uint64(0), acc.DebitsReserved)
			assert.Equal(tt, uint64(0), acc.CreditsAccepted)
			assert.Equal(tt, uint64(0), acc.CreditsReserved)
			assert.Equal(tt, identity.ID, acc.IdentityID)

			pClient.EXPECT().GetAccount(ctx, gomock.Any()).Return(&pacioliv1.Account{
				Id:              ledgerAccountID,
				DebitsReserved:  1, // return non-zero to make sure default values aren't used.
				DebitsAccepted:  2,
				CreditsAccepted: 3,
				CreditsReserved: 4,
			}, nil).Times(1)
			var freshAcc *Account
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_acc, err := as.GetByIdentityIDWithTrx(ctx, tx, identity.ID)
				if err != nil {
					return err
				}

				freshAcc = _acc
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, ledgerAccountID, freshAcc.LedgerAccountID)
			assert.Equal(tt, uint64(1), freshAcc.DebitsReserved)
			assert.Equal(tt, uint64(2), freshAcc.DebitsAccepted)
			assert.Equal(tt, uint64(3), freshAcc.CreditsAccepted)
			assert.Equal(tt, uint64(4), freshAcc.CreditsReserved)
			assert.Equal(tt, identity.ID, freshAcc.IdentityID)

			pClient.EXPECT().GetAccount(ctx, gomock.Any()).Return(&pacioliv1.Account{
				Id:              ledgerAccountID,
				DebitsReserved:  1, // return non-zero to make sure default values aren't used.
				DebitsAccepted:  2,
				CreditsAccepted: 3,
				CreditsReserved: 4,
			}, nil).Times(1)
			var freshAccGottenByID *Account
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_acc, err := as.Get(ctx, tx, acc.ID)
				if err != nil {
					return err
				}

				freshAccGottenByID = _acc
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, ledgerAccountID, freshAccGottenByID.LedgerAccountID)
			assert.Equal(tt, uint64(1), freshAccGottenByID.DebitsReserved)
			assert.Equal(tt, uint64(2), freshAccGottenByID.DebitsAccepted)
			assert.Equal(tt, uint64(3), freshAccGottenByID.CreditsAccepted)
			assert.Equal(tt, uint64(4), freshAccGottenByID.CreditsReserved)
			assert.Equal(tt, identity.ID, freshAccGottenByID.IdentityID)
		})

		t.Run("fails if not written to pacioli", func(tt *testing.T) {
			pClient.EXPECT().CreateAccount(ctx, gomock.Any()).
				Return(nil, status.Error(codes.Internal, "Failed to create account.")).Times(1)

			var acc *Account
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_acc, err := as.Create(ctx, tx, &CreateAccountArgs{
					IdentityID: identity.ID,
					Country:    identity.Country,
				})
				if err != nil {
					return err
				}

				acc = _acc
				return nil
			})

			assert.Error(tt, err)
			assert.Nil(tt, acc)
		})

		t.Run("validates arguments", func(tt *testing.T) {
			type scenario struct {
				Name          string
				Args          *CreateAccountArgs
				ExpectedError string
			}
			scenarios := []scenario{
				{
					Name: "Country code must be valid.",
					Args: &CreateAccountArgs{
						IdentityID: identity.ID,
						Country:    "XCV",
					},
					ExpectedError: "Key: 'CreateAccountArgs.Country' Error:Field validation for 'Country' failed on the 'iso3166_1_alpha2' tag",
				},
				{
					Name: "Identity must exist",
					Args: &CreateAccountArgs{
						IdentityID: uuid.NewString(),
						Country:    "US",
					},
					ExpectedError: "Identity must exist.",
				},
			}

			for _, scenario := range scenarios {
				var acc *Account
				err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
					_acc, err := as.Create(ctx, tx, scenario.Args)
					if err != nil {
						return err
					}

					acc = _acc
					return nil
				})

				assert.Equal(tt, scenario.ExpectedError, err.Error())
				assert.Nil(tt, acc, scenario.Name)
			}
		})
	})

	s.Run("GetAccountByID requires identityID", func(t *testing.T) {
		var acc *Account
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_acc, err := as.GetByIdentityIDWithTrx(ctx, tx, "")
			if err != nil {
				return err
			}

			acc = _acc
			return nil
		})

		assert.Equal(t, "identityID is required.", err.Error())
		assert.Nil(t, acc)
	})

	s.Run("Get requires accountID", func(t *testing.T) {
		var acc *Account
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_acc, err := as.Get(ctx, tx, "")
			if err != nil {
				return err
			}

			acc = _acc
			return nil
		})

		assert.Equal(t, "Accounts service: accountID is required.", err.Error())
		assert.Nil(t, acc)
	})
}
