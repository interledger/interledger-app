package admin

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/waitlist"
)

type Backends interface {
	DB() *sqlx.DB
	AdminAuth() auth.Service
	Validator() *validator.Validate
	Waitlist() waitlist.Client
}
