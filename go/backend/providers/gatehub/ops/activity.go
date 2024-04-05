package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends) *Activity {
	ec := external.NewClient(
		os.Getenv("GATEHUB_APP_ID"),
		os.Getenv("GATEHUB_SECRET"),
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		},
	)

	return &Activity{
		b:        b,
		external: ec,
	}
}

func (a *Activity) CreateGatehubUser(ctx context.Context, walletID string) (string, error) {
	ul, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(ul) < 1 {
		return "", fmt.Errorf("%w No Fynbos user found for walletID", gatehub.ErrInternal)
	}

	resp, err := a.external.CreateUser(ctx, ul[0].Email)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return resp.ID, nil
}

func (a *Activity) SaveGatehubUser(ctx context.Context, walletID, externalID string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;", externalID, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return nil
}

func (a *Activity) GetGatehubUser(ctx context.Context, walletID string) (string, error) {
	var externalID string
	err := a.b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", temporal.NewNonRetryableApplicationError("Gatehub user not found for wallet", "ErrNotFound", err)
	} else if err != nil {
		return "", err
	}

	return externalID, nil
}

func (a *Activity) CreateGatehubWalletLinkedAccount(ctx context.Context, walletID string) (*linkedaccounts.LinkedAccount, error) {
	w, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, wallets.ErrNoWalletFound) {
		return nil, temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", err)
	} else if err != nil {
		return nil, err
	}

	externalID, err := getExternalUserID(ctx, a.b, w.ID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Gatehub user not found for wallet", "ErrNotFound", err)
	} else if err != nil {
		return nil, err
	}

	userWallets, err := a.external.GetUserWallets(ctx, externalID)
	if err != nil {
		return nil, err
	}

	var primaryWallet *external.Wallet
	for _, w := range userWallets.Wallets {
		if w.Primary {
			primaryWallet = &w
			break
		}
	}
	if primaryWallet == nil {
		return nil, fmt.Errorf("%w Could not find a primary wallet for gatehub user", gatehub.ErrInternal)
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, w.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	for _, la := range las {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			return &la, nil
		}
	}

	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:        w.ID,
		Type:            gatehub.AccTypeBalance,
		Provider:        gatehub.ProviderName,
		ProviderID:      primaryWallet.UUID,
		Name:            "EUR Balance",
		Nickname:        "EUR Balance",
		CanReceive:      true,
		ReceiveCountry:  w.Country,
		ReceiveCurrency: currency.EUR,
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

func (a *Activity) CreateGatehubBalanceAccount(ctx context.Context, id string) error {
	accs, err := a.b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:                         id,
			LedgerID:                   gatehub.LedgerIDEUR,
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
		return fmt.Errorf("%w failed to setup account status(%s)", pti.ErrInternal, accs[0].Code)
	}

	return nil
}
