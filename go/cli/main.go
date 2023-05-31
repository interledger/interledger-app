package main

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/cli/identities"
	identities_client "gitlab.com/fynbos/cli/identities/client"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/spf13/viper"
	"gitlab.com/fynbos/cli/cmd"
)

var cfgFile string

func main() {
	b := &backends{}
	rootCmd := cmd.NewCmdRoot(b)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fynbos/config.json)")

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

type backends struct {
	config     *viper.Viper
	httpClient *http.Client
	validator  *validator.Validate
	identities identities.Client
}

func (b *backends) Config() *viper.Viper {
	if b.config == nil {
		c := viper.New()
		if cfgFile != "" {
			// Use config file from the flag.
			c.SetConfigFile(cfgFile)
		} else {
			// Find home directory.
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatal(err)
			}

			fynbosDir := path.Join(home, ".fynbos")
			_, err = os.ReadDir(fynbosDir)
			if os.IsNotExist(err) {
				err := os.Mkdir(fynbosDir, os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			}
			if err != nil && !os.IsNotExist(err) {
				log.Fatal(err)
			}

			// Search config in home directory with name "config.json".
			fynbosConfigFile := path.Join(fynbosDir, "config.json")
			c.SetConfigType("json")
			c.SetConfigFile(fynbosConfigFile)

			c.SetDefault("clientKeyPath", path.Join(fynbosDir, "clientKey.pem"))

			f, _ := os.Stat(fynbosConfigFile)
			if f == nil {
				err := os.WriteFile(fynbosConfigFile, []byte("{}"), os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		c.AutomaticEnv()

		if err := c.ReadInConfig(); err != nil {
			log.Fatal("Failed to read config file. ", "error: ", err)
		}

		b.config = c
	}

	return b.config
}

func (b *backends) HttpClient() *http.Client {
	if b.httpClient == nil {
		b.httpClient = http.DefaultClient
	}

	return b.httpClient
}

func (b *backends) Validator() *validator.Validate {
	b.validator = validator.New()

	return b.validator
}

func (b *backends) Identities() identities.Client {
	client := identities_client.New()
	b.identities = client

	return b.identities
}
