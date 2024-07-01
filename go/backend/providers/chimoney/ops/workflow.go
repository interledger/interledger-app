package ops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/pacioli"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

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

	err = workflow.ExecuteActivity(ctx, a.CreateBalanceAccount, walletID, exID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return exID, nil
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
		SendCurrency:    currency.EUR,
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

func (a *Activity) CreateChimoneyWallet(ctx context.Context, walletID string) (string, error) {
	ul, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(ul) < 1 {
		return "", fmt.Errorf("%w No Fynbos user found for walletID", chimoney.ErrInternal)
	}

	userInfo, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	email, err := GetInteracEmail(ctx, a.b, walletID)
	if err != nil {
		return "", err
	}

	exID, err := a.external.CreateWallet(ctx, external.CreateWalletReq{
		Name:        userInfo.FirstName + " " + userInfo.LastName,
		Email:       ul[0].Email,
		FirstName:   userInfo.FirstName,
		LastName:    userInfo.LastName,
		PhoneNumber: email,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return exID, nil
}

func CreateChimoneyWithdrawalWorkflow(
	ctx workflow.Context, walletID string, amount currency.Amount,
) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating chimoney withdrawal.")

	// create transaction in pending state
	var trxID string
	err := workflow.ExecuteActivity(ctx, a.CreateChimoneyWithdrawalTransaction, walletID, amount).Get(ctx, &trxID)
	if err != nil {
		return "", err
	}

	// start child workflow but don't wait for it to complete
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            fmt.Sprintf("chimoney_execute_withdrawal_%s", trxID),
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_ABANDON, // run independently of parent
	}
	childCtx := workflow.WithChildOptions(ctx, cwo)
	_ = workflow.ExecuteChildWorkflow(childCtx, ExecuteChimoneyWithdrawalWorkflow, walletID, trxID).GetChildWorkflowExecution()

	return trxID, nil
}

type withdrawalStage uint8

var (
	reserveLiquidity   withdrawalStage = 0
	externalWithdrawal withdrawalStage = 1
	finalizeReserve    withdrawalStage = 2
	updateTransaction  withdrawalStage = 3
)

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
	if trx.State != transactions.StatePending || !(trx.Provider == chimoney.ProviderName && trx.Type == transactions.TransactionTypeWithdrawal) {
		return temporal.NewNonRetryableApplicationError("chimoney withdrawal: transaction is either not pending or not a chimoney withdrawal", "ErrInternal", nil)
	}

	// reserve liquidity and surface Insufficient balance errors.
	err = workflow.ExecuteActivity(ctx, a.ReserveChimoneyBalance, walletID, trxID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, reserveLiquidity, walletID, trxID)
	}

	// perform the withdrawal. Rollback on failure
	err = workflow.ExecuteActivity(ctx, a.ChimoneyWithdraw, walletID, trxID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, externalWithdrawal, walletID, trxID)
	}

	// wait until it is succesful?

	// finalize balance
	err = workflow.ExecuteActivity(ctx, a.FinalizeChimoneyBalance, trxID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, finalizeReserve, walletID, trxID)
	}

	// TODO: update transaction foreignID
	err = workflow.ExecuteActivity(ctx, a.CompleteChimoneyTransaction, walletID, trxID).Get(ctx, nil)
	if err != nil {
		return rollBackWithdrawal(ctx, a, updateTransaction, walletID, trxID)
	}

	return nil
}

func (a *Activity) CreateChimoneyWithdrawalTransaction(ctx context.Context, walletID string, amount currency.Amount) (string, error) {
	trx, err := a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID: walletID,
		State:    transactions.StatePending,
		Provider: chimoney.ProviderName,
		Amount:   amount,
		Title:    "Withdrawal",
	})
	if err != nil {
		return "", err
	}

	return trx, nil
}

func (a *Activity) GetChimoneyTransaction(ctx context.Context, walletID, trxID string) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransaction(ctx, walletID, trxID)
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

func (a *Activity) ChimoneyWithdraw(ctx context.Context, walletID, trxID string) error {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
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

	interacEmail, err := GetInteracEmail(ctx, a.b, walletID)
	if err != nil {
		return err
	}

	a.b.Email().SendWithdrawalEmail(ctx, walletID, trx.Amount, interacEmail, trx.Timestamp.Format("02 Jan 2006"))

	return nil
}

func (a *Activity) FailChimoneyTransaction(ctx context.Context, walletID, trxID string) error {
	err := a.b.Transactions().SetTransactionState(ctx, trxID, transactions.StateFailed)
	if err != nil {
		return err
	}

	a.b.Email().SendWithdrawalFailedEmail(ctx, walletID)

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
			"Fynbot",
			fmt.Sprintf("Chimoney withdrawal failed after external api call. walletID=%s, transactionID=%s", walletID, trxID),
		)
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
