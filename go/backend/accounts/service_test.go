package accounts

import (
	"context"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestAccountsService(s *testing.T) {
	ctx := context.Background()

	pacioliContainer, err := test_utils.SetupPacioli(ctx)
	if err != nil {
		s.Fatal(err)
	}
	defer func() {
		err = pacioliContainer.Terminate(ctx)
		if err != nil {
			s.Fatal(err)
		}
	}()

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, cleanup := test_utils.MigrateCockroachDB(s, ctx)
	defer func() {
		cleanup()
	}()
	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}

	cs := _country.NewService(db)
	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	is = _identity.NewLoggingService(is, logger)

	pacioliLedgerID := uint16(1)
	conn, err := grpc.Dial(pacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.Fatal(err)
	}
	pClient := pacioliv1.NewPacioliServiceClient(conn)
	as, err := NewService(&ServiceArgs{
		Is:              is,
		Cs:              cs,
		PacioliLedgerID: pacioliLedgerID,
		PacioliClient:   pClient,
		PacioliTenant:   "dev",
		Db:              db,
	})
	if err != nil {
		s.Fatal(err)
	}
	as = NewLoggingService(as, logger)
	err = as.Init(ctx)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("GetByIdentityID requires identityID", func(t *testing.T) {
		var acc *Account
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_acc, err := as.GetByIdentityIDWithTrx(ctx, tx, "")
			if err != nil {
				return err
			}

			acc = _acc
			return nil
		})

		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "IdentityID is required.")
		assert.Nil(t, acc)
	})

	s.Run("Get requires accountID", func(t *testing.T) {
		var acc *Account
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_acc, err := as.Get(ctx, "")
			if err != nil {
				return err
			}

			acc = _acc
			return nil
		})

		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "AccountID is required.")
		assert.Nil(t, acc)
	})

	s.Run("verify account", func(t *testing.T) {
		var identity *_identity.Identity
		err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Create(ctx, &_identity.CreateArgs{
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

		t.Run("updates verification state to verified", func(tt *testing.T) {
			acc, err := as.Create(ctx, &CreateAccountArgs{
				IdentityID: identity.ID,
				Country:    identity.Country,
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.False(tt, acc.IsVerified())
			var verifiedAcc *Account
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				verifiedAcc, err = as.VerifyWithTx(ctx, tx, &VerifyArgs{
					AccountID:  acc.ID,
					Provider:   "noop",
					ProviderID: "test-customer1",
				})
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}

			assert.True(tt, verifiedAcc.IsVerified())
			assert.Equal(tt, "noop", verifiedAcc.Provider)
			assert.Equal(tt, "test-customer1", verifiedAcc.ProviderID)
		})
	})

	s.Run("create account", func(t *testing.T) {
		t.Run("writes to db if written to pacioli", func(tt *testing.T) {
			type scenario struct {
				Name                       string
				DebitsMustNotExceedCredits bool
				CreditsMustNotExceedDebits bool
			}
			scenarios := []scenario{
				{
					Name:                       "Sets DebitsMustNotExceedCredits",
					DebitsMustNotExceedCredits: true,
				},
				{
					Name:                       "Sets CreditsMustNotExceedDebits",
					CreditsMustNotExceedDebits: true,
				},
			}

			for _, scenario := range scenarios {
				var identity *_identity.Identity
				err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
					_identity, err := is.Create(ctx, &_identity.CreateArgs{
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
				assert.NotNil(t, identity)

				acc, err := as.Create(ctx, &CreateAccountArgs{
					IdentityID:                 identity.ID,
					Country:                    identity.Country,
					DebitMustNotExceedCredits:  scenario.DebitsMustNotExceedCredits,
					CreditsMustNotExceedDebits: scenario.CreditsMustNotExceedDebits,
				})
				if err != nil {
					tt.Fatal(err)
				}
				assert.Equal(tt, uint64(0), acc.DebitsAccepted)
				assert.Equal(tt, uint64(0), acc.DebitsReserved)
				assert.Equal(tt, uint64(0), acc.CreditsAccepted)
				assert.Equal(tt, uint64(0), acc.CreditsReserved)
				assert.Equal(tt, identity.ID, acc.IdentityID)
				assert.Equal(tt, scenario.CreditsMustNotExceedDebits, acc.CreditsMustNotExceedDebits)
				assert.Equal(tt, scenario.DebitsMustNotExceedCredits, acc.DebitsMustNotExceedCredits)

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
				assert.Equal(tt, uint64(0), freshAcc.DebitsReserved)
				assert.Equal(tt, uint64(0), freshAcc.DebitsAccepted)
				assert.Equal(tt, uint64(0), freshAcc.CreditsAccepted)
				assert.Equal(tt, uint64(0), freshAcc.CreditsReserved)
				assert.Equal(tt, identity.ID, freshAcc.IdentityID)
				assert.Equal(tt, scenario.CreditsMustNotExceedDebits, freshAcc.CreditsMustNotExceedDebits)
				assert.Equal(tt, scenario.DebitsMustNotExceedCredits, freshAcc.DebitsMustNotExceedCredits)

				freshAccGottenByID, err := as.Get(ctx, acc.ID)
				if err != nil {
					tt.Fatal(err)
				}
				assert.Equal(tt, uint64(0), freshAccGottenByID.DebitsReserved)
				assert.Equal(tt, uint64(0), freshAccGottenByID.DebitsAccepted)
				assert.Equal(tt, uint64(0), freshAccGottenByID.CreditsAccepted)
				assert.Equal(tt, uint64(0), freshAccGottenByID.CreditsReserved)
				assert.Equal(tt, identity.ID, freshAccGottenByID.IdentityID)
				assert.Equal(tt, scenario.CreditsMustNotExceedDebits, freshAccGottenByID.CreditsMustNotExceedDebits)
				assert.Equal(tt, scenario.DebitsMustNotExceedCredits, freshAccGottenByID.DebitsMustNotExceedCredits)
			}
		})

		t.Run("validates arguments", func(tt *testing.T) {
			var identity *_identity.Identity
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_identity, err := is.Create(ctx, &_identity.CreateArgs{
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
			assert.NotNil(t, identity)
			type scenario struct {
				Name                 string
				Args                 *CreateAccountArgs
				ExpectedErrorMessage string
				ExpectedError        error
			}
			scenarios := []scenario{
				{
					Name: "Country code must be valid.",
					Args: &CreateAccountArgs{
						IdentityID: identity.ID,
						Country:    "XCV",
					},
					ExpectedErrorMessage: "Key: 'CreateAccountArgs.Country' Error:Field validation for 'Country' failed on the 'iso3166_1_alpha2' tag",
					ExpectedError:        ErrInvalidArgument,
				},
				{
					Name: "Identity must exist",
					Args: &CreateAccountArgs{
						IdentityID: uuid.NewString(),
						Country:    "US",
					},
					ExpectedErrorMessage: "not found.",
					ExpectedError:        ErrInternal,
				},
			}

			for _, scenario := range scenarios {
				acc, err := as.Create(ctx, scenario.Args)

				assert.ErrorIs(tt, err, scenario.ExpectedError)
				assert.Contains(tt, err.Error(), scenario.ExpectedErrorMessage)
				assert.Nil(tt, acc, scenario.Name)
			}
		})

		// This test must come last as we close the connection to Pacioli.
		t.Run("fails if not written to pacioli", func(tt *testing.T) {
			conn.Close()
			assert.Equal(t, "SHUTDOWN", conn.GetState().String())
			var identity *_identity.Identity
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_identity, err := is.Create(ctx, &_identity.CreateArgs{
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
			assert.NotNil(t, identity)

			acc, err := as.Create(ctx, &CreateAccountArgs{
				IdentityID: identity.ID,
				Country:    identity.Country,
			})

			assert.Error(tt, err)
			assert.Nil(tt, acc)
		})
	})
}
