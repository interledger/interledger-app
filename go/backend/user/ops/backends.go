package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Kratos() *kratos.APIClient
	Analytics() analytics.Client
	Keys() keys.Client
}
