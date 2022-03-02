package fundingsources

import (
	"context"
	"fmt"
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
		c.Cleanup()
	})

	s.Run("validates create bank account arguments", func(t *testing.T) {
		fmt.Println("starting scenarios")
		type Scenario struct {
			Name          string
			Args          *CreateBankAccountArgs
			ExpectedError string
		}
		scenarios := []Scenario{
			{
				Name:          "IdentityID is required to create bank account",
				Args:          generateCreateBankAccountArgs(withIdentityID("")),
				ExpectedError: "Key: 'CreateBankAccountArgs.IdentityID' Error:Field validation for 'IdentityID' failed on the 'required' tag",
			},
			{
				Name:          "IdentityID must exist to create bank account",
				Args:          generateCreateBankAccountArgs(withIdentityID(uuid.NewString())),
				ExpectedError: "Identity must exist to create funding source.",
			},
			{
				Name:          "Name is required to create bank account",
				Args:          generateCreateBankAccountArgs(withName("")),
				ExpectedError: "Key: 'CreateBankAccountArgs.Name' Error:Field validation for 'Name' failed on the 'required' tag",
			},
			{
				Name:          "Institution is required to create bank account",
				Args:          generateCreateBankAccountArgs(withInstitution("")),
				ExpectedError: "Key: 'CreateBankAccountArgs.Institution' Error:Field validation for 'Institution' failed on the 'required' tag",
			},
			{
				Name:          "AccountNumber is required to create bank account",
				Args:          generateCreateBankAccountArgs(withAccountNumber("")),
				ExpectedError: "Key: 'CreateBankAccountArgs.AccountNumber' Error:Field validation for 'AccountNumber' failed on the 'required' tag",
			},
			{
				Name:          "RoutingNumber is required to create bank account",
				Args:          generateCreateBankAccountArgs(withRoutingNumber("")),
				ExpectedError: "Key: 'CreateBankAccountArgs.RoutingNumber' Error:Field validation for 'RoutingNumber' failed on the 'required' tag",
			},
			{
				Name:          "Type must be one of noop required to create bank account",
				Args:          generateCreateBankAccountArgs(withType("")),
				ExpectedError: "Key: 'CreateBankAccountArgs.Type' Error:Field validation for 'Type' failed on the 'required' tag",
			},
		}

		for _, scenario := range scenarios {
			fs, err := c.Fs.CreateBankAccount(c.Ctx, scenario.Args)
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.Equal(t, scenario.ExpectedError, err.Error())
			assert.Nil(t, fs)
		}
	})

	s.Run("creates bank account", func(t *testing.T) {
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

		args := generateCreateBankAccountArgs(withIdentityID(userID))
		fs, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.Name, fs.Name)
		assert.NotEqual(t, args.AccountNumber, fs.Mask)
		assert.Equal(t, "noop", fs.Type)
		assert.Equal(t, args.Type, fs.SubType)
		assert.Equal(t, userID, fs.IdentityID)
		assert.Equal(t, "required", fs.VerificationState)
	})

	s.Run("verifies bank account", func(t *testing.T) {
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
		args := generateCreateBankAccountArgs(withIdentityID(userID))
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
		_, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID:   myID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
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
		args := generateCreateBankAccountArgs(withIdentityID(myID))
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

		assert.EqualError(t, err, "Funding source not found.")
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
		assert.EqualError(t, err, "Funding source not found.")
	})

	s.Run("get user's funding sources", func(t *testing.T) {
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
		fundingsource, err := c.Fs.CreateBankAccount(ctx, generateCreateBankAccountArgs(withIdentityID(userID)))
		if err != nil {
			t.Fatal(err)
		}
		fundingsource1, err := c.Fs.CreateBankAccount(ctx, generateCreateBankAccountArgs(withIdentityID(userID)))
		if err != nil {
			t.Fatal(err)
		}

		t.Run("returns a list of all the users funding sources", func(tt *testing.T) {
			var fundingsources []*FundingSource
			err = crdbsqlx.ExecuteTx(ctx, c.Db, nil, func(tx *sqlx.Tx) error {
				_fs, err := c.Fs.GetByIdentityId(ctx, tx, userID)
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
			_, err := NewAccount(c, &onboarding.CreateAccountArgs{
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

			var fundingsources []*FundingSource
			err = crdbsqlx.ExecuteTx(ctx, c.Db, nil, func(tx *sqlx.Tx) error {
				_fs, err := c.Fs.GetByIdentityId(ctx, tx, otherUserID)
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
