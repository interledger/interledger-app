package external

import "context"

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (*CreateTransactionResponse, error)
	RetrieveTransaction(ctx context.Context, id string, currency string) (*RetrieveTransactionResponse, error)
	CreateAccount(ctx context.Context, args CreateAccountArgs) (*CreateAccountResponse, error)
	RetrieveAccount(ctx context.Context, id string) (*RetrieveAccountResponse, error)
	QueryCard(ctx context.Context, args QueryCardArgs) (*QueryCardResponse, error)
	Init3DS(ctx context.Context, args Init3DSArgs) (*Init3DSResponse, error)
	Lookup3DS(ctx context.Context, args Lookup3DSArgs) (*Lookup3DSResponse, error)
	Authenticate3DS(ctx context.Context, args Authenticate3DSArgs) (*Authenticate3DSResponse, error)
	DeleteTransaction(ctx context.Context, id string, deleteType DeleteType, currency string) (*DeleteTransactionResponse, error)
}
