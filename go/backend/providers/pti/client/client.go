package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/backend/providers/pti/ops"
)

var _ pti.Client = &Client{}

type Client struct {
	b        ops.Backends
	external external.Client
}

func New(b ops.Backends) *Client {
	ptiPrivateKey, err := jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
	if err != nil {
		log.Fatalln(err)
	}
	ptiExternal, err := external.NewWithOptions(
		external.WithBaseURL(os.Getenv("PTI_BASE_URL")),
		external.WithOTELLHTTPClient(),
		external.WithClientID(os.Getenv("PTI_CLIENT_ID")),
		external.WithDerivedKeys(ptiPrivateKey),
	)
	if err != nil {
		log.Fatalln(fmt.Errorf("%w Failed to create PTI external client: %s", pti.ErrInternal, err))
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

func (c Client) GetWidget(ctx context.Context, walletID string) (*pti.WidgetDetails, error) {
	return ops.GetKYCWidget(ctx, c.b, walletID)
}

func (c Client) CreateCard(ctx context.Context, walletID, tokenID string) (pti.Await, error) {
	return ops.CreateCard(ctx, c.b, walletID, tokenID)
}

func (c Client) GetLinkedAccountCardDetails(ctx context.Context, id string) (*pti.EncryptedCreditCardPaymentInformation, error) {
	return ops.GetLinkedAccountCardDetails(ctx, c.b, c.external, id)
}

func (c Client) CreateBankAccount(ctx context.Context, args pti.CreateBankAccountArgs) (pti.Await, error) {
	return ops.CreateBankAccount(ctx, c.b, args)
}

func (c Client) CreateDeposit(ctx context.Context, la *linkedaccounts.LinkedAccount, amount currency.Amount, note string) error {
	return ops.CreateDeposit(ctx, c.b, la, amount, note)
}
