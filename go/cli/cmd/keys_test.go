package cmd_test

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/cli/cmd"
)

func TestCreateKeys(t *testing.T) {
	b := backends{
		config: viper.New(),
	}
	keyFile, err := os.CreateTemp("", "*.pem")
	require.NoError(t, err)
	configFile, err := os.CreateTemp("", "*.json")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(keyFile.Name())
		_ = os.Remove(configFile.Name())
	})
	b.config.Set("clientKeyPath", keyFile.Name())
	b.config.SetConfigType("json")
	b.config.SetConfigFile(configFile.Name())

	cases := []struct {
		Name  string
		Error string
		Args  []string
	}{
		{
			Name:  "Force overwrites key file",
			Args:  []string{"create", "-k", "test", "-f"},
			Error: "",
		},
		{
			Name:  "Default is to not overwrite key file",
			Args:  []string{"create", "-k", "test"},
			Error: "A key already exists. Use the `-f` flag to overwrite.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			createKeyCmd := cmd.NewKeysCmd(b)
			createKeyCmd.SetArgs(tc.Args)
			err := createKeyCmd.Execute()

			if tc.Error == "" {
				require.NoError(st, err)

				key, keyErr := os.ReadFile(keyFile.Name())
				require.NoError(st, keyErr)
				assert.NotEmpty(st, key)
				assert.Equal(st, "test", b.config.GetString("clientKeyID"))
			} else {
				assert.ErrorContains(st, err, tc.Error)
			}
		})
	}
}
