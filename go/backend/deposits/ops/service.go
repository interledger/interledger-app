package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/deposits/flows"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

/*type ServiceArgs struct {
	Db *sqlx.DB               `validate:"required"`
	As accounts.Service       `validate:"required"`
	Is identity.Service       `validate:"required"`
	Fs fundingsources.Service `validate:"required"`
	Tp client.Client          `validate:"required"`
}
*/
/*type service struct {
	b         Backends
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Service
	is        identity.Service
	fs        fundingsources.Service
	tp        client.Client
}*/

/*func NewService(b Backends) (deposits.Client, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", deposits.ErrInvalidArgument, err.Error())
	}

	return &service{
		validator: v,
		b:         b,
		db:        args.Db,
		as:        args.As,
		is:        args.Is,
		fs:        args.Fs,
		tp:        args.Tp,
	}, nil
}*/

func InitiateDeposit(ctx context.Context, b Backends, args *deposits.InitiateDepositArgs) (*deposits.Deposit, error) {
	if err := b.Validator().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", deposits.ErrInvalidArgument, err.Error())
	}
	/*
		The flow should be as follows
		* Check if existing deposit already
		* Checks on Identity, Account, FS
		* Create Deposit Object (accountId, funding source etc)
		* Kickoff workflow
	*/
	id, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", deposits.ErrInternal, err.Error())
	}
	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", deposits.ErrInternal, err.Error())
	}
	fundingSource, err := b.FundingSources().Get(ctx, args.FundingSourceID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", deposits.ErrInternal, err.Error())
	}
	if fundingSource.AccountID != acc.ID {
		return nil, deposits.ErrNotFound
	}
	if fundingSource.VerificationState != "verified" {
		return nil, deposits.ErrUnverifiedFundingSource
	}

	if !b.Accounts().CanMakeDeposit(acc, id.ID) {
		return nil, deposits.ErrUnauthorized
	}

	// TODO should this be an idempotent key?
	var deposit deposits.Deposit
	err = b.DB().Get(&deposit, `INSERT INTO deposits
		(account_id, funding_source_id, amount, state) VALUES ($1, $2, $3, $4)
		RETURNING *;
		`, acc.ID, fundingSource.ID, args.Amount, deposits.Created)
	if err != nil {
		return nil, fmt.Errorf("failed to insert into db %s %w", err.Error(), deposits.ErrInternal)
	}

	// TODO add mechanism to handle if deposit is created but workflow is not
	workflowOptions := client.StartWorkflowOptions{
		ID:                    "deposit_" + deposit.ID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	_, err = b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, flows.DepositWorkflow, deposit.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert execute workflow %s %w", err.Error(), deposits.ErrInternal)
	}

	return &deposit, nil
}

func Get(ctx context.Context, b Backends, id string) (*deposits.Deposit, error) {
	var deposit deposits.Deposit
	err := b.DB().GetContext(ctx, &deposit, `select * from deposits where id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}

func SetState(ctx context.Context, b Backends, id string, state deposits.State) error {

	deposit, err := Get(ctx, b, id)
	if err != nil {
		return err
	}

	// TODO add checks for legitimate state changes

	_, err = b.DB().ExecContext(ctx, "update deposits set state = $1 where id = $2", state.String(), deposit.ID)
	if err != nil {
		return err
	}

	return nil
}
