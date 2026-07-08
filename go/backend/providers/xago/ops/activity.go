package ops

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/providers/xago/external"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/pacioli"
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

const travelRuleReportRowLimit = 30_000
const travelRuleReportAlertThreshold = travelRuleReportRowLimit * 3 / 4
const travelRuleKYCRetention = 30 * 24 * time.Hour

func (a *Activity) SendTravelRuleReport(ctx context.Context) error {
	if a.pgpRecipient == nil {
		return temporal.NewNonRetryableApplicationError("xago: TravelRulePGPPublicKey is not configured", "TravelRuleConfig", nil)
	}

	records, err := GetUnreportedTravelRuleRecords(ctx, a.b, travelRuleReportRowLimit)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	if err = sendTravelRuleReportEmail(ctx, a.b, sendTravelRuleReportEmailArgs{
		Records:   records,
		Recipient: a.pgpRecipient,
		Email:     a.travelRuleEmail,
	}); err != nil {
		return err
	}

	ids := make([]string, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}

	reportedAt := time.Now().UTC()
	if err = MarkTravelRuleRecordsAsReported(ctx, a.b, ids, reportedAt); err != nil {
		return err
	}

	logger := activity.GetLogger(ctx)
	logger.Info("reported xago travel rule records", "count", len(records), "reported_at", reportedAt.Format(time.RFC3339Nano))

	if len(records) >= travelRuleReportAlertThreshold {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot", fmt.Sprintf(
			"*:::[XAGO TRAVEL RULE]:::* \n *Reported rows:* %d,\n *Cap:* %d,\n *Note:* daily volume is near the cap; reports might start deferring records to the next run",
			len(records), travelRuleReportRowLimit))
	}
	return nil
}

func (a *Activity) ResendTravelRuleReport(ctx context.Context, reportedAt time.Time) error {
	if a.pgpRecipient == nil {
		return temporal.NewNonRetryableApplicationError("xago: TravelRulePGPPublicKey is not configured", "TravelRuleConfig", nil)
	}

	records, err := GetTravelRuleRecordsByReportedAt(ctx, a.b, reportedAt, travelRuleReportRowLimit)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("xago: no resendable travel rule records for reported_at %s", reportedAt.Format(time.RFC3339Nano)),
			"TravelRuleResendEmpty", nil)
	}

	if err = sendTravelRuleReportEmail(ctx, a.b, sendTravelRuleReportEmailArgs{
		Records:   records,
		Recipient: a.pgpRecipient,
		Email:     a.travelRuleEmail,
	}); err != nil {
		return err
	}

	activity.GetLogger(ctx).Info("resent xago travel rule records", "count", len(records), "reported_at", reportedAt.Format(time.RFC3339Nano))

	return nil
}

func (a *Activity) ClearTravelRuleKYC(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-travelRuleKYCRetention)
	cleared, err := ClearReportedTravelRuleKYC(ctx, a.b, cutoff, time.Now().UTC())
	if err != nil {
		return err
	}
	activity.GetLogger(ctx).Info("cleared xago travel rule KYC data", "count", cleared, "cutoff", cutoff)
	return nil
}

type sendTravelRuleReportEmailArgs struct {
	Records   []TravelRuleRecord
	Recipient *openpgp.Entity
	Email     string
}

func sendTravelRuleReportEmail(ctx context.Context, b Backends, args sendTravelRuleReportEmailArgs) error {
	csv, err := buildTravelRuleCSV(args.Records)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("build travel rule csv", "TravelRuleEncoding", err)
	}

	encryptedCsv, err := encryptPGP(csv, args.Recipient)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("encrypt travel rule csv", "TravelRuleEncryption", err)
	}

	return b.Email().SendXagoTravelRuleEmail(ctx, encryptedCsv, args.Email)
}

func buildTravelRuleCSV(records []TravelRuleRecord) ([]byte, error) {
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

	return buf.Bytes(), nil
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
