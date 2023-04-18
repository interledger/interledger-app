package platforms

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
	"go.temporal.io/sdk/workflow"
)

type twitter struct {
	platform identities.Platform
}

func newTwitter(platform identities.Platform) *twitter {
	return &twitter{platform: platform}
}

func (t *twitter) VerifyWorkflow() interface{} {
	return TwitterVerifyWorkflow
}

func (t *twitter) NewVerifyCode() string {
	return uuid.NewString()
}

func (t *twitter) VerifyInstructions() string {
	return `In this environment all you need to do to verify is to request it. Enjoy in a NON Production environment.`
}

type TwitterActivity struct {
	b Backends
}

func NewTwitterActivity(b Backends) *TwitterActivity {
	return &TwitterActivity{b: b}
}

func TwitterVerifyWorkflow(ctx workflow.Context, id, proof string) (string, error) {
	var a *TwitterActivity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("VerifyWorkflow for twitter platform started", "id", id, "proof", proof)

	err := workflow.ExecuteActivity(ctx, a.VerifyTwitter, id, proof).Get(ctx, nil)
	if err != nil {
		return "failed", err
	}

	return "OK", nil
}

func (a *TwitterActivity) VerifyTwitter(ctx context.Context, id, proof string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE identities SET proof=$1, state=$2, updated_at=now() WHERE id=$3",
		proof, identities.StateVerified, id)
	if err != nil {
		return fmt.Errorf("%w %s", identities.ErrInternal, err)
	}

	return nil
}
