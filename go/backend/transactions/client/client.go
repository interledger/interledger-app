package client

import (
	"context"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/db"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
)

var _ transactions.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) transactions.Client {
	return &client{
		b: b,
	}
}

func (c *client) CreateTransaction(ctx context.Context, args transactions.CreateTransactionArgs) (string, error) {
	return ops.CreateTransaction(ctx, c.b, args)
}

func (c *client) CreateTransactionTx(ctx context.Context, tx *sqlx.Tx, args transactions.CreateTransactionArgs) (string, error) {
	return ops.CreateTransactionTx(ctx, c.b, tx, args)
}

func (c *client) AddTransfers(ctx context.Context, trxID string, args []transactions.TransferArgs) error {
	return ops.AddTransfers(ctx, c.b, trxID, args)
}

func (c *client) AddTransfersTx(ctx context.Context, tx *sqlx.Tx, trxID string, args []transactions.TransferArgs) error {
	return ops.AddTransfersTx(ctx, c.b, tx, trxID, args)
}

func (c *client) SetTransactionForeignID(ctx context.Context, ID string, foreignID string) error {
	return ops.SetTransactionForeignID(ctx, c.b, ID, foreignID)
}

func (c *client) SetTransactionDestination(ctx context.Context, id, destination string) error {
	return ops.SetTransactionDestination(ctx, c.b, id, destination)
}

func (c *client) SetTransactionRefundState(ctx context.Context, id string, state transactions.RefundState) error {
	return ops.SetTransactionRefundState(ctx, c.b, id, state)
}

func (c *client) SetTransferForeignID(ctx context.Context, ID string, foreignID string) error {
	return ops.SetTransferForeignID(ctx, c.b, ID, foreignID)
}

func (c *client) SetTransactionState(ctx context.Context, ID string, state transactions.State) error {
	return ops.SetTransactionState(ctx, c.b, ID, state)
}

func (c *client) SetTransactionStateTx(ctx context.Context, tx *sqlx.Tx, ID string, state transactions.State) error {
	return ops.SetTransactionStateTx(ctx, c.b, tx, ID, state)
}

func (c *client) SetTransferState(ctx context.Context, ID string, state transactions.State) error {
	return ops.SetTransferState(ctx, c.b, ID, state)
}

func (c *client) SetTransactionAmountTx(ctx context.Context, tx *sqlx.Tx, ID string, amount currency.Amount) error {
	return ops.SetTransactionAmountTx(ctx, c.b, tx, ID, amount)
}

func (c *client) List(ctx context.Context, page db.Pagination, walletID string) ([]transactions.Transaction, error) {
	return ops.List(ctx, c.b, walletID, page)
}

func (c *client) ListCompleted(ctx context.Context, page db.Pagination, walletID string) ([]transactions.Transaction, error) {
	return ops.ListCompleted(ctx, c.b, walletID, page)
}

func (c *client) ListWithPending(ctx context.Context, page db.Pagination, walletID string) ([]transactions.Transaction, error) {
	return ops.ListWithPending(ctx, c.b, walletID, page)
}

func (c *client) ListTransactionsInRange(ctx context.Context, walletID string, inRange transactions.TransactionRangeFilter) ([]transactions.Transaction, error) {
	return ops.ListTransactionsInRange(ctx, c.b, walletID, inRange)
}

func (c *client) GetTransaction(ctx context.Context, walletID string, txID string) (*transactions.Transaction, error) {
	return ops.GetTransaction(ctx, c.b, walletID, txID)
}

func (c *client) GetTransactionByForeignID(ctx context.Context, walletID string, foreignID string) (*transactions.Transaction, error) {
	return ops.GetTransactionByForeignID(ctx, c.b, walletID, foreignID)
}

func (c *client) ListTransfers(ctx context.Context, trxID string) ([]transactions.Transfer, error) {
	return ops.ListTransfers(ctx, c.b, trxID)
}

func (c *client) GetHasTransacted(ctx context.Context, walletID, destination string) (bool, error) {
	return ops.GetHasTransacted(ctx, c.b, walletID, destination)
}

func (c *client) GetTransactedCount(ctx context.Context, walletID, destination string) (int, error) {
	return ops.GetTransactedCount(ctx, c.b, walletID, destination)
}

func (c *client) CountReceiveTransactions(ctx context.Context, walletID string) (int, error) {
	return ops.CountReceiveTransactions(ctx, c.b, walletID)
}

func (c *client) CountSendTransactions(ctx context.Context, walletID string) (int, error) {
	return ops.CountSendTransactions(ctx, c.b, walletID)
}

func (c *client) ListAll(ctx context.Context, page db.Pagination) ([]transactions.Transaction, error) {
	return ops.ListAllTransactions(ctx, c.b, page)
}

func (c *client) ListTransactionsForCard(ctx context.Context, page db.Pagination, walletID, cardID string) ([]transactions.Transaction, error) {
	return ops.ListTransactionsForCard(ctx, c.b, page, walletID, cardID)
}
