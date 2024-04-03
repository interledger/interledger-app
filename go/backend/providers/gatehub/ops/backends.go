package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/user"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	Temporal() temporal.Client
}
