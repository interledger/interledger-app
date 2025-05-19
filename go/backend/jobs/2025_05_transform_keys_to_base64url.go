package jobs

import (
	"encoding/base64"
	"os"
	"time"

	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const batchSize = 100

func TranformKeysToBase64URLJob(ctx workflow.Context, params MigrateWalletAddressesParams) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("Starting job TransformKeysToBase64URL: transforming public keys values from Base64 to Base64URL")

	err := workflow.ExecuteActivity(ctx, a.TransformWalletKeys).Get(ctx, nil)
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

type walletKey struct {
	ID        int    `db:"id"`
	PublicKey string `db:"public_key"`
}

func (a *Activity) TransformWalletKeys() error {
	for {
		var walletsKeys []walletKey

		err := a.b.DB().Select(&walletsKeys, "SELECT id, public_key FROM wallet_keys WHERE key_type = $1 AND location = 'database' AND deleted_at IS NULL LIMIT $2", keys.NonCustodial, batchSize)
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
			_, err = tx.Exec(`UPDATE keys SET x = $1 WHERE id = $2`, encoded, key.ID)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

type rafikiKey struct {
	ID int    `db:"id"`
	X  string `db:"x"`
}

func (a *Activity) TransformRafikiKeys() error {
	connString := os.Getenv("RAFIKI_DB_URL")

	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close()

	for {
		var rafikiKeys []rafikiKey

		err := a.b.DB().Select(&rafikiKeys, `SELECT id, x FROM \"walletAddressKeys\" WHERE LIMIT $1`, batchSize)
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
	}

	return nil
}
