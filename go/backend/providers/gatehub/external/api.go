package external

import "context"

type Client interface {
	IssueToken(ctx context.Context, product Product) (*IssueTokenResponse, error)
}
