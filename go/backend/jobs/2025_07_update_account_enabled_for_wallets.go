package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type WalletActive struct {
	IsActive bool     `json:"isActive"`
	Wallets  []string `json:"wallets"`
	Region   string   `json:"region"`
}

func UpdateAccountEnabledJob(ctx workflow.Context, params WalletActive) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting UpdateAccountEnabledJob", zap.Bool("account_enabled", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets))

	err := workflow.ExecuteActivity(ctx, a.UpdateAppStatus, params).Get(ctx, nil)
	if err != nil {
		log.Error("UpdateAccountEnabledJob failed", zap.Any("err", err))
		return err
	}

	log.Info("completed UpdateAccountEnabledJob", zap.Bool("account_enabled", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets))
	return nil
}

func (a *Activity) UpdateAppStatus(ctx context.Context, updateParams WalletActive) error {
	var (
		whereClauses []string
		args         []interface{}
		argIndex     = 2
	)

	args = append(args, updateParams.IsActive)

	if len(updateParams.Wallets) > 0 {
		placeholders := make([]string, len(updateParams.Wallets))
		for i, walletID := range updateParams.Wallets {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, walletID)
			argIndex++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("wallet_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if updateParams.Region != "" {
		if updateParams.Region == "EU" {
			whereClauses = append(whereClauses, `
				wallet_id IN (
					SELECT id FROM public.wallets 
					WHERE country IS NOT NULL AND country NOT IN ('US', 'ZA', 'CA')
				)
			`)
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf(`
				wallet_id IN (
					SELECT id FROM public.wallets 
					WHERE country IS NOT NULL AND country = $%d
				)`, argIndex))
			args = append(args, updateParams.Region)
		}
	}

	query := "UPDATE public.wallet_features SET account_enabled = $1"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	_, err := a.b.DB().ExecContext(ctx, query, args...)
	return err
}
