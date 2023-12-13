package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"

	"gitlab.com/fynbos/backend/currency"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
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
	var ptiPrivateKey any
	if env.IsLocal() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatalln(err)
		}

		ptiPrivateKey = privateKey
	} else {
		// pti TODO: parsing private key might change depending on file format
		privateKeyPEM, err := os.ReadFile(os.Getenv("PTI_PRIVATE_KEY_PATH"))
		if err != nil {
			log.Fatalln(err)
		}

		privateKey, _ := pem.Decode(privateKeyPEM)
		ptiPrivateKey, err = x509.ParsePKCS8PrivateKey(privateKey.Bytes)
		if err != nil {
			log.Fatalln(err)
		}
	}
	external := external.New(external.ClientArgs{
		Transport: &http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		},
		ClientID:   os.Getenv("PTI_CLIENT_ID"),
		PrivateKey: ptiPrivateKey,
	})

	return &Client{b, external}
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
