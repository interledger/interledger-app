package ops

import (
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/slack/external"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Payments() payments.Client
}
