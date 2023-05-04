package twitter

import (
	"context"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (*Authorization, error)
	CreateToken(ctx context.Context, args *CreateTokenArgs) (*Token, error)
	GetTokensByWalletID(ctx context.Context, args *GetTokensByWalletIDArgs) ([]Token, error)
}
