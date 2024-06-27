package ops

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	return exID, nil
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
