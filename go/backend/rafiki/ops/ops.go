package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
)

func CreatePaymentPointer(ctx context.Context, b Backends, w wallets.Wallet, assetCode string) error {
	if env.IsProd() {
		return nil
	}

	var ppID string
	err := b.DB().GetContext(ctx, &ppID, "SELECT payment_pointer_id FROM rafiki_payment_pointers WHERE wallet_id=$1", w.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if ppID != "" {
		return nil
	}

	ppID, err = b.External().CreatePaymentPointer(ctx, w, assetCode)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2)", w.ID, ppID)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	keys, err := b.Keys().List(ctx, w.ID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, key := range keys {
		err := CreatePaymentPointerKey(ctx, b, key.ID, w.ID)
		if err != nil {
			return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
		}
	}

	return nil
}

func LookupWalletID(ctx context.Context, b Backends, paymentPointerID string) (string, error) {
	var wid string
	err := b.DB().GetContext(ctx, &wid, "SELECT wallet_id FROM rafiki_payment_pointers WHERE payment_pointer_id=$1", paymentPointerID)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return wid, nil
}

func LookupPaymentPointerID(ctx context.Context, b Backends, walletID string) (string, error) {
	var ppID string
	err := b.DB().GetContext(ctx, &ppID, "SELECT payment_pointer_id FROM rafiki_payment_pointers WHERE wallet_id=$1", walletID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", rafiki.ErrNotFound, err)
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return ppID, nil
}

func FundOutgoingPayment(ctx context.Context, b Backends, paymentID string) error {
	var eventIDs []string
	err := b.DB().SelectContext(ctx, &eventIDs, "SELECT event_id FROM rafiki_outgoing_payments WHERE payment_id=$1", paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, id := range eventIDs {
		err = b.External().FundOutgoingPayment(ctx, id)
		if err != nil {
			return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
		}
	}

	return nil
}

func FinalizeWebMonetization(ctx context.Context, b Backends, paymentID string) error {
	var reserveIDs []string
	err := b.DB().SelectContext(ctx, &reserveIDs, "SELECT id FROM rafiki_outgoing_payments WHERE payment_id=$1", paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, id := range reserveIDs {
		err = b.PTI().FinaliseReserve(ctx, id)
		if err != nil {
			return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
		}
	}

	return nil
}

func CreatePaymentPointerKey(ctx context.Context, b Backends, keyID string, walletID string) error {
	key, err := b.Keys().GetPublicKey(ctx, keyID, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	ppID, err := LookupPaymentPointerID(ctx, b, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	externalID, err := b.External().CreatePaymentPointerKey(ctx, ppID, *key)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	result, err := b.DB().ExecContext(ctx, "INSERT INTO rafiki_wallet_keys (external_id, internal_id) VALUES ($1, $2);", externalID, key.ID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Mapping rafiki keys to fynbos key failed.", rafiki.ErrInternal)
	}

	return nil
}

func RevokePaymentPointerKey(ctx context.Context, b Backends, keyID string) error {
	var externalID string
	err := b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM rafiki_wallet_keys WHERE internal_id=$1;", keyID)
	if err != nil {
		return fmt.Errorf("%w %s ", rafiki.ErrInternal, err)
	}
	if externalID == "" {
		return rafiki.ErrNotFound
	}

	err = b.External().RevokePaymentPointerKey(ctx, externalID)
	if err != nil {
		return fmt.Errorf("%w %s ", rafiki.ErrInternal, err)
	}
	return nil
}

func ListGrants(ctx context.Context, b Backends, walletID string) ([]rafiki.Grant, error) {
	w, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	grants, err := b.External().ListGrants(ctx, w.AddressString())
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	var resp []rafiki.Grant
	for _, g := range grants {
		// 2024-02-01T14:17:12.219Z - comes from rafiki like this
		createdAt, _ := time.Parse(time.RFC3339, g.CreatedAt)
		resp = append(resp, rafiki.Grant{
			Id:                 g.Id,
			Client:             g.Client,
			State:              string(g.State),
			FinalizationReason: string(g.FinalizationReason),
			CreatedAt:          createdAt.Format("2 Jan 2006 - 15:04"),
		})
	}

	return resp, err
}

func GetGrant(ctx context.Context, b Backends, grantID string) (*rafiki.Grant, error) {
	g, err := b.External().GetGrant(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, g.CreatedAt)

	return &rafiki.Grant{
		Id:                 g.Id,
		Client:             g.Client,
		State:              string(g.State),
		FinalizationReason: string(g.FinalizationReason),
		CreatedAt:          createdAt.Format("2 Jan 2006 - 15:04"),
	}, nil
}

func RevokeGrant(ctx context.Context, b Backends, grantID string) error {
	err := b.External().RevokeGrant(ctx, grantID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func ListPendingWebMonetization(ctx context.Context, b Backends, walletID string) ([]transactions.Transaction, error) {
	var dbPayments []dbPayment
	err := b.DB().SelectContext(ctx, &dbPayments, `SELECT id, from_wallet, to_wallet, amount, amount_asset, created_at FROM rafiki_outgoing_payments
		WHERE payment_id is null AND (from_wallet=$1 OR to_wallet=$1) ORDER BY created_at desc`, walletID)
	if err != nil {
		return nil, err
	}

	fromTxs := make(map[string]transactions.Transaction)
	toTxs := make(map[string]transactions.Transaction)

	wList := make(map[string]*wallets.Wallet)

	lookup := func(id string) (*wallets.Wallet, error) {
		w, ok := wList[id]
		if ok {
			return w, nil
		}

		w, err := b.Wallets().Get(ctx, id)
		if err != nil {
			return nil, err
		}

		wList[id] = w
		return w, nil
	}

	for _, p := range dbPayments {
		if p.FromWalletID == walletID {
			// Outgoing transactions
			tx, ok := fromTxs[p.ToWalletID]
			if ok {
				tx.Amount = currency.FromUInt64(tx.Amount.Value+p.Amount, currency.ParseCurrency(p.Asset))
				fromTxs[p.ToWalletID] = tx
				continue
			}

			to, err := lookup(p.ToWalletID)
			if err != nil {
				return nil, err
			}
			from, err := lookup(p.FromWalletID)
			if err != nil {
				return nil, err
			}

			fromTxs[p.ToWalletID] = transactions.Transaction{
				ID:                      p.ID,
				ForeignID:               p.ID,
				Source:                  from.AddressString(),
				Destination:             to.AddressString(),
				Title:                   to.Name,
				Type:                    transactions.TransactionTypeWebMonetizationOutgoing,
				Timestamp:               p.Timestamp,
				Provider:                pti.ProviderName,
				State:                   transactions.StatePending,
				Amount:                  currency.FromUInt64(p.Amount, currency.ParseCurrency(p.Asset)),
				DestinationIdentity:     to.ID,
				DestinationIdentityType: "WalletID",
			}
			continue
		}
		// Incoming transactions
		tx, ok := toTxs[p.FromWalletID]
		if ok {
			tx.Amount = currency.FromUInt64(tx.Amount.Value+p.Amount, currency.ParseCurrency(p.Asset))
			toTxs[p.FromWalletID] = tx
			continue
		}
		to, err := lookup(p.ToWalletID)
		if err != nil {
			return nil, err
		}
		from, err := lookup(p.FromWalletID)
		if err != nil {
			return nil, err
		}
		toTxs[p.FromWalletID] = transactions.Transaction{
			ID:                      p.ID,
			ForeignID:               p.ID,
			Source:                  from.AddressString(),
			Destination:             to.AddressString(),
			Title:                   from.Name,
			Type:                    transactions.TransactionTypeWebMonetizationIncoming,
			Timestamp:               p.Timestamp,
			Provider:                pti.ProviderName,
			State:                   transactions.StatePending,
			Amount:                  currency.FromUInt64(p.Amount, currency.ParseCurrency(p.Asset)),
			DestinationIdentity:     to.ID,
			DestinationIdentityType: "WalletID",
		}
	}

	var resp []transactions.Transaction
	for _, v := range fromTxs {
		resp = append(resp, v)
	}
	for _, v := range toTxs {
		resp = append(resp, v)
	}
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].Timestamp.After(resp[j].Timestamp)
	})

	return resp, nil
}
