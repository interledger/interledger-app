package jobs

import (
	"encoding/base64"
	"time"

	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const limit = 100
const batchSize = 100

func TransformKeysToBase64URLJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
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

	log.Info("Completed job TransformKeysToBase64URL")
	return nil
}

func (a *Activity) TransformWalletKeys() error {
	type key struct {
		ID        string `db:"id"`
		PublicKey string `db:"public_key"`
	}

	offset := 0

	for {
		var wKeys []key

		err := a.b.DB().Select(&wKeys, "SELECT id, public_key FROM wallet_keys WHERE public_key IS NOT NULL AND key_type = $1 AND location = 'database' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT $2 OFFSET $3", keys.NonCustodial.String(), limit, offset)
		if err != nil {
			return err
		}

		if len(wKeys) == 0 {
			break
		}

		tx := a.b.DB().MustBegin()
		for _, key := range wKeys {
			decoded, err := base64.StdEncoding.DecodeString(key.PublicKey)
			if err != nil {
				log.Error("could not decode", zap.String("key", key.ID), zap.Any("err", err))
				continue
			}

			encoded := base64.RawURLEncoding.EncodeToString(decoded)

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

func (a *Activity) TransformRafikiKeys() error {
	type key struct {
		ID string `db:"id"`
		X  string `db:"x"`
	}

	offset := 0

	db, err := DbConnection(a.cfg.RafikiDBURL)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close()

	for {
		var rKeys []key

		err = db.Select(&rKeys, `SELECT id, x FROM "walletAddressKeys" WHERE revoked = false ORDER BY "createdAt" ASC LIMIT $1 OFFSET $2`, limit, offset)
		if err != nil {
			return err
		}

		if len(rKeys) == 0 {
			break
		}

		tx := db.MustBegin()
		for _, key := range rKeys {
			decoded, err := base64.StdEncoding.DecodeString(key.X)
			if err != nil {
				log.Error("could not decode", zap.String("key", key.ID), zap.Any("err", err))
				continue
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
