package fundingsources

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/onboarding"
)

func TestFundingSources(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err = c.Cleanup()
		if err != nil {
			s.Fatal(err)
		}
	})

	s.Run("validates create bank account arguments", func(t *testing.T) {
		userID := uuid.NewString()
		_, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   userID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		type Scenario struct {
			Name                 string
			Args                 *CreateBankAccountArgs
			ExpectedErrorMessage string
			ExpectedError        error
		}
		scenarios := []Scenario{
			{
				Name:                 "IdentityID is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withIdentityID("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.IdentityID' Error:Field validation for 'IdentityID' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "IdentityID must exist to create bank account",
				Args:                 generateCreateBankAccountArgs(withIdentityID(uuid.NewString())),
				ExpectedErrorMessage: "not found.",
				ExpectedError:        ErrInternal,
			},
			{
				Name:                 "AccountID is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withAccountID("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.AccountID' Error:Field validation for 'AccountID' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "AccountID must exist to create bank account",
				Args:                 generateCreateBankAccountArgs(withIdentityID(userID), withAccountID(uuid.NewString())),
				ExpectedErrorMessage: "not found.",
				ExpectedError:        ErrInternal,
			},
			{
				Name:                 "Name is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withName("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.Name' Error:Field validation for 'Name' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "Institution is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withInstitution("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.Institution' Error:Field validation for 'Institution' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "AccountNumber is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withAccountNumber("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.AccountNumber' Error:Field validation for 'AccountNumber' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "RoutingNumber is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withRoutingNumber("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.RoutingNumber' Error:Field validation for 'RoutingNumber' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "Type must be one of noop required to create bank account",
				Args:                 generateCreateBankAccountArgs(withType("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.Type' Error:Field validation for 'Type' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
		}

		for _, scenario := range scenarios {
			fs, err := c.Fs.CreateBankAccount(c.Ctx, scenario.Args)
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.ErrorIs(t, err, scenario.ExpectedError)
			assert.Contains(t, err.Error(), scenario.ExpectedErrorMessage)
			assert.Nil(t, fs)
		}
	})

	s.Run("creates bank account", func(t *testing.T) {
		userID := uuid.NewString()
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   userID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(userID),
			withAccountID(acc.ID),
		)
		fs, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.Name, fs.Name)
		assert.NotEqual(t, args.AccountNumber, fs.Mask)
		assert.Equal(t, "noop", fs.Type)
		assert.Equal(t, args.Type, fs.SubType)
		assert.Equal(t, acc.ID, fs.AccountID)
		assert.Equal(t, "required", fs.VerificationState)
	})

	s.Run("verifies bank account", func(t *testing.T) {
		userID := uuid.NewString()
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   userID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(userID),
			withAccountID(acc.ID),
		)
		bankAccount, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "required", bankAccount.VerificationState)

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      userID,
			FundingSourceID: bankAccount.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "verified", fs.VerificationState)
	})

	s.Run("returns not found if funding source does not belong to user", func(t *testing.T) {
		otherUserID := uuid.NewString()
		myID := uuid.NewString()
		myAcc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   myID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   otherUserID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(myID),
			withAccountID(myAcc.ID),
		)
		bankAccount, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "required", bankAccount.VerificationState)

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      otherUserID,
			FundingSourceID: bankAccount.ID,
		})
		if err == nil {
			t.Fatal("User must only be able to verify their own funding sources.")
		}

		assert.ErrorIs(t, err, ErrUnauthorized)
		assert.Contains(t, err.Error(), "unauthorized.")
		assert.Nil(t, fs)
	})

	s.Run("returns not found if funding source does not exist", func(t *testing.T) {
		userID := uuid.NewString()
		_, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   userID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      userID,
			FundingSourceID: uuid.NewString(),
		})
		if err == nil {
			t.Fatal("Must only be able to verify funding sources that exist.")
		}

		assert.Nil(t, fs)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Contains(t, err.Error(), "not found.")
	})

	s.Run("get user's funding sources", func(t *testing.T) {
		userID := uuid.NewString()
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   userID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		fundingsource, err := c.Fs.CreateBankAccount(
			ctx,
			generateCreateBankAccountArgs(withIdentityID(userID), withAccountID(acc.ID)),
		)
		if err != nil {
			t.Fatal(err)
		}
		fundingsource1, err := c.Fs.CreateBankAccount(
			ctx,
			generateCreateBankAccountArgs(withIdentityID(userID), withAccountID(acc.ID)),
		)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("returns a list of all the users funding sources", func(tt *testing.T) {
			var fundingsources []FundingSource
			err = crdbsqlx.ExecuteTx(ctx, c.Db, nil, func(tx *sqlx.Tx) error {
				_fs, err := c.Fs.GetByAccountId(ctx, tx, acc.ID)
				if err != nil {
					return err
				}
				fundingsources = _fs
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}

			fundingSourcesIDs := []string{fundingsources[1].ID, fundingsources[0].ID}

			assert.Equal(tt, 2, len(fundingsources))
			assert.Contains(tt, fundingSourcesIDs, fundingsource.ID)
			assert.Contains(tt, fundingSourcesIDs, fundingsource1.ID)
		})

		t.Run("returns an empty list if a user has no funding sources", func(tt *testing.T) {
			otherUserID := uuid.NewString()
			otherAcc, err := NewAccount(c, &onboarding.CreateAccountArgs{
				IdentityID:   otherUserID,
				FirstName:    faker.FirstName(),
				LastName:     faker.LastName(),
				MobileNumber: faker.E164PhoneNumber(),
				Email:        faker.Email(),
				Country:      "US",
			})
			if err != nil {
				t.Fatal(err)
			}

			var fundingsources []FundingSource
			err = crdbsqlx.ExecuteTx(ctx, c.Db, nil, func(tx *sqlx.Tx) error {
				_fs, err := c.Fs.GetByAccountId(ctx, tx, otherAcc.ID)
				if err != nil {
					return err
				}
				fundingsources = _fs
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(tt, 0, len(fundingsources))
		})
	})
}

func generateCreateBankAccountArgs(opts ...func(*CreateBankAccountArgs)) *CreateBankAccountArgs {
	args := &CreateBankAccountArgs{
		IdentityID:    uuid.NewString(),
		AccountID:     uuid.NewString(),
		Name:          faker.Name(),
		AccountNumber: faker.CCNumber(),
		RoutingNumber: faker.CCNumber(),
		Institution:   faker.Name(),
		Type:          "cheque",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withAccountID(id string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.AccountID = id
	}
}

func withIdentityID(id string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.IdentityID = id
	}
}

func withName(name string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.Name = name
	}
}

func withType(_type string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.Type = _type
	}
}

func withAccountNumber(num string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.AccountNumber = num
	}
}

func withRoutingNumber(num string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.RoutingNumber = num
	}
}

func withInstitution(name string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.Institution = name
	}
}
