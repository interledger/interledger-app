package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
}
