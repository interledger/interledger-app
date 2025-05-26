package jobs

import (
	"context"
	"encoding/base64"
	"os"
	"time"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const limit = 100
const batchSize = 100

func RemoveCustodialKeysAndTranformKeysJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("Starting job TransformKeysToBase64URL: transforming public keys values from Base64 to Base64URL")

	err := workflow.ExecuteActivity(ctx, a.RemoveCustodialKeys).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.TransformWalletKeys).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.TransformRafikiKeys).Get(ctx, nil)
	if err != nil {
		return err
	}

	log.Info("Completed job TransformKeysToBase64URL.")
	return nil
}

func (a *Activity) RemoveCustodialKeys(ctx context.Context) error {
	logger := activity.GetLogger(ctx)

	wallets, err := a.b.Wallets().ListAll(ctx, db.Pagination{
		PageToken: "",
		PageSize:  50,
	})

	if err != nil {
		return err
	}

	for _, wallet := range wallets {
		logger.Info("removing custodial keys for wallet", "wallet", wallet.ID)
		err = a.b.Keys().RemoveCustodialKeysForWallet(ctx, wallet.ID)
		if err != nil {
			return err
		}
	}

	rows, err := a.b.DB().ExecContext(ctx, "DELETE FROM wallet_keys WHERE key_type = $1", keys.Custodial)
	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	log.Info("deleted custodial wallet keys", zap.Int64("rows-affected", affected))
	return nil
}

type walletKey struct {
	ID        string `db:"id"`
	PublicKey string `db:"public_key"`
}

func (a *Activity) TransformWalletKeys() error {
	offset := 0

	for {
		var walletsKeys []walletKey

		err := a.b.DB().Select(&walletsKeys, "SELECT id, public_key FROM wallet_keys WHERE public_key IS NOT NULL AND key_type = $1 AND location = 'database' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT $2 OFFSET $3", keys.NonCustodial, limit, offset)
		if err != nil {
			return err
		}

		if len(walletsKeys) == 0 {
			break
		}

		tx := a.b.DB().MustBegin()
		for _, key := range walletsKeys {
			decoded, err := base64.StdEncoding.DecodeString(key.PublicKey)
			if err != nil {
				return err
			}

			encoded := base64.RawURLEncoding.EncodeToString(decoded)

			// Update the row
			_, err = tx.Exec(`UPDATE wallet_keys SET public_key = $1 WHERE id = $2`, encoded, key.ID)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
		offset += batchSize
	}

	return nil
}

type rafikiKey struct {
	ID string `db:"id"`
	X  string `db:"x"`
}

func (a *Activity) TransformRafikiKeys() error {
	offset := 100
	connString := os.Getenv("RAFIKI_DB_URL")

	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close()

	for {
		var rafikiKeys []rafikiKey

		err := a.b.DB().Select(&rafikiKeys, `SELECT id, x FROM \"walletAddressKeys\" WHERE revoked = false ORDER BY "createdAt" DESC LIMIT $1 OFFSET $2`, limit, offset)
		if err != nil {
			return err
		}

		if len(rafikiKeys) == 0 {
			break
		}

		tx := a.b.DB().MustBegin()
		for _, key := range rafikiKeys {
			decoded, err := base64.StdEncoding.DecodeString(key.X)
			if err != nil {
				return err
			}

			encoded := base64.RawURLEncoding.EncodeToString(decoded)

			_, err = tx.Exec(`UPDATE "walletAddressKeys" SET x = $1 WHERE id = $2`, encoded, key.ID)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
		offset += batchSize
	}

	return nil
}
