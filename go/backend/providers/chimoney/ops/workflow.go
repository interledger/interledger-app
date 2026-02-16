package ops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"gitlab.com/fynbos/pacioli"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const depositChannel = "chimoney_deposits"

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends) *Activity {
	ec := external.New(
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, external.Redact),
			),
		},
	)

	return &Activity{
		b:        b,
		external: ec,
	}
}

func ChimomeyCompleteKYC(ctx workflow.Context, walletID string, kycStatus kyc.Status) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	return workflow.ExecuteActivity(ctx, a.SetChimoneyKYCStatus, walletID, kycStatus).Get(ctx, nil)
}

func CreateChimoneyUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating chimoney sub account.")

	var exID string
	err := workflow.ExecuteActivity(ctx, a.CreateChimoneyWallet, walletID).Get(ctx, &exID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveChimoneyWallet, walletID, exID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.CreateLinkedAccount, walletID, exID).Get(ctx, &la)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateBalanceAccount, la.ID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return exID, nil
}

func (a *Activity) GetChimoneyWallet(ctx context.Context, walletID string) (*external.WalletResp, error) {
	externalID, err := GetChiWallet(ctx, a.b, walletID)
	if errors.Is(err, chimoney.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("chimoney: wallet not found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	return a.external.GetWallet(ctx, externalID)
}

func (a *Activity) SetChimoneyKYCStatus(ctx context.Context, walletID string, status kyc.Status) error {
	return a.b.KYC().SetKYCStatus(ctx, walletID, status)
}

func (a *Activity) CreateLinkedAccount(ctx context.Context, walletID, externalID string) (*linkedaccounts.LinkedAccount, error) {
	w, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, wallets.ErrNoWalletFound) {
		return nil, temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", err)
	} else if err != nil {
		return nil, err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, w.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	for _, la := range las {
		if la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeBalance {
			return &la, nil
		}
	}

	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:        w.ID,
		Type:            chimoney.AccTypeBalance,
		Provider:        chimoney.ProviderName,
		ProviderID:      externalID,
		Name:            "CAD Balance",
		Nickname:        "CAD Balance",
		CanReceive:      true,
		ReceiveCountry:  w.Country,
		ReceiveCurrency: currency.CAD,
		SendCountry:     w.Country,
		SendCurrency:    currency.CAD,
		CanSend:         true,
		State:           linkedaccounts.Verified,
	})
	if err != nil {
		return nil, err
	}

	return la, nil
}

func (a *Activity) CreateBalanceAccount(ctx context.Context, id string) error {
	accs, err := a.b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:                         id,
			LedgerID:                   chimoney.LedgerIDCAD,
			Code:                       1,
			DebitsMustNotExceedCredits: true,
			CreditsMustNotExceedDebits: false,
		},
	})
	if err != nil {
		return err
	}

	if len(accs) == 0 {
		// No error codes to speak of
		return nil
	}

	if accs[0].Code != pacioli.AccountOK && accs[0].Code != pacioli.AccountExists {
		return fmt.Errorf("%w failed to setup account status(%s)", chimoney.ErrInternal, accs[0].Code)
	}

	return nil
}

func (a *Activity) SaveChimoneyWallet(ctx context.Context, walletID, exID string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO chi_money_wallets (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;", exID, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return nil
}

func convertPhoneToE164(phoneNumber string) string {
	var digitsOnlyRegex = regexp.MustCompile(`\D`)
	digits := digitsOnlyRegex.ReplaceAllString(phoneNumber, "")
	return "+" + digits
}

func (a *Activity) CreateChimoneyWallet(ctx context.Context, walletID string) (string, error) {
	ul, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(ul) < 1 {
		return "", fmt.Errorf("%w No Interledger user found for walletID", chimoney.ErrInternal)
	}

	exID, err := a.external.CreateWallet(ctx, external.CreateWalletReq{
		Name:        ul[0].FirstName + " " + ul[0].LastName,
		Email:       ul[0].Email,
		FirstName:   ul[0].FirstName,
		LastName:    ul[0].LastName,
		PhoneNumber: convertPhoneToE164(ul[0].PhoneNumber),
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return exID, nil
}

type withdrawalStage uint8

var (
	reserveLiquidity         withdrawalStage = 0
	externalWithdrawal       withdrawalStage = 1
	finalizeReserve          withdrawalStage = 2
	updateTransaction        withdrawalStage = 3
	externalWithdrawalFailed withdrawalStage = 4
)

func ExecuteChimoneyFinishWithdrawalWorkflow(
	ctx workflow.Context, IssueID string, chiWalletID string, status string,
) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing chimoney finish withdrawal with status: ", zap.String("status", status))

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletIDFromChimoneyID, chiWalletID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	var trx transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetChimoneyTransactionByForeignID, walletID, IssueID).Get(ctx, &trx)
	if err != nil {
		return err
	}

	if trx.State == transactions.StateCompleted || trx.State == transactions.StateFailed {
		logger.Info("Chimoney withdrawal already finalized with status: ", zap.String("status", string(trx.State)), zap.String("issueID", IssueID))
		return nil
	}

	if status == "cancelled" || status == "expired" {
		return rollBackWithdrawal(ctx, a, externalWithdrawalFailed, trx.WalletID, trx.ID)
	}

	// finalize balance
	err = workflow.ExecuteActivity(ctx, a.FinalizeChimoneyBalance, trx.ID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, finalizeReserve, trx.WalletID, trx.ID)
	}

	// finalize transaction
	err = workflow.ExecuteActivity(ctx, a.CompleteChimoneyTransaction, trx.WalletID, trx.ID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, updateTransaction, trx.WalletID, trx.ID)
	}

	return nil

}
func ExecuteChimoneyWithdrawalWorkflow(
	ctx workflow.Context, walletID, trxID string,
) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing chimoney withdrawal.")

	var trx transactions.Transaction
	err := workflow.ExecuteActivity(ctx, a.GetChimoneyTransaction, walletID, trxID).Get(ctx, &trx)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SetWithdrawalTransfer, walletID, trxID).Get(ctx, &trx)
	if err != nil {
		return err
	}

	// reserve liquidity and surface Insufficient balance errors.
	err = workflow.ExecuteActivity(ctx, a.ReserveChimoneyBalance, walletID, trxID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, reserveLiquidity, walletID, trxID)
	}

	// perform the withdrawal. Rollback on failure
	var externalTrx external.WithdrawResponse
	err = workflow.ExecuteActivity(ctx, a.ChimoneyWithdraw, walletID, trxID).Get(ctx, &externalTrx)
	if err != nil {
		return rollBackWithdrawal(ctx, a, externalWithdrawal, walletID, trxID)
	}

	if len(externalTrx.Data) < 1 {
		logger.Error("chimoney withdrawal: No payout data found")
		return rollBackWithdrawal(ctx, a, externalWithdrawal, walletID, trxID)
	}
	// set the foreign ID to the issueID from chimoney so we can track it via webhooks
	err = workflow.ExecuteActivity(ctx, a.SetChimoneyTransactionForeignID, trxID, externalTrx.Data[0].IssueID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, externalWithdrawal, walletID, trxID)
	}

	return nil
}

func (a *Activity) ChimoneyPayoutStatus(ctx context.Context, walletID string, chiRef string) (*external.PayoutStatusResponse, error) {
	chiWallet, err := GetChiWallet(ctx, a.b, walletID)
	if errors.Is(err, chimoney.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	return a.external.PayoutStatus(ctx, external.PayoutStatusRequest{
		ChiWallet:           chiWallet,
		TurnOffNotification: true,
		Reference:           chiRef,
	})
}

func (a *Activity) CreateChimoneyWithdrawalTransaction(ctx context.Context, walletID string, amount currency.Amount) (string, error) {
	trx, err := a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:    walletID,
		State:       transactions.StatePending,
		Provider:    chimoney.ProviderName,
		Amount:      amount,
		Title:       "Withdrawal",
		ForeignType: transactions.TransactionTypeWithdrawal,
	})
	if err != nil {
		return "", err
	}

	return trx, nil
}

func (a *Activity) CreateChimoneyDepositTransaction(ctx context.Context, walletID, issueID string) (string, error) {
	chiWalletID, err := GetChiWallet(ctx, a.b, walletID)
	if errors.Is(err, chimoney.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", err)
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	depositDetails, err := a.external.VerifyPayment(ctx, external.VerifyPaymentReq{
		IssueID:   issueID,
		ChiWallet: chiWalletID,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	amount, err := currency.FromString(depositDetails.Amount, currency.ParseCurrency(depositDetails.Currency))
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	// find chimoney balance account
	balances, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return "", err
	}

	var chimoneyBalance *linkedaccounts.LinkedAccount
	for _, la := range balances {
		if la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeBalance {
			chimoneyBalance = &la
			break
		}
	}
	if chimoneyBalance == nil {
		return "", temporal.NewNonRetryableApplicationError("chimoney deposit: no balance account found", "ErrInternal", nil)
	}

	trx, err := a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                walletID,
		State:                   transactions.StatePending,
		Provider:                chimoney.ProviderName,
		Amount:                  amount,
		Title:                   "Deposit",
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		ForeignType:             transactions.TransactionTypeDeposit,
		ForeignID:               issueID,
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: chimoneyBalance.ID,
				Type:            transactions.TransferTypeCreditBalance,
				Amount:          amount,
				State:           transactions.StatePending,
				ForeignID:       issueID,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return trx, nil
}

func (a *Activity) GetChimoneyTransactionByForeignID(ctx context.Context, walletID string, foreignID string) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransactionByForeignID(ctx, walletID, foreignID)
}

func (a *Activity) GetWalletIDFromChimoneyID(ctx context.Context, chimoneyID string) (string, error) {
	return GetWalletID(ctx, a.b, chimoneyID)
}
func (a *Activity) GetChimoneyTransaction(ctx context.Context, walletID, trxID string) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransaction(ctx, walletID, trxID)
}

func (a *Activity) SetWithdrawalTransfer(ctx context.Context, walletID, trxID string) error {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
	}

	// look up chimoney balance account
	balanceAccs, err := a.b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return err
	}

	var interaccAcc *linkedaccounts.LinkedAccount
	for _, bal := range balanceAccs {
		if bal.Provider == chimoney.ProviderName && bal.Type == chimoney.AccTypeInterac {
			interaccAcc = &bal
			break
		}
	}
	if interaccAcc == nil {
		return temporal.NewNonRetryableApplicationError("chimoney withdrawal: interac account not found", "ErrInternal", nil)
	}

	return a.b.Transactions().AddTransfers(ctx, trxID, []transactions.TransferArgs{
		{
			LinkedAccountID: interaccAcc.ID,
			Type:            transactions.TransferTypeCreditBankAccount,
			State:           transactions.StateCompleted,
			Amount:          trx.Amount,
		},
	})
}

func (a *Activity) ReserveChimoneyBalance(ctx context.Context, walletID, trxID string) error {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
	}

	// look up chimoney balance account
	balanceAccs, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return err
	}

	var chimoneyAcc *linkedaccounts.LinkedAccount
	for _, bal := range balanceAccs {
		if bal.Provider == chimoney.ProviderName && bal.Type == chimoney.AccTypeBalance {
			chimoneyAcc = &bal
			break
		}
	}
	if chimoneyAcc == nil {
		return temporal.NewNonRetryableApplicationError("chimoney withdrawal: balance account not found", "ErrInternal", nil)
	}

	_, err = ReserveBalance(ctx, a.b, chimoneyAcc.ID, trxID, trx.Amount, time.Hour*24*365)

	return err
}

func (a *Activity) FinalizeChimoneyBalance(ctx context.Context, trxID string) error {
	return FinaliseReserve(ctx, a.b, trxID)
}

func (a *Activity) RollbackChimoneyBalance(ctx context.Context, trxID string) error {
	return RollbackReserve(ctx, a.b, trxID)
}

func (a *Activity) AssignChimoneyBalance(ctx context.Context, walletID, trxID string) error {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
	}

	// look up chimoney balance account
	balanceAccs, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return err
	}

	var chimoneyAcc *linkedaccounts.LinkedAccount
	for _, bal := range balanceAccs {
		if bal.Provider == chimoney.ProviderName && bal.Type == chimoney.AccTypeBalance {
			chimoneyAcc = &bal
			break
		}
	}
	if chimoneyAcc == nil {
		return temporal.NewNonRetryableApplicationError("chimoney withdrawal: balance account not found", "ErrInternal", nil)
	}

	_, err = AssignBalance(ctx, a.b, chimoneyAcc.ID, trxID, trx.Amount)

	return err
}

func (a *Activity) ChimoneyWithdraw(ctx context.Context, walletID, trxID string) (*external.WithdrawResponse, error) {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return nil, err
	}

	return Withdraw(ctx, a.b, a.external, walletID, trx.Amount)
}

func (a *Activity) SetChimoneyTransactionForeignID(ctx context.Context, trxID, foreignID string) error {
	return a.b.Transactions().SetTransactionForeignID(ctx, trxID, foreignID)
}

func (a *Activity) CompleteChimoneyTransaction(ctx context.Context, walletID, trxID string) error {
	err := a.b.Transactions().SetTransactionState(ctx, trxID, transactions.StateCompleted)
	if err != nil {
		return err
	}

	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
	}

	users, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return err
	}
	if len(users) < 1 {
		return nil
	}

	email := users[0].Email
	if trx.Type == transactions.TransactionTypeWithdrawal {
		a.b.Email().SendWithdrawalEmail(ctx, walletID, trx.Amount, email, trx.Timestamp.Format("02 Jan 2006"))
	} else if trx.Type == transactions.TransactionTypeDeposit {
		a.b.Email().SendDepositReceivedEmail(ctx, walletID, trx.Amount, email, trx.Timestamp.Format("02 Jan 2006"))
	}

	return nil
}

func (a *Activity) FailChimoneyTransaction(ctx context.Context, walletID, trxID string) error {
	err := a.b.Transactions().SetTransactionState(ctx, trxID, transactions.StateFailed)
	if err != nil {
		return err
	}

	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
	}

	if trx.Type == transactions.TransactionTypeWithdrawal {
		a.b.Email().SendWithdrawalFailedEmail(ctx, walletID)
	} else if trx.Type == transactions.TransactionTypeDeposit {
		a.b.Email().SendDepositFailedEmail(ctx, walletID)
	}

	return nil
}

func rollBackWithdrawal(ctx workflow.Context, a *Activity, stage withdrawalStage, walletID, trxID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Rolling back chimoney withdrawal. stage:" + string(stage))

	switch stage {
	case updateTransaction, finalizeReserve:
		// we don't rollback anything as we have reserved liquidity and made external api call.
		// Rather notify slack for manual intervention
		slack.SendToChannel(
			context.Background(),
			slack.ChannelNotifyEvents,
			"wallet-info-bot",
			fmt.Sprintf("Chimoney withdrawal failed after external api call. walletID=%s, transactionID=%s", walletID, trxID),
		)
	case externalWithdrawalFailed:
		slack.SendToChannel(
			context.Background(),
			slack.ChannelNotifyEvents,
			"wallet-info-bot",
			fmt.Sprintf("Chimoney withdrawal failed on Chimoney side. walletID=%s, transactionID=%s", walletID, trxID),
		)
		fallthrough
	case externalWithdrawal:
		err := workflow.ExecuteActivity(ctx, a.RollbackChimoneyBalance, trxID).Get(ctx, nil)
		if err != nil {
			return err
		}
		fallthrough
	case reserveLiquidity:
		err := workflow.ExecuteActivity(ctx, a.FailChimoneyTransaction, walletID, trxID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateChimoneyDepositWorkflow(ctx workflow.Context, issueID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating chimoney deposit.")

	// Try get the Interledger walletID so we know that it is for one of our wallets
	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletIDFromIssueID, issueID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	return workflow.ExecuteActivity(ctx, a.CreateChimoneyDepositTransaction, walletID, issueID).Get(ctx, nil)
}

func FinishChimoneyDepositWorkflow(ctx workflow.Context, issueID string, status string, chiWalletID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Finishing chimoney deposit.")

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletIDFromChimoneyID, chiWalletID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	var trx transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetChimoneyTransactionByForeignID, walletID, issueID).Get(ctx, &trx)
	if err != nil {
		return err
	}

	var externalPayment external.Payment
	err = workflow.ExecuteActivity(ctx, a.VerifyChimoneyPayment, walletID, issueID).Get(ctx, &externalPayment)
	if err != nil {
		return err
	}

	switch externalPayment.Status {
	case "failed", "expired":
		err = workflow.ExecuteActivity(ctx, a.FailChimoneyTransaction, walletID, trx.ID).Get(ctx, nil)
		if err != nil {
			return err
		}

		return nil
	case "redeemed":
		// continue to finalize the transaction
		err = workflow.ExecuteActivity(ctx, a.AssignChimoneyBalance, walletID, trx.ID).Get(ctx, nil)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(ctx, a.CompleteChimoneyTransaction, walletID, trx.ID).Get(ctx, nil)
		if err != nil {
			return err
		}
	default:
		logger.Info("Chimoney deposit in unrecognized state: ", zap.String("status", externalPayment.Status))
		err = workflow.ExecuteActivity(ctx, a.FailChimoneyTransaction, walletID, trx.ID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) GetWalletIDFromIssueID(ctx context.Context, issueID string) (string, error) {
	details, err := a.external.VerifyPayment(ctx, external.VerifyPaymentReq{
		IssueID: issueID,
	})
	if err != nil {
		return "", err
	}

	return GetWalletID(ctx, a.b, details.SubAccount)
}

func (a *Activity) VerifyChimoneyPayment(ctx context.Context, walletID, issueID string) (*external.Payment, error) {
	chiWalletID, err := GetChiWallet(ctx, a.b, walletID)
	if errors.Is(err, chimoney.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	return a.external.VerifyPayment(ctx, external.VerifyPaymentReq{
		IssueID:   issueID,
		ChiWallet: chiWalletID,
	})
}
