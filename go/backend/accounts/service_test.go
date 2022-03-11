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
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/proto/pacioli/v1"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	if err != nil {
		s.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			s.Fatal(err)
		}
	}()
	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()

	cs := _country.NewService(db)
	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
	})
	if err != nil {
		s.Fatal(err)
	}
	is = _identity.NewLoggingService(is, logger)

	pacioliLedgerID := uint16(1)
	pClient := mockPacioliV1.NewMockPacioliServiceClient(ctrl)
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
	err = as.Init(ctx)
	if err != nil {
		s.Fatal(err)
	}

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
			pClient.EXPECT().ConfigureAccounts(ctx, gomock.Any(), gomock.Any()).Return(
				&pacioli.ConfigureAccountsResponse{}, nil,
			).Times(1)
			pClient.EXPECT().GetAccounts(ctx, gomock.Any()).DoAndReturn(
				func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioli.GetAccountsResponse, error) {
					return &pacioli.GetAccountsResponse{
						Accounts: []*pacioli.Account{
							{
								Id:              args.Ids[0],
								DebitsReserved:  1, // return non-zero to make sure default values aren't used.
								DebitsAccepted:  2,
								CreditsAccepted: 3,
								CreditsReserved: 4,
							},
						},
					}, nil
				}).Times(2)
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
			assert.Equal(tt, uint64(0), acc.DebitsAccepted)
			assert.Equal(tt, uint64(0), acc.DebitsReserved)
			assert.Equal(tt, uint64(0), acc.CreditsAccepted)
			assert.Equal(tt, uint64(0), acc.CreditsReserved)
			assert.Equal(tt, identity.ID, acc.IdentityID)

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
			assert.Equal(tt, uint64(1), freshAcc.DebitsReserved)
			assert.Equal(tt, uint64(2), freshAcc.DebitsAccepted)
			assert.Equal(tt, uint64(3), freshAcc.CreditsAccepted)
			assert.Equal(tt, uint64(4), freshAcc.CreditsReserved)
			assert.Equal(tt, identity.ID, freshAcc.IdentityID)

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
			assert.Equal(tt, uint64(1), freshAccGottenByID.DebitsReserved)
			assert.Equal(tt, uint64(2), freshAccGottenByID.DebitsAccepted)
			assert.Equal(tt, uint64(3), freshAccGottenByID.CreditsAccepted)
			assert.Equal(tt, uint64(4), freshAccGottenByID.CreditsReserved)
			assert.Equal(tt, identity.ID, freshAccGottenByID.IdentityID)
		})

		t.Run("fails if not written to pacioli", func(tt *testing.T) {
			pClient.EXPECT().ConfigureAccounts(ctx, gomock.Any()).Return(
				nil, status.Error(codes.Internal, "Failed to create account."),
			).Times(1)

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
				var acc *Account
				err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
					_acc, err := as.Create(ctx, tx, scenario.Args)
					if err != nil {
						return err
					}

					acc = _acc
					return nil
				})

				assert.ErrorIs(tt, err, scenario.ExpectedError)
				assert.Contains(tt, err.Error(), scenario.ExpectedErrorMessage)
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

		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "IdentityID is required.")
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

		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "AccountID is required.")
		assert.Nil(t, acc)
	})

	s.Run("verify account", func(t *testing.T) {
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

		t.Run("updates verification state to verified", func(tt *testing.T) {
			pClient.EXPECT().ConfigureAccounts(ctx, gomock.Any()).Return(
				&pacioli.ConfigureAccountsResponse{}, nil,
			).Times(1)
			var verifiedAcc *Account
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				acc, err := as.Create(ctx, tx, &CreateAccountArgs{
					IdentityID: identity.ID,
					Country:    identity.Country,
				})
				if err != nil {
					return err
				}
				assert.False(tt, acc.IsVerified())

				pClient.EXPECT().GetAccounts(ctx, gomock.Any()).DoAndReturn(
					func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioli.GetAccountsResponse, error) {
						return &pacioli.GetAccountsResponse{
							Accounts: []*pacioli.Account{
								{
									Id: args.Ids[0],
								},
							},
						}, nil
					}).Times(1)
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
}
