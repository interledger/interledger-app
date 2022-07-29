package ops_test

import (
	"context"
	"testing"

	identity_client "gitlab.com/fynbos/backend/identity/client"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/accounts/ops"
	country_client "gitlab.com/fynbos/backend/country/client"
	_identity "gitlab.com/fynbos/backend/identity"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"go.uber.org/zap"
)

func TestAccountsService(s *testing.T) {
	ctx := context.Background()

	pacioliContainer := test_utils.SetupPacioli(s, ctx)

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db := test_utils.MigrateCockroachDB(s, ctx)
	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}

	cs := country_client.New(ops.NewTestBackends(s, db, nil, nil, nil))
	is := identity_client.New(ops.NewTestBackends(s, db, nil, cs, nil), logger)

	pacioliLedgerID := uint32(1)

	pClient, err := pacioli_client.New(pacioliContainer.PacioliUrl)
	if err != nil {
		s.Fatal(err)
	}

	b := ops.NewTestBackends(s, db, is, cs, pClient)

	s.Run("GetByIdentityID requires identityID", func(t *testing.T) {
		var acc *accounts.Account
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_acc, err := ops.GetByIdentityIDWithTrx(ctx, b, tx, "")
			if err != nil {
				return err
			}

			acc = _acc
			return nil
		})

		assert.ErrorIs(t, err, accounts.ErrInvalidArgument)
		assert.Contains(t, err.Error(), "IdentityID is required.")
		assert.Nil(t, acc)
	})

	s.Run("Get requires accountID", func(t *testing.T) {
		var acc *accounts.Account
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_acc, err := ops.Get(ctx, b, "")
			if err != nil {
				return err
			}

			acc = _acc
			return nil
		})

		assert.ErrorIs(t, err, accounts.ErrInvalidArgument)
		assert.Contains(t, err.Error(), "AccountID is required.")
		assert.Nil(t, acc)
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

				acc, err := ops.Create(ctx, b, pacioliLedgerID, &accounts.CreateAccountArgs{
					IdentityID:                 identity.ID,
					DebitMustNotExceedCredits:  scenario.DebitsMustNotExceedCredits,
					CreditsMustNotExceedDebits: scenario.CreditsMustNotExceedDebits,
					Provider:                   "unit",
					ProviderID:                 uuid.NewString(),
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

				var freshAcc *accounts.Account
				err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
					_acc, err := ops.GetByIdentityIDWithTrx(ctx, b, tx, identity.ID)
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

				freshAccGottenByID, err := ops.Get(ctx, b, acc.ID)
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
				Args                 *accounts.CreateAccountArgs
				ExpectedErrorMessage string
				ExpectedError        error
			}
			scenarios := []scenario{
				{
					Name: "Identity must exist",
					Args: &accounts.CreateAccountArgs{
						IdentityID: uuid.NewString(),
						Provider:   "unit",
						ProviderID: uuid.NewString(),
					},
					ExpectedErrorMessage: "not found.",
					ExpectedError:        accounts.ErrInternal,
				},
			}

			for _, scenario := range scenarios {
				acc, err := ops.Create(ctx, b, pacioliLedgerID, scenario.Args)

				assert.ErrorIs(tt, err, scenario.ExpectedError)
				assert.Contains(tt, err.Error(), scenario.ExpectedErrorMessage)
				assert.Nil(tt, acc, scenario.Name)
			}
		})
	})

	// This test must come last as we close the connection to Pacioli.
	//TODO: Do we need this test
	/*t.Run("fails if not written to pacioli", func(tt *testing.T) {
			pacioliContainer.Pacioli.Process.Kill()
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
				Provider:   "unit",
				ProviderID: uuid.NewString(),
			})

			assert.Error(tt, err)
			assert.Nil(tt, acc)
		})
	})*/
}
