package authorization

import (
	_ "embed"
	"github.com/osohq/go-oso"
	org "gitlab.com/fynbos/backend/organisation"
	"gitlab.com/fynbos/backend/user"
)

//go:embed main.polar
var policy []byte

func NewService() (*oso.Oso, error) {
	o, err := oso.NewOso()
	if err != nil {
		return nil, err
	}

	o.RegisterClass(org.Organisation{}, nil)
	o.RegisterClass(user.User{}, nil)

	err = o.LoadString(string(policy))
	if err != nil {
		return nil, err
	}

	return &o, nil
}
