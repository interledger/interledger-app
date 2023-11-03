package pacioli

import "context"

type Client interface {

	// This is declaritive and will not fail if the ledger exists. It will fail if one exists with
	// different fields.
	ConfigureLedgers(ctx context.Context, args []ConfigureLedgerArgs) ([]LedgerResult, error)
	GetLedgers(ctx context.Context, ledgerIDs []uint32) ([]Ledger, error)

	// This is declaritive and will not fail if the account exists. It will fail if one exists with
	// different fields.
	ConfigureAccounts(ctx context.Context, args []ConfigureAccountArgs) ([]AccountResult, error)
	GetAccounts(ctx context.Context, accountIDs []string) ([]Account, error)

	CreateTransfers(ctx context.Context, args []CreateTransferArgs) ([]TransferResult, error)
	GetTransfers(ctx context.Context, transferIDs []string) ([]Transfer, error)
	PostTransfers(ctx context.Context, transferIDs []string) ([]TransferResult, error)
	VoidTransfers(ctx context.Context, transferIDs []string) ([]TransferResult, error)
}
