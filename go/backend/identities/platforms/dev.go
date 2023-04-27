package platforms

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (d *dev) NewVerifyCode(args *NewVerifyCodeArgs) string {
	return uuid.NewString()
}

func (d *dev) VerifyInstructions() string {
	return `In this environment all you need to do to verify is to request it. Enjoy in a NON Production environment.`
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
