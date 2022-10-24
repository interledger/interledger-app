package actions

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	DB() *sqlx.DB
	Kratos() *kratos.APIClient
	Signup() signup.Client
	User() user.Client
	Validator() *validator.Validate
}
