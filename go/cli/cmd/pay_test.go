package cmd_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/cli/cmd"
)

func TestCreateOutgoingPayment(t *testing.T) {
	b := NewTestBackends(t)
	mockOpenPaymentsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/timmy" && r.Method == "POST" {
			// redirect to auth server
			http.Redirect(w, r, fmt.Sprintf("http://%s/grant", r.Host), http.StatusTemporaryRedirect)
			return
		}

		if r.URL.Path == "/grant" && r.Method == "POST" {
			g := cmd.Grant{
				ID:     uuid.NewString(),
				Client: fmt.Sprintf("http://%s/timmy", r.Host),
				Tokens: []cmd.AccessToken{
					{
						Value: uuid.NewString(),
						Access: []cmd.Access{
							{
								Type:      "outgoing-payment",
								Locations: []string{fmt.Sprintf("http://%s/outgoing", r.Host)},
							},
						},
					},
				},
			}
			jsonGrant, err := json.Marshal(g)
			require.NoError(t, err)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonGrant)
			return
		}

		if r.URL.Path == "/outgoing" && r.Method == "POST" {
			op := cmd.OutgoingPayment{
				ID: fmt.Sprintf("http://%s/%s", r.Host, uuid.NewString()),
			}
			jsonOp, err := json.Marshal(op)
			require.NoError(t, err)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonOp)
			return
		}

		w.WriteHeader(404)
	}))
	t.Cleanup(func() {
		mockOpenPaymentsServer.Close()
	})
	b.config.Set("wallet", fmt.Sprintf("%s/timmy", mockOpenPaymentsServer.URL))

	payCmd := cmd.NewPayCmd(b)
	payCmd.SetArgs([]string{fmt.Sprintf("%s/timmy", mockOpenPaymentsServer.URL), "-a", "10"})

	err := payCmd.Execute()
	require.NoError(t, err)
}

type backends struct {
	config     *viper.Viper
	httpClient *http.Client
}

func (b backends) Config() *viper.Viper {
	return b.config
}

func (b backends) HttpClient() *http.Client {
	return b.httpClient
}

func NewTestBackends(t *testing.T) *backends {
	b := &backends{
		config:     viper.New(),
		httpClient: http.DefaultClient,
	}
	b.config.Set("clientKeyPath", NewSavedPrivateKey(t))
	b.config.Set("clientKeyID", "test")
	b.config.AddConfigPath(os.TempDir())
	b.config.SetConfigType("json")

	return b
}

func NewSavedPrivateKey(t *testing.T) string {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	keyBytes := bytes.NewBuffer(nil)
	err = pem.Encode(keyBytes, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Key,
	})
	require.NoError(t, err)

	file, err := os.CreateTemp("", "*-key.pem")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(file.Name())
	})
	err = os.WriteFile(file.Name(), keyBytes.Bytes(), os.ModePerm)
	require.NoError(t, err)

	return file.Name()
}
