package mx

import "context"

type Client interface {
	CreateAccount(ctx context.Context, args CreateAccountArgs) (*Account, error)
	GetAccount(ctx context.Context, mxAccountGuid string) (*Account, error)
	GetAccountByFundingsource(ctx context.Context, fundingsourceID string) (*Account, error)
	StartIdentityAggregation(ctx context.Context, mxUserGuid, mxMemberGuid string) (*Member, error)
	GetMemberStatus(ctx context.Context, mxUserGuid, mxMemberGuid string) (*Member, error)
	// This will fetch the account owner information for the specified mx account. The identity
	// aggregation has to have been completed first.
	GetAccountOwner(ctx context.Context, args GetAccountOwnerArgs) (*AccountOwner, error)
	ReadAccount(ctx context.Context, mxAccountGuid string) (*AccountDetails, error)
	// The mx connect widget will allow the user to log into their bank and select an account.
	// They do not pass this to us on the front end and so we need to call out to find out the
	// mx account guid of the account that was selected.
	// Calling the users/:users/members/:members/account_numbers should only have the account selected
	// by the user.
	GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error)
	GetMxUserByAccountID(ctx context.Context, accountID string) (string, error)
	VerifyOwnership(ctx context.Context, args VerifyOwnershipArgs) error
	GetConnectWidget(ctx context.Context, accountID string, identityID string) (string, error)
	InitiateCreateAccount(ctx context.Context, args InitiateCreateAccountArgs) (string, error)
	// Blocks till the workflow is complete and returns the result.
	WaitForCreateAccount(ctx context.Context, fundingsourceID string) error
	InitiateCreateFundingsource(ctx context.Context, args InitiateCreateFundingsourceArgs) error
	StartBalanceAggregation(ctx context.Context, mxAccountGuid string) (*Member, error)
	GetAccountBalance(ctx context.Context, mxAccountGuid string) (*AccountBalance, error)
}
