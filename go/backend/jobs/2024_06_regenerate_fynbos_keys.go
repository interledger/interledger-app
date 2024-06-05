package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func RegenerateCustodialKeysJob(ctx workflow.Context) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var wallets []wallets.Wallet
	err := workflow.ExecuteActivity(ctx, a.ListAllWallets).Get(ctx, &wallets)
	if err != nil {
		return nil, err
	}

	var failedWallets []string
	for _, w := range wallets {
		var keys []string
		err := workflow.ExecuteActivity(ctx, a.ListCustodialKeys, w.ID).Get(ctx, &keys)
		if err != nil {
			failedWallets = append(failedWallets, w.ID)
			continue
		}

		for _, keyID := range keys {
			err = workflow.ExecuteActivity(ctx, a.RevokeRafikiKey, keyID).Get(ctx, nil)
			if err != nil {
				failedWallets = append(failedWallets, w.ID)
				break
			}

			err = workflow.ExecuteActivity(ctx, a.DeleteCustodialKey, keyID).Get(ctx, nil)
			if err != nil {
				failedWallets = append(failedWallets, w.ID)
				break
			}
		}

		var newCustodialKeyID string
		err = workflow.ExecuteActivity(ctx, a.RegenerateCustodialKey, w.ID).Get(ctx, &newCustodialKeyID)
		if err != nil {
			failedWallets = append(failedWallets, w.ID)
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.RegisterKeyInRafiki, w.ID, newCustodialKeyID).Get(ctx, nil)
		if err != nil {
			failedWallets = append(failedWallets, w.ID)
			continue
		}
	}

	return failedWallets, nil
}

func (a *Activity) ListCustodialKeys(ctx context.Context, walletID string) ([]string, error) {
	return listCustodialKeys(ctx, a.b, walletID)
}

func listCustodialKeys(ctx context.Context, b Backends, walletID string) ([]string, error) {
	keyset, err := b.Keys().List(ctx, walletID)
	if err != nil {
		return nil, err
	}

	var custodialKeys []string
	for _, key := range keyset {
		if key.Type == keys.Custodial {
			custodialKeys = append(custodialKeys, key.ID)
		}
	}

	return custodialKeys, nil
}

func (a *Activity) RegenerateCustodialKey(ctx context.Context, walletID string) (string, error) {
	keyset, err := listCustodialKeys(ctx, a.b, walletID)
	if err != nil {
		return "", err
	}
	if len(keyset) > 0 {
		return "", fmt.Errorf("regenerate custodial keys internal error: wallet already has a custodial key. walletID=" + walletID)
	}

	err = a.b.Keys().ProvisionPrivateKey(ctx, walletID)
	if err != nil {
		return "", err
	}

	custodialKeys, err := listCustodialKeys(ctx, a.b, walletID)
	if err != nil {
		return "", err
	}

	return custodialKeys[0], nil
}

func (a *Activity) DeleteCustodialKey(ctx context.Context, keyID string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE wallet_keys SET deleted_at=now()::TIMESTAMP WHERE id=$1 AND key_type=$2;", keyID, keys.Custodial)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) RevokeRafikiKey(ctx context.Context, keyID string) error {
	err := a.b.Rafiki().RevokePaymentPointerKey(ctx, keyID)
	if err != nil && !errors.Is(err, rafiki.ErrNotFound) {
		return err
	}

	return nil
}

func (a *Activity) RegisterKeyInRafiki(ctx context.Context, walletID, keyID string) error {
	return a.b.Rafiki().CreatePaymentPointerKey(ctx, keyID, walletID)
}
