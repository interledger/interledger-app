package workflow

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	_unit "gitlab.com/fynbos/backend/providers/unit"
)

var (
	ErrInternal = errors.New("create mx account activity: internal error.")
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
	mxAccount, err := a.mx.GetAccount(ctx, mxAccountID)
	if err != nil {
		return fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	acc, err := a.accountService.Get(ctx, mxAccount.AccountID)
	if err != nil {
		return fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	user, err := a.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	unitCustomer, err := a.unit.GetCustomerByAccountID(ctx, mxAccount.AccountID)
	if err != nil {
		return fmt.Errorf("%w No unit customer found for accountID=%s.", ErrInternal, mxAccount.AccountID)
	}

	// perform this just before creating the counter party as we get charged for Mx api calls.
	accountNumbers, err := a.mx.ReadAccount(ctx, mxAccount.Guid)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	idempotencyKey := sha256.Sum256([]byte(mxAccount.Guid))
	_, err = a.unit.CreateCounterParty(ctx, &_unit.CreateCounterPartyArgs{
		Name:            fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		RoutingNumber:   accountNumbers.RoutingNumber,
		AccountNumber:   accountNumbers.AccountNumber,
		AccountType:     accountNumbers.Type,
		Type:            "person",
		IdempotencyKey:  string(idempotencyKey[0:]),
		UnitCustomerID:  unitCustomer.ID,
		FundingsourceID: mxAccount.Guid, // TODO: confusing - refactor mx account model
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (a *Activity) CreateFundingSource(ctx context.Context, mxAccountID string) error {
	mxAccount, err := a.mx.GetAccount(ctx, mxAccountID)
	if err != nil {
		return fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	acc, err := a.accountService.Get(ctx, mxAccount.AccountID)
	if err != nil {
		return fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	user, err := a.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	// calling this in the activity so it's accidently stored in temporal state.
	accountNumbers, err := a.mx.ReadAccount(ctx, mxAccount.Guid)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	start := len(accountNumbers.AccountNumber) - 4
	if start < 0 {
		start = 0
	}
	// we use the last 4 digits. If it less than 4 digits then we use the whole thing.
	mask := accountNumbers.AccountNumber[start:]

	_, err = a.fundingsourceService.Create(ctx, &fundingsources.CreateArgs{
		IdentityID:        user.ID, // TODO: refactor ACL out of services
		AccountID:         mxAccount.AccountID,
		Name:              "string",
		Mask:              mask,
		VerificationState: string(fundingsources.VERIFIED), // TODO: remove verification state
		Type:              "mx",
		SubType:           "bank",
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}
