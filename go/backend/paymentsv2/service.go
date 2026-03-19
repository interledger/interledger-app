package paymentsv2

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/api/enums/v1"
	temporalClient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type Service struct {
	Repo         repository
	Temporal     temporalClient.Client
	LedgerClient pacioli.Client
}

func NewService(repo repository, ledgerClient pacioli.Client, temporal temporalClient.Client) *Service {
	fmt.Println("ledgerClient", ledgerClient)
	return &Service{Repo: repo, LedgerClient: ledgerClient, Temporal: temporal}

}

func (s Service) Get(ctx context.Context, id string) (*Payment, error) {
	return s.Repo.Get(ctx, id)
}

func (s Service) Store(ctx context.Context, payment *Payment) error {
	return s.Repo.Store(ctx, payment)
}

func (s Service) Process(ctx context.Context, payment *Payment) error {
	log.Info("Processing payment", zap.String("id", payment.ID))

	workflowOptions := temporalClient.StartWorkflowOptions{
		ID:                       "payments_v2_" + payment.ID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // Workflow has 8 days to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err := s.Temporal.ExecuteWorkflow(ctx, workflowOptions, PaymentWorkflowV2, payment)
	if err != nil {
		return err
	}

	return nil
}
