package mx

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
	_unit "gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/sdk/temporal"
)

type (
	Activity struct {
		validator            *validator.Validate
		unit                 _unit.Service
		mx                   Service
		accountService       accounts.Service
		identityService      identity.Service
		fundingsourceService fundingsources.Service
	}

	ActivityArgs struct {
		Mx                   Service                `validate:"required"`
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
	fundingsourceID string,
	accountID string,
	userGuid string, // from mx
	memberGuid string, // from mx
	accountGuid string, // from mx
) error {
	_, err := a.mx.CreateAccount(ctx, &CreateAccountArgs{
		Guid:            accountGuid,
		UserGuid:        userGuid,
		MemberGuid:      memberGuid,
		AccountID:       accountID,
		FundingsourceID: fundingsourceID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) StartIdentityAggregation(
	ctx context.Context,
	mxAccountGuid string,
) error {
	_, err := a.mx.StartIdentityAggregation(ctx, mxAccountGuid)
	if err != nil {
		return err
	}

	return nil
}

// Uses a ticker and go routine to poll for aggregation to be complete. Temporal recommends this
// over failing the activity task for polling at short intervals.
// This will always perform at least one api call. Thereafter it will retry up to maxRetries.
func (a *Activity) WaitForIdentityAggregation(
	ctx context.Context,
	mxAccountGuid string,
	maxRetries uint8,
	pollInterval time.Duration,
) error {
	ticker := time.NewTicker(pollInterval)
	retries := uint8(0)
	for range ticker.C {
		if retries > maxRetries {
			return errors.New("Timed out waiting for identity aggregation.")
		}

		member, err := a.mx.GetMemberStatus(ctx, mxAccountGuid)
		if err != nil {
			retries += 1
			continue
		}

		if !member.IsBeingAggregated {
			break
		}

		retries += 1
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

// This will first validate that the associated mxAccount, account, user and unit customer exist.
// It then fetches the account routing information from MX and uses it to create a counterparty
// on unit.
// Note: the sha256 of the fundingsourceID is used as the idempotency key when calling out to
// Unit.
func (a *Activity) CreateUnitCounterParty(ctx context.Context, mxAccountGuid string) error {
	mxAccount, err := a.mx.GetAccount(ctx, mxAccountGuid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}

		if errors.Is(err, ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}

		// retry here as this may be network related etc.
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	acc, err := a.accountService.Get(ctx, mxAccount.AccountID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", ErrInternal, err)
		if errors.Is(err, accounts.ErrNotFound) || errors.Is(err, accounts.ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(
				wrappedError.Error(),
				"ErrInternal",
				wrappedError,
			)
		}

		// retry here as this may be network related etc.
		return wrappedError
	}

	user, err := a.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", ErrInternal, err)
		if errors.Is(err, identity.ErrNotFound) || errors.Is(err, identity.ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(
				wrappedError.Error(),
				"ErrInternal",
				wrappedError,
			)
		}

		// retry here as this may be network related etc.
		return wrappedError
	}

	unitCustomer, err := a.unit.GetCustomerByIdentityID(ctx, user.ID)
	// TODO: unit needs to have ErrNotFound
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	// perform this just before creating the counter party as we get charged for Mx api calls.
	accountNumbers, err := a.mx.ReadAccount(ctx, mxAccount.Guid)
	if err != nil {
		// at this stage we know the mx account must exist so keep retrying.
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	idempotencyKey := sha256.Sum256([]byte(mxAccount.FundingsourceID))
	_, err = a.unit.CreateCounterParty(ctx, &_unit.CreateCounterPartyArgs{
		Name:            fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		RoutingNumber:   accountNumbers.RoutingNumber,
		AccountNumber:   accountNumbers.AccountNumber,
		AccountType:     accountNumbers.Type,
		Type:            "person",
		IdempotencyKey:  string(idempotencyKey[0:]),
		UnitCustomerID:  unitCustomer.ID,
		FundingsourceID: mxAccount.FundingsourceID,
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

// This will first validate that the associated mxAccount, account and user exist before creating
// a funding source.
// The account number is then retrieved from MX and the last 4 digits used as the funding source
// mask.
func (a *Activity) CreateFundingSource(
	ctx context.Context,
	mxAccountGuid string,
	fundingsourceName string,
) error {
	mxAccount, err := a.mx.GetAccount(ctx, mxAccountGuid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}

		if errors.Is(err, ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}

		// retry here as this may be network related etc.
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	acc, err := a.accountService.Get(ctx, mxAccount.AccountID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", ErrInternal, err)
		if errors.Is(err, accounts.ErrNotFound) || errors.Is(err, accounts.ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(
				wrappedError.Error(),
				"ErrInternal",
				wrappedError,
			)
		}

		// retry here as this may be network related etc.
		return wrappedError
	}

	user, err := a.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", ErrInternal, err)
		if errors.Is(err, identity.ErrNotFound) || errors.Is(err, identity.ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(
				wrappedError.Error(),
				"ErrInternal",
				wrappedError,
			)
		}

		// retry here as this may be network related etc.
		return wrappedError
	}

	// calling this in the activity so it's not accidently stored in temporal state.
	accountRoutingInfo, err := a.mx.ReadAccount(ctx, mxAccount.Guid)
	if err != nil {
		// at this stage we know the mx account must exist so keep retrying.
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	start := len(accountRoutingInfo.AccountNumber) - 4
	if start < 0 {
		start = 0
	}
	// we use the last 4 digits. If it less than 4 digits then we use the whole thing.
	mask := accountRoutingInfo.AccountNumber[start:]

	_, err = a.fundingsourceService.Create(ctx, &fundingsources.CreateArgs{
		IdentityID:        user.ID,
		AccountID:         mxAccount.AccountID,
		Name:              fundingsourceName,
		Mask:              mask,
		VerificationState: string(fundingsources.VERIFIED), // TODO: remove verification state
		Type:              "mx",
		SubType:           "bank",
		ID:                mxAccount.FundingsourceID,
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

// This will start the balance aggregation. You will have to get the member's status to see when the
// process has been completed.
func (a *Activity) StartBalanceAggregation(ctx context.Context, mxAccountGuid string) error {
	member, err := a.mx.StartBalanceAggregation(ctx, mxAccountGuid)
	if errors.Is(err, ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return err // retryable
	}

	if !CanAggregate(member.ConnectionStatus) {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("Cannot aggregrate member with status=%s", member.ConnectionStatus),
			"ErrNotFound",
			err,
		)
	}

	return nil
}
