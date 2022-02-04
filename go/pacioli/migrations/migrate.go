package migrations

import (
	"embed"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Migrate(fs *embed.FS) error {
	serviceName := "pacioli"
	baseConnString := os.Getenv("DB_URL")
	if baseConnString == "" {
		baseConnString = fmt.Sprintf("cockroach://%s@cockroachdb-public:26257/%s?sslmode=verify-full&max_conns=20&max_idle_conns=4", serviceName, serviceName)
	}

	// We read the ssl certs into memory and add them inline to the connection string.
	// This side steps the file permission issues (postgres spec requires
	// private key to have 0600) when mounting secrets into the pod and running
	// as non-root user.
	// https://github.com/hashicorp/vault/issues/10925
	// https://github.com/lib/pq/commit/b9bb726ebf154627a21b50f9ffa4b28c6ed3f4d8
	connString, err := InlineSslCreds(
		baseConnString,
		fmt.Sprintf("/cockroach-certs/client.%s.key", serviceName),
		fmt.Sprintf("/cockroach-certs/client.%s.crt", serviceName),
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		return err
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return err
	}

	// The expected behaviour is for it to return ErrNoChange
	if err := m.Up(); err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// We read the ssl certs into memory and add them inline to the connection string.
// This side steps the file permission issues (postgres spec requires
// private key to have 0600) when mounting secrets into the pod and running
// as non-root user.
// https://github.com/hashicorp/vault/issues/10925
// https://github.com/lib/pq/commit/b9bb726ebf154627a21b50f9ffa4b28c6ed3f4d8
func InlineSslCreds(baseUrl string, pathToPrivateKey string, pathToClientCert string, pathToRootCert string) (string, error) {
	connString := baseUrl + "&sslinline=true"
	sslkeyBytes, err := ioutil.ReadFile(pathToPrivateKey)
	if err != nil {
		return "", err
	}

	sslcertBytes, err := ioutil.ReadFile(pathToClientCert)
	if err != nil {
		return "", err
	}

	sslrootcertBytes, err := ioutil.ReadFile(pathToRootCert)
	if err != nil {
		return "", err
	}

	connString += "&sslinline=true"
	connString += "&sslkey=" + url.QueryEscape(string(sslkeyBytes))
	connString += "&sslcert=" + url.QueryEscape(string(sslcertBytes))
	connString += "&sslrootcert=" + url.QueryEscape(string(sslrootcertBytes))

	return connString, nil
}
