package cmd

import (
	"net/http"

	"github.com/spf13/viper"
)

type Backends interface {
	Config() *viper.Viper
	HttpClient() *http.Client
}
