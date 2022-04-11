package authorization

import (
	"embed"
	"reflect"

	"github.com/osohq/go-oso"
	_accounts "gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/user"
)

//go:embed *.polar
var f embed.FS

func NewService() (*oso.Oso, error) {
	o, err := oso.NewOso()
	if err != nil {
		return nil, err
	}

	policy, err := f.ReadFile("main.polar")
	if err != nil {
		return nil, err
	}

	err = o.RegisterClass(user.User{}, nil)
	if err != nil {
		return nil, err
	}

	err = o.LoadString(string(policy))
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func NewAdminService() (*oso.Oso, error) {
	o, err := oso.NewOso()
	if err != nil {
		return nil, err
	}

	adminPolicy, err := f.ReadFile("admin.polar")
	if err != nil {
		return nil, err
	}

	err = o.RegisterClass(reflect.TypeOf(auth.AdminUser{}), nil)
	if err != nil {
		return nil, err
	}

	err = o.RegisterClass(reflect.TypeOf(_accounts.Account{}), nil)
	if err != nil {
		return nil, err
	}

	err = o.LoadString(string(adminPolicy))
	if err != nil {
		return nil, err
	}

	return &o, nil
}
