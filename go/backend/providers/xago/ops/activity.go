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
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/providers/xago/external"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/log"
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

type travelRuleReportRow struct {
	TransactionReference   string
	OriginatorName         string
	OriginatorAccountID    string
	OriginatorAddress      string
	OriginatorPlaceOfBirth string
	OriginatorDateOfBirth  string
	BeneficiaryName        string
	BeneficiaryAccountID   string
}

const (
	travelRuleReportBatchSize = 10_000
	travelRuleFetchThrottle   = 300 * time.Millisecond
)

func (a *Activity) SendTravelRuleReport(ctx context.Context, cutoff time.Time) error {
	if a.pgpRecipient == nil {
		return temporal.NewNonRetryableApplicationError("xago: TravelRulePGPPublicKey is not configured", "TravelRuleConfig", nil)
	}

	records, err := GetUnreportedTravelRuleRecords(ctx, a.b, cutoff)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	resolver := newTravelRuleResolver(a.b)

	batchTotal := (len(records) + travelRuleReportBatchSize - 1) / travelRuleReportBatchSize

	var reported, skipped, batchNumber int

	for start := 0; start < len(records); start += travelRuleReportBatchSize {
		end := min(start+travelRuleReportBatchSize, len(records))
		batch := records[start:end]
		batchNumber++

		rows, resolvedIDs, batchSkipped := resolver.resolve(ctx, batch)
		skipped += batchSkipped
		if len(rows) == 0 {
			continue
		}

		encrypted, err := encryptTravelRuleRows(rows, a.pgpRecipient)
		if err != nil {
			return err
		}
		if err := a.b.Email().SendXagoTravelRuleEmail(ctx, encrypted, a.travelRuleEmail, cutoff, batchNumber, batchTotal, len(rows)); err != nil {
			return err
		}

		reportedAt := time.Now().UTC().Truncate(time.Microsecond)
		if err := MarkTravelRuleRecordsAsReported(ctx, a.b, resolvedIDs, reportedAt, batchNumber, batchTotal); err != nil {
			return err
		}
		reported += len(rows)
		log.Info("reported xago travel rule batch", zap.Int("batch", batchNumber), zap.Int("total", batchTotal), zap.Int("count", len(rows)), zap.Time("reported_at", reportedAt))
	}

	log.Info("finished xago travel rule report", zap.Int("reported", reported), zap.Int("skipped", skipped))

	if skipped > 0 {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot", fmt.Sprintf(
			"*:::[XAGO TRAVEL RULE]:::* %d some records were skipped", skipped))
	}

	return nil
}

func (a *Activity) ResendTravelRuleReport(ctx context.Context, reportedAt time.Time) error {
	if a.pgpRecipient == nil {
		return temporal.NewNonRetryableApplicationError("xago: TravelRulePGPPublicKey is not configured", "TravelRuleConfig", nil)
	}

	records, err := GetTravelRuleRecordsByReportedAt(ctx, a.b, reportedAt)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("xago: no resendable travel rule records for reported_at %s", reportedAt),
			"TravelRuleResendEmpty", nil)
	}

	rows, _, skipped := newTravelRuleResolver(a.b).resolve(ctx, records)
	if skipped > 0 {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot", fmt.Sprintf(
			"*:::[XAGO TRAVEL RULE]:::* resend for %s REFUSED: %d of %d records no longer resolve (KYC); see logs for payment IDs", reportedAt, skipped, len(records)))
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("xago: %d of %d records for reported_at %s could not be resolved; refusing to send a partial resend (see logs for payment IDs)", skipped, len(records), reportedAt),
			"TravelRuleResendUnresolved", nil)
	}

	encrypted, err := encryptTravelRuleRows(rows, a.pgpRecipient)
	if err != nil {
		return err
	}
	if err = a.b.Email().SendXagoTravelRuleEmail(ctx, encrypted, a.travelRuleEmail, reportedAt, records[0].BatchNumber, records[0].BatchTotal, len(rows)); err != nil {
		return err
	}

	log.Info("resent xago travel rule records", zap.Int("batch", records[0].BatchNumber), zap.Int("total", records[0].BatchTotal), zap.Int("count", len(rows)), zap.Time("reported_at", reportedAt))

	return nil
}

type travelRuleOriginator struct {
	name         string
	accountID    string
	address      string
	placeOfBirth string
	dateOfBirth  string
}
type travelRuleBeneficiary struct {
	name      string
	accountID string
}

type travelRuleResolver struct {
	b             Backends
	originators   map[string]travelRuleOriginator
	beneficiaries map[string]travelRuleBeneficiary
}

func newTravelRuleResolver(b Backends) *travelRuleResolver {
	return &travelRuleResolver{
		b:             b,
		originators:   map[string]travelRuleOriginator{},
		beneficiaries: map[string]travelRuleBeneficiary{},
	}
}

func (res *travelRuleResolver) resolve(ctx context.Context, records []dbTravelRuleRecord) (rows []travelRuleReportRow, resolvedIDs []string, skipped int) {
	for _, r := range records {
		originator, ok := res.originators[r.SenderWalletID]
		if !ok {
			user, err := res.b.Gatehub().GetUser(ctx, r.SenderWalletID)
			time.Sleep(travelRuleFetchThrottle)
			if err != nil {
				log.Warn("skipping travel rule record: gatehub lookup failed", zap.String("payment_id", r.PaymentID), zap.String("sender_wallet_id", r.SenderWalletID), zap.Error(err))
				skipped++
				continue
			}
			originator = travelRuleOriginator{
				name:      joinNonEmpty(" ", user.Profile.FirstName, user.Profile.LastName),
				accountID: user.UUID,
				address: joinNonEmpty(", ",
					user.Profile.AddressStreet1,
					user.Profile.AddressStreet2,
					user.Profile.AddressCity,
					user.Profile.AddressPostalCode,
					user.Profile.AddressCountryCode,
				),
				placeOfBirth: joinNonEmpty(", ", user.Profile.BirthCity, user.Profile.BirthCountryCode),
				dateOfBirth:  formatDateOfBirth(user.Profile.BirthYear, user.Profile.BirthMonth, user.Profile.BirthDay),
			}
			res.originators[r.SenderWalletID] = originator
		}

		beneficiary, ok := res.beneficiaries[r.ReceiverWalletID]
		if !ok {
			user, err := res.b.KYC().GetPersonaAccountAttributes(ctx, r.ReceiverWalletID)
			time.Sleep(travelRuleFetchThrottle)
			if err != nil {
				log.Warn("skipping travel rule record: persona lookup failed", zap.String("payment_id", r.PaymentID), zap.String("receiver_wallet_id", r.ReceiverWalletID), zap.Error(err))
				skipped++
				continue
			}
			sub, err := LookupSubAccount(ctx, res.b, r.ReceiverWalletID)
			if err != nil {
				log.Warn("skipping travel rule record: xago sub-account lookup failed", zap.String("payment_id", r.PaymentID), zap.String("receiver_wallet_id", r.ReceiverWalletID), zap.Error(err))
				skipped++
				continue
			}
			beneficiary = travelRuleBeneficiary{
				name:      joinNonEmpty(" ", user.FirstName, user.LastName),
				accountID: sub.AccountID,
			}
			res.beneficiaries[r.ReceiverWalletID] = beneficiary
		}

		rows = append(rows, travelRuleReportRow{
			TransactionReference:   r.PaymentID,
			OriginatorName:         originator.name,
			OriginatorAccountID:    originator.accountID,
			OriginatorAddress:      originator.address,
			OriginatorPlaceOfBirth: originator.placeOfBirth,
			OriginatorDateOfBirth:  originator.dateOfBirth,
			BeneficiaryName:        beneficiary.name,
			BeneficiaryAccountID:   beneficiary.accountID,
		})
		resolvedIDs = append(resolvedIDs, r.ID)
	}

	return rows, resolvedIDs, skipped
}

func encryptTravelRuleRows(rows []travelRuleReportRow, recipient *openpgp.Entity) ([]byte, error) {
	csv, err := buildTravelRuleCSV(rows)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError("build travel rule csv", "TravelRuleEncoding", err)
	}

	encrypted, err := encryptPGP(csv, recipient)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError("encrypt travel rule csv", "TravelRuleEncryption", err)
	}

	return encrypted, nil
}

func joinNonEmpty(sep string, parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

func formatDateOfBirth(year, month, day int) string {
	if year == 0 || month == 0 || day == 0 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func buildTravelRuleCSV(records []travelRuleReportRow) ([]byte, error) {
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
