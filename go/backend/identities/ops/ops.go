package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/identities/platforms"
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
)

const cols = ` id, wallet_id, platform, handle, state, public, proof, code `

func List(ctx context.Context, b Backends, walletID string) ([]identities.Identity, error) {
	var res []identities.Identity
	err := b.DB().SelectContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE wallet_id=$1", cols), walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return res, nil
}

func ListPublic(ctx context.Context, b Backends, walletID string) ([]identities.Identity, error) {
	var res []identities.Identity
	err := b.DB().SelectContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE wallet_id=$1 AND public=true AND state=$2", cols),
		walletID, identities.StateVerified)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return res, nil
}

func Add(ctx context.Context, b Backends, args identities.AddArgs) (*identities.VerifyInstructions, error) {
	err := b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	p, err := platforms.Get(b, args.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	var existing identities.Identity
	err = b.DB().GetContext(ctx, &existing, fmt.Sprintf("SELECT %s FROM identities WHERE platform=$1 AND lower(handle)=$2 AND state=$3", cols),
		args.Platform, strings.ToLower(args.Handle), identities.StateVerified)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	if existing.ID != "" {
		return nil, fmt.Errorf("%w %s handle %s has already been verified", identities.ErrInvalidArgument, args.Platform, args.Handle)
	}

	id := uuid.NewString()
	code, err := p.NewVerifyCode(ctx, &platforms.NewVerifyCodeArgs{
		WalletID:   args.WalletID,
		Identifier: args.Handle,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO identities(id, wallet_id, platform, handle, state, public, code, proof) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		id, args.WalletID, args.Platform, args.Handle, identities.StateUnverified, args.Public, code, "")
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return &identities.VerifyInstructions{
		IdentityID:   id,
		Code:         code,
		Instructions: p.VerifyInstructions(),
	}, nil
}

func Get(ctx context.Context, b Backends, id string) (*identities.Identity, error) {
	var res identities.Identity
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE id=$1", cols), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", identities.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return &res, nil

}

func VerifyInstructions(ctx context.Context, b Backends, id string) (*identities.VerifyInstructions, error) {
	ident, err := Get(ctx, b, id)
	if err != nil {
		return nil, err
	}

	p, err := platforms.Get(b, ident.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	return &identities.VerifyInstructions{
		IdentityID:   id,
		Code:         ident.VerificationCode,
		Instructions: p.VerifyInstructions(),
	}, nil
}

func Delete(ctx context.Context, b Backends, id, walletID string) error {
	res, err := b.DB().ExecContext(ctx, "DELETE FROM identities WHERE id=$1 AND wallet_id=$2", id, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	if rows != 1 {
		return fmt.Errorf("%w wrong number of rows deleted (%d)", identities.ErrInternal, rows)
	}

	return err
}

func SetPublic(ctx context.Context, b Backends, id, walletID string, public bool) (*identities.Identity, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE identities SET public=$1, updated_at=now() WHERE id=$2 AND wallet_id=$3", public, id, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return Get(ctx, b, id)
}

func StartVerification(ctx context.Context, b Backends, id, proof string) (*identities.Identity, error) {
	ident, err := Get(ctx, b, id)
	if err != nil {
		return nil, err
	}

	p, err := platforms.Get(b, ident.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	workflowOptions := temporal_client.StartWorkflowOptions{
		ID:                       "identities_verify_" + id,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24, // Workflow has a day to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	wf, err := b.Temporal().ExecuteWorkflow(ctx, workflowOptions, p.VerifyWorkflow(), id, proof)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}
	err = wf.Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return Get(ctx, b, id)
}
