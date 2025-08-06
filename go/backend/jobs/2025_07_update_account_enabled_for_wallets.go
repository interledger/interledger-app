package jobs

import (
	"context"
	"fmt"
	"os"
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
	connString := os.Getenv("RAFIKI_DB_URL")
	if connString == "" {
		return fmt.Errorf("RAFIKI_DB_URL environment variable is not set")
	}

	wallets, err := a.fetchWallets(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to fetch wallets: %w", err)
	}

	err = a.updateWalletAddresses(ctx, wallets, params.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update wallet addresses: %w", err)
	}

	return nil
}

func (a *Activity) fetchWallets(ctx context.Context, params WalletActive) ([]string, error) {
	whereClause, args := buildWhereClause(params)

	if whereClause == "" {
		return nil, nil
	}

	query := "SELECT name FROM public.wallets " + whereClause
	var wallets []string
	err := a.b.DB().GetContext(ctx, &wallets, query, args...)
	if err != nil {
		log.Error("Error fetching wallets", zap.Error(err))
		return nil, err
	}

	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets found for the given criteria")
	}

	return wallets, nil
}

func (a *Activity) updateWalletAddresses(ctx context.Context, wallets []string, isActive bool) error {

	connString := os.Getenv("RAFIKI_DB_URL")
	if connString == "" {
		return fmt.Errorf("RAFIKI_DB_URL environment variable is not set")
	}

	db, err := DbConnection(connString)
	if err != nil {
		return fmt.Errorf("failed to establish a new database connection: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Error("Failed to close the database connection", zap.Error(closeErr))
		}
	}()

	updateQuery := "UPDATE \"walletAddresses\" SET \"deactivatedAt\" = "
	if isActive {
		updateQuery += "null"
	} else {
		updateQuery += "NOW()"
	}

	var updateArgs []interface{}
	if len(wallets) > 0 {
		placeholders := make([]string, len(wallets))
		for i, w := range wallets {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			updateArgs = append(updateArgs, w)
		}
		updateQuery += fmt.Sprintf(" WHERE \"publicName\" IN (%s)", strings.Join(placeholders, ", "))
	}

	_, err = db.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	log.Info("Wallet addresses updated successfully", zap.Int("count", len(wallets)))
	return nil
}

func buildWhereClause(params WalletActive) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Region != "" {
		if params.Region == "EU" {
			conditions = append(conditions, "country IS NOT NULL AND country NOT IN ('US', 'ZA', 'CA')")
		} else {
			conditions = append(conditions, fmt.Sprintf("country IS NOT NULL AND country = $%d", argIndex))
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
		conditions = append(conditions, fmt.Sprintf("id IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}
