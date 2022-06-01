package fundingsources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/sdk/temporal"
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
	mxFundingSourceID string,
) error {
	if _, err := a.mx.StartIdentityAggregation(ctx, mxFundingSourceID); err != nil {
		return err
	}

	return nil
}

// Uses a ticker and go routine to poll every second for up to 10 seconds for aggregation to be
// complete. Temporal recommends this over failing the activity task for polling at short intervals.
func (a *Activity) WaitForIdentityAggregation(
	ctx context.Context,
	mxFundingSourceID string,
) error {
	now := time.Now()
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		elapsed := time.Since(now)
		if elapsed > 10*time.Second {
			return errors.New("Timed out waiting for identity aggregation.")
		}

		member, err := a.mx.GetMemberStatus(ctx, mxFundingSourceID)
		if err != nil {
			continue
		}

		if !member.IsBeingAggregated {
			return nil
		}
	}

	return nil
}

func (a *Activity) VerifyOwnership(
	ctx context.Context,
	fundingSourceID string,
	identityID string,
) error {
	_, err := a.fundingsourceService.VerifyMxBankAccount(ctx, identityID, fundingSourceID)
	if errors.Is(err, ErrUnauthorized) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrUnauthorized", err)
	} else if errors.Is(err, ErrInvalidArgument) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
	} else if err != nil {
		return temporal.NewApplicationError(err.Error(), "ErrInternal")
	}

	return nil
}

func (a *Activity) SetMask(ctx context.Context, fundingsourceID string) error {
	err := a.fundingsourceService.SetMxFundingSourceMask(ctx, fundingsourceID)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) CreateUnitCounterParty(ctx context.Context, fundingsourceID string) error {
	_, err := a.fundingsourceService.CreateUnitCounterPartyFromMxAccount(ctx, fundingsourceID)
	if err != nil {
		return err
	}

	return nil
}
