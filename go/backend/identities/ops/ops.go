package ops

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/identities/platforms"
)

const cols = ` id, wallet_id, platform, handle, state, public, proof `

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
	err := b.DB().SelectContext(ctx, &res, fmt.Sprintf("SELECT %s FROM identities WHERE wallet_id=$1 AND public=true", cols), walletID)
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

	p, err := platforms.Get(args.Platform)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identities.ErrInvalidArgument, err)
	}

	id := uuid.NewString()
	code := p.NewVerifyCode()

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
