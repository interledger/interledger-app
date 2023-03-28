package external

import "context"

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (*CreateTransactionResponse, error)
	RetrieveTransaction(ctx context.Context, id string) (*RetrieveTransactionResponse, error)
	CreateAccount(ctx context.Context, args CreateAccountArgs) (*CreateAccountResponse, error)
	RetrieveAccount(ctx context.Context, id string) (*RetrieveAccountResponse, error)
	QueryCard(ctx context.Context, args QueryCardArgs) (*QueryCardResponse, error)
}
