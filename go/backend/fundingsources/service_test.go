package fundingsources

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/identity/noop"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

func TestFundingSources(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}
	defer crdb.Container.Terminate(ctx)

	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}
	defer logger.Sync()

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()
	cs := country.NewService()
	provider := noop.NewMockProvider(ctrl)
	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		NoopProvider:   provider,
	})
	fs, err := NewService(&ServiceArgs{Identity: is})
	fs = NewLoggingService(fs, logger)

	s.Run("create funding source", func(t *testing.T) {
		var identity *_identity.Identity
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			id, err := is.Create(ctx, tx, _identity.CreateArgs{
				ID:           uuid.NewString(),
				FirstName:    faker.Name(),
				LastName:     faker.Name(),
				MobileNumber: faker.E164PhoneNumber(),
				Email:        faker.Email(),
				Country:      "US",
			})
			if err != nil {
				return err
			}

			identity = id
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("validates arguments", func(tt *testing.T) {
			type Scenario struct {
				Name          string
				Args          *CreateArgs
				ExpectedError string
			}
			scenarios := []Scenario{
				{
					Name:          "IdentityID is required to create funding source",
					Args:          generateCreateArgs(withIdentityID("")),
					ExpectedError: "Key: 'CreateArgs.IdentityID' Error:Field validation for 'IdentityID' failed on the 'required' tag",
				},
				{
					Name:          "IdentityID must exist to create funding source",
					Args:          generateCreateArgs(withIdentityID(uuid.NewString())),
					ExpectedError: "Identity must exist to create funding source.",
				},
				{
					Name:          "Name is required to create funding source",
					Args:          generateCreateArgs(withName("")),
					ExpectedError: "Key: 'CreateArgs.Name' Error:Field validation for 'Name' failed on the 'required' tag",
				},
				{
					Name:          "Mask is required to create funding source",
					Args:          generateCreateArgs(withMask("")),
					ExpectedError: "Key: 'CreateArgs.Mask' Error:Field validation for 'Mask' failed on the 'required' tag",
				},
				{
					Name:          "VerificationState is required to create funding source",
					Args:          generateCreateArgs(withVerificationState("")),
					ExpectedError: "Key: 'CreateArgs.VerificationState' Error:Field validation for 'VerificationState' failed on the 'required' tag",
				},
				{
					Name:          "Type must be one of noop required to create funding source",
					Args:          generateCreateArgs(withType("")),
					ExpectedError: "Key: 'CreateArgs.Type' Error:Field validation for 'Type' failed on the 'oneof' tag",
				},
				{
					Name:          "TypeID is required to create funding source",
					Args:          generateCreateArgs(withTypeID("")),
					ExpectedError: "Key: 'CreateArgs.TypeID' Error:Field validation for 'TypeID' failed on the 'required' tag",
				},
				{
					Name:          "SubType is required to create funding source",
					Args:          generateCreateArgs(withSubType("")),
					ExpectedError: "Key: 'CreateArgs.SubType' Error:Field validation for 'SubType' failed on the 'required' tag",
				},
			}

			for _, scenario := range scenarios {
				var fundingSource *FundingSource
				err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
					_fundingSource, err := fs.Create(ctx, tx, scenario.Args)
					if err != nil {
						return err
					}

					fundingSource = _fundingSource
					return nil
				})
				if err == nil {
					tt.Fatal(scenario.Name)
				}

				assert.Equal(tt, scenario.ExpectedError, err.Error())
				assert.Nil(tt, fundingSource)
			}
		})

		t.Run("creates db record", func(tt *testing.T) {
			var fundingsource *FundingSource
			args := generateCreateArgs(withIdentityID(identity.ID))
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_fs, err := fs.Create(ctx, tx, args)
				if err != nil {
					return err
				}
				fundingsource = _fs
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, args.Name, fundingsource.Name)
			assert.Equal(tt, args.Mask, fundingsource.Mask)
			assert.Equal(tt, args.IdentityID, identity.ID)
			assert.Equal(tt, args.Type, fundingsource.Type)
			assert.Equal(tt, args.TypeID, fundingsource.TypeID)
			assert.Equal(tt, args.SubType, fundingsource.SubType)
			assert.Equal(tt, args.VerificationState, fundingsource.VerificationState)

			var fetchedFS *FundingSource
			err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_fs, err := fs.Get(ctx, tx, fundingsource.ID)
				if err != nil {
					return err
				}
				fetchedFS = _fs
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(tt, args.Name, fetchedFS.Name)
			assert.Equal(tt, args.Mask, fetchedFS.Mask)
			assert.Equal(tt, args.IdentityID, identity.ID)
			assert.Equal(tt, args.Type, fetchedFS.Type)
			assert.Equal(tt, args.TypeID, fetchedFS.TypeID)
			assert.Equal(tt, args.SubType, fetchedFS.SubType)
			assert.Equal(tt, args.VerificationState, fetchedFS.VerificationState)
		})
	})
}

func generateCreateArgs(opts ...func(*CreateArgs)) *CreateArgs {
	args := &CreateArgs{
		IdentityID:        uuid.NewString(),
		Name:              faker.Name(),
		Mask:              "****",
		VerificationState: "pending",
		Type:              "noop",
		TypeID:            uuid.NewString(),
		SubType:           "cheque",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withIdentityID(id string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.IdentityID = id
	}
}

func withName(name string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.Name = name
	}
}

func withMask(mask string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.Mask = mask
	}
}

func withVerificationState(state string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.VerificationState = state
	}
}

func withTypeID(typeID string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.TypeID = typeID
	}
}

func withType(_type string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.Type = _type
	}
}

func withSubType(subtype string) func(args *CreateArgs) {
	return func(args *CreateArgs) {
		args.SubType = subtype
	}
}
