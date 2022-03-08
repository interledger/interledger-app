package authorization

import (
	_ "embed"

	"github.com/osohq/go-oso"
	"gitlab.com/fynbos/backend/user"
)

//go:embed main.polar
var policy []byte

func NewService() (*oso.Oso, error) {
	o, err := oso.NewOso()
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
