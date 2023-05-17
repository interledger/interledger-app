package platforms

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/identities"
	"go.temporal.io/sdk/workflow"
)

type dev struct {
	platform identities.Platform
}

func newDev(platform identities.Platform) *dev {
	return &dev{platform: platform}
}

func (d *dev) VerifyWorkflow() interface{} {
	return DevVerifyWorkflow
}

func (d *dev) GenerateSignedClaim(ctx context.Context, args *SignedClaimArgs) (*GeneratedSignedClaim, error) {

	claim := identities.Claim{
		Wallet:     args.WalletID,
		Type:       "",
		Identifier: "",
		Kid:        "",
		Ctime:      0,
	}

	return &GeneratedSignedClaim{
		Claim:         claim,
		Signature:     []byte(""),
		SignatureHash: []byte(""),
	}, nil
}

func (d *dev) VerifyInstructions(ctx context.Context, args *VerifyInstructionsArgs) (string, error) {
	return `In this environment all you need to do to verify is to request it. Enjoy in a NON Production environment.`, nil
}

type DevActivity struct {
	b Backends
}

func NewDevActivity(b Backends) *DevActivity {
	return &DevActivity{b: b}
}

func DevVerifyWorkflow(ctx workflow.Context, id, proof string) (string, error) {
	var a *DevActivity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("VerifyWorkflow for dev platform started", "id", id, "proof", proof)

	err := workflow.ExecuteActivity(ctx, a.VerifyDev, id, proof).Get(ctx, nil)
	if err != nil {
		return "failed", err
	}

	return "OK", nil
}

func (a *DevActivity) VerifyDev(ctx context.Context, id, proof string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE identities SET proof=$1, state=$2, updated_at=now() WHERE id=$3",
		proof, identities.StateVerified, id)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return nil
}
