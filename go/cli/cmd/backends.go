package cmd

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/cli/identities"
	"net/http"

	"github.com/spf13/viper"
)

type Backends interface {
	Config() *viper.Viper
	HttpClient() *http.Client
	Validator() *validator.Validate
	Identities() identities.Client
}
