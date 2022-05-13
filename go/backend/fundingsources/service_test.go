package fundingsources

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
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
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
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
				Args:                 generateCreateBankAccountArgs(withIdentityID(id.ID), withAccountID(uuid.NewString())),
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
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(id.ID),
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
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(id.ID),
			withAccountID(acc.ID),
		)
		bankAccount, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "required", bankAccount.VerificationState)

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      id.ID,
			FundingSourceID: bankAccount.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "verified", fs.VerificationState)
	})

	s.Run("returns not found if funding source does not belong to user", func(t *testing.T) {
		otherUser := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		otherUserID, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
			ID:           otherUser.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        otherUser.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		myID, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		myAcc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: myID.ID,
			Country:    myID.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: otherUserID.ID,
			Country:    otherUserID.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(myID.ID),
			withAccountID(myAcc.ID),
		)
		bankAccount, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "required", bankAccount.VerificationState)

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      otherUserID.ID,
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
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      id.ID,
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
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		fundingsource, err := c.Fs.CreateBankAccount(
			ctx,
			generateCreateBankAccountArgs(withIdentityID(id.ID), withAccountID(acc.ID)),
		)
		if err != nil {
			t.Fatal(err)
		}
		fundingsource1, err := c.Fs.CreateBankAccount(
			ctx,
			generateCreateBankAccountArgs(withIdentityID(id.ID), withAccountID(acc.ID)),
		)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("returns a list of all the users funding sources", func(tt *testing.T) {
			fs, err := c.Fs.GetByAccountId(ctx, acc.ID)
			if err != nil {
				tt.Fatal(err)
			}

			fundingSourcesIDs := []string{fs[1].ID, fs[0].ID}

			assert.Equal(tt, 2, len(fs))
			assert.Contains(tt, fundingSourcesIDs, fundingsource.ID)
			assert.Contains(tt, fundingSourcesIDs, fundingsource1.ID)
		})

		t.Run("returns an empty list if a user has no funding sources", func(tt *testing.T) {
			user := &_user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}

			id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
			otherAcc, err := NewAccount(c, &onboarding.CreateAccountArgs{
				IdentityID: id.ID,
				Country:    id.Country,
			})
			if err != nil {
				t.Fatal(err)
			}

			fs, err := c.Fs.GetByAccountId(ctx, otherAcc.ID)
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(tt, 0, len(fs))
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
