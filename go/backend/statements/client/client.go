package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
)

var _ statements.Client = client{}

type client struct{}

func New() *client {
	return &client{}
}

func (c client) GenerateWalletStatementPDF(ctx context.Context, wallet *machnet.Wallet, transactions []transactions.Transaction) ([]byte, error) {
	return []byte{}, nil
}
