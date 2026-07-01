package ops

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/pacioli"
)

func (a *Activity) PollDeposits(ctx context.Context) ([]external.Deposit, error) {
	var page int = 1
	var deposits []external.Deposit
	for {
		deps, err := a.b.External().ListDeposits(ctx, page)
		if err != nil {
			return nil, err
		}

		var lastDepID string
		for _, dep := range deps {
			if !strings.EqualFold(dep.Status, "success") {
				continue
			}
			lastDepID = dep.TransactionID
			deposits = append(deposits, dep)
		}

		// Lookup if we already have the deposit in our DB
		if lastDepID != "" {
			var txID string
			err = a.b.DB().GetContext(ctx, &txID, "SELECT transaction_id FROM xago_deposits WHERE transaction_id=$1", lastDepID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			/// We already have this deposit in our DB so we've reached to where we want to scan
			if txID != "" {
				break
			}
		}
		if len(deps) < 10 {
			break
		}
		page++
	}

	return deposits, nil
}

func (a *Activity) SaveDeposits(ctx context.Context, deposits []external.Deposit) error {
	stmt, err := a.b.DB().PrepareContext(ctx, "INSERT INTO xago_deposits (transaction_id, origin_amount, amount, status, account_id) VALUES ($1, $2,$3, $4,$5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, dep := range deposits {
		_, err = stmt.ExecContext(ctx, dep.TransactionID, dep.OriginAmount, dep.Amount, dep.Status, dep.AccountID)
		if db.IsErrorCode(err, db.UniqueViolationError) {
			continue
		}
		if err != nil {
			return err
		}

		subAcc, err := LookupByAccountID(ctx, a.b, dep.AccountID)
		if err != nil {
			return err
		}

		// Best effort
		a.b.Email().SendDepositReceivedEmail(ctx, subAcc.WalletID, currency.FromFloat64(dep.Amount, currency.ZAR), "", "")
	}

	return nil
}

func (a *Activity) CreateDepositTransactions(ctx context.Context, deposits []external.Deposit) error {
	for _, dep := range deposits {
		subAcc, err := LookupByAccountID(ctx, a.b, dep.AccountID)
		if err != nil {
			return err
		}

		lal, err := a.b.LinkedAccounts().ListByWalletId(ctx, subAcc.WalletID)
		if err != nil {
			return err
		}

		var acc linkedaccounts.LinkedAccount
		for _, la := range lal {
			if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && la.SendCurrency == currency.ZAR {
				acc = la
				break
			}
		}
		if acc.ID == "" {
			return fmt.Errorf("%w no ZAR account found for depost", xago.ErrNotFound)
		}

		// Som of these may actually be no ops because the were filled in by the webhook. All of it's idempotent so it's safe to rerun
		// Idempotent call
		_, err = a.b.Transactions().GetTransaction(ctx, acc.WalletID, dep.TransactionID)
		if errors.Is(err, transactions.ErrNotFound) {
			_, err = a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
				ID:                      dep.TransactionID,
				WalletID:                acc.WalletID,
				ForeignID:               dep.TransactionID,
				ForeignType:             transactions.TransactionTypeDeposit,
				Provider:                transactions.ProviderXago,
				State:                   transactions.StateCompleted,
				Title:                   "Deposit",
				Note:                    "Deposit received",
				Source:                  "Bank Deposit",
				Destination:             acc.WalletID,
				Amount:                  currency.FromFloat64(dep.Amount, currency.ZAR),
				LinkedAccountTitle:      acc.Title(),
				DestinationIdentity:     acc.WalletID,
				DestinationIdentityType: "WalletID",
			})
		}
		if err != nil {
			return err
		}

		// Also an idempotent call
		tr, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
			{
				ID:              dep.TransactionID,
				Amount:          currency.FromFloat64(dep.Amount, currency.ZAR).Value,
				DebitAccountID:  xago.ZAROpsAccount,
				CreditAccountID: acc.ID,
				Pending:         false,
				Code:            1,
				Ledger:          xago.LedgerIDZAR,
			},
		})
		if err != nil {
			return err
		}

		if len(tr) > 0 && tr[0].Code != 0 {
			return fmt.Errorf("failed to create pacioli transaction for xago deposit status (%s)", tr[0].Code)
		}
	}

	return nil
}

type TravelRuleRecord struct {
	ID                     string `db:"id"`
	TransactionReference   string `db:"transaction_reference"`
	OriginatorName         string `db:"originator_name"`
	OriginatorAccountID    string `db:"originator_account_id"`
	OriginatorAddress      string `db:"originator_address"`
	OriginatorPlaceOfBirth string `db:"originator_place_of_birth"`
	OriginatorDateOfBirth  string `db:"originator_date_of_birth"`
	BeneficiaryName        string `db:"beneficiary_name"`
	BeneficiaryAccountID   string `db:"beneficiary_account_id"`
}

func (a *Activity) GetTravelRuleRecords(ctx context.Context) ([]TravelRuleRecord, error) {
	var records []TravelRuleRecord
	err := a.b.DB().SelectContext(ctx, &records, `
		SELECT id,
		       transaction_reference,
		       originator_name,
		       originator_account_id,
		       originator_address,
		       originator_place_of_birth,
		       originator_date_of_birth,
		       beneficiary_name,
		       beneficiary_account_id
		FROM xago_travel_rule_records
		WHERE reported_at IS NULL
	`)
	return records, err
}

func (a *Activity) BuildTravelRuleCSV(_ context.Context, records []TravelRuleRecord) ([]byte, error) {
	if a.pgpRecipient == nil {
		return nil, fmt.Errorf("xago: TravelRulePGPPublicKey is not configured")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{
		"transaction_reference",
		"originator_name",
		"originator_account_id",
		"originator_address",
		"originator_place_of_birth",
		"originator_date_of_birth",
		"beneficiary_name",
		"beneficiary_account_id",
	}); err != nil {
		return nil, err
	}

	for _, r := range records {
		if err := w.Write([]string{
			r.TransactionReference,
			r.OriginatorName,
			r.OriginatorAccountID,
			r.OriginatorAddress,
			r.OriginatorPlaceOfBirth,
			r.OriginatorDateOfBirth,
			r.BeneficiaryName,
			r.BeneficiaryAccountID,
		}); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return encryptPGP(buf.Bytes(), a.pgpRecipient)
}

func encryptPGP(plaintext []byte, recipient *openpgp.Entity) ([]byte, error) {
	var buf bytes.Buffer
	w, err := openpgp.Encrypt(&buf, openpgp.EntityList{recipient}, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("pgp encrypt: %w", err)
	}
	if _, err = w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("pgp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("pgp close: %w", err)
	}
	return buf.Bytes(), nil
}

func (a *Activity) SendTravelRuleEmail(ctx context.Context, csvBytes []byte) error {
	return a.b.Email().SendXagoTravelRuleEmail(ctx, csvBytes, a.travelRuleEmail)
}

func (a *Activity) MarkTravelRuleRecordsAsReported(ctx context.Context, records []TravelRuleRecord) error {
	ids := make([]string, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	_, err := a.b.DB().ExecContext(ctx, `
		UPDATE xago_travel_rule_records SET
			originator_name           = '',
			originator_account_id     = '',
			originator_address        = '',
			originator_place_of_birth = '',
			originator_date_of_birth  = '',
			beneficiary_name          = '',
			beneficiary_account_id    = '',
			reported_at               = now()
		WHERE id = ANY($1::uuid[])`,
		pq.Array(ids),
	)
	return err
}
