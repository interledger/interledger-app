package jobs

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type PersonaAccounts []struct {
	Id       string `json:"ID"`
	WalletId string `json:"ReferenceID"`
}

func UpdatePersonaAccountIDs(ctx workflow.Context, params PersonaAccounts) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting UpdatePersonaAccountIDsActivity")

	err := workflow.ExecuteActivity(ctx, a.UpdatePersonaAccountIDsActivity, params).Get(ctx, nil)
	if err != nil {
		log.Error("UpdatePersonaAccountIDsActivity failed", zap.Any("err", err))
		return err
	}
	log.Info("completed UpdatePersonaAccountIDsActivity")
	return nil
}

func (a *Activity) UpdatePersonaAccountIDsActivity(ctx context.Context, params PersonaAccounts) error {
	for _, account := range params {
		if account.WalletId == "" || account.Id == "" {
			continue
		}
		log.Info("updating persona account id", zap.String("wallet_id", account.WalletId), zap.String("persona_account_id", account.Id))
		_, err := a.b.DB().ExecContext(ctx, "UPDATE kyc_persona_accounts SET external_id = $1, updated_at = now() WHERE wallet_id = $2", account.Id, account.WalletId)
		if err != nil {
			return fmt.Errorf("failed to update persona account id for wallet %s: %w", account.WalletId, err)
		}
	}

	return nil
}
