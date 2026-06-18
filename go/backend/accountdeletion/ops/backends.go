package ops

import "github.com/jmoiron/sqlx"

type Backends interface {
	DB() *sqlx.DB
}
