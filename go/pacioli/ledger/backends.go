package ledger

import (
	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	TigerBeetle() tigerbeetle_go.Client
	Validator() *validator.Validate
}
