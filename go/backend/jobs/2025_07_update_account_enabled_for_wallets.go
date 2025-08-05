package jobs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
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
		log.Error("UpdateAppStatus failed", zap.Any("err", err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateRafikiWalletActiveStatus, params).Get(ctx, nil)
	if err != nil {
		log.Error("UpdateRafikiWalletActiveStatus failed", zap.Any("err", err))
		return err
	}
	log.Info("completed UpdateAccountEnabledJob", zap.Bool("account_enabled", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets))
	return nil
}

func (a *Activity) UpdateAppStatus(ctx context.Context, updateParams WalletActive) error {

	query, args, err := buildUpdateAppStatusQuery(updateParams)
	if err != nil {
		log.Error("Failed to build query for updating app status", zap.Error(err))
		return err
	}

	result, err := a.b.DB().ExecContext(ctx, query, args...)
	if err != nil {
		log.Error("Failed to execute query for updating app status", zap.String("query", query), zap.Any("args", args), zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info("App status updated successfully", zap.Int64("rowsAffected", rowsAffected))

	return nil
}

func (a *Activity) UpdateRafikiWalletActiveStatus(ctx context.Context, params WalletActive) error {
	connString := os.Getenv("RAFIKI_DB_URL")
	if connString == "" {
		log.Error("RAFIKI_DB_URL environment variable is not set")
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

func buildUpdateAppStatusQuery(updateParams WalletActive) (string, []interface{}, error) {
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

	return query, args, nil
}

func (a *Activity) fetchWallets(ctx context.Context, params WalletActive) ([]string, error) {
	whereClause, args := buildWhereClause(params)

	if whereClause == "" {
		return nil, nil
	}

	query := "SELECT name FROM public.wallets " + whereClause
	rows, err := a.b.DB().QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("Error fetching wallets", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var wallets []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Error("Row scan error", zap.Error(err))
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		wallets = append(wallets, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
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

	db, err := sqlx.ConnectContext(ctx, "postgres", connString)
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
