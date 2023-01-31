package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/analytics"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Kratos() *kratos.APIClient
	Analytics() analytics.Client
}
