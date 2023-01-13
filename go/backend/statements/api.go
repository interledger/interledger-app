package statements

import (
	"context"
)

type Client interface {
	GenerateWalletStatementPDF(ctx context.Context, args GenerateWalletStatementArgs) ([]byte, error)
}
