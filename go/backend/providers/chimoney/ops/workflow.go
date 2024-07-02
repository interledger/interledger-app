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
	"gitlab.com/fynbos/backend/wallets"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
