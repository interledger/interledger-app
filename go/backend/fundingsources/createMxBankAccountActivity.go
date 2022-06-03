package fundingsources

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	_unit "gitlab.com/fynbos/backend/providers/unit"
)

type (
	Activity struct {
		validator            *validator.Validate
		unit                 _unit.Service
		mx                   mx.Service
		accountService       accounts.Service
		identityService      identity.Service
		fundingsourceService Service
	}

	ActivityArgs struct {
		Mx                   _mx.Service      `validate:"required"`
		Unit                 _unit.Service    `validate:"required"`
		AccountService       accounts.Service `validate:"required"`
		IdentityService      identity.Service `validate:"required"`
		FundingSourceService Service          `validate:"required"`
	}
)

func NewActivity(args *ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &Activity{
		validator:            v,
		unit:                 args.Unit,
		mx:                   args.Mx,
		accountService:       args.AccountService,
		identityService:      args.IdentityService,
		fundingsourceService: args.FundingSourceService,
	}, nil
}

func (a *Activity) GetSelectedMxAccountGuid(
	ctx context.Context,
	mxUserGuid string,
	mxMemberGuid string,
) (string, error) {
	return "", nil
}

func (a *Activity) CreateMxAccount(
	ctx context.Context,
	fundingSourceID string,
	mxUserGuid string,
	mxMemberGuid string,
	mxAccountGuid string,
) error {
	fs, err := a.fundingsourceService.Get(ctx, fundingSourceID)
	if err != nil {
		return err
	}

	_, err = a.mx.CreateMxFundingSource(ctx, &_mx.CreateMxFundingSourceArgs{
		ID:            fundingSourceID, // we map 1-1 to fundingsource
		AccountID:     fs.AccountID,
		MxUserGuid:    mxUserGuid,
		MxMemberGuid:  mxMemberGuid,
		MxAccountGuid: mxAccountGuid,
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) StartIdentityAggregation(
	ctx context.Context,
	mxUserGuid string,
	mxMemberGuid string,
) error {
	return nil
}

func (a *Activity) WaitForIdentityAggregation(
	ctx context.Context,
	mxUserGuid string,
	mxMemberGuid string,
) error {
	return nil
}

func (a *Activity) VerifyOwnership(
	ctx context.Context,
	fundingSourceID string,
	mxUserGuid string,
	mxMemberGuid string,
	mxAccountGuid string,
) error {
	return nil
}

func (a *Activity) GetBankAccountInfo(
	ctx context.Context,
	mxUserGuid string,
	mxAccountGuid string,
) (*AccountInfo, error) {
	return nil, nil
}

func (a *Activity) SetMask(ctx context.Context, fundingsourceID string, accountNumber string) error {
	return nil
}

func (a *Activity) CreateUnitCounterParty(ctx context.Context, fundingsourceID string, accountInfo *AccountInfo) error {
	return nil
}
