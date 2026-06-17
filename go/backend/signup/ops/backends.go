package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/twilio"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Twilio() twilio.Service
}
