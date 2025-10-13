package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/rafiki"
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

func UpdateRafikiWalletEnabledJob(ctx workflow.Context, params WalletActive) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting UpdateRafikiWalletEnabledJob", zap.Bool("is Active", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets))

	err := workflow.ExecuteActivity(ctx, a.UpdateRafikiWalletActiveStatus, params).Get(ctx, nil)
	if err != nil {
		log.Error("UpdateRafikiWalletEnabledJob failed", zap.Any("err", err))
		return err
	}
	log.Info("completed UpdateRafikiWalletEnabledJob", zap.Bool("is Active", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets))
	return nil
}

func (a *Activity) UpdateRafikiWalletActiveStatus(ctx context.Context, params WalletActive) error {
	rafikiWallets, err := a.FetchRafikiWalletIds(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to fetch wallets: %w", err)
	}
	err = a.MutateRafikiWallets(ctx, rafikiWallets, params.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update wallet addresses: %w", err)
	}

	return nil
}

func (a *Activity) MutateRafikiWallets(ctx context.Context, rafikiWallet []rafiki.UpdateAddressStatus, isActive bool) error {
	client := a.b.Rafiki()
	for _, walletId := range rafikiWallet {
		err := client.UpdateWalletAddressStatus(ctx, walletId, isActive)
		if err != nil {
			return fmt.Errorf("failed to update wallet address status for %s: %w", walletId, err)
		}
	}
	return nil
}

func (a *Activity) FetchRafikiWalletIds(ctx context.Context, params WalletActive) ([]rafiki.UpdateAddressStatus, error) {
	whereClause, args := BuildWhereClause(params)

	query := "SELECT rafiki.payment_pointer_id, wallets.name FROM public.rafiki_payment_pointers as rafiki INNER JOIN wallets as wallets ON rafiki.wallet_id = wallets.id " + whereClause

	var walletList []rafiki.UpdateAddressStatus
	err := a.b.DB().SelectContext(ctx, &walletList, query, args...)
	if err != nil {

		return nil, err
	}

	if len(walletList) == 0 {
		return nil, fmt.Errorf("no wallets found for the given criteria")
	}

	return walletList, nil
}

func BuildWhereClause(params WalletActive) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Region != "" {
		if params.Region == "EU" {
			conditions = append(conditions, "wallets.country IS NOT NULL AND wallets.country NOT IN ('US', 'ZA', 'CA')")
		} else {
			conditions = append(conditions, fmt.Sprintf("wallets.country IS NOT NULL AND wallets.country = $%d", argIndex))
			args = append(args, params.Region)
			argIndex++
		}
	}

	if len(params.Wallets) > 0 {
		placeholders := make([]string, len(params.Wallets))
		for i, w := range params.Wallets {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, w)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("wallets.id IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
