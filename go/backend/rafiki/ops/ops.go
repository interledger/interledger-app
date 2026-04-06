package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
)

func CreatePaymentPointer(ctx context.Context, b Backends, w wallets.Wallet) (string, error) {
	var ppID string
	err := b.DB().GetContext(ctx, &ppID, "SELECT payment_pointer_id FROM rafiki_payment_pointers WHERE wallet_id=$1", w.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if ppID != "" {
		return ppID, nil
	}

	ppID, err = b.External().CreatePaymentPointer(ctx, w)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2)", w.ID, ppID)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return ppID, nil
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	// This shouldn't really happen from now on since we are not provisioning a
	// custodial key anymore when the wallet is created.

	// keys, err := b.Keys().List(ctx, w.ID)
	// if err != nil {
	// 	return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	// }
	//
	// for _, key := range keys {
	// 	err := CreatePaymentPointerKey(ctx, b, key.ID, w.ID)
	// 	if err != nil {
	// 		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	// 	}
	// }

	return ppID, nil
}

func GetWalletAddress(ctx context.Context, b Backends, walletID string) (*rafiki.WalletAddress, error) {
	externalID, err := LookupPaymentPointerID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	address, err := b.External().GetWalletAddress(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return &rafiki.WalletAddress{
		ID:         address.Id,
		AssetCode:  address.Asset.Code,
		AssetScale: address.Asset.Scale,
		URL:        address.Url,
	}, nil
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

	payment, err := b.Payments().Lookup(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, id := range reserveIDs {
		if senderAcc.Provider == xago.ProviderName {
			err = b.Xago().FinaliseReserve(ctx, id)
		} else if senderAcc.Provider == pti.ProviderName {
			err = b.PTI().FinaliseReserve(ctx, id)
		} else if senderAcc.Provider == gatehub.ProviderName {
			err = b.Gatehub().FinaliseReserve(ctx, id)
		} else if senderAcc.Provider == chimoney.ProviderName {
			err = b.Chimoney().FinaliseReserve(ctx, id)
		}
		if errors.Is(err, xago.ErrTimedOut) || errors.Is(err, chimoney.ErrTimedOut) || errors.Is(err, gatehub.ErrTimedOut) || errors.Is(err, pti.ErrTimedOut) {
			return fmt.Errorf("%w %s", rafiki.ErrTimedOut, err)
		}
		if errors.Is(err, xago.ErrNotFound) || errors.Is(err, chimoney.ErrNotFound) || errors.Is(err, gatehub.ErrNotFound) || errors.Is(err, pti.ErrNotFound) {
			return fmt.Errorf("%w %s", rafiki.ErrTimedOut, err)
		}
		if err != nil {
			return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
		}
	}

	return nil
}

func RollbackWebMonetization(ctx context.Context, b Backends, paymentID string) error {
	var reserveIDs []string
	err := b.DB().SelectContext(ctx, &reserveIDs, "SELECT id FROM rafiki_outgoing_payments WHERE payment_id=$1", paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	payment, err := b.Payments().Lookup(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, id := range reserveIDs {
		if senderAcc.Provider == xago.ProviderName {
			err = b.Xago().RollbackReserve(ctx, id)
		} else if senderAcc.Provider == pti.ProviderName {
			err = b.PTI().RollbackReserve(ctx, id)
		} else if senderAcc.Provider == gatehub.ProviderName {
			err = b.Gatehub().RollbackReserve(ctx, id)
		} else if senderAcc.Provider == chimoney.ProviderName {
			err = b.Chimoney().RollbackReserve(ctx, id)
		}
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
		return fmt.Errorf("%w Mapping rafiki keys to interledger key failed.", rafiki.ErrInternal)
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

		var access []rafiki.Access
		for _, a := range g.Access {
			var debitAmount int64
			if a.Limits.DebitAmount.Value != "" {
				debitAmount, err = strconv.ParseInt(a.Limits.DebitAmount.Value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
				}
			}

			var recvAmount int64
			if a.Limits.ReceiveAmount.Value != "" {
				recvAmount, err = strconv.ParseInt(a.Limits.ReceiveAmount.Value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
				}
			}

			access = append(access, rafiki.Access{
				ID:         a.Id,
				Identifier: a.Identifier,
				Type:       a.Type,
				Actions:    a.Actions,
				Limits: rafiki.Limits{
					Receiver:      a.Limits.Receiver,
					Interval:      a.Limits.Interval,
					DebitAmount:   currency.FromUInt64(debitAmount, currency.ParseCurrency(a.Limits.DebitAmount.AssetCode)),
					ReceiveAmount: currency.FromUInt64(recvAmount, currency.ParseCurrency(a.Limits.ReceiveAmount.AssetCode)),
				},
			})
		}

		resp = append(resp, rafiki.Grant{
			Id:                 g.Id,
			Client:             g.Client,
			State:              string(g.State),
			FinalizationReason: string(g.FinalizationReason),
			CreatedAt:          createdAt.Format("2 Jan 2006 - 15:04"),
			Access:             access,
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
	var access []rafiki.Access
	for _, a := range g.Access {
		var debitAmount int64
		if a.Limits.DebitAmount.Value != "" {
			debitAmount, err = strconv.ParseInt(a.Limits.DebitAmount.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
			}
		}

		var recvAmount int64
		if a.Limits.ReceiveAmount.Value != "" {
			recvAmount, err = strconv.ParseInt(a.Limits.ReceiveAmount.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
			}
		}

		access = append(access, rafiki.Access{
			ID:         a.Id,
			Identifier: a.Identifier,
			Type:       a.Type,
			Actions:    a.Actions,
			Limits: rafiki.Limits{
				Receiver:      a.Limits.Receiver,
				Interval:      a.Limits.Interval,
				DebitAmount:   currency.FromUInt64(debitAmount, currency.ParseCurrency(a.Limits.DebitAmount.AssetCode)),
				ReceiveAmount: currency.FromUInt64(recvAmount, currency.ParseCurrency(a.Limits.ReceiveAmount.AssetCode)),
			},
		})
	}

	return &rafiki.Grant{
		Id:                 g.Id,
		Client:             g.Client,
		State:              string(g.State),
		FinalizationReason: string(g.FinalizationReason),
		CreatedAt:          createdAt.Format("2 Jan 2006 - 15:04"),
		Access:             access,
	}, nil
}

func RevokeGrant(ctx context.Context, b Backends, grantID string) error {
	err := b.External().RevokeGrant(ctx, grantID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func UpdateWalletAddressStatus(ctx context.Context, b Backends, walletId rafiki.UpdateAddressStatus, status bool) error {
	return b.External().UpdateWalletAddressStatus(ctx, walletId, status)

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
