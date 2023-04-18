package ops

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type (
	Backends interface {
		DB() *sqlx.DB
	}
)

func CreateAuthURL(ctx context.Context, b Backends) (string, error) {
	return "", nil
}
