package client

import (
	"context"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_client "gitlab.com/fynbos/backend/providers/tabapay/external/client"
	mock_client "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay/ops"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	temporal "go.temporal.io/sdk/client"
)

var _ tabapay.Client = &Client{}

type NewClientArgs struct {
	BasisTheoryProxyApiKey string
	ClientID               string
	SubClientID            string
	BearerToken            string
	SettlementAccountID    string
}

type Backends interface {
	DB() *sqlx.DB
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Temporal() temporal.Client
}

type opsBackends struct {
	external external.Client
	b        Backends
}

func (ob *opsBackends) DB() *sqlx.DB {
	return ob.b.DB()
}

func (ob *opsBackends) External() external.Client {
	return ob.external
}

func (ob *opsBackends) KYC() kyc.Client {
	return ob.b.KYC()
}

func (ob *opsBackends) LinkedAccounts() linkedaccounts.Client {
	return ob.b.LinkedAccounts()
}

func (ob *opsBackends) Temporal() temporal.Client {
	return ob.b.Temporal()
}

func New(args NewClientArgs, b Backends) (*Client, error) {
	var fynbosAddress *external.Address
	var externalClient external.Client
	if env.IsLocal() {
		externalClient = mock_client.SetupDevMock(nil)
	} else {
		var err error
		externalClient, err = external_client.New(external_client.NewClientArgs{
			BasisTheoryProxyApiKey: args.BasisTheoryProxyApiKey,
			ClientID:               args.ClientID,
			SubClientID:            args.SubClientID,
			BearerToken:            args.BearerToken,
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		})
		if err != nil {
			return nil, err
		}

		if os.Getenv("FYNBOS_ADDRESS_STATE") != "" && os.Getenv("FYNBOS_ADDRESS_CITY") != "" &&
			os.Getenv("FYNBOS_ADDRESS_ZIPCODE") != "" && os.Getenv("FYNBOS_ADDRESS_COUNTRY") != "" {
			fynbosAddress = &external.Address{
				City:    os.Getenv("FYNBOS_ADDRESS_CITY"),
				State:   os.Getenv("FYNBOS_ADDRESS_STATE"),
				ZipCode: os.Getenv("FYNBOS_ADDRESS_ZIPCODE"),
				Country: os.Getenv("FYNBOS_ADDRESS_COUNTRY"),
			}
		} else {
			log.Warn("Fynbos address not configured. Soft descriptors will not be sent to Tabapay")
		}
	}

	if fynbosAddress == nil {
		log.Warn("tabapay client is not configured with Fynbos address. Soft descriptor won't be sent.")
	}

	return &Client{
		b: &opsBackends{
			external: externalClient,
			b:        b,
		},
		settlementAccountID: args.SettlementAccountID,
		fynbosAddress:       fynbosAddress,
	}, nil
}

type Client struct {
	b                   ops.Backends
	settlementAccountID string
	fynbosAddress       *external.Address
}

func (c *Client) CreateCard(ctx context.Context, args tabapay.CreateCardArgs) (tabapay.Await, error) {
	return ops.CreateCard(ctx, c.b, args)
}

func (c *Client) PullFromCard(ctx context.Context, args tabapay.PullFromCardArgs) (*tabapay.Transaction, error) {
	var softDescriptor *external.SoftDescriptor
	if args.SoftDescriptor != "" && c.fynbosAddress != nil {
		softDescriptor = &external.SoftDescriptor{
			Name:    args.SoftDescriptor,
			Address: c.fynbosAddress,
		}
	} else {
		log.Warn("Fynbos address not configured. Soft descriptors will not be sent to Tabapay")
	}

	return ops.PullFromCard(ctx, c.b, ops.PullFromCardArgs{
		WalletID:            args.WalletID,
		ProviderID:          args.ProviderID,
		ReferenceID:         args.ReferenceID,
		Amount:              args.Amount,
		SettlementAccountID: c.settlementAccountID,
		ThreeDSID:           args.ThreeDSID,
		SoftDescriptor:      softDescriptor,
	})
}

func (c *Client) PushToCard(ctx context.Context, args tabapay.PushToCardArgs) (*tabapay.Transaction, error) {
	var softDescriptor *external.SoftDescriptor
	if args.SoftDescriptor != "" && c.fynbosAddress != nil {
		softDescriptor = &external.SoftDescriptor{
			Name:    args.SoftDescriptor,
			Address: c.fynbosAddress,
		}
	} else {
		log.Warn("Fynbos address not configured. Soft descriptors will not be sent to Tabapay")
	}

	return ops.PushToCard(ctx, c.b, ops.PullFromCardArgs{
		WalletID:            args.WalletID,
		ProviderID:          args.ProviderID,
		ReferenceID:         args.ReferenceID,
		Amount:              args.Amount,
		SettlementAccountID: c.settlementAccountID,
		ThreeDSID:           args.ThreeDSID,
		SoftDescriptor:      softDescriptor,
	})
}

func (c *Client) GetTransaction(ctx context.Context, id string) (*tabapay.Transaction, error) {
	return ops.GetTransaction(ctx, c.b, id)
}

func (c *Client) Init3DS(ctx context.Context, args tabapay.Init3DSArgs) (*tabapay.Init3DSResponse, error) {
	return ops.Init3DS(ctx, c.b, args)
}

func (c *Client) Lookup3DS(ctx context.Context, args tabapay.Lookup3DSArgs) (*tabapay.Lookup3DSResponse, error) {
	return ops.Lookup3DS(ctx, c.b, args)
}

func (c *Client) Authenticate3DS(ctx context.Context, args tabapay.Authenticate3DSArgs) (*tabapay.Authenticate3DSResponse, error) {
	return ops.Authenticate3DS(ctx, c.b, args)
}

func (c *Client) Get3DSSession(ctx context.Context, id string) (*tabapay.ThreeDSSession, error) {
	return ops.Get3DSSession(ctx, c.b, id)
}

func (c *Client) ReverseTransaction(ctx context.Context, id string, txSettled bool) error {
	return ops.ReverseTransaction(ctx, c.b, id, txSettled)
}
