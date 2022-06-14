package workflow

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
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
		fundingsourceService fundingsources.Service
	}

	ActivityArgs struct {
		Mx                   _mx.Service            `validate:"required"`
		Unit                 _unit.Service          `validate:"required"`
		AccountService       accounts.Service       `validate:"required"`
		IdentityService      identity.Service       `validate:"required"`
		FundingSourceService fundingsources.Service `validate:"required"`
	}
)

func NewActivity(args *ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, err
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
	mxAccountGuid, err := a.mx.GetSelectedAccountGuid(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return "", err
	}

	return mxAccountGuid, nil
}

func (a *Activity) CreateMxAccount(
	ctx context.Context,
	id string, // id that will be used for our database row
	accountID string,
	mxUserGuid string,
	mxMemberGuid string,
	mxAccountGuid string, // id the mx generates
) (string, error) {
	return "", nil
}

func (a *Activity) StartIdentityAggregation(
	ctx context.Context,
	mxAccountID string,
) error {
	_, err := a.mx.StartIdentityAggregation(ctx, mxAccountID)
	if err != nil {
		return err
	}

	return nil
}

// Uses a ticker and go routine to poll every second for up to 10 seconds for aggregation to be
// complete. Temporal recommends this over failing the activity task for polling at short intervals.
func (a *Activity) WaitForIdentityAggregation(
	ctx context.Context,
	mxAccountID string,
) error {
	now := time.Now()
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		elapsed := time.Since(now)
		if elapsed > 10*time.Second {
			return errors.New("Timed out waiting for identity aggregation.")
		}

		member, err := a.mx.GetMemberStatus(ctx, mxAccountID)
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
	mxAccountID string,
) error {
	err := a.mx.VerifyOwnership(ctx, mxAccountID)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) CreateUnitCounterParty(ctx context.Context, mxAccountID string) error {
	return nil
}

func (a *Activity) CreateFundingSource(ctx context.Context, mxAccountID string) error {
	return nil
}
