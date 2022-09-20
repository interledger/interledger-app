package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"go.temporal.io/sdk/temporal"
)

type (
	Activity struct {
		b Backends
	}
)

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

func (a *Activity) GetSelectedMxAccountGuid(
	ctx context.Context,
	mxUserGuid string,
	mxMemberGuid string,
) (string, error) {
	mxAccountGuid, err := a.b.MX().GetSelectedAccountGuid(ctx, mxUserGuid, mxMemberGuid)
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
	_, err := a.b.MX().CreateAccount(ctx, mx.CreateAccountArgs{
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
	mxUserGuid,
	mxMemberGuid string,
) error {
	_, err := a.b.MX().StartIdentityAggregation(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return err
	}

	return nil
}

type WaitForAggregationArgs struct {
	MaxRetries   uint8
	PollInterval time.Duration
	MxMemberGuid string
	MxUserGuid   string
}

// Uses a ticker and go routine to poll for aggregation to be complete. Temporal recommends this
// over failing the activity task for polling at short intervals.
// This will always perform at least one api call. Thereafter it will retry up to maxRetries.
func (a *Activity) WaitForAggregation(
	ctx context.Context,
	args *WaitForAggregationArgs,
) error {
	ticker := time.NewTicker(args.PollInterval)
	retries := uint8(0)
	for range ticker.C {
		if retries > args.MaxRetries {
			return errors.New("Timed out waiting for aggregation.")
		}

		member, err := a.b.MX().GetMemberStatus(ctx, args.MxUserGuid, args.MxMemberGuid)
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
	args mx.VerifyOwnershipArgs,
) error {
	err := a.b.MX().VerifyOwnership(ctx, args)
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
	mxAccount, err := a.b.MX().GetAccount(ctx, mxAccountGuid)
	if err != nil {
		if errors.Is(err, mx.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}

		if errors.Is(err, mx.ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}

		// retry here as this may be network related etc.
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	acc, err := a.b.Accounts().Get(ctx, mxAccount.AccountID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", mx.ErrInternal, err)
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

	_, err = a.b.Identity().Get(ctx, acc.IdentityID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", mx.ErrInternal, err)
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

	// perform this just before creating the counter party as we get charged for Mx api calls.
	_, err = a.b.MX().ReadAccount(ctx, mxAccount.Guid)
	if err != nil {
		// at this stage we know the mx account must exist so keep retrying.
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
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
	mxAccount, err := a.b.MX().GetAccount(ctx, mxAccountGuid)
	if err != nil {
		if errors.Is(err, mx.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}

		if errors.Is(err, mx.ErrInvalidArgument) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}

		// retry here as this may be network related etc.
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	acc, err := a.b.Accounts().Get(ctx, mxAccount.AccountID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", mx.ErrInternal, err)
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

	user, err := a.b.Identity().Get(ctx, acc.IdentityID)
	if err != nil {
		wrappedError := fmt.Errorf("%w %s", mx.ErrInternal, err)
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
	accountRoutingInfo, err := a.b.MX().ReadAccount(ctx, mxAccount.Guid)
	if err != nil {
		// at this stage we know the mx account must exist so keep retrying.
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	start := len(accountRoutingInfo.AccountNumber) - 4
	if start < 0 {
		start = 0
	}
	// we use the last 4 digits. If it less than 4 digits then we use the whole thing.
	mask := accountRoutingInfo.AccountNumber[start:]

	_, err = a.b.FundingSources().Create(ctx, &fundingsources.CreateArgs{
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
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return nil
}

// This will start the balance aggregation. You will have to get the member's status to see when the
// process has been completed.
func (a *Activity) StartBalanceAggregation(ctx context.Context, mxAccountGuid string) error {
	member, err := a.b.MX().StartBalanceAggregation(ctx, mxAccountGuid)
	if errors.Is(err, mx.ErrNotFound) {
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

func (a *Activity) GetMxAccountByFundingsource(
	ctx context.Context,
	fundingsourceID string,
) (*mx.Account, error) {
	mxAcc, err := a.b.MX().GetAccountByFundingsource(ctx, fundingsourceID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err // retryable
	}

	return mxAcc, nil
}

func (a *Activity) GetMxAccountBalance(
	ctx context.Context,
	mxAccountGuid string,
) (*mx.AccountBalance, error) {
	balance, err := a.b.MX().GetAccountBalance(ctx, mxAccountGuid)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err // retryable
	}

	return balance, nil
}

func CanAggregate(memberConnectionStatus string) bool {
	switch memberConnectionStatus {
	case external.CONNECTION_STATUS_CONNECTED,
		external.CONNECTION_STATUS_CREATED,
		external.CONNECTION_STATUS_DEGRADED,
		external.CONNECTION_STATUS_DISCONNECTED,
		external.CONNECTION_STATUS_EXPIRED,
		external.CONNECTION_STATUS_FAILED,
		external.CONNECTION_STATUS_IMPEDED,
		external.CONNECTION_STATUS_RECONNECTED,
		external.CONNECTION_STATUS_UPDATED,
		external.CONNECTION_STATUS_DELAYED,
		external.CONNECTION_STATUS_REJECTED,
		external.CONNECTION_STATUS_RESUMED:
		return true
	default:
		return false
	}
}

func IsSavings(accountType string) bool {
	return strings.ToLower(strings.TrimSpace(accountType)) == "savings"
}
