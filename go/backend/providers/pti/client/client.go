package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"gitlab.com/fynbos/backend/currency"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	external_mock "gitlab.com/fynbos/backend/providers/pti/external/mock"
	"gitlab.com/fynbos/backend/providers/pti/ops"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ pti.Client = &Client{}

type Client struct {
	b        ops.Backends
	external external.Client
}

func New(b ops.Backends) *Client {
	var ptiPrivateKey jwk.Key
	var err error
	var ptiExternal external.Client
	if env.IsTest() {
		ptiExternal = external_mock.SetupDevMock(nil)
	} else if env.IsLocal() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatalln(err)
		}

		ptiPrivateKey, err = jwk.FromRaw(privateKey)
		if err != nil {
			log.Fatalln(err)
		}
		ptiExternal = external.New(external.ClientArgs{
			Transport: &http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, nil),
				),
			},
			ClientID:   "LOCAL",
			PrivateKey: ptiPrivateKey,
		})
	} else {
		ptiPrivateKey, err = jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
		if err != nil {
			log.Fatalln(err)
		}
		ptiExternal = external.New(external.ClientArgs{
			Transport: &http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, nil),
				),
			},
			ClientID:   os.Getenv("PTI_CLIENT_ID"),
			PrivateKey: ptiPrivateKey,
		})
	}

	return &Client{b: b, external: ptiExternal}
}

func (c Client) CreateWallet(ctx context.Context, walletID string, currency currency.Currency) (pti.Await, error) {
	return ops.CreateWallet(ctx, c.b, pti.CreateWalletArgs{
		WalletID: walletID,
		Currency: currency,
		Nickname: "USD Balance",
		Title:    "USD Balance",
	})
}

func (c Client) GetWallet(ctx context.Context, linkedAccountID string) (*pti.Wallet, error) {
	return ops.GetWallet(ctx, c.b, c.external, linkedAccountID)
}

func (c Client) GetBalance(ctx context.Context, linkedAccountID string) (*pti.Balance, error) {
	return ops.GetBalance(ctx, c.b, linkedAccountID)
}

func (c Client) DepositToWallet(ctx context.Context, args pti.TransactionArgs) (string, error) {
	return ops.DepositToWallet(ctx, c.b, c.external, args)
}

func (c Client) WithdrawalFromWallet(ctx context.Context, args pti.TransactionArgs) (string, error) {
	return ops.WithdrawFromWallet(ctx, c.b, c.external, args)
}

func (c Client) UpdateTransactionStatus(ctx context.Context, args pti.TransactionStatusArgs) error {
	return ops.UpdateTransactionStatus(ctx, c.b, c.external, args)
}

func (c Client) ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*pti.Balance, error) {
	return ops.ReserveBalance(ctx, c.b, linkedAccountID, txID, amt, timeout)
}

func (c Client) FinaliseReserve(ctx context.Context, trxID string) error {
	return ops.FinaliseReserve(ctx, c.b, trxID)
}

func (c Client) AssignBalance(ctx context.Context, linkedAccountID string, txID string, amt currency.Amount) (*pti.Balance, error) {
	return ops.AssignBalance(ctx, c.b, linkedAccountID, txID, amt)
}

func (c Client) RollbackReserve(ctx context.Context, trxID string) error {
	return ops.RollbackReserve(ctx, c.b, trxID)
}

func (c Client) ReserveTransfer(ctx context.Context, fromAccount, toAccount, txID string, amt currency.Amount, timeout time.Duration) error {
	return ops.ReserveTransfer(ctx, c.b, fromAccount, toAccount, txID, amt, timeout)
}

func (c Client) CreateJWT(ctx context.Context, args pti.TokenArgs) (*pti.TokenResponse, error) {
	return c.external.CreateJWT(ctx, args)
}

func (c Client) GetKYCWidget(ctx context.Context, walletID string) (*pti.KYCWidgetDetails, error) {
	return ops.GetKYCWidget(ctx, c.b, walletID)
}
